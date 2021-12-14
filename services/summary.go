package services

import (
	"errors"
	"fmt"
	"github.com/emvi/logbuch"
	"github.com/leandro-lugaresi/hub"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
	"github.com/patrickmn/go-cache"
	"sort"
	"strings"
	"time"
)

type SummaryService struct {
	config              *config.Config
	cache               *cache.Cache
	eventBus            *hub.Hub
	repository          repositories.ISummaryRepository
	durationService     IDurationService
	aliasService        IAliasService
	projectLabelService IProjectLabelService
}

type SummaryRetriever func(f, t time.Time, u *models.User) (*models.Summary, error)

func NewSummaryService(summaryRepo repositories.ISummaryRepository, durationService IDurationService, aliasService IAliasService, projectLabelService IProjectLabelService) *SummaryService {
	srv := &SummaryService{
		config:              config.Get(),
		cache:               cache.New(24*time.Hour, 24*time.Hour),
		eventBus:            config.EventBus(),
		repository:          summaryRepo,
		durationService:     durationService,
		aliasService:        aliasService,
		projectLabelService: projectLabelService,
	}

	sub1 := srv.eventBus.Subscribe(0, config.TopicProjectLabel)
	go func(sub *hub.Subscription) {
		for m := range sub.Receiver {
			userId := m.Fields[config.FieldUserId].(string)
			for key := range srv.cache.Items() {
				if strings.HasSuffix(key, fmt.Sprintf("__%s__--aliased", userId)) {
					srv.cache.Delete(key)
				}
			}
		}
	}(&sub1)

	return srv
}

// Public summary generation methods

// Aliased retrieves or computes a new summary based on the given SummaryRetriever and augments it with entity aliases and project labels
func (srv *SummaryService) Aliased(from, to time.Time, user *models.User, f SummaryRetriever, skipCache bool) (*models.Summary, error) {
	// Check cache
	cacheKey := srv.getHash(from.String(), to.String(), user.ID, "--aliased")
	if cacheResult, ok := srv.cache.Get(cacheKey); ok && !skipCache {
		return cacheResult.(*models.Summary), nil
	}

	// Wrap alias resolution
	resolve := func(t uint8, k string) string {
		s, _ := srv.aliasService.GetAliasOrDefault(user.ID, t, k)
		return s
	}

	// Initialize alias resolver service
	if err := srv.aliasService.InitializeUser(user.ID); err != nil {
		return nil, err
	}

	// Get actual summary
	s, err := f(from, to, user)
	if err != nil {
		return nil, err
	}

	// Post-process summary and cache it
	summary := s.WithResolvedAliases(resolve)
	summary = srv.withProjectLabels(summary)
	summary.FillBy(models.SummaryProject, models.SummaryLabel) // first fill up labels from projects
	summary.FillMissing()                                      // then, full up types which are entirely missing

	srv.cache.SetDefault(cacheKey, summary)
	return summary.Sorted(), nil
}

func (srv *SummaryService) Retrieve(from, to time.Time, user *models.User) (*models.Summary, error) {
	// Get all already existing, pre-generated summaries that fall into the requested interval
	summaries, err := srv.repository.GetByUserWithin(user, from, to)
	if err != nil {
		return nil, err
	}

	// Generate missing slots (especially before and after existing summaries) from durations (formerly raw heartbeats)
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

	return summary.Sorted(), nil
}

