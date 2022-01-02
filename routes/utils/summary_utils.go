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

	summary, err := ss.Aliased(summaryParams.From, summaryParams.To, summaryParams.User, retrieveSummary, summaryParams.Filters, summaryParams.Recompute)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}

	summary.FromTime = models.CustomTime(summary.FromTime.T().In(user.TZ()))
	summary.ToTime = models.CustomTime(summary.ToTime.T().In(user.TZ()))

	return summary, nil, http.StatusOK
}
