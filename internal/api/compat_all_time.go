package api

import (
	"net/http"
	"time"

	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/internal/utilities"
	"github.com/muety/wakapi/models"
	v1 "github.com/muety/wakapi/models/compat/wakatime/v1"
	"github.com/muety/wakapi/models/types"
)

// type AllTimeHandler struct {
// 	config      *conf.Config
// 	userSrvc    services.IUserService
// 	summarySrvc services.ISummaryService
// }

// func NewAllTimeHandler(services services.IServices) *AllTimeHandler {
// 	return &AllTimeHandler{
// 		userSrvc:    services.Users(),
// 		summarySrvc: services.Summary(),
// 		config:      conf.Get(),
// 	}
// }

// func (h *AllTimeHandler) RegisterRoutes(router chi.Router) {
// 	router.Group(func(r chi.Router) {
// 		r.Use(middlewares.NewAuthenticateMiddleware(h.userSrvc).Handler)
// 		r.Get("/compat/wakatime/v1/users/{user}/all_time_since_today", h.Get)
// 	})
// }

// @Summary Retrieve summary for all time
// @Description Mimics https://wakatime.com/developers#all_time_since_today
// @ID get-all-time
// @Tags wakatime
// @Produce json
// @Param user path string true "User ID to fetch data for (or 'current')"
// @Security ApiKeyAuth
// @Success 200 {object} v1.AllTimeViewModel
// @Router /compat/wakatime/v1/users/{user}/all_time_since_today [get]
func (a *APIv1) GetAllTime(w http.ResponseWriter, r *http.Request) {
	user, err := utilities.CheckEffectiveUser(w, r, a.services.Users(), "current")
	if err != nil {
		return // response was already sent by util function
	}

	summaryParams := &models.SummaryParams{
		From:      time.Time{},
		To:        time.Now(),
		User:      user,
		Recompute: false,
	}

	summary, err, status := a.loadUserSummary(summaryParams, helpers.ParseSummaryFilters(r).WithSelectFilteredOnly())
	if err != nil {
		w.WriteHeader(status)
		w.Write([]byte(err.Error()))
		return
	}

	vm := v1.NewAllTimeFrom(summary)
	helpers.RespondJSON(w, r, http.StatusOK, vm)
}

func (a *APIv1) loadUserSummary(summaryParams *models.SummaryParams, filters *models.Filters) (*models.Summary, error, int) {
	var retrieveSummary types.SummaryRetriever = a.services.Summary().Retrieve
	if summaryParams.Recompute {
		retrieveSummary = a.services.Summary().Summarize
	}

	summary, err := a.services.Summary().Aliased(
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
