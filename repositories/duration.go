package repositories

import (
	"time"

	"github.com/duke-git/lancet/v2/condition"
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

func (r *DurationRepository) GetAll() ([]*models.Duration, error) {
	var durations []*models.Duration
	q := r.db.Model(&models.Duration{})
	q = r.queryAddTimeSorting(q, false)
	if err := q.Find(&durations).Error; err != nil {
		return nil, err
	}
	return durations, nil
}

func (r *DurationRepository) StreamAllBatched(batchSize int) (chan []*models.Duration, error) {
	out := make(chan []*models.Duration)

	q := r.db.Model(&models.Duration{})
	q = r.queryAddTimeSorting(q, false)
	rows, err := q.Rows()
	if err != nil {
		return nil, err
	}

	go streamRowsBatched[models.Duration](rows, out, r.db, batchSize, func(err error) {
		conf.Log().Error("failed to scan duration row", "error", err)
	})
	return out, nil
}

func (r *DurationRepository) StreamByUserBatched(user *models.User, batchSize int) (chan []*models.Duration, error) {
	out := make(chan []*models.Duration)

	q := r.db.Model(&models.Duration{}).Where(&models.Duration{UserID: user.ID})
	q = r.queryAddTimeSorting(q, false)
	rows, err := q.Rows()
	if err != nil {
		return nil, err
	}

	go streamRowsBatched[models.Duration](rows, out, r.db, batchSize, func(err error) {
		conf.Log().Error("failed to scan duration row", "error", err)
	})
	return out, nil
}

func (r *DurationRepository) GetAllWithin(from, to time.Time, user *models.User) ([]*models.Duration, error) {
	return r.GetAllWithinByFilters(from, to, user, map[string][]string{})
}

func (r *DurationRepository) GetAllWithinByFilters(from, to time.Time, user *models.User, filterMap map[string][]string) ([]*models.Duration, error) {
	var durations []*models.Duration

	q := r.db.Model(&models.Duration{}).Where(&models.Duration{UserID: user.ID})
	q = r.queryAddTimeFilterBetween(q, from.Local(), to.Local())
	q = r.queryAddTimeSorting(q, false)

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
	q := r.db.Model(&models.Duration{}).Where(&models.Duration{UserID: user.ID})
	q = r.queryAddTimeSorting(q, true)
	err := q.First(&duration).Error
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
	q := r.db.Model(models.Duration{}).Where("user_id = ?", user.ID)
	q = r.queryAddTimeFilterLessEqual(q, t.Local())
	if err := q.Delete(models.Duration{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *DurationRepository) queryAddTimeFilterBetween(q *gorm.DB, from, to time.Time) *gorm.DB {
	if r.config.Db.IsSQLite() {
		q = q.
			Where("time_real >= julianday(?)", from.Local()).
			Where("time_real < julianday(?)", to.Local())
	} else {
		q = q.
			Where("time >= ?", from.Local()).
			Where("time < ?", to.Local())
	}
	return q
}

func (r *DurationRepository) queryAddTimeFilterLessEqual(q *gorm.DB, t time.Time) *gorm.DB {
	if r.config.Db.IsSQLite() {
		return q.Where("time_real <= julianday(?)", t.Local())
	}
	return q.Where("time <= ?", t.Local())
}

func (r *DurationRepository) queryAddTimeSorting(q *gorm.DB, desc bool) *gorm.DB {
	order := condition.Ternary(desc, "desc", "asc")
	if r.config.Db.IsSQLite() {
		return q.Order("time_real " + order)
	}
	return q.Order("time " + order)
}
