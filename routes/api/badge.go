package api

import (
	"fmt"
	"github.com/duke-git/lancet/v2/maputil"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/go-chi/chi/v5"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	v1 "github.com/muety/wakapi/models/compat/shields/v1"
	routeutils "github.com/muety/wakapi/routes/utils"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"github.com/narqo/go-badge"
	"github.com/patrickmn/go-cache"
	"net/http"
	"time"
)

type BadgeHandler struct {
	config      *conf.Config
	cache       *cache.Cache
	userSrvc    services.IUserService
	summarySrvc services.ISummaryService
}

func NewBadgeHandler(userService services.IUserService, summaryService services.ISummaryService) *BadgeHandler {
	return &BadgeHandler{
		config:      conf.Get(),
		cache:       cache.New(time.Hour, time.Hour),
		userSrvc:    userService,
		summarySrvc: summaryService,
	}
}

func (h *BadgeHandler) RegisterRoutes(router chi.Router) {
	r := chi.NewRouter()
	r.Use(middlewares.NewAuthenticateMiddleware(h.userSrvc).WithOptionalFor("/api/badge/").Handler)
	r.Get("/{user}/*", h.Get)
	router.Mount("/badge", r)
}

func (h *BadgeHandler) Get(w http.ResponseWriter, r *http.Request) {
	authorizedUser := middlewares.GetPrincipal(r)
	user, err := h.userSrvc.GetUserById(chi.URLParam(r, "user"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	interval, filters, err := routeutils.GetBadgeParams(r.URL.Path, authorizedUser, user)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(err.Error()))
		return
	}

	cacheKey := fmt.Sprintf("%s_%v_%s_%s", user.ID, *interval.Key, filters.Hash(), r.URL.RawQuery)
	noCache := utils.IsNoCache(r, 1*time.Hour)
	if cacheResult, ok := h.cache.Get(cacheKey); ok && !noCache {
		respondSvg(w, cacheResult.([]byte))
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

	badgeData := v1.NewBadgeDataFrom(summary)
	if customLabel := r.URL.Query().Get("label"); customLabel != "" {
		badgeData.Label = customLabel
	}
	if customColor := r.URL.Query().Get("color"); customColor != "" {
		badgeData.Color = customColor
	}

	if badgeData.Color[0:1] != "#" && !slice.Contain(maputil.Keys(badge.ColorScheme), badgeData.Color) {
		badgeData.Color = "#" + badgeData.Color
	}

	badgeSvg, err := badge.RenderBytes(badgeData.Label, badgeData.Message, badge.Color(badgeData.Color))
	h.cache.SetDefault(cacheKey, badgeSvg)
	respondSvg(w, badgeSvg)
}

func respondSvg(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "max-age=3600")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
