package services

import (
	"math"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/n1try/wakapi/models"
)

type SummaryService struct {
	Config           *models.Config
	Db               *gorm.DB
	HeartbeatService *HeartbeatService
}

func (srv *SummaryService) GetSummary(from, to time.Time, user *models.User) (*models.Summary, error) {
	heartbeats, err := srv.HeartbeatService.GetAllWithin(from, to, user)
	if err != nil {
		return nil, err
	}

	summary := &models.Summary{
		UserID:           user.ID,
		FromTime:         &from,
		ToTime:           &to,
		Projects:         srv.aggregateBy(heartbeats, models.SummaryProject),
		Languages:        srv.aggregateBy(heartbeats, models.SummaryLanguage),
		Editors:          srv.aggregateBy(heartbeats, models.SummaryEditor),
		OperatingSystems: srv.aggregateBy(heartbeats, models.SummaryOS),
	}

	return summary, nil
}

func (srv *SummaryService) aggregateBy(heartbeats []*models.Heartbeat, aggregationType uint8) []models.SummaryItem {
	durations := make(map[string]time.Duration)

	for i, h := range heartbeats {
		var key string
		switch aggregationType {
		case models.SummaryProject:
			key = h.Project
		case models.SummaryEditor:
			key = h.Editor
		case models.SummaryLanguage:
			key = h.Language
		case models.SummaryOS:
			key = h.OperatingSystem
		}

		if _, ok := durations[key]; !ok {
			durations[key] = time.Duration(0)
		}

		if i == 0 {
			continue
		}

		timePassed := h.Time.Time().Sub(heartbeats[i-1].Time.Time())
		timeThresholded := math.Min(float64(timePassed), float64(time.Duration(2)*time.Minute))
		durations[key] += time.Duration(int64(timeThresholded))
	}

	items := make([]models.SummaryItem, 0)
	for k, v := range durations {
		items = append(items, models.SummaryItem{
			Key:   k,
			Total: v / time.Second,
		})
	}

	return items
}
