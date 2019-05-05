package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/n1try/wakapi/models"
	"github.com/n1try/wakapi/services"
)

var HeartbeatSrvc services.HeartbeatService

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

	// Init Services
	HeartbeatSrvc = services.HeartbeatService{db}

	// Define Routes
	http.HandleFunc("/api/heartbeat", HeartbeatHandler)

	// Listen HTTP
	portString := ":" + strconv.Itoa(config.Port)
	s := &http.Server{
		Addr:         portString,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Printf("Listening on %+s\n", portString)
	s.ListenAndServe()
}
