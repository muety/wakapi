package utils

import (
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/services"
	"net/http"
)

func LoadUserSummary(ss services.ISummaryService, r *http.Request) (*models.Summary, error, int) {
	summaryParams, err := helpers.ParseSummaryParams(r)
	if err != nil {
		return nil, err, http.StatusBadRequest
	}
	return LoadUserSummaryByParams(ss, summaryParams)
}

func LoadUserSummaryByParams(ss services.ISummaryService, params *models.SummaryParams) (*models.Summary, error, int) {
	var retrieveSummary services.SummaryRetriever = ss.Retrieve
	if params.Recompute {
		retrieveSummary = ss.Summarize
	}

	summary, err := ss.Aliased(
		params.From,
		params.To,
		params.User,
		retrieveSummary,
		params.Filters,
		params.Recompute,
	)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}

	summary.FromTime = models.CustomTime(summary.FromTime.T().In(params.User.TZ()))
	summary.ToTime = models.CustomTime(summary.ToTime.T().In(params.User.TZ()))

	return summary, nil, http.StatusOK
}
