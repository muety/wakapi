package v1

import (
	"errors"
	"github.com/gorilla/mux"
	"github.com/muety/wakapi/models"
	v1 "github.com/muety/wakapi/models/compat/v1"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"net/http"
	"strings"
	"time"
)

type CompatV1SummariesHandler struct {
	summarySrvc *services.SummaryService
	config      *models.Config
}

func NewCompatV1SummariesHandler(summaryService *services.SummaryService) *CompatV1SummariesHandler {
	return &CompatV1SummariesHandler{
		summarySrvc: summaryService,
		config:      models.GetConfig(),
	}
}

/*
TODO: support parameters: branches, timeout, writes_only, timezone
https://wakatime.com/developers#summaries
timezone can be specified via an offset suffix (e.g. +02:00) in date strings
*/

func (h *CompatV1SummariesHandler) ApiGet(w http.ResponseWriter, r *http.Request) {
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

	vm := v1.NewSummariesFrom(summaries, &v1.Filters{})
	utils.RespondJSON(w, http.StatusOK, vm)
}

func (h *CompatV1SummariesHandler) loadUserSummaries(r *http.Request) ([]*models.Summary, error, int) {
	user := r.Context().Value(models.UserKey).(*models.User)
	params := r.URL.Query()

	var start, end time.Time
	// TODO: find out what other special dates are supported by wakatime (e.g. tomorrow, yesterday, ...?)
	if startKey, endKey := params.Get("start"), params.Get("end"); startKey == "today" && startKey == endKey {
		start = utils.StartOfToday()
		end = time.Now()
	} else {
		var err error

		start, err = time.Parse(time.RFC3339, strings.Replace(startKey, " ", "+", 1))
		if err != nil {
			return nil, errors.New("missing required 'start' parameter"), http.StatusBadRequest
		}

		end, err = time.Parse(time.RFC3339, strings.Replace(endKey, " ", "+", 1))
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
		summary, err := h.summarySrvc.Construct(interval[0], interval[1], user, false) // 'to' is always constant
		if err != nil {
			return nil, err, http.StatusInternalServerError
		}
		summaries[i] = summary
	}

	return summaries, nil, http.StatusOK
}
