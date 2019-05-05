package routes

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/n1try/wakapi/services"

	_ "github.com/go-sql-driver/mysql"
	"github.com/n1try/wakapi/models"
)

type HeartbeatHandler struct {
	HeartbeatSrvc *services.HeartbeatService
}

func (h *HeartbeatHandler) Post(w http.ResponseWriter, r *http.Request) {
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

	user := r.Context().Value(models.UserKey).(*models.User)
	err = h.HeartbeatSrvc.InsertBatch(heartbeats, user)
	if err != nil {
		w.WriteHeader(500)
		os.Stderr.WriteString(err.Error())
		return
	}

	w.WriteHeader(201)
}
