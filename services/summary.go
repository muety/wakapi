package services

import (
	"crypto/md5"
	"errors"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
	"github.com/patrickmn/go-cache"
	"math"
	"sort"
	"time"
)

const HeartbeatDiffThreshold = 2 * time.Minute

type SummaryService struct {
	config           *config.Config
	cache            *cache.Cache
	repository       repositories.ISummaryRepository
	heartbeatService IHeartbeatService
	aliasService     IAliasService
}

type SummaryRetriever func(f, t time.Time, u *models.User) (*models.Summary, error)

func NewSummaryService(summaryRepo repositories.ISummaryRepository, heartbeatService IHeartbeatService, aliasService IAliasService) *SummaryService {
	return &SummaryService{
		config:           config.Get(),
		cache:            cache.New(24*time.Hour, 24*time.Hour),
		repository:       summaryRepo,
		heartbeatService: heartbeatService,
		aliasService:     aliasService,
	}
}

// Public summary generation methods

func (srv *SummaryService) Aliased(from, to time.Time, user *models.User, f SummaryRetriever) (*models.Summary, error) {
	// Check cache
	cacheKey := srv.getHash(from.String(), to.String(), user.ID, "--aliased")
	if cacheResult, ok := srv.cache.Get(cacheKey); ok {
		return cacheResult.(*models.Summary), nil
	}

	// Wrap alias resolution
	resolve := func(t uint8, k string) string {
		s, _ := srv.aliasService.GetAliasOrDefault(user.ID, t, k)
		return s
	}

	// Initialize alias resolver service
	if err := srv.aliasService.LoadUserAliases(user.ID); err != nil {
		return nil, err
	}

	// Get actual summary
	s, err := f(from, to, user)
	if err != nil {
		return nil, err
	}

	// Post-process summary and cache it
	summary := s.WithResolvedAliases(resolve)
	srv.cache.SetDefault(cacheKey, summary)
	return summary.Sorted(), nil
}

func (srv *SummaryService) Retrieve(from, to time.Time, user *models.User) (*models.Summary, error) {
	// Check cache
	cacheKey := srv.getHash(from.String(), to.String(), user.ID, "--aliased")
	if cacheResult, ok := srv.cache.Get(cacheKey); ok {
		return cacheResult.(*models.Summary), nil
	}

	// Get all already existing, pre-generated summaries that fall into the requested interval
	summaries, err := srv.repository.GetByUserWithin(user, from, to)
	if err != nil {
		return nil, err
	}

	// Generate missing slots (especially before and after existing summaries) from raw heartbeats
	missingIntervals := srv.getMissingIntervals(from, to, summaries)
	for _, interval := range missingIntervals {
		if s, err := srv.Summarize(interval.Start, interval.End, user); err == nil {
			summaries = append(summaries, s)
		} else {
			return nil, err
		}
	}

	// Merge existing and newly generated summary snippets
	summary, err := srv.mergeSummaries(summaries)
	if err != nil {
		return nil, err
	}

	// Cache 'em
	srv.cache.SetDefault(cacheKey, summary)
	return summary.Sorted(), nil
}

