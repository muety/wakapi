package helpers

import (
	"encoding/json"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/utils"
	"net/http"
)

func ExtractCookieAuth(r *http.Request) (username *string, err error) {
	return utils.ExtractCookieAuth(r, config.Get().Security.SecureCookie)
}

func RespondJSON(w http.ResponseWriter, r *http.Request, status int, object interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(object); err != nil {
		config.Log().Request(r).Error("error while writing json response: %v", err)
	}
}
