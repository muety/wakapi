package repositories

import (
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type HeartbeatRepository struct {
	db *gorm.DB
}

func NewHeartbeatRepository(db *gorm.DB) *HeartbeatRepository {
	return &HeartbeatRepository{db: db}
}

// Use with caution!!
func (r *HeartbeatRepository) GetAll() ([]*models.Heartbeat, error) {
	var heartbeats []*models.Heartbeat
	if err := r.db.Find(&heartbeats).Error; err != nil {
		return nil, err
	}
	return heartbeats, nil
}

func (r *HeartbeatRepository) InsertBatch(heartbeats []*models.Heartbeat) error {
	if err := r.db.
		Clauses(clause.OnConflict{
			DoNothing: true,
		}).
		Create(&heartbeats).Error; err != nil {
		return err
	}
	return nil
}

func (r *HeartbeatRepository) GetLatestByOriginAndUser(origin string, user *models.User) (*models.Heartbeat, error) {
	var heartbeat models.Heartbeat
	if err := r.db.
		Model(&models.Heartbeat{}).
		Where(&models.Heartbeat{
			UserID: user.ID,
			Origin: origin,
		}).
		Order("time desc").
		First(&heartbeat).Error; err != nil {
		return nil, err
	}
	return &heartbeat, nil
}

func (r *HeartbeatRepository) GetAllWithin(from, to time.Time, user *models.User) ([]*models.Heartbeat, error) {
	var heartbeats []*models.Heartbeat
	if err := r.db.
		Where(&models.Heartbeat{UserID: user.ID}).
		Where("time >= ?", from.Local()).
		Where("time < ?", to.Local()).
		Order("time asc").
		Find(&heartbeats).Error; err != nil {
		return nil, err
	}
	return heartbeats, nil
}

func (r *HeartbeatRepository) GetFirstByUsers() ([]*models.TimeByUser, error) {
	var result []*models.TimeByUser
	r.db.Model(&models.User{}).
		Select("users.id as user, min(time) as time").
		Joins("left join heartbeats on users.id = heartbeats.user_id").
		Group("user").
		Scan(&result)
	return result, nil
}

func (r *HeartbeatRepository) GetLastByUsers() ([]*models.TimeByUser, error) {
	var result []*models.TimeByUser
	r.db.Model(&models.User{}).
		Select("users.id as user, max(time) as time").
		Joins("left join heartbeats on users.id = heartbeats.user_id").
		Group("user").
		Scan(&result)
	return result, nil
}

func (r *HeartbeatRepository) Count() (int64, error) {
	var count int64
	if err := r.db.
		Model(&models.Heartbeat{}).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *HeartbeatRepository) CountByUser(user *models.User) (int64, error) {
	var count int64
	if err := r.db.
		Model(&models.Heartbeat{}).
		Where(&models.Heartbeat{UserID: user.ID}).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *HeartbeatRepository) CountByUsers(users []*models.User) ([]*models.CountByUser, error) {
	var counts []*models.CountByUser

	userIds := make([]string, len(users))
	for i, u := range users {
		userIds[i] = u.ID
	}

	if err := r.db.
		Model(&models.User{}).
		Select("users.id as user, count(heartbeats.id) as count").
		Joins("left join heartbeats on users.id = heartbeats.user_id").
		Where("user_id in ?", userIds).
		Group("user").
		Find(&counts).Error; err != nil {
		return counts, err
	}
	return counts, nil
}

func (r *HeartbeatRepository) DeleteBefore(t time.Time) error {
	if err := r.db.
		Where("time <= ?", t.Local()).
		Delete(models.Heartbeat{}).Error; err != nil {
		return err
	}
	return nil
}
