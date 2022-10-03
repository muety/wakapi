package services

import (
	"github.com/emvi/logbuch"
	"github.com/go-co-op/gocron"
	"github.com/leandro-lugaresi/hub"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
	"github.com/muety/wakapi/utils"
	"github.com/patrickmn/go-cache"
	"reflect"
	"strings"
	"time"
)

type LeaderboardService struct {
	config         *config.Config
	cache          *cache.Cache
	eventBus       *hub.Hub
	repository     repositories.ILeaderboardRepository
	summaryService ISummaryService
	userService    IUserService
}

func NewLeaderboardService(leaderboardRepo repositories.ILeaderboardRepository, summaryService ISummaryService, userService IUserService) *LeaderboardService {
	srv := &LeaderboardService{
		config:         config.Get(),
		cache:          cache.New(6*time.Hour, 6*time.Hour),
		eventBus:       config.EventBus(),
		repository:     leaderboardRepo,
		summaryService: summaryService,
		userService:    userService,
	}

	onUserUpdate := srv.eventBus.Subscribe(0, config.EventUserUpdate)
	go func(sub *hub.Subscription) {
		for m := range sub.Receiver {

			// generate leaderboard for updated user, if leaderboard enabled and none present, yet
			user := m.Fields[config.FieldPayload].(*models.User)
			if user.PublicLeaderboard {
				exists, err := srv.ExistsAnyByUser(user.ID)
				if err != nil {
					config.Log().Error("failed to check existing leaderboards upon user update - %v", err)
				}
				if !exists {
					logbuch.Info("generating leaderboard for '%s' after settings update", user.ID)
					srv.Run([]*models.User{user}, models.IntervalPast7Days, []uint8{models.SummaryLanguage})
				}
			}

		}
	}(&onUserUpdate)

	return srv
}

func (srv *LeaderboardService) ScheduleDefault() {
	runAllUsers := func(interval *models.IntervalKey, by []uint8) {
		users, err := srv.userService.GetAllByLeaderboard(true)
		if err != nil {
			config.Log().Error("failed to get users for leaderboard generation - %v", err)
			return
		}

		srv.Run(users, interval, by)
	}

	s := gocron.NewScheduler(time.Local)
	s.Every(1).Day().At(srv.config.App.LeaderboardGenerationTime).Do(runAllUsers, models.IntervalPast7Days, []uint8{models.SummaryLanguage})
	s.StartBlocking()
}

func (srv *LeaderboardService) Run(users []*models.User, interval *models.IntervalKey, by []uint8) error {
	logbuch.Info("generating leaderboard (%s) for %d users (%d aggregations)", (*interval)[0], len(users), len(by))

	for _, user := range users {
		if err := srv.repository.DeleteByUserAndInterval(user.ID, interval); err != nil {
			config.Log().Error("failed to delete leaderboard items for user %s (interval %s) - %v", user.ID, (*interval)[0], err)
			continue
		}

		item, err := srv.GenerateByUser(user, interval)
		if err != nil {
			config.Log().Error("failed to generate general leaderboard for user %s - %v", user.ID, err)
			continue
		}

		if err := srv.repository.InsertBatch([]*models.LeaderboardItem{item}); err != nil {
			config.Log().Error("failed to persist general leaderboard for user %s - %v", user.ID, err)
			continue
		}

		for _, by := range by {
			items, err := srv.GenerateAggregatedByUser(user, interval, by)
			if err != nil {
				config.Log().Error("failed to generate aggregated (by %s) leaderboard for user %s - %v", models.GetEntityColumn(by), user.ID, err)
				continue
			}

			if len(items) == 0 {
				continue
			}

			if err := srv.repository.InsertBatch(items); err != nil {
				config.Log().Error("failed to persist aggregated (by %s) leaderboard for user %s - %v", models.GetEntityColumn(by), user.ID, err)
				continue
			}
		}
	}

	srv.cache.Flush()
	logbuch.Info("finished leaderboard generation")
	return nil
}

func (srv *LeaderboardService) ExistsAnyByUser(userId string) (bool, error) {
	count, err := srv.repository.CountAllByUser(userId)
	return count > 0, err
}

func (srv *LeaderboardService) GetByInterval(interval *models.IntervalKey, resolveUsers bool) (models.Leaderboard, error) {
	return srv.GetAggregatedByInterval(interval, nil, resolveUsers)
}

func (srv *LeaderboardService) GetAggregatedByInterval(interval *models.IntervalKey, by *uint8, resolveUsers bool) (models.Leaderboard, error) {
	// check cache
	cacheKey := srv.getHash(interval, by)
	if cacheResult, ok := srv.cache.Get(cacheKey); ok {
		return cacheResult.([]*models.LeaderboardItem), nil
	}

	items, err := srv.repository.GetAllAggregatedByInterval(interval, by)
	if err != nil {
		return nil, err
	}

	if resolveUsers {
		a := models.Leaderboard(items).UserIDs()
		println(a)
		users, err := srv.userService.GetManyMapped(models.Leaderboard(items).UserIDs())
		if err != nil {
			config.Log().Error("failed to resolve users for leaderboard item - %v", err)
		} else {
			for _, item := range items {
				if u, ok := users[item.UserID]; ok {
					item.User = u
				}
			}
		}
	}

	srv.cache.SetDefault(cacheKey, items)
	return items, nil
}

func (srv *LeaderboardService) GenerateByUser(user *models.User, interval *models.IntervalKey) (*models.LeaderboardItem, error) {
	err, from, to := utils.ResolveIntervalTZ(interval, user.TZ())
	if err != nil {
		return nil, err
	}

	summary, err := srv.summaryService.Aliased(from, to, user, srv.summaryService.Retrieve, nil, false)
	if err != nil {
		return nil, err
	}

	return &models.LeaderboardItem{
		User:     user,
		UserID:   user.ID,
		Interval: (*interval)[0],
		Total:    summary.TotalTime(),
	}, nil
}

func (srv *LeaderboardService) GenerateAggregatedByUser(user *models.User, interval *models.IntervalKey, by uint8) ([]*models.LeaderboardItem, error) {
	err, from, to := utils.ResolveIntervalTZ(interval, user.TZ())
	if err != nil {
		return nil, err
	}

	summary, err := srv.summaryService.Aliased(from, to, user, srv.summaryService.Retrieve, nil, false)
	if err != nil {
		return nil, err
	}

	summaryItems := *summary.ItemsByType(by)
	items := make([]*models.LeaderboardItem, summaryItems.Len())

	for i := 0; i < summaryItems.Len(); i++ {
		key := summaryItems[i].Key
		items[i] = &models.LeaderboardItem{
			User:     user,
			UserID:   user.ID,
			Interval: (*interval)[0],
			By:       &by,
			Total:    summary.TotalTimeByKey(by, key),
			Key:      &key,
		}
	}

	return items, nil
}

func (srv *LeaderboardService) getHash(interval *models.IntervalKey, by *uint8) string {
	k := strings.Join(*interval, "__")
	if by != nil && !reflect.ValueOf(by).IsNil() {
		k += "__" + models.GetEntityColumn(*by)
	}
	return k
}
