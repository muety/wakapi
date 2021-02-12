package api

import (
	"github.com/emvi/logbuch"
	"github.com/gorilla/mux"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	v1 "github.com/muety/wakapi/models/compat/wakatime/v1"
	mm "github.com/muety/wakapi/models/metrics"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"net/http"
	"sort"
	"strconv"
	"time"
)

const (
	MetricsPrefix = "wakatime"

	DescAllTime          = "Total seconds (all time)."
	DescTotal            = "Total seconds."
	DescEditors          = "Total seconds for each editor."
	DescProjects         = "Total seconds for each project."
	DescLanguages        = "Total seconds for each language."
	DescOperatingSystems = "Total seconds for each operating system."
	DescMachines         = "Total seconds for each machine."

	DescAdminTotalTime       = "Total seconds (all users, all time)"
	DescAdminTotalHeartbeats = "Total number of tracked heartbeats (all users, all time)"
	DescAdminTotalUser       = "Total number of registered users"
)

type MetricsHandler struct {
	config        *conf.Config
	userSrvc      services.IUserService
	summarySrvc   services.ISummaryService
	heartbeatSrvc services.IHeartbeatService
	keyValueSrvc  services.IKeyValueService
}

func NewMetricsHandler(userService services.IUserService, summaryService services.ISummaryService, heartbeatService services.IHeartbeatService, keyValueService services.IKeyValueService) *MetricsHandler {
	return &MetricsHandler{
		userSrvc:      userService,
		summarySrvc:   summaryService,
		heartbeatSrvc: heartbeatService,
		keyValueSrvc:  keyValueService,
		config:        conf.Get(),
	}
}

func (h *MetricsHandler) RegisterRoutes(router *mux.Router) {
	if !h.config.Security.ExposeMetrics {
		return
	}

	logbuch.Info("exposing prometheus metrics under /api/metrics")

	r := router.PathPrefix("/metrics").Subrouter()
	r.Use(
		middlewares.NewAuthenticateMiddleware(h.userSrvc).Handler,
	)
	r.Methods(http.MethodGet).HandlerFunc(h.Get)
}

func (h *MetricsHandler) Get(w http.ResponseWriter, r *http.Request) {
	var metrics mm.Metrics

	user := r.Context().Value(models.UserKey).(*models.User)
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	summaryAllTime, err := h.summarySrvc.Aliased(time.Time{}, time.Now(), user, h.summarySrvc.Retrieve)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	from, to := utils.MustResolveIntervalRaw("today")

	summaryToday, err := h.summarySrvc.Aliased(from, to, user, h.summarySrvc.Retrieve)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// User Metrics

	metrics = append(metrics, &mm.CounterMetric{
		Name:   MetricsPrefix + "_cumulative_seconds_total",
		Desc:   DescAllTime,
		Value:  int(v1.NewAllTimeFrom(summaryAllTime, &models.Filters{}).Data.TotalSeconds),
		Labels: []mm.Label{},
	})

	metrics = append(metrics, &mm.CounterMetric{
		Name:   MetricsPrefix + "_seconds_total",
		Desc:   DescTotal,
		Value:  int(summaryToday.TotalTime().Seconds()),
		Labels: []mm.Label{},
	})

	for _, p := range summaryToday.Projects {
		metrics = append(metrics, &mm.CounterMetric{
			Name:   MetricsPrefix + "_project_seconds_total",
			Desc:   DescProjects,
			Value:  int(summaryToday.TotalTimeByKey(models.SummaryProject, p.Key).Seconds()),
			Labels: []mm.Label{{Key: "name", Value: p.Key}},
		})
	}

	for _, l := range summaryToday.Languages {
		metrics = append(metrics, &mm.CounterMetric{
			Name:   MetricsPrefix + "_language_seconds_total",
			Desc:   DescLanguages,
			Value:  int(summaryToday.TotalTimeByKey(models.SummaryLanguage, l.Key).Seconds()),
			Labels: []mm.Label{{Key: "name", Value: l.Key}},
		})
	}

	for _, e := range summaryToday.Editors {
		metrics = append(metrics, &mm.CounterMetric{
			Name:   MetricsPrefix + "_editor_seconds_total",
			Desc:   DescEditors,
			Value:  int(summaryToday.TotalTimeByKey(models.SummaryEditor, e.Key).Seconds()),
			Labels: []mm.Label{{Key: "name", Value: e.Key}},
		})
	}

	for _, o := range summaryToday.OperatingSystems {
		metrics = append(metrics, &mm.CounterMetric{
			Name:   MetricsPrefix + "_operating_system_seconds_total",
			Desc:   DescOperatingSystems,
			Value:  int(summaryToday.TotalTimeByKey(models.SummaryOS, o.Key).Seconds()),
			Labels: []mm.Label{{Key: "name", Value: o.Key}},
		})
	}

	for _, m := range summaryToday.Machines {
		metrics = append(metrics, &mm.CounterMetric{
			Name:   MetricsPrefix + "_machine_seconds_total",
			Desc:   DescMachines,
			Value:  int(summaryToday.TotalTimeByKey(models.SummaryMachine, m.Key).Seconds()),
			Labels: []mm.Label{{Key: "name", Value: m.Key}},
		})
	}

	// Admin metrics

	if user.IsAdmin {
		var (
			totalSeconds    int
			totalUsers      int
			totalHeartbeats int
		)

		if t, err := h.keyValueSrvc.GetString(conf.KeyLatestTotalTime); err == nil && t != nil && t.Value != "" {
			if d, err := time.ParseDuration(t.Value); err == nil {
				totalSeconds = int(d.Seconds())
			}
		}

		if t, err := h.keyValueSrvc.GetString(conf.KeyLatestTotalUsers); err == nil && t != nil && t.Value != "" {
			if d, err := strconv.Atoi(t.Value); err == nil {
				totalUsers = d
			}
		}

		if t, err := h.keyValueSrvc.GetString(conf.KeyLatestTotalUsers); err == nil && t != nil && t.Value != "" {
			if d, err := strconv.Atoi(t.Value); err == nil {
				totalUsers = d
			}
		}

		if t, err := h.heartbeatSrvc.Count(); err == nil {
			totalHeartbeats = int(t)
		}

		metrics = append(metrics, &mm.CounterMetric{
			Name:   MetricsPrefix + "_admin_seconds_total",
			Desc:   DescAdminTotalTime,
			Value:  totalSeconds,
			Labels: []mm.Label{},
		})

		metrics = append(metrics, &mm.CounterMetric{
			Name:   MetricsPrefix + "_admin_users_total",
			Desc:   DescAdminTotalUser,
			Value:  totalUsers,
			Labels: []mm.Label{},
		})

		metrics = append(metrics, &mm.CounterMetric{
			Name:   MetricsPrefix + "_admin_heartbeats_total",
			Desc:   DescAdminTotalHeartbeats,
			Value:  totalHeartbeats,
			Labels: []mm.Label{},
		})
	}

	sort.Sort(metrics)

	w.Header().Set("content-type", "text/plain; charset=utf-8")
	w.Write([]byte(metrics.Print()))
}
