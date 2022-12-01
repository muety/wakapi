package services

import (
	"errors"
	datastructure "github.com/duke-git/lancet/v2/datastructure/set"
	"github.com/emvi/logbuch"
	"github.com/muety/artifex/v2"
	"github.com/muety/wakapi/config"
	"sync"
	"time"

	"github.com/muety/wakapi/models"
)

const (
	aggregateIntervalDays int = 1
)

var aggregationLock = sync.Mutex{}

type AggregationService struct {
	config           *config.Config
	userService      IUserService
	summaryService   ISummaryService
	heartbeatService IHeartbeatService
	inProgress       datastructure.Set[string]
	queueDefault     *artifex.Dispatcher
	queueWorkers     *artifex.Dispatcher
}

func NewAggregationService(userService IUserService, summaryService ISummaryService, heartbeatService IHeartbeatService) *AggregationService {
	return &AggregationService{
		config:           config.Get(),
		userService:      userService,
		summaryService:   summaryService,
		heartbeatService: heartbeatService,
		inProgress:       datastructure.NewSet[string](),
		queueDefault:     config.GetDefaultQueue(),
		queueWorkers:     config.GetQueue(config.QueueProcessing),
	}
}

type AggregationJob struct {
	UserID string
	From   time.Time
	To     time.Time
}

// Schedule a job to (re-)generate summaries every day shortly after midnight
func (srv *AggregationService) Schedule() {
	logbuch.Info("scheduling summary aggregation")

	if _, err := srv.queueDefault.DispatchCron(func() {
		if err := srv.AggregateSummaries(datastructure.NewSet[string]()); err != nil {
			config.Log().Error("failed to generate summaries, %v", err)
		}
	}, srv.config.App.GetAggregationTimeCron()); err != nil {
		config.Log().Error("failed to schedule summary generation, %v", err)
	}
}

func (srv *AggregationService) AggregateSummaries(userIds datastructure.Set[string]) error {
	if err := srv.lockUsers(userIds); err != nil {
		return err
	}
	defer srv.unlockUsers(userIds)

	logbuch.Info("generating summaries")

	// Get a map from user ids to the time of their latest summary or nil if none exists yet
	lastUserSummaryTimes, err := srv.summaryService.GetLatestByUser()
	if err != nil {
		config.Log().Error(err.Error())
		return err
	}

	// Get a map from user ids to the time of their earliest heartbeats or nil if none exists yet
	firstUserHeartbeatTimes, err := srv.heartbeatService.GetFirstByUsers()
	if err != nil {
		config.Log().Error(err.Error())
		return err
	}

	// Build actual lookup table from it
	firstUserHeartbeatLookup := make(map[string]models.CustomTime)
	for _, e := range firstUserHeartbeatTimes {
		firstUserHeartbeatLookup[e.User] = e.Time
	}

	// Dispatch summary generation jobs
	jobs := make(chan *AggregationJob)
	defer close(jobs)
	go func() {
		for jobRef := range jobs {
			job := *jobRef
			if err := srv.queueWorkers.Dispatch(func() {
				srv.process(job)
			}); err != nil {
				config.Log().Error("failed to dispatch summary generation job for user '%s'", job.UserID)
			}
		}
	}()

	// Generate summary aggregation jobs
	for _, e := range lastUserSummaryTimes {
		if userIds != nil && !userIds.IsEmpty() && !userIds.Contain(e.User) {
			continue
		}

		if e.Time.Valid() {
			// Case 1: User has aggregated summaries already
			// -> Spawn jobs to create summaries from their latest aggregation to now
			generateUserJobs(e.User, e.Time.T(), jobs)
		} else if t := firstUserHeartbeatLookup[e.User]; t.Valid() {
			// Case 2: User has no aggregated summaries, yet, but has heartbeats
			// -> Spawn jobs to create summaries from their first heartbeat to now
			generateUserJobs(e.User, t.T(), jobs)
		}
		// Case 3: User doesn't have heartbeats at all
		// -> Nothing to do
	}

	return nil
}

func (srv *AggregationService) process(job AggregationJob) {
	if summary, err := srv.summaryService.Summarize(job.From, job.To, &models.User{ID: job.UserID}, nil); err != nil {
		config.Log().Error("failed to generate summary (%v, %v, %s) - %v", job.From, job.To, job.UserID, err)
	} else {
		logbuch.Info("successfully generated summary (%v, %v, %s)", job.From, job.To, job.UserID)
		if err := srv.summaryService.Insert(summary); err != nil {
			config.Log().Error("failed to save summary (%v, %v, %s) - %v", summary.UserID, summary.FromTime, summary.ToTime, err)
		}
	}
}

func generateUserJobs(userId string, from time.Time, jobs chan<- *AggregationJob) {
	var to time.Time

	// Go to next day of either user's first heartbeat or latest aggregation
	from = from.Add(-1 * time.Second)
	from = time.Date(
		from.Year(),
		from.Month(),
		from.Day()+aggregateIntervalDays,
		0, 0, 0, 0,
		from.Location(),
	)

	// Iteratively aggregate per-day summaries until end of yesterday is reached
	end := getStartOfToday().Add(-1 * time.Second)
	for from.Before(end) && to.Before(end) {
		to = time.Date(
			from.Year(),
			from.Month(),
			from.Day()+aggregateIntervalDays,
			0, 0, 0, 0,
			from.Location(),
		)
		jobs <- &AggregationJob{userId, from, to}
		from = to
	}
}

func (srv *AggregationService) lockUsers(userIds datastructure.Set[string]) error {
	aggregationLock.Lock()
	defer aggregationLock.Unlock()
	for uid := range userIds {
		if srv.inProgress.Contain(uid) {
			return errors.New("aggregation already in progress for at least of the request users")
		}
	}
	srv.inProgress = srv.inProgress.Union(userIds)
	return nil
}

func (srv *AggregationService) unlockUsers(userIds datastructure.Set[string]) {
	aggregationLock.Lock()
	defer aggregationLock.Unlock()
	for uid := range userIds {
		srv.inProgress.Delete(uid)
	}
}

func getStartOfToday() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 1, now.Location())
}
