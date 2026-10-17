package utils

import (
	"errors"
	"net/http"
	"time"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
)

const DefaultSessionTimeout = 2 * time.Hour

type CookieKeyData struct {
	Username string `json:"username,omitempty"`
	Expiry   int64  `json:"expiry,omitempty"`
}

func GetAuthCookieDuration() time.Duration {
	config := conf.Get()
	if config.Security.CookieMaxAgeSec > 0 {
		return time.Duration(config.Security.CookieMaxAgeSec) * time.Second
	}
	return DefaultSessionTimeout
}

func CreateAuthCookieWithExpiry(username string, expiry time.Time) (*http.Cookie, error) {
	config := conf.Get()

	cookieData := CookieKeyData{
		Username: username,
		Expiry:   expiry.Unix(),
	}
	encoded, err := conf.GetAuthCookie().Encode(models.AuthCookieKey, cookieData)
	if err != nil {
		return nil, err
	}

	return config.CreateCookie(models.AuthCookieKey, encoded), nil
}

func CreateAuthCookie(username string) (*http.Cookie, error) {
	return CreateAuthCookieWithExpiry(username, time.Now().Add(GetAuthCookieDuration()))
}

func ExtractCookieAuthData(r *http.Request) (*CookieKeyData, error) {
	cookie, err := r.Cookie(models.AuthCookieKey)
	if err != nil {
		return nil, errors.New("missing authentication")
	}

	var cookieData CookieKeyData
	if err := conf.GetAuthCookie().Decode(models.AuthCookieKey, cookie.Value, &cookieData); err != nil {
		return nil, errors.New("cookie is invalid")
	}

	if cookieData.Username == "" {
		return nil, errors.New("missing username")
	}
	if time.Now().After(time.Unix(cookieData.Expiry, 0)) {
		return nil, errors.New("cookie is expired")
	}

	return &cookieData, nil
}

func ExtractCookieAuth(r *http.Request) (username *string, err error) {
	cookieData, err := ExtractCookieAuthData(r)
	if err != nil {
		return nil, err
	}

	return &cookieData.Username, nil
}

func ShouldRefreshCookie(expiry int64) bool {
	ttl := GetAuthCookieDuration()
	remaining := time.Until(time.Unix(expiry, 0))
	return remaining <= ttl/2
}

func AuthenticateCookie(w http.ResponseWriter, r *http.Request) (username *string, err error) {
	cookieData, err := ExtractCookieAuthData(r)
	if err != nil {
		return nil, err
	}

	if w != nil && ShouldRefreshCookie(cookieData.Expiry) {
		if cookie, err := CreateAuthCookie(cookieData.Username); err == nil {
			http.SetCookie(w, cookie)
		}
		_ = RefreshSession(r, w)
	}

	return &cookieData.Username, nil
}
