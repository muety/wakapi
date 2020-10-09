package services

import (
	"crypto/md5"
	"errors"
	"github.com/muety/wakapi/config"
	"github.com/patrickmn/go-cache"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/muety/wakapi/models"
)

type SummaryService struct {
	Config           *config.Config
	Cache            *cache.Cache
	Db               *gorm.DB
	HeartbeatService *HeartbeatService
	AliasService     *AliasService
}

func NewSummaryService(db *gorm.DB, heartbeatService *HeartbeatService, aliasService *AliasService) *SummaryService {
	return &SummaryService{
		Config:           config.Get(),
		Cache:            cache.New(24*time.Hour, 24*time.Hour),
		Db:               db,
		HeartbeatService: heartbeatService,
		AliasService:     aliasService,
	}
}

type Interval struct {
	Start time.Time
	End   time.Time
}

// TODO: simplify!
func (srv *SummaryService) Construct(from, to time.Time, user *models.User, recompute bool) (*models.Summary, error) {
	var existingSummaries []*models.Summary
	var cacheKey string

	if recompute {
		existingSummaries = make([]*models.Summary, 0)
	} else {
		cacheKey = getHash([]time.Time{from, to}, user)
		if result, ok := srv.Cache.Get(cacheKey); ok {
			return result.(*models.Summary), nil
		}
		summaries, err := srv.GetByUserWithin(user, from, to)
		if err != nil {
			return nil, err
		}
		existingSummaries = summaries
	}

	missingIntervals := getMissingIntervals(from, to, existingSummaries)

	heartbeats := make([]*models.Heartbeat, 0)
	for _, interval := range missingIntervals {
		hb, err := srv.HeartbeatService.GetAllWithin(interval.Start, interval.End, user)
		if err != nil {
			return nil, err
		}
		heartbeats = append(heartbeats, hb...)
	}

	types := models.SummaryTypes()

	var projectItems []*models.SummaryItem
	var languageItems []*models.SummaryItem
	var editorItems []*models.SummaryItem
	var osItems []*models.SummaryItem
	var machineItems []*models.SummaryItem

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
		case models.SummaryMachine:
			machineItems = item.Items
		}
	}
	close(c)

	realFrom, realTo := from, to
	if len(existingSummaries) > 0 {
		realFrom = existingSummaries[0].FromTime
		realTo = existingSummaries[len(existingSummaries)-1].ToTime

		for _, summary := range existingSummaries {
			summary.FillUnknown()
		}
	}
	if len(heartbeats) > 0 {
		t1, t2 := time.Time(heartbeats[0].Time), time.Time(heartbeats[len(heartbeats)-1].Time)
		if t1.After(realFrom) && t1.Before(time.Date(realFrom.Year(), realFrom.Month(), realFrom.Day()+1, 0, 0, 0, 0, realFrom.Location())) {
			realFrom = t1
		}
		if t2.Before(realTo) && t2.After(time.Date(realTo.Year(), realTo.Month(), realTo.Day()-1, 0, 0, 0, 0, realTo.Location())) {
			realTo = t2
		}
	}

	aggregatedSummary := &models.Summary{
		UserID:           user.ID,
		FromTime:         realFrom,
		ToTime:           realTo,
		Projects:         projectItems,
		Languages:        languageItems,
		Editors:          editorItems,
		OperatingSystems: osItems,
		Machines:         machineItems,
	}

	allSummaries := []*models.Summary{aggregatedSummary}
	allSummaries = append(allSummaries, existingSummaries...)

	summary, err := mergeSummaries(allSummaries)
	if err != nil {
		return nil, err
	}

	if cacheKey != "" {
		srv.Cache.SetDefault(cacheKey, summary)
	}

	return summary, nil
}

func (srv *SummaryService) Insert(summary *models.Summary) error {
	if err := srv.Db.Create(summary).Error; err != nil {
		return err
	}
	return nil
}

func (srv *SummaryService) GetByUserWithin(user *models.User, from, to time.Time) ([]*models.Summary, error) {
	var summaries []*models.Summary
	if err := srv.Db.
		Where(&models.Summary{UserID: user.ID}).
		Where("from_time >= ?", from).
		Where("to_time <= ?", to).
		Order("from_time asc").
		Preload("Projects", "type = ?", models.SummaryProject).
		Preload("Languages", "type = ?", models.SummaryLanguage).
		Preload("Editors", "type = ?", models.SummaryEditor).
		Preload("OperatingSystems", "type = ?", models.SummaryOS).
		Preload("Machines", "type = ?", models.SummaryMachine).
		Find(&summaries).Error; err != nil {
		return nil, err
	}
	return summaries, nil
}

