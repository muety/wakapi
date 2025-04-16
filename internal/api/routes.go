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
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		helpers.RespondJSON(w, r, http.StatusNotFound, map[string]interface{}{
			"error":   "not found",
			"message": "The requested resource was not found",
		})
	})
	// Global routes
	r.Get("/health", api.HealthCheck)
	r.Get("/api/compat/shields/v1/{user}/*", api.GetShield)
	r.Post("/plugins/errors", api.PostDiagnostics)
	r.Get("/captcha/{id}.png", captcha.Server(captcha.StdWidth, captcha.StdHeight).ServeHTTP)

	r.Group(func(r chi.Router) {
		r.Use(middlewares.NewAuthenticateMiddleware(api.services.Users()).WithOptionalFor("/api/badge/").Handler)
		r.Get("/api/badge/{user}/*", api.GetBadge)
	})

	r.Route("/api/chart", func(r chi.Router) {
		r.Use(
			middlewares.NewAuthenticateMiddleware(api.services.Users()).WithOptionalFor("/api/activity/chart/").Handler,
			mw.Compress(9, "image/svg+xml"),
		)
		r.Get("/{userWithExt}", api.GetActivityChart)
	})

	// Avatar routes
	r.Route("/avatar", func(r chi.Router) {
		r.Use(mw.Compress(9, "image/svg+xml"))
		r.Get("/{hash}.svg", api.GetAvatarHash)
	})

	// Heartbeat routes
	r.Group(func(r chi.Router) {
		r.Use(
			middlewares.NewAuthenticateMiddleware(api.services.Users()).WithOptionalForMethods(http.MethodOptions).Handler,
			customMiddleware.NewWakatimeRelayMiddleware().Handler,
		)

		if api.config.IsDev() {
			r.Use(customMiddleware.NewWakatimeRelayMiddleware().OtherInstancesHandler)
		}

		// Heartbeat endpoints - consolidated with common handler
		heartbeatRoutes := []string{
			"/api/heartbeat",
			"/api/heartbeats",
			"/api/users/{user}/heartbeats",
			"/api/users/{user}/heartbeats.bulk",
			"/api/v1/users/{user}/heartbeats",
			"/api/v1/users/{user}/heartbeats.bulk",
			"/api/compat/wakatime/v1/users/{user}/heartbeats",
			"/api/compat/wakatime/v1/users/{user}/heartbeats.bulk",
		}

		for _, route := range heartbeatRoutes {
			r.Post(route, api.ProcessHeartBeat)
			r.Options(route, cors.AllowAll().HandlerFunc)
		}
	})

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

	r.Route("/api", func(r chi.Router) {
		r.Use(middlewares.NewAuthenticateMiddleware(api.services.Users()).Handler)
		r.Get("/users/{user}/statusbar/today", api.GetStatusBarRange)
	})

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

		// Authenticated profile endpoints
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
			r.Use(middlewares.NewAuthenticateMiddleware(api.services.Users()).Handler)

			r.Get("/user-agents", api.FetchUserAgents)
			r.Get("/summaries", api.GetSummaries)
			r.Get("/projects", api.GetProjects)
			r.Get("/projects/{id}", api.GetProject)
			r.Get("/durations", api.GetDurations)

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
		})
	})
}
