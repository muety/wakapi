package main

//go:generate $GOPATH/bin/pkger

import (
	"github.com/emvi/logbuch"
	"github.com/gorilla/handlers"
	"github.com/markbates/pkger"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/migrations"
	"github.com/muety/wakapi/repositories"
	"gorm.io/gorm/logger"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/muety/wakapi/middlewares"
	customMiddleware "github.com/muety/wakapi/middlewares/custom"
	"github.com/muety/wakapi/routes"
	shieldsV1Routes "github.com/muety/wakapi/routes/compat/shields/v1"
	wtV1Routes "github.com/muety/wakapi/routes/compat/wakatime/v1"
	"github.com/muety/wakapi/services"
	_ "gorm.io/driver/mysql"
	_ "gorm.io/driver/postgres"
	_ "gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	db     *gorm.DB
	config *conf.Config
)

var (
	aliasRepository           repositories.IAliasRepository
	heartbeatRepository       repositories.IHeartbeatRepository
	userRepository            repositories.IUserRepository
	languageMappingRepository repositories.ILanguageMappingRepository
	summaryRepository         repositories.ISummaryRepository
	keyValueRepository        repositories.IKeyValueRepository
)

var (
	aliasService           services.IAliasService
	heartbeatService       services.IHeartbeatService
	userService            services.IUserService
	languageMappingService services.ILanguageMappingService
	summaryService         services.ISummaryService
	aggregationService     services.IAggregationService
	keyValueService        services.IKeyValueService
	miscService            services.IMiscService
)

// TODO: Refactor entire project to be structured after business domains

func main() {
	config = conf.Load()

	// Set log level
	if config.IsDev() {
		logbuch.SetLevel(logbuch.LevelDebug)
	} else {
		logbuch.SetLevel(logbuch.LevelInfo)
	}

	// Set up GORM
	gormLogger := logger.New(
		log.New(os.Stdout, "", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Minute,
			Colorful:      false,
			LogLevel:      logger.Silent,
		},
	)

	// Connect to database
	var err error
	db, err = gorm.Open(config.Db.GetDialector(), &gorm.Config{Logger: gormLogger})
	if config.Db.Dialect == "sqlite3" {
		db.Raw("PRAGMA foreign_keys = ON;")
	}

	if config.IsDev() {
		db = db.Debug()
	}
	sqlDb, _ := db.DB()
	sqlDb.SetMaxIdleConns(int(config.Db.MaxConn))
	sqlDb.SetMaxOpenConns(int(config.Db.MaxConn))
	if err != nil {
		logbuch.Error(err.Error())
		logbuch.Fatal("could not connect to database")
	}
	defer sqlDb.Close()

	// Migrate database schema
	migrations.RunPreMigrations(db, config)
	runDatabaseMigrations()
	migrations.RunCustomPostMigrations(db, config)

	// Repositories
	aliasRepository = repositories.NewAliasRepository(db)
	heartbeatRepository = repositories.NewHeartbeatRepository(db)
	userRepository = repositories.NewUserRepository(db)
	languageMappingRepository = repositories.NewLanguageMappingRepository(db)
	summaryRepository = repositories.NewSummaryRepository(db)
	keyValueRepository = repositories.NewKeyValueRepository(db)

	// Services
	aliasService = services.NewAliasService(aliasRepository)
	userService = services.NewUserService(userRepository)
	languageMappingService = services.NewLanguageMappingService(languageMappingRepository)
	heartbeatService = services.NewHeartbeatService(heartbeatRepository, languageMappingService)
	summaryService = services.NewSummaryService(summaryRepository, heartbeatService, aliasService)
	aggregationService = services.NewAggregationService(userService, summaryService, heartbeatService)
	keyValueService = services.NewKeyValueService(keyValueRepository)
	miscService = services.NewMiscService(userService, summaryService, keyValueService)

	// Schedule background tasks
	go aggregationService.Schedule()
	go miscService.ScheduleCountTotalTime()

	// TODO: move endpoint registration to the respective routes files

	routes.Init()

	// Handlers
	summaryHandler := routes.NewSummaryHandler(summaryService)
	healthHandler := routes.NewHealthHandler(db)
	heartbeatHandler := routes.NewHeartbeatHandler(heartbeatService, languageMappingService)
	settingsHandler := routes.NewSettingsHandler(userService, summaryService, aliasService, aggregationService, languageMappingService)
	homeHandler := routes.NewHomeHandler(keyValueService)
	loginHandler := routes.NewLoginHandler(userService)
	imprintHandler := routes.NewImprintHandler(keyValueService)
	wakatimeV1AllHandler := wtV1Routes.NewAllTimeHandler(summaryService)
	wakatimeV1SummariesHandler := wtV1Routes.NewSummariesHandler(summaryService)
	shieldV1BadgeHandler := shieldsV1Routes.NewBadgeHandler(summaryService, userService)

	// Setup Routers
	router := mux.NewRouter()
	publicRouter := router.PathPrefix("/").Subrouter()
	settingsRouter := publicRouter.PathPrefix("/settings").Subrouter()
	summaryRouter := publicRouter.PathPrefix("/summary").Subrouter()
	apiRouter := router.PathPrefix("/api").Subrouter()
	summaryApiRouter := apiRouter.PathPrefix("/summary").Subrouter()
	heartbeatApiRouter := apiRouter.PathPrefix("/heartbeat").Subrouter()
	healthApiRouter := apiRouter.PathPrefix("/health").Subrouter()
	compatRouter := apiRouter.PathPrefix("/compat").Subrouter()
	wakatimeV1Router := compatRouter.PathPrefix("/wakatime/v1/users/{user}").Subrouter()
	shieldsV1Router := compatRouter.PathPrefix("/shields/v1/{user}").Subrouter()

	// Middlewares
	recoveryMiddleware := handlers.RecoveryHandler()
	loggingMiddleware := middlewares.NewLoggingMiddleware(
		// Use logbuch here once https://github.com/emvi/logbuch/issues/4 is realized
		log.New(os.Stdout, "", log.LstdFlags),
	)
	corsMiddleware := handlers.CORS()
	authenticateMiddleware := middlewares.NewAuthenticateMiddleware(
		userService,
		[]string{"/api/health", "/api/compat/shields/v1"},
	).Handler
	wakatimeRelayMiddleware := customMiddleware.NewWakatimeRelayMiddleware().Handler

	// Router configs
	router.Use(loggingMiddleware, recoveryMiddleware)
	summaryRouter.Use(authenticateMiddleware)
	settingsRouter.Use(authenticateMiddleware)
	apiRouter.Use(corsMiddleware, authenticateMiddleware)
	heartbeatApiRouter.Use(wakatimeRelayMiddleware)

	// Route registrations
	homeHandler.RegisterRoutes(publicRouter)
	loginHandler.RegisterRoutes(publicRouter)
	imprintHandler.RegisterRoutes(publicRouter)
	summaryHandler.RegisterRoutes(summaryRouter)
	settingsHandler.RegisterRoutes(settingsRouter)

	// API Route registrations
	summaryHandler.RegisterAPIRoutes(summaryApiRouter)
	healthHandler.RegisterAPIRoutes(healthApiRouter)
	heartbeatHandler.RegisterAPIRoutes(heartbeatApiRouter)
	wakatimeV1AllHandler.RegisterAPIRoutes(wakatimeV1Router)
	wakatimeV1SummariesHandler.RegisterAPIRoutes(wakatimeV1Router)
	shieldV1BadgeHandler.RegisterAPIRoutes(shieldsV1Router)

	// Static Routes
	router.PathPrefix("/assets").Handler(http.FileServer(pkger.Dir("/static")))

	// Listen HTTP
	listen(router)
}

