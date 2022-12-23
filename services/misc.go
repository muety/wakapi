package services

import (
	"fmt"
	"github.com/emvi/logbuch"
	"github.com/muety/artifex/v2"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/utils"
	"go.uber.org/atomic"
	"strconv"
	"sync"
	"time"

	"github.com/muety/wakapi/models"
)

const (
	countUsersEvery        = 1 * time.Hour
	computeOldestDataEvery = 6 * time.Hour
)

var countLock = sync.Mutex{}
var firstDataLock = sync.Mutex{}

type MiscService struct {
	config           *config.Config
	userService      IUserService
	heartbeatService IHeartbeatService
	summaryService   ISummaryService
	keyValueService  IKeyValueService
	queueDefault     *artifex.Dispatcher
	queueWorkers     *artifex.Dispatcher
}

func NewMiscService(userService IUserService, heartbeatService IHeartbeatService, summaryService ISummaryService, keyValueService IKeyValueService) *MiscService {
	return &MiscService{
		config:           config.Get(),
		userService:      userService,
		heartbeatService: heartbeatService,
		summaryService:   summaryService,
		keyValueService:  keyValueService,
		queueDefault:     config.GetDefaultQueue(),
		queueWorkers:     config.GetQueue(config.QueueProcessing),
	}
}

func (srv *MiscService) Schedule() {
	logbuch.Info("scheduling total time counting")
	if _, err := srv.queueDefault.DispatchEvery(srv.CountTotalTime, countUsersEvery); err != nil {
		config.Log().Error("failed to schedule user counting jobs, %v", err)
	}

	logbuch.Info("scheduling first data computing")
	if _, err := srv.queueDefault.DispatchEvery(srv.ComputeOldestHeartbeats, computeOldestDataEvery); err != nil {
		config.Log().Error("failed to schedule first data computing jobs, %v", err)
	}

	// run once initially for a fresh instance
	if !srv.existsUsersTotalTime() {
		if err := srv.queueDefault.Dispatch(srv.CountTotalTime); err != nil {
			config.Log().Error("failed to dispatch user counting jobs, %v", err)
		}
	}
	if !srv.existsUsersFirstData() {
		if err := srv.queueDefault.Dispatch(srv.ComputeOldestHeartbeats); err != nil {
			config.Log().Error("failed to dispatch first data computing jobs, %v", err)
		}
	}
}

func (srv *MiscService) CountTotalTime() {
	logbuch.Info("counting users total time")
	if ok := countLock.TryLock(); !ok {
		config.Log().Warn("couldn't acquire lock for counting users total time, job is still pending")
	}
	defer countLock.Unlock()

	users, err := srv.userService.GetAll()
	if err != nil {
		config.Log().Error("failed to fetch users for time counting, %v", err)
	}

	var totalTime = atomic.NewDuration(0)
	var pendingJobs sync.WaitGroup
	pendingJobs.Add(len(users))

	for _, u := range users {
		user := *u
		if err := srv.queueWorkers.Dispatch(func() {
			defer pendingJobs.Done()
			totalTime.Add(srv.countUserTotalTime(user.ID))
		}); err != nil {
			config.Log().Error("failed to enqueue counting job for user '%s'", user.ID)
			pendingJobs.Done()
		}
	}

	// persist
	go func(wg *sync.WaitGroup) {
		if !utils.WaitTimeout(&pendingJobs, 2*countUsersEvery) {
			if err := srv.keyValueService.PutString(&models.KeyStringValue{
				Key:   config.KeyLatestTotalTime,
				Value: totalTime.Load().String(),
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

func (srv *MiscService) ComputeOldestHeartbeats() {
	logbuch.Info("computing users' first data")

	if err := srv.queueWorkers.Dispatch(func() {
		if ok := firstDataLock.TryLock(); !ok {
			config.Log().Warn("couldn't acquire lock for computing users' first data, job is still pending")
			return
		}
		defer firstDataLock.Unlock()

		results, err := srv.heartbeatService.GetFirstByUsers()
		if err != nil {
			config.Log().Error("failed to compute users' first data, %v", err)
		}

		for _, entry := range results {
			if entry.Time.T().IsZero() {
				continue
			}

			kvKey := fmt.Sprintf("%s_%s", config.KeyFirstHeartbeat, entry.User)
			if err := srv.keyValueService.PutString(&models.KeyStringValue{
				Key:   kvKey,
				Value: entry.Time.T().Format(time.RFC822Z),
			}); err != nil {
				config.Log().Error("failed to save user's first heartbeat time: %v", err)
			}
		}
	}); err != nil {
		config.Log().Error("failed to enqueue computing first data for user, %v", err)
	}
}

func (srv *MiscService) countUserTotalTime(userId string) time.Duration {
	result, err := srv.summaryService.Aliased(time.Time{}, time.Now(), &models.User{ID: userId}, srv.summaryService.Retrieve, nil, false)
	if err != nil {
		config.Log().Error("failed to count total for user %s: %v", userId, err)
		return 0
	}
	return result.TotalTime()
}

func (srv *MiscService) existsUsersTotalTime() bool {
	results, _ := srv.keyValueService.GetByPrefix(config.KeyLatestTotalTime)
	return len(results) > 0
}

func (srv *MiscService) existsUsersFirstData() bool {
	results, _ := srv.keyValueService.GetByPrefix(config.KeyFirstHeartbeat)
	return len(results) > 0
}
