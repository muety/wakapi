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
		// branch summaries are currently not persisted, as only relevant in combination with project filter
		Find(&summaries).Error; err != nil {
		return nil, err
	}

	if err := r.populateItems(summaries); err != nil {
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
		// branch summaries are currently not persisted, as only relevant in combination with project filter
		Find(&summaries).Error; err != nil {
		return nil, err
	}

	if err := r.populateItems(summaries); err != nil {
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

// inplace
func (r *SummaryRepository) populateItems(summaries []*models.Summary) error {
	summaryMap := map[uint]*models.Summary{}
	summaryIds := make([]uint, len(summaries))
	for i, s := range summaries {
		if s.NumHeartbeats == 0 {
			continue
		}
		summaryMap[s.ID] = s
		summaryIds[i] = s.ID
	}

	var items []*models.SummaryItem

	if err := r.db.
		Model(&models.SummaryItem{}).
		Where("summary_id in ?", summaryIds).
		Find(&items).Error; err != nil {
		return err
	}

	for _, item := range items {
		l := summaryMap[item.SummaryID].ItemsByType(item.Type)
		*l = append(*l, item)
	}

	return nil
}
