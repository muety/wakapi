package services

import (
	"errors"
	"strings"
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
	srv := &ApiKeyService{
		config:     config.Get(),
		eventBus:   config.EventBus(),
		repository: apiKeyRepository,
		cache:      cache.New(24*time.Hour, 24*time.Hour),
	}

	onApiKeyCreate := srv.eventBus.Subscribe(0, config.EventApiKeyCreate)
	go func(sub *hub.Subscription) {
		for m := range sub.Receiver {
			srv.invalidateUserCache(m.Fields[config.FieldUserId].(string))
		}
	}(&onApiKeyCreate)

	onApiKeyDelete := srv.eventBus.Subscribe(0, config.EventApiKeyDelete)
	go func(sub *hub.Subscription) {
		for m := range sub.Receiver {
			srv.invalidateUserCache(m.Fields[config.FieldUserId].(string))
		}
	}(&onApiKeyDelete)

	return srv
}

func (srv *ApiKeyService) GetByApiKey(apiKey string, requireFullAccessKey bool) (*models.ApiKey, error) {
	return srv.repository.GetByApiKey(apiKey, requireFullAccessKey)
}

func (srv *ApiKeyService) GetByUser(userId string) ([]*models.ApiKey, error) {
	if userApiKeys, found := srv.cache.Get(userId); found {
		return userApiKeys.([]*models.ApiKey), nil
	}

	userApiKeys, err := srv.repository.GetByUser(userId)
	if err != nil {
		return nil, err
	}
	srv.cache.Set(userId, userApiKeys, cache.DefaultExpiration)
	return userApiKeys, nil
}

func (srv *ApiKeyService) Create(apiKey *models.ApiKey) (*models.ApiKey, error) {
	result, err := srv.repository.Insert(apiKey)
	if err != nil {
		return nil, err
	}

	srv.cache.Delete(result.UserID)
	srv.notifyUpdate(apiKey, false)
	return result, nil
}

func (srv *ApiKeyService) Delete(apiKey *models.ApiKey) error {
	if apiKey.UserID == "" {
		return errors.New("no user id specified")
	}
	err := srv.repository.Delete(apiKey.ApiKey)
	srv.cache.Delete(apiKey.UserID)
	srv.notifyUpdate(apiKey, true)
	return err
}

func (srv *ApiKeyService) notifyUpdate(apiKey *models.ApiKey, isDelete bool) {
	name := config.EventApiKeyCreate
	if isDelete {
		name = config.EventApiKeyDelete
	}
	srv.eventBus.Publish(hub.Message{
		Name:   name,
		Fields: map[string]interface{}{config.FieldPayload: apiKey, config.FieldUserId: apiKey.UserID},
	})
}

func (srv *ApiKeyService) invalidateUserCache(userId string) {
	for key := range srv.cache.Items() {
		if strings.Contains(key, userId) {
			srv.cache.Delete(key)
		}
	}
}