func (srv *SummaryService) Summarize(from, to time.Time, user *models.User) (*models.Summary, error) {
	// Initialize and fetch data
	var durations models.Durations
	if result, err := srv.durationService.Get(from, to, user); err == nil {
		durations = result
	} else {
		return nil, err
	}

	types := models.NativeSummaryTypes()

	typedAggregations := make(chan models.SummaryItemContainer)
	defer close(typedAggregations)
	for _, t := range types {
		go srv.aggregateBy(durations, t, typedAggregations)
	}

	// Aggregate durations (formerly raw heartbeats) by types in parallel and collect them
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

	if durations.Len() > 0 {
		from = time.Time(durations.First().Time)
		to = time.Time(durations.Last().Time)
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

	return summary.Sorted(), nil
}

// CRUD methods

func (srv *SummaryService) GetLatestByUser() ([]*models.TimeByUser, error) {
	return srv.repository.GetLastByUser()
}

func (srv *SummaryService) DeleteByUser(userId string) error {
	srv.invalidateUserCache(userId)
	return srv.repository.DeleteByUser(userId)
}

func (srv *SummaryService) Insert(summary *models.Summary) error {
	srv.invalidateUserCache(summary.UserID)
	return srv.repository.Insert(summary)
}

// Private summary generation and utility methods

func (srv *SummaryService) aggregateBy(durations []*models.Duration, summaryType uint8, c chan models.SummaryItemContainer) {
	mapping := make(map[string]time.Duration)

	for _, d := range durations {
		mapping[d.GetKey(summaryType)] += d.Duration
	}

	items := make([]*models.SummaryItem, 0)
	for k, v := range mapping {
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

func (srv *SummaryService) withProjectLabels(summary *models.Summary) *models.Summary {
	newEntry := func(key string, total time.Duration) *models.SummaryItem {
		return &models.SummaryItem{
			Type:  models.SummaryLabel,
			Key:   key,
			Total: total,
		}
	}

	allLabels, err := srv.projectLabelService.GetByUser(summary.UserID)
	if err != nil {
		logbuch.Error("failed to retrieve project labels for user summary ('%s', '%s', '%s')", summary.UserID, summary.FromTime.String(), summary.ToTime.String())
		return summary
	}

	mappedProjects := make(map[string]*models.SummaryItem, len(summary.Projects))
	for _, p := range summary.Projects {
		mappedProjects[p.Key] = p
	}

	var totalLabelTime time.Duration
	labelMap := make(map[string]*models.SummaryItem, 0)
	for _, l := range allLabels {
		if p, ok := mappedProjects[l.ProjectKey]; ok {
			if _, ok2 := labelMap[l.Label]; !ok2 {
				labelMap[l.Label] = newEntry(l.Label, 0)
			}
			labelMap[l.Label].Total += p.Total
			totalLabelTime += p.Total
		}
	}
	//labelMap[models.DefaultProjectLabel] = newEntry(models.DefaultProjectLabel, summary.TotalTimeBy(models.SummaryProject) / time.Second-totalLabelTime)

	labels := make([]*models.SummaryItem, 0, len(labelMap))
	for _, v := range labelMap {
		if v.Total > 0 {
			labels = append(labels, v)
		}
	}
	summary.Labels = labels
	return summary
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
		Labels:           make([]*models.SummaryItem, 0),
	}

	var processed = map[time.Time]bool{}

	for _, s := range summaries {
		hash := s.FromTime.T()
		if _, found := processed[hash]; found {
			logbuch.Warn("summary from %v to %v (user '%s') was attempted to be processed more often than once", s.FromTime, s.ToTime, s.UserID)
			continue
		}

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
		finalSummary.Labels = srv.mergeSummaryItems(finalSummary.Labels, s.Labels)

		processed[hash] = true
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
		// we assume that, if summary for any time range within a day is present, no further heartbeats exist on that day before 'from' and after 'to' time of that summary
		// this requires that a summary exists for every single day in a year and none is skipped, which shouldn't ever happen
		td1 := time.Date(t1.Year(), t1.Month(), t1.Day()+1, 0, 0, 0, 0, t1.Location())
		td2 := time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, t2.Location())

		// we always want to jump to beginning of next day
		// however, if left summary ends already at midnight, we would instead jump to beginning of second-next day -> go back again
		if td1.Sub(t1) == 24*time.Hour {
			td1 = td1.Add(-1 * time.Hour)
		}

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
	return strings.Join(args, "__")
}

func (srv *SummaryService) invalidateUserCache(userId string) {
	for key := range srv.cache.Items() {
		if strings.Contains(key, userId) {
			srv.cache.Delete(key)
		}
	}
}
