package services

import (
	"fmt"
	"strings"
	"sync"
	"time"

	datastructure "github.com/duke-git/lancet/v2/datastructure/set"
	"github.com/leandro-lugaresi/hub"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/repositories"
	"github.com/patrickmn/go-cache"

	"github.com/muety/wakapi/models"
)

type HeartbeatService struct {
	config              *config.Config
	cache               *cache.Cache
	eventBus            *hub.Hub
	repository          repositories.IHeartbeatRepository
	languageMappingSrvc ILanguageMappingService
	cacheLock           *sync.RWMutex
}

func NewHeartbeatService(heartbeatRepo repositories.IHeartbeatRepository, languageMappingService ILanguageMappingService) *HeartbeatService {
	srv := &HeartbeatService{
		config:              config.Get(),
		cache:               cache.New(24*time.Hour, 24*time.Hour),
		eventBus:            config.EventBus(),
		repository:          heartbeatRepo,
		languageMappingSrvc: languageMappingService,
		cacheLock:           &sync.RWMutex{}, // if this turns out as a bottleneck, we might switch to per-user caches organized in a map
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
			srv.checkInvalidateRangeCache(heartbeat)
			srv.updateLatestHeartbeatCache(heartbeat)
		}
	}(&sub1)

	return srv
}

func (srv *HeartbeatService) Insert(heartbeat *models.Heartbeat) error {
	go srv.updateEntityUserCacheByHeartbeat(heartbeat)
	return srv.repository.InsertBatch([]*models.Heartbeat{heartbeat})
}

func (srv *HeartbeatService) InsertBatch(heartbeats []*models.Heartbeat) error {
	if len(heartbeats) == 0 {
		return nil
	}

	hashes := datastructure.New[string]()

	// https://github.com/muety/wakapi/issues/139
	filteredHeartbeats := make([]*models.Heartbeat, 0, len(heartbeats))
	for _, hb := range heartbeats {
		if !hashes.Contain(hb.Hash) {
			hb = hb.Sanitize()
			filteredHeartbeats = append(filteredHeartbeats, hb)
			hashes.Add(hb.Hash)
		}
		go srv.updateEntityUserCacheByHeartbeat(hb)
	}

	err := srv.repository.InsertBatch(filteredHeartbeats)
	if err == nil {
		go srv.notifyBatch(filteredHeartbeats)
	}
	return err
}

func (srv *HeartbeatService) Count(approximate bool) (int64, error) {
	result, ok := srv.cache.Get(srv.countTotalCacheKey())
	if ok {
		return result.(int64), nil
	}
	count, err := srv.repository.Count(approximate)
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
	heartbeats, err := srv.repository.GetWithin(from, to, user)
	if err != nil {
		return nil, err
	}
	return srv.augmented(heartbeats, user.ID)
}

func (srv *HeartbeatService) StreamAllWithin(from, to time.Time, user *models.User) (chan *models.Heartbeat, error) {
	languageMapping, err := srv.languageMappingSrvc.ResolveByUser(user.ID)
	if err != nil {
		return nil, err
	}

	c, err := srv.repository.StreamWithin(from, to, user)
	if err != nil {
		return nil, err
	}
	return srv.augmentedAsync(c, languageMapping)
}

func (srv *HeartbeatService) StreamAllWithinRaw(from, to time.Time, user *models.User) (chan *models.Heartbeat, error) {
	return srv.repository.StreamWithin(from, to, user) // no augmentation
}

// StreamAllWithinExcludingHeartbeats streams heartbeats within the given interval without augmenting them,
// excluding heartbeats matching any of the given exclusions (see models.ExcludeFromDurations).
func (srv *HeartbeatService) StreamAllWithinExcludingHeartbeats(from, to time.Time, user *models.User, exclusions []models.HeartbeatExclusionFilter) (chan *models.Heartbeat, error) {
	return srv.repository.StreamWithinExcludingHeartbeats(from, to, user, exclusions) // no augmentation
}

