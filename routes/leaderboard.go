package routes

import (
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/emvi/logbuch"
	"github.com/go-chi/chi/v5"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/models/view"
	routeutils "github.com/muety/wakapi/routes/utils"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
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

func (h *LeaderboardHandler) RegisterRoutes(router chi.Router) {
	r := chi.NewRouter()
	r.Use(
		middlewares.NewAuthenticateMiddleware(h.userService).
			WithRedirectTarget(defaultErrorRedirectTarget()).
			WithRedirectErrorMessage("unauthorized").
			WithOptionalFor([]string{"/"}).Handler,
	)
	r.Get("/", h.GetIndex)

	router.Mount("/leaderboard", r)
}

func (h *LeaderboardHandler) GetIndex(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}
	if err := templates[conf.LeaderboardTemplate].Execute(w, h.buildViewModel(r, w)); err != nil {
		logbuch.Error(err.Error())
	}
}

func (h *LeaderboardHandler) buildViewModel(r *http.Request, w http.ResponseWriter) *view.LeaderboardViewModel {
	user := middlewares.GetPrincipal(r)
	byParam := strings.ToLower(r.URL.Query().Get("by"))
	keyParam := strings.ToLower(r.URL.Query().Get("key"))
	pageParams := utils.ParsePageParamsWithDefault(r, 1, 100)
	// note: pagination is not fully implemented, yet
	// count function to get total item / total pages is missing
	// and according ui (+ optionally search bar) is missing, too

	var err error
	var leaderboard models.Leaderboard
	var userLanguages map[string][]string
	var topKeys []string

	if byParam == "" {
		leaderboard, err = h.leaderboardService.GetByInterval(models.IntervalPast7Days, pageParams, true)
		if err != nil {
			conf.Log().Request(r).Error("error while fetching general leaderboard items - %v", err)
			return &view.LeaderboardViewModel{
				Messages: view.Messages{Error: criticalError},
			}
		}

		// regardless of page, always show own rank
		if user != nil && !leaderboard.HasUser(user.ID) {
			// but only if leaderboard spans multiple pages
			if count, err := h.leaderboardService.CountUsers(); err == nil && count > int64(pageParams.PageSize) {
				if l, err := h.leaderboardService.GetByIntervalAndUser(models.IntervalPast7Days, user.ID, true); err == nil && len(l) > 0 {
					leaderboard = append(leaderboard, l[0])
				}
			}
		}
	} else {
		if by, ok := allowedAggregations[byParam]; ok {
			leaderboard, err = h.leaderboardService.GetAggregatedByInterval(models.IntervalPast7Days, &by, pageParams, true)
			if err != nil {
				conf.Log().Request(r).Error("error while fetching general leaderboard items - %v", err)
				return &view.LeaderboardViewModel{
					Messages: view.Messages{Error: criticalError},
				}
			}

			// regardless of page, always show own rank
			if user != nil {
				// but only if leaderboard could, in theory, span multiple pages
				if count, err := h.leaderboardService.CountUsers(); err == nil && count > int64(pageParams.PageSize) {
					if l, err := h.leaderboardService.GetAggregatedByIntervalAndUser(models.IntervalPast7Days, user.ID, &by, true); err == nil {
						leaderboard.AddMany(l)
					} else {
						conf.Log().Request(r).Error("error while fetching own aggregated user leaderboard - %v", err)
					}
				}
			}

			userLeaderboards := slice.GroupWith[*models.LeaderboardItemRanked, string](leaderboard, func(item *models.LeaderboardItemRanked) string {
				return item.UserID
			})
			userLanguages = map[string][]string{}
			for u, items := range userLeaderboards {
				userLanguages[u] = models.Leaderboard(items).TopKeysByUser(models.SummaryLanguage, u)
			}

			topKeys = leaderboard.TopKeys(by)
			if len(topKeys) > 0 {
				if keyParam == "" {
					keyParam = topKeys[0]
				}
				keyParam = strings.ToLower(keyParam)
				leaderboard = leaderboard.TopByKey(by, keyParam)
			}
		} else {
			return &view.LeaderboardViewModel{
				Messages: view.Messages{Error: fmt.Sprintf("unsupported aggregation '%s'", byParam)},
			}
		}
	}

	var apiKey string
	if user != nil {
		apiKey = user.ApiKey
	}

	vm := &view.LeaderboardViewModel{
		User:          user,
		By:            byParam,
		Key:           keyParam,
		Items:         leaderboard,
		UserLanguages: userLanguages,
		TopKeys:       topKeys,
		ApiKey:        apiKey,
		PageParams:    pageParams,
	}
	return routeutils.WithSessionMessages(vm, r, w)
}
