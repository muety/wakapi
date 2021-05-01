package v1

import (
	"errors"
	"github.com/gorilla/mux"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	v1 "github.com/muety/wakapi/models/compat/wakatime/v1"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"net/http"
	"strings"
	"time"
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

func (h *SummariesHandler) RegisterRoutes(router *mux.Router) {
	r := router.PathPrefix("/compat/wakatime/v1/users/{user}/summaries").Subrouter()
	r.Use(
		middlewares.NewAuthenticateMiddleware(h.userSrvc).Handler,
	)
	r.Path("").Methods(http.MethodGet).HandlerFunc(h.Get)
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
// @Param range query string false "Range interval identifier" Enums(today, yesterday, week, month, year, 7_days, last_7_days, 30_days, last_30_days, 12_months, last_12_months, any)
// @Param start query string false "Start date (e.g. '2021-02-07')"
// @Param end query string false "End date (e.g. '2021-02-08')"
// @Security ApiKeyAuth
// @Success 200 {object} v1.SummariesViewModel
// @Router /compat/wakatime/v1/users/{user}/summaries [get]
func (h *SummariesHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestedUser := vars["user"]
	authorizedUser := middlewares.GetPrincipal(r)

	if requestedUser != authorizedUser.ID && requestedUser != "current" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	summaries, err, status := h.loadUserSummaries(r)
	if err != nil {
		w.WriteHeader(status)
		w.Write([]byte(err.Error()))
		return
	}

	filters := &models.Filters{}
	if projectQuery := r.URL.Query().Get("project"); projectQuery != "" {
		filters.Project = projectQuery
	}

	vm := v1.NewSummariesFrom(summaries, filters)
	utils.RespondJSON(w, r, http.StatusOK, vm)
}

func (h *SummariesHandler) loadUserSummaries(r *http.Request) ([]*models.Summary, error, int) {
	user := middlewares.GetPrincipal(r)
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
		if err, parsedFrom, parsedTo := utils.ResolveIntervalRawTZ(rangeParam, timezone); err == nil {
			start, end = parsedFrom, parsedTo
		} else {
			return nil, errors.New("invalid 'range' parameter"), http.StatusBadRequest
		}
	} else if err, parsedFrom, parsedTo := utils.ResolveIntervalRawTZ(startParam, timezone); err == nil && startParam == endParam {
		// also accept start param to be a range param
		start, end = parsedFrom, parsedTo
	} else {
		// eventually, consider start and end params a date
		var err error

		start, err = utils.ParseDateTimeTZ(strings.Replace(startParam, " ", "+", 1), timezone)
		if err != nil {
			return nil, errors.New("missing required 'start' parameter"), http.StatusBadRequest
		}

		end, err = utils.ParseDateTimeTZ(strings.Replace(endParam, " ", "+", 1), timezone)
		if err != nil {
			return nil, errors.New("missing required 'end' parameter"), http.StatusBadRequest
		}
	}

	// wakatime interprets end date as "inclusive", wakapi usually as "exclusive"
	// i.e. for wakatime, an interval 2021-04-29 - 2021-04-29 is actually 2021-04-29 - 2021-04-30,
	// while for wakapi it would be empty
	// see https://github.com/muety/wakapi/issues/192
	end = utils.EndOfDay(end).Add(-1 * time.Second)

	overallParams := &models.SummaryParams{
		From:      start,
		To:        end,
		User:      user,
		Recompute: false,
	}

	intervals := utils.SplitRangeByDays(overallParams.From, overallParams.To)
	summaries := make([]*models.Summary, len(intervals))

	for i, interval := range intervals {
		summary, err := h.summarySrvc.Aliased(interval[0], interval[1], user, h.summarySrvc.Retrieve, false)
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
