package utils

import (
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"net/http"
)

func LoadUserSummary(ss services.ISummaryService, r *http.Request) (*models.Summary, error, int) {
	user := middlewares.GetPrincipal(r)
	summaryParams, err := utils.ParseSummaryParams(r)
	if err != nil {
		return nil, err, http.StatusBadRequest
	}

	var retrieveSummary services.SummaryRetriever = ss.Retrieve
	if summaryParams.Recompute {
		retrieveSummary = ss.Summarize
	}

	summary, err := ss.Aliased(summaryParams.From, summaryParams.To, summaryParams.User, retrieveSummary, ParseFilters(r), summaryParams.Recompute)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}

	summary.FromTime = models.CustomTime(summary.FromTime.T().In(user.TZ()))
	summary.ToTime = models.CustomTime(summary.ToTime.T().In(user.TZ()))

	return summary, nil, http.StatusOK
}

func ParseFilters(r *http.Request) *models.Filters {
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
