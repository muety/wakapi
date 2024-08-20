package services

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/muety/artifex/v2"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"log/slog"
	"time"
)

type HousekeepingService struct {
	config        *config.Config
	userSrvc      IUserService
	heartbeatSrvc IHeartbeatService
	summarySrvc   ISummaryService
	queueDefault  *artifex.Dispatcher
	queueWorkers  *artifex.Dispatcher
}

func NewHousekeepingService(userService IUserService, heartbeatService IHeartbeatService, summaryService ISummaryService) *HousekeepingService {
	return &HousekeepingService{
		config:        config.Get(),
		userSrvc:      userService,
		heartbeatSrvc: heartbeatService,
		summarySrvc:   summaryService,
		queueDefault:  config.GetDefaultQueue(),
		queueWorkers:  config.GetQueue(config.QueueHousekeeping),
	}
}

func (s *HousekeepingService) Schedule() {
	s.scheduleDataCleanups()
	s.scheduleInactiveUsersCleanup()
	s.scheduleProjectStatsCacheWarming()
}

func (s *HousekeepingService) CleanUserDataBefore(user *models.User, before time.Time) error {
	slog.Warn("cleaning up user data older than", "userID", user.ID, "date", before)
	if s.config.App.DataCleanupDryRun {
		slog.Info("skipping actual data deletion for dry run", "userID", user.ID)
		return nil
	}

	// clear old heartbeats
	if err := s.heartbeatSrvc.DeleteByUserBefore(user, before); err != nil {
		return err
	}

	// clear old summaries
	slog.Info("clearing summaries for user older than", "userID", user.ID, "date", before)
	if err := s.summarySrvc.DeleteByUserBefore(user.ID, before); err != nil {
		return err
	}

	return nil
}

func (s *HousekeepingService) CleanInactiveUsers(before time.Time) error {
	slog.Info("cleaning up users inactive since", "date", before)
	users, err := s.userSrvc.GetAll()
	if err != nil {
		return err
	}

	var i int
	for _, u := range users {
		if u.LastLoggedInAt.T().After(before) || u.HasData {
			continue
		}

		slog.Warn("deleting user due to inactivity and no data", "userID", u.ID)
		if err := s.userSrvc.Delete(u); err != nil {
			config.Log().Error("failed to delete user", "userID", u.ID)
		} else {
			i++
		}
	}
	slog.Info("deleted users due to inactivity", "deletedCount", i, "totalCount", len(users))

	return nil
}

func (s *HousekeepingService) WarmUserProjectStatsCache(user *models.User) error {
	slog.Info("pre-warming project stats cache for user", "userID", user.ID)
	if _, err := s.heartbeatSrvc.GetUserProjectStats(user, time.Time{}, utils.BeginOfToday(time.Local), nil, true); err != nil {
		config.Log().Error("failed to pre-warm project stats cache", "userID", user.ID, "error", err)
	}
	return nil
}

func (s *HousekeepingService) runWarmProjectStatsCache() {
	// fetch active users
	users, err := s.userSrvc.GetActive(false)
	if err != nil {
		config.Log().Error("failed to get active users for project stats cache warming", "error", err)
		return
	}

	// fetch user heartbeat counts
	userHeartbeatCounts, err := s.heartbeatSrvc.CountByUsers(users)
	if err != nil {
		config.Log().Error("failed to count user heartbeats for project stats cache warming", "error", err)
		return
	}

	// schedule jobs
	for _, c := range userHeartbeatCounts {
		// only warm cache for users with >= 100k heartbeats (where calculation is expected to take unbearably long)
		if c.Count < 100_000 {
			continue
		}

		user, _ := slice.FindBy[*models.User](users, func(i int, u *models.User) bool {
			return u.ID == c.User
		})
		s.queueWorkers.Dispatch(func() {
			if err := s.WarmUserProjectStatsCache(user); err != nil {
				config.Log().Error("failed to pre-warm project stats cache", "userID", user.ID)
			}
		})
	}
}

func (s *HousekeepingService) runCleanData() {
	// fetch all users
	users, err := s.userSrvc.GetAll()
	if err != nil {
		config.Log().Error("failed to get users for data cleanup", "error", err)
		return
	}

	// schedule jobs
	for _, u := range users {
		// don't clean data for subscribed users or when they otherwise have unlimited data access
		if u.MinDataAge().IsZero() {
			continue
		}

		user := *u
		s.queueWorkers.Dispatch(func() {
			if err := s.CleanUserDataBefore(&user, user.MinDataAge()); err != nil {
				config.Log().Error("failed to clear old user data", "userID", user.ID)
			}
		})
	}
}

func (s *HousekeepingService) runCleanInactiveUsers() {
	s.queueWorkers.Dispatch(func() {
		if s.config.App.MaxInactiveMonths <= 0 {
			return
		}
		if err := s.CleanInactiveUsers(time.Now().AddDate(0, -s.config.App.MaxInactiveMonths, 0)); err != nil {
			config.Log().Error("failed to clean up inactive users", "error", err)
		}
	})
}

// individual scheduling functions

func (s *HousekeepingService) scheduleDataCleanups() {
	if s.config.App.DataRetentionMonths <= 0 {
		return
	}

	slog.Info("scheduling data cleanup")

	_, err := s.queueDefault.DispatchCron(s.runCleanData, s.config.App.DataCleanupTime)
	if err != nil {
		config.Log().Error("failed to dispatch data cleanup jobs", "error", err)
	}
}

func (s *HousekeepingService) scheduleInactiveUsersCleanup() {
	if s.config.App.MaxInactiveMonths <= 0 {
		return
	}

	slog.Info("scheduling inactive users cleanup")

	_, err := s.queueDefault.DispatchCron(s.runCleanInactiveUsers, s.config.App.DataCleanupTime)
	if err != nil {
		config.Log().Error("failed to dispatch inactive users cleanup job", "error", err)
	}
}

func (s *HousekeepingService) scheduleProjectStatsCacheWarming() {
	slog.Info("scheduling project stats cache pre-warming")

	_, err := s.queueDefault.DispatchEvery(s.runWarmProjectStatsCache, 12*time.Hour)
	if err != nil {
		config.Log().Error("failed to dispatch pre-warming project stats cache", "error", err)
	}

	// run once initially, 1 min after start
	if !s.config.QuickStart {
		if err := s.queueDefault.DispatchIn(s.runWarmProjectStatsCache, 1*time.Minute); err != nil {
			config.Log().Error("failed to dispatch pre-warming project stats cache", "error", err)
		}
	}
}
