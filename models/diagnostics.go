package models

type Diagnostics struct {
	ID           uint   `gorm:"primary_key"`
	User         *User  `json:"-" gorm:"not null; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	UserID       string `json:"-" gorm:"not null; index:idx_diagnostics_user"`
	Platform     string `json:"platform"`
	Architecture string `json:"architecture"`
	Plugin       string `json:"plugin"`
	CliVersion   string `json:"cli_version"`
	Logs         string `json:"logs" gorm:"type:text"`
	StackTrace   string `json:"stacktrace" gorm:"type:text"`
}
