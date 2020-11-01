package services

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
	"github.com/patrickmn/go-cache"
	"time"
)

type CustomRuleService struct {
	config     *config.Config
	repository *repositories.CustomRuleRepository
	cache      *cache.Cache
}

func NewCustomRuleService(customRuleRepo *repositories.CustomRuleRepository) *CustomRuleService {
	return &CustomRuleService{
		config:     config.Get(),
		repository: customRuleRepo,
		cache:      cache.New(1*time.Hour, 2*time.Hour),
	}
}

func (srv *CustomRuleService) GetCustomRuleById(id uint) (*models.CustomRule, error) {
	return srv.repository.GetById(id)
}

func (srv *CustomRuleService) GetCustomRuleForUser(userId string) ([]*models.CustomRule, error) {
	if rules, found := srv.cache.Get(userId); found {
		return rules.([]*models.CustomRule), nil
	}

	rules, err := srv.repository.GetByUser(userId)
	if err != nil {
		return nil, err
	}
	srv.cache.Set(userId, rules, cache.DefaultExpiration)
	return rules, nil
}

func (srv *CustomRuleService) Create(rule *models.CustomRule) (*models.CustomRule, error) {
	result, err := srv.repository.Insert(rule)
	if err != nil {
		return nil, err
	}

	srv.cache.Delete(result.UserID)
	return result, nil
}

func (srv *CustomRuleService) Delete(rule *models.CustomRule) error {
	err := srv.repository.Delete(rule.ID)
	srv.cache.Delete(rule.UserID)
	return err
}
