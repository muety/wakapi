package services

import (
	"errors"
	"sort"
	"strings"
	"time"

	"log/slog"

	"github.com/becheran/wildmatch-go"
	"github.com/duke-git/lancet/v2/datetime"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/leandro-lugaresi/hub"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
	summarytypes "github.com/muety/wakapi/types"
	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"
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

func NewSummaryService(db *gorm.DB) *SummaryService {
	summaryRepository := repositories.NewSummaryRepository(db)
	aliasService := NewAliasService(db)
	projectLabelService := NewProjectLabelService(db)
	heartbeatService := NewHeartbeatService(db)
	durationService := NewDurationService(db)

	srv := &SummaryService{
		config:              config.Get(),
		cache:               cache.New(24*time.Hour, 24*time.Hour),
		eventBus:            config.EventBus(),
		repository:          summaryRepository,
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

func NewTestSummaryService(summaryRepo repositories.ISummaryRepository, heartbeatService IHeartbeatService, durationService IDurationService, aliasService IAliasService, projectLabelService IProjectLabelService) *SummaryService {
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

// Generate is the main method that tells the complete story of summary generation
func (srv *SummaryService) Generate(request *summarytypes.SummaryRequest, options *summarytypes.ProcessingOptions) (*models.Summary, error) {
	// Step 1: Validate the request
	if err := request.Validate(); err != nil {
		return nil, err
	}

	// Step 2: Try to retrieve existing summaries
	summary, err := srv.RetrieveFromStorage(request)
	if err != nil || srv.needsComputation(summary, request) {
		// Step 3: Compute missing parts
		computed, err := srv.ComputeFromDurations(request)
		if err != nil {
			return nil, err
		}
		summary = srv.mergeSummaryWithComputed(summary, computed)
	}

	// Step 4: Apply enhancements based on options
	if options.ApplyAliases || options.ApplyProjectLabels {
		summary = srv.applyEnhancements(summary, request.User, options)
	}

	// Step 5: Cache if needed
	if options.UseCache && !request.SkipCache {
		srv.cacheResult(summary, request)
	}

	return summary, nil
}

// QuickSummary provides a simple way to get a summary with default settings
func (srv *SummaryService) QuickSummary(from, to time.Time, user *models.User) (*models.Summary, error) {
	request := summarytypes.NewSummaryRequest(from, to, user)
	options := summarytypes.DefaultProcessingOptions()
	return srv.Generate(request, options)
}

// DetailedSummary provides a detailed summary with full processing
func (srv *SummaryService) DetailedSummary(request *summarytypes.SummaryRequest) (*models.Summary, error) {
	options := summarytypes.DefaultProcessingOptions()
	return srv.Generate(request, options)
}

// RetrieveFromStorage gets summaries from the database
func (srv *SummaryService) RetrieveFromStorage(request *summarytypes.SummaryRequest) (*models.Summary, error) {
	summaries := make([]*models.Summary, 0)

	// Filtered summaries are not persisted currently
	// Special case: if (a) filters apply to only one entity type and (b) we're only interested in the summary items of that particular entity type,
	// we can still fetch the persisted summary and drop all irrelevant parts from it
	if request.Filters == nil || request.Filters.IsEmpty() || (request.Filters.CountDistinctTypes() == 1 && request.Filters.SelectFilteredOnly) {
		// Get all already existing, pre-generated summaries that fall into the requested interval
		result, err := srv.repository.GetByUserWithin(request.User, request.From, request.To)
		if err == nil {
			summaries = result
		} else {
			return nil, err
		}
	}

	// Generate missing slots (especially before and after existing summaries) from durations (formerly raw heartbeats)
	missingIntervals := srv.getMissingIntervals(request.From, request.To, summaries, false)
	for _, interval := range missingIntervals {
		if s, err := srv.ComputeFromDurations(summarytypes.NewSummaryRequest(interval.Start, interval.End, request.User).WithFilters(request.Filters)); err == nil {
			if len(missingIntervals) > 2 && s.FromTime.T().Equal(s.ToTime.T()) {
				// little hack here: GetAllWithin will query for >= from_date
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

	if request.Filters != nil && request.Filters.CountDistinctTypes() == 1 && request.Filters.SelectFilteredOnly {
		filter := request.Filters.OneOrEmpty()
		summary.KeepOnly(map[uint8]bool{filter.Entity: true}).ApplyFilter(filter)
	}

	return summary.Sorted(), nil
}

// ComputeFromDurations computes fresh summaries from duration data
func (srv *SummaryService) ComputeFromDurations(request *summarytypes.SummaryRequest) (*models.Summary, error) {
	// Initialize and fetch data
	durations, err := srv.durationService.Get(request.From, request.To, request.User, request.Filters, "project")
	if err != nil {
		return nil, err
	}

	types := models.PersistedSummaryTypes()
	if request.Filters != nil && request.Filters.IsProjectDetails() {
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

	from := request.From
	to := request.To
	if durations.Len() > 0 {
		from = time.Time(durations.First().Time)
		to = time.Time(durations.Last().Time)
	}

	summary := &models.Summary{
		UserID:           request.User.ID,
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

	return summary.Sorted(), nil
}

// Helper methods for the Generate method

func (srv *SummaryService) needsComputation(summary *models.Summary, request *summarytypes.SummaryRequest) bool {
	// If we have no summary, we definitely need computation
	if summary == nil {
		return true
	}

	// If the summary doesn't cover the full requested range, we need computation
	if summary.FromTime.T().After(request.From) || summary.ToTime.T().Before(request.To) {
		return true
	}

	return false
}

func (srv *SummaryService) mergeSummaryWithComputed(existing, computed *models.Summary) *models.Summary {
	if existing == nil {
		return computed
	}
	if computed == nil {
		return existing
	}

	// Use the existing merge logic
	merged, _ := srv.mergeSummaries([]*models.Summary{existing, computed})
	return merged
}

func (srv *SummaryService) applyEnhancements(summary *models.Summary, user *models.User, options *summarytypes.ProcessingOptions) *models.Summary {
	if options.ApplyAliases {
		// Apply alias resolution logic from the existing Aliased method
		resolveAliases := srv.getAliasResolver(user)
		summary = summary.WithResolvedAliases(resolveAliases)
	}

	if options.ApplyProjectLabels {
		// Apply project labels
		summary = srv.withProjectLabels(summary)
	}

	return summary
}

func (srv *SummaryService) cacheResult(summary *models.Summary, request *summarytypes.SummaryRequest) {
	// Use existing cache logic
	cacheKey := srv.getHash(request.From.String(), request.To.String(), request.User.ID, request.Filters.Hash(), "--generated")
	if request.To.Truncate(time.Second).Equal(request.To) && request.From.Truncate(time.Second).Equal(request.From) {
		srv.cache.Set(cacheKey, summary, 0)
	}
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
	return srv.repository.Insert(summary)
}

func (srv *SummaryService) GetHeartbeatsWritePercentage(userID string, start time.Time, end time.Time) (float64, error) {
	return srv.heartbeatService.GetHeartbeatsWritePercentage(userID, start, end)
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

	var processed = map[time.Time]bool{}

	for i, s := range summaries {
		hash := s.FromTime.T()
		if _, found := processed[hash]; found {
			slog.Warn("summary was attempted to be processed more often than once", "fromTime", s.FromTime, "toTime", s.ToTime, "userID", s.UserID)
			continue
		}

		if i > 0 {
			if prev := summaries[i-1]; s.FromTime.T().Before(prev.ToTime.T()) {
				slog.Warn("got overlapping summaries for user", "prevID", prev.ID, "currentID", s.ID, "userID", s.UserID, "fromTime", s.FromTime, "prevToTime", prev.ToTime)
			}
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
		finalSummary.Branches = srv.mergeSummaryItems(finalSummary.Branches, s.Branches)
		finalSummary.Entities = srv.mergeSummaryItems(finalSummary.Entities, s.Entities)
		finalSummary.Categories = srv.mergeSummaryItems(finalSummary.Categories, s.Categories)
		finalSummary.NumHeartbeats += s.NumHeartbeats

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

func (srv *SummaryService) getMissingIntervals(from, to time.Time, summaries []*models.Summary, precise bool) []*models.Interval {
	if len(summaries) == 0 {
		return []*models.Interval{{Start: from, End: to}}
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
