package services

import (
	"fmt"
	"github.com/leandro-lugaresi/hub"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/repositories"
	"github.com/muety/wakapi/utils"
	"github.com/patrickmn/go-cache"
	"strings"
	"sync"
	"time"

	"github.com/muety/wakapi/models"
)

type HeartbeatService struct {
	config              *config.Config
	cache               *cache.Cache
	eventBus            *hub.Hub
	repository          repositories.IHeartbeatRepository
	languageMappingSrvc ILanguageMappingService
	entityCacheLock     *sync.RWMutex
}

func NewHeartbeatService(heartbeatRepo repositories.IHeartbeatRepository, languageMappingService ILanguageMappingService) *HeartbeatService {
	srv := &HeartbeatService{
		config:              config.Get(),
		cache:               cache.New(24*time.Hour, 24*time.Hour),
		eventBus:            config.EventBus(),
		repository:          heartbeatRepo,
		languageMappingSrvc: languageMappingService,
		entityCacheLock:     &sync.RWMutex{},
	}

	// using event hub is an unnecessary indirection here, however, we might
	// potentially need heartbeat events elsewhere throughout the application some time
	// so it's more consistent to already have it this way
	sub1 := srv.eventBus.Subscribe(0, config.EventHeartbeatCreate)
	go func(sub *hub.Subscription) {
		for m := range sub.Receiver {
			heartbeat := m.Fields[config.FieldPayload].(*models.Heartbeat)
			srv.cache.IncrementInt64(srv.countByUserCacheKey(heartbeat.UserID), 1) // increment doesn't update expiration time
			srv.cache.IncrementInt64(srv.countTotalCacheKey(), 1)
		}
	}(&sub1)

	return srv
}

func (srv *HeartbeatService) Insert(heartbeat *models.Heartbeat) error {
	go srv.updateEntityUserCacheByHeartbeat(heartbeat)
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
		go srv.updateEntityUserCacheByHeartbeat(hb)
	}

	err := srv.repository.InsertBatch(filteredHeartbeats)
	if err == nil {
		go srv.notifyBatch(filteredHeartbeats)
	}
	return err
}

func (srv *HeartbeatService) Count() (int64, error) {
	result, ok := srv.cache.Get(srv.countTotalCacheKey())
	if ok {
		return result.(int64), nil
	}
	count, err := srv.repository.Count()
	if err == nil {
		srv.cache.Set(srv.countTotalCacheKey(), count, srv.countCacheTtl())
	}
	return count, err
}

func (srv *HeartbeatService) CountByUser(user *models.User) (int64, error) {
	key := srv.countByUserCacheKey(user.ID)
	result, ok := srv.cache.Get(key)
	if ok {
		return result.(int64), nil
	}
	count, err := srv.repository.CountByUser(user)
	if err == nil {
		srv.cache.Set(key, count, srv.countCacheTtl())
	}
	return count, err
}

func (srv *HeartbeatService) CountByUsers(users []*models.User) ([]*models.CountByUser, error) {
	missingUsers := make([]*models.User, 0, len(users))
	userCounts := make([]*models.CountByUser, 0, len(users))

	for _, u := range users {
		key := srv.countByUserCacheKey(u.ID)
		result, ok := srv.cache.Get(key)
		if ok {
			userCounts = append(userCounts, &models.CountByUser{User: u.ID, Count: result.(int64)})
		} else {
			missingUsers = append(missingUsers, u)
		}
	}

	counts, err := srv.repository.CountByUsers(missingUsers)
	if err != nil {
		return nil, err
	}

	for _, uc := range counts {
		key := srv.countByUserCacheKey(uc.User)
		srv.cache.Set(key, uc.Count, srv.countCacheTtl())
		userCounts = append(userCounts, uc)
	}

	return userCounts, nil
}

func (srv *HeartbeatService) GetAllWithin(from, to time.Time, user *models.User) ([]*models.Heartbeat, error) {
	heartbeats, err := srv.repository.GetAllWithin(from, to, user)
	if err != nil {
		return nil, err
	}
	return srv.augmented(heartbeats, user.ID)
}

func (srv *HeartbeatService) GetAllWithinByFilters(from, to time.Time, user *models.User, filters *models.Filters) ([]*models.Heartbeat, error) {
	heartbeats, err := srv.repository.GetAllWithinByFilters(from, to, user, srv.filtersToColumnMap(filters))
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
		srv.entityCacheLock.RLock()
		defer srv.entityCacheLock.RUnlock()
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

	srv.cache.Set(cacheKey, utils.StringsToSet(filtered), cache.NoExpiration)
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
		entitySet := entities.(map[string]bool)

		srv.entityCacheLock.Lock()
		defer srv.entityCacheLock.Unlock()

		if _, ok := entitySet[entityKey]; !ok {
			entitySet[entityKey] = true
			// new project / language / ..., which is not yet present in cache, arrived as part of a heartbeats
			// -> update cache instead of just invalidating it, because rebuilding is expensive here
			srv.cache.Set(cacheKey, entitySet, cache.NoExpiration)
		}
	}
}

func (srv *HeartbeatService) updateEntityUserCacheByHeartbeat(hb *models.Heartbeat) {
	go srv.updateEntityUserCache(models.SummaryProject, hb.Project, hb.User)
	go srv.updateEntityUserCache(models.SummaryLanguage, hb.Language, hb.User)
	go srv.updateEntityUserCache(models.SummaryEditor, hb.Editor, hb.User)
	go srv.updateEntityUserCache(models.SummaryOS, hb.OperatingSystem, hb.User)
	go srv.updateEntityUserCache(models.SummaryMachine, hb.Machine, hb.User)
	go srv.updateEntityUserCache(models.SummaryBranch, hb.Branch, hb.User)
}

func (srv *HeartbeatService) notifyBatch(heartbeats []*models.Heartbeat) {
	for _, hb := range heartbeats {
		srv.eventBus.Publish(hub.Message{
			Name:   config.EventHeartbeatCreate,
			Fields: map[string]interface{}{config.FieldPayload: hb},
		})
	}
}

func (srv *HeartbeatService) countByUserCacheKey(userId string) string {
	return fmt.Sprintf("%s--hearbeat-count", userId)
}

func (srv *HeartbeatService) countTotalCacheKey() string {
	return "heartbeat-count"
}

func (srv *HeartbeatService) countCacheTtl() time.Duration {
	return time.Duration(srv.config.App.CountCacheTTLMin) * time.Minute
}

func (srv *HeartbeatService) filtersToColumnMap(filters *models.Filters) map[string][]string {
	columnMap := map[string][]string{}
	for _, t := range models.SummaryTypes() {
		f := filters.ResolveEntity(t)
		if len(*f) > 0 {
			columnMap[models.GetEntityColumn(t)] = *f
		}
	}
	return columnMap
}
