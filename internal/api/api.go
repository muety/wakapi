package api

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	mw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sirupsen/logrus"
	_ "gorm.io/driver/mysql"
	_ "gorm.io/driver/postgres"
	_ "gorm.io/driver/sqlite"
	_ "gorm.io/driver/sqlserver"
	"gorm.io/gorm"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/internal/observability"
	"github.com/muety/wakapi/internal/utilities"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/repositories"
	"github.com/muety/wakapi/routes/api"
	shieldsV1Routes "github.com/muety/wakapi/routes/compat/shields/v1"
	wtV1Routes "github.com/muety/wakapi/routes/compat/wakatime/v1"
	"github.com/muety/wakapi/routes/relay"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/services/mail"
	"github.com/sebest/xff"

	_ "net/http/pprof"
)

const (
	audHeaderName        = "X-JWT-AUD"
	defaultVersion       = "unknown version"
	APIVersionHeaderName = "api-version"
)

type API struct {
	handler http.Handler
	db      *gorm.DB
	config  *conf.Config

	// overrideTime can be used to override the clock used by handlers. Should only be used in tests!
	overrideTime func() time.Time
}

func (a *API) Now() time.Time {
	if a.overrideTime != nil {
		return a.overrideTime()
	}

	return time.Now()
}

// NewAPI instantiates a new REST API
// func NewAPI(globalConfig *conf.Config, db *gorm.DB) *API {
// 	return NewAPIWithVersion(globalConfig, db)
// }

func NewAPI(globalConfig *conf.Config, db *gorm.DB) *API {
	api := &API{config: globalConfig, db: db}

	r := newRouter()
	r.chi.NotFound(func(w http.ResponseWriter, r *http.Request) {
		sendJSON(w, http.StatusNotFound, nil, "Resource not found", "The resource you're looking for cannot be found")
	})
	setupGlobalMiddleware(r, api.config)

	r.Get("/health", api.HealthCheck)

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Client-IP", "X-Client-Info", audHeaderName, APIVersionHeaderName},
		ExposedHeaders:   []string{"X-Total-Count", "Link", APIVersionHeaderName},
		AllowCredentials: true,
	})

	api.handler = corsHandler.Handler(r)

	return api
}

func setupGlobalMiddleware(r *router, globalConfig *conf.Config) {
	xffmw, _ := xff.Default() // handles x forwarded by
	logger := observability.NewStructuredLogger(logrus.StandardLogger(), globalConfig)
	if err := observability.ConfigureLogging(&globalConfig.Logging); err != nil {
		logrus.WithError(err).Error("unable to configure logging")
	}
	r.UseBypass(observability.AddRequestID(globalConfig))
	r.UseBypass(logger)
	r.UseBypass(xffmw.Handler)
	r.UseBypass(recoverer)
	r.UseBypass(mw.CleanPath)
	r.UseBypass(mw.StripSlashes)
	r.UseBypass(middlewares.NewPrincipalMiddleware())
	r.UseBypass(
		middlewares.NewLoggingMiddleware(slog.Info, []string{
			"/assets",
			"/favicon",
			"/service-worker.js",
			"/api/health",
			"/api/avatar",
		}),
	)
}

func InitializeJobs(config *conf.Config) {
	// Schedule background tasks
	go conf.StartJobs()
	go aggregationService.Schedule()
	go reportService.Schedule()
	go housekeepingService.Schedule()
	go miscService.Schedule()

	if config.App.LeaderboardEnabled {
		go leaderboardService.Schedule()
	}
}

var (
	aliasRepository           repositories.IAliasRepository
	heartbeatRepository       repositories.IHeartbeatRepository
	userRepository            repositories.IUserRepository
	languageMappingRepository repositories.ILanguageMappingRepository
	projectLabelRepository    repositories.IProjectLabelRepository
	summaryRepository         repositories.ISummaryRepository
	leaderboardRepository     *repositories.LeaderboardRepository
	keyValueRepository        repositories.IKeyValueRepository
	diagnosticsRepository     repositories.IDiagnosticsRepository
	metricsRepository         *repositories.MetricsRepository
	goalRepository            repositories.IGoalRepository
	oauthUserRepository       repositories.IOauthUserRepository
	userAgentPluginRepository repositories.PluginUserAgentRepository
	clientRepository          repositories.IClientRepository
	invoiceRepository         *repositories.InvoiceRepository
)

var (
	aliasService           services.IAliasService
	heartbeatService       services.IHeartbeatService
	userService            services.IUserService
	languageMappingService services.ILanguageMappingService
	projectLabelService    services.IProjectLabelService
	durationService        services.IDurationService
	summaryService         services.ISummaryService
	leaderboardService     services.ILeaderboardService
	aggregationService     services.IAggregationService
	mailService            services.IMailService
	keyValueService        services.IKeyValueService
	reportService          services.IReportService
	activityService        services.IActivityService
	diagnosticsService     services.IDiagnosticsService
	housekeepingService    services.IHousekeepingService
	miscService            services.IMiscService
	goalService            services.IGoalService
	oauthUserService       services.IUserOauthService
	userAgentPluginService services.IPluginUserAgentService
	clientService          services.IClientService
	invoiceService         services.InvoiceService
)

