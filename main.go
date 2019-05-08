package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"github.com/go-sql-driver/mysql"
	"github.com/n1try/wakapi/middlewares"
	"github.com/n1try/wakapi/models"
	"github.com/n1try/wakapi/routes"
	"github.com/n1try/wakapi/services"
)

func readConfig() models.Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	port, err := strconv.Atoi(os.Getenv("WAKAPI_PORT"))
	dbUser, valid := os.LookupEnv("WAKAPI_DB_USER")
	dbPassword, valid := os.LookupEnv("WAKAPI_DB_PASSWORD")
	dbHost, valid := os.LookupEnv("WAKAPI_DB_HOST")
	dbName, valid := os.LookupEnv("WAKAPI_DB_NAME")

	if err != nil {
		log.Fatal(err)
	}
	if !valid {
		log.Fatal("Config parameters missing.")
	}

	return models.Config{
		Port:       port,
		DbHost:     dbHost,
		DbUser:     dbUser,
		DbPassword: dbPassword,
		DbName:     dbName,
	}
}

func main() {
	// Read Config
	config := readConfig()

	// Connect Database
	dbConfig := mysql.Config{
		User:                 config.DbUser,
		Passwd:               config.DbPassword,
		Net:                  "tcp",
		Addr:                 config.DbHost,
		DBName:               config.DbName,
		AllowNativePasswords: true,
		ParseTime:            true,
	}
	db, _ := sql.Open("mysql", dbConfig.FormatDSN())
	defer db.Close()
	err := db.Ping()
	if err != nil {
		log.Fatal("Could not connect to database.")
	}

	// Services
	heartbeatSrvc := &services.HeartbeatService{db}
	userSrvc := &services.UserService{db}
	aggregationSrvc := &services.AggregationService{db, heartbeatSrvc}

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
