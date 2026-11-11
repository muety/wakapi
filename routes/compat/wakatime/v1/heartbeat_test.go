package v1

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHeartbeatHandler_Get_AICachedInputTokens(t *testing.T) {
	config.Set(config.Empty())

	router := chi.NewRouter()
	apiRouter := chi.NewRouter()
	apiRouter.Use(middlewares.NewSharedDataMiddleware())
	router.Mount("/api", apiRouter)

	testUser := &models.User{
		ID:        "TestUser",
		ApiKey:    "test-user-api-key",
		CreatedAt: models.CustomTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
	}

	userServiceMock := new(mocks.UserServiceMock)
	userServiceMock.On("GetUserById", "TestUser").Return(testUser, nil)
	userServiceMock.On("GetUserByKey", "test-user-api-key", false).Return(testUser, nil)

	heartbeatServiceMock := new(mocks.HeartbeatServiceMock)
	heartbeats := []*models.Heartbeat{
		{
			ID:                  101,
			UserID:              "TestUser",
			Entity:              "server.go",
			Project:             "wakapi",
			Time:                models.CustomTime(time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)),
			CreatedAt:           models.CustomTime(time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)),
			AIInputTokens:       250,
			AIOutputTokens:      125,
			AICachedInputTokens: 8500,
		},
	}
	heartbeatServiceMock.On("GetAllWithin", mock.Anything, mock.Anything, testUser).Return(heartbeats, nil)

	handler := NewHeartbeatHandler(userServiceMock, heartbeatServiceMock)
	handler.RegisterRoutes(apiRouter)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/compat/wakatime/v1/users/{user}/heartbeats?date=2026-10-15", nil)
	req = withUrlParam(req, "user", "current")
	req.Header.Add(
		"Authorization",
		fmt.Sprintf("Bearer %s", base64.StdEncoding.EncodeToString([]byte(testUser.ApiKey))),
	)

	router.ServeHTTP(rec, req)
	res := rec.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)
	data, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	bodyStr := string(data)
	assert.Contains(t, bodyStr, `"ai_cached_input_tokens":8500`)
	assert.Contains(t, bodyStr, `"ai_input_tokens":250`)
	assert.Contains(t, bodyStr, `"ai_output_tokens":125`)
}