func listen(handler http.Handler) {
	var s4, s6 *http.Server

	// IPv4
	if config.Server.ListenIpV4 != "" {
		bindString4 := config.Server.ListenIpV4 + ":" + strconv.Itoa(config.Server.Port)
		s4 = &http.Server{
			Handler:      handler,
			Addr:         bindString4,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}
	}

	// IPv6
	if config.Server.ListenIpV6 != "" {
		bindString6 := "[" + config.Server.ListenIpV6 + "]:" + strconv.Itoa(config.Server.Port)
		s6 = &http.Server{
			Handler:      handler,
			Addr:         bindString6,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}
	}

	if config.UseTLS() {
		if s4 != nil {
			logbuch.Info("--> Listening for HTTPS on %s... ✅", s4.Addr)
			go func() {
				if err := s4.ListenAndServeTLS(config.Server.TlsCertPath, config.Server.TlsKeyPath); err != nil {
					logbuch.Fatal(err.Error())
				}
			}()
		}
		if s6 != nil {
			logbuch.Info("--> Listening for HTTPS on %s... ✅", s6.Addr)
			go func() {
				if err := s6.ListenAndServeTLS(config.Server.TlsCertPath, config.Server.TlsKeyPath); err != nil {
					logbuch.Fatal(err.Error())
				}
			}()
		}
	} else {
		if s4 != nil {
			logbuch.Info("--> Listening for HTTP on %s... ✅", s4.Addr)
			go func() {
				if err := s4.ListenAndServe(); err != nil {
					logbuch.Fatal(err.Error())
				}
			}()
		}
		if s6 != nil {
			logbuch.Info("--> Listening for HTTP on %s... ✅", s6.Addr)
			go func() {
				if err := s6.ListenAndServe(); err != nil {
					logbuch.Fatal(err.Error())
				}
			}()
		}
	}

	<-make(chan interface{}, 1)
}

func runDatabaseMigrations() {
	if err := config.GetMigrationFunc(config.Db.Dialect)(db); err != nil {
		logbuch.Fatal(err.Error())
	}
}
