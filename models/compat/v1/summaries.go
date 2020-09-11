package v1

import (
	"fmt"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"math"
	"sync"
	"time"
)

// https://wakatime.com/developers#summaries
// https://pastr.de/v/736450

type WakatimeSummaries struct {
	Data  []*wakatimeSummariesData `json:"data"`
	End   time.Time                `json:"end"`
	Start time.Time                `json:"start"`
}

type wakatimeSummariesData struct {
	Categories       []*wakatimeSummariesEntry    `json:"categories"`
	Dependencies     []*wakatimeSummariesEntry    `json:"dependencies"`
	Editors          []*wakatimeSummariesEntry    `json:"editors"`
	Languages        []*wakatimeSummariesEntry    `json:"languages"`
	Machines         []*wakatimeSummariesEntry    `json:"machines"`
	OperatingSystems []*wakatimeSummariesEntry    `json:"operating_systems"`
	Projects         []*wakatimeSummariesEntry    `json:"projects"`
	GrandTotal       *wakatimeSummariesGrandTotal `json:"grand_total"`
	Range            *wakatimeSummariesRange      `json:"range"`
}

type wakatimeSummariesEntry struct {
	Digital      string  `json:"digital"`
	Hours        int     `json:"hours"`
	Minutes      int     `json:"minutes"`
	Name         string  `json:"name"`
	Percent      float64 `json:"percent"`
	Seconds      int     `json:"seconds"`
	Text         string  `json:"text"`
	TotalSeconds float64 `json:"total_seconds"`
}

type wakatimeSummariesGrandTotal struct {
	Digital      string  `json:"digital"`
	Hours        int     `json:"hours"`
	Minutes      int     `json:"minutes"`
	Text         string  `json:"text"`
	TotalSeconds float64 `json:"total_seconds"`
}

type wakatimeSummariesRange struct {
	Date     string    `json:"date"`
	End      time.Time `json:"end"`
	Start    time.Time `json:"start"`
	Text     string    `json:"text"`
	Timezone string    `json:"timezone"`
}

func NewSummariesFrom(summaries []*models.Summary, filters *Filters) *WakatimeSummaries {
	data := make([]*wakatimeSummariesData, len(summaries))
	minDate, maxDate := time.Now().Add(1*time.Second), time.Time{}

	for i, s := range summaries {
		data[i] = newDataFrom(s)

		if s.FromTime.Before(minDate) {
			minDate = s.FromTime
		}
		if s.ToTime.After(maxDate) {
			maxDate = s.ToTime
		}
	}

	return &WakatimeSummaries{
		Data:  data,
		End:   maxDate,
		Start: minDate,
	}
}

func newDataFrom(s *models.Summary) *wakatimeSummariesData {
	zone, _ := time.Now().Zone()
	total := s.TotalTime()
	totalHrs, totalMins := int(total.Hours()), int((total - time.Duration(total.Hours())*time.Hour).Minutes())

	data := &wakatimeSummariesData{
		Categories:       make([]*wakatimeSummariesEntry, 0),
		Dependencies:     make([]*wakatimeSummariesEntry, 0),
		Editors:          make([]*wakatimeSummariesEntry, len(s.Editors)),
		Languages:        make([]*wakatimeSummariesEntry, len(s.Languages)),
		Machines:         make([]*wakatimeSummariesEntry, len(s.Machines)),
		OperatingSystems: make([]*wakatimeSummariesEntry, len(s.OperatingSystems)),
		Projects:         make([]*wakatimeSummariesEntry, len(s.Projects)),
		GrandTotal: &wakatimeSummariesGrandTotal{
			Digital:      fmt.Sprintf("%d:%d", totalHrs, totalMins),
			Hours:        totalHrs,
			Minutes:      totalMins,
			Text:         utils.FmtWakatimeDuration(total),
			TotalSeconds: total.Seconds(),
		},
		Range: &wakatimeSummariesRange{
			Date:     time.Now().Format(time.RFC3339),
			End:      s.ToTime,
			Start:    s.FromTime,
			Text:     "",
			Timezone: zone,
		},
	}

	var wg sync.WaitGroup
	wg.Add(5)

	go func(data *wakatimeSummariesData) {
		defer wg.Done()
		for i, e := range s.Projects {
			data.Projects[i] = convertEntry(e, s.TotalTimeBy(models.SummaryProject))
		}
	}(data)

	go func(data *wakatimeSummariesData) {
		defer wg.Done()
		for i, e := range s.Editors {
			data.Editors[i] = convertEntry(e, s.TotalTimeBy(models.SummaryEditor))
		}
	}(data)

	go func(data *wakatimeSummariesData) {
		defer wg.Done()
		for i, e := range s.Languages {
			data.Languages[i] = convertEntry(e, s.TotalTimeBy(models.SummaryLanguage))

		}
	}(data)

	go func(data *wakatimeSummariesData) {
		defer wg.Done()
		for i, e := range s.OperatingSystems {
			data.OperatingSystems[i] = convertEntry(e, s.TotalTimeBy(models.SummaryOS))
		}
	}(data)

	go func(data *wakatimeSummariesData) {
		defer wg.Done()
		for i, e := range s.Machines {
			data.Machines[i] = convertEntry(e, s.TotalTimeBy(models.SummaryMachine))
		}
	}(data)

	wg.Wait()
	return data
}

func convertEntry(e *models.SummaryItem, entityTotal time.Duration) *wakatimeSummariesEntry {
	// this is a workaround, since currently, the total time of a summary item is mistakenly represented in seconds
	// TODO: fix some day, while migrating persisted summary items
	total := e.Total * time.Second
	hrs := int(total.Hours())
	mins := int((total - time.Duration(hrs)*time.Hour).Minutes())
	secs := int((total - time.Duration(hrs)*time.Hour - time.Duration(mins)*time.Minute).Seconds())

	return &wakatimeSummariesEntry{
		Digital:      fmt.Sprintf("%d:%d:%d", hrs, mins, secs),
		Hours:        hrs,
		Minutes:      mins,
		Name:         e.Key,
		Percent:      math.Round((total.Seconds()/entityTotal.Seconds())*1e4) / 100,
		Seconds:      secs,
		Text:         utils.FmtWakatimeDuration(total),
		TotalSeconds: total.Seconds(),
	}
}
