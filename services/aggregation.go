package services

import (
	"log"
	"runtime"
	"time"

	"github.com/jasonlvhit/gocron"
	"github.com/jinzhu/gorm"
	"github.com/muety/wakapi/models"
)

const (
	aggregateIntervalDays int = 1 // TODO: Make configurable
)

type AggregationService struct {
	Config           *models.Config
	Db               *gorm.DB
	UserService      *UserService
	SummaryService   *SummaryService
	HeartbeatService *HeartbeatService
}

type AggregationJob struct {
	UserID string
	From   time.Time
	To     time.Time
}

func (srv *AggregationService) Init() {}

// Schedule a job to (re-)generate summaries every day shortly after midnight
// TODO: Make configurable
func (srv *AggregationService) Schedule() {
	jobs := make(chan *AggregationJob)
	summaries := make(chan *models.Summary)
	defer close(jobs)
	defer close(summaries)

	for i := 0; i < runtime.NumCPU(); i++ {
		go srv.summaryWorker(jobs, summaries)
	}

	for i := 0; i < int(srv.Config.DbMaxConn); i++ {
		go srv.persistWorker(summaries)
	}

	// Run once initially
	srv.trigger(jobs)

	gocron.Every(1).Day().At("02:15").Do(srv.trigger, jobs)
	<-gocron.Start()
}

func (srv *AggregationService) summaryWorker(jobs <-chan *AggregationJob, summaries chan<- *models.Summary) {
	for job := range jobs {
		if summary, err := srv.SummaryService.Construct(job.From, job.To, &models.User{ID: job.UserID}, true); err != nil {
			log.Printf("Failed to generate summary (%v, %v, %s) – %v.\n", job.From, job.To, job.UserID, err)
		} else {
			log.Printf("Successfully generated summary (%v, %v, %s).\n", job.From, job.To, job.UserID)
			summaries <- summary
		}
	}
}

func (srv *AggregationService) persistWorker(summaries <-chan *models.Summary) {
	for summary := range summaries {
		if err := srv.SummaryService.Insert(summary); err != nil {
			log.Printf("Failed to save summary (%v, %v, %s) – %v.\n", summary.UserID, summary.FromTime, summary.ToTime, err)
		}
	}
}

func (srv *AggregationService) trigger(jobs chan<- *AggregationJob) error {
	log.Println("Generating summaries.")

	users, err := srv.UserService.GetAll()
	if err != nil {
		log.Println(err)
		return err
	}

	latestSummaries, err := srv.SummaryService.GetLatestByUser()
	if err != nil {
		log.Println(err)
		return err
	}

	userSummaryTimes := make(map[string]time.Time)
	for _, s := range latestSummaries {
		userSummaryTimes[s.UserID] = s.ToTime
	}

	missingUserIDs := make([]string, 0)
	for _, u := range users {
		if _, ok := userSummaryTimes[u.ID]; !ok {
			missingUserIDs = append(missingUserIDs, u.ID)
		}
	}

	firstHeartbeats, err := srv.HeartbeatService.GetFirstUserHeartbeats(missingUserIDs)
	if err != nil {
		log.Println(err)
		return err
	}

	for id, t := range userSummaryTimes {
		generateUserJobs(id, t, jobs)
	}

	for _, h := range firstHeartbeats {
		generateUserJobs(h.UserID, time.Time(h.Time), jobs)
	}

	return nil
}

func generateUserJobs(userId string, lastAggregation time.Time, jobs chan<- *AggregationJob) {
	var from, to time.Time
	end := getStartOfToday().Add(-1 * time.Second)

	if lastAggregation.Hour() == 0 {
		from = lastAggregation
	} else {
		from = time.Date(
			lastAggregation.Year(),
			lastAggregation.Month(),
			lastAggregation.Day()+aggregateIntervalDays,
			0, 0, 0, 0,
			lastAggregation.Location(),
		)
	}

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
