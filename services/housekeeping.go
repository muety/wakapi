package services

import (
	"github.com/emvi/logbuch"
	"github.com/muety/artifex/v2"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
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
	if s.config.App.DataRetentionMonths <= 0 {
		return
	}

	logbuch.Info("scheduling data cleanup")

	_, err := s.queueDefault.DispatchCron(func() {
		// fetch all users
		users, err := s.userSrvc.GetAll()
		if err != nil {
			config.Log().Error("failed to get users for data cleanup, %v", err)
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
					config.Log().Error("failed to clear old user data for '%s'", user.ID)
				}
			})
		}
	}, s.config.App.DataCleanupTime)

	if err != nil {
		config.Log().Error("failed to dispatch data cleanup jobs, %v", err)
	}
}

func (s *HousekeepingService) CleanUserDataBefore(user *models.User, before time.Time) error {
	logbuch.Warn("cleaning up user data for '%s' older than %v", user.ID, before)
	if s.config.App.DataCleanupDryRun {
		logbuch.Info("skipping actual data deletion for '%v', because this is just a dry run", user.ID)
		return nil
	}

	// clear old heartbeats
	if err := s.heartbeatSrvc.DeleteByUserBefore(user, before); err != nil {
		return err
	}

	// clear old summaries
	logbuch.Info("clearing summaries for user '%s' older than %v", user.ID, before)
	if err := s.summarySrvc.DeleteByUserBefore(user.ID, before); err != nil {
		return err
	}

	return nil
}
