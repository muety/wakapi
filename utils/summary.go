package utils

import (
	"errors"
	"github.com/muety/wakapi/models"
	"net/http"
	"time"
)

func ResolveInterval(interval string) (err error, from, to time.Time) {
	to = time.Now()

	switch interval {
	case models.IntervalToday:
		from = StartOfToday()
	case models.IntervalYesterday:
		from = StartOfToday().Add(-24 * time.Hour)
		to = StartOfToday()
	case models.IntervalThisWeek:
		from = StartOfWeek()
	case models.IntervalThisMonth:
		from = StartOfMonth()
	case models.IntervalThisYear:
		from = StartOfYear()
	case models.IntervalPast7Days:
		from = StartOfToday().AddDate(0, 0, -7)
	case models.IntervalPast30Days:
		from = StartOfToday().AddDate(0, 0, -30)
	case models.IntervalPast12Months:
		from = StartOfToday().AddDate(0, -12, 0)
	case models.IntervalAny:
		from = time.Time{}
	default:
		err = errors.New("invalid interval")
	}

	return err, from, to
}

func ParseSummaryParams(r *http.Request) (*models.SummaryParams, error) {
	user := r.Context().Value(models.UserKey).(*models.User)
	params := r.URL.Query()

	var err error
	var from, to time.Time

	if interval := params.Get("interval"); interval != "" {
		err, from, to = ResolveInterval(interval)
	} else {
		from, err = ParseDate(params.Get("from"))
		if err != nil {
			return nil, errors.New("missing 'from' parameter")
		}

		to, err = ParseDate(params.Get("to"))
		if err != nil {
			return nil, errors.New("missing 'to' parameter")
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
