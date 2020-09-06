package v1

// https://wakatime.com/developers#all_time_since_today

type AllTimeVieModel struct {
	Data *AllTimeVieModelData `json:"data"`
}

type AllTimeVieModelData struct {
	Seconds    float32 `json:"seconds"`       // total number of seconds logged since account created
	Text       string  `json:"text"`          // total time logged since account created as human readable string>
	IsUpToDate bool    `json:"is_up_to_date"` // true if the stats are up to date; when false, a 202 response code is returned and stats will be refreshed soon>
}
