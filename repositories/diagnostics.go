package repositories

import (
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

type DiagnosticsRepository struct {
	db *gorm.DB
}

func NewDiagnosticsRepository(db *gorm.DB) *DiagnosticsRepository {
	return &DiagnosticsRepository{db: db}
}

func (r *DiagnosticsRepository) Insert(diagnostics *models.Diagnostics) (*models.Diagnostics, error) {
	return diagnostics, r.db.Create(diagnostics).Error
}
