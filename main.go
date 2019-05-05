package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"

	_ "github.com/go-sql-driver/mysql"
	"github.com/n1try/wakapi/middlewares"
	"github.com/n1try/wakapi/models"
	"github.com/n1try/wakapi/routes"
	"github.com/n1try/wakapi/services"
)

func getConfig() models.Config {
	portPtr := flag.Int("port", 8080, "Port for the webserver to listen on")
	flag.Parse()
	return models.Config{
		Port: *portPtr,
	}
}

func main() {
	// Read Config
	config := getConfig()

	// Connect Database
	db, _ := sql.Open("mysql", "fakatime_user:eB2zyLt2heqWj5Y9@tcp(muetsch.io:3306)/fakatime")
	defer db.Close()
	err := db.Ping()
	if err != nil {
		fmt.Println("Could not connect to database.")
		os.Exit(1)
	}

	// Services
	heartbeatSrvc := &services.HeartbeatService{db}
	userSrvc := &services.UserService{db}

	// Handlers
	heartbeatHandler := &routes.HeartbeatHandler{HeartbeatSrvc: heartbeatSrvc}

	// Middlewares
	authenticate := &middlewares.AuthenticateMiddleware{UserSrvc: userSrvc}

	// Setup Routing
	router := mux.NewRouter()
	apiRouter := mux.NewRouter().PathPrefix("/api").Subrouter()
	n := negroni.Classic()
	n.UseHandler(router)

	// API Routes
	heartbeats := apiRouter.Path("/heartbeat").Subrouter()
	heartbeats.Methods("POST").HandlerFunc(heartbeatHandler.Post)

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
	fmt.Printf("Listening on %+s\n", portString)
	s.ListenAndServe()
}
