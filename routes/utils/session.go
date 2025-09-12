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
