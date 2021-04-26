package v1

import (
	"github.com/muety/wakapi/models"
	"math"
	"time"
)

// https://wakatime.com/api/v1/users/current/stats/last_7_days
// https://pastr.de/p/f2fxg6ragj7z5e7fhsow9rb6

type StatsViewModel struct {
	Data *StatsData `json:"data"`
}

type StatsData struct {
	Username              string            `json:"username"`
	UserId                string            `json:"user_id"`
	Start                 time.Time         `json:"start"`
	End                   time.Time         `json:"end"`
	TotalSeconds          float64           `json:"total_seconds"`
	DailyAverage          float64           `json:"daily_average"`
	DaysIncludingHolidays int               `json:"days_including_holidays"`
	Editors               []*SummariesEntry `json:"editors"`
	Languages             []*SummariesEntry `json:"languages"`
	Machines              []*SummariesEntry `json:"machines"`
	Projects              []*SummariesEntry `json:"projects"`
	OperatingSystems      []*SummariesEntry `json:"operating_systems"`
}

func NewStatsFrom(summary *models.Summary, filters *models.Filters) *StatsViewModel {
	totalTime := summary.TotalTime()
	numDays := int(summary.ToTime.T().Sub(summary.FromTime.T()).Hours() / 24)
	if math.IsInf(float64(numDays), 0) {
		numDays = 0
	}

	data := &StatsData{
		Username:              summary.UserID,
		UserId:                summary.UserID,
		Start:                 summary.FromTime.T(),
		End:                   summary.ToTime.T(),
		TotalSeconds:          totalTime.Seconds(),
		DailyAverage:          totalTime.Seconds() / float64(numDays),
		DaysIncludingHolidays: numDays,
	}

	editors := make([]*SummariesEntry, len(summary.Editors))
	for i, e := range summary.Editors {
		editors[i] = convertEntry(e, summary.TotalTimeBy(models.SummaryEditor))
	}

	languages := make([]*SummariesEntry, len(summary.Languages))
	for i, e := range summary.Languages {
		languages[i] = convertEntry(e, summary.TotalTimeBy(models.SummaryLanguage))
	}

	machines := make([]*SummariesEntry, len(summary.Machines))
	for i, e := range summary.Machines {
		machines[i] = convertEntry(e, summary.TotalTimeBy(models.SummaryMachine))
	}

	projects := make([]*SummariesEntry, len(summary.Projects))
	for i, e := range summary.Projects {
		projects[i] = convertEntry(e, summary.TotalTimeBy(models.SummaryProject))
	}

	oss := make([]*SummariesEntry, len(summary.OperatingSystems))
	for i, e := range summary.OperatingSystems {
		oss[i] = convertEntry(e, summary.TotalTimeBy(models.SummaryOS))
	}

	data.Editors = editors
	data.Languages = languages
	data.Machines = machines
	data.Projects = projects
	data.OperatingSystems = oss

	return &StatsViewModel{
		Data: data,
	}
}
