package main

import (
	"github.com/gorilla/handlers"
	conf "github.com/muety/wakapi/config"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/routes"
	shieldsV1Routes "github.com/muety/wakapi/routes/compat/shields/v1"
	wtV1Routes "github.com/muety/wakapi/routes/compat/wakatime/v1"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
)

var (
	db     *gorm.DB
	config *conf.Config
)

var (
	aliasService       *services.AliasService
	heartbeatService   *services.HeartbeatService
	userService        *services.UserService
	summaryService     *services.SummaryService
	aggregationService *services.AggregationService
	keyValueService    *services.KeyValueService
)

// TODO: Refactor entire project to be structured after business domains

func main() {
	config = conf.Load()

	// Enable line numbers in logging
	if config.IsDev() {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	// Show data loss warning
	if config.App.CleanUp {
		promptAbort("`CLEANUP` is set to `true`, which may cause data loss. Are you sure to continue?", 5)
	}

	// Connect to database
	var err error
	db, err = gorm.Open(config.Db.Dialect, utils.MakeConnectionString(config))
	if config.Db.Dialect == "sqlite3" {
		db.DB().Exec("PRAGMA foreign_keys = ON;")
	}
	db.LogMode(config.IsDev())
	db.DB().SetMaxIdleConns(int(config.Db.MaxConn))
	db.DB().SetMaxOpenConns(int(config.Db.MaxConn))
	if err != nil {
		log.Println(err)
		log.Fatal("could not connect to database")
	}
	defer db.Close()

	// Migrate database schema
	runDatabaseMigrations()
	applyFixtures()

	// Services
	aliasService = services.NewAliasService(db)
	heartbeatService = services.NewHeartbeatService(db)
	userService = services.NewUserService(db)
	summaryService = services.NewSummaryService(db, heartbeatService, aliasService)
	aggregationService = services.NewAggregationService(db, userService, summaryService, heartbeatService)
	keyValueService = services.NewKeyValueService(db)

	// Custom migrations and initial data
	migrateLanguages()

	// Aggregate heartbeats to summaries and persist them
	go aggregationService.Schedule()

	if config.App.CleanUp {
		go heartbeatService.ScheduleCleanUp()
	}

	// TODO: move endpoint registration to the respective routes files

	// Handlers
	heartbeatHandler := routes.NewHeartbeatHandler(heartbeatService)
	summaryHandler := routes.NewSummaryHandler(summaryService)
	healthHandler := routes.NewHealthHandler(db)
	settingsHandler := routes.NewSettingsHandler(userService)
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

func applyFixtures() {
	if err := config.GetFixturesFunc(config.Db.Dialect)(db); err != nil {
		log.Fatal(err)
	}
}

func migrateLanguages() {
	for k, v := range config.App.CustomLanguages {
		result := db.Model(models.Heartbeat{}).
			Where("language = ?", "").
			Where("entity LIKE ?", "%."+k).
			Updates(models.Heartbeat{Language: v})
		if result.Error != nil {
			log.Fatal(result.Error)
		}
		if result.RowsAffected > 0 {
			log.Printf("Migrated %+v rows for custom language %+s.\n", result.RowsAffected, k)
		}
	}
}

func promptAbort(message string, timeoutSec int) {
	log.Printf("[WARNING] %s.\nTo abort server startup, press Ctrl+C.\n", message)
	for i := timeoutSec; i > 0; i-- {
		log.Printf("Starting in %d seconds ...\n", i)
		time.Sleep(1 * time.Second)
	}
}
