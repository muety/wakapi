package utils

import (
	"errors"
	"net/http"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
)

func ExtractCookieAuth(r *http.Request) (username *string, err error) {
	cookie, err := r.Cookie(models.AuthCookieKey)
	if err != nil {
		return nil, errors.New("missing authentication")
	}

	if err := config.GetAuthCookie().Decode(models.AuthCookieKey, cookie.Value, &username); err != nil {
		return nil, errors.New("cookie is invalid")
	}

	return username, nil
}
