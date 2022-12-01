package v1

import (
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/models"
	"time"
)

// https://wakatime.com/developers#all_time_since_today

type AllTimeViewModel struct {
	Data *AllTimeData `json:"data"`
}

type AllTimeData struct {
	TotalSeconds float32       `json:"total_seconds"` // total number of seconds logged since account created
	Text         string        `json:"text"`          // total time logged since account created as human readable string>
	IsUpToDate   bool          `json:"is_up_to_date"` // true if the stats are up to date; when false, a 202 response code is returned and stats will be refreshed soon>
	Range        *AllTimeRange `json:"range"`
}

type AllTimeRange struct {
	End       string `json:"end"`
	EndDate   string `json:"end_date"`
	Start     string `json:"start"`
	StartDate string `json:"start_date"`
	Timezone  string `json:"timezone"`
}

func NewAllTimeFrom(summary *models.Summary) *AllTimeViewModel {
	total := summary.TotalTime()
	tzName, _ := summary.FromTime.T().Zone()
	return &AllTimeViewModel{
		Data: &AllTimeData{
			TotalSeconds: float32(total.Seconds()),
			Text:         helpers.FmtWakatimeDuration(total),
			IsUpToDate:   true,
			Range: &AllTimeRange{
				End:       summary.ToTime.T().Format(time.RFC3339),
				EndDate:   helpers.FormatDate(summary.ToTime.T()),
				Start:     summary.FromTime.T().Format(time.RFC3339),
				StartDate: helpers.FormatDate(summary.FromTime.T()),
				Timezone:  tzName,
			},
		},
	}
}
