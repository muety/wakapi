package services

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
)

type DiagnosticsService struct {
	config     *config.Config
	repository repositories.IDiagnosticsRepository
}

func NewDiagnosticsService(diagnosticsRepo repositories.IDiagnosticsRepository) *DiagnosticsService {
	return &DiagnosticsService{
		config:     config.Get(),
		repository: diagnosticsRepo,
	}
}

func (srv *DiagnosticsService) Create(diagnostics *models.Diagnostics) (*models.Diagnostics, error) {
	diagnostics.ID = 0
	return srv.repository.Insert(diagnostics)
}
