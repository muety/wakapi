package models

import (
	"errors"
	"math"
	"sort"
	"time"

	"github.com/duke-git/lancet/v2/mathutil"
	"github.com/duke-git/lancet/v2/slice"
)

const (
	NSummaryTypes   uint8 = 99
	SummaryUnknown  uint8 = 98
	SummaryProject  uint8 = 0
	SummaryLanguage uint8 = 1
	SummaryEditor   uint8 = 2
	SummaryOS       uint8 = 3
	SummaryMachine  uint8 = 4
	SummaryLabel    uint8 = 5
	SummaryBranch   uint8 = 6
	SummaryEntity   uint8 = 7
	SummaryCategory uint8 = 8
)

const UnknownSummaryKey = "unknown"
const DefaultProjectLabel = "default"

type Summaries []*Summary

type Summary struct {
	ID       uint       `json:"-" gorm:"primary_key; size:32"`
	User     *User      `json:"-" gorm:"not null; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	UserID   string     `json:"user_id" gorm:"not null; index:idx_time_summary_user"`
	FromTime CustomTime `json:"from" gorm:"not null; default:CURRENT_TIMESTAMP; index:idx_time_summary_user" swaggertype:"string" format:"date" example:"2006-01-02 15:04:05.000"`
	ToTime   CustomTime `json:"to" gorm:"not null; default:CURRENT_TIMESTAMP; index:idx_time_summary_user" swaggertype:"string" format:"date" example:"2006-01-02 15:04:05.000"`

	// Previously, all the following properties created a cascade foreign key constraint on the summary_items table
	// back to this summary table resulting in 5 identical foreign key constraints on the summary_items table.
	// This is not a problem for PostgreSQL, MySQL and SQLite, but for MSSQL, which complains about circular cascades on
	// update/delete between these two tables. All of these created foreign key constraints are identical, so only one constraint is enough.
	// MySQL will create a foreign key constraint for every property referencing other structs, even no constraint is specified in tags.
	// So explicitly set gorm:"-" in all other properties to avoid creating duplicate foreign key constraints
	Projects         SummaryItems `json:"projects" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Languages        SummaryItems `json:"languages" gorm:"-"`
	Editors          SummaryItems `json:"editors" gorm:"-"`
	OperatingSystems SummaryItems `json:"operating_systems" gorm:"-"`
	Machines         SummaryItems `json:"machines" gorm:"-"`
	Labels           SummaryItems `json:"labels" gorm:"-"`   // labels are not persisted, but calculated at runtime, i.e. when summary is retrieved
	Branches         SummaryItems `json:"branches" gorm:"-"` // branches are not persisted, but calculated at runtime in case a project Filter is applied
	Entities         SummaryItems `json:"entities" gorm:"-"` // entities are not persisted, but calculated at runtime in case a project Filter is applied
	Categories       SummaryItems `json:"categories" gorm:"-"`
	NumHeartbeats    int          `json:"-"`
}

type SummaryItems []*SummaryItem

