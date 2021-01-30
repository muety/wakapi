package routes

import (
	"fmt"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"net/http"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) RegisterRoutes(router *mux.Router) {}

func (h *HealthHandler) RegisterAPIRoutes(router *mux.Router) {
	router.Methods(http.MethodGet).HandlerFunc(h.ApiGet)
}

func (h *HealthHandler) ApiGet(w http.ResponseWriter, r *http.Request) {
	var dbStatus int
	if sqlDb, err := h.db.DB(); err == nil {
		if err := sqlDb.Ping(); err == nil {
			dbStatus = 1
		}
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(fmt.Sprintf("app=1\ndb=%d", dbStatus)))
}
