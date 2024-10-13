package repositories

import (
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

type InvoiceRepository struct {
	db *gorm.DB
}

func NewInvoiceRepository(db *gorm.DB) *InvoiceRepository {
	return &InvoiceRepository{db: db}
}

func (r *InvoiceRepository) FindOne(attributes models.Invoice) (*models.Invoice, error) {
	u := &models.Invoice{}
	result := r.db.Where(&attributes).First(u)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // No record found
		}
		return nil, result.Error
	}
	return u, nil
}

func (r *InvoiceRepository) Create(client *models.Invoice) (*models.Invoice, error) {
	result := r.db.Create(client)
	if err := result.Error; err != nil {
		return nil, err
	}
	return client, nil
}

func (r *InvoiceRepository) Update(client *models.Invoice, update *models.InvoiceUpdate) (*models.Invoice, error) {

	result := r.db.Model(client).Updates(update)
	if err := result.Error; err != nil {
		return nil, err
	}

	return client, nil
}

func (r *InvoiceRepository) FetchUserInvoices(userID, query string) ([]*models.Invoice, error) {
	var invoices []*models.Invoice

	builtQuery := r.db.
		Order("created_at desc").
		Limit(100). // TODO: paginate - when this becomes necessary. The average user has a limited number of invoices - more an enterprise thingy?
		Where(&models.Invoice{UserID: userID})

	if err := builtQuery.Preload("Client").Find(&invoices).Error; err != nil {
		return nil, err
	}
	return invoices, nil
}

func (r *InvoiceRepository) DeleteByIdAndUser(invoiceID, userID string) error {
	if err := r.db.
		Where("id = ?", invoiceID).
		Where("user_id = ?", userID).
		Delete(models.Invoice{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *InvoiceRepository) GetByIdForUser(invoiceID, userID string) (*models.Invoice, error) {
	g := &models.Invoice{}

	err := r.db.Where(models.Invoice{ID: invoiceID, UserID: userID}).First(g).Error
	if err != nil {
		return g, err
	}

	if g.ID != "" {
		return g, nil
	}
	return nil, err
}
