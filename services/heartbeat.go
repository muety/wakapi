package services

import (
	"github.com/jasonlvhit/gocron"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/repositories"
	"github.com/muety/wakapi/utils"
	"log"
	"time"

	"github.com/muety/wakapi/models"
)

const (
	cleanUpInterval = time.Duration(aggregateIntervalDays) * 2 * 24 * time.Hour
)

type HeartbeatService struct {
	config              *config.Config
	repository          *repositories.HeartbeatRepository
	languageMappingSrvc *LanguageMappingService
}

func NewHeartbeatService(heartbeatRepo *repositories.HeartbeatRepository, languageMappingService *LanguageMappingService) *HeartbeatService {
	return &HeartbeatService{
		config:              config.Get(),
		repository:          heartbeatRepo,
		languageMappingSrvc: languageMappingService,
	}
}

func (srv *HeartbeatService) InsertBatch(heartbeats []*models.Heartbeat) error {
	return srv.repository.InsertBatch(heartbeats)
}

func (srv *HeartbeatService) GetAllWithin(from, to time.Time, user *models.User) ([]*models.Heartbeat, error) {
	heartbeats, err := srv.repository.GetAllWithin(from, to, user)
	if err != nil {
		return nil, err
	}
	return srv.augmented(heartbeats, user.ID)
}

func (srv *HeartbeatService) GetFirstUserHeartbeats(userIds []string) ([]*models.Heartbeat, error) {
	return srv.repository.GetFirstByUsers(userIds)
}

func (srv *HeartbeatService) DeleteBefore(t time.Time) error {
	return srv.repository.DeleteBefore(t)
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

func (srv *HeartbeatService) augmented(heartbeats []*models.Heartbeat, userId string) ([]*models.Heartbeat, error) {
	languageMapping, err := srv.languageMappingSrvc.ResolveByUser(userId)
	if err != nil {
		return nil, err
	}

	for i := range heartbeats {
		heartbeats[i].Augment(languageMapping)
	}

	return heartbeats, nil
}
