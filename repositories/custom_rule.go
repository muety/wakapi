package repositories

import (
	"github.com/jinzhu/gorm"
	"github.com/muety/wakapi/models"
)

type CustomRuleRepository struct {
	db *gorm.DB
}

func NewCustomRuleRepository(db *gorm.DB) *CustomRuleRepository {
	return &CustomRuleRepository{db: db}
}

func (r *CustomRuleRepository) GetById(id uint) (*models.CustomRule, error) {
	rule := &models.CustomRule{}
	if err := r.db.Where(&models.CustomRule{ID: id}).First(rule).Error; err != nil {
		return rule, err
	}
	return rule, nil
}

func (r *CustomRuleRepository) GetByUser(userId string) ([]*models.CustomRule, error) {
	var rules []*models.CustomRule
	if err := r.db.
		Where(&models.CustomRule{UserID: userId}).
		Find(&rules).Error; err != nil {
		return rules, err
	}
	return rules, nil
}

func (r *CustomRuleRepository) Insert(rule *models.CustomRule) (*models.CustomRule, error) {
	result := r.db.Create(rule)
	if err := result.Error; err != nil {
		return nil, err
	}
	return rule, nil
}

func (r *CustomRuleRepository) Delete(id uint) error {
	return r.db.
		Where("id = ?", id).
		Delete(models.CustomRule{}).Error
}
