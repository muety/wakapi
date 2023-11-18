package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"net/http"

	conf "github.com/muety/wakapi/config"
	v1 "github.com/muety/wakapi/models/compat/wakatime/v1"
	"github.com/muety/wakapi/services"
)

type LeadersHandler struct {
	config          *conf.Config
	userSrvc        services.IUserService
	leaderboardSrvc services.ILeaderboardService
}

func NewLeadersHandler(userService services.IUserService, leaderboardService services.ILeaderboardService) *LeadersHandler {
	return &LeadersHandler{
		userSrvc:        userService,
		leaderboardSrvc: leaderboardService,
		config:          conf.Get(),
	}
}

func (h *LeadersHandler) RegisterRoutes(router chi.Router) {
	router.Group(func(r chi.Router) {
		r.Use(middlewares.NewAuthenticateMiddleware(h.userSrvc).Handler)
		r.Get("/compat/wakatime/v1/leaders", h.Get)
	})
}

// @Summary List of users ranked by coding activity in descending order.
// @Description Mimics https://wakatime.com/developers#leaders
// @ID get-wakatime-leaders
// @Tags wakatime
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} v1.LeaderboardViewModel
// @Router /compat/wakatime/v1/leaders [get]
func (h *LeadersHandler) Get(w http.ResponseWriter, r *http.Request) {
	wakapiUser := middlewares.GetPrincipal(r)
	if wakapiUser == nil {
		helpers.RespondJSON(w, r, http.StatusUnauthorized, nil)
		return
	}
	pageParams := utils.ParsePageParamsWithDefault(r, 1, 100)

	/*
		Not implemented query params:
			language - String - optional - Filter leaders by a specific language.
			is_hireable - Boolean - optional - Filter leaders by the hireable badge.
			country_code - String - optional - Filter leaders by two-character country code.
	*/

	by := models.SummaryLanguage
	leaderboard, err := h.leaderboardSrvc.GetAggregatedByInterval(models.IntervalPast7Days, &by, pageParams, true)
	if err != nil {
		conf.Log().Request(r).Error("error while fetching general leaderboard items - %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("something went wrong"))
		return
	}

	// regardless of page, always show own rank
	if wakapiUser != nil && !leaderboard.HasUser(wakapiUser.ID) {
		// but only if leaderboard spans multiple pages
		if count, err := h.leaderboardSrvc.CountUsers(); err == nil && count > int64(pageParams.PageSize) {
			if l, err := h.leaderboardSrvc.GetByIntervalAndUser(models.IntervalPast7Days, wakapiUser.ID, true); err == nil && len(l) > 0 {
				leaderboard = append(leaderboard, l[0])
			}
		}
	}

	user := v1.NewLeaderboardViewModel(&leaderboard, wakapiUser)
	helpers.RespondJSON(w, r, http.StatusOK, user)
}
