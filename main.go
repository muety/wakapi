package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/emvi/logbuch"
	"github.com/gorilla/handlers"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/migrations"
	"github.com/muety/wakapi/repositories"
	"github.com/muety/wakapi/routes/api"
	"github.com/muety/wakapi/services/mail"
	"github.com/muety/wakapi/utils"
	"gorm.io/gorm/logger"

	"github.com/gorilla/mux"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/routes"
	shieldsV1Routes "github.com/muety/wakapi/routes/compat/shields/v1"
	wtV1Routes "github.com/muety/wakapi/routes/compat/wakatime/v1"
	"github.com/muety/wakapi/services"
	_ "gorm.io/driver/mysql"
	_ "gorm.io/driver/postgres"
	_ "gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Embed version.txt
//go:embed version.txt
var version string

// Embed static files
//go:embed static
var staticFiles embed.FS

var (
	db     *gorm.DB
	config *conf.Config
)

var (
	aliasRepository           repositories.IAliasRepository
	heartbeatRepository       repositories.IHeartbeatRepository
	userRepository            repositories.IUserRepository
	languageMappingRepository repositories.ILanguageMappingRepository
	projectLabelRepository    repositories.IProjectLabelRepository
	summaryRepository         repositories.ISummaryRepository
	keyValueRepository        repositories.IKeyValueRepository
)

var (
	aliasService           services.IAliasService
	heartbeatService       services.IHeartbeatService
	userService            services.IUserService
	languageMappingService services.ILanguageMappingService
	projectLabelService    services.IProjectLabelService
	summaryService         services.ISummaryService
	aggregationService     services.IAggregationService
	mailService            services.IMailService
	keyValueService        services.IKeyValueService
	reportService          services.IReportService
	miscService            services.IMiscService
)

// TODO: Refactor entire project to be structured after business domains

// @title Wakapi API
// @version 1.0
// @description REST API to interact with [Wakapi](https://wakapi.dev)
// @description
// @description ## Authentication
// @description Set header `Authorization` to your API Key encoded as Base64 and prefixed with `Basic`
// @description **Example:** `Basic ODY2NDhkNzQtMTljNS00NTJiLWJhMDEtZmIzZWM3MGQ0YzJmCg==`

// @contact.name Ferdinand Mütsch
// @contact.url https://github.com/muety
// @contact.email ferdinand@muetsch.io

// @license.name GPL-3.0
// @license.url https://github.com/muety/wakapi/blob/master/LICENSE

