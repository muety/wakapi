package main

import (
	"encoding/json"
	"github.com/gorilla/handlers"
	"github.com/gorilla/securecookie"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gobuffalo/packr/v2"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
	"github.com/rubenv/sql-migrate"
	ini "gopkg.in/ini.v1"

	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/routes"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
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
)

// TODO: Refactor entire project to be structured after business domains

func readConfig() *models.Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	env := utils.LookupFatal("ENV")
	dbType := utils.LookupFatal("WAKAPI_DB_TYPE")
	dbUser := utils.LookupFatal("WAKAPI_DB_USER")
	dbPassword := utils.LookupFatal("WAKAPI_DB_PASSWORD")
	dbHost := utils.LookupFatal("WAKAPI_DB_HOST")
	dbName := utils.LookupFatal("WAKAPI_DB_NAME")
	dbPortStr := utils.LookupFatal("WAKAPI_DB_PORT")
	defaultUserName := utils.LookupFatal("WAKAPI_DEFAULT_USER_NAME")
	defaultUserPassword := utils.LookupFatal("WAKAPI_DEFAULT_USER_PASSWORD")
	dbPort, err := strconv.Atoi(dbPortStr)

	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Fatalf("Fail to read file: %v", err)
	}

	if dbType == "" {
		dbType = "mysql"
	}

	dbMaxConn := cfg.Section("database").Key("max_connections").MustUint(1)
	addr := cfg.Section("server").Key("listen").MustString("127.0.0.1")
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		port = cfg.Section("server").Key("port").MustInt()
	}

	basePath := cfg.Section("server").Key("base_path").MustString("/")
	if strings.HasSuffix(basePath, "/") {
		basePath = basePath[:len(basePath)-1]
	}

	cleanUp := cfg.Section("app").Key("cleanup").MustBool(false)

	// Read custom languages
	customLangs := make(map[string]string)
	languageKeys := cfg.Section("languages").Keys()
	for _, k := range languageKeys {
		customLangs[k.Name()] = k.MustString("unknown")
	}

	// Read language colors
	// Source: https://raw.githubusercontent.com/ozh/github-colors/master/colors.json
	var colors = make(map[string]string)
	var rawColors map[string]struct {
		Color string `json:"color"`
		Url   string `json:"url"`
	}

	data, err := ioutil.ReadFile("data/colors.json")
	if err != nil {
		log.Fatal(err)
	}

	if err := json.Unmarshal(data, &rawColors); err != nil {
		log.Fatal(err)
	}

	for k, v := range rawColors {
		colors[strings.ToLower(k)] = v.Color
	}

	// TODO: Read keys from env, so that users are not logged out every time the server is restarted
	secureCookie := securecookie.New(
		securecookie.GenerateRandomKey(64),
		securecookie.GenerateRandomKey(32),
	)

	return &models.Config{
		Env:                 env,
		Port:                port,
		Addr:                addr,
		BasePath:            basePath,
		DbHost:              dbHost,
		DbPort:              uint(dbPort),
		DbUser:              dbUser,
		DbPassword:          dbPassword,
		DbName:              dbName,
		DbDialect:           dbType,
		DbMaxConn:           dbMaxConn,
		CleanUp:             cleanUp,
		SecureCookie:        secureCookie,
		DefaultUserName:     defaultUserName,
		DefaultUserPassword: defaultUserPassword,
		CustomLanguages:     customLangs,
		LanguageColors:      colors,
	}
}

