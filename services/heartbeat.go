package services

import (
	"github.com/jasonlvhit/gocron"
	"github.com/muety/wakapi/utils"
	"log"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/muety/wakapi/models"
	gormbulk "github.com/t-tiger/gorm-bulk-insert"
)

const (
	TableHeartbeat  = "heartbeat"
	cleanUpInterval = time.Duration(aggregateIntervalDays) * 2 * 24 * time.Hour
)

type HeartbeatService struct {
	Config *models.Config
	Db     *gorm.DB
}

func NewHeartbeatService(db *gorm.DB) *HeartbeatService {
	return &HeartbeatService{
		Config: models.GetConfig(),
		Db:     db,
	}
}

func (srv *HeartbeatService) InsertBatch(heartbeats []*models.Heartbeat) error {
	var batch []interface{}
	for _, h := range heartbeats {
		batch = append(batch, *h)
	}

	if err := gormbulk.BulkInsert(srv.Db, batch, 3000); err != nil {
		return err
	}
	return nil
}

func (srv *HeartbeatService) GetAllWithin(from, to time.Time, user *models.User) ([]*models.Heartbeat, error) {
	var heartbeats []*models.Heartbeat
	if err := srv.Db.
		Where(&models.Heartbeat{UserID: user.ID}).
		Where("time >= ?", from).
		Where("time <= ?", to).
		Order("time asc").
		Find(&heartbeats).Error; err != nil {
		return nil, err
	}
	return heartbeats, nil
}

// Will return *models.Heartbeat object with only user_id and time fields filled
func (srv *HeartbeatService) GetFirstUserHeartbeats(userIds []string) ([]*models.Heartbeat, error) {
	var heartbeats []*models.Heartbeat
	if err := srv.Db.
		Table("heartbeats").
		Select("user_id, min(time) as time").
		Where("user_id IN (?)", userIds).
		Group("user_id").
		Scan(&heartbeats).Error; err != nil {
		return nil, err
	}
	return heartbeats, nil
}

func (srv *HeartbeatService) DeleteBefore(t time.Time) error {
	if err := srv.Db.
		Where("time <= ?", t).
		Delete(models.Heartbeat{}).Error; err != nil {
		return err
	}
	return nil
}

func (srv *HeartbeatService) CleanUp() error {
	refTime := utils.StartOfToday().Add(-cleanUpInterval)
	if err := srv.DeleteBefore(refTime); err != nil {
		log.Printf("Failed to clean up heartbeats older than %v – %v\n", refTime, err)
		return err
	}
	log.Printf("Successfully cleaned up heartbeats older than %v\n", refTime)
	return nil
}

func (srv *HeartbeatService) ScheduleCleanUp() {
	srv.CleanUp()

	gocron.Every(1).Day().At("02:30").Do(srv.CleanUp)
	<-gocron.Start()
}
