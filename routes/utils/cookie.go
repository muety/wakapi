package utils

import (
	"errors"
	"net/http"
	"time"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
)

type cookieKeyData struct {
	Username string `json:"username,omitempty"`
	Expiry   int64  `json:"expiry,omitempty"`
}

func CreateAuthCookie(username string) (*http.Cookie, error) {
	config := conf.Get()

	var expiry time.Time
	if config.Security.CookieMaxAgeSec > 0 {
		expiry = time.Now().Add(time.Duration(config.Security.CookieMaxAgeSec) * time.Second) // expiration time is set, encode it in the cookie, so they cannot be used after expiry
	} else {
		expiry = time.Now().Add(2 * time.Hour) // cookies only last for the session, so we set the expiry to 2h, which should last long enough for most sessions
	}

	cookieData := cookieKeyData{
		Username: username,
		Expiry:   expiry.Unix(),
	}
	encoded, err := conf.GetAuthCookie().Encode(models.AuthCookieKey, cookieData)
	if err != nil {
		return nil, err
	}

	return config.CreateCookie(models.AuthCookieKey, encoded), nil
}

func ExtractCookieAuth(r *http.Request) (username *string, err error) {
	cookie, err := r.Cookie(models.AuthCookieKey)
	if err != nil {
		return nil, errors.New("missing authentication")
	}

	var cookieData cookieKeyData
	if err := conf.GetAuthCookie().Decode(models.AuthCookieKey, cookie.Value, &cookieData); err != nil {
		return nil, errors.New("cookie is invalid")
	}

	if cookieData.Username == "" {
		return nil, errors.New("missing username")
	}
	if time.Now().After(time.Unix(cookieData.Expiry, 0)) {
		return nil, errors.New("cookie is expired")
	}

	return &cookieData.Username, nil
}
