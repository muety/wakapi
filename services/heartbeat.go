package services

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/repositories"
	"time"

	"github.com/muety/wakapi/models"
)

type HeartbeatService struct {
	config              *config.Config
	repository          repositories.IHeartbeatRepository
	languageMappingSrvc ILanguageMappingService
}

func NewHeartbeatService(heartbeatRepo repositories.IHeartbeatRepository, languageMappingService ILanguageMappingService) *HeartbeatService {
	return &HeartbeatService{
		config:              config.Get(),
		repository:          heartbeatRepo,
		languageMappingSrvc: languageMappingService,
	}
}

func (srv *HeartbeatService) Insert(heartbeat *models.Heartbeat) error {
	return srv.repository.InsertBatch([]*models.Heartbeat{heartbeat})
}

func (srv *HeartbeatService) InsertBatch(heartbeats []*models.Heartbeat) error {
	hashes := make(map[string]bool)

	// https://github.com/muety/wakapi/issues/139
	filteredHeartbeats := make([]*models.Heartbeat, 0, len(heartbeats))
	for _, hb := range heartbeats {
		if _, ok := hashes[hb.Hash]; !ok {
			filteredHeartbeats = append(filteredHeartbeats, hb)
			hashes[hb.Hash] = true
		}
	}

	return srv.repository.InsertBatch(filteredHeartbeats)
}

func (srv *HeartbeatService) Count() (int64, error) {
	return srv.repository.Count()
}

func (srv *HeartbeatService) CountByUser(user *models.User) (int64, error) {
	return srv.repository.CountByUser(user)
}

func (srv *HeartbeatService) CountByUsers(users []*models.User) ([]*models.CountByUser, error) {
	return srv.repository.CountByUsers(users)
}

func (srv *HeartbeatService) GetAllWithin(from, to time.Time, user *models.User) ([]*models.Heartbeat, error) {
	heartbeats, err := srv.repository.GetAllWithin(from, to, user)
	if err != nil {
		return nil, err
	}
	return srv.augmented(heartbeats, user.ID)
}

func (srv *HeartbeatService) GetLatestByUser(user *models.User) (*models.Heartbeat, error) {
	return srv.repository.GetLatestByUser(user)
}

func (srv *HeartbeatService) GetLatestByOriginAndUser(origin string, user *models.User) (*models.Heartbeat, error) {
	return srv.repository.GetLatestByOriginAndUser(origin, user)
}

func (srv *HeartbeatService) GetFirstByUsers() ([]*models.TimeByUser, error) {
	return srv.repository.GetFirstByUsers()
}

func (srv *HeartbeatService) DeleteBefore(t time.Time) error {
	return srv.repository.DeleteBefore(t)
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
