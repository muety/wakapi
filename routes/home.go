package routes

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/models/view"
	"github.com/muety/wakapi/services"
	"net/http"
	"strconv"
	"time"
)

type HomeHandler struct {
	config       *conf.Config
	keyValueSrvc services.IKeyValueService
}

var loginDecoder = schema.NewDecoder()
var signupDecoder = schema.NewDecoder()
var resetPasswordDecoder = schema.NewDecoder()

func NewHomeHandler(keyValueService services.IKeyValueService) *HomeHandler {
	return &HomeHandler{
		config:       conf.Get(),
		keyValueSrvc: keyValueService,
	}
}

func (h *HomeHandler) RegisterRoutes(router *mux.Router) {
	router.Path("/").Methods(http.MethodGet).HandlerFunc(h.GetIndex)
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
	var totalHours int
	var totalUsers int

	if t, err := h.keyValueSrvc.GetString(conf.KeyLatestTotalTime); err == nil && t != nil && t.Value != "" {
		if d, err := time.ParseDuration(t.Value); err == nil {
			totalHours = int(d.Hours())
		}
	}

	if t, err := h.keyValueSrvc.GetString(conf.KeyLatestTotalUsers); err == nil && t != nil && t.Value != "" {
		if d, err := strconv.Atoi(t.Value); err == nil {
			totalUsers = d
		}
	}

	return &view.HomeViewModel{
		Success:    r.URL.Query().Get("success"),
		Error:      r.URL.Query().Get("error"),
		TotalHours: totalHours,
		TotalUsers: totalUsers,
	}
}
