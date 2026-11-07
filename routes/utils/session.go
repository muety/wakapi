package utils

import (
	"net/http"

	"github.com/muety/wakapi/config"
)

func ClearSession(r *http.Request, w http.ResponseWriter) {
	session, _ := config.GetSessionStore().Get(r, config.CookieKeySession)
	clear(session.Values)
	session.Save(r, w)
}

func RefreshSession(r *http.Request, w http.ResponseWriter) error {
	if w == nil || r == nil {
		return nil
	}
	if _, err := r.Cookie(config.CookieKeySession); err == nil {
		session, err := config.GetSessionStore().Get(r, config.CookieKeySession)
		if err != nil {
			return err
		}
		return session.Save(r, w)
	}
	return nil
}
