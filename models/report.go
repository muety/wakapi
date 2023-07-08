package models

import "time"

type Report struct {
	From           time.Time
	To             time.Time
	User           *User
	Summary        *Summary
	DailySummaries []*Summary
}
