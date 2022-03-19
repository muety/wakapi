package models

type Diagnostics struct {
	ID           uint   `gorm:"primary_key"`
	Platform     string `json:"platform"`
	Architecture string `json:"architecture"`
	Plugin       string `json:"plugin"`
	CliVersion   string `json:"cli_version"`
	Logs         string `json:"logs" gorm:"type:text"`
	StackTrace   string `json:"stacktrace" gorm:"type:text"`
}
