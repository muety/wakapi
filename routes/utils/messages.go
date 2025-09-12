package utils

import (
	"net/http"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models/view"
)

func SetError(r *http.Request, w http.ResponseWriter, message string) {
	setMessage(r, w, message, "error")
}

func SetSuccess(r *http.Request, w http.ResponseWriter, message string) {
	setMessage(r, w, message, "success")
}

func WithSessionMessages[T view.BasicViewModel](vm T, r *http.Request, w http.ResponseWriter) T {
	session, _ := conf.GetSessionStore().Get(r, conf.CookieKeySession)
	if errors := session.Flashes("error"); len(errors) > 0 {
		vm.SetError(errors[0].(string))
	}
	if successes := session.Flashes("success"); len(successes) > 0 {
		vm.SetSuccess(successes[0].(string))
	}
	session.Save(r, w)
	return vm
}

func setMessage(r *http.Request, w http.ResponseWriter, message, key string) {
	session, _ := conf.GetSessionStore().Get(r, conf.CookieKeySession)
	session.AddFlash(message, key)
	session.Save(r, w)
}
