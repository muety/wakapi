package utils

import (
	"errors"
	"github.com/muety/wakapi/models"
	"net/http"
	"time"
)

func ParseSummaryParams(r *http.Request) (*models.SummaryParams, error) {
	user := r.Context().Value(models.UserKey).(*models.User)
	params := r.URL.Query()
	interval := params.Get("interval")

	from, err := ParseDate(params.Get("from"))
	if err != nil {
		switch interval {
		case models.IntervalToday:
			from = StartOfToday()
		case models.IntervalLastDay:
			from = StartOfToday().Add(-24 * time.Hour)
		case models.IntervalLastWeek:
			from = StartOfWeek()
		case models.IntervalLastMonth:
			from = StartOfMonth()
		case models.IntervalLastYear:
			from = StartOfYear()
		case models.IntervalAny:
			from = time.Time{}
		default:
			return nil, errors.New("missing 'from' parameter")
		}
	}

	live := (params.Get("live") != "" && params.Get("live") != "false") || interval == models.IntervalToday

	recompute := params.Get("recompute") != "" && params.Get("recompute") != "false"

	to := StartOfToday()
	if live {
		to = time.Now()
	}

	return &models.SummaryParams{
		From:      from,
		To:        to,
		User:      user,
		Recompute: recompute,
	}, nil
}
