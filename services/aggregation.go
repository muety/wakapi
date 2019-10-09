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
	"container/list"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/n1try/wakapi/models"
)

type AggregationService struct {
	Config           *models.Config
	Db               *gorm.DB
	UserService      *UserService
	SummaryService   *SummaryService
	HeartbeatService *HeartbeatService
}

type AggregationJob struct {
	UserId string
	From   time.Time
	To     time.Time
}

// Use https://godoc.org/github.com/jasonlvhit/gocron to trigger jobs on a regular basis.
func (srv *AggregationService) Start(interval time.Duration) {
}

func (srv *AggregationService) generateJobs() (*list.List, error) {
	var aggregationJobs *list.List = list.New()

	users, err := srv.UserService.GetAll()
	if err != nil {
		return nil, err
	}

	latestSummaries, err := srv.SummaryService.GetLatestUserSummaries()
	if err != nil {
		return nil, err
	}

	userSummaryTimes := make(map[string]*time.Time)
	for _, s := range latestSummaries {
		userSummaryTimes[s.UserID] = s.ToTime
	}

	missingUserIds := make([]string, 0)
	for _, u := range users {
		if _, ok := userSummaryTimes[u.ID]; !ok {
			missingUserIds = append(missingUserIds, u.ID)
		}
	}

	firstHeartbeats, err := srv.HeartbeatService.GetFirstUserHeartbeats(missingUserIds)
	if err != nil {
		return nil, err
	}

	for id, t := range userSummaryTimes {
		var from time.Time
		if t.Hour() == 0 {
			from = *t
		} else {
			nextDay := t.Add(24 * time.Hour)
			from = time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 0, 0, 0, 0, t.Location())
		}

		aggregationJobs.PushBack(&AggregationJob{id, from, from.Add(24 * time.Hour)})
	}

	for _, h := range firstHeartbeats {
		var from time.Time
		var t time.Time = time.Time(*(h.Time))
		if t.Hour() == 0 {
			from = time.Time(*(h.Time))
		} else {
			nextDay := t.Add(24 * time.Hour)
			from = time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 0, 0, 0, 0, t.Location())
		}

		aggregationJobs.PushBack(&AggregationJob{h.UserID, from, from.Add(24 * time.Hour)})
	}

	return aggregationJobs, nil
}
