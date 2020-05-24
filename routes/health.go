package routes

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"net/http"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) ApiGet(w http.ResponseWriter, r *http.Request) {
	var dbStatus int
	if err := h.db.DB().Ping(); err == nil {
		dbStatus = 1
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(fmt.Sprintf("app=1\ndb=%d", dbStatus)))
}
