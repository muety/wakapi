package services

import (
	"time"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
	"github.com/patrickmn/go-cache"
)

type InvoiceService struct {
	config     *config.Config
	cache      *cache.Cache
	repository *repositories.InvoiceRepository
}

func NewInvoiceService(repository *repositories.InvoiceRepository) *InvoiceService {
	return &InvoiceService{
		config:     config.Get(),
		cache:      cache.New(1*time.Hour, 2*time.Hour),
		repository: repository,
	}
}

func (srv *InvoiceService) Create(newInvoice *models.Invoice) (*models.Invoice, error) {
	return srv.repository.Create(newInvoice)
}

func (srv *InvoiceService) Update(client *models.Invoice, update *models.InvoiceUpdate) (*models.Invoice, error) {
	return srv.repository.Update(client, update)
}

func (srv *InvoiceService) GetInvoiceForUser(id, userID string) (*models.Invoice, error) {
	return srv.repository.GetByIdForUser(id, userID)
}

func (srv *InvoiceService) DeleteInvoice(id, userID string) error {
	return srv.repository.DeleteByIdAndUser(id, userID)
}

func (srv *InvoiceService) FetchUserInvoices(id, query string) ([]*models.Invoice, error) {
	return srv.repository.FetchUserInvoices(id, query)
}
