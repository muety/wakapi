package services

import (
	"errors"
	"log/slog"
	"sync"
	"time"

	datastructure "github.com/duke-git/lancet/v2/datastructure/set"
	"github.com/muety/artifex/v2"
	"github.com/muety/wakapi/config"

	"github.com/muety/wakapi/models"
)

const (
	aggregateIntervalDays int = 1
)

var aggregationLock = sync.Mutex{}

type AggregationService struct {
	config                *config.Config
	userService           IUserService
	summaryService        ISummaryService
	heartbeatService      IHeartbeatService
	durationService       IDurationService
	inProgress            datastructure.Set[string]
	allInProgress         bool
	queueDefault          *artifex.Dispatcher
	queueSummaryWorkers   *artifex.Dispatcher
	queuedDurationWorkers *artifex.Dispatcher
}

func NewAggregationService(userService IUserService, summaryService ISummaryService, heartbeatService IHeartbeatService, durationService IDurationService) *AggregationService {
	return &AggregationService{
		config:                config.Get(),
		userService:           userService,
		summaryService:        summaryService,
		heartbeatService:      heartbeatService,
		durationService:       durationService,
		inProgress:            datastructure.New[string](),
		queueDefault:          config.GetDefaultQueue(),
		queueSummaryWorkers:   config.GetQueue(config.QueueProcessing),
		queuedDurationWorkers: config.GetQueue(config.QueueProcessing2),
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
		if err := srv.AggregateSummaries(datastructure.New[string](), false); err != nil {
			config.Log().Error("failed to regenerate summaries", "error", err)
		}
	}, srv.config.App.GetAggregationTimeCron()); err != nil {
		config.Log().Error("failed to schedule summary generation", "error", err)
	}
}

// AggregateSummaries generates summaries for the given users. If includeDurations is true, it also regenerates durations from scratch.
func (srv *AggregationService) AggregateSummaries(userIds datastructure.Set[string], includeDurations bool) error {
	if err := srv.lockUsers(userIds); err != nil {
		return err
	}
	defer srv.unlockUsers(userIds)

	slog.Info("generating summaries", "num_users", len(userIds))

	// Get a map from user ids to the time of their latest summary or nil if none exists yet
	lastUserSummaryTimes, err := srv.summaryService.GetLatestByUser() // TODO: build user-specific variant of this query for efficiency
	if err != nil {
		config.Log().Error("error occurred", "error", err.Error())
		return err
	}

	// Get a map from user ids to the time of their earliest heartbeats or nil if none exists yet
	firstUserHeartbeatTimes, err := srv.heartbeatService.GetFirstAll() // TODO: build user-specific variant of this query for efficiency
	if err != nil {
		config.Log().Error("error occurred", "error", err.Error())
		return err
	}

	// Build actual lookup table from it
	firstUserHeartbeatLookup := make(map[string]models.CustomTime)
	for _, e := range firstUserHeartbeatTimes {
		firstUserHeartbeatLookup[e.User] = e.Time
	}

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
	var allWg sync.WaitGroup
	for _, user := range users {
		u := *user
		jobs := make([]*AggregationJob, 0)

		durationsWg := &sync.WaitGroup{}

		// regenerate durations for the user if requested
		// generally not needed, because event listener in durations service will take care of generating new durations on a regular interval
		if includeDurations {
			durationsWg.Add(1)
			allWg.Add(1)
			if err := srv.queuedDurationWorkers.Dispatch(func() {
				defer allWg.Done()
				defer durationsWg.Done()
				slog.Info("regenerating user durations as part of summary aggregation", "user", u.ID)
				srv.durationService.Regenerate(&u, true)
			}); err != nil {
				durationsWg.Done()
				allWg.Done()
				config.Log().Error("failed to dispatch durations generation job", "userID", u.ID, "error", err)
			}
		}

		// generate actual summary aggregation jobs
		for _, e := range lastUserSummaryTimes {
			if e.User != user.ID {
				continue
			}

			if e.Time.Valid() {
				// Case 1: User has aggregated summaries already
				// -> Spawn jobs to create summaries from their latest aggregation to now
				slog.Info("generating summary aggregation jobs for user", "user", u.ID, "from", e.Time.T())
				jobs = append(jobs, generateUserJobs(&u, e.Time.T())...)
			} else if t := firstUserHeartbeatLookup[e.User]; t.Valid() {
				// Case 2: User has no aggregated summaries, yet, but has heartbeats
				// -> Spawn jobs to create summaries from their first heartbeat to now
				slog.Info("generating summary aggregation jobs for user", "user", u.ID, "from", t.T())
				jobs = append(jobs, generateUserJobs(&u, t.T())...)
			} else {
				// Case 3: User doesn't have heartbeats at all
				// -> Nothing to do
				slog.Info("skipping summary aggregation because user has no heartbeats", "user", u.ID)
			}
		}

		// dispatch the jobs for current user
		for _, jobRef := range jobs {
			job := *jobRef
			allWg.Add(1)
			if err := srv.queueSummaryWorkers.Dispatch(func() {
				defer allWg.Done()
				durationsWg.Wait()
				srv.process(job)
			}); err != nil {
				allWg.Done()
				config.Log().Error("failed to dispatch summary generation job", "userID", job.User.ID)
			}
		}
	}

	allWg.Wait()
	return nil
}