type SummaryItem struct {
	ID        uint64        `json:"-" gorm:"primary_key"`
	Summary   *Summary      `json:"-" gorm:"not null; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	SummaryID uint          `json:"-" gorm:"size:32"`
	Type      uint8         `json:"-" gorm:"index:idx_type"`
	Key       string        `json:"key" gorm:"size:255"`
	Total     time.Duration `json:"total" swaggertype:"primitive,integer"`
}

type SummaryItemContainer struct {
	Type  uint8
	Items []*SummaryItem
}

type SummaryParams struct {
	From      time.Time
	To        time.Time
	User      *User
	Filters   *Filters
	Recompute bool
}

func SummaryTypes() []uint8 {
	return []uint8{SummaryProject, SummaryLanguage, SummaryEditor, SummaryOS, SummaryMachine, SummaryLabel, SummaryBranch, SummaryEntity, SummaryCategory}
}

func NativeSummaryTypes() []uint8 {
	return []uint8{SummaryProject, SummaryLanguage, SummaryEditor, SummaryOS, SummaryMachine, SummaryBranch, SummaryEntity, SummaryCategory}
}

func PersistedSummaryTypes() []uint8 {
	return []uint8{SummaryProject, SummaryLanguage, SummaryEditor, SummaryOS, SummaryMachine, SummaryCategory}
}

func NewEmptySummary() *Summary {
	return &Summary{
		Projects:         SummaryItems{},
		Languages:        SummaryItems{},
		Editors:          SummaryItems{},
		OperatingSystems: SummaryItems{},
		Machines:         SummaryItems{},
		Labels:           SummaryItems{},
		Branches:         SummaryItems{},
		Entities:         SummaryItems{},
		Categories:       SummaryItems{},
	}
}

func (s *Summary) Sorted() *Summary {
	sort.Sort(sort.Reverse(s.Projects))
	sort.Sort(sort.Reverse(s.Machines))
	sort.Sort(sort.Reverse(s.OperatingSystems))
	sort.Sort(sort.Reverse(s.Languages))
	sort.Sort(sort.Reverse(s.Editors))
	sort.Sort(sort.Reverse(s.Labels))
	sort.Sort(sort.Reverse(s.Branches))
	sort.Sort(sort.Reverse(s.Entities))
	sort.Sort(sort.Reverse(s.Categories))
	return s
}

func (s *Summary) Types() []uint8 {
	return SummaryTypes()
}

func (s *Summary) MappedItems() map[uint8]*SummaryItems {
	return map[uint8]*SummaryItems{
		SummaryProject:  &s.Projects,
		SummaryLanguage: &s.Languages,
		SummaryEditor:   &s.Editors,
		SummaryOS:       &s.OperatingSystems,
		SummaryMachine:  &s.Machines,
		SummaryLabel:    &s.Labels,
		SummaryBranch:   &s.Branches,
		SummaryEntity:   &s.Entities,
		SummaryCategory: &s.Categories,
	}
}

func (s *Summary) GetByType(summaryType uint8) *SummaryItems {
	switch summaryType {
	case SummaryProject:
		return &s.Projects
	case SummaryLanguage:
		return &s.Languages
	case SummaryEditor:
		return &s.Editors
	case SummaryOS:
		return &s.OperatingSystems
	case SummaryMachine:
		return &s.Machines
	case SummaryLabel:
		return &s.Labels
	case SummaryBranch:
		return &s.Branches
	case SummaryEntity:
		return &s.Entities
	case SummaryCategory:
		return &s.Categories
	}
	return nil
}

func (s *Summary) SetByType(summaryType uint8, items *SummaryItems) {
	switch summaryType {
	case SummaryProject:
		s.Projects = *items
		break
	case SummaryLanguage:
		s.Languages = *items
		break
	case SummaryEditor:
		s.Editors = *items
		break
	case SummaryOS:
		s.OperatingSystems = *items
		break
	case SummaryMachine:
		s.Machines = *items
		break
	case SummaryLabel:
		s.Labels = *items
		break
	case SummaryBranch:
		s.Branches = *items
		break
	case SummaryEntity:
		s.Entities = *items
		break
	case SummaryCategory:
		s.Categories = *items
		break
	}
}

func (s *Summary) KeepOnly(types map[uint8]bool) *Summary {
	if len(types) == 0 {
		return s
	}

	for _, t := range SummaryTypes() {
		if keep, ok := types[t]; !keep || !ok {
			*s.GetByType(t) = []*SummaryItem{}
		}
	}

	return s
}

// ApplyFilter drops all summary elements of the given type that don't match the given query.
// Please note: this only makes sense if you're eventually interested in nothing but the total time of that specific type,
// because the summary will be inconsistent after this operation (e.g. when filtering by project, languages, editors, etc. won't match up anymore).
// Therefore, use with caution.
func (s *Summary) ApplyFilter(filter FilterElement) *Summary {
	items := SummaryItems(slice.Filter[*SummaryItem](*s.GetByType(filter.Entity), func(i int, item *SummaryItem) bool {
		return filter.Filter.MatchAny(item.Key)
	}))
	s.SetByType(filter.Entity, &items)
	return s
}

/*
Augments the summary in a way that at least one item is present for every type.

If a summary has zero items for a given type, but one or more for any of the other types,
the total summary duration can be derived from those and inserted as a dummy-item with key "unknown"
for the missing type.
For instance, the machine type was introduced post hoc. Accordingly, no "machine"-information is present in
the data for old heartbeats and summaries. If a user has two years of data without machine information and
one day with such, a "machine"-chart plotted from that data will reference a way smaller absolute total amount
of time than the other ones.
To avoid having to modify persisted data retrospectively, i.e. inserting a dummy SummaryItem for the new type,
such is generated dynamically here, considering the "machine" for all old heartbeats "unknown".
*/
func (s *Summary) FillMissing() {
	types := s.Types()
	typeItems := s.MappedItems()
	missingTypes := make([]uint8, 0)

	for _, t := range types {
		if len(*typeItems[t]) == 0 {
			missingTypes = append(missingTypes, t)
		}
	}

	// can't proceed if entire summary is empty
	if len(missingTypes) == len(types) {
		return
	}

	// construct dummy item for all missing types
	presentType, err := s.findFirstPresentType()
	if err != nil {
		return // all types are either zero or missing entirely, nothing to fill
	}
	for _, t := range missingTypes {
		s.FillBy(presentType, t)
	}
}

// inplace!
func (s *Summary) FillBy(fromType uint8, toType uint8) {
	typeItems := s.MappedItems()
	totalWanted := s.TotalTimeBy(fromType)
	totalActual := s.TotalTimeBy(toType)

	key := UnknownSummaryKey
	if toType == SummaryLabel {
		key = DefaultProjectLabel
	}

	existingEntryIdx := -1
	for i, item := range *typeItems[toType] {
		if item.Key == key {
			existingEntryIdx = i
			break
		}
	}

	total := (totalWanted - totalActual) / time.Second // workaround
	if total > 0 {
		if existingEntryIdx >= 0 {
			(*typeItems[toType])[existingEntryIdx].Total = total
		} else {
			*typeItems[toType] = append(*typeItems[toType], &SummaryItem{
				Type:  toType,
				Key:   key,
				Total: total,
			})
		}
	}
}

func (s *Summary) TotalTime() time.Duration {
	var timeSum time.Duration

	mappedItems := s.MappedItems()
	t, err := s.findFirstPresentType()
	if err != nil {
		return 0
	}
	for _, item := range *mappedItems[t] {
		timeSum += item.Total
	}

	return timeSum * time.Second
}

func (s *Summary) TotalTimeBy(entityType uint8) (timeSum time.Duration) {
	mappedItems := s.MappedItems()
	if items := mappedItems[entityType]; len(*items) > 0 {
		for _, item := range *items {
			timeSum = timeSum + item.Total*time.Second
		}
	}
	return timeSum
}

func (s *Summary) TotalTimeByKey(entityType uint8, key string) (timeSum time.Duration) {
	mappedItems := s.MappedItems()
	if items := mappedItems[entityType]; len(*items) > 0 {
		for _, item := range *items {
			if item.Key != key {
				continue
			}
			timeSum = timeSum + item.Total*time.Second
		}
	}
	return timeSum
}

func (s *Summary) TotalTimeByFilter(filter FilterElement) time.Duration {
	var total time.Duration
	for _, f := range filter.Filter {
		total += s.TotalTimeByKey(filter.Entity, f)
	}
	return total
}

func (s *Summary) MaxBy(entityType uint8) *SummaryItem {
	var max *SummaryItem
	mappedItems := s.MappedItems()
	if items := mappedItems[entityType]; len(*items) > 0 {
		for _, item := range *items {
			if max == nil || item.Total > max.Total {
				max = item
			}
		}
	}
	return max
}

func (s *Summary) MaxByToString(entityType uint8) string {
	max := s.MaxBy(entityType)
	if max == nil {
		return "-"
	}
	return max.Key
}

func (s *Summary) WithResolvedAliases(resolve AliasResolver) *Summary {
	processAliases := func(items []*SummaryItem) []*SummaryItem {
		if items == nil {
			return nil
		}
		itemsAliased := make([]*SummaryItem, 0)

		findItem := func(key string) *SummaryItem {
			for _, item := range itemsAliased {
				if item.Key == key {
					return item
				}
			}
			return nil
		}

		for _, item := range items {
			// Add all "top-level" items, i.e. such without aliases
			if key := resolve(item.Type, item.Key); key == item.Key {
				itemsAliased = append(itemsAliased, item)
			}
		}

		for _, item := range items {
			// Add all remaining projects and merge with their alias
			if key := resolve(item.Type, item.Key); key != item.Key {
				if targetItem := findItem(key); targetItem != nil {
					targetItem.Total += item.Total
				} else {
					itemsAliased = append(itemsAliased, &SummaryItem{
						ID:        item.ID,
						SummaryID: item.SummaryID,
						Type:      item.Type,
						Key:       key,
						Total:     item.Total,
					})
				}
			}
		}

		return itemsAliased
	}

	// Resolve aliases
	s.Projects = processAliases(s.Projects)
	s.Editors = processAliases(s.Editors)
	s.Languages = processAliases(s.Languages)
	s.OperatingSystems = processAliases(s.OperatingSystems)
	s.Machines = processAliases(s.Machines)
	s.Labels = processAliases(s.Labels)
	s.Branches = processAliases(s.Branches)
	s.Categories = processAliases(s.Categories)
	// no aliases for entities / files

	return s
}

// inplace!
func (s *Summary) InTZ(tz *time.Location) *Summary {
	// time zone madness, see https://github.com/muety/wakapi/issues/719#issuecomment-2599365514
	s.FromTime = CustomTime(s.FromTime.T().In(tz))
	s.ToTime = CustomTime(s.ToTime.T().In(tz))
	return s
}

func (s *Summary) findFirstPresentType() (uint8, error) {
	for _, t := range s.Types() {
		if s.TotalTimeBy(t) != 0 {
			return t, nil
		}
	}
	return 127, errors.New("no type present")
}

func (s *SummaryParams) HasFilters() bool {
	return s.Filters != nil && !s.Filters.IsEmpty()
}

func (s *SummaryParams) IsProjectDetails() bool {
	if !s.HasFilters() {
		return false
	}
	_, entity, filters := s.Filters.One()
	return entity == SummaryProject && len(filters) == 1 // exactly one
}

func (s *SummaryParams) GetProjectFilter() string {
	if !s.IsProjectDetails() {
		return ""
	}
	_, _, filters := s.Filters.One()
	return filters[0]
}

func (s *SummaryParams) RangeDays() int {
	return int(math.Floor(s.To.Sub(s.From).Hours() / 24))
}

func (s *SummaryItem) TotalFixed() time.Duration {
	// this is a workaround, since currently, the total time of a summary item is mistakenly represented in seconds
	// TODO: fix some day, while migrating persisted summary items
	return s.Total * time.Second
}

func (s Summaries) MaxTotalTime() time.Duration {
	return mathutil.Max(slice.Map[*Summary, time.Duration](s, func(i int, item *Summary) time.Duration {
		return item.TotalTime()
	})...)
}

func (s Summaries) Len() int {
	return len(s)
}

func (s Summaries) Less(i, j int) bool {
	return s[i].FromTime.T().Before(s[j].FromTime.T())
}

func (s Summaries) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SummaryItems) Len() int {
	return len(s)
}

func (s SummaryItems) Less(i, j int) bool {
	return s[i].Total < s[j].Total
}

func (s SummaryItems) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
