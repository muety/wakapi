package services

import (
	"errors"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/becheran/wildmatch-go"
	"github.com/duke-git/lancet/v2/condition"
	"github.com/duke-git/lancet/v2/datetime"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/leandro-lugaresi/hub"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/models/types"
	"github.com/muety/wakapi/repositories"
	"github.com/patrickmn/go-cache"
)

type SummaryService struct {
	config              *config.Config
	cache               *cache.Cache
	eventBus            *hub.Hub
	repository          repositories.ISummaryRepository
	heartbeatService    IHeartbeatService
	durationService     IDurationService
	aliasService        IAliasService
	projectLabelService IProjectLabelService
}

func NewSummaryService(summaryRepo repositories.ISummaryRepository, heartbeatService IHeartbeatService, durationService IDurationService, aliasService IAliasService, projectLabelService IProjectLabelService) *SummaryService {
	srv := &SummaryService{
		config:              config.Get(),
		cache:               cache.New(24*time.Hour, 24*time.Hour),
		eventBus:            config.EventBus(),
		repository:          summaryRepo,
		heartbeatService:    heartbeatService,
		durationService:     durationService,
		aliasService:        aliasService,
		projectLabelService: projectLabelService,
	}

	sub1 := srv.eventBus.Subscribe(0, config.TopicProjectLabel)
	go func(sub *hub.Subscription) {
		for m := range sub.Receiver {
			srv.invalidateUserCache(m.Fields[config.FieldUserId].(string))
		}
	}(&sub1)

	return srv
}

// Public summary generation methods

// Aliased retrieves or computes a new summary based on the given SummaryRetriever and augments it with entity aliases and project labels
func (srv *SummaryService) Aliased(from, to time.Time, user *models.User, f types.SummaryRetriever, filters *models.Filters, customTimeout *time.Duration, skipCache bool) (*models.Summary, error) {
	requestedTimeout := getEffectiveTimeout(user, customTimeout)

	// Check cache (or skip for sub second-level date precision)
	cacheKey := srv.getHash(from.String(), to.String(), user.ID, filters.Hash(), strconv.Itoa(int(requestedTimeout)), "--aliased")
	if to.Truncate(time.Second).Equal(to) && from.Truncate(time.Second).Equal(from) {
		if cacheResult, ok := srv.cache.Get(cacheKey); ok && !skipCache {
			return cacheResult.(*models.Summary).Sorted().InTZ(user.TZ()), nil
		}
	}

	// Resolver functions
	resolveAliases := srv.getAliasResolver(user)
	resolveAliasesReverse := srv.getAliasReverseResolver(user)
	resolveProjectLabelsReverse := srv.getProjectLabelsReverseResolver(user)

	// Post-process filters
	if filters != nil {
		if !filters.Project.Exists() {
			filters = filters.WithProjectLabels(resolveProjectLabelsReverse)
		}
		filters = filters.WithAliases(resolveAliasesReverse)
	}

	// Initialize alias resolver service
	if err := srv.aliasService.InitializeUser(user.ID); err != nil {
		return nil, err
	}

	// Get actual summary
	s, err := f(from, to, user, filters, customTimeout)
	if err != nil {
		return nil, err
	}

	// Post-process summary and cache it
	summary := s.WithResolvedAliases(resolveAliases)
	summary = srv.withProjectLabels(summary)
	summary.FillBy(models.SummaryProject, models.SummaryLabel) // first fill up labels from projects
	summary.FillMissing()                                      // then, full up types which are entirely missing

	if withDetails := filters != nil && filters.IsProjectDetails(); !withDetails {
		summary.Branches = nil
		summary.Entities = nil
	}

	srv.cache.SetDefault(cacheKey, summary)
	return summary.Sorted().InTZ(user.TZ()), nil
}

