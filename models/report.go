package models

import "time"

type Report struct {
	From           time.Time
	To             time.Time
	User           *User
	Summary        *Summary
	DailySummaries []*Summary
}

func (r *Report) DailyAverage() time.Duration {
	numberOfDays := len(r.DailySummaries)
	if numberOfDays == 0 {
		return 0
	}
	dailyAverage := r.Summary.TotalTime() / time.Duration(numberOfDays)
	return dailyAverage
}
