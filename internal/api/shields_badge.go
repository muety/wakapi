package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/muety/wakapi/helpers"
	routeutils "github.com/muety/wakapi/routes/utils"

	"github.com/muety/wakapi/models"
	v1 "github.com/muety/wakapi/models/compat/shields/v1"
)

// @Summary Get badge data
// @Description Retrieve total time for a given entity (e.g. a project) within a given range (e.g. one week) in a format compatible with [Shields.io](https://shields.io/endpoint). Requires public data access to be allowed.
// @ID get-badge
// @Tags badges
// @Produce json
// @Param user path string true "User ID to fetch data for"
// @Param interval path string true "Interval to aggregate data for" Enums(today, yesterday, week, month, year, 7_days, last_7_days, 30_days, last_30_days, 6_months, last_6_months, 12_months, last_12_months, last_year, any, all_time)
// @Param filter path string true "Filter to apply (e.g. 'project:wakapi' or 'language:Go')"
// @Success 200 {object} v1.BadgeData
// @Router /compat/shields/v1/{user}/{interval}/{filter} [get]
func (a *APIv1) GetShield(w http.ResponseWriter, r *http.Request) {
	user, err := a.services.Users().GetUserById(chi.URLParam(r, "user"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	interval, filters, err := routeutils.GetBadgeParams(r.URL.Path, nil, user)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(err.Error()))
		return
	}
	filters.WithSelectFilteredOnly()

	cacheKey := fmt.Sprintf("%s_%v_%s", user.ID, *interval.Key, filters.Hash())
	if cacheResult, ok := a.cache.Get(cacheKey); ok {
		helpers.RespondJSON(w, r, http.StatusOK, cacheResult.(*v1.BadgeData))
		return
	}

	params := &models.SummaryParams{
		From:    interval.Start,
		To:      interval.End,
		User:    user,
		Filters: filters,
	}

	summary, err, status := routeutils.LoadUserSummaryByParams(a.services.Summary(), params)
	if err != nil {
		w.WriteHeader(status)
		w.Write([]byte(err.Error()))
		return
	}

	vm := v1.NewBadgeDataFrom(summary)
	a.cache.SetDefault(cacheKey, vm)
	helpers.RespondJSON(w, r, http.StatusOK, vm)
}