func (srv *SummaryService) Retrieve(from, to time.Time, user *models.User, filters *models.Filters, customTimeout *time.Duration) (*models.Summary, error) {
	summaries := make([]*models.Summary, 0)
	requestedTimeout := getEffectiveTimeout(user, customTimeout)

	// Filtered summaries or summaries at alternative timeouts are not persisted currently
	// Special case: if (a) filters apply to only one entity type and (b) we're only interested in the summary items of that particular entity type,
	// we can still fetch the persisted summary and drop all irrelevant parts from it
	requiresFiltering := filters != nil && !filters.IsEmpty() && (filters.CountDistinctTypes() > 1 || !filters.SelectFilteredOnly)
	mustRecompute := requiresFiltering || requestedTimeout != user.HeartbeatsTimeout()

	if !mustRecompute {
		// Get all already existing, pre-generated summaries that fall into the requested interval
		result, err := srv.repository.GetByUserWithin(user, from, to)
		if err == nil {
			summaries = srv.fixZeroDuration(result)
		} else {
			return nil, err
		}
	}

	// Generate missing slots (especially before and after existing summaries) from durations (formerly raw heartbeats)
	missingIntervals := srv.getMissingIntervals(from, to, summaries, false)
	for _, interval := range missingIntervals {
		if s, err := srv.Summarize(interval.Start, interval.End, user, filters, customTimeout); err == nil {
			if len(missingIntervals) > 2 && s.FromTime.T().Equal(s.ToTime.T()) {
				// little hack here: GetWithin will query for >= from_date
				// however, for "in-between" / intra-day missing intervals, we want strictly > from_date to prevent double-counting
				// to not have to rewrite many interfaces, we skip these summaries here
				continue
			}
			summaries = append(summaries, s)
		} else {
			return nil, err
		}
	}

	// Merge existing and newly generated summary snippets
	sort.Sort(models.Summaries(summaries))
	summary, err := srv.mergeSummaries(summaries)
	if err != nil {
		return nil, err
	}

	// prevent 0001-01-01T00:00:00 caused by empty "pre" missing interval, see https://github.com/muety/wakapi/issues/843
	summary.FromTime = models.CustomTime(condition.Ternary(summary.FromTime.T().Before(from), from, summary.FromTime.T()))

	if summary.TotalTime() == 0 {
		summary.FromTime = models.CustomTime(from)
		summary.ToTime = models.CustomTime(to)
	}

	if filters != nil && filters.CountDistinctTypes() == 1 && filters.SelectFilteredOnly {
		filter := filters.OneOrEmpty()
		summary.KeepOnly(map[uint8]bool{filter.Entity: true}).ApplyFilter(filter)
	}

	return summary.Sorted().InTZ(user.TZ()), nil
}

func (srv *SummaryService) Summarize(from, to time.Time, user *models.User, filters *models.Filters, customTimeout *time.Duration) (*models.Summary, error) {
	// Initialize and fetch data
	durations, err := srv.durationService.Get(from, to, user, filters, customTimeout, false)
	if err != nil {
		return nil, err
	}

	types := models.PersistedSummaryTypes()
	if filters != nil && filters.IsProjectDetails() {
		types = append(types, models.SummaryBranch)
		types = append(types, models.SummaryEntity)
	}

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
	var branchItems []*models.SummaryItem
	var entityItems []*models.SummaryItem
	var categoryItems []*models.SummaryItem

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
		case models.SummaryBranch:
			branchItems = item.Items
		case models.SummaryEntity:
			entityItems = item.Items
		case models.SummaryCategory:
			categoryItems = item.Items
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
		Branches:         branchItems,
		Entities:         entityItems,
		Categories:       categoryItems,
		NumHeartbeats:    durations.TotalNumHeartbeats(),
	}

	return summary.Sorted().InTZ(user.TZ()), nil
}

// CRUD methods

func (srv *SummaryService) GetLatestByUser() ([]*models.TimeByUser, error) {
	return srv.repository.GetLastByUser()
}

func (srv *SummaryService) DeleteByUser(userId string) error {
	srv.invalidateUserCache(userId)
	return srv.repository.DeleteByUser(userId)
}

func (srv *SummaryService) DeleteByUserBefore(userId string, t time.Time) error {
	srv.invalidateUserCache(userId)
	return srv.repository.DeleteByUserBefore(userId, t)
}

