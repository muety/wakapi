package services

import (
	"errors"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
	"github.com/patrickmn/go-cache"
	"time"
)

type LanguageMappingService struct {
	config     *config.Config
	cache      *cache.Cache
	repository repositories.ILanguageMappingRepository
}

func NewLanguageMappingService(languageMappingsRepo repositories.ILanguageMappingRepository) *LanguageMappingService {
	return &LanguageMappingService{
		config:     config.Get(),
		repository: languageMappingsRepo,
		cache:      cache.New(24*time.Hour, 24*time.Hour),
	}
}

func (srv *LanguageMappingService) GetById(id uint) (*models.LanguageMapping, error) {
	return srv.repository.GetById(id)
}

func (srv *LanguageMappingService) GetByUser(userId string) ([]*models.LanguageMapping, error) {
	if mappings, found := srv.cache.Get(userId); found {
		return mappings.([]*models.LanguageMapping), nil
	}

	mappings, err := srv.repository.GetByUser(userId)
	if err != nil {
		return nil, err
	}
	srv.cache.Set(userId, mappings, cache.DefaultExpiration)
	return mappings, nil
}

func (srv *LanguageMappingService) ResolveByUser(userId string) (map[string]string, error) {
	mappings := srv.getServerMappings()
	userMappings, err := srv.GetByUser(userId)
	if err != nil {
		return nil, err
	}

	for _, m := range userMappings {
		mappings[m.Extension] = m.Language
	}
	return mappings, nil
}

func (srv *LanguageMappingService) Create(mapping *models.LanguageMapping) (*models.LanguageMapping, error) {
	result, err := srv.repository.Insert(mapping)
	if err != nil {
		return nil, err
	}

	srv.cache.Delete(result.UserID)
	return result, nil
}

func (srv *LanguageMappingService) Delete(mapping *models.LanguageMapping) error {
	if mapping.UserID == "" {
		return errors.New("no user id specified")
	}
	err := srv.repository.Delete(mapping.ID)
	srv.cache.Delete(mapping.UserID)
	return err
}

func (srv *LanguageMappingService) getServerMappings() map[string]string {
	// https://dave.cheney.net/2017/04/30/if-a-map-isnt-a-reference-variable-what-is-it
	return srv.config.App.GetCustomLanguages()
}
