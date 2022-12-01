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

	// this is not exactly precise, because of summer / winter time, etc.
	retentionDuration := time.Now().Sub(time.Now().AddDate(0, -s.config.App.DataRetentionMonths, 0))

	_, err := s.queueDefault.DispatchCron(func() {
		// fetch all users
		users, err := s.userSrvc.GetAll()
		if err != nil {
			config.Log().Error("failed to get users for data cleanup, %v", err)
			return
		}

		// schedule jobs
		for _, u := range users {
			user := *u
			s.queueWorkers.Dispatch(func() {
				if err := s.ClearOldUserData(&user, retentionDuration); err != nil {
					config.Log().Error("failed to clear old user data for '%s'", user.ID)
				}
			})
		}
	}, s.config.App.DataCleanupTime)

	if err != nil {
		config.Log().Error("failed to dispatch data cleanup jobs, %v", err)
	}
}

func (s *HousekeepingService) ClearOldUserData(user *models.User, maxAge time.Duration) error {
	before := time.Now().Add(-maxAge)
	logbuch.Warn("cleaning up user data for '%s' older than %v", user.ID, before)

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
