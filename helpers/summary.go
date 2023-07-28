package helpers

import (
	"errors"
	"github.com/muety/wakapi/models"
	"net/http"
	"time"
)

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
	if q := r.URL.Query().Get("entity"); q != "" {
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
