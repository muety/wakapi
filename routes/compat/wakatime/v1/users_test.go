package v1

import (
	"encoding/base64"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var (
	adminUser = &models.User{
		ID:             "AdminUser",
		ApiKey:         "admin-user-api-key",
		Email:          "admin@user.com",
		IsAdmin:        true,
		CreatedAt:      models.CustomTime(time.Date(2022, 2, 2, 22, 22, 22, 1, time.UTC)),
		LastLoggedInAt: models.CustomTime(time.Date(2022, 2, 2, 22, 22, 22, 2, time.UTC)),
	}

	basicUser = &models.User{
		ID:             "BasicUser",
		ApiKey:         "basic-user-api-key",
		Email:          "basic@user.com",
		IsAdmin:        false,
		CreatedAt:      models.CustomTime(time.Date(2022, 2, 2, 22, 22, 22, 3, time.UTC)),
		LastLoggedInAt: models.CustomTime(time.Date(2022, 2, 2, 22, 22, 22, 4, time.UTC)),
	}
)

func TestUsersHandler_Get(t *testing.T) {
	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api").Subrouter().StrictSlash(true)
	apiRouter.Use(middlewares.NewPrincipalMiddleware())

	userServiceMock := new(mocks.UserServiceMock)
	userServiceMock.On("GetUserById", "AdminUser").Return(adminUser, nil)
	userServiceMock.On("GetUserByKey", "admin-user-api-key").Return(adminUser, nil)
	userServiceMock.On("GetUserById", "BasicUser").Return(basicUser, nil)
	userServiceMock.On("GetUserByKey", "basic-user-api-key").Return(basicUser, nil)

	heartbeatServiceMock := new(mocks.HeartbeatServiceMock)
	heartbeatServiceMock.On("GetLatestByUser", adminUser).Return(&models.Heartbeat{
		CreatedAt: models.CustomTime(time.Date(2022, 2, 2, 22, 22, 22, 5, time.UTC)),
		Time:      models.CustomTime(time.Date(2022, 2, 2, 22, 22, 22, 6, time.UTC)),
	}, nil)
	heartbeatServiceMock.On("GetLatestByUser", basicUser).Return(&models.Heartbeat{
		CreatedAt: models.CustomTime(time.Date(2022, 2, 2, 22, 22, 22, 5, time.UTC)),
		Time:      models.CustomTime(time.Date(2022, 2, 2, 22, 22, 22, 6, time.UTC)),
	}, nil)

	usersHandler := NewUsersHandler(userServiceMock, heartbeatServiceMock)
	usersHandler.RegisterRoutes(apiRouter)

	t.Run("when requesting own user data", func(t *testing.T) {
		t.Run("should return own data", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/compat/wakatime/v1/users/AdminUser", nil)
			req.Header.Add(
				"Authorization",
				fmt.Sprintf("Bearer %s", base64.StdEncoding.EncodeToString([]byte(adminUser.ApiKey))),
			)
			requestRecorder := httptest.NewRecorder()
			apiRouter.ServeHTTP(requestRecorder, req)
			res := requestRecorder.Result()
			defer res.Body.Close()

			data, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Errorf("unextected error. Error: %s", err)
			}

			if !strings.Contains(string(data), "\"username\":\"AdminUser\"") {
				t.Errorf("invalid response received. Expected json Received: %s", string(data))
			}
		})
	})

	t.Run("when requesting another users data", func(t *testing.T) {
		t.Run("should respond with '401 unauthorized' if not an admin user", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/compat/wakatime/v1/users/AdminUser", nil)
			req.Header.Add(
				"Authorization",
				fmt.Sprintf("Bearer %s", base64.StdEncoding.EncodeToString([]byte(basicUser.ApiKey))),
			)
			requestRecorder := httptest.NewRecorder()
			apiRouter.ServeHTTP(requestRecorder, req)
			res := requestRecorder.Result()
			defer res.Body.Close()

			data, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Errorf("unextected error. Error: %s", err)
			}

			if string(data) != "401 unauthorized" {
				t.Errorf("invalid response received. Expected: '401 unauthorized' Received: %s", string(data))
			}
		})

		t.Run("should receive user data if requesting user is an admin", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/compat/wakatime/v1/users/BasicUser", nil)
			req.Header.Add(
				"Authorization",
				fmt.Sprintf("Bearer %s", base64.StdEncoding.EncodeToString([]byte(adminUser.ApiKey))),
			)
			requestRecorder := httptest.NewRecorder()
			apiRouter.ServeHTTP(requestRecorder, req)
			res := requestRecorder.Result()
			defer res.Body.Close()

			data, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Errorf("unextected error. Error: %s", err)
			}

			if !strings.Contains(string(data), "\"username\":\"BasicUser\"") {
				t.Errorf("invalid response received. Expected 'BasicUser' info Received: %s", string(data))
			}
		})
	})
}
