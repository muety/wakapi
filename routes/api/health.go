package api

import (
	"fmt"
	"github.com/muety/wakapi/helpers"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type HealthApiHandler struct {
	db *gorm.DB
}

type HealthResponse struct {
	App int `json:"app"`
	DB  int `json:"db"`
}

func NewHealthApiHandler(db *gorm.DB) *HealthApiHandler {
	return &HealthApiHandler{db: db}
}

func (h *HealthApiHandler) RegisterRoutes(router chi.Router) {
	router.Get("/health", h.Get)
}

// @Summary Check the application's health status
// @ID get-health
// @Tags misc
// @Produce plain
// @Success 200 {string} string
// @Router /health [get]
func (h *HealthApiHandler) Get(w http.ResponseWriter, r *http.Request) {
	var dbStatus int
	if sqlDb, err := h.db.DB(); err == nil {
		if err := sqlDb.Ping(); err == nil {
			dbStatus = 1
		}
	}

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		helpers.RespondJSON(w, r, http.StatusOK, HealthResponse{App: 1, DB: dbStatus})
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(fmt.Sprintf("app=1\ndb=%d", dbStatus)))
}
