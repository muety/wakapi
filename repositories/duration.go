package repositories

import (
	"time"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

type DurationRepository struct {
	BaseRepository
	config *conf.Config
}

func NewDurationRepository(db *gorm.DB) *DurationRepository {
	return &DurationRepository{BaseRepository: NewBaseRepository(db), config: conf.Get()}
}

func (r *DurationRepository) GetAllWithin(from, to time.Time, user *models.User) ([]*models.Duration, error) {
	return r.GetAllWithinByFilters(from, to, user, map[string][]string{})
}

func (r *DurationRepository) GetAllWithinByFilters(from, to time.Time, user *models.User, filterMap map[string][]string) ([]*models.Duration, error) {
	var durations []*models.Duration

	q := r.db.
		Where(&models.Duration{UserID: user.ID}).
		Where("time >= ?", from.Local()).
		Where("time < ?", to.Local()).
		Order("time asc")

	if len(filterMap) > 0 {
		q = filteredQuery(q, filterMap)
	}

	if err := q.Find(&durations).Error; err != nil {
		return nil, err
	}
	return durations, nil
}

func (r *DurationRepository) GetLatestByUser(user *models.User) (*models.Duration, error) {
	var duration *models.Duration
	err := r.db.
		Where(&models.Duration{UserID: user.ID}).
		Order("time desc").
		First(&duration).
		Error
	return duration, err
}

func (r *DurationRepository) InsertBatch(durations []*models.Duration) error {
	return InsertBatchChunked[*models.Duration](durations, &models.Duration{}, r.db)
}

func (r *DurationRepository) DeleteByUser(user *models.User) error {
	if err := r.db.
		Where("user_id = ?", user.ID).
		Delete(models.Duration{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *DurationRepository) DeleteByUserBefore(user *models.User, t time.Time) error {
	if err := r.db.
		Where("user_id = ?", user.ID).
		Where("time <= ?", t.Local()).
		Delete(models.Duration{}).Error; err != nil {
		return err
	}
	return nil
}
