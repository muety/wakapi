package models

import (
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
	ID               uint           `json:"-" gorm:"primary_key"`
	UserID           string         `json:"user_id" gorm:"not null; index:idx_time_summary_user"`
	FromTime         time.Time      `json:"from" gorm:"not null; type:timestamp; default:CURRENT_TIMESTAMP; index:idx_time_summary_user"`
	ToTime           time.Time      `json:"to" gorm:"not null; type:timestamp; default:CURRENT_TIMESTAMP; index:idx_time_summary_user"`
	Projects         []*SummaryItem `json:"projects"`
	Languages        []*SummaryItem `json:"languages"`
	Editors          []*SummaryItem `json:"editors"`
	OperatingSystems []*SummaryItem `json:"operating_systems"`
	Machines         []*SummaryItem `json:"machines"`
}

type SummaryItem struct {
	ID        uint          `json:"-" gorm:"primary_key"`
	SummaryID uint          `json:"-"`
	Type      uint8         `json:"-"`
	Key       string        `json:"key"`
	Total     time.Duration `json:"total"`
}

type SummaryItemContainer struct {
	Type  uint8
	Items []*SummaryItem
}

type SummaryViewModel struct {
	*Summary
	LanguageColors map[string]string
	Error          string
	Success        string
	ApiKey         string
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
	types := []uint8{SummaryProject, SummaryLanguage, SummaryEditor, SummaryOS, SummaryMachine}
	missingTypes := make([]uint8, 0)
	typeItems := map[uint8]*[]*SummaryItem{
		SummaryProject:  &s.Projects,
		SummaryLanguage: &s.Languages,
		SummaryEditor:   &s.Editors,
		SummaryOS:       &s.OperatingSystems,
		SummaryMachine:  &s.Machines,
	}
	var somePresentType uint8

	for _, t := range types {
		if len(*typeItems[t]) == 0 {
			missingTypes = append(missingTypes, t)
		} else {
			somePresentType = t
		}
	}

	// can't proceed if entire summary is empty
	if len(missingTypes) == len(types) {
		return
	}

	// calculate total duration from any of the present sets of items
	var timeSum time.Duration
	for _, item := range *typeItems[somePresentType] {
		timeSum += item.Total
	}

	// construct dummy item for all missing types
	for _, t := range missingTypes {
		*typeItems[t] = append(*typeItems[t], &SummaryItem{
			Type:  t,
			Key:   UnknownSummaryKey,
			Total: timeSum,
		})
	}
}
