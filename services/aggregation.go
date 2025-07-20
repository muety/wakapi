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
	summarytypes "github.com/muety/wakapi/types"
	"gorm.io/gorm"
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
	progressTracker  map[string]time.Time // Track last processed time for each user
	progressMutex    sync.RWMutex
}

func NewAggregationService(db *gorm.DB) *AggregationService {
	summaryService := NewSummaryService(db)
	userService := NewUserService(db)
	heartbeatService := NewHeartbeatService(db)
	return &AggregationService{
		config:           config.Get(),
		userService:      userService,
		summaryService:   summaryService,
		heartbeatService: heartbeatService,
		inProgress:       datastructure.New[string](),
		queueDefault:     config.GetDefaultQueue(),
		queueWorkers:     config.GetQueue(config.QueueProcessing),
		progressTracker:  make(map[string]time.Time),
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

	// Process users in batches to avoid memory issues
	const batchSize = 100
	userIdsList := make([]string, 0)

	// Collect user IDs to process
	for _, e := range lastUserSummaryTimes {
		if userIds != nil && !userIds.IsEmpty() && !userIds.Contain(e.User) {
			continue
		}
		userIdsList = append(userIdsList, e.User)
	}

	// Process users in batches
	for i := 0; i < len(userIdsList); i += batchSize {
		end := i + batchSize
		if end > len(userIdsList) {
			end = len(userIdsList)
		}

		batchIds := userIdsList[i:end]

		// Fetch user objects for this batch
		users, err := srv.userService.GetManyMapped(batchIds)
		if err != nil {
			config.Log().Error("failed to fetch user batch", "batch_start", i, "batch_end", end, "error", err)
			continue // Process other batches even if one fails
		}

		// Generate summary aggregation jobs for this batch
		for userId, user := range users {
			// Find the summary time for this user
			var summaryTime *models.TimeByUser
			for _, e := range lastUserSummaryTimes {
				if e.User == userId {
					summaryTime = e
					break
				}
			}

			if summaryTime == nil {
				continue
			}

			if summaryTime.Time.Valid() {
				// Case 1: User has aggregated summaries already
				// -> Spawn jobs to create summaries from their latest aggregation to now
				generateUserJobs(user, summaryTime.Time.T(), jobs)
			} else if t := firstUserHeartbeatLookup[userId]; t.Valid() {
				// Case 2: User has no aggregated summaries, yet, but has heartbeats
				// -> Spawn jobs to create summaries from their first heartbeat to now
				generateUserJobs(user, t.T(), jobs)
			}
			// Case 3: User doesn't have heartbeats at all
			// -> Nothing to do
		}
	}

	return nil
}

func (srv *AggregationService) process(job AggregationJob) {
	request := summarytypes.NewSummaryRequest(job.From, job.To, job.User)

	// Generate summary with retry on transient errors
	var summary *models.Summary
	var err error
	retryCount := 3

	for i := range retryCount {
		summary, err = srv.summaryService.ComputeFromDurations(request)
		if err == nil {
			break
		}

		// Log retry attempts
		if i < retryCount-1 {
			config.Log().Warn("retrying summary generation", "attempt", i+1, "from", job.From, "to", job.To, "userID", job.User.ID, "error", err)
			time.Sleep(time.Second * time.Duration(i+1)) // Exponential backoff
		}
	}

	if err != nil {
		config.Log().Error("failed to generate summary after retries", "from", job.From, "to", job.To, "userID", job.User.ID, "error", err)
		return
	}

	slog.Info("successfully generated summary", "from", job.From, "to", job.To, "userID", job.User.ID)

	// Save summary with retry
	for i := 0; i < retryCount; i++ {
		err = srv.summaryService.Insert(summary)
		if err == nil {
			break
		}

		if i < retryCount-1 {
			config.Log().Warn("retrying summary save", "attempt", i+1, "userID", summary.UserID, "fromTime", summary.FromTime, "toTime", summary.ToTime, "error", err)
			time.Sleep(time.Second * time.Duration(i+1))
		}
	}

	if err != nil {
		config.Log().Error("failed to save summary after retries", "userID", summary.UserID, "fromTime", summary.FromTime, "toTime", summary.ToTime, "error", err)
	} else {
		// Update progress tracker
		srv.progressMutex.Lock()
		srv.progressTracker[job.User.ID] = job.To
		srv.progressMutex.Unlock()
	}
}

func generateUserJobs(user *models.User, from time.Time, jobs chan<- *AggregationJob) {
	var to time.Time
	userTZ := user.TZ()

	// Convert to user's timezone
	from = from.In(userTZ)

	// Move to start of next day in user's timezone
	// We don't subtract 1 second here to avoid boundary issues
	from = time.Date(
		from.Year(),
		from.Month(),
		from.Day(),
		0, 0, 0, 0,
		userTZ,
	).AddDate(0, 0, aggregateIntervalDays)

	// Iteratively aggregate per-day summaries until end of yesterday is reached
	// Get "today" in user's timezone
	end := getStartOfTodayForUser(user).Add(-1 * time.Second)
	for from.Before(end) && to.Before(end) {
		to = time.Date(
			from.Year(),
			from.Month(),
			from.Day()+aggregateIntervalDays,
			0, 0, 0, 0,
			userTZ, // Use user's timezone
		)
		jobs <- &AggregationJob{user, from, to}
		from = to
	}
}

func (srv *AggregationService) lockUsers(userIds datastructure.Set[string]) error {
	aggregationLock.Lock()
	defer aggregationLock.Unlock()

	// Keep track of users we've added so we can rollback on failure
	addedUsers := datastructure.New[string]()

	// Check and set atomically
	for uid := range userIds {
		if srv.inProgress.Contain(uid) {
			// Rollback any users we already added
			for addedId := range addedUsers {
				srv.inProgress.Delete(addedId)
			}
			return errors.New("aggregation already in progress for at least one of the requested users")
		}
		srv.inProgress.Add(uid)
		addedUsers.Add(uid)
	}
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

// getStartOfTodayForUser returns the start of today in the user's timezone
func getStartOfTodayForUser(user *models.User) time.Time {
	now := time.Now().In(user.TZ())
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, user.TZ())
}
