package services

import (
	"github.com/jinzhu/gorm"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
)

type CustomRuleService struct {
	Config *config.Config
	Db     *gorm.DB
}

func NewCustomRuleService(db *gorm.DB) *CustomRuleService {
	return &CustomRuleService{
		Config: config.Get(),
		Db:     db,
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
	if err := srv.Db.
		Where(&models.CustomRule{UserID: userId}).
		Find(&rules).Error; err != nil {
		return rules, err
	}

	return rules, nil
}

func (srv *CustomRuleService) Create(rule *models.CustomRule) (*models.CustomRule, error) {
	result := srv.Db.Create(rule)
	if err := result.Error; err != nil {
		return nil, err
	}

	return rule, nil
}

func (srv *CustomRuleService) Delete(rule *models.CustomRule) {
	srv.Db.
		Where("id = ?", rule.ID).
		Where("user_id = ?", rule.UserID).
		Delete(models.CustomRule{})
}
