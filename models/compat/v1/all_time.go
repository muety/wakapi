package v1

import (
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"time"
)

// https://wakatime.com/developers#all_time_since_today

type WakatimeAllTime struct {
	Data *wakatimeAllTimeData `json:"data"`
}

type wakatimeAllTimeData struct {
	TotalSeconds float32 `json:"total_seconds"` // total number of seconds logged since account created
	Text         string  `json:"text"`          // total time logged since account created as human readable string>
	IsUpToDate   bool    `json:"is_up_to_date"` // true if the stats are up to date; when false, a 202 response code is returned and stats will be refreshed soon>
}

func NewAllTimeFrom(summary *models.Summary, filters *Filters) *WakatimeAllTime {
	var total time.Duration
	if key := filters.Project; key != "" {
		total = summary.TotalTimeByKey(models.SummaryProject, key)
	} else {
		total = summary.TotalTime()
	}

	return &WakatimeAllTime{
		Data: &wakatimeAllTimeData{
			TotalSeconds: float32(total.Seconds()),
			Text:         utils.FmtWakatimeDuration(total),
			IsUpToDate:   true,
		},
	}
}
