package v1

import (
	"github.com/gorilla/mux"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	v1 "github.com/muety/wakapi/models/compat/shields/v1"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	intervalPattern     = `interval:([a-z0-9_]+)`
	entityFilterPattern = `(project|os|editor|language|machine):([_a-zA-Z0-9-]+)`
)

type BadgeHandler struct {
	config      *conf.Config
	userSrvc    services.IUserService
	summarySrvc services.ISummaryService
}

func NewBadgeHandler(summaryService services.ISummaryService, userService services.IUserService) *BadgeHandler {
	return &BadgeHandler{
		summarySrvc: summaryService,
		userSrvc:    userService,
		config:      conf.Get(),
	}
}

func (h *BadgeHandler) RegisterRoutes(router *mux.Router) {
	// no auth middleware here, handler itself resolves the user
	r := router.PathPrefix("/shields/v1/{user}").Subrouter()
	r.Methods(http.MethodGet).HandlerFunc(h.Get)
}

func (h *BadgeHandler) Get(w http.ResponseWriter, r *http.Request) {
	intervalReg := regexp.MustCompile(intervalPattern)
	entityFilterReg := regexp.MustCompile(entityFilterPattern)

	if userAgent := r.Header.Get("user-agent"); !strings.HasPrefix(userAgent, "Shields.io/") && !h.config.IsDev() {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var filterEntity, filterKey string
	if groups := entityFilterReg.FindStringSubmatch(r.URL.Path); len(groups) > 2 {
		filterEntity, filterKey = groups[1], groups[2]
	}

	var interval = models.IntervalPast30Days
	if groups := intervalReg.FindStringSubmatch(r.URL.Path); len(groups) > 1 {
		if i, err := utils.ParseInterval(groups[1]); err == nil {
			interval = i
		}
	}

	requestedUserId := mux.Vars(r)["user"]
	user, err := h.userSrvc.GetUserById(requestedUserId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_, rangeFrom, rangeTo := utils.ResolveInterval(interval)
	minStart := utils.StartOfDay(rangeTo.Add(-24 * time.Hour * time.Duration(user.ShareDataMaxDays)))
	if rangeFrom.Before(minStart) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("requested time range too broad"))
		return
	}

	var filters *models.Filters
	switch filterEntity {
	case "project":
		filters = models.NewFiltersWith(models.SummaryProject, filterKey)
	case "os":
		filters = models.NewFiltersWith(models.SummaryOS, filterKey)
	case "editor":
		filters = models.NewFiltersWith(models.SummaryEditor, filterKey)
	case "language":
		filters = models.NewFiltersWith(models.SummaryLanguage, filterKey)
	case "machine":
		filters = models.NewFiltersWith(models.SummaryMachine, filterKey)
	default:
		filters = &models.Filters{}
	}

	summary, err, status := h.loadUserSummary(user, interval)
	if err != nil {
		w.WriteHeader(status)
		w.Write([]byte(err.Error()))
		return
	}

	vm := v1.NewBadgeDataFrom(summary, filters)
	utils.RespondJSON(w, http.StatusOK, vm)
}

func (h *BadgeHandler) loadUserSummary(user *models.User, interval *models.IntervalKey) (*models.Summary, error, int) {
	err, from, to := utils.ResolveInterval(interval)
	if err != nil {
		return nil, err, http.StatusBadRequest
	}

	summaryParams := &models.SummaryParams{
		From: from,
		To:   to,
		User: user,
	}

	var retrieveSummary services.SummaryRetriever = h.summarySrvc.Retrieve
	if summaryParams.Recompute {
		retrieveSummary = h.summarySrvc.Summarize
	}

	summary, err := h.summarySrvc.Aliased(summaryParams.From, summaryParams.To, summaryParams.User, retrieveSummary)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}

	return summary, nil, http.StatusOK
}
