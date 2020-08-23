package main

import (
	"github.com/gorilla/handlers"
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
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
)

var (
	db     *gorm.DB
	config *models.Config
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
	config = models.GetConfig()

	// Enable line numbers in logging
	if config.IsDev() {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	// Show data loss warning
	if config.CleanUp {
		promptAbort("`CLEANUP` is set to `true`, which may cause data loss. Are you sure to continue?", 5)
	}

	// Connect to database
	var err error
	db, err = gorm.Open(config.DbDialect, utils.MakeConnectionString(config))
	if config.DbDialect == "sqlite3" {
		db.DB().Exec("PRAGMA foreign_keys = ON;")
	}
	db.LogMode(config.IsDev())
	db.DB().SetMaxIdleConns(int(config.DbMaxConn))
	db.DB().SetMaxOpenConns(int(config.DbMaxConn))
	if err != nil {
		log.Println(err)
		log.Fatal("could not connect to database")
	}
	// TODO: Graceful shutdown
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

	if config.CleanUp {
		go heartbeatService.ScheduleCleanUp()
	}

	// Handlers
	heartbeatHandler := routes.NewHeartbeatHandler(heartbeatService)
	summaryHandler := routes.NewSummaryHandler(summaryService)
	healthHandler := routes.NewHealthHandler(db)
	settingsHandler := routes.NewSettingsHandler(userService)
	publicHandler := routes.NewIndexHandler(userService, keyValueService)

	// Setup Routers
	router := mux.NewRouter()
	publicRouter := router.PathPrefix("/").Subrouter()
	settingsRouter := publicRouter.PathPrefix("/settings").Subrouter()
	summaryRouter := publicRouter.PathPrefix("/summary").Subrouter()
	apiRouter := router.PathPrefix("/api").Subrouter()

	// Middlewares
	recoveryMiddleware := handlers.RecoveryHandler()
	loggingMiddleware := middlewares.NewLoggingMiddleware().Handler
	corsMiddleware := handlers.CORS()
	authenticateMiddleware := middlewares.NewAuthenticateMiddleware(
		userService,
		[]string{"/api/health"},
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

	// API Routes
	apiRouter.Path("/heartbeat").Methods(http.MethodPost).HandlerFunc(heartbeatHandler.ApiPost)
	apiRouter.Path("/summary").Methods(http.MethodGet).HandlerFunc(summaryHandler.ApiGet)
	apiRouter.Path("/health").Methods(http.MethodGet).HandlerFunc(healthHandler.ApiGet)

	// Static Routes
	router.PathPrefix("/assets").Handler(http.FileServer(http.Dir("./static")))

	// Listen HTTP
	portString := config.Addr + ":" + strconv.Itoa(config.Port)
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
	if err := config.GetMigrationFunc(config.DbDialect)(db); err != nil {
		log.Fatal(err)
	}
}

func applyFixtures() {
	if err := config.GetFixturesFunc(config.DbDialect)(db); err != nil {
		log.Fatal(err)
	}
}

func migrateLanguages() {
	for k, v := range config.CustomLanguages {
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
