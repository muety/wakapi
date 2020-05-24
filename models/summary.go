package models

import (
	"time"
)

const (
	NSummaryTypes   uint8 = 4
	SummaryProject  uint8 = 0
	SummaryLanguage uint8 = 1
	SummaryEditor   uint8 = 2
	SummaryOS       uint8 = 3
)

type Summary struct {
	ID               uint           `json:"-" gorm:"primary_key"`
	UserID           string         `json:"user_id" gorm:"not null; index:idx_time_summary_user"`
	FromTime         time.Time      `json:"from" gorm:"not null; type:timestamp; default:CURRENT_TIMESTAMP; index:idx_time_summary_user"`
	ToTime           time.Time      `json:"to" gorm:"not null; type:timestamp; default:CURRENT_TIMESTAMP; index:idx_time_summary_user"`
	Projects         []*SummaryItem `json:"projects"`
	Languages        []*SummaryItem `json:"languages"`
	Editors          []*SummaryItem `json:"editors"`
	OperatingSystems []*SummaryItem `json:"operating_systems"`
}

type SummaryItem struct {
	ID        uint          `json:"-" gorm:"primary_key"`
	SummaryID uint          `json:"-"`
	Type      uint8         `json:"-"`
	Key       string        `json:"key"`
	Total     time.Duration `json:"total"`
}

type SummaryItemContainer struct {
	Type  uint8
	Items []*SummaryItem
}

type SummaryViewModel struct {
	*Summary
	LanguageColors map[string]string
	Error          string
	Success        string
	ApiKey         string
}
