package api

import (
	"encoding/json"
	"net/http"

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
// @Produce json
// @Success 200 {object} map[string]int
// @Router /health [get]
func (h *HealthApiHandler) Get(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{App: 1, DB: 0}

	if sqlDb, err := h.db.DB(); err == nil {
		if err := sqlDb.Ping(); err == nil {
			response.DB = 1
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
