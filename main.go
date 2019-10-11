package main

import (
	"crypto/md5"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	uuid "github.com/satori/go.uuid"
	ini "gopkg.in/ini.v1"

	"github.com/n1try/wakapi/middlewares"
	"github.com/n1try/wakapi/models"
	"github.com/n1try/wakapi/routes"
	"github.com/n1try/wakapi/services"
	"github.com/n1try/wakapi/utils"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func readConfig() *models.Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	dbUser, valid := os.LookupEnv("WAKAPI_DB_USER")
	dbPassword, valid := os.LookupEnv("WAKAPI_DB_PASSWORD")
	dbHost, valid := os.LookupEnv("WAKAPI_DB_HOST")
	dbName, valid := os.LookupEnv("WAKAPI_DB_NAME")

	if !valid {
		log.Fatal("Environment variables missing.")
	}

	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Fatalf("Fail to read file: %v", err)
	}

	dbMaxConn := cfg.Section("database").Key("max_connections").MustUint(1)
	addr := cfg.Section("server").Key("listen").MustString("127.0.0.1")
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		port = cfg.Section("server").Key("port").MustInt()
	}

	// Read custom languages
	customLangs := make(map[string]string)
	languageKeys := cfg.Section("languages").Keys()
	for _, k := range languageKeys {
		customLangs[k.Name()] = k.MustString("unknown")
	}

	return &models.Config{
		Port:            port,
		Addr:            addr,
		DbHost:          dbHost,
		DbUser:          dbUser,
		DbPassword:      dbPassword,
		DbName:          dbName,
		DbDialect:       "mysql",
		DbMaxConn:       dbMaxConn,
		CustomLanguages: customLangs,
	}
}

func main() {
	// Read Config
	config := readConfig()

	// Connect to database
	db, err := gorm.Open(config.DbDialect, utils.MakeConnectionString(config))
	db.LogMode(false)
	db.DB().SetMaxIdleConns(1)
	db.DB().SetMaxOpenConns(10)
	if err != nil {
		// log.Fatal("Could not connect to database.")
		log.Fatal(err)
	}
	defer db.Close()

	// Migrate database schema
	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.Alias{})
	db.AutoMigrate(&models.Summary{})
	db.AutoMigrate(&models.SummaryItem{})
	db.AutoMigrate(&models.Heartbeat{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
	db.AutoMigrate(&models.SummaryItem{}).AddForeignKey("summary_id", "summaries(id)", "CASCADE", "CASCADE")

	// Custom migrations and initial data
	addDefaultUser(db, config)
	migrateLanguages(db, config)

	// Services
	aliasSrvc := &services.AliasService{config, db}
	heartbeatSrvc := &services.HeartbeatService{config, db}
	userSrvc := &services.UserService{config, db}
	summarySrvc := &services.SummaryService{config, db, heartbeatSrvc, aliasSrvc}
	aggregationSrvc := &services.AggregationService{config, db, userSrvc, summarySrvc, heartbeatSrvc}

	// Aggregate heartbeats to summaries and persist them
	go aggregationSrvc.Schedule()

	// Handlers
	heartbeatHandler := &routes.HeartbeatHandler{HeartbeatSrvc: heartbeatSrvc}
	summaryHandler := &routes.SummaryHandler{SummarySrvc: summarySrvc}

	// Middlewares
	authenticate := &middlewares.AuthenticateMiddleware{UserSrvc: userSrvc}
	cors := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"*"},
		Debug:          false,
	})

	// Setup Routing
	router := mux.NewRouter()
	apiRouter := mux.NewRouter().PathPrefix("/api").Subrouter()

	// API Routes
	heartbeats := apiRouter.Path("/heartbeat").Subrouter()
	heartbeats.Methods(http.MethodPost).HandlerFunc(heartbeatHandler.Post)

	aggreagations := apiRouter.Path("/summary").Subrouter()
	aggreagations.Methods(http.MethodGet).HandlerFunc(summaryHandler.Get)

	// Sub-Routes Setup
	router.PathPrefix("/api").Handler(negroni.Classic().
		With(cors).
		With(
			negroni.HandlerFunc(authenticate.Handle),
			negroni.Wrap(apiRouter),
		))
	router.PathPrefix("/").Handler(negroni.Classic().With(negroni.Wrap(http.FileServer(http.Dir("./static")))))

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

func migrateLanguages(db *gorm.DB, cfg *models.Config) {
	for k, v := range cfg.CustomLanguages {
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

func addDefaultUser(db *gorm.DB, cfg *models.Config) {
	pw := md5.Sum([]byte(models.DefaultPassword))
	pwString := hex.EncodeToString(pw[:])
	var err error
	apiKey := uuid.Must(uuid.NewV4(), err).String()
	if err != nil {
		log.Println("Unable to create api key.")
		log.Fatal(err)
	}
	u := &models.User{ID: models.DefaultUser, Password: pwString, ApiKey: apiKey}
	result := db.FirstOrCreate(u, &models.User{ID: u.ID})
	if result.Error != nil {
		log.Println("Unable to create default user.")
		log.Fatal(result.Error)
	}
	if result.RowsAffected > 0 {
		log.Printf("Created default user '%s' with password '%s' and API key '%s'.\n", u.ID, models.DefaultPassword, u.ApiKey)
	}
}
