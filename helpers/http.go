package helpers

import (
	"encoding/json"
	"net/http"

	"github.com/muety/wakapi/config"
)

func RespondJSON(w http.ResponseWriter, r *http.Request, status int, object interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(object); err != nil {
		config.Log().Request(r).Error("error while writing json response", "error", err)
	}
}
