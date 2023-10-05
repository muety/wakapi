package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"net/http"
	"time"
)

type ActivityApiHandler struct {
	config          *conf.Config
	userService     services.IUserService
	activityService services.IActivityService
}

func NewActivityApiHandler(userService services.IUserService, activityService services.IActivityService) *ActivityApiHandler {
	return &ActivityApiHandler{
		activityService: activityService,
		userService:     userService,
		config:          conf.Get(),
	}
}

func (h *ActivityApiHandler) RegisterRoutes(router chi.Router) {
	r := chi.NewRouter()
	r.Use(
		middlewares.NewAuthenticateMiddleware(h.userService).WithOptionalFor([]string{"/api/activity/chart/"}).Handler,
		middleware.Compress(9, "image/svg+xml"),
	)
	r.Get("/chart/{user}.svg", h.GetActivityChart)

	router.Mount("/activity", r)
}

func (h *ActivityApiHandler) GetActivityChart(w http.ResponseWriter, r *http.Request) {
	authorizedUser := middlewares.GetPrincipal(r)
	requestedUser, err := h.userService.GetUserById(chi.URLParam(r, "user"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if authorizedUser == nil || authorizedUser.ID != requestedUser.ID {
		if _, userRange := helpers.ResolveMaximumRange(requestedUser.ShareDataMaxDays); userRange != models.IntervalPast12Months && userRange != models.IntervalAny { // TODO: build "hierarchy" of intervals to easily check if one is contained in another
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}

	paramDark := r.URL.Query().Has("dark") && r.URL.Query().Get("dark") != "false"
	paramNoAttr := r.URL.Query().Has("noattr") && r.URL.Query().Get("noattr") != "false" // no attribution (no wakapi logo in bottom left corner)

	chart, err := h.activityService.GetChart(requestedUser, models.IntervalPast12Months, paramDark, paramNoAttr, utils.IsNoCache(r, 6*time.Hour))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		conf.Log().Request(r).Error("failed to get activity chart for user %s - %v", err)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "max-age=21600") // 6 hours
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(chart))
}
