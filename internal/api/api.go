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
	lru "github.com/hashicorp/golang-lru"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/internal/mail"
	"github.com/muety/wakapi/internal/observability"
	"github.com/muety/wakapi/internal/utilities"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/routes/relay"
	"github.com/muety/wakapi/services"
	"github.com/patrickmn/go-cache"
	"github.com/sebest/xff"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type APIv1 struct {
	db     *gorm.DB
	config *conf.Config

	// overrideTime can be used to override the clock used by handlers. Should only be used in tests!
	overrideTime func() time.Time
	mailService  mail.IMailService
	services     services.IServices
	httpClient   *http.Client
	cache        *cache.Cache
	lruCache     *lru.Cache
}

func (a *APIv1) Now() time.Time {
	if a.overrideTime != nil {
		return a.overrideTime()
	}

	return time.Now()
}

func setupGlobalMiddleware(r *chi.Mux, globalConfig *conf.Config) {
	xffmw, _ := xff.Default() // handles x forwarded by
	logger := observability.NewStructuredLogger(logrus.StandardLogger(), globalConfig)
	if err := observability.ConfigureLogging(&globalConfig.Logging); err != nil {
		logrus.WithError(err).Error("unable to configure logging")
	}
	r.Use(observability.AddRequestID(globalConfig))
	r.Use(logger)
	r.Use(xffmw.Handler)
	r.Use(mw.CleanPath)
	r.Use(mw.StripSlashes)
	r.Use(middlewares.NewPrincipalMiddleware())
	r.Use(
		middlewares.NewLoggingMiddleware(slog.Info, []string{
			"/assets",
			"/favicon",
			"/service-worker.js",
			"/api/health",
			"/api/avatar",
		}),
	)
}

func NewAPIv1(globalConfig *conf.Config, db *gorm.DB) *APIv1 {
	api := &APIv1{
		config:      globalConfig,
		db:          db,
		mailService: mail.NewMailService(),
		services:    services.NewServices(db),
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		cache:       cache.New(6*time.Hour, 6*time.Hour),
	}

	lruCache, err := lru.New(1 * 1000 * 64)
	if err != nil {
		panic(err)
	}

	api.lruCache = lruCache

	return api
}

func (a *APIv1) initializeJobs() {
	// Schedule background tasks
	go conf.StartJobs()
	go a.services.Aggregation().Schedule()
	go a.services.Report().Schedule()
	go a.services.HouseKeeping().Schedule()
	go a.services.Misc().Schedule()

	if a.config.App.LeaderboardEnabled {
		go a.services.LeaderBoard().Schedule()
	}
}

func corsSetup(r *chi.Mux) {
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
}

func StartApi(config *conf.Config) {
	db, sqlDB, err := utilities.InitDB(config)

	if err != nil {
		conf.Log().Fatal("could not connect to database", "error", err)
		os.Exit(1)
		return
	}

	defer sqlDB.Close()

	api := NewAPIv1(config, db)

	// Other Handlers
	relayHandler := relay.NewRelayHandler()

	// Setup Routing
	router := chi.NewRouter()
	corsSetup(router)

	router.Use(
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

	setupGlobalMiddleware(router, api.config)

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		sendJSON(w, http.StatusNotFound, nil, "Resource not found", "The resource you're looking for cannot be found")
	})

	// Hook sub routers
	router.Group(func(r chi.Router) {
		r.Use(middlewares.NewSecurityMiddleware())
		relayHandler.RegisterRoutes(r)
	})

	if config.EnablePprof {
		slog.Info("profiling enabled, exposing pprof data", "url", "http://127.0.0.1:6060/debug/pprof")
		go func() {
			_ = http.ListenAndServe("127.0.0.1:6060", nil)
		}()
	}

	// Listen HTTP
	api.RegisterApiV1Routes(router)
	api.initializeJobs()
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
