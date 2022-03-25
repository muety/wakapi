package utils

import (
	"fmt"
	"github.com/duke-git/lancet/v2/datetime"
	"time"
)

func BeginOfToday(tz *time.Location) time.Time {
	return datetime.BeginOfDay(time.Now().In(tz))
}

func BeginOfThisWeek(tz *time.Location) time.Time {
	return datetime.BeginOfWeek(time.Now().In(tz))
}

func BeginOfThisMonth(tz *time.Location) time.Time {
	return datetime.BeginOfMonth(time.Now().In(tz))
}

func BeginOfThisYear(tz *time.Location) time.Time {
	return datetime.BeginOfYear(time.Now().In(tz))
}

// CeilDate rounds date up to the start of next day if date is not already a start (00:00:00)
func CeilDate(date time.Time) time.Time {
	floored := datetime.BeginOfDay(date)
	if floored == date {
		return floored
	}
	return floored.AddDate(0, 0, 1)
}

// SplitRangeByDays creates a slice of intervals between from and to, each of which is at max of 24 hours length and has its split at midnight
func SplitRangeByDays(from time.Time, to time.Time) [][]time.Time {
	intervals := make([][]time.Time, 0)

	for t1 := from; t1.Before(to); {
		t2 := datetime.BeginOfDay(t1).AddDate(0, 0, 1)
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

// LocalTZOffset returns the time difference between server local time and UTC
func LocalTZOffset() time.Duration {
	_, offset := time.Now().Zone()
	return time.Duration(offset * int(time.Second))
}
