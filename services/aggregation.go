package services

import (
	"errors"
	datastructure "github.com/duke-git/lancet/v2/datastructure/set"
	"github.com/muety/artifex/v2"
	"github.com/muety/wakapi/config"
	"log/slog"
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
		inProgress:       datastructure.New[string](),
		queueDefault:     config.GetDefaultQueue(),
		queueWorkers:     config.GetQueue(config.QueueProcessing),
	}
}

type AggregationJob struct {
	User *models.User
	From time.Time
	To   time.Time
}

// Schedule a job to (re-)generate summaries every day shortly after midnight
func (srv *AggregationService) Schedule() {
	slog.Info("scheduling summary aggregation")

	if _, err := srv.queueDefault.DispatchCron(func() {
		if err := srv.AggregateSummaries(datastructure.New[string]()); err != nil {
			config.Log().Error("failed to generate summaries", "error", err)
		}
	}, srv.config.App.GetAggregationTimeCron()); err != nil {
		config.Log().Error("failed to schedule summary generation", "error", err)
	}
}

func (srv *AggregationService) AggregateSummaries(userIds datastructure.Set[string]) error {
	if err := srv.lockUsers(userIds); err != nil {
		return err
	}
	defer srv.unlockUsers(userIds)

	slog.Info("generating summaries")

	// Get a map from user ids to the time of their latest summary or nil if none exists yet
	lastUserSummaryTimes, err := srv.summaryService.GetLatestByUser()
	if err != nil {
		config.Log().Error("error occurred", "error", err)
		return err
	}

	// Get a map from user ids to the time of their earliest heartbeats or nil if none exists yet
	firstUserHeartbeatTimes, err := srv.heartbeatService.GetFirstByUsers()
	if err != nil {
		config.Log().Error("error occurred", "error", err)
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
				config.Log().Error("failed to dispatch summary generation job", "userID", job.User.ID)
			}
		}
	}()

	// Fetch complete user objects
	var users map[string]*models.User
	if userIds != nil && !userIds.IsEmpty() {
		users, err = srv.userService.GetManyMapped(userIds.Values())
	} else {
		users, err = srv.userService.GetAllMapped()
	}
	if err != nil {
		return err
	}

	// Generate summary aggregation jobs
	for _, e := range lastUserSummaryTimes {
		if userIds != nil && !userIds.IsEmpty() && !userIds.Contain(e.User) {
			continue
		}

		u, _ := users[e.User]

		if e.Time.Valid() {
			// Case 1: User has aggregated summaries already
			// -> Spawn jobs to create summaries from their latest aggregation to now
			generateUserJobs(u, e.Time.T(), jobs)
		} else if t := firstUserHeartbeatLookup[e.User]; t.Valid() {
			// Case 2: User has no aggregated summaries, yet, but has heartbeats
			// -> Spawn jobs to create summaries from their first heartbeat to now
			generateUserJobs(u, t.T(), jobs)
		}
		// Case 3: User doesn't have heartbeats at all
		// -> Nothing to do
	}

	return nil
}

func (srv *AggregationService) process(job AggregationJob) {
	if summary, err := srv.summaryService.Summarize(job.From, job.To, job.User, nil); err != nil {
		config.Log().Error("failed to generate summary", "from", job.From, "to", job.To, "userID", job.User.ID, "error", err)
	} else {
		slog.Info("successfully generated summary", "from", job.From, "to", job.To, "userID", job.User.ID)
		if err := srv.summaryService.Insert(summary); err != nil {
			config.Log().Error("failed to save summary", "userID", summary.UserID, "fromTime", summary.FromTime, "toTime", summary.ToTime, "error", err)
		}
	}
}

func generateUserJobs(user *models.User, from time.Time, jobs chan<- *AggregationJob) {
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
		jobs <- &AggregationJob{user, from, to}
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
