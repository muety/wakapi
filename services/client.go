package services

import (
	"time"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
	"github.com/patrickmn/go-cache"
)

type ClientService struct {
	config     *config.Config
	cache      *cache.Cache
	repository repositories.IClientRepository
}

func NewClientService(repository repositories.IClientRepository) *ClientService {
	return &ClientService{
		config:     config.Get(),
		cache:      cache.New(1*time.Hour, 2*time.Hour),
		repository: repository,
	}
}

func (srv *ClientService) Create(newClient *models.Client) (*models.Client, error) {
	return srv.repository.Create(newClient)
}

func (srv *ClientService) Update(client *models.Client, update *models.ClientUpdate) (*models.Client, error) {
	return srv.repository.Update(client, update)
}

func (srv *ClientService) GetClientForUser(id, userID string) (*models.Client, error) {
	return srv.repository.GetByIdForUser(id, userID)
}

func (srv *ClientService) DeleteClient(id, userID string) error {
	return srv.repository.DeleteByIdAndUser(id, userID)
}

func (srv *ClientService) FetchUserClients(id, query string) ([]*models.Client, error) {
	return srv.repository.FetchUserClients(id, query)
}

type IClientService interface {
	Create(client *models.Client) (*models.Client, error)
	Update(client *models.Client, update *models.ClientUpdate) (*models.Client, error)
	GetClientForUser(id, userID string) (*models.Client, error)
	DeleteClient(id, userID string) error
	FetchUserClients(id, query string) ([]*models.Client, error)
}
