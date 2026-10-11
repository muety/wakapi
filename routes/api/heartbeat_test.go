package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	v1 "github.com/muety/wakapi/models/compat/wakatime/v1"
	routeutils "github.com/muety/wakapi/routes/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHeartbeatHandler_Options(t *testing.T) {
	config.Set(config.Empty())

	router := chi.NewRouter()
	apiRouter := chi.NewRouter()
	apiRouter.Use(middlewares.NewSharedDataMiddleware())
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

func TestHeartbeatHandler_Post_Timeliness(t *testing.T) {
	cfg := config.Empty()
	cfg.App.HeartbeatMaxAge = "168h" // 7 days
	config.Set(cfg)

	userServiceMock := new(mocks.UserServiceMock)
	heartbeatServiceMock := new(mocks.HeartbeatServiceMock)
	handler := NewHeartbeatApiHandler(userServiceMock, heartbeatServiceMock, nil)

	user := &models.User{ID: "testuser", HasData: true}

	t.Run("should only insert timely heartbeats", func(t *testing.T) {
		now := time.Now()
		timelyTime := now.Unix()
		untimelyTime := now.Add(-200 * time.Hour).Unix()

		body := fmt.Sprintf(`[{"entity": "timely.go", "time": %d}, {"entity": "untimely.go", "time": %d}]`, timelyTime, untimelyTime)
		req := httptest.NewRequest(http.MethodPost, "/heartbeats", bytes.NewBufferString(body))

		sharedData := config.NewSharedData()
		ctx := context.WithValue(req.Context(), config.KeySharedData, sharedData)
		req = req.WithContext(ctx)
		routeutils.SetPrincipal(req, user)

		rec := httptest.NewRecorder()

		// Expectation: only 1 heartbeat should be passed to InsertBatch
		heartbeatServiceMock.On("InsertBatch", mock.MatchedBy(func(hbs []*models.Heartbeat) bool {
			return len(hbs) == 1 && hbs[0].Entity == "timely.go"
		})).Return(nil)

		handler.Post(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var response v1.HeartbeatResponseViewModel
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(response.Responses))

		// Check first heartbeat (timely) - should be 201
		assert.Equal(t, float64(http.StatusCreated), response.Responses[0][1])
		assert.Nil(t, response.Responses[0][0].(map[string]interface{})["error"])

		// Check second heartbeat (untimely) - should be 400
		assert.Equal(t, float64(http.StatusBadRequest), response.Responses[1][1])
		assert.Equal(t, "invalid heartbeat object", response.Responses[1][0].(map[string]interface{})["error"])

		heartbeatServiceMock.AssertExpectations(t)
	})

	t.Run("should parse ai_model from payload and user-agent", func(t *testing.T) {
		now := time.Now().Unix()

		body := fmt.Sprintf(`[{"entity": "file1.go", "time": %d, "ai_model": "claude-3-5-sonnet"}, {"entity": "file2.go", "time": %d, "user_agent": "wakatime/v1.0 (linux-6.6.0-x86_64) go1.21 opus/4.1-medium claude-code/2.1.45"}]`, now, now)
		req := httptest.NewRequest(http.MethodPost, "/heartbeats", bytes.NewBufferString(body))

		sharedData := config.NewSharedData()
		ctx := context.WithValue(req.Context(), config.KeySharedData, sharedData)
		req = req.WithContext(ctx)
		routeutils.SetPrincipal(req, user)

		rec := httptest.NewRecorder()

		heartbeatServiceMock.On("InsertBatch", mock.MatchedBy(func(hbs []*models.Heartbeat) bool {
			return len(hbs) == 2 && hbs[0].AIModel == "claude-3-5-sonnet" && hbs[1].AIModel == "opus" && hbs[1].Editor == "Claude"
		})).Return(nil)

		handler.Post(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		heartbeatServiceMock.AssertExpectations(t)
	})
}

func TestHeartbeatHandler_Post_Placeholders(t *testing.T) {
	cfg := config.Empty()
	cfg.App.HeartbeatMaxAge = "168h" // 7 days
	config.Set(cfg)

	userServiceMock := new(mocks.UserServiceMock)
	heartbeatServiceMock := new(mocks.HeartbeatServiceMock)
	handler := NewHeartbeatApiHandler(userServiceMock, heartbeatServiceMock, nil)

	user := &models.User{ID: "testuser", HasData: true}

	t.Run("should resolve placeholders once per distinct project per request", func(t *testing.T) {
		now := time.Now().Unix()
		body := fmt.Sprintf(`[{"entity": "file1.go", "time": %d, "project": "p1", "branch": "<<LAST_BRANCH>>", "language": "<<LAST_LANGUAGE>>"}, {"entity": "file2.go", "time": %d, "project": "p1", "branch": "<<LAST_BRANCH>>"}, {"entity": "file3.go", "time": %d, "project": "p2", "branch": "<<LAST_BRANCH>>"}]`, now, now, now)
		req := httptest.NewRequest(http.MethodPost, "/heartbeats", bytes.NewBufferString(body))

		sharedData := config.NewSharedData()
		ctx := context.WithValue(req.Context(), config.KeySharedData, sharedData)
		req = req.WithContext(ctx)
		routeutils.SetPrincipal(req, user)

		rec := httptest.NewRecorder()

		heartbeatServiceMock.On("GetLatestByFilters", mock.Anything, mock.Anything, false).Return(&models.Heartbeat{Branch: "last-branch", Language: "last-language"}, nil).Times(2)
		heartbeatServiceMock.On("InsertBatch", mock.MatchedBy(func(hbs []*models.Heartbeat) bool {
			return len(hbs) == 3 &&
				hbs[0].Branch == "last-branch" && hbs[0].Language == "last-language" &&
				hbs[1].Branch == "last-branch" && hbs[1].Language == "" &&
				hbs[2].Branch == "last-branch"
		})).Return(nil)

		handler.Post(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		heartbeatServiceMock.AssertExpectations(t)
	})
}

func Test_fillPlaceholders(t *testing.T) {
	newMock := func() *mocks.HeartbeatServiceMock {
		m := new(mocks.HeartbeatServiceMock)
		m.On("GetLatestByUser", mock.Anything, mock.Anything).Return(&models.Heartbeat{
			Project:  "project1",
			Branch:   "replaced2_by_project",
			Language: "replaced3_by_project",
		}, nil)
		m.On("GetLatestByFilters", mock.Anything, mock.Anything, mock.Anything).Return(&models.Heartbeat{
			Project:  "must not be used",
			Branch:   "replaced2",
			Language: "replaced3",
		}, nil)
		return m
	}

	t.Run("when filling placeholders", func(t *testing.T) {
		t.Run("should replace project, language and branch properly", func(t *testing.T) {
			srv := newMock()
			hb := &models.Heartbeat{
				Project:  "project9",
				Branch:   "<<LAST_BRANCH>>",
				Language: "<<LAST_LANGUAGE>>",
			}
			hb = fillPlaceholders(hb, newPlaceholderCache(&models.User{}, srv))

			assert.Len(t, srv.Calls, 1)
			filters := srv.Calls[0].Arguments.Get(1).(*models.Filters)

			assert.Equal(t, "project9", filters.Project[0])
			assert.Equal(t, "project9", hb.Project)
			assert.Equal(t, "replaced2", hb.Branch)
			assert.Equal(t, "replaced3", hb.Language)
		})

		t.Run("should replace project, language and branch properly with a single lookup", func(t *testing.T) {
			srv := newMock()
			hb := &models.Heartbeat{
				Project:  "<<LAST_PROJECT>>",
				Branch:   "<<LAST_BRANCH>>",
				Language: "<<LAST_LANGUAGE>>",
			}
			hb = fillPlaceholders(hb, newPlaceholderCache(&models.User{}, srv))

			assert.Len(t, srv.Calls, 1) // only one call because project-resolved heartbeat is reused

			assert.Equal(t, "project1", hb.Project)
			assert.Equal(t, "replaced2_by_project", hb.Branch)
			assert.Equal(t, "replaced3_by_project", hb.Language)
		})

		t.Run("should replace nothing if no placeholders given", func(t *testing.T) {
			srv := newMock()
			hb := &models.Heartbeat{
				Project:  "project2",
				Branch:   "branch2",
				Language: "language2",
			}
			hb = fillPlaceholders(hb, newPlaceholderCache(&models.User{}, srv))
			assert.Equal(t, "project2", hb.Project)
			assert.Equal(t, "branch2", hb.Branch)
			assert.Equal(t, "language2", hb.Language)
			assert.Empty(t, srv.Calls)
		})

		t.Run("should clear placeholders without replacement for browsing heartbeats", func(t *testing.T) {
			srv := newMock()
			hb := &models.Heartbeat{
				Project:  "<<LAST_PROJECT>>",
				Branch:   "<<LAST_BRANCH>>",
				Language: "<<LAST_LANGUAGE>>",
				Type:     "url",
			}
			hb = fillPlaceholders(hb, newPlaceholderCache(&models.User{}, srv))
			assert.Empty(t, hb.Project)
			assert.Empty(t, hb.Branch)
			assert.Empty(t, hb.Language)
		})

		t.Run("should dedupe language and branch lookups within a single heartbeat", func(t *testing.T) {
			srv := newMock()
			hb := &models.Heartbeat{
				Project:  "project3",
				Branch:   "<<LAST_BRANCH>>",
				Language: "<<LAST_LANGUAGE>>",
			}
			hb = fillPlaceholders(hb, newPlaceholderCache(&models.User{}, srv))
			assert.Len(t, srv.Calls, 1)
			assert.Equal(t, "project3", srv.Calls[0].Arguments.Get(1).(*models.Filters).Project[0])
			assert.Equal(t, "replaced2", hb.Branch)
			assert.Equal(t, "replaced3", hb.Language)
		})

		t.Run("should dedupe lookups across heartbeats sharing a cache", func(t *testing.T) {
			srv := newMock()
			cache := newPlaceholderCache(&models.User{}, srv)

			hb1 := fillPlaceholders(&models.Heartbeat{Project: "project3", Branch: "<<LAST_BRANCH>>", Language: "<<LAST_LANGUAGE>>"}, cache)
			hb2 := fillPlaceholders(&models.Heartbeat{Project: "project3", Branch: "<<LAST_BRANCH>>"}, cache)
			hb3 := fillPlaceholders(&models.Heartbeat{Project: "project4", Branch: "<<LAST_BRANCH>>"}, cache)

			assert.Len(t, srv.Calls, 2)
			assert.Equal(t, "project3", srv.Calls[0].Arguments.Get(1).(*models.Filters).Project[0])
			assert.Equal(t, "project4", srv.Calls[1].Arguments.Get(1).(*models.Filters).Project[0])
			assert.Equal(t, "replaced2", hb1.Branch)
			assert.Equal(t, "replaced2", hb2.Branch)
			assert.Equal(t, "replaced2", hb3.Branch)
			assert.Equal(t, "replaced3", hb1.Language)
		})

		t.Run("should dedupe latest-by-user lookups across heartbeats", func(t *testing.T) {
			srv := newMock()
			cache := newPlaceholderCache(&models.User{}, srv)

			hb1 := fillPlaceholders(&models.Heartbeat{Project: "<<LAST_PROJECT>>"}, cache)
			hb2 := fillPlaceholders(&models.Heartbeat{Project: "<<LAST_PROJECT>>"}, cache)

			assert.Len(t, srv.Calls, 1)
			assert.Equal(t, "project1", hb1.Project)
			assert.Equal(t, "project1", hb2.Project)
		})
	})
}
