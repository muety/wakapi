package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/duke-git/lancet/v2/datetime"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"
)

type ClientService struct {
	config *config.Config
	cache  *cache.Cache
	db     *gorm.DB
}

func NewClientService(db *gorm.DB) *ClientService {
	return &ClientService{
		config: config.Get(),
		cache:  cache.New(1*time.Hour, 2*time.Hour),
		db:     db,
	}
}

func (srv *ClientService) Create(newClient *models.Client) (*models.Client, error) {
	result := srv.db.Create(newClient)
	if err := result.Error; err != nil {
		return nil, err
	}
	return newClient, nil
}

func (srv *ClientService) Update(client *models.Client, update *models.ClientUpdate) (*models.Client, error) {
	result := srv.db.Model(client).Updates(update)
	if err := result.Error; err != nil {
		return nil, err
	}

	return client, nil
}

func (srv *ClientService) GetClientForUser(clientID, userID string) (*models.Client, error) {
	g := &models.Client{}

	err := srv.db.Where(models.Client{ID: clientID, UserID: userID}).First(g).Error
	if err != nil {
		return g, err
	}

	if g.ID != "" {
		return g, nil
	}
	return nil, err
}

func (srv *ClientService) DeleteClient(clientID, userID string) error {
	if err := srv.db.
		Where("id = ?", clientID).
		Where("user_id = ?", userID).
		Delete(models.Client{}).Error; err != nil {
		return err
	}
	return nil
}

func (srv *ClientService) FetchUserClients(userID, query string) ([]*models.Client, error) {
	var clients []*models.Client

	builtQuery := srv.db.
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

func (srv *ClientService) FetchClientInvoiceLineItems(client *models.Client, user *models.User, summarySrvc ISummaryService, start, end time.Time) (*models.Summary, error) {
	end = datetime.EndOfDay(end)

	if !end.After(start) {
		return nil, errors.New("'end' date must be after 'start' date")
	}

	filters := client.GetSummaryFilters()
	summary, err := summarySrvc.Aliased(start, end, user, summarySrvc.Retrieve, filters, end.After(time.Now()))

	if err != nil {
		return nil, err
	}

	summary.FromTime = models.CustomTime(start)
	summary.ToTime = models.CustomTime(end.Add(-1 * time.Second))

	return summary, nil
}

type IClientService interface {
	Create(client *models.Client) (*models.Client, error)
	Update(client *models.Client, update *models.ClientUpdate) (*models.Client, error)
	GetClientForUser(id, userID string) (*models.Client, error)
	DeleteClient(id, userID string) error
	FetchUserClients(id, query string) ([]*models.Client, error)
	FetchClientInvoiceLineItems(client *models.Client, user *models.User, summarySrvc ISummaryService, start, end time.Time) (*models.Summary, error)
}
