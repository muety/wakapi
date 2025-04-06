package api

import (
	"log/slog"
	"net/http"
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
	"github.com/muety/wakapi/internal/mail"
	"github.com/muety/wakapi/internal/observability"
	"github.com/muety/wakapi/middlewares"

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
	mailService  mail.IMailService
}

func (a *API) Now() time.Time {
	if a.overrideTime != nil {
		return a.overrideTime()
	}

	return time.Now()
}

func NewAPI(globalConfig *conf.Config, db *gorm.DB) *API {
	api := &API{
		config:      globalConfig,
		db:          db,
		mailService: mail.NewMailService(),
	}

	r := newRouter()
	r.chi.NotFound(func(w http.ResponseWriter, r *http.Request) {
		sendJSON(w, http.StatusNotFound, nil, "Resource not found", "The resource you're looking for cannot be found")
	})
	// setupGlobalMiddleware(r, api.config)

	// registerRoutes(r, api)

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

func setupGlobalMiddleware(r *chi.Mux, globalConfig *conf.Config) {
	xffmw, _ := xff.Default() // handles x forwarded by
	logger := observability.NewStructuredLogger(logrus.StandardLogger(), globalConfig)
	if err := observability.ConfigureLogging(&globalConfig.Logging); err != nil {
		logrus.WithError(err).Error("unable to configure logging")
	}
	r.Use(observability.AddRequestID(globalConfig))
	r.Use(logger)
	r.Use(xffmw.Handler)
	r.Use(recoverer)
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

// ServeHTTP implements the http.Handler interface by passing the request along
// to its underlying Handler.
func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.handler.ServeHTTP(w, r)
}