func (srv *SummaryService) Insert(summary *models.Summary) error {
	srv.invalidateUserCache(summary.UserID)
	return srv.repository.InsertWithRetry(summary)
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
		config.Log().Error("failed to retrieve project labels for user summary", "userID", summary.UserID, "fromTime", summary.FromTime.String(), "toTime", summary.ToTime.String())
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
	// labelMap[models.DefaultProjectLabel] = newEntry(models.DefaultProjectLabel, summary.TotalTimeBy(models.SummaryProject) / time.Second-totalLabelTime)

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
	// summaries must be sorted by from_date
	// also, this function implicitly assumes summaries are distinct, i.e. don't cover overlapping time intervals
	// if they do, activity within the overlap would be counted double

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
		Branches:         make([]*models.SummaryItem, 0),
		Entities:         make([]*models.SummaryItem, 0),
		Categories:       make([]*models.SummaryItem, 0),
	}

	var processed = map[time.Time]*models.Summary{}

	for i, s := range summaries {
		hash := s.FromTime.T()
		if s2, found := processed[hash]; found {
			if !s.ToTime.T().Equal(s2.ToTime.T()) {
				// TODO: heuristic for which one to use (more recent one? larger interval? more heartbeats included?)
				slog.Warn("got multiple summaries for same start date but different intervals", "id1", s.ID, "id2", s2.ID, "fromTime1", s.FromTime.T(), "fromTime2", s2.FromTime.T(), "toTime1", s.ToTime.T(), "toTime2", s2.ToTime.T(), "userID", s.UserID)
			}
			continue
		}

		if i > 0 {
			if prev := summaries[i-1]; s.FromTime.T().Before(prev.ToTime.T()) {
				slog.Warn("got overlapping summaries for user", "prevID", prev.ID, "currentID", s.ID, "userID", s.UserID, "fromTime", s.FromTime.T(), "prevToTime", prev.ToTime.T())
			}
		}

		if s.UserID != finalSummary.UserID {
			return nil, errors.New("users don't match")
		}

		totalTime := s.TotalTime()

		if s.FromTime.T().Before(minTime) && totalTime > 0 { // only consider non-empty summaries
			minTime = s.FromTime.T()
		}
		if s.ToTime.T().After(maxTime) && totalTime > 0 { // only consider non-empty summaries
			maxTime = s.ToTime.T()
		}

		finalSummary.Projects = srv.mergeSummaryItems(finalSummary.Projects, s.Projects)
		finalSummary.Languages = srv.mergeSummaryItems(finalSummary.Languages, s.Languages)
		finalSummary.Editors = srv.mergeSummaryItems(finalSummary.Editors, s.Editors)
		finalSummary.OperatingSystems = srv.mergeSummaryItems(finalSummary.OperatingSystems, s.OperatingSystems)
		finalSummary.Machines = srv.mergeSummaryItems(finalSummary.Machines, s.Machines)
		finalSummary.Labels = srv.mergeSummaryItems(finalSummary.Labels, s.Labels)
		finalSummary.Branches = srv.mergeSummaryItems(finalSummary.Branches, s.Branches)
		finalSummary.Entities = srv.mergeSummaryItems(finalSummary.Entities, s.Entities)
		finalSummary.Categories = srv.mergeSummaryItems(finalSummary.Categories, s.Categories)
		finalSummary.NumHeartbeats += s.NumHeartbeats

		processed[hash] = s
	}

	finalSummary.FromTime = models.CustomTime(minTime)
	finalSummary.ToTime = models.CustomTime(condition.Ternary(maxTime.Before(minTime), minTime, maxTime))

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

func (srv *SummaryService) getMissingIntervals(from, to time.Time, summaries []*models.Summary, precise bool) []*models.Interval {
	if len(summaries) == 0 {
		return []*models.Interval{{from, to}}
	}

	intervals := make([]*models.Interval, 0)

	// Pre
	if from.Before(summaries[0].FromTime.T()) {
		intervals = append(intervals, &models.Interval{Start: from, End: summaries[0].FromTime.T()})
	}

	// Between
	for i := 0; i < len(summaries)-1; i++ {
		t1, t2 := summaries[i].ToTime.T(), summaries[i+1].FromTime.T()
		if t1.Equal(t2) || t1.Equal(to) || t1.After(to) {
			continue
		}

		td1 := t1
		td2 := t2

		// round to end of day / start of day, assuming that summaries are always generated on a per-day basis
		// we assume that, if summary for any time range within a day is present, no further heartbeats exist on that day before 'from' and after 'to' time of that summary
		// this requires that a summary exists for every single day in a year and none is skipped, which shouldn't ever happen
		// non-precise mode is mainly for speed when fetching summaries over large intervals and trades speed for summary accuracy / comprehensiveness
		if !precise {
			td1 = datetime.BeginOfDay(t1).AddDate(0, 0, 1)
			td2 = datetime.BeginOfDay(t2)

			// we always want to jump to beginning of next day
			// however, if left summary ends already at midnight, we would instead jump to beginning of second-next day -> go back again
			if td1.AddDate(0, 0, 1).Equal(t1) {
				td1 = td1.Add(-1 * time.Hour)
			}
		}

		// one or more day missing in between?
		if td1.Before(td2) {
			intervals = append(intervals, &models.Interval{Start: t1, End: t2})
		}
	}

	// Post
	if to.After(summaries[len(summaries)-1].ToTime.T()) {
		intervals = append(intervals, &models.Interval{Start: summaries[len(summaries)-1].ToTime.T(), End: to})
	}

	return intervals
}

