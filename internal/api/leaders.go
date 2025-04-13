package api

import (
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"

	conf "github.com/muety/wakapi/config"
	v1 "github.com/muety/wakapi/models/compat/wakatime/v1"
)

// @Summary List of users ranked by coding activity in descending order.
// @Description Mimics https://wakatime.com/developers#leaders
// @ID get-wakatime-leaders
// @Tags wakatime
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} v1.LeadersViewModel
// @Router /compat/wakatime/v1/leaders [get]
func (a *APIv1) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	languageParam := strings.ToLower(r.URL.Query().Get("language"))
	pageParams := utils.ParsePageParamsWithDefault(r, 1, 100)
	by := models.SummaryLanguage

	loadPrimaryLeaderboard := func() (models.Leaderboard, error) {
		if languageParam == "" {
			return a.services.LeaderBoard().GetByInterval(a.services.LeaderBoard().GetDefaultScope(), pageParams, true)
		} else {
			l, err := a.services.LeaderBoard().GetAggregatedByInterval(a.services.LeaderBoard().GetDefaultScope(), &by, pageParams, true)
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
			return a.services.LeaderBoard().GetByIntervalAndUser(a.services.LeaderBoard().GetDefaultScope(), user.ID, true)
		} else {
			l, err := a.services.LeaderBoard().GetAggregatedByIntervalAndUser(a.services.LeaderBoard().GetDefaultScope(), user.ID, &by, true)
			if err == nil {
				return l.TopByKey(by, languageParam), err
			}
			return nil, err
		}
	}

	primaryLeaderboard, err := loadPrimaryLeaderboard()
	if err != nil {
		conf.Log().Request(r).Error("error while fetching general leaderboard items", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("something went wrong"))
		return
	}
	primaryLeaderboard.FilterEmpty()

	languageLeaderboard, err := a.services.LeaderBoard().GetAggregatedByInterval(a.services.LeaderBoard().GetDefaultScope(), &by, &utils.PageParams{Page: 1, PageSize: math.MaxUint16}, true)
	if err != nil {
		conf.Log().Request(r).Error("error while fetching language-specific leaderboard items", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("something went wrong"))
		return
	}

	// regardless of page, always show own rank
	if user != nil && !primaryLeaderboard.HasUser(user.ID) {
		if l, err := loadPrimaryUserLeaderboard(); err == nil {
			primaryLeaderboard.AddMany(l)
			primaryLeaderboard = primaryLeaderboard.Deduplicate()
		} else {
			conf.Log().Request(r).Error("error while fetching own general user leaderboard", "userID", user.ID, "error", err)
		}
		// no need to fetch language-leaderboard for user, because not using pagination above
	}

	vm := a.buildLeadersViewModel(primaryLeaderboard, languageLeaderboard, user, a.services.LeaderBoard().GetDefaultScope(), pageParams)
	vm.Language = languageParam
	helpers.RespondJSON(w, r, http.StatusOK, vm)
}

func (a *APIv1) buildLeadersViewModel(globalLeaderboard, languageLeaderboard models.Leaderboard, user *models.User, interval *models.IntervalKey, pageParams *utils.PageParams) *v1.LeadersViewModel {
	var currentUserGlobal []*models.LeaderboardItemRanked
	if user != nil {
		currentUserGlobal = *globalLeaderboard.GetByUser(user.ID)
	}

	totalUsers, _ := a.services.LeaderBoard().CountUsers(true)
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
				Languages: slice.Map(languageLeaderboard.TopKeysTotalsByUser(models.SummaryLanguage, entry.UserID), func(i int, item models.LeaderboardKeyTotal) *v1.LeadersLanguage {
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

// Deduplicate returns a new leaderboard with duplicate user entries removed
// keeping only the first occurrence of each user
func DeduplicateLeaderboard(l models.Leaderboard) models.Leaderboard {
	if len(l) == 0 {
		return l
	}

	seen := make(map[string]bool)
	result := make([]*models.LeaderboardItemRanked, 0, len(l))

	for _, item := range l {
		if !seen[item.UserID] {
			seen[item.UserID] = true
			result = append(result, item)
		}
	}

	return result
}
