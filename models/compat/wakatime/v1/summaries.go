package v1

import (
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/duke-git/lancet/v2/mathutil"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
)

// https://wakatime.com/developers#summaries
// https://pastr.de/v/736450

type SummariesViewModel struct {
	Data            []*SummariesData          `json:"data"`
	End             time.Time                 `json:"end"`
	Start           time.Time                 `json:"start"`
	CumulativeTotal *SummariesCumulativeTotal `json:"cumulative_total"`
	DailyAverage    *SummariesDailyAverage    `json:"daily_average"`
}

type SummariesCumulativeTotal struct {
	Decimal string  `json:"decimal"`
	Digital string  `json:"digital"`
	Seconds float64 `json:"seconds"`
	Text    string  `json:"text"`
}

type SummariesDailyAverage struct {
	DaysIncludingHolidays         int    `json:"days_including_holidays"`
	DaysMinusHolidays             int    `json:"days_minus_holidays"`
	Holidays                      int    `json:"holidays"`
	Seconds                       int64  `json:"seconds"`
	SecondsIncludingOtherLanguage int64  `json:"seconds_including_other_language"`
	Text                          string `json:"text"`
	TextIncludingOtherLanguage    string `json:"text_including_other_language"`
}

type SummariesData struct {
	Categories       []*SummariesEntry    `json:"categories"`
	Dependencies     []*SummariesEntry    `json:"dependencies"`
	Editors          []*SummariesEntry    `json:"editors"`
	Languages        []*SummariesEntry    `json:"languages"`
	Machines         []*SummariesEntry    `json:"machines"`
	OperatingSystems []*SummariesEntry    `json:"operating_systems"`
	Projects         []*SummariesEntry    `json:"projects"`
	Branches         []*SummariesEntry    `json:"branches"`
	Entities         []*SummariesEntry    `json:"entities"`
	GrandTotal       *SummariesGrandTotal `json:"grand_total"`
	Range            *SummariesRange      `json:"range"`
}

type SummariesEntry struct {
	Digital      string  `json:"digital"`
	Hours        int     `json:"hours"`
	Minutes      int     `json:"minutes"`
	Name         string  `json:"name"`
	Percent      float64 `json:"percent"`
	Seconds      int     `json:"seconds"`
	Text         string  `json:"text"`
	TotalSeconds float64 `json:"total_seconds"`
}

type SummariesGrandTotal struct {
	Digital      string  `json:"digital"`
	Hours        int     `json:"hours"`
	Minutes      int     `json:"minutes"`
	Text         string  `json:"text"`
	TotalSeconds float64 `json:"total_seconds"`
}

type SummariesRange struct {
	Date     string    `json:"date"`
	End      time.Time `json:"end"`
	Start    time.Time `json:"start"`
	Text     string    `json:"text"`
	Timezone string    `json:"timezone"`
}

// MarshalJSON adds a customized JSON serialization that will include the `branches` and `entities` fields if set to empty arrays, but exclude them if otherwise considered empty.
func (s *SummariesData) MarshalJSON() ([]byte, error) {
	type alias SummariesData
	if s.Branches == nil || s.Entities == nil {
		return json.Marshal(&struct {
			*alias
			Branches []*SummariesEntry `json:"branches,omitempty"`
			Entities []*SummariesEntry `json:"entities,omitempty"`
		}{alias: (*alias)(s)})
	}
	return json.Marshal((*alias)(s))
}

func NewSummariesFrom(summaries []*models.Summary) *SummariesViewModel {
	data := make([]*SummariesData, len(summaries))
	minDate, maxDate := time.Now().Add(1*time.Second), time.Time{}

	for i, s := range summaries {
		data[i] = newDataFrom(s)

		if s.FromTime.T().Before(minDate) {
			minDate = s.FromTime.T()
		}
		if s.ToTime.T().After(maxDate) {
			maxDate = s.ToTime.T()
		}
	}

	var totalTime time.Duration
	var totalTimeKnown time.Duration // total duration in non-unknown languages
	for _, s := range summaries {
		total := s.TotalTime()
		totalTime += total
		totalTimeKnown += (total - s.TotalTimeByKey(models.SummaryLanguage, models.UnknownSummaryKey))
	}

	totalHrs, totalMins, totalSecs := totalTime.Hours(), (totalTime - time.Duration(totalTime.Hours())*time.Hour).Minutes(), totalTime.Seconds()
	totalDays := mathutil.Max[int](len(utils.SplitRangeByDays(minDate, maxDate)), 1)
	totalTimeAvg, totalTimeKnownAvg := totalTime/time.Duration(totalDays), totalTimeKnown/time.Duration(totalDays)
	totalSecsAvg, totalSecsKnownAvg := int64(totalTimeAvg.Seconds()), int64(totalTimeKnownAvg.Seconds())

	return &SummariesViewModel{
		Data:  data,
		End:   maxDate,
		Start: minDate,
		CumulativeTotal: &SummariesCumulativeTotal{
			Decimal: fmt.Sprintf("%.2f", totalHrs),
			Digital: fmt.Sprintf("%d:%d", int(totalHrs), int(totalMins)),
			Seconds: totalSecs,
			Text:    helpers.FmtWakatimeDuration(totalTime),
		},
		DailyAverage: &SummariesDailyAverage{
			DaysIncludingHolidays:         totalDays,
			DaysMinusHolidays:             totalDays,
			Holidays:                      0, // not implemented, because we don't track user location
			Seconds:                       totalSecsKnownAvg,
			SecondsIncludingOtherLanguage: totalSecsAvg,
			Text:                          helpers.FmtWakatimeDuration(totalTimeKnownAvg),
			TextIncludingOtherLanguage:    helpers.FmtWakatimeDuration(totalTimeAvg),
		},
	}
}

