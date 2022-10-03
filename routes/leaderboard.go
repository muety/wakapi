package routes

import (
	"fmt"
	"github.com/emvi/logbuch"
	"github.com/gorilla/mux"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/models/view"
	"github.com/muety/wakapi/services"
	"net/http"
	"strings"
)

type LeaderboardHandler struct {
	config             *conf.Config
	userService        services.IUserService
	leaderboardService services.ILeaderboardService
}

var allowedAggregations = map[string]uint8{
	"language": models.SummaryLanguage,
}

func NewLeaderboardHandler(userService services.IUserService, leaderboardService services.ILeaderboardService) *LeaderboardHandler {
	return &LeaderboardHandler{
		config:             conf.Get(),
		userService:        userService,
		leaderboardService: leaderboardService,
	}
}

func (h *LeaderboardHandler) RegisterRoutes(router *mux.Router) {
	r := router.PathPrefix("/leaderboard").Subrouter()
	r.Use(
		middlewares.NewAuthenticateMiddleware(h.userService).
			WithRedirectTarget(defaultErrorRedirectTarget()).
			WithOptionalFor([]string{"/"}).
			Handler,
	)
	r.Methods(http.MethodGet).HandlerFunc(h.GetIndex)
}

func (h *LeaderboardHandler) GetIndex(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}
	if err := templates[conf.LeaderboardTemplate].Execute(w, h.buildViewModel(r)); err != nil {
		logbuch.Error(err.Error())
	}
}

func (h *LeaderboardHandler) buildViewModel(r *http.Request) *view.LeaderboardViewModel {
	user := middlewares.GetPrincipal(r)
	byParam := strings.ToLower(r.URL.Query().Get("by"))
	keyParam := strings.ToLower(r.URL.Query().Get("key"))

	var err error
	var leaderboard models.Leaderboard
	var topKeys []string

	if byParam == "" {
		leaderboard, err = h.leaderboardService.GetByInterval(models.IntervalPast7Days, true)
		if err != nil {
			conf.Log().Request(r).Error("error while fetching general leaderboard items - %v", err)
			return &view.LeaderboardViewModel{Error: criticalError}
		}
	} else {
		if by, ok := allowedAggregations[byParam]; ok {
			leaderboard, err = h.leaderboardService.GetAggregatedByInterval(models.IntervalPast7Days, &by, true)
			if err != nil {
				conf.Log().Request(r).Error("error while fetching general leaderboard items - %v", err)
				return &view.LeaderboardViewModel{Error: criticalError}
			}

			topKeys = leaderboard.TopKeys(by)
			if len(topKeys) > 0 {
				if keyParam == "" {
					keyParam = strings.ToLower(topKeys[0])
				}
				leaderboard = leaderboard.TopByKey(by, keyParam)
			}
		} else {
			return &view.LeaderboardViewModel{Error: fmt.Sprintf("unsupported aggregation '%s'", byParam)}
		}
	}

	var apiKey string
	if user != nil {
		apiKey = user.ApiKey
	}

	return &view.LeaderboardViewModel{
		User:    user,
		By:      byParam,
		Key:     keyParam,
		Items:   leaderboard,
		TopKeys: topKeys,
		ApiKey:  apiKey,
		Success: r.URL.Query().Get("success"),
		Error:   r.URL.Query().Get("error"),
	}
}
