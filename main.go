package main

import (
	"github.com/gorilla/handlers"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/migrations/common"
	"github.com/muety/wakapi/repositories"
	"log"
	"net/http"
	"strconv"
	"time"

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

var (
	db     *gorm.DB
	config *conf.Config
)

var (
	aliasRepository           *repositories.AliasRepository
	heartbeatRepository       *repositories.HeartbeatRepository
	userRepository            *repositories.UserRepository
	languageMappingRepository *repositories.LanguageMappingRepository
	summaryRepository         *repositories.SummaryRepository
	keyValueRepository        *repositories.KeyValueRepository
)

var (
	aliasService           *services.AliasService
	heartbeatService       *services.HeartbeatService
	userService            *services.UserService
	languageMappingService *services.LanguageMappingService
	summaryService         *services.SummaryService
	aggregationService     *services.AggregationService
	keyValueService        *services.KeyValueService
)

// TODO: Refactor entire project to be structured after business domains

func main() {
	config = conf.Load()

	// Enable line numbers in logging
	if config.IsDev() {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	// Connect to database
	var err error
	db, err = gorm.Open(config.Db.GetDialector(), &gorm.Config{})
	if config.Db.Dialect == "sqlite3" {
		db.Raw("PRAGMA foreign_keys = ON;")
	}

	sqlDb, _ := db.DB()
	sqlDb.SetMaxIdleConns(int(config.Db.MaxConn))
	sqlDb.SetMaxOpenConns(int(config.Db.MaxConn))
	if err != nil {
		log.Println(err)
		log.Fatal("could not connect to database")
	}
	defer sqlDb.Close()

	// Migrate database schema
	common.RunCustomPreMigrations(db, config)
	runDatabaseMigrations()
	common.RunCustomPostMigrations(db, config)

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

	// Aggregate heartbeats to summaries and persist them
	go aggregationService.Schedule()

	// TODO: move endpoint registration to the respective routes files

	// Handlers
	summaryHandler := routes.NewSummaryHandler(summaryService)
	healthHandler := routes.NewHealthHandler(db)
	heartbeatHandler := routes.NewHeartbeatHandler(heartbeatService, languageMappingService)
	settingsHandler := routes.NewSettingsHandler(userService, languageMappingService)
	publicHandler := routes.NewIndexHandler(userService, keyValueService)
	wakatimeV1AllHandler := wtV1Routes.NewAllTimeHandler(summaryService)
	wakatimeV1SummariesHandler := wtV1Routes.NewSummariesHandler(summaryService)
	shieldV1BadgeHandler := shieldsV1Routes.NewBadgeHandler(summaryService, userService)

	// Setup Routers
	router := mux.NewRouter()
	publicRouter := router.PathPrefix("/").Subrouter()
	settingsRouter := publicRouter.PathPrefix("/settings").Subrouter()
	summaryRouter := publicRouter.PathPrefix("/summary").Subrouter()
	apiRouter := router.PathPrefix("/api").Subrouter()
	compatRouter := apiRouter.PathPrefix("/compat").Subrouter()
	wakatimeV1Router := compatRouter.PathPrefix("/wakatime/v1").Subrouter()
	shieldsV1Router := compatRouter.PathPrefix("/shields/v1").Subrouter()

	// Middlewares
	recoveryMiddleware := handlers.RecoveryHandler()
	loggingMiddleware := middlewares.NewLoggingMiddleware().Handler
	corsMiddleware := handlers.CORS()
	authenticateMiddleware := middlewares.NewAuthenticateMiddleware(
		userService,
		[]string{"/api/health", "/api/compat/shields/v1"},
	).Handler

	// Router configs
	router.Use(loggingMiddleware, recoveryMiddleware)
	summaryRouter.Use(authenticateMiddleware)
	settingsRouter.Use(authenticateMiddleware)
	apiRouter.Use(corsMiddleware, authenticateMiddleware)

	// Public Routes
	publicRouter.Path("/").Methods(http.MethodGet).HandlerFunc(publicHandler.GetIndex)
	publicRouter.Path("/login").Methods(http.MethodPost).HandlerFunc(publicHandler.PostLogin)
	publicRouter.Path("/logout").Methods(http.MethodPost).HandlerFunc(publicHandler.PostLogout)
	publicRouter.Path("/signup").Methods(http.MethodGet).HandlerFunc(publicHandler.GetSignup)
	publicRouter.Path("/signup").Methods(http.MethodPost).HandlerFunc(publicHandler.PostSignup)
	publicRouter.Path("/imprint").Methods(http.MethodGet).HandlerFunc(publicHandler.GetImprint)

	// Summary Routes
	summaryRouter.Methods(http.MethodGet).HandlerFunc(summaryHandler.GetIndex)

	// Settings Routes
	settingsRouter.Methods(http.MethodGet).HandlerFunc(settingsHandler.GetIndex)
	settingsRouter.Path("/credentials").Methods(http.MethodPost).HandlerFunc(settingsHandler.PostCredentials)
	settingsRouter.Path("/language_mappings").Methods(http.MethodPost).HandlerFunc(settingsHandler.PostCreateLanguageMapping)
	settingsRouter.Path("/language_mappings/delete").Methods(http.MethodPost).HandlerFunc(settingsHandler.DeleteLanguageMapping)
	settingsRouter.Path("/reset").Methods(http.MethodPost).HandlerFunc(settingsHandler.PostResetApiKey)
	settingsRouter.Path("/badges").Methods(http.MethodPost).HandlerFunc(settingsHandler.PostToggleBadges)

	// API Routes
	apiRouter.Path("/heartbeat").Methods(http.MethodPost).HandlerFunc(heartbeatHandler.ApiPost)
	apiRouter.Path("/summary").Methods(http.MethodGet).HandlerFunc(summaryHandler.ApiGet)
	apiRouter.Path("/health").Methods(http.MethodGet).HandlerFunc(healthHandler.ApiGet)

	// Wakatime compat V1 API Routes
	wakatimeV1Router.Path("/users/{user}/all_time_since_today").Methods(http.MethodGet).HandlerFunc(wakatimeV1AllHandler.ApiGet)
	wakatimeV1Router.Path("/users/{user}/summaries").Methods(http.MethodGet).HandlerFunc(wakatimeV1SummariesHandler.ApiGet)

	// Shields.io compat API Routes
	shieldsV1Router.PathPrefix("/{user}").Methods(http.MethodGet).HandlerFunc(shieldV1BadgeHandler.ApiGet)

	// Static Routes
	router.PathPrefix("/assets").Handler(http.FileServer(http.Dir("./static")))

	// Listen HTTP
	portString := config.Server.ListenIpV4 + ":" + strconv.Itoa(config.Server.Port)
	s := &http.Server{
		Handler:      router,
		Addr:         portString,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Printf("Listening on %+s\n", portString)
	s.ListenAndServe()
}

func runDatabaseMigrations() {
	if err := config.GetMigrationFunc(config.Db.Dialect)(db); err != nil {
		log.Fatal(err)
	}
}
