package repositories

import (
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
	"time"
)

type SummaryRepository struct {
	db *gorm.DB
}

func NewSummaryRepository(db *gorm.DB) *SummaryRepository {
	return &SummaryRepository{db: db}
}

func (r *SummaryRepository) GetAll() ([]*models.Summary, error) {
	var summaries []*models.Summary
	if err := r.db.
		Order("from_time asc").
		Preload("Projects", "type = ?", models.SummaryProject).
		Preload("Languages", "type = ?", models.SummaryLanguage).
		Preload("Editors", "type = ?", models.SummaryEditor).
		Preload("OperatingSystems", "type = ?", models.SummaryOS).
		Preload("Machines", "type = ?", models.SummaryMachine).
		Find(&summaries).Error; err != nil {
		return nil, err
	}
	return summaries, nil
}

func (r *SummaryRepository) Insert(summary *models.Summary) error {
	if err := r.db.Create(summary).Error; err != nil {
		return err
	}
	return nil
}

func (r *SummaryRepository) GetByUserWithin(user *models.User, from, to time.Time) ([]*models.Summary, error) {
	var summaries []*models.Summary
	if err := r.db.
		Where(&models.Summary{UserID: user.ID}).
		Where("from_time >= ?", from.Local()).
		Where("to_time <= ?", to.Local()).
		Order("from_time asc").
		Preload("Projects", "type = ?", models.SummaryProject).
		Preload("Languages", "type = ?", models.SummaryLanguage).
		Preload("Editors", "type = ?", models.SummaryEditor).
		Preload("OperatingSystems", "type = ?", models.SummaryOS).
		Preload("Machines", "type = ?", models.SummaryMachine).
		Find(&summaries).Error; err != nil {
		return nil, err
	}
	return summaries, nil
}

func (r *SummaryRepository) GetLastByUser() ([]*models.TimeByUser, error) {
	var result []*models.TimeByUser
	r.db.Model(&models.User{}).
		Select("users.id as user, max(to_time) as time").
		Joins("left join summaries on users.id = summaries.user_id").
		Group("user").
		Scan(&result)
	return result, nil
}

func (r *SummaryRepository) DeleteByUser(userId string) error {
	if err := r.db.
		Where("user_id = ?", userId).
		Delete(models.Summary{}).Error; err != nil {
		return err
	}
	return nil
}
