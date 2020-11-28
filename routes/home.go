package routes

import (
	"fmt"
	"github.com/gorilla/schema"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/models/view"
	"net/http"
)

type HomeHandler struct {
	config *conf.Config
}

var loginDecoder = schema.NewDecoder()
var signupDecoder = schema.NewDecoder()

func NewHomeHandler() *HomeHandler {
	return &HomeHandler{
		config: conf.Get(),
	}
}

func (h *HomeHandler) GetIndex(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	if cookie, err := r.Cookie(models.AuthCookieKey); err == nil && cookie.Value != "" {
		http.Redirect(w, r, fmt.Sprintf("%s/summary", h.config.Server.BasePath), http.StatusFound)
		return
	}

	templates[conf.IndexTemplate].Execute(w, h.buildViewModel(r))
}

func (h *HomeHandler) buildViewModel(r *http.Request) *view.HomeViewModel {
	return &view.HomeViewModel{
		Success: r.URL.Query().Get("success"),
		Error:   r.URL.Query().Get("error"),
	}
}
