package services

import (
	"github.com/emvi/logbuch"
	"github.com/muety/artifex"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/utils"
	"strconv"
	"sync"
	"time"

	"github.com/muety/wakapi/models"
)

type MiscService struct {
	config          *config.Config
	userService     IUserService
	summaryService  ISummaryService
	keyValueService IKeyValueService
	queueDefault    *artifex.Dispatcher
	queueWorkers    *artifex.Dispatcher
}

func NewMiscService(userService IUserService, summaryService ISummaryService, keyValueService IKeyValueService) *MiscService {
	return &MiscService{
		config:          config.Get(),
		userService:     userService,
		summaryService:  summaryService,
		keyValueService: keyValueService,
		queueDefault:    config.GetDefaultQueue(),
		queueWorkers:    config.GetQueue(config.QueueProcessing),
	}
}

func (srv *MiscService) ScheduleCountTotalTime() {
	logbuch.Info("scheduling total time counting")
	if _, err := srv.queueDefault.DispatchEvery(srv.CountTotalTime, 1*time.Hour); err != nil {
		config.Log().Error("failed to schedule user counting jobs, %v", err)
	}
}

func (srv *MiscService) CountTotalTime() {
	users, err := srv.userService.GetAll()
	if err != nil {
		config.Log().Error("failed to fetch users for time counting, %v", err)
	}

	var totalTime time.Duration = 0
	var pendingJobs sync.WaitGroup
	pendingJobs.Add(len(users))

	for _, u := range users {
		if err := srv.queueWorkers.Dispatch(func() {
			defer pendingJobs.Done()
			totalTime += srv.countUserTotalTime(u.ID)
		}); err != nil {
			config.Log().Error("failed to enqueue counting job for user '%s'", u.ID)
			pendingJobs.Done()
		}
	}

	// persist
	go func(wg *sync.WaitGroup) {
		if utils.WaitTimeout(&pendingJobs, 10*time.Minute) {
			if err := srv.keyValueService.PutString(&models.KeyStringValue{
				Key:   config.KeyLatestTotalTime,
				Value: totalTime.String(),
			}); err != nil {
				config.Log().Error("failed to save total time count: %v", err)
			}

			if err := srv.keyValueService.PutString(&models.KeyStringValue{
				Key:   config.KeyLatestTotalUsers,
				Value: strconv.Itoa(len(users)),
			}); err != nil {
				config.Log().Error("failed to save total users count: %v", err)
			}
		} else {
			config.Log().Error("waiting for user counting jobs timed out")
		}
	}(&pendingJobs)
}

func (srv *MiscService) countUserTotalTime(userId string) time.Duration {
	result, err := srv.summaryService.Aliased(time.Time{}, time.Now(), &models.User{ID: userId}, srv.summaryService.Retrieve, nil, false)
	if err != nil {
		config.Log().Error("failed to count total for user %s: %v", userId, err)
		return 0
	}
	return result.TotalTime()
}
