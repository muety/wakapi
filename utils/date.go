package utils

import (
	"os"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/datetime"
)

func UnixEra() time.Time {
	return time.Unix(0, 0)
}

func MustParseTime(layout, value string) time.Time {
	t, _ := time.Parse(layout, value)
	return t
}

func BeginOfToday(tz *time.Location) time.Time {
	return datetime.BeginOfDay(time.Now().In(tz))
}

func BeginOfThisWeek(tz *time.Location, startOfWeek time.Weekday) time.Time {
	return datetime.BeginOfWeek(time.Now().In(tz), startOfWeek)
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
		// https://github.com/muety/wakapi/issues/779
		if t2 == t1 {
			t1 = datetime.BeginOfDay(t1).Add(24 * time.Hour)
			continue
		}
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

// SetZone overwrites a date's timezone without reinterpreting at
func SetZone(t time.Time, loc *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
}

func ResolveIANAZone(tz *time.Location) string {
	if tz.String() == "Local" {
		link, err := os.Readlink("/etc/localtime") // only on Linux / Unix
		if err == nil {
			if parts := strings.Split(link, "zoneinfo/"); len(parts) > 1 {
				return parts[1]
			}
		}
	}
	return tz.String()
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
