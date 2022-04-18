package v1

import (
	"fmt"
	routeutils "github.com/muety/wakapi/routes/utils"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	v1 "github.com/muety/wakapi/models/compat/shields/v1"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"github.com/patrickmn/go-cache"
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
	user, err := h.userSrvc.GetUserById(mux.Vars(r)["user"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	interval, filters, err := routeutils.GetBadgeParams(r, user)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(err.Error()))
		return
	}

	cacheKey := fmt.Sprintf("%s_%v_%s", user.ID, *interval.Key, filters.Hash())
	if cacheResult, ok := h.cache.Get(cacheKey); ok {
		utils.RespondJSON(w, r, http.StatusOK, cacheResult.(*v1.BadgeData))
		return
	}

	params := &models.SummaryParams{
		From:    interval.Start,
		To:      interval.End,
		User:    user,
		Filters: filters,
	}

	summary, err, status := routeutils.LoadUserSummaryByParams(h.summarySrvc, params)
	if err != nil {
		w.WriteHeader(status)
		w.Write([]byte(err.Error()))
		return
	}

	vm := v1.NewBadgeDataFrom(summary)
	h.cache.SetDefault(cacheKey, vm)
	utils.RespondJSON(w, r, http.StatusOK, vm)
}

func (h *BadgeHandler) loadUserSummary(user *models.User, interval *models.IntervalKey, filters *models.Filters) (*models.Summary, error, int) {
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

	summary, err := h.summarySrvc.Aliased(
		summaryParams.From,
		summaryParams.To,
		summaryParams.User,
		retrieveSummary,
		filters,
		summaryParams.Recompute,
	)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}

	return summary, nil, http.StatusOK
}