// Will return *models.Index objects with only user_id and to_time filled
func (srv *SummaryService) GetLatestByUser() ([]*models.Summary, error) {
	var summaries []*models.Summary
	if err := srv.Db.
		Table("summaries").
		Select("user_id, max(to_time) as to_time").
		Group("user_id").
		Scan(&summaries).Error; err != nil {
		return nil, err
	}
	return summaries, nil
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
		case models.SummaryMachine:
			key = h.Machine
		}

		if key == "" {
			key = models.UnknownSummaryKey
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

	items := make([]*models.SummaryItem, 0)
	for k, v := range durations {
		items = append(items, &models.SummaryItem{
			Key:   k,
			Total: v / time.Second,
			Type:  summaryType,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Total > items[j].Total
	})

	c <- models.SummaryItemContainer{Type: summaryType, Items: items}
}

func getMissingIntervals(from, to time.Time, existingSummaries []*models.Summary) []*Interval {
	if len(existingSummaries) == 0 {
		return []*Interval{{from, to}}
	}

	intervals := make([]*Interval, 0)

	// Pre
	if from.Before(existingSummaries[0].FromTime) {
		intervals = append(intervals, &Interval{from, existingSummaries[0].FromTime})
	}

	// Between
	for i := 0; i < len(existingSummaries)-1; i++ {
		t1, t2 := existingSummaries[i].ToTime, existingSummaries[i+1].FromTime
		if t1.Equal(t2) {
			continue
		}

		// round to end of day / start of day, assuming that summaries are always generated on a per-day basis
		td1 := time.Date(t1.Year(), t1.Month(), t1.Day()+1, 0, 0, 0, 0, t1.Location())
		td2 := time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, t2.Location())
		// one or more day missing in between?
		if td1.Before(td2) {
			intervals = append(intervals, &Interval{existingSummaries[i].ToTime, existingSummaries[i+1].FromTime})
		}
	}

	// Post
	if to.After(existingSummaries[len(existingSummaries)-1].ToTime) {
		intervals = append(intervals, &Interval{to, existingSummaries[len(existingSummaries)-1].ToTime})
	}

	return intervals
}

func mergeSummaries(summaries []*models.Summary) (*models.Summary, error) {
	if len(summaries) < 1 {
		return nil, errors.New("no summaries given")
	}

	var minTime, maxTime time.Time
	minTime = time.Now()

	finalSummary := &models.Summary{
		UserID:           summaries[0].UserID,
		Projects:         make([]*models.SummaryItem, 0),
		Languages:        make([]*models.SummaryItem, 0),
		Editors:          make([]*models.SummaryItem, 0),
		OperatingSystems: make([]*models.SummaryItem, 0),
		Machines:         make([]*models.SummaryItem, 0),
	}

	for _, s := range summaries {
		if s.UserID != finalSummary.UserID {
			return nil, errors.New("users don't match")
		}

		if s.FromTime.Before(minTime) {
			minTime = s.FromTime
		}

		if s.ToTime.After(maxTime) {
			maxTime = s.ToTime
		}

		finalSummary.Projects = mergeSummaryItems(finalSummary.Projects, s.Projects)
		finalSummary.Languages = mergeSummaryItems(finalSummary.Languages, s.Languages)
		finalSummary.Editors = mergeSummaryItems(finalSummary.Editors, s.Editors)
		finalSummary.OperatingSystems = mergeSummaryItems(finalSummary.OperatingSystems, s.OperatingSystems)
		finalSummary.Machines = mergeSummaryItems(finalSummary.Machines, s.Machines)
	}

	finalSummary.FromTime = minTime
	finalSummary.ToTime = maxTime

	return finalSummary, nil
}

func mergeSummaryItems(existing []*models.SummaryItem, new []*models.SummaryItem) []*models.SummaryItem {
	items := make(map[string]*models.SummaryItem)

	// Build map from existing
	for _, item := range existing {
		items[item.Key] = item
	}

	for _, item := range new {
		if it, ok := items[item.Key]; !ok {
			items[item.Key] = item
		} else {
			(*it).Total += item.Total
		}
	}

	var i int
	itemList := make([]*models.SummaryItem, len(items))
	for k, v := range items {
		itemList[i] = &models.SummaryItem{Key: k, Total: v.Total, Type: v.Type}
		i++
	}

	sort.Slice(itemList, func(i, j int) bool {
		return itemList[i].Total > itemList[j].Total
	})

	return itemList
}

func getHash(times []time.Time, user *models.User) string {
	digest := md5.New()
	for _, t := range times {
		digest.Write([]byte(strconv.Itoa(int(t.Unix()))))
	}
	digest.Write([]byte(user.ID))
	return string(digest.Sum(nil))
}
