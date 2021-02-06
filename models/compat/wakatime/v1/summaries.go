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

type SummariesViewModel struct {
	Data  []*SummariesData `json:"data"`
	End   time.Time        `json:"end"`
	Start time.Time        `json:"start"`
}

type SummariesData struct {
	Categories       []*SummariesEntry    `json:"categories"`
	Dependencies     []*SummariesEntry    `json:"dependencies"`
	Editors          []*SummariesEntry    `json:"editors"`
	Languages        []*SummariesEntry    `json:"languages"`
	Machines         []*SummariesEntry    `json:"machines"`
	OperatingSystems []*SummariesEntry    `json:"operating_systems"`
	Projects         []*SummariesEntry    `json:"projects"`
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

func NewSummariesFrom(summaries []*models.Summary, filters *models.Filters) *SummariesViewModel {
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

	return &SummariesViewModel{
		Data:  data,
		End:   maxDate,
		Start: minDate,
	}
}

func newDataFrom(s *models.Summary) *SummariesData {
	zone, _ := time.Now().Zone()
	total := s.TotalTime()
	totalHrs, totalMins := int(total.Hours()), int((total - time.Duration(total.Hours())*time.Hour).Minutes())

	data := &SummariesData{
		Categories:       make([]*SummariesEntry, 0),
		Dependencies:     make([]*SummariesEntry, 0),
		Editors:          make([]*SummariesEntry, len(s.Editors)),
		Languages:        make([]*SummariesEntry, len(s.Languages)),
		Machines:         make([]*SummariesEntry, len(s.Machines)),
		OperatingSystems: make([]*SummariesEntry, len(s.OperatingSystems)),
		Projects:         make([]*SummariesEntry, len(s.Projects)),
		GrandTotal: &SummariesGrandTotal{
			Digital:      fmt.Sprintf("%d:%d", totalHrs, totalMins),
			Hours:        totalHrs,
			Minutes:      totalMins,
			Text:         utils.FmtWakatimeDuration(total),
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
	wg.Add(5)

	go func(data *SummariesData) {
		defer wg.Done()
		for i, e := range s.Projects {
			data.Projects[i] = convertEntry(e, s.TotalTimeBy(models.SummaryProject))
		}
	}(data)

	go func(data *SummariesData) {
		defer wg.Done()
		for i, e := range s.Editors {
			data.Editors[i] = convertEntry(e, s.TotalTimeBy(models.SummaryEditor))
		}
	}(data)

	go func(data *SummariesData) {
		defer wg.Done()
		for i, e := range s.Languages {
			data.Languages[i] = convertEntry(e, s.TotalTimeBy(models.SummaryLanguage))

		}
	}(data)

	go func(data *SummariesData) {
		defer wg.Done()
		for i, e := range s.OperatingSystems {
			data.OperatingSystems[i] = convertEntry(e, s.TotalTimeBy(models.SummaryOS))
		}
	}(data)

	go func(data *SummariesData) {
		defer wg.Done()
		for i, e := range s.Machines {
			data.Machines[i] = convertEntry(e, s.TotalTimeBy(models.SummaryMachine))
		}
	}(data)

	wg.Wait()
	return data
}

func convertEntry(e *models.SummaryItem, entityTotal time.Duration) *SummariesEntry {
	// this is a workaround, since currently, the total time of a summary item is mistakenly represented in seconds
	// TODO: fix some day, while migrating persisted summary items
	total := e.Total * time.Second
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
		Text:         utils.FmtWakatimeDuration(total),
		TotalSeconds: total.Seconds(),
	}
}
