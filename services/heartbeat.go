package services

import (
	"fmt"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/repositories"
	"github.com/muety/wakapi/utils"
	"github.com/patrickmn/go-cache"
	"strings"
	"time"

	"github.com/muety/wakapi/models"
)

type HeartbeatService struct {
	config              *config.Config
	cache               *cache.Cache
	repository          repositories.IHeartbeatRepository
	languageMappingSrvc ILanguageMappingService
}

func NewHeartbeatService(heartbeatRepo repositories.IHeartbeatRepository, languageMappingService ILanguageMappingService) *HeartbeatService {
	return &HeartbeatService{
		config:              config.Get(),
		cache:               cache.New(24*time.Hour, 24*time.Hour),
		repository:          heartbeatRepo,
		languageMappingSrvc: languageMappingService,
	}
}

func (srv *HeartbeatService) Insert(heartbeat *models.Heartbeat) error {
	srv.updateEntityUserCacheByHeartbeat(heartbeat)
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
		srv.updateEntityUserCacheByHeartbeat(hb)
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

func (srv *HeartbeatService) GetEntitySetByUser(entityType uint8, user *models.User) ([]string, error) {
	cacheKey := srv.getEntityUserCacheKey(entityType, user)
	if results, found := srv.cache.Get(cacheKey); found {
		return utils.SetToStrings(results.(map[string]bool)), nil
	}

	results, err := srv.repository.GetEntitySetByUser(entityType, user)
	if err != nil {
		return nil, err
	}

	filtered := make([]string, 0, len(results))
	for _, r := range results {
		if strings.TrimSpace(r) != "" {
			filtered = append(filtered, r)
		}
	}

	srv.cache.Set(cacheKey, utils.StringsToSet(filtered), cache.DefaultExpiration)
	return filtered, nil
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

func (srv *HeartbeatService) getEntityUserCacheKey(entityType uint8, user *models.User) string {
	return fmt.Sprintf("entity_set_%d_%s", entityType, user.ID)
}

func (srv *HeartbeatService) updateEntityUserCache(entityType uint8, entityKey string, user *models.User) {
	cacheKey := srv.getEntityUserCacheKey(entityType, user)
	if entities, found := srv.cache.Get(cacheKey); found {
		if _, ok := entities.(map[string]bool)[entityKey]; !ok {
			// new project / language / ..., which is not yet present in cache, arrived as part of a heartbeats
			// -> invalidate cache
			srv.cache.Delete(cacheKey)
		}
	}
}

func (srv *HeartbeatService) updateEntityUserCacheByHeartbeat(hb *models.Heartbeat) {
	srv.updateEntityUserCache(models.SummaryProject, hb.Project, hb.User)
	srv.updateEntityUserCache(models.SummaryLanguage, hb.Language, hb.User)
	srv.updateEntityUserCache(models.SummaryEditor, hb.Editor, hb.User)
	srv.updateEntityUserCache(models.SummaryOS, hb.OperatingSystem, hb.User)
	srv.updateEntityUserCache(models.SummaryMachine, hb.Machine, hb.User)
}