func StartApi(config *conf.Config) {
	db, sqlDB, err := utilities.InitDB(config)

	defer sqlDB.Close()

	if err != nil {
		conf.Log().Fatal("could not connect to database", "error", err)
		os.Exit(1)
		return
	}

	fmt.Println("Connected to database __ successfully")

	// Repositories
	aliasRepository = repositories.NewAliasRepository(db)
	heartbeatRepository = repositories.NewHeartbeatRepository(db)
	userRepository = repositories.NewUserRepository(db)
	languageMappingRepository = repositories.NewLanguageMappingRepository(db)
	projectLabelRepository = repositories.NewProjectLabelRepository(db)
	summaryRepository = repositories.NewSummaryRepository(db)
	leaderboardRepository = repositories.NewLeaderboardRepository(db)
	keyValueRepository = repositories.NewKeyValueRepository(db)
	diagnosticsRepository = repositories.NewDiagnosticsRepository(db)
	metricsRepository = repositories.NewMetricsRepository(db)

	userAgentPluginRepository = repositories.NewPluginUserAgentRepository(db)
	oauthUserRepository = repositories.NewUserOauthRepository(db)
	goalRepository = repositories.NewGoalRepository(db)
	clientRepository = repositories.NewClientRepository(db)
	invoiceRepository = repositories.NewInvoiceRepository(db)

	// Services
	mailService = mail.NewMailService()
	aliasService = services.NewAliasService(aliasRepository)
	userService = services.NewUserService(mailService, userRepository)
	userAgentPluginService = services.NewPluginUserAgentService(&userAgentPluginRepository)
	languageMappingService = services.NewLanguageMappingService(languageMappingRepository)
	projectLabelService = services.NewProjectLabelService(projectLabelRepository)
	heartbeatService = services.NewHeartbeatService(heartbeatRepository, languageMappingService)
	durationService = services.NewDurationService(heartbeatService)
	summaryService = services.NewSummaryService(summaryRepository, heartbeatService, durationService, aliasService, projectLabelService)
	goalService = services.NewGoalService(goalRepository)
	oauthUserService = services.NewUserOauthService(oauthUserRepository)
	aggregationService = services.NewAggregationService(userService, summaryService, heartbeatService)
	keyValueService = services.NewKeyValueService(keyValueRepository)
	reportService = services.NewReportService(summaryService, userService, mailService)
	activityService = services.NewActivityService(summaryService)
	diagnosticsService = services.NewDiagnosticsService(diagnosticsRepository)
	housekeepingService = services.NewHousekeepingService(userService, heartbeatService, summaryService)
	miscService = services.NewMiscService(userService, heartbeatService, summaryService, keyValueService, mailService)
	clientService = services.NewClientService(clientRepository)
	invoiceService = *services.NewInvoiceService(invoiceRepository)
	leaderboardService = services.NewLeaderboardService(leaderboardRepository, summaryService, userService)

	InitializeJobs(config)
	// API Handlers
	authApiHandler := api.NewAuthApiHandler(db, userService, oauthUserService, mailService, aggregationService, summaryService)
	settingsApiHandler := api.NewSettingsHandler(userService, db)
	heartbeatApiHandler := api.NewHeartbeatApiHandler(userService, heartbeatService, languageMappingService, userAgentPluginService)
	summaryApiHandler := api.NewSummaryApiHandler(userService, summaryService)
	metricsHandler := api.NewMetricsHandler(userService, summaryService, heartbeatService, leaderboardService, keyValueService, metricsRepository)
	diagnosticsHandler := api.NewDiagnosticsApiHandler(userService, diagnosticsService)
	avatarHandler := api.NewAvatarHandler()
	activityHandler := api.NewActivityApiHandler(userService, activityService)
	badgeHandler := api.NewBadgeHandler(userService, summaryService)
	captchaHandler := api.NewCaptchaHandler()

	// Compat Handlers
	wakatimeV1StatusBarHandler := wtV1Routes.NewStatusBarHandler(userService, summaryService)
	wakatimeV1AllHandler := wtV1Routes.NewAllTimeHandler(userService, summaryService)
	wakatimeV1SummariesHandler := wtV1Routes.NewSummariesHandler(userService, summaryService)
	wakatimeV1GoalsHandler := wtV1Routes.NewGoalsApiHandler(db, goalService, userService, summaryService)
	wakatimeV1UserAgentsHandler := wtV1Routes.NewUserAgentApiHandler(db, userAgentPluginService, userService)
	wakatimeV1StatsHandler := wtV1Routes.NewStatsHandler(userService, summaryService)
	wakatimeV1UsersHandler := wtV1Routes.NewUsersHandler(userService, heartbeatService)
	wakatimeV1ProjectsHandler := wtV1Routes.NewProjectsHandler(userService, heartbeatService)
	wakatimeV1HeartbeatsHandler := wtV1Routes.NewHeartbeatHandler(userService, heartbeatService)
	wakatimeV1LeadersHandler := wtV1Routes.NewLeadersHandler(userService, leaderboardService)
	shieldV1BadgeHandler := shieldsV1Routes.NewBadgeHandler(summaryService, userService)
	wakatimeV1ClientsHandler := wtV1Routes.NewClientsApiHandler(db, clientService, userService, summaryService)
	wakatimeV1InvoiceHandler := wtV1Routes.NewInvoicesApiHandler(db, invoiceService, userService, summaryService, clientService)

	// Other Handlers
	relayHandler := relay.NewRelayHandler()

	corsSetup := func(r *chi.Mux) {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"https://*", "http://*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300,
		}))
	}

	// Setup Routing
	router := chi.NewRouter()
	corsSetup(router)

	router.Use(
		mw.CleanPath,
		mw.StripSlashes,
		mw.Recoverer,
		middlewares.NewPrincipalMiddleware(),
		middlewares.NewLoggingMiddleware(slog.Info, []string{
			"/assets",
			"/favicon",
			"/service-worker.js",
			"/api/health",
			"/api/avatar",
		}),
	)
	if config.Sentry.Dsn != "" {
		router.Use(middlewares.NewSentryMiddleware())
	}

	// Setup Sub Routers
	rootRouter := chi.NewRouter()
	corsSetup(rootRouter)
	rootRouter.Use(middlewares.NewSecurityMiddleware())

	apiRouter := chi.NewRouter()
	corsSetup(apiRouter)

	// Hook sub routers
	router.Mount("/", rootRouter)
	router.Mount("/api", apiRouter)

	// Route registrations
	relayHandler.RegisterRoutes(rootRouter)

	// API route registrations
	summaryApiHandler.RegisterRoutes(apiRouter)
	// healthApiHandler.RegisterRoutes(apiRouter)
	authApiHandler.RegisterRoutes(apiRouter)
	settingsApiHandler.RegisterRoutes(apiRouter)
	heartbeatApiHandler.RegisterRoutes(apiRouter)
	metricsHandler.RegisterRoutes(apiRouter)
	diagnosticsHandler.RegisterRoutes(apiRouter)
	avatarHandler.RegisterRoutes(apiRouter)
	activityHandler.RegisterRoutes(apiRouter)
	badgeHandler.RegisterRoutes(apiRouter)
	wakatimeV1StatusBarHandler.RegisterRoutes(apiRouter)
	wakatimeV1AllHandler.RegisterRoutes(apiRouter)
	wakatimeV1SummariesHandler.RegisterRoutes(apiRouter)
	wakatimeV1StatsHandler.RegisterRoutes(apiRouter)
	wakatimeV1GoalsHandler.RegisterRoutes(apiRouter)
	wakatimeV1ClientsHandler.RegisterRoutes(apiRouter)
	wakatimeV1InvoiceHandler.RegisterRoutes(apiRouter)
	wakatimeV1UserAgentsHandler.RegisterRoutes(apiRouter)
	wakatimeV1UsersHandler.RegisterRoutes(apiRouter)
	wakatimeV1ProjectsHandler.RegisterRoutes(apiRouter)
	wakatimeV1HeartbeatsHandler.RegisterRoutes(apiRouter)
	wakatimeV1LeadersHandler.RegisterRoutes(apiRouter)
	shieldV1BadgeHandler.RegisterRoutes(apiRouter)
	captchaHandler.RegisterRoutes(apiRouter)

	if config.EnablePprof {
		slog.Info("profiling enabled, exposing pprof data", "url", "http://127.0.0.1:6060/debug/pprof")
		go func() {
			_ = http.ListenAndServe("127.0.0.1:6060", nil)
		}()
	}

	// Listen HTTP
	listen(router, config)
}

