package services

import (
	"fmt"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"
)

type InvoiceService struct {
	config *config.Config
	cache  *cache.Cache
	db     *gorm.DB
}

func NewInvoiceService(db *gorm.DB) IInvoiceService {
	return &InvoiceService{
		config: config.Get(),
		cache:  cache.New(1*time.Hour, 2*time.Hour),
		db:     db,
	}
}

func (srv *InvoiceService) Create(newInvoice *models.Invoice) (*models.Invoice, error) {
	nanoID, err := gonanoid.Generate("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ", 8) // Custom length/charset
	if err != nil {
		return nil, err
	}
	newInvoice.InvoiceID = fmt.Sprintf("INV-%s", nanoID)
	result := srv.db.Create(newInvoice)
	if err := result.Error; err != nil {
		return nil, err
	}
	return newInvoice, nil
}

func (srv *InvoiceService) Update(invoice *models.Invoice, update *models.InvoiceUpdate) (*models.Invoice, error) {
	result := srv.db.Model(invoice).Updates(update)
	if err := result.Error; err != nil {
		return nil, err
	}

	return invoice, nil
}

func (srv *InvoiceService) GetInvoiceForUser(invoiceID, userID string) (*models.Invoice, error) {
	invoice := &models.Invoice{}

	result := srv.db.Where(models.Invoice{ID: invoiceID, UserID: userID}).Preload("Client").First(invoice)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // No record found
		}
		return nil, result.Error
	}
	return invoice, nil
}

func (srv *InvoiceService) DeleteInvoice(invoiceID, userID string) error {
	if err := srv.db.
		Where("id = ?", invoiceID).
		Where("user_id = ?", userID).
		Delete(models.Invoice{}).Error; err != nil {
		return err
	}
	return nil
}

func (srv *InvoiceService) FetchUserInvoices(userID, query string) ([]*models.Invoice, error) {
	var invoices []*models.Invoice

	builtQuery := srv.db.
		Order("created_at desc").
		Limit(100). // TODO: paginate - when this becomes necessary. The average user has a limited number of invoices - more an enterprise thingy?
		Where(&models.Invoice{UserID: userID})

	if err := builtQuery.Preload("Client").Find(&invoices).Error; err != nil {
		return nil, err
	}
	return invoices, nil
}

type IInvoiceService interface {
	Create(*models.Invoice) (*models.Invoice, error)
	Update(*models.Invoice, *models.InvoiceUpdate) (*models.Invoice, error)
	GetInvoiceForUser(invoiceID, userID string) (*models.Invoice, error)
	DeleteInvoice(invoiceID, userID string) error
	FetchUserInvoices(userID, query string) ([]*models.Invoice, error)
}
