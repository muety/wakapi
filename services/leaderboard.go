package services

import (
	"github.com/emvi/logbuch"
	"github.com/leandro-lugaresi/hub"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
	"github.com/muety/wakapi/utils"
	"github.com/patrickmn/go-cache"
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
	return &LeaderboardService{
		config:         config.Get(),
		cache:          cache.New(24*time.Hour, 24*time.Hour),
		eventBus:       config.EventBus(),
		repository:     leaderboardRepo,
		summaryService: summaryService,
		userService:    userService,
	}
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

	runAllUsers(models.IntervalPast7Days, []uint8{models.SummaryLanguage})

	//s := gocron.NewScheduler(time.Local)
	//s.Every(1).Day().At(srv.config.App.LeaderboardGenerationTime).Do(runAllUsers, models.IntervalPast7Days, []uint8{models.SummaryLanguage})
	//s.StartBlocking()
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

	logbuch.Info("finished leaderboard generation")

	return nil
}

func (srv *LeaderboardService) GetByInterval(interval *models.IntervalKey) ([]*models.LeaderboardItem, error) {
	return srv.GetAggregatedByInterval(interval, nil)
}

func (srv *LeaderboardService) GetAggregatedByInterval(interval *models.IntervalKey, by *uint8) ([]*models.LeaderboardItem, error) {
	return srv.repository.GetAllAggregatedByInterval(interval, by)
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