// Modify the listen function to store HTTP server references and use the WaitGroup
func listen(handler http.Handler, config *conf.Config) {
	var s4, s6, sSocket *http.Server

	// IPv4
	if config.Server.ListenIpV4 != "-" && config.Server.ListenIpV4 != "" {
		bindString4 := config.Server.ListenIpV4 + ":" + strconv.Itoa(config.Server.Port)
		s4 = &http.Server{
			Handler:      handler,
			Addr:         bindString4,
			ReadTimeout:  time.Duration(config.Server.TimeoutSec) * time.Second,
			WriteTimeout: time.Duration(config.Server.TimeoutSec) * time.Second,
		}
	}

	// IPv6
	if config.Server.ListenIpV6 != "-" && config.Server.ListenIpV6 != "" {
		bindString6 := "[" + config.Server.ListenIpV6 + "]:" + strconv.Itoa(config.Server.Port)
		s6 = &http.Server{
			Handler:      handler,
			Addr:         bindString6,
			ReadTimeout:  time.Duration(config.Server.TimeoutSec) * time.Second,
			WriteTimeout: time.Duration(config.Server.TimeoutSec) * time.Second,
		}
	}

	// UNIX domain socket
	if config.Server.ListenSocket != "-" && config.Server.ListenSocket != "" {
		// Remove if exists
		if _, err := os.Stat(config.Server.ListenSocket); err == nil {
			slog.Info("👉 Removing unix socket", "listenSocket", config.Server.ListenSocket)
			if err := os.Remove(config.Server.ListenSocket); err != nil {
				conf.Log().Fatal(err.Error())
			}
		}
		sSocket = &http.Server{
			Handler:      handler,
			ReadTimeout:  time.Duration(config.Server.TimeoutSec) * time.Second,
			WriteTimeout: time.Duration(config.Server.TimeoutSec) * time.Second,
		}
	}

	if config.UseTLS() {
		if s4 != nil {
			slog.Info("👉 Listening for HTTPS... ✅", "address", s4.Addr)
			go func() {
				if err := s4.ListenAndServeTLS(config.Server.TlsCertPath, config.Server.TlsKeyPath); err != nil {
					conf.Log().Fatal(err.Error())
				}
			}()
		}
		if s6 != nil {
			slog.Info("👉 Listening for HTTPS... ✅", "address", s6.Addr)
			go func() {
				if err := s6.ListenAndServeTLS(config.Server.TlsCertPath, config.Server.TlsKeyPath); err != nil {
					conf.Log().Fatal(err.Error())
				}
			}()
		}
		if sSocket != nil {
			slog.Info("👉 Listening for HTTPS... ✅", "address", config.Server.ListenSocket)
			go func() {
				unixListener, err := net.Listen("unix", config.Server.ListenSocket)
				if err != nil {
					conf.Log().Fatal(err.Error())
				}
				if err := os.Chmod(config.Server.ListenSocket, os.FileMode(config.Server.ListenSocketMode)); err != nil {
					slog.Warn("failed to set user permissions for unix socket", "error", err)
				}
				if err := sSocket.ServeTLS(unixListener, config.Server.TlsCertPath, config.Server.TlsKeyPath); err != nil {
					conf.Log().Fatal(err.Error())
				}
			}()
		}
	} else {
		if s4 != nil {
			slog.Info("👉 Listening for HTTP... ✅", "address", s4.Addr)
			go func() {
				if err := s4.ListenAndServe(); err != nil {
					conf.Log().Fatal(err.Error())
				}
			}()
		}
		if s6 != nil {
			slog.Info("👉 Listening for HTTP... ✅", "address", s6.Addr)
			go func() {
				if err := s6.ListenAndServe(); err != nil {
					conf.Log().Fatal(err.Error())
				}
			}()
		}
		if sSocket != nil {
			slog.Info("👉 Listening for HTTP... ✅", "address", config.Server.ListenSocket)
			go func() {
				unixListener, err := net.Listen("unix", config.Server.ListenSocket)
				if err != nil {
					conf.Log().Fatal(err.Error())
				}
				if err := os.Chmod(config.Server.ListenSocket, os.FileMode(config.Server.ListenSocketMode)); err != nil {
					slog.Warn("failed to set user permissions for unix socket", "error", err)
				}
				if err := sSocket.Serve(unixListener); err != nil {
					conf.Log().Fatal(err.Error())
				}
			}()
		}
	}

	<-make(chan interface{}, 1)
}

type HealthCheckResponse struct {
	Version        string `json:"version"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	DatabaseStatus int    `json:"database_status"`
}

// HealthCheck endpoint indicates if the gotrue api service is available
func (a *API) HealthCheck(w http.ResponseWriter, r *http.Request) error {
	var dbStatus int
	if sqlDb, err := a.db.DB(); err == nil {
		if err := sqlDb.Ping(); err == nil {
			dbStatus = 1
		}
	}
	return sendJSON(w, http.StatusOK, HealthCheckResponse{
		Version:        "0.0.1",
		Name:           "Wakana",
		Description:    "Wakana is an api for developer activity logs generated by IDE plugins",
		DatabaseStatus: dbStatus,
	}, "", "")
}

// ServeHTTP implements the http.Handler interface by passing the request along
// to its underlying Handler.
func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.handler.ServeHTTP(w, r)
}
