package v1

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/muety/wakapi/helpers"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	v1 "github.com/muety/wakapi/models/compat/wakatime/v1"
	"github.com/muety/wakapi/services"
)

type StatsHandler struct {
	config      *conf.Config
	userSrvc    services.IUserService
	summarySrvc services.ISummaryService
}

func NewStatsHandler(userService services.IUserService, summaryService services.ISummaryService) *StatsHandler {
	return &StatsHandler{
		userSrvc:    userService,
		summarySrvc: summaryService,
		config:      conf.Get(),
	}
}

func (h *StatsHandler) RegisterRoutes(router chi.Router) {
	router.Group(func(r chi.Router) {
		r.Use(
			middlewares.NewAuthenticateMiddleware(h.userSrvc).WithOptionalFor("/").Handler,
		)
		r.Get("/v1/users/{user}/stats/{range}", h.Get)
		r.Get("/compat/wakatime/v1/users/{user}/stats/{range}", h.Get)

		// Also works without range, see https://github.com/anuraghazra/github-readme-stats/issues/865#issuecomment-776186592
		r.Get("/v1/users/{user}/stats", h.Get)
		r.Get("/compat/wakatime/v1/users/{user}/stats", h.Get)
	})
}

// TODO: support filtering (requires https://github.com/muety/wakapi/issues/108)

// @Summary Retrieve statistics for a given user
// @Description Mimics https://wakatime.com/developers#stats
// @ID get-wakatimes-tats
// @Tags wakatime
// @Produce json
// @Param user path string true "User ID to fetch data for (or 'current')"
// @Param range path string false "Range interval identifier" Enums(today, yesterday, week, month, year, 7_days, last_7_days, 30_days, last_30_days, 6_months, last_6_months, 12_months, last_12_months, last_year, any, all_time)
// @Param project query string false "Project to filter by"
// @Param language query string false "Language to filter by"
// @Param editor query string false "Editor to filter by"
// @Param operating_system query string false "OS to filter by"
// @Param machine query string false "Machine to filter by"
// @Param label query string false "Project label to filter by"
// @Security ApiKeyAuth
// @Success 200 {object} v1.StatsViewModel
// @Router /compat/wakatime/v1/users/{user}/stats/{range} [get]
func (h *StatsHandler) Get(w http.ResponseWriter, r *http.Request) {
	userParam := chi.URLParam(r, "user")
	rangeParam := chi.URLParam(r, "range")
	var authorizedUser, requestedUser *models.User

	authorizedUser = middlewares.GetPrincipal(r)
	if authorizedUser != nil && userParam == "current" {
		userParam = authorizedUser.ID
	}

	requestedUser, err := h.userSrvc.GetUserById(userParam)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("user not found"))
		return
	}

	// if no range was requested, get the maximum allowed range given the users max shared days, otherwise default to past 7 days (which will fail in the next step, because user didn't allow any sharing)
	// this "floors" the user's maximum shared date to the supported range buckets (e.g. if user opted to share 12 days, we'll still fallback to "last_7_days") for consistency with wakatime
	if rangeParam == "" {
		if _, userRange := helpers.ResolveMaximumRange(requestedUser.ShareDataMaxDays); userRange != nil {
			rangeParam = (*userRange)[1]
		} else {
			rangeParam = (*models.IntervalPast7Days)[1]
		}
	}

	err, rangeFrom, rangeTo := helpers.ResolveIntervalRawTZ(rangeParam, requestedUser.TZ(), requestedUser.StartOfWeekDay())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid range"))
		return
	}

	minStart := rangeTo.AddDate(0, 0, -requestedUser.ShareDataMaxDays)
	if (authorizedUser == nil || requestedUser.ID != authorizedUser.ID) &&
		rangeFrom.Before(minStart) && requestedUser.ShareDataMaxDays >= 0 {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("requested time range too broad"))
		return
	}

	summary, err, status := h.loadUserSummary(requestedUser, rangeFrom, rangeTo, helpers.ParseSummaryFilters(r))
	if err != nil {
		w.WriteHeader(status)
		w.Write([]byte(err.Error()))
		return
	}

	if requestedUser.ID != summary.UserID {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("mismatch between requested user and summary owner")) // should never happen
		conf.Log().Request(r).Error("mismatch between requested user and summary owner", "requested_user", requestedUser.ID, "summary_user", summary.UserID, "summary_id", summary.ID)
		return
	}
	summary.User = requestedUser

	stats := v1.NewStatsFrom(summary, &models.Filters{})
	stats.Data.Range = rangeParam
	stats.Data.HumanReadableRange = helpers.MustParseInterval(rangeParam).GetHumanReadable()
	stats.Data.IsCodingActivityVisible = requestedUser.ShareDataMaxDays != 0
	stats.Data.IsOtherUsageVisible = requestedUser.AnyDataShared()

	if authorizedUser == nil || requestedUser.ID != authorizedUser.ID {
		// post filter stats according to user's given sharing permissions
		if !requestedUser.ShareEditors {
			stats.Data.Editors = make([]*v1.SummariesEntry, 0)
		}
		if !requestedUser.ShareLanguages {
			stats.Data.Languages = make([]*v1.SummariesEntry, 0)
		}
		if !requestedUser.ShareProjects {
			stats.Data.Projects = make([]*v1.SummariesEntry, 0)
		}
		if !requestedUser.ShareOSs {
			stats.Data.OperatingSystems = make([]*v1.SummariesEntry, 0)
		}
		if !requestedUser.ShareMachines {
			stats.Data.Machines = make([]*v1.SummariesEntry, 0)
		}
	}

	helpers.RespondJSON(w, r, http.StatusOK, stats)
}

func (h *StatsHandler) loadUserSummary(user *models.User, start, end time.Time, filters *models.Filters) (*models.Summary, error, int) {
	overallParams := &models.SummaryParams{
		From:      start,
		To:        end,
		User:      user,
		Recompute: false,
	}

	summary, err := h.summarySrvc.Aliased(overallParams.From, overallParams.To, user, h.summarySrvc.Retrieve, filters, nil, false)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}

	return summary, nil, http.StatusOK
}
