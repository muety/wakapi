package services

import (
	"errors"

	"github.com/jinzhu/gorm"
	"github.com/n1try/wakapi/models"
)

type AliasService struct {
	Config *models.Config
	Db     *gorm.DB
}

var userAliases map[string][]*models.Alias

func (srv *AliasService) InitUser(userId string) error {
	if userAliases == nil {
		userAliases = make(map[string][]*models.Alias)
	}

	var aliases []*models.Alias
	if err := srv.Db.
		Where(&models.Alias{UserID: userId}).
		Find(&aliases).Error; err != nil {
		return err
	}

	userAliases[userId] = aliases
	return nil
}

func (srv *AliasService) GetAliasOrDefault(userId string, summaryType uint8, value string) (string, error) {
	if userAliases, ok := userAliases[userId]; ok {
		for _, a := range userAliases {
			if a.Type == summaryType && a.Value == value {
				return a.Key, nil
			}
		}
		return value, nil
	}
	return "", errors.New("User aliases not initialized")
}

func (src *AliasService) IsInitialized(userId string) bool {
	if _, ok := userAliases[userId]; ok {
		return true
	}
	return false
}
