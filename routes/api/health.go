package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type HealthApiHandler struct {
	db *gorm.DB
}

func NewHealthApiHandler(db *gorm.DB) *HealthApiHandler {
	return &HealthApiHandler{db: db}
}

func (h *HealthApiHandler) RegisterRoutes(router *mux.Router) {
	r := router.PathPrefix("/health").Subrouter()
	r.Path("").Methods(http.MethodGet).HandlerFunc(h.Get)
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

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(fmt.Sprintf("app=1\ndb=%d", dbStatus)))
}
