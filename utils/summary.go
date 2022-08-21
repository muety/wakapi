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
	now := time.Now().In(tz)
	to = now

	switch interval {
	case models.IntervalToday:
		from = BeginOfToday(tz)
	case models.IntervalYesterday:
		from = BeginOfToday(tz).Add(-24 * time.Hour)
		to = BeginOfToday(tz)
	case models.IntervalThisWeek:
		from = BeginOfThisWeek(tz)
	case models.IntervalLastWeek:
		from = BeginOfThisWeek(tz).AddDate(0, 0, -7)
		to = BeginOfThisWeek(tz)
	case models.IntervalThisMonth:
		from = BeginOfThisMonth(tz)
	case models.IntervalLastMonth:
		from = BeginOfThisMonth(tz).AddDate(0, -1, 0)
		to = BeginOfThisMonth(tz)
	case models.IntervalThisYear:
		from = BeginOfThisYear(tz)
	case models.IntervalPast7Days:
		from = now.AddDate(0, 0, -7)
	case models.IntervalPast7DaysYesterday:
		from = BeginOfToday(tz).AddDate(0, 0, -1).AddDate(0, 0, -7)
		to = BeginOfToday(tz).AddDate(0, 0, -1)
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

	filters := ParseSummaryFilters(r)

	return &models.SummaryParams{
		From:      from,
		To:        to,
		User:      user,
		Recompute: recompute,
		Filters:   filters,
	}, nil
}

func ParseSummaryFilters(r *http.Request) *models.Filters {
	filters := &models.Filters{}
	if q := r.URL.Query().Get("project"); q != "" {
		filters.With(models.SummaryProject, q)
	}
	if q := r.URL.Query().Get("language"); q != "" {
		filters.With(models.SummaryLanguage, q)
	}
	if q := r.URL.Query().Get("editor"); q != "" {
		filters.With(models.SummaryEditor, q)
	}
	if q := r.URL.Query().Get("machine"); q != "" {
		filters.With(models.SummaryMachine, q)
	}
	if q := r.URL.Query().Get("operating_system"); q != "" {
		filters.With(models.SummaryOS, q)
	}
	if q := r.URL.Query().Get("label"); q != "" {
		filters.With(models.SummaryLabel, q)
	}
	if q := r.URL.Query().Get("branch"); q != "" {
		filters.With(models.SummaryBranch, q)
	}
	return filters
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
