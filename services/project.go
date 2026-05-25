package services

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	datastructure "github.com/duke-git/lancet/v2/datastructure/set"
	"github.com/duke-git/lancet/v2/maputil"
	"github.com/leandro-lugaresi/hub"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
	"github.com/muety/wakapi/utils"
	"github.com/patrickmn/go-cache"
)

type ProjectService struct {
	config        *config.Config
	cache         *cache.Cache
	eventBus      *hub.Hub
	repository    repositories.IHeartbeatRepository
	aliasService  IAliasService
	heartbeatSrvc IHeartbeatService
}

func NewProjectService(aliasService IAliasService, heartbeatRepo repositories.IHeartbeatRepository, heartbeatSrvc IHeartbeatService) *ProjectService {
	srv := &ProjectService{
		config:        config.Get(),
		cache:         cache.New(24*time.Hour, 24*time.Hour),
		eventBus:      config.EventBus(),
		repository:    heartbeatRepo,
		aliasService:  aliasService,
		heartbeatSrvc: heartbeatSrvc,
	}

	sub1 := srv.eventBus.Subscribe(0, config.EventHeartbeatCreate)
	go func(sub *hub.Subscription) {
		for m := range sub.Receiver {
			heartbeat := m.Fields[config.FieldPayload].(*models.Heartbeat)
			srv.checkInvalidateProjectStatsCache(heartbeat)
		}
	}(&sub1)

	sub2 := srv.eventBus.Subscribe(0, config.TopicAlias)
	go func(sub *hub.Subscription) {
		for m := range sub.Receiver {
			userId := m.Fields[config.FieldUserId].(string)
			srv.invalidateProjectStatsCache(userId)
		}
	}(&sub2)

	return srv
}

func (srv *ProjectService) GetUserProjectStats(user *models.User, from, to time.Time, search string, pageParams *utils.PageParams, skipCache bool) ([]*models.ProjectStats, error) {
	var (
		limit  = math.MaxInt32
		offset = 0
	)

	if pageParams != nil {
		limit = pageParams.Limit()
		offset = pageParams.Offset()
	}

	cacheKey := fmt.Sprintf("project_stats_%s_%d_%d_%d_%d_%s", user.ID, from.Unix(), to.Unix(), limit, offset, search)
	if results, found := srv.cache.Get(cacheKey); found && !skipCache {
		return results.([]*models.ProjectStats), nil
	} else if search == "" {
		if results, found := srv.cache.Get(fmt.Sprintf("project_stats_%s_%d_%d_%d_%d_", user.ID, from.Unix(), to.Unix(), math.MaxInt32, 0)); found && !skipCache {
			return utils.SubSlice[*models.ProjectStats](results.([]*models.ProjectStats), uint(offset), uint(offset+limit)), nil
		}
	}

	if to.IsZero() {
		to = time.Now()
	}

	rawResults, err := srv.repository.GetUserProjectStats(user, from, to)
	if err != nil {
		return nil, err
	}

	merged := make(map[string]*models.ProjectStats)
	maxCounts := make(map[string]int64)

	for _, stats := range rawResults {
		aliasKey, _ := srv.aliasService.GetAliasOrDefault(user.ID, models.SummaryProject, stats.Project)

		if existing, ok := merged[aliasKey]; ok {
			existing.Count += stats.Count

			if stats.First.T().Before(existing.First.T()) {
				existing.First = stats.First
			}

			if stats.Last.T().After(existing.Last.T()) {
				existing.Last = stats.Last
			}

			if stats.Count > maxCounts[aliasKey] {
				existing.TopLanguage = stats.TopLanguage
				maxCounts[aliasKey] = stats.Count
			}
		} else {
			merged[aliasKey] = &models.ProjectStats{
				UserId:      stats.UserId,
				Project:     aliasKey,
				TopLanguage: stats.TopLanguage,
				Count:       stats.Count,
				First:       stats.First,
				Last:        stats.Last,
			}
			maxCounts[aliasKey] = stats.Count
		}
	}

	aggregatedResults := make([]*models.ProjectStats, 0, len(merged))
	for _, v := range merged {
		if search == "" || strings.Contains(strings.ToLower(v.Project), strings.ToLower(search)) {
			aggregatedResults = append(aggregatedResults, v)
		}
	}

	sort.Slice(aggregatedResults, func(i, j int) bool {
		return aggregatedResults[i].Last.T().After(aggregatedResults[j].Last.T())
	})

	paginatedResults := utils.SubSlice[*models.ProjectStats](aggregatedResults, uint(offset), uint(offset+limit))

	srv.cache.Set(cacheKey, paginatedResults, 12*time.Hour)
	if search == "" && (limit != math.MaxInt32 || offset != 0) {
		srv.cache.Set(fmt.Sprintf("project_stats_%s_%d_%d_%d_%d_", user.ID, from.Unix(), to.Unix(), math.MaxInt32, 0), aggregatedResults, 12*time.Hour)
	}

	go srv.populateUniqueUserProjects(user.ID)

	return paginatedResults, nil
}

func (srv *ProjectService) populateUniqueUserProjects(userId string) {
	userProjectsCacheKey := srv.getUserProjectsCacheKey(userId)
	if _, found := srv.cache.Get(userProjectsCacheKey); !found {
		projects, _ := srv.heartbeatSrvc.GetEntitySetByUser(models.SummaryProject, userId)
		srv.cache.Set(userProjectsCacheKey, datastructure.New[string](projects...), cache.NoExpiration)
	}
}

func (srv *ProjectService) invalidateProjectStatsCache(userId string) {
	var invalidated bool
	for _, k := range maputil.Keys[string, cache.Item](srv.cache.Items()) {
		if strings.HasPrefix(k, fmt.Sprintf("project_stats_%s_", userId)) {
			srv.cache.Delete(k)
			invalidated = true
		}
	}
	if invalidated {
		srv.cache.Delete(srv.getUserProjectsCacheKey(userId))
	}
}

func (srv *ProjectService) checkInvalidateProjectStatsCache(newHeartbeat *models.Heartbeat) {
	if uniqueProjects, found := srv.cache.Get(srv.getUserProjectsCacheKey(newHeartbeat.UserID)); found && !uniqueProjects.(datastructure.Set[string]).Contain(newHeartbeat.Project) {
		srv.invalidateProjectStatsCache(newHeartbeat.UserID)
	}
}

func (srv *ProjectService) getUserProjectsCacheKey(userId string) string {
	return fmt.Sprintf("unique_projects_%s", userId)
}