func newDataFrom(s *models.Summary) *SummariesData {
	zone, _ := time.Now().Zone()
	total := s.TotalTime()
	totalHrs, totalMins := int(total.Hours()), int((total - time.Duration(total.Hours())*time.Hour).Minutes())

	data := &SummariesData{
		Dependencies:     make([]*SummariesEntry, 0),
		Editors:          make([]*SummariesEntry, len(s.Editors)),
		Languages:        make([]*SummariesEntry, len(s.Languages)),
		Machines:         make([]*SummariesEntry, len(s.Machines)),
		OperatingSystems: make([]*SummariesEntry, len(s.OperatingSystems)),
		Projects:         make([]*SummariesEntry, len(s.Projects)),
		Branches:         make([]*SummariesEntry, len(s.Branches)),
		Entities:         make([]*SummariesEntry, len(s.Entities)),
		Categories:       make([]*SummariesEntry, len(s.Categories)),
		GrandTotal: &SummariesGrandTotal{
			Digital:      fmt.Sprintf("%d:%d", totalHrs, totalMins),
			Hours:        totalHrs,
			Minutes:      totalMins,
			Text:         helpers.FmtWakatimeDuration(total),
			TotalSeconds: total.Seconds(),
		},
		Range: &SummariesRange{
			Date:     time.Now().Format(time.RFC3339),
			End:      s.ToTime.T(),
			Start:    s.FromTime.T(),
			Text:     "",
			Timezone: zone,
		},
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go utils.WithRecovery1[*SummariesData](func(data *SummariesData) {
		defer wg.Done()
		for i, e := range s.Projects {
			data.Projects[i] = convertEntry(e, s.TotalTimeBy(models.SummaryProject))
		}
	}, data)

	wg.Add(1)
	go utils.WithRecovery1[*SummariesData](func(data *SummariesData) {
		defer wg.Done()
		for i, e := range s.Editors {
			data.Editors[i] = convertEntry(e, s.TotalTimeBy(models.SummaryEditor))
		}
	}, data)

	wg.Add(1)
	go utils.WithRecovery1[*SummariesData](func(data *SummariesData) {
		defer wg.Done()
		for i, e := range s.Languages {
			data.Languages[i] = convertEntry(e, s.TotalTimeBy(models.SummaryLanguage))
		}
	}, data)

	wg.Add(1)
	go utils.WithRecovery1[*SummariesData](func(data *SummariesData) {
		defer wg.Done()
		for i, e := range s.OperatingSystems {
			data.OperatingSystems[i] = convertEntry(e, s.TotalTimeBy(models.SummaryOS))
		}
	}, data)

	wg.Add(1)
	go utils.WithRecovery1[*SummariesData](func(data *SummariesData) {
		defer wg.Done()
		for i, e := range s.Machines {
			data.Machines[i] = convertEntry(e, s.TotalTimeBy(models.SummaryMachine))
		}
	}, data)

	wg.Add(1)
	go utils.WithRecovery1[*SummariesData](func(data *SummariesData) {
		defer wg.Done()
		for i, e := range s.Branches {
			data.Branches[i] = convertEntry(e, s.TotalTimeBy(models.SummaryBranch))
		}
	}, data)

	wg.Add(1)
	go utils.WithRecovery1[*SummariesData](func(data *SummariesData) {
		defer wg.Done()
		for i, e := range s.Entities {
			data.Entities[i] = convertEntry(e, s.TotalTimeBy(models.SummaryEntity))
		}
	}, data)

	wg.Add(1)
	go utils.WithRecovery1[*SummariesData](func(data *SummariesData) {
		defer wg.Done()
		for i, e := range s.Categories {
			data.Categories[i] = convertEntry(e, s.TotalTimeBy(models.SummaryCategory))
		}
	}, data)

	if s.Branches == nil {
		data.Branches = nil
	}
	if s.Entities == nil {
		data.Entities = nil
	}

	wg.Wait()
	return data
}

func convertEntry(e *models.SummaryItem, entityTotal time.Duration) *SummariesEntry {
	total := e.TotalFixed()
	hrs := int(total.Hours())
	mins := int((total - time.Duration(hrs)*time.Hour).Minutes())
	secs := int((total - time.Duration(hrs)*time.Hour - time.Duration(mins)*time.Minute).Seconds())
	percentage := math.Round((total.Seconds()/entityTotal.Seconds())*1e4) / 100
	if math.IsNaN(percentage) || math.IsInf(percentage, 0) {
		percentage = 0
	}

	return &SummariesEntry{
		Digital:      fmt.Sprintf("%d:%d:%d", hrs, mins, secs),
		Hours:        hrs,
		Minutes:      mins,
		Name:         e.Key,
		Percent:      percentage,
		Seconds:      secs,
		Text:         helpers.FmtWakatimeDuration(total),
		TotalSeconds: total.Seconds(),
	}
}
