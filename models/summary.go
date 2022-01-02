package models

import (
	"errors"
	"sort"
	"time"
)

const (
	NSummaryTypes   uint8 = 99
	SummaryProject  uint8 = 0
	SummaryLanguage uint8 = 1
	SummaryEditor   uint8 = 2
	SummaryOS       uint8 = 3
	SummaryMachine  uint8 = 4
	SummaryLabel    uint8 = 5
	SummaryBranch   uint8 = 6
)

const UnknownSummaryKey = "unknown"
const DefaultProjectLabel = "default"

type Summary struct {
	ID               uint         `json:"-" gorm:"primary_key"`
	User             *User        `json:"-" gorm:"not null; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	UserID           string       `json:"user_id" gorm:"not null; index:idx_time_summary_user"`
	FromTime         CustomTime   `json:"from" gorm:"not null; type:timestamp; default:CURRENT_TIMESTAMP; index:idx_time_summary_user" swaggertype:"string" format:"date" example:"2006-01-02 15:04:05.000"`
	ToTime           CustomTime   `json:"to" gorm:"not null; type:timestamp; default:CURRENT_TIMESTAMP; index:idx_time_summary_user" swaggertype:"string" format:"date" example:"2006-01-02 15:04:05.000"`
	Projects         SummaryItems `json:"projects" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Languages        SummaryItems `json:"languages" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Editors          SummaryItems `json:"editors" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	OperatingSystems SummaryItems `json:"operating_systems" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Machines         SummaryItems `json:"machines" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Labels           SummaryItems `json:"labels" gorm:"-"`   // labels are not persisted, but calculated at runtime, i.e. when summary is retrieved
	Branches         SummaryItems `json:"branches" gorm:"-"` // branches are not persisted, but calculated at runtime in case a project filter is applied
	NumHeartbeats    int          `json:"-" gorm:"default:0"`
}

type SummaryItems []*SummaryItem

type SummaryItem struct {
	ID        uint64        `json:"-" gorm:"primary_key"`
	Summary   *Summary      `json:"-" gorm:"not null; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	SummaryID uint          `json:"-"`
	Type      uint8         `json:"-" gorm:"index:idx_type"`
	Key       string        `json:"key"`
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
	Recompute bool
}

func SummaryTypes() []uint8 {
	return []uint8{SummaryProject, SummaryLanguage, SummaryEditor, SummaryOS, SummaryMachine, SummaryLabel, SummaryBranch}
}

func NativeSummaryTypes() []uint8 {
	return []uint8{SummaryProject, SummaryLanguage, SummaryEditor, SummaryOS, SummaryMachine, SummaryBranch}
}

func PersistedSummaryTypes() []uint8 {
	return []uint8{SummaryProject, SummaryLanguage, SummaryEditor, SummaryOS, SummaryMachine}
}

func (s *Summary) Sorted() *Summary {
	sort.Sort(sort.Reverse(s.Projects))
	sort.Sort(sort.Reverse(s.Machines))
	sort.Sort(sort.Reverse(s.OperatingSystems))
	sort.Sort(sort.Reverse(s.Languages))
	sort.Sort(sort.Reverse(s.Editors))
	sort.Sort(sort.Reverse(s.Labels))
	sort.Sort(sort.Reverse(s.Branches))
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
	}
}

func (s *Summary) ItemsByType(summaryType uint8) *SummaryItems {
	return s.MappedItems()[summaryType]
}

/* Augments the summary in a way that at least one item is present for every type.
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
	for _, f := range filter.filter {
		total += s.TotalTimeByKey(filter.entity, f)
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
	processAliases := func(origin []*SummaryItem) []*SummaryItem {
		if origin == nil {
			return nil
		}

		target := make([]*SummaryItem, 0)

		findItem := func(key string) *SummaryItem {
			for _, item := range target {
				if item.Key == key {
					return item
				}
			}
			return nil
		}

		for _, item := range origin {
			// Add all "top-level" items, i.e. such without aliases
			if key := resolve(item.Type, item.Key); key == item.Key {
				target = append(target, item)
			}
		}

		for _, item := range origin {
			// Add all remaining projects and merge with their alias
			if key := resolve(item.Type, item.Key); key != item.Key {
				if targetItem := findItem(key); targetItem != nil {
					targetItem.Total += item.Total
				} else {
					target = append(target, &SummaryItem{
						ID:        item.ID,
						SummaryID: item.SummaryID,
						Type:      item.Type,
						Key:       key,
						Total:     item.Total,
					})
				}
			}
		}

		return target
	}

	// Resolve aliases
	s.Projects = processAliases(s.Projects)
	s.Editors = processAliases(s.Editors)
	s.Languages = processAliases(s.Languages)
	s.OperatingSystems = processAliases(s.OperatingSystems)
	s.Machines = processAliases(s.Machines)
	s.Labels = processAliases(s.Labels)
	s.Branches = processAliases(s.Branches)

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

func (s *SummaryItem) TotalFixed() time.Duration {
	// this is a workaround, since currently, the total time of a summary item is mistakenly represented in seconds
	// TODO: fix some day, while migrating persisted summary items
	return s.Total * time.Second
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
