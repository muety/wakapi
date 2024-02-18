package utils

import (
	"github.com/duke-git/lancet/v2/datetime"
	"strings"
	"time"
)

func MustParseTime(layout, value string) time.Time {
	t, _ := time.Parse(layout, value)
	return t
}

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

// LocalTZOffset returns the time difference between server local time and UTC
func LocalTZOffset() time.Duration {
	_, offset := time.Now().Zone()
	return time.Duration(offset * int(time.Second))
}

func ParseWeekday(s string) time.Weekday {
	switch strings.ToLower(s) {
	case "mon", strings.ToLower(time.Monday.String()):
		return time.Monday
	case "tue", strings.ToLower(time.Tuesday.String()):
		return time.Tuesday
	case "wed", strings.ToLower(time.Wednesday.String()):
		return time.Wednesday
	case "thu", strings.ToLower(time.Thursday.String()):
		return time.Thursday
	case "fri", strings.ToLower(time.Friday.String()):
		return time.Friday
	case "sat", strings.ToLower(time.Saturday.String()):
		return time.Saturday
	case "sun", strings.ToLower(time.Sunday.String()):
		return time.Sunday
	}
	return time.Monday
}
