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

	port := cfg.Section("server").Key("port").MustInt()
	intervalStr := cfg.Section("data").Key("aggregation_interval").String()
	interval, _ := time.ParseDuration(intervalStr)

	return &models.Config{
		Port:                port,
		DbHost:              dbHost,
		DbUser:              dbUser,
		DbPassword:          dbPassword,
		DbName:              dbName,
		DbDialect:           "mysql",
		AggregationInterval: interval,
	}
}

func main() {
	// Read Config
	config := readConfig()

	// Connect to database
	db, err := gorm.Open(config.DbDialect, utils.MakeConnectionString(config))
	db.LogMode(true)
	if err != nil {
		// log.Fatal("Could not connect to database.")
		log.Fatal(err)
	}
	defer db.Close()

	// Migrate database schema
	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.Heartbeat{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
	db.AutoMigrate(&models.Aggregation{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
	db.AutoMigrate(&models.AggregationItem{}).AddForeignKey("aggregation_id", "aggregations(id)", "RESTRICT", "RESTRICT")

	// Services
	heartbeatSrvc := &services.HeartbeatService{config, db}
	userSrvc := &services.UserService{config, db}
	aggregationSrvc := &services.AggregationService{config, db, heartbeatSrvc}

	// Handlers
	heartbeatHandler := &routes.HeartbeatHandler{HeartbeatSrvc: heartbeatSrvc}
	aggregationHandler := &routes.AggregationHandler{AggregationSrvc: aggregationSrvc}

	// Middlewares
	authenticate := &middlewares.AuthenticateMiddleware{UserSrvc: userSrvc}

	// Setup Routing
	router := mux.NewRouter()
	apiRouter := mux.NewRouter().PathPrefix("/api").Subrouter()

	// API Routes
	heartbeats := apiRouter.Path("/heartbeat").Subrouter()
	heartbeats.Methods("POST").HandlerFunc(heartbeatHandler.Post)

	aggreagations := apiRouter.Path("/aggregation").Subrouter()
	aggreagations.Methods("GET").HandlerFunc(aggregationHandler.Get)

	// Sub-Routes Setup
	router.PathPrefix("/api").Handler(negroni.Classic().With(
		negroni.HandlerFunc(authenticate.Handle),
		negroni.Wrap(apiRouter),
	))

	// Listen HTTP
	portString := ":" + strconv.Itoa(config.Port)
	s := &http.Server{
		Handler:      router,
		Addr:         portString,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Printf("Listening on %+s\n", portString)
	s.ListenAndServe()
}