func (srv *HeartbeatService) GetAllWithinByFilters(from, to time.Time, user *models.User, filters *models.Filters) ([]*models.Heartbeat, error) {
	heartbeats, err := srv.repository.GetAllWithinByFilters(from, to, user, srv.filtersToColumnMap(filters))
	if err != nil {
		return nil, err
	}
	return srv.augmented(heartbeats, user.ID)
}

func (srv *HeartbeatService) StreamAllWithinByFilters(from, to time.Time, user *models.User, filters *models.Filters) (chan *models.Heartbeat, error) {
	languageMapping, err := srv.languageMappingSrvc.ResolveByUser(user.ID)
	if err != nil {
		return nil, err
	}

	c, err := srv.repository.StreamWithinByFilters(from, to, user, srv.filtersToColumnMap(filters))
	if err != nil {
		return nil, err
	}
	return srv.augmentedAsync(c, languageMapping)
}

func (srv *HeartbeatService) GetLatestByUser(user *models.User, skipCache bool) (*models.Heartbeat, error) {
	cacheKey := srv.latestByUserCacheKey(user.ID)
	if !skipCache {
		if result, found := srv.cache.Get(cacheKey); found {
			srv.cacheLock.RLock()
			defer srv.cacheLock.RUnlock()
			return result.(*models.Heartbeat), nil
		}
	}

	heartbeat, err := srv.repository.GetLatestByUser(user)
	if err != nil {
		return nil, err
	}

	// cache even when no row was found
	// such an entry is treated as time-zero by refreshLatestCacheEntry and hence upgraded automatically once a heartbeat arrives
	srv.cacheLock.Lock()
	defer srv.cacheLock.Unlock()
	srv.cache.SetDefault(cacheKey, cloneHeartbeat(heartbeat))
	return heartbeat, nil
}

func (srv *HeartbeatService) GetLatestByOriginAndUser(origin string, user *models.User) (*models.Heartbeat, error) {
	return srv.repository.GetLatestByOriginAndUser(origin, user)
}

func (srv *HeartbeatService) GetLatestByFilters(user *models.User, filters *models.Filters, skipCache bool) (*models.Heartbeat, error) {
	columnMap := srv.filtersToColumnMap(filters)

	// cached because see above (GetLatestByUser); cache misses are served by idx_user_project_time
	projectFilter, isSingleProjectFilter := isSingleProjectFilter(columnMap)
	cacheKey := srv.latestByProjectCacheKey(user.ID, projectFilter)

	if isSingleProjectFilter && !skipCache {
		if result, found := srv.cache.Get(cacheKey); found {
			srv.cacheLock.RLock()
			defer srv.cacheLock.RUnlock()
			return result.(*models.Heartbeat), nil
		}
	}

	heartbeat, err := srv.repository.GetLatestByFilters(user, columnMap)
	if err != nil {
		return nil, err
	}

	if isSingleProjectFilter {
		srv.cacheLock.Lock()
		defer srv.cacheLock.Unlock()
		if heartbeat != nil {
			srv.cache.SetDefault(cacheKey, cloneHeartbeat(heartbeat))
		} else {
			// cache misses (typed-nil sentinel) as well
			// the sentinel is treated as time-zero by refreshLatestCacheEntry and hence upgraded automatically once a heartbeat for that project arrives
			srv.cache.SetDefault(cacheKey, heartbeat)
		}
	}
	return heartbeat, nil
}

func (srv *HeartbeatService) GetFirstAll() ([]*models.TimeByUser, error) {
	return srv.repository.GetFirstAll()
}

func (srv *HeartbeatService) GetLastAll() ([]*models.TimeByUser, error) {
	return srv.repository.GetLastAll()
}

func (srv *HeartbeatService) GetFirstByUser(user *models.User) (time.Time, error) {
	cacheKey := srv.getUserFirstCacheKey(user.ID)
	if result, found := srv.cache.Get(cacheKey); found {
		return result.(time.Time), nil
	}

	result, err := srv.repository.GetRangeByUser(user)
	if err != nil {
		return time.Time{}, err
	}

	srv.cache.Set(cacheKey, result.First.T(), cache.NoExpiration)
	return result.First.T(), nil
}

