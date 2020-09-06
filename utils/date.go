package utils

import (
	"fmt"
	"time"
)

func StartOfDay() time.Time {
	ref := time.Now()
	return time.Date(ref.Year(), ref.Month(), ref.Day(), 0, 0, 0, 0, ref.Location())
}

func StartOfWeek() time.Time {
	ref := time.Now()
	year, week := ref.ISOWeek()
	return firstDayOfISOWeek(year, week, ref.Location())
}

func StartOfMonth() time.Time {
	ref := time.Now()
	return time.Date(ref.Year(), ref.Month(), 1, 0, 0, 0, 0, ref.Location())
}

func StartOfYear() time.Time {
	ref := time.Now()
	return time.Date(ref.Year(), time.January, 1, 0, 0, 0, 0, ref.Location())
}

func FmtWakatimeDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	return fmt.Sprintf("%d hrs %d mins", h, m)
}

// https://stackoverflow.com/a/18632496
func firstDayOfISOWeek(year int, week int, timezone *time.Location) time.Time {
	date := time.Date(year, 0, 0, 0, 0, 0, 0, timezone)
	isoYear, isoWeek := date.ISOWeek()
	for date.Weekday() != time.Monday { // iterate back to Monday
		date = date.AddDate(0, 0, -1)
		isoYear, isoWeek = date.ISOWeek()
	}
	for isoYear < year { // iterate forward to the first day of the first week
		date = date.AddDate(0, 0, 1)
		isoYear, isoWeek = date.ISOWeek()
	}
	for isoWeek < week { // iterate forward to the first day of the given week
		date = date.AddDate(0, 0, 1)
		isoYear, isoWeek = date.ISOWeek()
	}
	return date
}
