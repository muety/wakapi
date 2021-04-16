package services

import (
	"github.com/emvi/logbuch"
	"github.com/muety/wakapi/config"
	"runtime"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/muety/wakapi/models"
)

const (
	aggregateIntervalDays int = 1
)

type AggregationService struct {
	config           *config.Config
	userService      IUserService
	summaryService   ISummaryService
	heartbeatService IHeartbeatService
}

func NewAggregationService(userService IUserService, summaryService ISummaryService, heartbeatService IHeartbeatService) *AggregationService {
	return &AggregationService{
		config:           config.Get(),
		userService:      userService,
		summaryService:   summaryService,
		heartbeatService: heartbeatService,
	}
}

type AggregationJob struct {
	UserID string
	From   time.Time
	To     time.Time
}

// Schedule a job to (re-)generate summaries every day shortly after midnight
func (srv *AggregationService) Schedule() {
	// Run once initially
	if err := srv.Run(nil); err != nil {
		logbuch.Fatal("failed to run AggregationJob: %v", err)
	}

	s := gocron.NewScheduler(time.Local)
	s.Every(1).Day().At(srv.config.App.AggregationTime).Do(srv.Run, map[string]bool{})
	s.StartBlocking()
}

func (srv *AggregationService) Run(userIds map[string]bool) error {
	jobs := make(chan *AggregationJob)
	summaries := make(chan *models.Summary)

	for i := 0; i < runtime.NumCPU(); i++ {
		go srv.summaryWorker(jobs, summaries)
	}

	for i := 0; i < int(srv.config.Db.MaxConn); i++ {
		go srv.persistWorker(summaries)
	}

	// don't leak open channels
	go func(c1 chan *AggregationJob, c2 chan *models.Summary) {
		defer close(c1)
		defer close(c2)
		time.Sleep(1 * time.Hour)
	}(jobs, summaries)

	return srv.trigger(jobs, userIds)
}

func (srv *AggregationService) summaryWorker(jobs <-chan *AggregationJob, summaries chan<- *models.Summary) {
	for job := range jobs {
		if summary, err := srv.summaryService.Summarize(job.From, job.To, &models.User{ID: job.UserID}); err != nil {
			config.Log().Error("failed to generate summary (%v, %v, %s) – %v", job.From, job.To, job.UserID, err)
		} else {
			logbuch.Info("successfully generated summary (%v, %v, %s)", job.From, job.To, job.UserID)
			summaries <- summary
		}
	}
}

func (srv *AggregationService) persistWorker(summaries <-chan *models.Summary) {
	for summary := range summaries {
		if err := srv.summaryService.Insert(summary); err != nil {
			config.Log().Error("failed to save summary (%v, %v, %s) – %v", summary.UserID, summary.FromTime, summary.ToTime, err)
		}
	}
}

func (srv *AggregationService) trigger(jobs chan<- *AggregationJob, userIds map[string]bool) error {
	logbuch.Info("generating summaries")

	var users []*models.User
	if allUsers, err := srv.userService.GetAll(); err != nil {
		config.Log().Error(err.Error())
		return err
	} else if userIds != nil && len(userIds) > 0 {
		users = make([]*models.User, 0)
		for _, u := range allUsers {
			if yes, ok := userIds[u.ID]; yes && ok {
				users = append(users, u)
			}
		}
	} else {
		users = allUsers
	}

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

	// Generate summary aggregation jobs
	for _, e := range lastUserSummaryTimes {
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

func getStartOfToday() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 1, now.Location())
}
