package services

import (
	"errors"
	"sync"

	"github.com/jinzhu/gorm"
	"github.com/muety/wakapi/models"
)

type AliasService struct {
	Config *models.Config
	Db     *gorm.DB
}

func NewAliasService(db *gorm.DB) *AliasService {
	return &AliasService{
		Config: models.GetConfig(),
		Db:     db,
	}
}

var userAliases sync.Map

func (srv *AliasService) LoadUserAliases(userId string) error {
	var aliases []*models.Alias
	if err := srv.Db.
		Where(&models.Alias{UserID: userId}).
		Find(&aliases).Error; err != nil {
		return err
	}

	userAliases.Store(userId, aliases)
	return nil
}

func (srv *AliasService) GetAliasOrDefault(userId string, summaryType uint8, value string) (string, error) {
	if ua, ok := userAliases.Load(userId); ok {
		for _, a := range ua.([]*models.Alias) {
			if a.Type == summaryType && a.Value == value {
				return a.Key, nil
			}
		}
		return value, nil
	}
	return "", errors.New("user aliases not initialized")
}

func (src *AliasService) IsInitialized(userId string) bool {
	if _, ok := userAliases.Load(userId); ok {
		return true
	}
	return false
}
