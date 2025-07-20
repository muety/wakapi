package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/datetime"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/internal/utilities"
	"github.com/muety/wakapi/models"
	v1 "github.com/muety/wakapi/models/compat/wakatime/v1"
	summarytypes "github.com/muety/wakapi/types"
	"github.com/muety/wakapi/utils"
)

// TODO: Support parameters: project, branches, timeout, writes_only
// See https://wakatime.com/developers#summaries.
// Timezone can be specified via an offset suffix (e.g. +02:00) in date strings.
// Requires https://github.com/muety/wakapi/issues/108.

// @Summary Retrieve WakaTime-compatible summaries
// @Description Mimics https://wakatime.com/developers#summaries.
// @ID get-wakatime-summaries
// @Tags wakatime
// @Produce json
// @Param user path string true "User ID to fetch data for (or 'current')"
// @Param range query string false "Range interval identifier" Enums(today, yesterday, week, month, year, 7_days, last_7_days, 30_days, last_30_days, 6_months, last_6_months, 12_months, last_12_months, last_year, any, all_time)
// @Param start query string false "Start date (e.g. '2021-02-07')"
// @Param end query string false "End date (e.g. '2021-02-08')"
// @Param project query string false "Project to filter by"
// @Param language query string false "Language to filter by"
// @Param editor query string false "Editor to filter by"
// @Param operating_system query string false "OS to filter by"
// @Param machine query string false "Machine to filter by"
// @Param label query string false "Project label to filter by"
// @Security ApiKeyAuth
// @Success 200 {object} v1.SummariesViewModel
// @Router /compat/wakatime/v1/users/{user}/summaries [get]
func (a *APIv1) GetSummaries(w http.ResponseWriter, r *http.Request) {
	user, err := utilities.CheckEffectiveUser(w, r, a.services.Users(), "current")
	if err != nil {
		return // response was already sent by util function
	}

	start, end, err, status := a.ComputeTimeRange(r, user)
	if err != nil {
		w.WriteHeader(status)
		w.Write([]byte(err.Error()))
		return
	}

	summaries, status, err := a.loadUserSummaries(r, user)
	if err != nil {
		w.WriteHeader(status)
		w.Write([]byte(err.Error()))
		return
	}

	writePercentage, err := a.services.Summary().GetHeartbeatsWritePercentage(user.ID, start, end)
	if err != nil {
		w.WriteHeader(status)
		w.Write([]byte(err.Error()))
		return
	}

	vm := v1.NewSummariesFrom(summaries)
	vm.WritePercentage = writePercentage

	helpers.RespondJSON(w, r, http.StatusOK, vm)
}

// ComputeTimeRange extracts and computes the start and end time from request parameters
// Returns the computed start and end times, along with any error and HTTP status code
func (a *APIv1) ComputeTimeRange(r *http.Request, user *models.User) (time.Time, time.Time, error, int) {
	params := r.URL.Query()
	rangeParam, startParam, endParam, tzParam := params.Get("range"), params.Get("start"), params.Get("end"), params.Get("timezone")

	timezone := user.TZ()
	if tzParam != "" {
		if tz, err := time.LoadLocation(tzParam); err == nil {
			timezone = tz
		}
	}

	var start, end time.Time
	if rangeParam != "" {
		// range param takes precedence
		if err, parsedFrom, parsedTo := helpers.ResolveIntervalRawTZ(rangeParam, timezone); err == nil {
			start, end = parsedFrom, parsedTo
		} else {
			return time.Time{}, time.Time{}, errors.New("invalid 'range' parameter"), http.StatusBadRequest
		}
	} else if err, parsedFrom, parsedTo := helpers.ResolveIntervalRawTZ(startParam, timezone); err == nil && startParam == endParam {
		// also accept start param to be a range param
		start, end = parsedFrom, parsedTo
	} else {
		// eventually, consider start and end params a date
		var err error

		start, err = helpers.ParseDateTimeTZ(strings.Replace(startParam, " ", "+", 1), timezone)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("missing required 'start' parameter"), http.StatusBadRequest
		}

		end, err = helpers.ParseDateTimeTZ(strings.Replace(endParam, " ", "+", 1), timezone)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("missing required 'end' parameter"), http.StatusBadRequest
		}
	}

	// wakatime interprets end date as "inclusive", wakapi usually as "exclusive"
	// i.e. for wakatime, an interval 2021-04-29 - 2021-04-29 is actually 2021-04-29 - 2021-04-30,
	// while for wakapi it would be empty
	// see https://github.com/muety/wakapi/issues/192
	end = datetime.EndOfDay(end)

	if !end.After(start) {
		return time.Time{}, time.Time{}, errors.New("'end' date must be after 'start' date"), http.StatusBadRequest
	}

	return start, end, nil, http.StatusOK
}

func (a *APIv1) loadUserSummaries(r *http.Request, user *models.User) ([]*models.Summary, int, error) {
	start, end, err, status := a.ComputeTimeRange(r, user)
	if err != nil {
		return nil, status, err
	}

	overallParams := &models.SummaryParams{
		From: start,
		To:   end,
	}

	intervals := utils.SplitRangeByDays(overallParams.From, overallParams.To)
	summaries := make([]*models.Summary, len(intervals))

	// filtering
	filters := helpers.ParseSummaryFilters(r)

	for i, interval := range intervals {
		request := summarytypes.NewSummaryRequest(interval.Start, interval.End, user).WithFilters(filters)
		if end.After(time.Now()) {
			request = request.WithoutCache()
		}
		options := summarytypes.DefaultProcessingOptions()
		summary, err := a.services.Summary().Generate(request, options)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		// wakatime returns requested instead of actual summary range
		summary.FromTime = models.CustomTime(interval.Start)
		summary.ToTime = models.CustomTime(interval.End.Add(-1 * time.Second))
		summaries[i] = summary
	}

	return summaries, http.StatusOK, err
}
