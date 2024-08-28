package repositories

import (
	"fmt"

	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

type ClientRepository struct {
	db *gorm.DB
}

func NewClientRepository(db *gorm.DB) *ClientRepository {
	return &ClientRepository{db: db}
}

func (r *ClientRepository) FindOne(attributes models.Client) (*models.Client, error) {
	u := &models.Client{}
	result := r.db.Where(&attributes).First(u)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // No record found
		}
		return nil, result.Error
	}
	return u, nil
}

func (r *ClientRepository) Create(client *models.Client) (*models.Client, error) {
	result := r.db.Create(client)
	if err := result.Error; err != nil {
		return nil, err
	}
	return client, nil
}

func (r *ClientRepository) Update(client *models.Client, update *models.ClientUpdate) (*models.Client, error) {

	result := r.db.Model(client).Updates(update)
	if err := result.Error; err != nil {
		return nil, err
	}

	return client, nil
}

func (r *ClientRepository) FetchUserClients(userID, query string) ([]*models.Client, error) {
	var clients []*models.Client

	builtQuery := r.db.
		Order("created_at desc").
		Limit(100). // TODO: paginate - when this becomes necessary. The average user has a limited number of clients - more an enterprise thingy?
		Where(&models.Client{UserID: userID})

	if query != "" {
		likeQuery := fmt.Sprintf("%%%s%%", query)
		builtQuery.Where("name iLIKE ?", likeQuery) // ILIKE IS MOSTLY PG - !#postgresql
	}
	if err := builtQuery.Find(&clients).Error; err != nil {
		return nil, err
	}
	return clients, nil
}

func (r *ClientRepository) DeleteByIdAndUser(clientID, userID string) error {
	if err := r.db.
		Where("id = ?", clientID).
		Where("user_id = ?", userID).
		Delete(models.Client{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *ClientRepository) GetByIdForUser(clientID, userID string) (*models.Client, error) {
	g := &models.Client{}

	err := r.db.Where(models.Client{ID: clientID, UserID: userID}).First(g).Error
	if err != nil {
		return g, err
	}

	if g.ID != "" {
		return g, nil
	}
	return nil, err
}

type IClientRepository interface {
	FindOne(attributes models.Client) (*models.Client, error)
	Create(client *models.Client) (*models.Client, error)
	Update(client *models.Client, update *models.ClientUpdate) (*models.Client, error)
	FetchUserClients(userID, query string) ([]*models.Client, error)
	DeleteByIdAndUser(clientID, userID string) error
	GetByIdForUser(clientID, userID string) (*models.Client, error)
}
