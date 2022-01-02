package utils

import (
	"fmt"
	"time"
)

func StartOfDay(date time.Time) time.Time {
	return FloorDate(date)
}

func StartOfToday(tz *time.Location) time.Time {
	return StartOfDay(FloorDate(time.Now().In(tz)))
}

func EndOfDay(date time.Time) time.Time {
	floored := FloorDate(date)
	if floored == date {
		date = date.Add(1 * time.Second)
	}
	return CeilDate(date)
}

func EndOfToday(tz *time.Location) time.Time {
	return EndOfDay(time.Now().In(tz))
}

func StartOfThisWeek(tz *time.Location) time.Time {
	return StartOfWeek(time.Now().In(tz))
}

func StartOfWeek(date time.Time) time.Time {
	year, week := date.ISOWeek()
	return firstDayOfISOWeek(year, week, date.Location())
}

func StartOfThisMonth(tz *time.Location) time.Time {
	return StartOfMonth(time.Now().In(tz))
}

func StartOfMonth(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
}

func StartOfThisYear(tz *time.Location) time.Time {
	return StartOfYear(time.Now().In(tz))
}

func StartOfYear(date time.Time) time.Time {
	return time.Date(date.Year(), time.January, 1, 0, 0, 0, 0, date.Location())
}

// FloorDate rounds date down to the start of the day and keeps the time zone
func FloorDate(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
}

// FloorDateHour rounds date down to the start of the current hour and keeps the time zone
func FloorDateHour(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), date.Hour(), 0, 0, 0, date.Location())
}

// CeilDate rounds date up to the start of next day if date is not already a start (00:00:00)
func CeilDate(date time.Time) time.Time {
	floored := FloorDate(date)
	if floored == date {
		return floored
	}
	return floored.AddDate(0, 0, 1)
}

// SetLocation resets the time zone information of a date without converting it, i.e. 19:00 UTC will result in 19:00 CET, for instance
func SetLocation(date time.Time, tz *time.Location) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, tz)
}

// WithOffset adds the time zone difference between Local and tz to a date, i.e. 19:00 UTC will result in 21:00 CET (or 22:00 CEST), for instance
func WithOffset(date time.Time, tz *time.Location) time.Time {
	now := time.Now()
	_, localOffset := now.Zone()
	_, targetOffset := now.In(tz).Zone()
	dateTz := date.Add(time.Duration((targetOffset - localOffset) * int(time.Second)))
	return time.Date(dateTz.Year(), dateTz.Month(), dateTz.Day(), dateTz.Hour(), dateTz.Minute(), dateTz.Second(), dateTz.Nanosecond(), dateTz.Location()).In(tz)
}

// SplitRangeByDays creates a slice of intervals between from and to, each of which is at max of 24 hours length and has its split at midnight
func SplitRangeByDays(from time.Time, to time.Time) [][]time.Time {
	intervals := make([][]time.Time, 0)

	for t1 := from; t1.Before(to); {
		t2 := StartOfDay(t1).AddDate(0, 0, 1)
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
