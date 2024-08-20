package helpers

import (
	"encoding/json"
	"errors"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"net/http"
)

func ExtractCookieAuth(r *http.Request, config *config.Config) (username *string, err error) {
	cookie, err := r.Cookie(models.AuthCookieKey)
	if err != nil {
		return nil, errors.New("missing authentication")
	}

	if err := config.Security.SecureCookie.Decode(models.AuthCookieKey, cookie.Value, &username); err != nil {
		return nil, errors.New("cookie is invalid")
	}

	return username, nil
}

func RespondJSON(w http.ResponseWriter, r *http.Request, status int, object interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(object); err != nil {
		config.Log().Request(r).Error("error while writing json response", "error", err)
	}
}
