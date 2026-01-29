package v1

import (
	"math"
	"time"

	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
)

// https://wakatime.com/api/v1/users/current/stats/last_7_days
// https://pastr.de/p/vovw1912mlpf9wwycclrfezx

type StatsViewModel struct {
	Data *StatsData `json:"data"`
}

type StatsData struct {
	Username                  string            `json:"username"`
	UserId                    string            `json:"user_id"`
	Start                     string            `json:"start"`
	End                       string            `json:"end"`
	Status                    string            `json:"status"`
	Timezone                  string            `json:"timezone"`
	TotalSeconds              float64           `json:"total_seconds"`
	DailyAverage              float64           `json:"daily_average"`
	DaysIncludingHolidays     int               `json:"days_including_holidays"`
	Range                     string            `json:"range"`
	HumanReadableRange        string            `json:"human_readable_range"`
	HumanReadableTotal        string            `json:"human_readable_total"`
	HumanReadableDailyAverage string            `json:"human_readable_daily_average"`
	IsCodingActivityVisible   bool              `json:"is_coding_activity_visible"`
	IsOtherUsageVisible       bool              `json:"is_other_usage_visible"`
	Editors                   []*SummariesEntry `json:"editors"`
	Languages                 []*SummariesEntry `json:"languages"`
	Machines                  []*SummariesEntry `json:"machines"`
	Projects                  []*SummariesEntry `json:"projects"`
	OperatingSystems          []*SummariesEntry `json:"operating_systems"`
	Branches                  []*SummariesEntry `json:"branches,omitempty"`
	Categories                []*SummariesEntry `json:"categories"`
}

func NewStatsFrom(summary *models.Summary, filters *models.Filters) *StatsViewModel {
	totalTime := summary.TotalTime()
	numDays := int(summary.ToTime.T().Sub(summary.FromTime.T()).Hours() / 24)

	data := &StatsData{
		Username:              summary.UserID,
		UserId:                summary.UserID,
		Start:                 summary.FromTime.T().Format(time.RFC3339),
		End:                   summary.ToTime.T().Format(time.RFC3339),
		Status:                "ok",
		Timezone:              utils.ResolveIANAZone(summary.User.TZ()),
		TotalSeconds:          totalTime.Seconds(),
		DaysIncludingHolidays: numDays,
		HumanReadableTotal:    helpers.FmtWakatimeDuration(totalTime),
	}

	if numDays > 0 {
		data.DailyAverage = totalTime.Seconds() / float64(numDays)
		data.HumanReadableDailyAverage = helpers.FmtWakatimeDuration(totalTime / time.Duration(numDays))
	}
	if math.IsInf(data.DailyAverage, 0) || math.IsNaN(data.DailyAverage) {
		data.DailyAverage = 0
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

	branches := make([]*SummariesEntry, len(summary.Branches))
	for i, e := range summary.Branches {
		branches[i] = convertEntry(e, summary.TotalTimeBy(models.SummaryBranch))
	}

	categories := make([]*SummariesEntry, len(summary.Categories))
	for i, e := range summary.Categories {
		categories[i] = convertEntry(e, summary.TotalTimeBy(models.SummaryCategory))
	}

	// entities omitted intentionally

	data.Editors = editors
	data.Languages = languages
	data.Machines = machines
	data.Projects = projects
	data.OperatingSystems = oss
	data.Branches = branches
	data.Categories = categories

	if summary.Branches == nil {
		data.Branches = nil
	}

	return &StatsViewModel{
		Data: data,
	}
}