func (srv *SummaryService) Summarize(from, to time.Time, user *models.User) (*models.Summary, error) {
	// Initialize and fetch data
	var heartbeats models.Heartbeats
	if rawHeartbeats, err := srv.heartbeatService.GetAllWithin(from, to, user); err == nil {
		heartbeats = rawHeartbeats
	} else {
		return nil, err
	}

	types := models.SummaryTypes()

	typedAggregations := make(chan models.SummaryItemContainer)
	defer close(typedAggregations)
	for _, t := range types {
		go srv.aggregateBy(heartbeats, t, typedAggregations)
	}

	// Aggregate raw heartbeats by types in parallel and collect them
	var projectItems []*models.SummaryItem
	var languageItems []*models.SummaryItem
	var editorItems []*models.SummaryItem
	var osItems []*models.SummaryItem
	var machineItems []*models.SummaryItem

	for i := 0; i < len(types); i++ {
		item := <-typedAggregations
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

	if heartbeats.Len() > 0 {
		from = time.Time(heartbeats.First().Time)
		to = time.Time(heartbeats.Last().Time)
	}

	summary := &models.Summary{
		UserID:           user.ID,
		FromTime:         models.CustomTime(from),
		ToTime:           models.CustomTime(to),
		Projects:         projectItems,
		Languages:        languageItems,
		Editors:          editorItems,
		OperatingSystems: osItems,
		Machines:         machineItems,
	}

	//summary.FillUnknown()

	return summary.Sorted(), nil
}

// CRUD methods

func (srv *SummaryService) GetLatestByUser() ([]*models.TimeByUser, error) {
	return srv.repository.GetLastByUser()
}

func (srv *SummaryService) DeleteByUser(userId string) error {
	return srv.repository.DeleteByUser(userId)
}

func (srv *SummaryService) Insert(summary *models.Summary) error {
	return srv.repository.Insert(summary)
}

// Private summary generation and utility methods

func (srv *SummaryService) aggregateBy(heartbeats []*models.Heartbeat, summaryType uint8, c chan models.SummaryItemContainer) {
	durations := make(map[string]time.Duration)

	for i, h := range heartbeats {
		key := h.GetKey(summaryType)

		if _, ok := durations[key]; !ok {
			durations[key] = time.Duration(0)
		}

		if i == 0 {
			continue
		}

		t1, t2, tdiff := h.Time.T(), heartbeats[i-1].Time.T(), time.Duration(0)
		// This is a hack. The time difference between two heartbeats from two subsequent day (e.g. 23:59:59 and 00:00:01) are ignored.
		// This is to prevent a discrepancy between summaries computed solely from heartbeats and summaries involving pre-aggregated per-day summaries.
		// For the latter, a duration is already pre-computed and information about individual heartbeats is lost, so there can be no cross-day overflow.
		// Essentially, we simply ignore such edge-case heartbeats here, which makes the eventual total duration potentially a bit shorter.
		if t1.Day() == t2.Day() {
			timePassed := t1.Sub(t2)
			tdiff = time.Duration(int64(math.Min(float64(timePassed), float64(HeartbeatDiffThreshold))))
		}
		durations[key] += tdiff
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

func (srv *SummaryService) mergeSummaries(summaries []*models.Summary) (*models.Summary, error) {
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

		if s.FromTime.T().Before(minTime) {
			minTime = s.FromTime.T()
		}

		if s.ToTime.T().After(maxTime) {
			maxTime = s.ToTime.T()
		}

		finalSummary.Projects = srv.mergeSummaryItems(finalSummary.Projects, s.Projects)
		finalSummary.Languages = srv.mergeSummaryItems(finalSummary.Languages, s.Languages)
		finalSummary.Editors = srv.mergeSummaryItems(finalSummary.Editors, s.Editors)
		finalSummary.OperatingSystems = srv.mergeSummaryItems(finalSummary.OperatingSystems, s.OperatingSystems)
		finalSummary.Machines = srv.mergeSummaryItems(finalSummary.Machines, s.Machines)
	}

	finalSummary.FromTime = models.CustomTime(minTime)
	finalSummary.ToTime = models.CustomTime(maxTime)

	return finalSummary, nil
}

func (srv *SummaryService) mergeSummaryItems(existing []*models.SummaryItem, new []*models.SummaryItem) []*models.SummaryItem {
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

func (srv *SummaryService) getMissingIntervals(from, to time.Time, summaries []*models.Summary) []*models.Interval {
	if len(summaries) == 0 {
		return []*models.Interval{{from, to}}
	}

	intervals := make([]*models.Interval, 0)

	// Pre
	if from.Before(summaries[0].FromTime.T()) {
		intervals = append(intervals, &models.Interval{from, summaries[0].FromTime.T()})
	}

	// Between
	for i := 0; i < len(summaries)-1; i++ {
		t1, t2 := summaries[i].ToTime.T(), summaries[i+1].FromTime.T()
		if t1.Equal(t2) {
			continue
		}

		// round to end of day / start of day, assuming that summaries are always generated on a per-day basis
		td1 := time.Date(t1.Year(), t1.Month(), t1.Day()+1, 0, 0, 0, 0, t1.Location())
		td2 := time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, t2.Location())
		// one or more day missing in between?
		if td1.Before(td2) {
			intervals = append(intervals, &models.Interval{summaries[i].ToTime.T(), summaries[i+1].FromTime.T()})
		}
	}

	// Post
	if to.After(summaries[len(summaries)-1].ToTime.T()) {
		intervals = append(intervals, &models.Interval{summaries[len(summaries)-1].ToTime.T(), to})
	}

	return intervals
}

func (srv *SummaryService) getHash(args ...string) string {
	digest := md5.New()
	for _, a := range args {
		digest.Write([]byte(a))
	}
	return string(digest.Sum(nil))
}