// Since summary timestamps are only second-level precision, we rarely observe examples where from- and to-time are allegedly equal.
// We artificially modify those to give them a one second duration and potentially fix the subsequent summary as well to prevent overlaps.
// Assumes summaries slice to be sorted by from time.
func (s *SummaryService) fixZeroDuration(summaries []*models.Summary) []*models.Summary {
	for i, summary := range summaries {
		if summary.FromTime.T().Equal(summary.ToTime.T()) {
			summary.ToTime = models.CustomTime(summary.ToTime.T().Add(1 * time.Second))

			if i < len(summaries)-1 {
				summaryNext := summaries[i+1]
				if summaryNext.FromTime.T().Before(summary.ToTime.T()) {
					// intentionally not trying to resolve larger overlaps that were there before (even though they shouldn't happen in theory)
					summaryNext.FromTime = models.CustomTime(summaryNext.FromTime.T().Add(1 * time.Second))
				}
			}
		}
	}

	return summaries
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

func (srv *SummaryService) getAliasResolver(user *models.User) models.AliasResolver {
	return func(t uint8, k string) string {
		s, _ := srv.aliasService.GetAliasOrDefault(user.ID, t, k)
		return s
	}
}

func (srv *SummaryService) getAliasReverseResolver(user *models.User) models.AliasReverseResolver {
	return func(t uint8, k string) []string {
		aliases, err := srv.aliasService.GetByUserAndKeyAndType(user.ID, k, t)
		if err != nil {
			config.Log().Error("failed to fetch aliases for user", "user", user.ID, "error", err)
			aliases = []*models.Alias{}
		}

		projects, err := srv.heartbeatService.GetEntitySetByUser(models.SummaryProject, user.ID)
		if err != nil {
			config.Log().Error("failed to fetch projects for alias resolution for user", "user", user.ID, "error", err)
		}

		isWildcard := func(alias string) bool {
			return strings.Contains(alias, "*") || strings.Contains(alias, "?")
		}

		// for wildcard patterns like "anchr-" (e.g. resolving to "anchr-mobile", "anchr-web", ...), we need to fetch all projects matching the pattern
		// this is mainly used for the filtering functionality
		// proper way would be to make the filters support wildcards as well instead
		matchProjects := func(aliasWildcard string) []string {
			pattern := wildmatch.NewWildMatch(aliasWildcard)
			return slice.Filter[string](projects, func(i int, project string) bool {
				return pattern.IsMatch(project)
			})
		}

		aliasStrings := make([]string, 0, len(aliases))
		for _, a := range aliases {
			if isWildcard(a.Value) {
				aliasStrings = append(aliasStrings, matchProjects(a.Value)...)
			} else {
				aliasStrings = append(aliasStrings, a.Value)
			}
		}
		return aliasStrings
	}
}

func (srv *SummaryService) getProjectLabelsReverseResolver(user *models.User) models.ProjectLabelReverseResolver {
	return func(k string) []string {
		var labels []*models.ProjectLabel
		allLabels, err := srv.projectLabelService.GetByUserGroupedInverted(user.ID)
		if err == nil {
			if l, ok := allLabels[k]; ok {
				labels = l
			}
		}
		projectStrings := make([]string, 0, len(labels))
		for _, l := range labels {
			projectStrings = append(projectStrings, l.ProjectKey)
		}
		return projectStrings
	}
}
