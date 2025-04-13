package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/muety/wakapi/helpers"

	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	v1 "github.com/muety/wakapi/models/compat/wakatime/v1"
)

// TODO: support filtering (requires https://github.com/muety/wakapi/issues/108)

// @Summary Retrieve statistics for a given user
// @Description Mimics https://wakatime.com/developers#stats
// @ID get-wakatimes-tats
// @Tags wakatime
// @Produce json
// @Param user path string true "User ID to fetch data for (or 'current')"
// @Param range path string false "Range interval identifier" Enums(today, yesterday, week, month, year, 7_days, last_7_days, 30_days, last_30_days, 6_months, last_6_months, 12_months, last_12_months, last_year, any, all_time)
// @Param project query string false "Project to filter by"
// @Param language query string false "Language to filter by"
// @Param editor query string false "Editor to filter by"
// @Param operating_system query string false "OS to filter by"
// @Param machine query string false "Machine to filter by"
// @Param label query string false "Project label to filter by"
// @Security ApiKeyAuth
// @Success 200 {object} v1.StatsViewModel
// @Router /compat/wakatime/v1/users/{user}/stats/{range} [get]
func (a *APIv1) GetUserStats(w http.ResponseWriter, r *http.Request) {
	userParam := chi.URLParam(r, "user")
	rangeParam := chi.URLParam(r, "range")
	var authorizedUser, requestedUser *models.User

	authorizedUser = middlewares.GetPrincipal(r)
	if authorizedUser != nil && userParam == "current" {
		userParam = authorizedUser.ID
	}

	requestedUser, err := a.services.Users().GetUserById(userParam)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("user not found"))
		return
	}

	// if no range was requested, get the maximum allowed range given the users max shared days, otherwise default to past 7 days (which will fail in the next step, because user didn't allow any sharing)
	// this "floors" the user's maximum shared date to the supported range buckets (e.g. if user opted to share 12 days, we'll still fallback to "last_7_days") for consistency with wakatime
	if rangeParam == "" {
		if _, userRange := helpers.ResolveMaximumRange(requestedUser.ShareDataMaxDays); userRange != nil {
			rangeParam = (*userRange)[1]
		} else {
			rangeParam = (*models.IntervalPast7Days)[1]
		}
	}

	err, rangeFrom, rangeTo := helpers.ResolveIntervalRawTZ(rangeParam, requestedUser.TZ())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid range"))
		return
	}

	minStart := rangeTo.AddDate(0, 0, -requestedUser.ShareDataMaxDays)
	if (authorizedUser == nil || requestedUser.ID != authorizedUser.ID) &&
		rangeFrom.Before(minStart) && requestedUser.ShareDataMaxDays >= 0 {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("requested time range too broad"))
		return
	}

	summaryParams := &models.SummaryParams{
		From:      rangeFrom,
		To:        rangeTo,
		User:      authorizedUser,
		Recompute: false,
	}

	summary, err, status := a.loadUserSummary(summaryParams, helpers.ParseSummaryFilters(r))
	if err != nil {
		w.WriteHeader(status)
		w.Write([]byte(err.Error()))
		return
	}

	stats := v1.NewStatsFrom(summary, &models.Filters{})
	stats.Data.Range = rangeParam
	stats.Data.HumanReadableRange = helpers.MustParseInterval(rangeParam).GetHumanReadable()
	stats.Data.IsCodingActivityVisible = requestedUser.ShareDataMaxDays != 0
	stats.Data.IsOtherUsageVisible = requestedUser.AnyDataShared()

	if authorizedUser == nil || requestedUser.ID != authorizedUser.ID {
		// post filter stats according to user's given sharing permissions
		if !requestedUser.ShareEditors {
			stats.Data.Editors = make([]*v1.SummariesEntry, 0)
		}
		if !requestedUser.ShareLanguages {
			stats.Data.Languages = make([]*v1.SummariesEntry, 0)
		}
		if !requestedUser.ShareProjects {
			stats.Data.Projects = make([]*v1.SummariesEntry, 0)
		}
		if !requestedUser.ShareOSs {
			stats.Data.OperatingSystems = make([]*v1.SummariesEntry, 0)
		}
		if !requestedUser.ShareMachines {
			stats.Data.Machines = make([]*v1.SummariesEntry, 0)
		}
	}

	helpers.RespondJSON(w, r, http.StatusOK, stats)
}
