package main

import (
	"fmt"
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
		log.Fatal(fmt.Sprintf("Fail to read file: %v", err))
	}

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
	db.AutoMigrate(&models.Heartbeat{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")

	// Services
	heartbeatSrvc := &services.HeartbeatService{config, db}
	userSrvc := &services.UserService{config, db}
	summarySrvc := &services.SummaryService{config, db, heartbeatSrvc}

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
