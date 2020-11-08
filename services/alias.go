package services

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/repositories"
	"sync"

	"github.com/muety/wakapi/models"
)

type AliasService struct {
	config     *config.Config
	repository repositories.IAliasRepository
}

func NewAliasService(aliasRepo repositories.IAliasRepository) *AliasService {
	return &AliasService{
		config:     config.Get(),
		repository: aliasRepo,
	}
}

var userAliases sync.Map

func (srv *AliasService) LoadUserAliases(userId string) error {
	aliases, err := srv.repository.GetByUser(userId)
	if err == nil {
		userAliases.Store(userId, aliases)
	}
	return err
}

func (srv *AliasService) GetAliasOrDefault(userId string, summaryType uint8, value string) (string, error) {
	if !srv.IsInitialized(userId) {
		if err := srv.LoadUserAliases(userId); err != nil {
			return "", err
		}
	}

	aliases, _ := userAliases.Load(userId)
	for _, a := range aliases.([]*models.Alias) {
		if a.Type == summaryType && a.Value == value {
			return a.Key, nil
		}
	}
	return value, nil
}

func (srv *AliasService) IsInitialized(userId string) bool {
	if _, ok := userAliases.Load(userId); ok {
		return true
	}
	return false
}
