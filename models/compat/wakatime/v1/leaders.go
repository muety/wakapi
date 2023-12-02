package v1

// partially compatible with https://wakatime.com/developers#leaders

type LeadersViewModel struct {
	CurrentUser *LeadersCurrentUser `json:"current_user"`
	Data        []*LeadersEntry     `json:"data"`
	Page        int                 `json:"page"`
	TotalPages  int                 `json:"total_pages"`
	Language    string              `json:"language"`
	Range       *LeadersRange       `json:"range"`
}

type LeadersCurrentUser struct {
	Rank int   `json:"rank"`
	Page int   `json:"page"`
	User *User `json:"user"`
}

type LeadersEntry struct {
	Rank         int                  `json:"rank"`
	RunningTotal *LeadersRunningTotal `json:"running_total"`
	User         *User                `json:"user"`
}

type LeadersRunningTotal struct {
	TotalSeconds              float64            `json:"total_seconds"`
	HumanReadableTotal        string             `json:"human_readable_total"`
	DailyAverage              float64            `json:"daily_average"`
	HumanReadableDailyAverage string             `json:"human_readable_daily_average"`
	Languages                 []*LeadersLanguage `json:"languages"`
}

type LeadersLanguage struct {
	Name         string  `json:"name"`
	TotalSeconds float64 `json:"total_seconds"`
}

type LeadersRange struct {
	EndText   string `json:"end_text"`
	EndDate   string `json:"end_date"`
	StartText string `json:"start_text"`
	StartDate string `json:"start_date"`
	Name      string `json:"name"`
	Text      string `json:"text"`
}
