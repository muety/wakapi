package utils

import (
	"encoding/json"
	"github.com/muety/wakapi/config"
	"net/http"
)

func RespondJSON(w http.ResponseWriter, status int, object interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(object); err != nil {
		config.Log().Error("error while writing json response: %v", err)
	}
}
