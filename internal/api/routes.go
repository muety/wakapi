package api

import (
	"net/http"

	"github.com/dchest/captcha"
	"github.com/go-chi/chi/v5"
	mw "github.com/go-chi/chi/v5/middleware"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/middlewares"
	customMiddleware "github.com/muety/wakapi/middlewares/custom"
	"github.com/muety/wakapi/services"
	"github.com/rs/cors"
)

func (api *APIv1) RegisterApiV1Routes(r *chi.Mux) {
	r.Get("/health", api.HealthCheck)
	r.Get("/api/compat/shields/v1/{user}/*", api.GetShield)
	r.Post("/plugins/errors", api.PostDiagnostics)
	r.Get("/captcha/{id}.png", captcha.Server(captcha.StdWidth, captcha.StdHeight).ServeHTTP)

	r.Group(func(r chi.Router) {
		r.Use(middlewares.NewAuthenticateMiddleware(api.services.Users()).WithOptionalFor("/api/badge/").Handler)
		r.Get("/api/badge/{user}/*", api.GetBadge)
	})

	r.Route("/api", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(
				middlewares.NewAuthenticateMiddleware(api.services.Users()).WithOptionalFor("/api/activity/chart/").Handler,
				mw.Compress(9, "image/svg+xml"),
			)
			r.Get("/chart/{userWithExt}", api.GetActivityChart)
		})

		r.Group(func(r chi.Router) {
			r.Use(
				middlewares.NewAuthenticateMiddleware(api.services.Users()).WithOptionalForMethods(http.MethodOptions).Handler,
				customMiddleware.NewWakatimeRelayMiddleware().Handler,
			)
			if api.config.IsDev() {
				r.Use(
					customMiddleware.NewWakatimeRelayMiddleware().OtherInstancesHandler,
				)
			}
			// see https://github.com/muety/wakapi/issues/203
			r.Post("/heartbeat", api.ProcessHeartBeat)
			r.Post("/heartbeats", api.ProcessHeartBeat)
			r.Post("/users/{user}/heartbeats", api.ProcessHeartBeat)
			r.Post("/users/{user}/heartbeats.bulk", api.ProcessHeartBeat)
			r.Post("/v1/users/{user}/heartbeats", api.ProcessHeartBeat)
			r.Post("/v1/users/{user}/heartbeats.bulk", api.ProcessHeartBeat)
			r.Post("/compat/wakatime/v1/users/{user}/heartbeats", api.ProcessHeartBeat)
			r.Post("/compat/wakatime/v1/users/{user}/heartbeats.bulk", api.ProcessHeartBeat)

			// https://github.com/muety/wakapi/issues/690
			for _, route := range r.Routes() {
				r.Options(route.Pattern, cors.AllowAll().HandlerFunc)
			}
		})
	})
	r.Group(func(r chi.Router) {
		r.Use(
			mw.Compress(9, "image/svg+xml"),
		)
		r.Get("/avatar/{hash}.svg", api.GetAvatarHash)
	})

	// compat/wakatime/v1/leaders

	r.Route("/api/compat/wakatime/v1/users/{user}", func(r chi.Router) {
		r.Use(middlewares.NewAuthenticateMiddleware(api.services.Users()).Handler)
		r.Get("/all_time_since_today", api.GetAllTime)
		r.Get("/heartbeats", api.GetHeartBeats)
		r.Get("/", api.GetUser)

		r.Get("/stats", api.GetUserStats)
		r.Get("/stats/{range}", api.GetUserStats)

		r.Get("/statusbar/{range}", api.GetStatusBarRange)
	})

	r.Get("/api/v1/leaders", api.GetLeaderboard)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", api.HealthCheck)

		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", api.Signin)
			r.Post("/signup", api.Signup)
			r.Post("/oauth/github", api.GithubOauth)
			r.Get("/validate", api.ValidateAuthToken)
			r.Post("/forgot-password", api.ForgotPassword)

			r.Post("/otp/create", services.CreateOTPHandler(api.services.Otp()))
			r.Post("/otp/verify", services.VerifyOTPHandler(api.services.Otp()))

			r.Group(func(r chi.Router) {
				r.Use(middlewares.NewAuthenticateMiddleware(api.services.Users()).Handler)
				r.Get("/api-key", api.GetApiKey)
				r.Post("/api-key/refresh", api.RefreshApiKey)
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(middlewares.NewAuthenticateMiddleware(api.services.Users()).Handler)
			r.Post("/settings", api.UpdateWakatimeSettings)
			r.Get("/profile", api.GetProfile)
			r.Put("/profile", api.SaveProfile)

			r.Get("/summary", api.GetSummary)

			if api.config.Security.ExposeMetrics {
				r.Get("/metrics", api.GetMetrics)
			}

		})
		r.Route("/users/{user}", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(
					middlewares.NewAuthenticateMiddleware(api.services.Users()).Handler,
				)
				r.Get("/user-agents", api.FetchUserAgents)
				r.Get("/summaries", api.GetSummaries)

				r.Route("/clients", func(r chi.Router) {
					r.Post("/", api.CreateClient)
					r.Get("/", api.FetchUserClients)
					r.Get("/{id}", api.GetClient)
					r.Put("/{id}", api.UpdateClient)
					r.Delete("/{id}", api.DeleteClient)
					r.Get("/{id}/invoice/items", api.FetchInvoiceLineItemsForClient)
				})

				r.Route("/goals", func(r chi.Router) {
					r.Post("/", api.CreateGoal)
					r.Get("/", api.FetchUserGoals)
					r.Get("/{id}", api.GetGoal)
					r.Put("/{id}", api.UpdateGoal)
					r.Delete("/{id}", api.DeleteGoal)
				})

				r.Route("/invoices", func(r chi.Router) {
					r.Post("/", api.CreateInvoice)
					r.Get("/", api.FetchUserInvoices)
					r.Get("/{id}", api.GetInvoice)
					r.Put("/{id}", api.UpdateInvoice)
					r.Delete("/{id}", api.DeleteInvoice)
				})

				r.Route("/stats", func(r chi.Router) {
					r.Get("/", api.GetUserStats)
					r.Get("/{range}", api.GetUserStats)
				})

				r.Route("/statusbar", func(r chi.Router) {
					r.Get("/", api.GetStatusBarRange)
					r.Get("/{range}", api.GetStatusBarRange)
				})

				r.Get("/projects", api.GetProjects)
				r.Get("/projects/{id}", api.GetProject)
			})
		})
	})
}

type HealthCheckResponse struct {
	Version        string `json:"version"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	DatabaseStatus int    `json:"database_status"`
}

// HealthCheck endpoint indicates if the gotrue api service is available
func (a *APIv1) HealthCheck(w http.ResponseWriter, r *http.Request) {
	var dbStatus int
	if sqlDb, err := a.db.DB(); err == nil {
		if err := sqlDb.Ping(); err == nil {
			dbStatus = 1
		}
	}
	helpers.RespondJSON(w, r, http.StatusOK, HealthCheckResponse{
		Version:        "0.0.1",
		Name:           "Wakana",
		Description:    "Wakana is an api for developer activity logs generated by IDE plugins",
		DatabaseStatus: dbStatus,
	})
}