func (srv *AggregationService) AggregateDurations(userIds datastructure.Set[string]) (err error) {
	if err := srv.lockUsers(userIds); err != nil {
		return err
	}
	defer srv.unlockUsers(userIds)

	slog.Info("generating durations", "num_users", len(userIds))

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

	var wg sync.WaitGroup
	for _, u := range users {
		user := &(*u)
		wg.Add(1)
		if err := srv.queuedDurationWorkers.Dispatch(func() {
			defer wg.Done()
			srv.durationService.Regenerate(user, true)
		}); err != nil {
			wg.Done()
			config.Log().Error("failed to dispatch durations generation job", "userID", user.ID, "error", err)
		}
	}
	wg.Wait()

	return nil
}

func (srv *AggregationService) process(job AggregationJob) {
	// process single summary interval for single user
	slog.Info("regenerating actual user summaries as part of summary aggregation", "user", job.User.ID, "from", job.From, "to", job.To)
	if summary, err := srv.summaryService.Summarize(job.From, job.To, job.User, nil, nil); err != nil {
		config.Log().Error("failed to regenerate summary", "from", job.From, "to", job.To, "userID", job.User.ID, "error", err)
	} else {
		slog.Info("successfully generated summary", "from", job.From, "to", job.To, "userID", job.User.ID)
		if err := srv.summaryService.Insert(summary); err != nil {
			config.Log().Error("failed to save summary", "userID", summary.UserID, "fromTime", summary.FromTime.T(), "toTime", summary.ToTime.T(), "error", err)
		}
	}
}

func generateUserJobs(user *models.User, from time.Time) (jobs []*AggregationJob) {
	var to time.Time
	tz := time.Local
	if user != nil {
		tz = user.TZ()
	}

	// Go to next day of either user's first heartbeat or latest aggregation
	from = from.In(tz).Add(-1 * time.Second)
	from = time.Date(
		from.Year(),
		from.Month(),
		from.Day()+aggregateIntervalDays,
		0, 0, 0, 0,
		tz,
	)

	// Iteratively aggregate per-day summaries until end of yesterday is reached
	end := getStartOfToday(tz).Add(-1 * time.Second)
	for from.Before(end) && to.Before(end) {
		to = time.Date(
			from.Year(),
			from.Month(),
			from.Day()+aggregateIntervalDays,
			0, 0, 0, 0,
			tz,
		)
		jobs = append(jobs, &AggregationJob{user, from, to})
		from = to
	}

	return jobs
}

func (srv *AggregationService) IsLocked(userId string) bool {
	aggregationLock.Lock()
	defer aggregationLock.Unlock()
	return srv.allInProgress || (srv.inProgress != nil && srv.inProgress.Contain(userId))
}

func (srv *AggregationService) lockUsers(userIds datastructure.Set[string]) error {
	aggregationLock.Lock()
	defer aggregationLock.Unlock()

	if srv.inProgress == nil {
		srv.inProgress = datastructure.New[string]()
	}

	if userIds == nil || userIds.IsEmpty() {
		if srv.allInProgress || !srv.inProgress.IsEmpty() {
			return errors.New("aggregation already in progress for at least of the request users")
		}
		srv.allInProgress = true
		return nil
	}

	if srv.allInProgress {
		return errors.New("aggregation already in progress for at least of the request users")
	}

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

	if userIds == nil || userIds.IsEmpty() {
		srv.allInProgress = false
		return
	}

	if srv.inProgress == nil {
		return
	}

	for uid := range userIds {
		srv.inProgress.Delete(uid)
	}
}

func getStartOfToday(tz *time.Location) time.Time {
	now := time.Now().In(tz)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 1, tz)
}
