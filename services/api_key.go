package services

import (
	"errors"
	"time"

	"github.com/leandro-lugaresi/hub"
	"github.com/patrickmn/go-cache"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
)

type ApiKeyService struct {
	config     *config.Config
	cache      *cache.Cache
	eventBus   *hub.Hub
	repository repositories.IApiKeyRepository
}

func NewApiKeyService(apiKeyRepository repositories.IApiKeyRepository) *ApiKeyService {
	return &ApiKeyService{
		config:     config.Get(),
		eventBus:   config.EventBus(),
		repository: apiKeyRepository,
		cache:      cache.New(24*time.Hour, 24*time.Hour),
	}
}

func (srv *ApiKeyService) GetById(id uint) (*models.ApiKey, error) {
	return srv.repository.GetById(id)
}

func (srv *ApiKeyService) GetByApiKey(apiKey string) (*models.ApiKey, error) {
	return srv.repository.GetByApiKey(apiKey)
}

func (srv *ApiKeyService) GetByRWApiKey(apiKey string) (*models.ApiKey, error) {
	return srv.repository.GetByRWApiKey(apiKey)
}

func (srv *ApiKeyService) GetByUser(userId string) ([]*models.ApiKey, error) {
	if labels, found := srv.cache.Get(userId); found {
		return labels.([]*models.ApiKey), nil
	}

	labels, err := srv.repository.GetByUser(userId)
	if err != nil {
		return nil, err
	}
	srv.cache.Set(userId, labels, cache.DefaultExpiration)
	return labels, nil
}

func (srv *ApiKeyService) Create(label *models.ApiKey) (*models.ApiKey, error) {
	result, err := srv.repository.Insert(label)
	if err != nil {
		return nil, err
	}

	srv.cache.Delete(result.UserID)
	srv.notifyUpdate(label, false)
	return result, nil
}

func (srv *ApiKeyService) Delete(label *models.ApiKey) error {
	if label.UserID == "" {
		return errors.New("no user id specified")
	}
	err := srv.repository.Delete(label.ID)
	srv.cache.Delete(label.UserID)
	srv.notifyUpdate(label, true)
	return err
}

func (srv *ApiKeyService) notifyUpdate(label *models.ApiKey, isDelete bool) {
	name := config.EventApiKeyCreate
	if isDelete {
		name = config.EventApiKeyDelete
	}
	srv.eventBus.Publish(hub.Message{
		Name:   name,
		Fields: map[string]interface{}{config.FieldPayload: label, config.FieldUserId: label.UserID},
	})
}
