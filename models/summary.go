package models

import (
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
)

const UnknownSummaryKey = "unknown"

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
}

type SummaryItems []*SummaryItem

type SummaryItem struct {
	ID        uint          `json:"-" gorm:"primary_key"`
	Summary   *Summary      `json:"-" gorm:"not null; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	SummaryID uint          `json:"-"`
	Type      uint8         `json:"-"`
	Key       string        `json:"key"`
	Total     time.Duration `json:"total" swaggertype:"primitive,integer"`
}

type SummaryItemContainer struct {
	Type  uint8
	Items []*SummaryItem
}

type SummaryViewModel struct {
	*Summary
	LanguageColors map[string]string
	EditorColors   map[string]string
	OSColors       map[string]string
	Error          string
	Success        string
	ApiKey         string
}

type SummaryParams struct {
	From      time.Time
	To        time.Time
	User      *User
	Recompute bool
}

type AliasResolver func(t uint8, k string) string

func SummaryTypes() []uint8 {
	return []uint8{SummaryProject, SummaryLanguage, SummaryEditor, SummaryOS, SummaryMachine}
}

func (s *Summary) Sorted() *Summary {
	sort.Sort(sort.Reverse(s.Projects))
	sort.Sort(sort.Reverse(s.Machines))
	sort.Sort(sort.Reverse(s.OperatingSystems))
	sort.Sort(sort.Reverse(s.Languages))
	sort.Sort(sort.Reverse(s.Editors))
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
	}
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
func (s *Summary) FillUnknown() {
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

	timeSum := s.TotalTime()

	// construct dummy item for all missing types
	for _, t := range missingTypes {
		*typeItems[t] = append(*typeItems[t], &SummaryItem{
			Type:  t,
			Key:   UnknownSummaryKey,
			Total: timeSum,
		})
	}
}

func (s *Summary) TotalTime() time.Duration {
	var timeSum time.Duration

	mappedItems := s.MappedItems()
	// calculate total duration from any of the present sets of items
	for _, t := range s.Types() {
		if items := mappedItems[t]; len(*items) > 0 {
			for _, item := range *items {
				timeSum += item.Total
			}
			break
		}
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

func (s *Summary) TotalTimeByFilters(filters *Filters) time.Duration {
	do, typeId, key := filters.One()
	if do {
		return s.TotalTimeByKey(typeId, key)
	}
	return 0
}

func (s *Summary) WithResolvedAliases(resolve AliasResolver) *Summary {
	processAliases := func(origin []*SummaryItem) []*SummaryItem {
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

	return s
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
