/*
	<< WORK IN PROGRESS >>
	Don't use theses classes, yet.

	This aims to implement https://github.com/n1try/wakapi/issues/1.
	Idea is to have regularly running, cron-like background jobs that request a summary
	from SummaryService for a pre-defined time interval, e.g. 24 hours. Those are persisted
	to the database. Once a user request a summary for a certain time frame that partilly
	overlaps with pre-generated summaries, those will be aggregated together with actual heartbeats
	for the non-overlapping time frames left and right.
*/

package services

import (
	"log"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/n1try/wakapi/models"
)

const (
	summaryInterval time.Duration = 24 * time.Hour
	nSummaryWorkers int           = 8
	nPersistWorkers int           = 8
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

// Use https://godoc.org/github.com/jasonlvhit/gocron to trigger jobs on a regular basis.
func (srv *AggregationService) Start(interval time.Duration) {
	jobs := make(chan *AggregationJob)
	summaries := make(chan *models.Summary)

	for i := 0; i < nSummaryWorkers; i++ {
		go srv.summaryWorker(jobs, summaries)
	}

	for i := 0; i < nPersistWorkers; i++ {
		go srv.persistWorker(summaries)
	}

	srv.generateJobs(jobs)
}

func (srv *AggregationService) summaryWorker(jobs <-chan *AggregationJob, summaries chan<- *models.Summary) {
	for job := range jobs {
		if summary, err := srv.SummaryService.CreateSummary(job.From, job.To, &models.User{ID: job.UserID}); err != nil {
			log.Printf("Failed to generate summary (%v, %v, %s) – %v.", job.From, job.To, job.UserID, err)
		} else {
			summaries <- summary
		}
	}
}

func (srv *AggregationService) persistWorker(summaries <-chan *models.Summary) {
	for summary := range summaries {
		if err := srv.SummaryService.SaveSummary(summary); err != nil {
			log.Printf("Failed to save summary (%v, %v, %s) – %v.", summary.UserID, summary.FromTime, summary.ToTime, err)
		}
	}
}

func (srv *AggregationService) generateJobs(jobs chan<- *AggregationJob) error {
	users, err := srv.UserService.GetAll()
	if err != nil {
		return err
	}

	latestSummaries, err := srv.SummaryService.GetLatestUserSummaries()
	if err != nil {
		return err
	}

	userSummaryTimes := make(map[string]*time.Time)
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
		return err
	}

	for id, t := range userSummaryTimes {
		generateUserJobs(id, *t, jobs)
	}

	for _, h := range firstHeartbeats {
		generateUserJobs(h.UserID, time.Time(*(h.Time)), jobs)
	}

	return nil
}

func generateUserJobs(userId string, lastAggregation time.Time, jobs chan<- *AggregationJob) {
	var from, to time.Time
	end := getStartOfToday().Add(-1 * time.Second)

	if lastAggregation.Hour() == 0 {
		from = lastAggregation
	} else {
		nextDay := lastAggregation.Add(24 * time.Hour)
		from = time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 0, 0, 0, 0, lastAggregation.Location())
	}

	for from.Before(end) && to.Before(end) {
		to = from.Add(24 * time.Hour)
		jobs <- &AggregationJob{userId, from, to}
		from = to
	}
}

func getStartOfToday() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 1, now.Location())
}
