package services

import (
	"errors"
	"iter"
	"log/slog"
	"sync"
	"time"

	datastructure "github.com/duke-git/lancet/v2/datastructure/set"
	"github.com/muety/artifex/v2"
	"github.com/muety/wakapi/config"

	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/models/types"
)

const (
	aggregateIntervalDays  int = 1
	maxDispatchSummaryJobs int = 64
)

var aggregationLock = sync.Mutex{}

type AggregationService struct {
	config                *config.Config
	userService           IUserService
	summaryService        ISummaryService
	heartbeatService      IHeartbeatService
	durationService       IDurationService
	inProgress            datastructure.Set[string]
	inProgressAll         bool
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

	// Note that QueueDefault (single worker) will be blocked while summary aggregation is running, but should be acceptable for now.
	// Alternative would be to introduce another queue (QueueAggregation or such) solely for this purpose.
	// actual aggregation jobs run on QueueProcessing, durations generation runs on QueueProcessing2.
	if _, err := srv.queueDefault.DispatchCron(func() {
		if err := srv.AggregateSummaries(datastructure.New[string](), types.DurationsRegenerationIncremental); err != nil {
			config.Log().Error("failed to regenerate summaries", "error", err)
		}
	}, srv.config.App.GetAggregationTimeCron()); err != nil {
		config.Log().Error("failed to schedule summary generation", "error", err)
	}
}

// AggregateSummaries generates summaries for the given users. Depending on the mode, it additionally regenerates users' durations, either only incrementally or entirely. Blocks until all done.
func (srv *AggregationService) AggregateSummaries(userIds datastructure.Set[string], durationsMode types.DurationsRegenerationMode) error {
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

	firstUserHeartbeatLookup := make(map[string]models.CustomTime)
	for _, e := range firstUserHeartbeatTimes {
		firstUserHeartbeatLookup[e.User] = e.Time
	}

	lastUserSummaryLookup := make(map[string]models.CustomTime)
	for _, e := range lastUserSummaryTimes {
		lastUserSummaryLookup[e.User] = e.Time
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
	wgGlobal := &sync.WaitGroup{}
	// artifex does not block upon dispatching into a saturated worker pool, but instead eagerly spawns goroutines
	// This semaphore enforces actual backpressure so that the lazy user jobs generator only yields up to 64 at a time
	// See https://github.com/muety/wakapi/issues/997
	// To clarify: jobs are still only being _processed_ with the queue's configured concurrency, but we still don't want to _dispatch_ thousands of (waiting) goroutines to save memory
	sem := make(chan struct{}, maxDispatchSummaryJobs)

	for _, user := range users {
		u := *user

		// determine start times for summary aggregation (actual jobs are only generated lazily upon dispatch to keep memory usage low, see #997)
		var starts []time.Time
		if lastSummary, ok := lastUserSummaryLookup[u.ID]; ok && lastSummary.Valid() {
			// Case 1: User has aggregated summaries already
			// -> Spawn jobs to create summaries from their latest aggregation to now
			slog.Info("generating summary aggregation jobs for user", "user", u.ID, "from", lastSummary.T())
			starts = append(starts, lastSummary.T())
		} else if firstHeartbeat, ok := firstUserHeartbeatLookup[u.ID]; ok && firstHeartbeat.Valid() {
			// Case 2: User has no aggregated summaries, yet, but has heartbeats
			// -> Spawn jobs to create summaries from their first heartbeat to now
			slog.Info("generating summary aggregation jobs for user", "user", u.ID, "from", firstHeartbeat.T())
			starts = append(starts, firstHeartbeat.T())
		} else {
			// Case 3: User doesn't have heartbeats at all
			// -> Nothing to do
			slog.Info("skipping summary aggregation because user has no heartbeats", "user", u.ID)
		}

		// generate and dispatch this user's summary jobs lazily under backpressure
		dispatchSummaryJobs := func() {
			for _, start := range starts {
				for job := range generateUserJobs(&u, start) {
					sem <- struct{}{}
					wgGlobal.Add(1)

					if err := srv.queueSummaryWorkers.Dispatch(func() {
						defer func() {
							<-sem
							wgGlobal.Done()
						}()
						srv.process(job)
					}); err != nil {
						<-sem
						wgGlobal.Done()
						config.Log().Error("failed to dispatch summary generation job", "userID", job.User.ID)
					}
				}
			}
		}

		// regenerate durations for the user
		// incremental regeneration mainly serves as catch-up fallback, as the event listener in durations service will also take care of generating new durations on a regular interval
		if durationsMode != types.DurationsRegenerationNone {
			forceAll := durationsMode == types.DurationsRegenerationFull

			wgGlobal.Add(1)

			if err := srv.queuedDurationWorkers.Dispatch(func() {
				defer wgGlobal.Done()
				slog.Info("regenerating user durations as part of summary aggregation", "user", u.ID, "full", forceAll)
				srv.durationService.Regenerate(&u, forceAll)
				dispatchSummaryJobs()
			}); err != nil {
				wgGlobal.Done()
				config.Log().Error("failed to dispatch durations generation job", "userID", u.ID, "error", err)
				dispatchSummaryJobs() // only dispatch after durations generation is done to not block worker slots while just waiting for this to finish
			}
		} else {
			dispatchSummaryJobs()
		}
	}

	wgGlobal.Wait() // optionally for one durations generation job, followed by n summary aggregation jobs
	return nil
}

// AggregateDurations generates durations for the given users. Blocks until all done.
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

	wg := &sync.WaitGroup{}

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

func generateUserJobs(user *models.User, start time.Time) iter.Seq[AggregationJob] {
	return func(yield func(AggregationJob) bool) {
		var to time.Time
		tz := time.Local
		if user != nil {
			tz = user.TZ()
		}

		// go to next day of either user's first heartbeat or latest aggregation
		from := start.In(tz).Add(-1 * time.Second)
		from = time.Date(
			from.Year(),
			from.Month(),
			from.Day()+aggregateIntervalDays,
			0, 0, 0, 0,
			tz,
		)

		// iteratively aggregate per-day summaries until end of yesterday is reached
		end := getStartOfToday(tz).Add(-1 * time.Second)
		for from.Before(end) {
			to = time.Date(
				from.Year(),
				from.Month(),
				from.Day()+aggregateIntervalDays,
				0, 0, 0, 0,
				tz,
			)
			if !to.After(from) {
				// in case of midnight DST transitions (e.g. falling back to 23:00 of the previous day), time.Date won't advance -> ensure strict forward progress by at least 24h
				// see https://github.com/muety/wakapi/issues/997
				to = from.Add(24 * time.Hour)
			}
			if !yield(AggregationJob{user, from, to}) {
				return
			}
			from = to
		}
	}
}

func (srv *AggregationService) IsLocked(userId string) bool {
	aggregationLock.Lock()
	defer aggregationLock.Unlock()
	return srv.inProgressAll || srv.inProgress.Contain(userId)
}

func (srv *AggregationService) lockUsers(userIds datastructure.Set[string]) error {
	aggregationLock.Lock()
	defer aggregationLock.Unlock()

	if userIds == nil || userIds.IsEmpty() {
		if srv.inProgressAll || !srv.inProgress.IsEmpty() {
			return errors.New("aggregation already in progress for at least of the requested users")
		}
		srv.inProgressAll = true
		return nil
	}

	if srv.inProgressAll {
		return errors.New("aggregation already in progress for at least of the requested users")
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
		srv.inProgressAll = false
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
