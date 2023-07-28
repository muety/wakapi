package helpers

import (
	"errors"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"time"
)

func ParseInterval(interval string) (*models.IntervalKey, error) {
	for _, i := range models.AllIntervals {
		if i.HasAlias(interval) {
			return i, nil
		}
	}
	return nil, errors.New("not a valid interval")
}

func MustParseInterval(interval string) *models.IntervalKey {
	key, _ := ParseInterval(interval)
	return key
}

func MustResolveIntervalRawTZ(interval string, tz *time.Location) (from, to time.Time) {
	_, from, to = ResolveIntervalRawTZ(interval, tz)
	return from, to
}

func ResolveIntervalRawTZ(interval string, tz *time.Location) (err error, from, to time.Time) {
	parsed, err := ParseInterval(interval)
	if err != nil {
		return err, time.Time{}, time.Time{}
	}
	return ResolveIntervalTZ(parsed, tz)
}

func ResolveIntervalTZ(interval *models.IntervalKey, tz *time.Location) (err error, from, to time.Time) {
	now := time.Now().In(tz)
	to = now

	switch interval {
	case models.IntervalToday:
		from = utils.BeginOfToday(tz)
	case models.IntervalYesterday:
		from = utils.BeginOfToday(tz).Add(-24 * time.Hour)
		to = utils.BeginOfToday(tz)
	case models.IntervalPastDay:
		from = now.Add(-24 * time.Hour)
	case models.IntervalThisWeek:
		from = utils.BeginOfThisWeek(tz)
	case models.IntervalLastWeek:
		from = utils.BeginOfThisWeek(tz).AddDate(0, 0, -7)
		to = utils.BeginOfThisWeek(tz)
	case models.IntervalThisMonth:
		from = utils.BeginOfThisMonth(tz)
	case models.IntervalLastMonth:
		from = utils.BeginOfThisMonth(tz).AddDate(0, -1, 0)
		to = utils.BeginOfThisMonth(tz)
	case models.IntervalThisYear:
		from = utils.BeginOfThisYear(tz)
	case models.IntervalPast7Days:
		from = now.AddDate(0, 0, -7)
	case models.IntervalPast7DaysYesterday:
		from = utils.BeginOfToday(tz).AddDate(0, 0, -1).AddDate(0, 0, -7)
		to = utils.BeginOfToday(tz).AddDate(0, 0, -1)
	case models.IntervalPast14Days:
		from = now.AddDate(0, 0, -14)
	case models.IntervalPast30Days:
		from = now.AddDate(0, 0, -30)
	case models.IntervalPast6Months:
		from = now.AddDate(0, -6, 0)
	case models.IntervalPast12Months:
		from = now.AddDate(0, -12, 0)
	case models.IntervalAny:
		from = time.Time{}
	default:
		err = errors.New("invalid interval")
	}

	return err, from, to
}

// ResolveClosestRange returns the interval label (e.g. "last_7_days") of the maximum allowed range when having opted to share this many days or an error for days == 0.
func ResolveMaximumRange(days int) (error, *models.IntervalKey) {
	if days == 0 {
		return errors.New("no matching interval"), nil
	}
	if days < 0 {
		return nil, models.IntervalAny
	}
	if days < 7 {
		return nil, models.IntervalPastDay
	}
	if days < 14 {
		return nil, models.IntervalPast7Days
	}
	if days < 30 {
		return nil, models.IntervalPast14Days
	}
	if days < 181 { // 3*31 + 2*30 + 1*28
		return nil, models.IntervalPast30Days
	}
	if days < 365 { // 7*31 + 4*30 + 1*28
		return nil, models.IntervalPast6Months
	}
	return nil, models.IntervalPast12Months
}
