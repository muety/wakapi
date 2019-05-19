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
	UserID           string        `json:"user_id"`
	FromTime         *time.Time    `json:"from"`
	ToTime           *time.Time    `json:"to"`
	Projects         []SummaryItem `json:"projects"`
	Languages        []SummaryItem `json:"languages"`
	Editors          []SummaryItem `json:"editors"`
	OperatingSystems []SummaryItem `json:"operating_systems"`
}

type SummaryItem struct {
	Key   string        `json:"key"`
	Total time.Duration `json:"total"`
}
