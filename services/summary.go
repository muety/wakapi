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
	AliasService     *AliasService
}

func (srv *SummaryService) GetSummary(from, to time.Time, user *models.User) (*models.Summary, error) {
	heartbeats, err := srv.HeartbeatService.GetAllWithin(from, to, user)
	if err != nil {
		return nil, err
	}

	types := []uint8{models.SummaryProject, models.SummaryLanguage, models.SummaryEditor, models.SummaryOS}

	var projectItems []models.SummaryItem
	var languageItems []models.SummaryItem
	var editorItems []models.SummaryItem
	var osItems []models.SummaryItem

	if err := srv.AliasService.LoadUserAliases(user.ID); err != nil {
		return nil, err
	}

	c := make(chan models.SummaryItemContainer)
	for _, t := range types {
		go srv.aggregateBy(heartbeats, t, user, c)
	}

	for i := 0; i < len(types); i++ {
		item := <-c
		switch item.Type {
		case models.SummaryProject:
			projectItems = item.Items
		case models.SummaryLanguage:
			languageItems = item.Items
		case models.SummaryEditor:
			editorItems = item.Items
		case models.SummaryOS:
			osItems = item.Items
		}
	}
	close(c)

	summary := &models.Summary{
		UserID:           user.ID,
		FromTime:         &from,
		ToTime:           &to,
		Projects:         projectItems,
		Languages:        languageItems,
		Editors:          editorItems,
		OperatingSystems: osItems,
	}

	return summary, nil
}

func (srv *SummaryService) aggregateBy(heartbeats []*models.Heartbeat, summaryType uint8, user *models.User, c chan models.SummaryItemContainer) {
	durations := make(map[string]time.Duration)

	for i, h := range heartbeats {
		var key string
		switch summaryType {
		case models.SummaryProject:
			key = h.Project
		case models.SummaryEditor:
			key = h.Editor
		case models.SummaryLanguage:
			key = h.Language
		case models.SummaryOS:
			key = h.OperatingSystem
		}

		if key == "" {
			key = "unknown"
		}

		if aliasedKey, err := srv.AliasService.GetAliasOrDefault(user.ID, summaryType, key); err == nil {
			key = aliasedKey
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

	c <- models.SummaryItemContainer{Type: summaryType, Items: items}
}
