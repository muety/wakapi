package services

import (
	"log"
	"math"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/n1try/wakapi/models"
)

type AggregationService struct {
	Config           *models.Config
	Db               *gorm.DB
	HeartbeatService *HeartbeatService
}

func (srv *AggregationService) SaveAggregation(aggregation *models.Aggregation) error {
	if err := srv.Db.Save(aggregation).Error; err != nil {
		return err
	}
	return nil
}

func (srv *AggregationService) DeleteAggregations(from, to time.Time) error {
	// TODO
	return nil
}

func (srv *AggregationService) FindOrAggregate(from, to time.Time, user *models.User) ([]*models.Aggregation, error) {
	var existingAggregations []*models.Aggregation
	if err := srv.Db.
		Where(&models.Aggregation{UserID: user.ID}).
		Where("from_time <= ?", from).
		Where("to_time <= ?", to).
		Order("to_time desc").
		Limit(models.NAggregationTypes).
		Find(&existingAggregations).Error; err != nil {
		return nil, err
	}

	maxTo := getMaxTo(existingAggregations)

	if len(existingAggregations) == 0 {
		newAggregations := srv.aggregate(from, to, user)
		for i := 0; i < len(newAggregations); i++ {
			srv.SaveAggregation(newAggregations[i])
		}
		return newAggregations, nil
	} else if maxTo.Before(to) {
		// newAggregations := srv.aggregate(maxTo, to, user)
		// TODO: compute aggregation(s) for remaining heartbeats
		// TODO: if these aggregations are more than 24h, save them
		// NOTE: never save aggregations that are less than 24h -> no need to delete some later
	} else if maxTo.Equal(to) {
		return existingAggregations, nil
	}

	// Should never occur
	return make([]*models.Aggregation, 0), nil
}

func (srv *AggregationService) aggregate(from, to time.Time, user *models.User) []*models.Aggregation {
	// TODO: Handle case that a time frame >= 24h is requested -> more than 4 will be returned
	types := []uint8{models.AggregationProject, models.AggregationLanguage, models.AggregationEditor, models.AggregationOS}
	heartbeats, err := srv.HeartbeatService.GetAllFrom(from, user)
	if err != nil {
		log.Fatal(err)
	}

	var aggregations []*models.Aggregation
	for _, t := range types {
		aggregation := &models.Aggregation{
			UserID:   user.ID,
			FromTime: &from,
			ToTime:   &to,
			Duration: to.Sub(from),
			Type:     t,
			Items:    srv.aggregateBy(heartbeats, t)[0:1], //make([]*models.AggregationItem, 0),
		}
		aggregations = append(aggregations, aggregation)
	}

	return aggregations
}

func (srv *AggregationService) aggregateBy(heartbeats []*models.Heartbeat, aggregationType uint8) []models.AggregationItem {
	durations := make(map[string]time.Duration)

	for i, h := range heartbeats {
		var key string
		switch aggregationType {
		case models.AggregationProject:
			key = h.Project
		case models.AggregationEditor:
			key = h.Editor
		case models.AggregationLanguage:
			key = h.Language
		case models.AggregationOS:
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

	var items []models.AggregationItem
	for k, v := range durations {
		items = append(items, models.AggregationItem{
			AggregationID: 9,
			Key:           k,
			Total:         models.ScannableDuration(v),
		})
	}

	return items
}

func (srv *AggregationService) MergeAggregations(aggregations []*models.Aggregation) []*models.Aggregation {
	// TODO
	return make([]*models.Aggregation, 0)
}

func getMaxTo(aggregations []*models.Aggregation) time.Time {
	var maxTo time.Time
	for i := 0; i < len(aggregations); i++ {
		agg := aggregations[i]
		if agg.ToTime.After(maxTo) {
			maxTo = *agg.ToTime
		}
	}
	return maxTo
}
