package services

import (
	"github.com/jinzhu/gorm"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/patrickmn/go-cache"
	"time"
)

type CustomRuleService struct {
	Config *config.Config
	Db     *gorm.DB
	cache  *cache.Cache
}

func NewCustomRuleService(db *gorm.DB) *CustomRuleService {
	return &CustomRuleService{
		Config: config.Get(),
		Db:     db,
		cache:  cache.New(1*time.Hour, 2*time.Hour),
	}
}

func (srv *CustomRuleService) GetCustomRuleById(customRuleId uint) (*models.CustomRule, error) {
	r := &models.CustomRule{}
	if err := srv.Db.Where(&models.CustomRule{ID: customRuleId}).First(r).Error; err != nil {
		return r, err
	}
	return r, nil
}

func (srv *CustomRuleService) GetCustomRuleForUser(userId string) ([]*models.CustomRule, error) {
	var rules []*models.CustomRule
	if rules, found := srv.cache.Get(userId); found {
		return rules.([]*models.CustomRule), nil
	}

	if err := srv.Db.
		Where(&models.CustomRule{UserID: userId}).
		Find(&rules).Error; err != nil {
		return rules, err
	}
	srv.cache.Set(userId, rules, cache.DefaultExpiration)
	return rules, nil
}

func (srv *CustomRuleService) Create(rule *models.CustomRule) (*models.CustomRule, error) {
	result := srv.Db.Create(rule)
	if err := result.Error; err != nil {
		return nil, err
	}
	srv.cache.Delete(rule.UserID)

	return rule, nil
}

func (srv *CustomRuleService) Delete(rule *models.CustomRule) {
	srv.Db.
		Where("id = ?", rule.ID).
		Where("user_id = ?", rule.UserID).
		Delete(models.CustomRule{})
	srv.cache.Delete(rule.UserID)
}
