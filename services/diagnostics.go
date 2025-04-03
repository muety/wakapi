package services

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
	"gorm.io/gorm"
)

type DiagnosticsService struct {
	config     *config.Config
	repository repositories.IDiagnosticsRepository
}

func NewDiagnosticsService(db *gorm.DB) *DiagnosticsService {
	diagnosticsRepository := repositories.NewDiagnosticsRepository(db)
	return &DiagnosticsService{
		config:     config.Get(),
		repository: diagnosticsRepository,
	}
}

func (srv *DiagnosticsService) Create(diagnostics *models.Diagnostics) (*models.Diagnostics, error) {
	diagnostics.ID = 0
	return srv.repository.Insert(diagnostics)
}
