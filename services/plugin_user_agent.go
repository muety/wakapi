package services

import (
	"time"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
	"github.com/patrickmn/go-cache"
)

type PluginUserAgentService struct {
	config     *config.Config
	cache      *cache.Cache
	repository repositories.IPluginUserAgentRepository
}

func NewPluginUserAgentService(repo repositories.IPluginUserAgentRepository) *PluginUserAgentService {
	return &PluginUserAgentService{
		config:     config.Get(),
		cache:      cache.New(1*time.Hour, 2*time.Hour),
		repository: repo,
	}
}

func (srv *PluginUserAgentService) CreateOrUpdate(useragent, user_id string) (*models.PluginUserAgent, error) {
	return srv.repository.CreateOrUpdate(useragent, user_id)
}

func (srv *PluginUserAgentService) FetchUserAgents(user_id string) ([]*models.PluginUserAgent, error) {
	return srv.repository.FetchUserAgents(user_id)
}

type IPluginUserAgentService interface {
	CreateOrUpdate(useragent, user_id string) (*models.PluginUserAgent, error)
	FetchUserAgents(user_id string) ([]*models.PluginUserAgent, error)
}