func (srv *HeartbeatService) GetLastByUser(user *models.User) (time.Time, error) {
	cacheKey := srv.getUserLastCacheKey(user.ID)
	if result, found := srv.cache.Get(cacheKey); found {
		return result.(time.Time), nil
	}

	result, err := srv.repository.GetRangeByUser(user)
	if err != nil {
		return time.Time{}, err
	}

	srv.cache.SetDefault(cacheKey, result.Last.T())
	return result.Last.T(), nil
}

func (srv *HeartbeatService) GetRangeByUser(user *models.User) (*models.RangeByUser, error) {
	return srv.repository.GetRangeByUser(user)
}

func (srv *HeartbeatService) GetEntitySetByUser(entityType uint8, userId string) ([]string, error) {
	cacheKey := srv.getEntityUserCacheKey(entityType, userId)
	if results, found := srv.cache.Get(cacheKey); found {
		srv.cacheLock.RLock()
		defer srv.cacheLock.RUnlock()
		return results.(datastructure.Set[string]).Values(), nil
	}

	results, err := srv.repository.GetEntitySetByUser(entityType, userId)
	if err != nil {
		return nil, err
	}

	filtered := make([]string, 0, len(results))
	for _, r := range results {
		if strings.TrimSpace(r) != "" {
			filtered = append(filtered, r)
		}
	}

	srv.cache.SetDefault(cacheKey, datastructure.New(filtered...))
	return filtered, nil
}

func (srv *HeartbeatService) DeleteBefore(t time.Time) error {
	go srv.cache.Flush()
	return srv.repository.DeleteBefore(t)
}

func (srv *HeartbeatService) DeleteByUser(user *models.User) error {
	go srv.cache.Flush()
	return srv.repository.DeleteByUser(user)
}

func (srv *HeartbeatService) DeleteByUserBefore(user *models.User, t time.Time) error {
	go srv.cache.Flush()
	return srv.repository.DeleteByUserBefore(user, t)
}

