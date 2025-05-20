package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/mileusna/useragent"
	"net/http"
	"strings"
)

type ApiRootHandler struct {
}

func NewApiRootHandler() *ApiRootHandler {
	return &ApiRootHandler{}
}

func (h *ApiRootHandler) RegisterRoutes(router chi.Router) {
	router.Get("/", h.Get)
}

func (h *ApiRootHandler) Get(w http.ResponseWriter, r *http.Request) {
	ua := useragent.Parse(r.UserAgent())
	if (ua.Desktop || ua.Tablet || ua.Mobile) && !strings.Contains(ua.String, "wakatime") {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	http.NotFound(w, r)
}