func main() {
	// Read Config
	config = readConfig()

	// Enable line numbers in logging
	if config.IsDev() {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
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
		log.Fatal("Could not connect to database.")
	}
	// TODO: Graceful shutdown
	defer db.Close()

	// Migrate database schema
	migrateDo := databaseMigrateActions(config.DbDialect)
	migrateDo(db)

	// Services
	aliasService = &services.AliasService{Config: config, Db: db}
	heartbeatService = &services.HeartbeatService{Config: config, Db: db}
	userService = &services.UserService{Config: config, Db: db}
	summaryService = &services.SummaryService{Config: config, Db: db, HeartbeatService: heartbeatService, AliasService: aliasService}
	aggregationService = &services.AggregationService{Config: config, Db: db, UserService: userService, SummaryService: summaryService, HeartbeatService: heartbeatService}

	svcs := []services.Initializable{aliasService, heartbeatService, userService, summaryService, aggregationService}
	for _, s := range svcs {
		s.Init()
	}

	// Custom migrations and initial data
	addDefaultUser()
	migrateLanguages()

	// Aggregate heartbeats to summaries and persist them
	go aggregationService.Schedule()

	if config.CleanUp {
		go heartbeatService.ScheduleCleanUp()
	}

	// Handlers
	heartbeatHandler := routes.NewHeartbeatHandler(config, heartbeatService)
	summaryHandler := routes.NewSummaryHandler(config, summaryService)
	healthHandler := routes.NewHealthHandler(db)
	indexHandler := routes.NewIndexHandler(config, userService)

	// Setup Routers
	router := mux.NewRouter()
	indexRouter := router.PathPrefix("/").Subrouter()
	summaryRouter := indexRouter.PathPrefix("/summary").Subrouter()
	apiRouter := router.PathPrefix("/api").Subrouter()

	// Middlewares
	recoveryMiddleware := handlers.RecoveryHandler()
	loggingMiddleware := middlewares.NewLoggingMiddleware().Handler
	corsMiddleware := handlers.CORS()
	authenticateMiddleware := middlewares.NewAuthenticateMiddleware(
		config,
		userService,
		[]string{"/api/health"},
	).Handler

	// Router configs
	router.Use(loggingMiddleware, recoveryMiddleware)
	summaryRouter.Use(authenticateMiddleware)
	apiRouter.Use(corsMiddleware, authenticateMiddleware)

	// Public Routes
	indexRouter.Path("/").Methods(http.MethodGet).HandlerFunc(indexHandler.Index)
	indexRouter.Path("/login").Methods(http.MethodPost).HandlerFunc(indexHandler.Login)
	indexRouter.Path("/logout").Methods(http.MethodPost).HandlerFunc(indexHandler.Logout)
	indexRouter.Path("/signup").Methods(http.MethodGet, http.MethodPost).HandlerFunc(indexHandler.Signup)

	// Summary Routes
	summaryRouter.Methods(http.MethodGet).HandlerFunc(summaryHandler.Index)

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

func databaseMigrateActions(dbDialect string) func(db *gorm.DB) {
	var migrateDo func(db *gorm.DB)
	if dbDialect == "sqlite3" {
		migrations := &migrate.PackrMigrationSource{
			Box: packr.New("migrations", "./migrations/sqlite3"),
		}
		migrateDo = func(db *gorm.DB) {
			n, err := migrate.Exec(db.DB(), "sqlite3", migrations, migrate.Up)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Applied %d migrations!\n", n)
		}
	} else {
		migrateDo = func(db *gorm.DB) {
			db.AutoMigrate(&models.Alias{})
			db.AutoMigrate(&models.Summary{})
			db.AutoMigrate(&models.SummaryItem{})
			db.AutoMigrate(&models.User{})
			db.AutoMigrate(&models.Heartbeat{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
			db.AutoMigrate(&models.SummaryItem{}).AddForeignKey("summary_id", "summaries(id)", "CASCADE", "CASCADE")
		}
	}
	return migrateDo
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

func addDefaultUser() {
	u, created, err := userService.CreateOrGet(&models.Signup{
		Username: config.DefaultUserName,
		Password: config.DefaultUserPassword,
	})

	if err != nil {
		log.Println("unable to create default user")
		log.Fatal(err)
	} else if created {
		log.Printf("created default user '%s' with password '%s' and API key '%s'\n", u.ID, config.DefaultUserPassword, u.ApiKey)
	} else {
		log.Printf("default user '%s' already existing\n", u.ID)
	}
}
