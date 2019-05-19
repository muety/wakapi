package routes

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/n1try/wakapi/services"
	"github.com/n1try/wakapi/utils"

	"github.com/n1try/wakapi/models"
)

type HeartbeatHandler struct {
	HeartbeatSrvc *services.HeartbeatService
}

func (h *HeartbeatHandler) Post(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var heartbeats []*models.Heartbeat
	user := r.Context().Value(models.UserKey).(*models.User)
	opSys, editor, _ := utils.ParseUserAgent(r.Header.Get("User-Agent"))

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&heartbeats); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	for _, h := range heartbeats {
		h.OperatingSystem = opSys
		h.Editor = editor
		h.User = user
		h.UserID = user.ID

		if !h.Valid() {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid heartbeat object."))
			return
		}
	}

	if err := h.HeartbeatSrvc.InsertBatch(heartbeats); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		os.Stderr.WriteString(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}
