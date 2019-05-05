package main

import (
	"encoding/json"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/n1try/wakapi/models"
)

func Authenticate(w http.ResponseWriter, r *http.Request) {

}

func HeartbeatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(415)
		return
	}
	dec := json.NewDecoder(r.Body)
	var heartbeats []models.Heartbeat
	err := dec.Decode(&heartbeats)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	err = HeartbeatSrvc.InsertMulti(heartbeats)
	if err != nil {
		w.WriteHeader(500)
		os.Stderr.WriteString(err.Error())
		return
	}

	w.WriteHeader(201)
}
