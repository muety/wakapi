package v1

import (
	"errors"
	"github.com/gorilla/mux"
	conf "github.com/muety/wakapi/config"
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
	summarySrvc services.ISummaryService
}

func NewSummariesHandler(summaryService services.ISummaryService) *SummariesHandler {
	return &SummariesHandler{
		summarySrvc: summaryService,
		config:      conf.Get(),
	}
}

func (h *SummariesHandler) RegisterRoutes(router *mux.Router) {
	router.Path("/wakatime/v1/users/{user}/summaries").Methods(http.MethodGet).HandlerFunc(h.Get)
}

// TODO: Support parameters: project, branches, timeout, writes_only, timezone
// See https://wakatime.com/developers#summaries.
// Timezone can be specified via an offset suffix (e.g. +02:00) in date strings.
// Requires https://github.com/muety/wakapi/issues/108.

func (h *SummariesHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestedUser := vars["user"]
	authorizedUser := r.Context().Value(models.UserKey).(*models.User)

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
	utils.RespondJSON(w, http.StatusOK, vm)
}

func (h *SummariesHandler) loadUserSummaries(r *http.Request) ([]*models.Summary, error, int) {
	user := r.Context().Value(models.UserKey).(*models.User)
	params := r.URL.Query()
	rangeParam, startParam, endParam := params.Get("range"), params.Get("start"), params.Get("end")

	var start, end time.Time
	if rangeParam != "" {
		// range param takes precedence
		if err, parsedFrom, parsedTo := utils.ResolveInterval(rangeParam); err == nil {
			start, end = parsedFrom, parsedTo
		} else {
			return nil, errors.New("invalid 'range' parameter"), http.StatusBadRequest
		}
	} else if err, parsedFrom, parsedTo := utils.ResolveInterval(startParam); err == nil && startParam == endParam {
		// also accept start param to be a range param
		start, end = parsedFrom, parsedTo
	} else {
		// eventually, consider start and end params a date
		var err error

		start, err = time.Parse(time.RFC3339, strings.Replace(startParam, " ", "+", 1))
		if err != nil {
			return nil, errors.New("missing required 'start' parameter"), http.StatusBadRequest
		}

		end, err = time.Parse(time.RFC3339, strings.Replace(endParam, " ", "+", 1))
		if err != nil {
			return nil, errors.New("missing required 'end' parameter"), http.StatusBadRequest
		}
	}

	overallParams := &models.SummaryParams{
		From:      start,
		To:        end,
		User:      user,
		Recompute: false,
	}

	intervals := utils.SplitRangeByDays(overallParams.From, overallParams.To)
	summaries := make([]*models.Summary, len(intervals))

	for i, interval := range intervals {
		summary, err := h.summarySrvc.Aliased(interval[0], interval[1], user, h.summarySrvc.Retrieve)
		if err != nil {
			return nil, err, http.StatusInternalServerError
		}
		summaries[i] = summary
	}

	return summaries, nil, http.StatusOK
}
