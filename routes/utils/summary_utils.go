package utils

import (
	"net/http"
	"strings"

	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/services"
	summarytypes "github.com/muety/wakapi/types"
)

func LoadUserSummary(ss services.ISummaryService, r *http.Request) (*models.Summary, error, int) {
	summaryParams, err := helpers.ParseSummaryParams(r)
	if err != nil {
		return nil, err, http.StatusBadRequest
	}
	return LoadUserSummaryByParams(ss, summaryParams)
}

func LoadUserSummaryByParams(ss services.ISummaryService, params *models.SummaryParams) (*models.Summary, error, int) {
	var summary *models.Summary
	var err error

	request := summarytypes.NewSummaryRequest(params.From, params.To, params.User).WithFilters(params.Filters)
	if params.Recompute {
		request = request.WithoutCache()
	}
	options := summarytypes.DefaultProcessingOptions()
	
	if params.Recompute {
		summary, err = ss.ComputeFromDurations(request)
		if err == nil {
			// Apply aliases and project labels manually for computed summaries
			summary, err = ss.Generate(request, options)
		}
	} else {
		summary, err = ss.Generate(request, options)
	}
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}

	summary.FromTime = models.CustomTime(summary.FromTime.T().In(params.User.TZ()))
	summary.ToTime = models.CustomTime(summary.ToTime.T().In(params.User.TZ()))

	return summary, nil, http.StatusOK
}

func FilterColors(all map[string]string, haystack models.SummaryItems) map[string]string {
	subset := make(map[string]string)
	for _, item := range haystack {
		if c, ok := all[strings.ToLower(item.Key)]; ok {
			subset[strings.ToLower(item.Key)] = c
		}
	}
	return subset
}
