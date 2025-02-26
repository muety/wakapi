package v1

import (
	"errors"
	"github.com/duke-git/lancet/v2/datetime"
	"github.com/go-chi/chi/v5"
	"github.com/muety/wakapi/helpers"
	"net/http"
	"strings"
	"time"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	v1 "github.com/muety/wakapi/models/compat/wakatime/v1"
	routeutils "github.com/muety/wakapi/routes/utils"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
)

type SummariesHandler struct {
	config      *conf.Config
	userSrvc    services.IUserService
	summarySrvc services.ISummaryService
}

func NewSummariesHandler(userService services.IUserService, summaryService services.ISummaryService) *SummariesHandler {
	return &SummariesHandler{
		userSrvc:    userService,
		summarySrvc: summaryService,
		config:      conf.Get(),
	}
}

func (h *SummariesHandler) RegisterRoutes(router chi.Router) {
	router.Group(func(r chi.Router) {
		r.Use(middlewares.NewAuthenticateMiddleware(h.userSrvc).Handler)
		r.Get("/compat/wakatime/v1/users/{user}/summaries", h.Get)
	})
}

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
func (h *SummariesHandler) Get(w http.ResponseWriter, r *http.Request) {
	user, err := routeutils.CheckEffectiveUser(w, r, h.userSrvc, "current")
	if err != nil {
		return // response was already sent by util function
	}

	summaries, err, status := h.loadUserSummaries(r, user)
	if err != nil {
		w.WriteHeader(status)
		w.Write([]byte(err.Error()))
		return
	}

	vm := v1.NewSummariesFrom(summaries)
	helpers.RespondJSON(w, r, http.StatusOK, vm)
}

func (h *SummariesHandler) loadUserSummaries(r *http.Request, user *models.User) ([]*models.Summary, error, int) {
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
			return nil, errors.New("invalid 'range' parameter"), http.StatusBadRequest
		}
	} else if err, parsedFrom, parsedTo := helpers.ResolveIntervalRawTZ(startParam, timezone); err == nil && startParam == endParam {
		// also accept start param to be a range param
		start, end = parsedFrom, parsedTo
	} else {
		// eventually, consider start and end params a date
		var err error

		start, err = helpers.ParseDateTimeTZ(strings.Replace(startParam, " ", "+", 1), timezone)
		if err != nil {
			return nil, errors.New("missing required 'start' parameter"), http.StatusBadRequest
		}

		end, err = helpers.ParseDateTimeTZ(strings.Replace(endParam, " ", "+", 1), timezone)
		if err != nil {
			return nil, errors.New("missing required 'end' parameter"), http.StatusBadRequest
		}
	}

	// wakatime interprets end date as "inclusive", wakapi usually as "exclusive"
	// i.e. for wakatime, an interval 2021-04-29 - 2021-04-29 is actually 2021-04-29 - 2021-04-30,
	// while for wakapi it would be empty
	// see https://github.com/muety/wakapi/issues/192
	end = datetime.EndOfDay(end)

	if !end.After(start) {
		return nil, errors.New("'end' date must be after 'start' date"), http.StatusBadRequest
	}

	overallParams := &models.SummaryParams{
		From: start,
		To:   end,
		User: user,
	}

	intervals := utils.SplitRangeByDays(overallParams.From, overallParams.To)
	summaries := make([]*models.Summary, len(intervals))

	// filtering
	filters := helpers.ParseSummaryFilters(r)

	for i, interval := range intervals {
		summary, err := h.summarySrvc.Aliased(interval[0], interval[1], user, h.summarySrvc.Retrieve, filters, nil, end.After(time.Now()))
		if err != nil {
			return nil, err, http.StatusInternalServerError
		}
		// wakatime returns requested instead of actual summary range
		summary.FromTime = models.CustomTime(interval[0])
		summary.ToTime = models.CustomTime(interval[1].Add(-1 * time.Second))
		summaries[i] = summary
	}

	return summaries, nil, http.StatusOK
}
