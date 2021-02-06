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
	"time"
)

type StatsHandler struct {
	config      *conf.Config
	summarySrvc services.ISummaryService
}

func NewStatsHandler(summaryService services.ISummaryService) *StatsHandler {
	return &StatsHandler{
		summarySrvc: summaryService,
		config:      conf.Get(),
	}
}

func (h *StatsHandler) RegisterRoutes(router *mux.Router) {
	router.Path("/wakatime/v1/users/{user}/stats/{range}").Methods(http.MethodGet).HandlerFunc(h.Get)
}

// TODO: support filtering (requires https://github.com/muety/wakapi/issues/108)

func (h *StatsHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestedUser := vars["user"]
	requestedRange := vars["range"]

	user := r.Context().Value(models.UserKey).(*models.User)

	if requestedUser != user.ID && requestedUser != "current" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	summary, err, status := h.loadUserSummary(user, requestedRange)
	if err != nil {
		w.WriteHeader(status)
		w.Write([]byte(err.Error()))
		return
	}

	filters := &models.Filters{}
	if projectQuery := r.URL.Query().Get("project"); projectQuery != "" {
		filters.Project = projectQuery
	}

	vm := v1.NewStatsFrom(summary, filters)
	utils.RespondJSON(w, http.StatusOK, vm)
}

func (h *StatsHandler) loadUserSummary(user *models.User, rangeKey string) (*models.Summary, error, int) {
	var start, end time.Time

	if err, parsedFrom, parsedTo := utils.ResolveIntervalRaw(rangeKey); err == nil {
		start, end = parsedFrom, parsedTo
	} else {
		return nil, errors.New("invalid 'range' parameter"), http.StatusBadRequest
	}

	overallParams := &models.SummaryParams{
		From:      start,
		To:        end,
		User:      user,
		Recompute: false,
	}

	summary, err := h.summarySrvc.Aliased(overallParams.From, overallParams.To, user, h.summarySrvc.Retrieve)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}

	return summary, nil, http.StatusOK
}
