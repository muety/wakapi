package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/muety/wakapi/helpers"
	"gorm.io/gorm"
)

type HealthApiHandler struct {
	db *gorm.DB
}

func NewHealthApiHandler(db *gorm.DB) *HealthApiHandler {
	return &HealthApiHandler{db: db}
}

func (h *HealthApiHandler) RegisterRoutes(router chi.Router) {
	router.Get("/health", h.Get)
	router.Get("/up", h.Get)
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

	helpers.RespondJSON(w, r, http.StatusAccepted, map[string]interface{}{
		"message": "All systems operating within normal parameters",
		"status":  http.StatusAccepted,
		"db":      dbStatus,
		"app":     1,
	})
}
