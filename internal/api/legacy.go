package api

import (
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	mw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/internal/utilities"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/routes/api"
	shieldsV1Routes "github.com/muety/wakapi/routes/compat/shields/v1"
	wtV1Routes "github.com/muety/wakapi/routes/compat/wakatime/v1"
	"github.com/muety/wakapi/routes/relay"
	"github.com/muety/wakapi/services"
	"gorm.io/gorm"
)

func InitializeJobs(config *conf.Config, services services.IServices) {
	// Schedule background tasks
	go conf.StartJobs()
	go services.Aggregation().Schedule()
	go services.Report().Schedule()
	go services.HouseKeeping().Schedule()
	go services.Misc().Schedule()

	if config.App.LeaderboardEnabled {
		go services.LeaderBoard().Schedule()
	}
}

func RegisterApiRoutes(db *gorm.DB, services services.IServices, apiRouter *chi.Mux) {
	// API route registrations
	api.NewAuthApiHandler(db, services).RegisterRoutes(apiRouter)
	api.NewSettingsHandler(db, services).RegisterRoutes(apiRouter)
	api.NewHeartbeatApiHandler(services).RegisterRoutes(apiRouter)
	api.NewSummaryApiHandler(services).RegisterRoutes(apiRouter)
	api.NewMetricsHandler(db).RegisterRoutes(apiRouter)
	api.NewDiagnosticsApiHandler(services).RegisterRoutes(apiRouter)
	api.NewAvatarHandler().RegisterRoutes(apiRouter)
	api.NewActivityApiHandler(services).RegisterRoutes(apiRouter)
	api.NewBadgeHandler(services).RegisterRoutes(apiRouter)
	api.NewCaptchaHandler().RegisterRoutes(apiRouter)

	// WakaTime v1 routes
	wtV1Routes.NewStatusBarHandler(services).RegisterRoutes(apiRouter)
	wtV1Routes.NewAllTimeHandler(services).RegisterRoutes(apiRouter)
	wtV1Routes.NewSummariesHandler(services).RegisterRoutes(apiRouter)
	wtV1Routes.NewGoalsApiHandler(db, services).RegisterRoutes(apiRouter)
	wtV1Routes.NewUserAgentApiHandler(db, services).RegisterRoutes(apiRouter)
	wtV1Routes.NewStatsHandler(services).RegisterRoutes(apiRouter)
	wtV1Routes.NewClientsApiHandler(db, services).RegisterRoutes(apiRouter)
	wtV1Routes.NewInvoicesApiHandler(db, services).RegisterRoutes(apiRouter)
	wtV1Routes.NewUsersHandler(services).RegisterRoutes(apiRouter)
	wtV1Routes.NewProjectsHandler(services).RegisterRoutes(apiRouter)
	wtV1Routes.NewHeartbeatHandler(services).RegisterRoutes(apiRouter)
	wtV1Routes.NewLeadersHandler(services).RegisterRoutes(apiRouter)

	// Shields v1 routes
	shieldsV1Routes.NewBadgeHandler(services).RegisterRoutes(apiRouter)
}

