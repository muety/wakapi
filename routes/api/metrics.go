package api

import (
	"errors"
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
	"time"
)

const (
	MetricsPrefix = "wakatime"

	DescHeartbeats       = "Total number of tracked heartbeats."
	DescAllTime          = "Total seconds (all time)."
	DescTotal            = "Total seconds."
	DescEditors          = "Total seconds for each editor."
	DescProjects         = "Total seconds for each project."
	DescLanguages        = "Total seconds for each language."
	DescOperatingSystems = "Total seconds for each operating system."
	DescMachines         = "Total seconds for each machine."

	DescAdminTotalTime       = "Total seconds (all users, all time)."
	DescAdminTotalHeartbeats = "Total number of tracked heartbeats (all users, all time)"
	DescAdminUserHeartbeats  = "Total number of tracked heartbeats by user (all time)."
	DescAdminTotalUsers      = "Total number of registered users."
	DescAdminActiveUsers     = "Number of active users."
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
	reqUser := middlewares.GetPrincipal(r)
	if reqUser == nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(conf.ErrUnauthorized))
		return
	}

	var metrics mm.Metrics

	if userMetrics, err := h.getUserMetrics(reqUser); err != nil {
		conf.Log().Request(r).Error("%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(conf.ErrInternalServerError))
		return
	} else {
		for _, m := range *userMetrics {
			metrics = append(metrics, m)
		}
	}

	if reqUser.IsAdmin {
		if adminMetrics, err := h.getAdminMetrics(reqUser); err != nil {
			conf.Log().Request(r).Error("%v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(conf.ErrInternalServerError))
			return
		} else {
			for _, m := range *adminMetrics {
				metrics = append(metrics, m)
			}
		}
	}

	sort.Sort(metrics)

	w.Header().Set("content-type", "text/plain; charset=utf-8")
	w.Write([]byte(metrics.Print()))
}

func (h *MetricsHandler) getUserMetrics(user *models.User) (*mm.Metrics, error) {
	var metrics mm.Metrics

	summaryAllTime, err := h.summarySrvc.Aliased(time.Time{}, time.Now(), user, h.summarySrvc.Retrieve, false)
	if err != nil {
		logbuch.Error("failed to retrieve all time summary for user '%s' for metric", user.ID)
		return nil, err
	}

	from, to := utils.MustResolveIntervalRaw("today")

	summaryToday, err := h.summarySrvc.Aliased(from, to, user, h.summarySrvc.Retrieve, false)
	if err != nil {
		logbuch.Error("failed to retrieve today's summary for user '%s' for metric", user.ID)
		return nil, err
	}

	heartbeatCount, err := h.heartbeatSrvc.CountByUser(user)
	if err != nil {
		logbuch.Error("failed to count heartbeats for user '%s' for metric", user.ID)
		return nil, err
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

	metrics = append(metrics, &mm.CounterMetric{
		Name:   MetricsPrefix + "_heartbeats_total",
		Desc:   DescHeartbeats,
		Value:  int(heartbeatCount),
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

	return &metrics, nil
}

func (h *MetricsHandler) getAdminMetrics(user *models.User) (*mm.Metrics, error) {
	var metrics mm.Metrics

	if !user.IsAdmin {
		return nil, errors.New("unauthorized")
	}

	var totalSeconds int
	if t, err := h.keyValueSrvc.GetString(conf.KeyLatestTotalTime); err == nil && t != nil && t.Value != "" {
		if d, err := time.ParseDuration(t.Value); err == nil {
			totalSeconds = int(d.Seconds())
		}
	}

	totalUsers, _ := h.userSrvc.Count()
	totalHeartbeats, _ := h.heartbeatSrvc.Count()

	activeUsers, err := h.userSrvc.GetActive()
	if err != nil {
		logbuch.Error("failed to retrieve active users for metric – %v", err)
		return nil, err
	}

	metrics = append(metrics, &mm.CounterMetric{
		Name:   MetricsPrefix + "_admin_seconds_total",
		Desc:   DescAdminTotalTime,
		Value:  totalSeconds,
		Labels: []mm.Label{},
	})

	metrics = append(metrics, &mm.CounterMetric{
		Name:   MetricsPrefix + "_admin_heartbeats_total",
		Desc:   DescAdminTotalHeartbeats,
		Value:  int(totalHeartbeats),
		Labels: []mm.Label{},
	})

	metrics = append(metrics, &mm.CounterMetric{
		Name:   MetricsPrefix + "_admin_users_total",
		Desc:   DescAdminTotalUsers,
		Value:  int(totalUsers),
		Labels: []mm.Label{},
	})

	metrics = append(metrics, &mm.CounterMetric{
		Name:   MetricsPrefix + "_admin_users_active_total",
		Desc:   DescAdminActiveUsers,
		Value:  len(activeUsers),
		Labels: []mm.Label{},
	})

	// Count per-user heartbeats

	userCounts, err := h.heartbeatSrvc.CountByUsers(activeUsers)
	if err != nil {
		logbuch.Error("failed to count heartbeats for active users", err.Error())
		return nil, err
	}

	for _, uc := range userCounts {
		metrics = append(metrics, &mm.CounterMetric{
			Name:   MetricsPrefix + "_admin_user_heartbeats_total",
			Desc:   DescAdminUserHeartbeats,
			Value:  int(uc.Count),
			Labels: []mm.Label{{Key: "user", Value: uc.User}},
		})
	}

	return &metrics, nil
}