// GetUserAgentsByUser returns a list of all user agents that have been recorded for the given user.
func (srv *HeartbeatService) GetUserAgentsByUser(user *models.User) ([]*models.UserAgent, error) {
	userAgents, err := srv.repository.GetUserAgentsByUser(user)
	if err != nil {
		return nil, err
	}
	for _, ua := range userAgents {
		ua.WithId()
	}
	return userAgents, nil
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

func (srv *HeartbeatService) augmentedAsync(in chan *models.Heartbeat, languageMapping map[string]string) (chan *models.Heartbeat, error) {
	// if this method made the query to fetch langauge mapping itself, it would produce a dead loop in case there are less than 2 database connections
	out := make(chan *models.Heartbeat)
	go func(in, out chan *models.Heartbeat) {
		defer close(out)
		for hb := range in {
			hb.Augment(languageMapping)
			out <- hb
		}
	}(in, out)
	return out, nil
}

func (srv *HeartbeatService) getEntityUserCacheKey(entityType uint8, userId string) string {
	return fmt.Sprintf("entity_set_%d_%s", entityType, userId)
}

func (srv *HeartbeatService) getUserFirstCacheKey(userId string) string {
	return fmt.Sprintf("user_first_%s", userId)
}

func (srv *HeartbeatService) getUserLastCacheKey(userId string) string {
	return fmt.Sprintf("user_last_%s", userId)
}

func (srv *HeartbeatService) latestByUserCacheKey(userId string) string {
	return fmt.Sprintf("latest_heartbeat_user_%s", userId)
}

func (srv *HeartbeatService) latestByProjectCacheKey(userId, project string) string {
	return fmt.Sprintf("latest_heartbeat_project_%s_%s", userId, project)
}

func (srv *HeartbeatService) updateEntityUserCache(entityType uint8, entityKey string, userId string) {
	cacheKey := srv.getEntityUserCacheKey(entityType, userId)
	if entities, found := srv.cache.Get(cacheKey); found {
		entitySet := entities.(datastructure.Set[string])

		srv.cacheLock.Lock()
		defer srv.cacheLock.Unlock()

		if !entitySet.Contain(entityKey) {
			entitySet.Add(entityKey)
			// new project / language / ..., which is not yet present in cache, arrived as part of a heartbeats
			// -> update cache instead of just invalidating it, because rebuilding is expensive here
			srv.cache.Set(cacheKey, entitySet, cache.NoExpiration)
		}
	}
}

func (srv *HeartbeatService) updateEntityUserCacheByHeartbeat(hb *models.Heartbeat) {
	go srv.updateEntityUserCache(models.SummaryProject, hb.Project, hb.UserID)
	go srv.updateEntityUserCache(models.SummaryLanguage, hb.Language, hb.UserID)
	go srv.updateEntityUserCache(models.SummaryEditor, hb.Editor, hb.UserID)
	go srv.updateEntityUserCache(models.SummaryOS, hb.OperatingSystem, hb.UserID)
	go srv.updateEntityUserCache(models.SummaryMachine, hb.Machine, hb.UserID)
	go srv.updateEntityUserCache(models.SummaryBranch, hb.Branch, hb.UserID)
	go srv.updateEntityUserCache(models.SummaryEntity, hb.Entity, hb.UserID)
	go srv.updateEntityUserCache(models.SummaryCategory, hb.Category, hb.UserID)
	go srv.updateEntityUserCache(models.SummaryAiModel, hb.AIModel, hb.UserID)
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
	for _, t := range models.NativeSummaryTypes() {
		f := filters.ResolveType(t)
		if len(*f) > 0 {
			columnMap[models.GetEntityColumn(t)] = *f
		}
	}
	return columnMap
}

func (srv *HeartbeatService) checkInvalidateRangeCache(newHeartbeat *models.Heartbeat) {
	keyFirst, keyLast := srv.getUserFirstCacheKey(newHeartbeat.UserID), srv.getUserLastCacheKey(newHeartbeat.UserID)

	first, found := srv.cache.Get(keyFirst)
	if found && newHeartbeat.Time.T().Before(first.(time.Time)) {
		srv.cache.Delete(keyFirst)
	}

	last, found := srv.cache.Get(keyLast)
	if found && newHeartbeat.Time.T().After(last.(time.Time)) {
		srv.cache.Delete(keyLast)
	}
}

// updateLatestHeartbeatCache keeps the latest heartbeat-cache entries along the user and (user, project) dimensions up to date, analogous to how checkInvalidateRangeCache
// like updateEntityUserCache, it only refreshes entries that are already populated, rather than eagerly creating one for every heartbeat seen
func (srv *HeartbeatService) updateLatestHeartbeatCache(newHeartbeat *models.Heartbeat) {
	srv.refreshLatestCacheEntry(srv.latestByUserCacheKey(newHeartbeat.UserID), newHeartbeat)
	srv.refreshLatestCacheEntry(srv.latestByProjectCacheKey(newHeartbeat.UserID, newHeartbeat.Project), newHeartbeat)
}

func (srv *HeartbeatService) refreshLatestCacheEntry(key string, newHeartbeat *models.Heartbeat) {
	srv.cacheLock.Lock() // write lock throughout, as this is a read-modify-write operation
	defer srv.cacheLock.Unlock()

	cached, found := srv.cache.Get(key)
	if !found {
		return // intentionally don't pre-fill cache from incoming heartbeats, because we can't be sure there are no newer ones in the databaes
	}
	// a cached nil pointer (or zero-value struct) indicates "known to have no heartbeats" and is effectively time-zero, i.e. always replaced
	if latest, ok := cached.(*models.Heartbeat); ok && (latest == nil || newHeartbeat.Time.T().After(latest.Time.T())) {
		srv.cache.SetDefault(key, cloneHeartbeat(newHeartbeat))
	}
}

// isSingleProjectFilter tells whether the given repository filter map consists of exactly one entry, which is a single project value
func isSingleProjectFilter(columnMap map[string][]string) (string, bool) {
	if len(columnMap) != 1 {
		return "", false
	}
	values, isProject := columnMap[models.GetEntityColumn(models.SummaryProject)]
	if !isProject || len(values) != 1 {
		return "", false
	}
	return values[0], true
}

func cloneHeartbeat(heartbeat *models.Heartbeat) *models.Heartbeat {
	clone := *heartbeat
	return &clone
}
