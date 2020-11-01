package repositories

import (
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
	"time"
)

type HeartbeatRepository struct {
	db *gorm.DB
}

func NewHeartbeatRepository(db *gorm.DB) *HeartbeatRepository {
	return &HeartbeatRepository{db: db}
}

func (r *HeartbeatRepository) InsertBatch(heartbeats []*models.Heartbeat) error {
	if err := r.db.Create(&heartbeats).Error; err != nil {
		return err
	}
	return nil
}

func (r *HeartbeatRepository) GetAllWithin(from, to time.Time, user *models.User) ([]*models.Heartbeat, error) {
	var heartbeats []*models.Heartbeat
	if err := r.db.
		Where(&models.Heartbeat{UserID: user.ID}).
		Where("time >= ?", from).
		Where("time < ?", to).
		Order("time asc").
		Find(&heartbeats).Error; err != nil {
		return nil, err
	}
	return heartbeats, nil
}

// Will return *models.Heartbeat object with only user_id and time fields filled
func (r *HeartbeatRepository) GetFirstByUsers(userIds []string) ([]*models.Heartbeat, error) {
	var heartbeats []*models.Heartbeat
	if err := r.db.
		Table("heartbeats").
		Select("user_id, min(time) as time").
		Where("user_id IN (?)", userIds).
		Group("user_id").
		Scan(&heartbeats).Error; err != nil {
		return nil, err
	}
	return heartbeats, nil
}

func (r *HeartbeatRepository) DeleteBefore(t time.Time) error {
	if err := r.db.
		Where("time <= ?", t).
		Delete(models.Heartbeat{}).Error; err != nil {
		return err
	}
	return nil
}
