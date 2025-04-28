package utils

import (
	"net/http"

	conf "github.com/muety/wakapi/config"
)

func SetError(r *http.Request, w http.ResponseWriter, message string) {
	setMessage(r, w, message, "error")
}

func SetSuccess(r *http.Request, w http.ResponseWriter, message string) {
	setMessage(r, w, message, "success")
}

func setMessage(r *http.Request, w http.ResponseWriter, message, key string) {
	session, _ := conf.GetSessionStore().Get(r, conf.SessionKeyDefault)
	session.AddFlash(message, key)
	session.Save(r, w)
}
