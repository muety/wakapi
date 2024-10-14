package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/mocks"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHeartbeatHandler_Options(t *testing.T) {
	config.Set(config.Empty())

	router := chi.NewRouter()
	apiRouter := chi.NewRouter()
	apiRouter.Use(middlewares.NewPrincipalMiddleware())
	router.Mount("/api", apiRouter)

	userServiceMock := new(mocks.UserServiceMock)
	heartbeatServiceMock := new(mocks.HeartbeatServiceMock)

	heartbeatHandler := NewHeartbeatApiHandler(userServiceMock, heartbeatServiceMock, nil)
	heartbeatHandler.RegisterRoutes(apiRouter)

	t.Run("when receiving cors preflight request", func(t *testing.T) {
		t.Run("should respond with anything allowed", func(t *testing.T) {
			rec := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodOptions, "/api/compat/wakatime/v1/users/current/heartbeats.bulk", nil)
			req.Header.Add("Access-Control-Request-Method", "POST")
			req.Header.Add("Origin", "https://wakapi.dev")

			router.ServeHTTP(rec, req)
			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, http.StatusNoContent, res.StatusCode)
			assert.Equal(t, "*", res.Header.Get("Access-Control-Allow-Origin"))
			assert.Equal(t, "POST", res.Header.Get("Access-Control-Allow-Methods"))
		})
	})
}