func StartApi(config *conf.Config) {
	db, sqlDB, err := utilities.InitDB(config)

	if err != nil {
		conf.Log().Fatal("could not connect to database", "error", err)
		os.Exit(1)
		return
	}

	defer sqlDB.Close()

	services := services.NewServices(db)
	InitializeJobs(config, services)

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

	relayHandler.RegisterRoutes(rootRouter)
	RegisterApiRoutes(db, services, apiRouter)

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
// listen initializes and starts HTTP/HTTPS servers based on the configuration.
// It supports IPv4, IPv6, and UNIX domain sockets.
func listen(handler http.Handler, config *conf.Config) {
	var s4, s6, sSocket *http.Server

	// Configure IPv4 server
	if config.Server.ListenIpV4 != "-" && config.Server.ListenIpV4 != "" {
		bindString4 := config.Server.ListenIpV4 + ":" + strconv.Itoa(config.Server.Port)
		s4 = createServer(handler, bindString4, config)
	}

	// Configure IPv6 server
	if config.Server.ListenIpV6 != "-" && config.Server.ListenIpV6 != "" {
		bindString6 := "[" + config.Server.ListenIpV6 + "]:" + strconv.Itoa(config.Server.Port)
		s6 = createServer(handler, bindString6, config)
	}

	// Configure UNIX domain socket server
	if config.Server.ListenSocket != "-" && config.Server.ListenSocket != "" {
		sSocket = configureUnixSocket(handler, config)
	}

	// Start servers based on TLS configuration
	if config.UseTLS() {
		startTLSServers(s4, s6, sSocket, config)
	} else {
		startHTTPServers(s4, s6, sSocket, config)
	}

	// Block the main goroutine to keep the servers running
	<-make(chan interface{}, 1)
}

// createServer creates an HTTP server with the given handler and address.
func createServer(handler http.Handler, address string, config *conf.Config) *http.Server {
	return &http.Server{
		Handler:      handler,
		Addr:         address,
		ReadTimeout:  time.Duration(config.Server.TimeoutSec) * time.Second,
		WriteTimeout: time.Duration(config.Server.TimeoutSec) * time.Second,
	}
}

// configureUnixSocket sets up a server for a UNIX domain socket.
func configureUnixSocket(handler http.Handler, config *conf.Config) *http.Server {
	// Remove existing socket file if it exists
	if _, err := os.Stat(config.Server.ListenSocket); err == nil {
		slog.Info("👉 Removing existing UNIX socket", "listenSocket", config.Server.ListenSocket)
		if err := os.Remove(config.Server.ListenSocket); err != nil {
			conf.Log().Fatal(err.Error())
		}
	}

	return &http.Server{
		Handler:      handler,
		ReadTimeout:  time.Duration(config.Server.TimeoutSec) * time.Second,
		WriteTimeout: time.Duration(config.Server.TimeoutSec) * time.Second,
	}
}

// startTLSServers starts the servers with TLS (HTTPS) configuration.
func startTLSServers(s4, s6, sSocket *http.Server, config *conf.Config) {
	if s4 != nil {
		slog.Info("👉 Listening for HTTPS on IPv4... ✅", "address", s4.Addr)
		go func() {
			if err := s4.ListenAndServeTLS(config.Server.TlsCertPath, config.Server.TlsKeyPath); err != nil {
				conf.Log().Fatal(err.Error())
			}
		}()
	}

	if s6 != nil {
		slog.Info("👉 Listening for HTTPS on IPv6... ✅", "address", s6.Addr)
		go func() {
			if err := s6.ListenAndServeTLS(config.Server.TlsCertPath, config.Server.TlsKeyPath); err != nil {
				conf.Log().Fatal(err.Error())
			}
		}()
	}

	if sSocket != nil {
		slog.Info("👉 Listening for HTTPS on UNIX socket... ✅", "address", config.Server.ListenSocket)
		go func() {
			unixListener, err := net.Listen("unix", config.Server.ListenSocket)
			if err != nil {
				conf.Log().Fatal(err.Error())
			}
			if err := os.Chmod(config.Server.ListenSocket, os.FileMode(config.Server.ListenSocketMode)); err != nil {
				slog.Warn("Failed to set permissions for UNIX socket", "error", err)
			}
			if err := sSocket.ServeTLS(unixListener, config.Server.TlsCertPath, config.Server.TlsKeyPath); err != nil {
				conf.Log().Fatal(err.Error())
			}
		}()
	}
}

// startHTTPServers starts the servers without TLS (HTTP) configuration.
func startHTTPServers(s4, s6, sSocket *http.Server, config *conf.Config) {
	if s4 != nil {
		slog.Info("👉 Listening for HTTP on IPv4... ✅", "address", s4.Addr)
		go func() {
			if err := s4.ListenAndServe(); err != nil {
				conf.Log().Fatal(err.Error())
			}
		}()
	}

	if s6 != nil {
		slog.Info("👉 Listening for HTTP on IPv6... ✅", "address", s6.Addr)
		go func() {
			if err := s6.ListenAndServe(); err != nil {
				conf.Log().Fatal(err.Error())
			}
		}()
	}

	if sSocket != nil {
		slog.Info("👉 Listening for HTTP on UNIX socket... ✅", "address", config.Server.ListenSocket)
		go func() {
			unixListener, err := net.Listen("unix", config.Server.ListenSocket)
			if err != nil {
				conf.Log().Fatal(err.Error())
			}
			if err := os.Chmod(config.Server.ListenSocket, os.FileMode(config.Server.ListenSocketMode)); err != nil {
				slog.Warn("Failed to set permissions for UNIX socket", "error", err)
			}
			if err := sSocket.Serve(unixListener); err != nil {
				conf.Log().Fatal(err.Error())
			}
		}()
	}
}
