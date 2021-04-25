package utils

import (
	"errors"
	"github.com/muety/wakapi/models"
	"net/http"
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
	to = time.Now().In(tz)

	switch interval {
	case models.IntervalToday:
		from = StartOfToday(tz)
	case models.IntervalYesterday:
		from = StartOfToday(tz).Add(-24 * time.Hour)
		to = StartOfToday(tz)
	case models.IntervalThisWeek:
		from = StartOfThisWeek(tz)
	case models.IntervalLastWeek:
		from = StartOfThisWeek(tz).AddDate(0, 0, -7)
		to = StartOfThisWeek(tz)
	case models.IntervalThisMonth:
		from = StartOfThisMonth(tz)
	case models.IntervalLastMonth:
		from = StartOfThisMonth(tz).AddDate(0, -1, 0)
		to = StartOfThisMonth(tz)
	case models.IntervalThisYear:
		from = StartOfThisYear(tz)
	case models.IntervalPast7Days:
		from = StartOfToday(tz).AddDate(0, 0, -7)
	case models.IntervalPast7DaysYesterday:
		from = StartOfToday(tz).AddDate(0, 0, -1).AddDate(0, 0, -7)
		to = StartOfToday(tz).AddDate(0, 0, -1)
	case models.IntervalPast14Days:
		from = StartOfToday(tz).AddDate(0, 0, -14)
	case models.IntervalPast30Days:
		from = StartOfToday(tz).AddDate(0, 0, -30)
	case models.IntervalPast12Months:
		from = StartOfToday(tz).AddDate(0, -12, 0)
	case models.IntervalAny:
		from = time.Time{}
	default:
		err = errors.New("invalid interval")
	}

	return err, from, to
}

func ParseSummaryParams(r *http.Request) (*models.SummaryParams, error) {
	user := extractUser(r)
	params := r.URL.Query()

	var err error
	var from, to time.Time

	if interval := params.Get("interval"); interval != "" {
		err, from, to = ResolveIntervalRawTZ(interval, user.TZ())
	} else if start := params.Get("start"); start != "" {
		err, from, to = ResolveIntervalRawTZ(start, user.TZ())
	} else {
		from, err = ParseDateTimeTZ(params.Get("from"), user.TZ())
		if err != nil {
			return nil, errors.New("missing or invalid 'from' parameter")
		}

		to, err = ParseDateTimeTZ(params.Get("to"), user.TZ())
		if err != nil {
			return nil, errors.New("missing or invalid 'to' parameter")
		}
	}

	recompute := params.Get("recompute") != "" && params.Get("recompute") != "false"

	return &models.SummaryParams{
		From:      from,
		To:        to,
		User:      user,
		Recompute: recompute,
	}, nil
}

func extractUser(r *http.Request) *models.User {
	type principalGetter interface {
		GetPrincipal() *models.User
	}
	if p := r.Context().Value("principal"); p != nil {
		return p.(principalGetter).GetPrincipal()
	}
	return nil
}