// @securitydefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// @BasePath /api
func main() {
	config = conf.Load(version)

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
	sqlDb, err := db.DB()
	sqlDb.SetMaxIdleConns(int(config.Db.MaxConn))
	sqlDb.SetMaxOpenConns(int(config.Db.MaxConn))
	if err != nil {
		logbuch.Error(err.Error())
		logbuch.Fatal("could not connect to database")
	}
	defer sqlDb.Close()

	// Migrate database schema
	migrations.Run(db, config)

	// Repositories
	aliasRepository = repositories.NewAliasRepository(db)
	heartbeatRepository = repositories.NewHeartbeatRepository(db)
	userRepository = repositories.NewUserRepository(db)
	languageMappingRepository = repositories.NewLanguageMappingRepository(db)
	projectLabelRepository = repositories.NewProjectLabelRepository(db)
	summaryRepository = repositories.NewSummaryRepository(db)
	keyValueRepository = repositories.NewKeyValueRepository(db)

	// Services
	aliasService = services.NewAliasService(aliasRepository)
	userService = services.NewUserService(userRepository)
	languageMappingService = services.NewLanguageMappingService(languageMappingRepository)
	projectLabelService = services.NewProjectLabelService(projectLabelRepository)
	heartbeatService = services.NewHeartbeatService(heartbeatRepository, languageMappingService)
	summaryService = services.NewSummaryService(summaryRepository, heartbeatService, aliasService, projectLabelService)
	aggregationService = services.NewAggregationService(userService, summaryService, heartbeatService)
	mailService = mail.NewMailService()
	keyValueService = services.NewKeyValueService(keyValueRepository)
	reportService = services.NewReportService(summaryService, userService, mailService)
	miscService = services.NewMiscService(userService, summaryService, keyValueService)

	// Schedule background tasks
	go aggregationService.Schedule()
	go miscService.ScheduleCountTotalTime()
	go reportService.Schedule()

	routes.Init()

	// API Handlers
	healthApiHandler := api.NewHealthApiHandler(db)
	heartbeatApiHandler := api.NewHeartbeatApiHandler(userService, heartbeatService, languageMappingService)
	summaryApiHandler := api.NewSummaryApiHandler(userService, summaryService)
	metricsHandler := api.NewMetricsHandler(userService, summaryService, heartbeatService, keyValueService)

	// Compat Handlers
	wakatimeV1AllHandler := wtV1Routes.NewAllTimeHandler(userService, summaryService)
	wakatimeV1SummariesHandler := wtV1Routes.NewSummariesHandler(userService, summaryService)
	wakatimeV1StatsHandler := wtV1Routes.NewStatsHandler(userService, summaryService)
	wakatimeV1UsersHandler := wtV1Routes.NewUsersHandler(userService, heartbeatService)
	wakatimeV1ProjectsHandler := wtV1Routes.NewProjectsHandler(userService, heartbeatService)
	shieldV1BadgeHandler := shieldsV1Routes.NewBadgeHandler(summaryService, userService)

	// MVC Handlers
	summaryHandler := routes.NewSummaryHandler(summaryService, userService)
	settingsHandler := routes.NewSettingsHandler(userService, heartbeatService, summaryService, aliasService, aggregationService, languageMappingService, projectLabelService, keyValueService, mailService)
	homeHandler := routes.NewHomeHandler(keyValueService)
	loginHandler := routes.NewLoginHandler(userService, mailService)
	imprintHandler := routes.NewImprintHandler(keyValueService)

	// Setup Routers
	router := mux.NewRouter()
	rootRouter := router.PathPrefix("/").Subrouter()
	apiRouter := router.PathPrefix("/api").Subrouter().StrictSlash(true)

	// Globally used middlewares
	router.Use(middlewares.NewPrincipalMiddleware())
	router.Use(middlewares.NewLoggingMiddleware(logbuch.Info, []string{"/assets"}))
	router.Use(handlers.RecoveryHandler())
	if config.Sentry.Dsn != "" {
		router.Use(middlewares.NewSentryMiddleware())
	}
	rootRouter.Use(middlewares.NewSecurityMiddleware())

	// Route registrations
	homeHandler.RegisterRoutes(rootRouter)
	loginHandler.RegisterRoutes(rootRouter)
	imprintHandler.RegisterRoutes(rootRouter)
	summaryHandler.RegisterRoutes(rootRouter)
	settingsHandler.RegisterRoutes(rootRouter)

	// API route registrations
	summaryApiHandler.RegisterRoutes(apiRouter)
	healthApiHandler.RegisterRoutes(apiRouter)
	heartbeatApiHandler.RegisterRoutes(apiRouter)
	metricsHandler.RegisterRoutes(apiRouter)
	wakatimeV1AllHandler.RegisterRoutes(apiRouter)
	wakatimeV1SummariesHandler.RegisterRoutes(apiRouter)
	wakatimeV1StatsHandler.RegisterRoutes(apiRouter)
	wakatimeV1UsersHandler.RegisterRoutes(apiRouter)
	wakatimeV1ProjectsHandler.RegisterRoutes(apiRouter)
	shieldV1BadgeHandler.RegisterRoutes(apiRouter)

	// Static Routes
	// https://github.com/golang/go/issues/43431
	embeddedStatic, _ := fs.Sub(staticFiles, "static")
	static := conf.ChooseFS("static", embeddedStatic)
	fileServer := http.FileServer(utils.NeuteredFileSystem{Fs: http.FS(static)})
	router.PathPrefix("/contribute.json").Handler(fileServer)
	router.PathPrefix("/assets").Handler(fileServer)
	router.PathPrefix("/swagger-ui").Handler(fileServer)
	router.PathPrefix("/docs").Handler(
		middlewares.NewFileTypeFilterMiddleware([]string{".go"})(fileServer),
	)

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
