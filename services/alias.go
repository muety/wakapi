package services

import (
	"errors"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
	"log"
	"sync"
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

func (srv *AliasService) IsInitialized(userId string) bool {
	if _, ok := userAliases.Load(userId); ok {
		return true
	}
	return false
}

func (srv *AliasService) InitializeUser(userId string) error {
	aliases, err := srv.repository.GetByUser(userId)
	if err == nil {
		userAliases.Store(userId, aliases)
	}
	return err
}

func (srv *AliasService) GetByUser(userId string) ([]*models.Alias, error) {
	aliases, err := srv.repository.GetByUser(userId)
	if err != nil {
		return nil, err
	}
	return aliases, nil
}

func (srv *AliasService) GetByUserAndKeyAndType(userId, key string, summaryType uint8) ([]*models.Alias, error) {
	aliases, err := srv.repository.GetByUserAndKeyAndType(userId, key, summaryType)
	if err != nil {
		return nil, err
	}
	return aliases, nil
}

func (srv *AliasService) GetAliasOrDefault(userId string, summaryType uint8, value string) (string, error) {
	if !srv.IsInitialized(userId) {
		if err := srv.InitializeUser(userId); err != nil {
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

func (srv *AliasService) Create(alias *models.Alias) (*models.Alias, error) {
	result, err := srv.repository.Insert(alias)
	if err != nil {
		return nil, err
	}
	go srv.reinitUser(alias.UserID)
	return result, nil
}

func (srv *AliasService) Delete(alias *models.Alias) error {
	if alias.UserID == "" {
		return errors.New("no user id specified")
	}
	err := srv.repository.Delete(alias.ID)
	go srv.reinitUser(alias.UserID)
	return err
}

func (srv *AliasService) DeleteMulti(aliases []*models.Alias) error {
	ids := make([]uint, len(aliases))
	affectedUsers := make(map[string]bool)
	for i, a := range aliases {
		if a.UserID == "" {
			return errors.New("no user id specified")
		}
		affectedUsers[a.UserID] = true
		ids[i] = a.ID
	}

	err := srv.repository.DeleteBatch(ids)

	for k := range affectedUsers {
		go srv.reinitUser(k)
	}

	return err
}

func (srv *AliasService) reinitUser(userId string) {
	if err := srv.InitializeUser(userId); err != nil {
		log.Printf("error initializing user aliases – %v\n", err)
	}
}
