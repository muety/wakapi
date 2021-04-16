package utils

import (
	"fmt"
	"time"
)

func StartOfToday() time.Time {
	return StartOfDay(time.Now())
}

func StartOfDay(date time.Time) time.Time {
	return FloorDate(date)
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

// FloorDate rounds date down to the start of the day
func FloorDate(date time.Time) time.Time {
	return date.Truncate(24 * time.Hour)
}

// CeilDate rounds date up to the start of next day if date is not already a start (00:00:00)
func CeilDate(date time.Time) time.Time {
	floored := FloorDate(date)
	if floored == date {
		return floored
	}
	return floored.Add(24 * time.Hour)
}

func SplitRangeByDays(from time.Time, to time.Time) [][]time.Time {
	intervals := make([][]time.Time, 0)

	for t1 := from; t1.Before(to); {
		t2 := StartOfDay(t1).Add(24 * time.Hour)
		if t2.After(to) {
			t2 = to
		}
		intervals = append(intervals, []time.Time{t1, t2})
		t1 = t2
	}

	return intervals
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
