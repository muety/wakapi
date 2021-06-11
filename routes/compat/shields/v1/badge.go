package v1

import (
	"fmt"
	"github.com/gorilla/mux"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	v1 "github.com/muety/wakapi/models/compat/shields/v1"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"github.com/patrickmn/go-cache"
	"net/http"
	"regexp"
	"time"
)

const (
	intervalPattern     = `interval:([a-z0-9_]+)`
	entityFilterPattern = `(project|os|editor|language|machine):([_a-zA-Z0-9-\s]+)`
)

type BadgeHandler struct {
	config      *conf.Config
	userSrvc    services.IUserService
	summarySrvc services.ISummaryService
	cache       *cache.Cache
}

func NewBadgeHandler(summaryService services.ISummaryService, userService services.IUserService) *BadgeHandler {
	return &BadgeHandler{
		summarySrvc: summaryService,
		userSrvc:    userService,
		cache:       cache.New(time.Hour, time.Hour),
		config:      conf.Get(),
	}
}

func (h *BadgeHandler) RegisterRoutes(router *mux.Router) {
	// no auth middleware here, handler itself resolves the user
	r := router.PathPrefix("/compat/shields/v1/{user}").Subrouter()
	r.Methods(http.MethodGet).HandlerFunc(h.Get)
}

// @Summary Get badge data
// @Description Retrieve total time for a given entity (e.g. a project) within a given range (e.g. one week) in a format compatible with [Shields.io](https://shields.io/endpoint). Requires public data access to be allowed.
// @ID get-badge
// @Tags badges
// @Produce json
// @Param user path string true "User ID to fetch data for"
// @Param interval path string true "Interval to aggregate data for" Enums(today, yesterday, week, month, year, 7_days, last_7_days, 30_days, last_30_days, 12_months, last_12_months, any)
// @Param filter path string true "Filter to apply (e.g. 'project:wakapi' or 'language:Go')"
// @Success 200 {object} v1.BadgeData
// @Router /compat/shields/v1/{user}/{interval}/{filter} [get]
func (h *BadgeHandler) Get(w http.ResponseWriter, r *http.Request) {
	intervalReg := regexp.MustCompile(intervalPattern)
	entityFilterReg := regexp.MustCompile(entityFilterPattern)

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

	_, rangeFrom, rangeTo := utils.ResolveIntervalTZ(interval, user.TZ())
	minStart := rangeTo.Add(-24 * time.Hour * time.Duration(user.ShareDataMaxDays))
	// negative value means no limit
	if rangeFrom.Before(minStart) && user.ShareDataMaxDays >= 0 {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("requested time range too broad"))
		return
	}

	var permitEntity bool
	var filters *models.Filters
	switch filterEntity {
	case "project":
		permitEntity = user.ShareProjects
		filters = models.NewFiltersWith(models.SummaryProject, filterKey)
	case "os":
		permitEntity = user.ShareOSs
		filters = models.NewFiltersWith(models.SummaryOS, filterKey)
	case "editor":
		permitEntity = user.ShareEditors
		filters = models.NewFiltersWith(models.SummaryEditor, filterKey)
	case "language":
		permitEntity = user.ShareLanguages
		filters = models.NewFiltersWith(models.SummaryLanguage, filterKey)
	case "machine":
		permitEntity = user.ShareMachines
		filters = models.NewFiltersWith(models.SummaryMachine, filterKey)
	case "label":
		permitEntity = user.ShareLabels
		filters = models.NewFiltersWith(models.SummaryLabel, filterKey)
	default:
		permitEntity = true
		filters = &models.Filters{}
	}

	if !permitEntity {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("user did not opt in to share entity-specific data"))
		return
	}

	cacheKey := fmt.Sprintf("%s_%v_%s_%s", user.ID, *interval, filterEntity, filterKey)
	if cacheResult, ok := h.cache.Get(cacheKey); ok {
		utils.RespondJSON(w, r, http.StatusOK, cacheResult.(*v1.BadgeData))
		return
	}

	summary, err, status := h.loadUserSummary(user, interval)
	if err != nil {
		w.WriteHeader(status)
		w.Write([]byte(err.Error()))
		return
	}

	vm := v1.NewBadgeDataFrom(summary, filters)
	h.cache.SetDefault(cacheKey, vm)
	utils.RespondJSON(w, r, http.StatusOK, vm)
}

func (h *BadgeHandler) loadUserSummary(user *models.User, interval *models.IntervalKey) (*models.Summary, error, int) {
	err, from, to := utils.ResolveIntervalTZ(interval, user.TZ())
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

	summary, err := h.summarySrvc.Aliased(summaryParams.From, summaryParams.To, summaryParams.User, retrieveSummary, summaryParams.Recompute)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}

	return summary, nil, http.StatusOK
}
