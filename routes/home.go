package routes

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/schema"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/models/view"
	routeutils "github.com/muety/wakapi/routes/utils"
	"github.com/muety/wakapi/services"
	"net/http"
	"strconv"
	"strings"
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

func (h *HomeHandler) RegisterRoutes(router chi.Router) {
	router.Get("/", h.GetIndex)
}

func (h *HomeHandler) GetIndex(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	if cookie, err := r.Cookie(models.AuthCookieKey); err == nil && cookie.Value != "" {
		http.Redirect(w, r, fmt.Sprintf("%s/summary", h.config.Server.BasePath), http.StatusFound)
		return
	}
	if h.config.Security.DisableFrontpage {
		http.Redirect(w, r, fmt.Sprintf("%s/login", h.config.Server.BasePath), http.StatusFound)
		return
	}

	templates[conf.IndexTemplate].Execute(w, h.buildViewModel(r, w))
}

func (h *HomeHandler) buildViewModel(r *http.Request, w http.ResponseWriter) *view.HomeViewModel {
	var totalHours int
	var totalUsers int
	var newsbox view.Newsbox

	if kv, err := h.keyValueSrvc.GetString(conf.KeyLatestTotalTime); err == nil && kv != nil && kv.Value != "" {
		if d, err := time.ParseDuration(kv.Value); err == nil {
			totalHours = int(d.Hours())
		}
	}

	if kv, err := h.keyValueSrvc.GetString(conf.KeyLatestTotalUsers); err == nil && kv != nil && kv.Value != "" {
		if d, err := strconv.Atoi(kv.Value); err == nil {
			totalUsers = d
		}
	}

	if kv, err := h.keyValueSrvc.GetString(conf.KeyNewsbox); err == nil && kv != nil && kv.Value != "" {
		if err := json.NewDecoder(strings.NewReader(kv.Value)).Decode(&newsbox); err != nil {
			conf.Log().Request(r).Error("failed to decode newsbox message - %v", err)
		}
	}

	vm := &view.HomeViewModel{
		TotalHours: totalHours,
		TotalUsers: totalUsers,
		Newsbox:    &newsbox,
	}
	return routeutils.WithSessionMessages(vm, r, w)
}
