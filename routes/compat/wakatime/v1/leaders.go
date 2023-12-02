package v1

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/go-chi/chi/v5"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"math"
	"net/http"
	"strings"
	"time"

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
		r.Use(middlewares.NewAuthenticateMiddleware(h.userSrvc).WithOptionalFor("/").Handler)
		r.Get("/compat/wakatime/v1/leaders", h.Get)
	})
}

// @Summary List of users ranked by coding activity in descending order.
// @Description Mimics https://wakatime.com/developers#leaders
// @ID get-wakatime-leaders
// @Tags wakatime
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} v1.LeadersViewModel
// @Router /compat/wakatime/v1/leaders [get]
func (h *LeadersHandler) Get(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	languageParam := strings.ToLower(r.URL.Query().Get("language"))
	pageParams := utils.ParsePageParamsWithDefault(r, 1, 100)
	by := models.SummaryLanguage

	loadPrimaryLeaderboard := func() (models.Leaderboard, error) {
		if languageParam == "" {
			return h.leaderboardSrvc.GetByInterval(h.leaderboardSrvc.GetDefaultScope(), pageParams, true)
		} else {
			l, err := h.leaderboardSrvc.GetAggregatedByInterval(h.leaderboardSrvc.GetDefaultScope(), &by, pageParams, true)
			if err == nil {
				return l.TopByKey(by, languageParam), err
			}
			return nil, err
		}
	}

	loadPrimaryUserLeaderboard := func() (models.Leaderboard, error) {
		if user == nil {
			return []*models.LeaderboardItemRanked{}, nil
		}
		if languageParam == "" {
			return h.leaderboardSrvc.GetByIntervalAndUser(h.leaderboardSrvc.GetDefaultScope(), user.ID, true)
		} else {
			l, err := h.leaderboardSrvc.GetAggregatedByIntervalAndUser(h.leaderboardSrvc.GetDefaultScope(), user.ID, &by, true)
			if err == nil {
				return l.TopByKey(by, languageParam), err
			}
			return nil, err
		}
	}

	primaryLeaderboard, err := loadPrimaryLeaderboard()
	if err != nil {
		conf.Log().Request(r).Error("error while fetching general leaderboard items - %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("something went wrong"))
		return
	}
	primaryLeaderboard.FilterEmpty()

	languageLeaderboard, err := h.leaderboardSrvc.GetAggregatedByInterval(h.leaderboardSrvc.GetDefaultScope(), &by, &utils.PageParams{Page: 1, PageSize: math.MaxUint16}, true)
	if err != nil {
		conf.Log().Request(r).Error("error while fetching language-specific leaderboard items - %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("something went wrong"))
		return
	}

	// regardless of page, always show own rank
	if user != nil && !primaryLeaderboard.HasUser(user.ID) {
		if l, err := loadPrimaryUserLeaderboard(); err == nil {
			primaryLeaderboard.AddMany(l)
		} else {
			conf.Log().Request(r).Error("error while fetching own general user leaderboard - %v", err)
		}
		// no need to fetch language-leaderboard for user, because not using pagination above
	}

	vm := h.buildViewModel(primaryLeaderboard, languageLeaderboard, user, h.leaderboardSrvc.GetDefaultScope(), pageParams)
	vm.Language = languageParam
	helpers.RespondJSON(w, r, http.StatusOK, vm)
}

func (h *LeadersHandler) buildViewModel(globalLeaderboard, languageLeaderboard models.Leaderboard, user *models.User, interval *models.IntervalKey, pageParams *utils.PageParams) *v1.LeadersViewModel {
	var currentUserGlobal []*models.LeaderboardItemRanked
	if user != nil {
		currentUserGlobal = *globalLeaderboard.GetByUser(user.ID)
	}

	totalUsers, _ := h.leaderboardSrvc.CountUsers(true)
	totalPages := int(totalUsers/int64(pageParams.PageSize) + 1)

	_, from, to := helpers.ResolveIntervalTZ(interval, time.UTC)
	numDays := len(utils.SplitRangeByDays(from, to))

	vm := &v1.LeadersViewModel{
		Data:       make([]*v1.LeadersEntry, 0, len(languageLeaderboard.UserIDs())),
		Page:       pageParams.Page,
		TotalPages: totalPages,
		Range: &v1.LeadersRange{
			EndText:   helpers.FormatDateHuman(to),
			EndDate:   to.Format(time.RFC3339),
			StartText: helpers.FormatDateHuman(from),
			StartDate: from.Format(time.RFC3339),
			Name:      (*interval)[0],
			Text:      interval.GetHumanReadable(),
		},
	}

	if len(currentUserGlobal) > 0 {
		vm.CurrentUser = &v1.LeadersCurrentUser{
			Rank: int(currentUserGlobal[0].Rank),
			Page: 1,
			User: v1.NewFromUser(currentUserGlobal[0].User),
		}
	}

	for _, entry := range globalLeaderboard {
		dailyAverage := entry.Total / time.Duration(numDays)

		vm.Data = append(vm.Data, &v1.LeadersEntry{
			Rank: int(entry.Rank),
			RunningTotal: &v1.LeadersRunningTotal{
				TotalSeconds:              float64(entry.Total / time.Second),
				HumanReadableTotal:        helpers.FmtWakatimeDuration(entry.Total),
				DailyAverage:              float64(dailyAverage / time.Second),
				HumanReadableDailyAverage: helpers.FmtWakatimeDuration(dailyAverage),
				Languages: slice.Map[models.LeaderboardKeyTotal, *v1.LeadersLanguage](languageLeaderboard.TopKeysTotalsByUser(models.SummaryLanguage, entry.UserID), func(i int, item models.LeaderboardKeyTotal) *v1.LeadersLanguage {
					return &v1.LeadersLanguage{
						Name:         item.Key,
						TotalSeconds: float64(item.Total / time.Second),
					}
				}),
			},
			User: v1.NewFromUser(entry.User),
		})
	}

	return vm
}
