package routes

import (
	"fmt"
	"net/http"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/go-chi/chi/v5"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/models/view"
	su "github.com/muety/wakapi/routes/utils"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
)

const (
	dailyStatsMinRangeDays = 3
	dailyStatsMaxRangeDays = 31
)

type SummaryHandler struct {
	config         *conf.Config
	userSrvc       services.IUserService
	summarySrvc    services.ISummaryService
	durationSrvc   services.IDurationService
	aliasSrvc      services.IAliasService
	heartbeatsSrvc services.IHeartbeatService
}

func NewSummaryHandler(summaryService services.ISummaryService, userService services.IUserService, heartbeatsService services.IHeartbeatService, durationService services.IDurationService, aliasService services.IAliasService) *SummaryHandler {
	return &SummaryHandler{
		summarySrvc:    summaryService,
		userSrvc:       userService,
		heartbeatsSrvc: heartbeatsService,
		durationSrvc:   durationService,
		aliasSrvc:      aliasService,
		config:         conf.Get(),
	}
}

func (h *SummaryHandler) RegisterRoutes(router chi.Router) {
	r := chi.NewRouter()
	r.Use(middlewares.NewAuthenticateMiddleware(h.userSrvc).
		WithRedirectTarget(defaultErrorRedirectTarget()).
		WithRedirectErrorMessage("unauthorized").Handler,
	)
	r.Get("/", h.GetIndex)

	router.Mount("/summary", r)
}

func (h *SummaryHandler) GetIndex(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	rawQuery := r.URL.RawQuery
	q := r.URL.Query()
	if q.Get("interval") == "" && q.Get("from") == "" {
		// If the PersistentIntervalKey cookie is set, redirect to the correct summary page
		if intervalCookie, _ := r.Cookie(models.PersistentIntervalKey); intervalCookie != nil {
			redirectAddress := fmt.Sprintf("%s/summary?interval=%s", h.config.Server.BasePath, intervalCookie.Value)
			http.Redirect(w, r, redirectAddress, http.StatusFound)
		}

		q.Set("interval", "today")
		r.URL.RawQuery = q.Encode()
	} else if q.Get("interval") != "" {
		// Send a Set-Cookie header to persist the interval
		headerValue := fmt.Sprintf("%s=%s", models.PersistentIntervalKey, q.Get("interval"))
		w.Header().Add("Set-Cookie", headerValue)
	}

	summaryParams, _ := helpers.ParseSummaryParams(r)
	summary, err, status := su.LoadUserSummaryByParams(h.summarySrvc, summaryParams)
	if err != nil {
		conf.Log().Request(r).Error("failed to load summary", "error", err)
		w.WriteHeader(status)
		templates[conf.SummaryTemplate].Execute(w, h.buildViewModel(r, w).WithError(err.Error()))
		return
	}
	// retrieved for showing all available filters
	summaryWithoutFilter, err, status := su.LoadUserSummaryWithoutFilter(h.summarySrvc, summaryParams)
	if err != nil {
		conf.Log().Request(r).Error("failed to load summary", "error", err)
		w.WriteHeader(status)
		templates[conf.SummaryTemplate].Execute(w, h.buildViewModel(r, w).WithError(err.Error()))
		return
	}
	availableFilters := h.extractAvailableFilters(summaryWithoutFilter)

	user := middlewares.GetPrincipal(r)
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		templates[conf.SummaryTemplate].Execute(w, h.buildViewModel(r, w).WithError("unauthorized"))
		return
	}

	// user first data
	firstData, err := h.heartbeatsSrvc.GetFirstByUser(user)
	if err != nil {
		conf.Log().Request(r).Error("error while user's heartbeats range", "user", user.ID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		templates[conf.SummaryTemplate].Execute(w, h.buildViewModel(r, w).WithError(err.Error()))
		return
	}

	// timeline data (daily stats)
	var timeline []*view.TimelineViewModel
	if rangeDays := summaryParams.RangeDays(); rangeDays >= dailyStatsMinRangeDays && rangeDays <= dailyStatsMaxRangeDays {
		dailyStatsSummaries, err := h.fetchSplitSummaries(summaryParams)
		if err != nil {
			conf.Log().Request(r).Error("failed to load timeline stats", "error", err)
		} else {
			timeline = view.NewTimelineViewModel(dailyStatsSummaries)
		}
	}

	// hourly breakdown data
	var hourlyBreakdown view.HourlyBreakdownsViewModel
	hourlyBreakdownFrom := summaryParams.From
	if summaryParams.RangeDays() > 1 { // get at most 24 hours of hourly breakdown
		hourlyBreakdownFrom = summaryParams.To.Add(-24 * time.Hour)
	}
	if summaries, err := h.durationSrvc.Get(hourlyBreakdownFrom, summaryParams.To, summaryParams.User, summaryParams.Filters, nil, false); err == nil {
		hourlyBreakdown = view.NewHourlyBreakdownViewModel(view.NewHourlyBreakdownItems(summaries, func(t uint8, k string) string {
			s, _ := h.aliasSrvc.GetAliasOrDefault(user.ID, t, k)
			return s
		}))
	} else {
		conf.Log().Request(r).Error("failed to load hourly breakdown stats", "error", err)
	}

	vm := view.SummaryViewModel{
		SharedLoggedInViewModel: view.SharedLoggedInViewModel{
			SharedViewModel: view.NewSharedViewModel(h.config, nil),
			User:            user,
		},
		AvailableFilters:    availableFilters,
		Summary:             summary,
		SummaryParams:       summaryParams,
		EditorColors:        su.FilterColors(h.config.App.GetEditorColors(), summary.Editors),
		LanguageColors:      su.FilterColors(h.config.App.GetLanguageColors(), summary.Languages),
		OSColors:            su.FilterColors(h.config.App.GetOSColors(), summary.OperatingSystems),
		RawQuery:            rawQuery,
		UserFirstData:       firstData,
		DataRetentionMonths: h.config.App.DataRetentionMonths,
		Timeline:            timeline,
		HourlyBreakdown:     hourlyBreakdown,
		HourlyBreakdownFrom: hourlyBreakdownFrom,
	}

	templates[conf.SummaryTemplate].Execute(w, vm)
}

