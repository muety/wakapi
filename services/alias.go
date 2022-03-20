package services

import (
	"errors"
	"fmt"
	datastructure "github.com/duke-git/lancet/v2/datastructure/set"
	"github.com/emvi/logbuch"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
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

func (srv *AliasService) MayInitializeUser(userId string) {
	if err := srv.InitializeUser(userId); err != nil {
		logbuch.Error("failed to initialize user alias map for user %s", userId)
	}
}

func (srv *AliasService) GetByUser(userId string) ([]*models.Alias, error) {
	if !srv.IsInitialized(userId) {
		srv.MayInitializeUser(userId)
	}
	if aliases, ok := userAliases.Load(userId); ok {
		return aliases.([]*models.Alias), nil
	} else {
		return nil, errors.New(fmt.Sprintf("no user aliases loaded for user %s", userId))
	}
}

func (srv *AliasService) GetByUserAndType(userId string, summaryType uint8) ([]*models.Alias, error) {
	check := func(a *models.Alias) bool {
		return a.Type == summaryType
	}
	return srv.getFiltered(userId, check)
}

func (srv *AliasService) GetByUserAndKeyAndType(userId, key string, summaryType uint8) ([]*models.Alias, error) {
	check := func(a *models.Alias) bool {
		return a.Key == key && a.Type == summaryType
	}
	return srv.getFiltered(userId, check)
}

func (srv *AliasService) GetAliasOrDefault(userId string, summaryType uint8, value string) (string, error) {
	if !srv.IsInitialized(userId) {
		srv.MayInitializeUser(userId)
	}

	if aliases, ok := userAliases.Load(userId); ok {
		for _, a := range aliases.([]*models.Alias) {
			if a.Type == summaryType && a.Value == value {
				return a.Key, nil
			}
		}
	}

	return value, nil
}

func (srv *AliasService) Create(alias *models.Alias) (*models.Alias, error) {
	result, err := srv.repository.Insert(alias)
	if err != nil {
		return nil, err
	}
	// manually update cache
	srv.updateCache(alias, false)
	// reload entire cache (async, though)
	go srv.MayInitializeUser(alias.UserID)

	return result, nil
}

func (srv *AliasService) Delete(alias *models.Alias) error {
	if alias.UserID == "" {
		return errors.New("no user id specified")
	}
	err := srv.repository.Delete(alias.ID)

	// manually update cache
	if err == nil {
		srv.updateCache(alias, false)
	}
	// reload entire cache (async, though)
	go srv.MayInitializeUser(alias.UserID)

	return err
}

func (srv *AliasService) DeleteMulti(aliases []*models.Alias) error {
	ids := make([]uint, len(aliases))
	affectedUsers := datastructure.NewSet[string]()

	for i, a := range aliases {
		if a.UserID == "" {
			return errors.New("no user id specified")
		}
		affectedUsers.Add(a.UserID)
		ids[i] = a.ID
	}

	err := srv.repository.DeleteBatch(ids)

	// manually update cache
	if err == nil {
		for _, a := range aliases {
			srv.updateCache(a, true)
		}
	}
	// reload entire cache (async, though)
	for k := range affectedUsers {
		go srv.MayInitializeUser(k)
	}

	return err
}

func (srv *AliasService) updateCache(reason *models.Alias, removal bool) {
	if !removal {
		if aliases, ok := userAliases.Load(reason.UserID); ok {
			updatedAliases := aliases.([]*models.Alias)
			updatedAliases = append(updatedAliases, reason)
			userAliases.Store(reason.UserID, updatedAliases)
		}
	} else {
		if aliases, ok := userAliases.Load(reason.UserID); ok {
			updatedAliases := make([]*models.Alias, 0, len(aliases.([]*models.Alias))) // if we only had generics...
			for _, a := range aliases.([]*models.Alias) {
				if a.ID != reason.ID {
					updatedAliases = append(updatedAliases, a)
				}
			}
			userAliases.Store(reason.UserID, updatedAliases)
		}
	}
}

func (srv *AliasService) getFiltered(userId string, check func(alias *models.Alias) bool) ([]*models.Alias, error) {
	if !srv.IsInitialized(userId) {
		srv.MayInitializeUser(userId)
	}
	if aliases, ok := userAliases.Load(userId); ok {
		filteredAliases := make([]*models.Alias, 0, len(aliases.([]*models.Alias)))
		for _, a := range aliases.([]*models.Alias) {
			if check(a) {
				filteredAliases = append(filteredAliases, a)
			}
		}
		return filteredAliases, nil
	} else {
		return nil, errors.New(fmt.Sprintf("no user aliases loaded for user %s", userId))
	}
}
