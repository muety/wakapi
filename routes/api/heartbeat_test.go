package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func Test_fillPlaceholders(t *testing.T) {
	heartbeatServiceMock := new(mocks.HeartbeatServiceMock)
	heartbeatServiceMock.On("GetLatestByUser", mock.Anything).Return(&models.Heartbeat{
		Project: "project1",
	}, nil)

	heartbeatServiceMock.On("GetLatestByFilters", mock.Anything, mock.Anything).Return(&models.Heartbeat{
		Project:  "must not be used",
		Branch:   "replaced2",
		Language: "replaced3",
	}, nil)

	t.Run("when filling placeholders", func(t *testing.T) {
		t.Run("should replace project, language and branch properly", func(t *testing.T) {
			hb := &models.Heartbeat{
				Project:  "<<LAST_PROJECT>>",
				Branch:   "<<LAST_BRANCH>>",
				Language: "<<LAST_LANGUAGE>>",
			}
			hb = fillPlaceholders(hb, &models.User{}, heartbeatServiceMock)

			filters1 := heartbeatServiceMock.Calls[1].Arguments.Get(1).(*models.Filters)
			filters2 := heartbeatServiceMock.Calls[2].Arguments.Get(1).(*models.Filters)

			assert.Equal(t, len(heartbeatServiceMock.Calls), 3)
			assert.Equal(t, "project1", filters1.Project[0])
			assert.Equal(t, "project1", filters2.Project[0])
			assert.Equal(t, "project1", hb.Project)
			assert.Equal(t, "replaced2", hb.Branch)
			assert.Equal(t, "replaced3", hb.Language)
		})

		t.Run("should replace nothing if no placeholders given", func(t *testing.T) {
			hb := &models.Heartbeat{
				Project:  "project2",
				Branch:   "branch2",
				Language: "language2",
			}
			hb = fillPlaceholders(hb, &models.User{}, heartbeatServiceMock)
			assert.Equal(t, "project2", hb.Project)
			assert.Equal(t, "branch2", hb.Branch)
			assert.Equal(t, "language2", hb.Language)
		})

		t.Run("should clear placeholders without replacement for browsing heartbeats", func(t *testing.T) {
			hb := &models.Heartbeat{
				Project:  "<<LAST_PROJECT>>",
				Branch:   "<<LAST_BRANCH>>",
				Language: "<<LAST_LANGUAGE>>",
				Type:     "url",
			}
			hb = fillPlaceholders(hb, &models.User{}, heartbeatServiceMock)
			assert.Empty(t, hb.Project)
			assert.Empty(t, hb.Branch)
			assert.Empty(t, hb.Language)
		})
	})
}