func (h *SummaryHandler) buildViewModel(r *http.Request, w http.ResponseWriter) *view.SummaryViewModel {
	user := middlewares.GetPrincipal(r)
	return su.WithSessionMessages(&view.SummaryViewModel{
		SharedLoggedInViewModel: view.SharedLoggedInViewModel{
			User:            user,
			SharedViewModel: view.NewSharedViewModel(h.config, nil),
		},
	}, r, w)
}

func (h *SummaryHandler) fetchSplitSummaries(params *models.SummaryParams) ([]*models.Summary, error) {
	summaries := make([]*models.Summary, 0)
	intervals := utils.SplitRangeByDays(params.From, params.To)
	for _, interval := range intervals {
		curSummary, err := h.summarySrvc.Aliased(interval[0], interval[1], params.User, h.summarySrvc.Retrieve, params.Filters, nil, false)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, curSummary)
	}
	return summaries, nil
}

// extractAvailableFilters extracts available filter names from a summary's various item collections.
func (h *SummaryHandler) extractAvailableFilters(summary *models.Summary) view.AvailableFilters {
	return view.AvailableFilters{
		ProjectNames:  slice.Map(summary.Projects, func(_ int, item *models.SummaryItem) string { return item.Key }),
		LanguageNames: slice.Map(summary.Languages, func(_ int, item *models.SummaryItem) string { return item.Key }),
		MachineNames:  slice.Map(summary.Machines, func(_ int, item *models.SummaryItem) string { return item.Key }),
		LabelNames:    slice.Map(summary.Labels, func(_ int, item *models.SummaryItem) string { return item.Key }),
		CategoryNames: slice.Map(summary.Categories, func(_ int, item *models.SummaryItem) string { return item.Key }),
	}
}
