package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

var (
	user1 = models.User{
		ID:               "user1",
		ShareDataMaxDays: 30,
		ShareLanguages:   true,
		ApiKey:           "fakekey",
	}
)

func TestHomeHandler_Get_NotLoggedIn(t *testing.T) {
	config.Set(config.Empty())
	config.Get().Env = "dev"

	if cwd, _ := os.Getwd(); strings.HasSuffix(cwd, "routes") {
		os.Chdir("..")
	}

	router := chi.NewRouter()
	router.Use(middlewares.NewSharedDataMiddleware())

	userServiceMock := new(mocks.UserServiceMock)
	userServiceMock.On("GetUserById", user1.ID).Return(&user1, nil)
	userServiceMock.On("CountCurrentlyOnline").Return(0, nil)

	keyValueServiceMock := new(mocks.KeyValueServiceMock)
	keyValueServiceMock.On("GetString", config.KeyLatestTotalTime).Return(&models.KeyStringValue{Key: config.KeyLatestTotalTime, Value: "0"}, nil)
	keyValueServiceMock.On("GetString", config.KeyLatestTotalUsers).Return(&models.KeyStringValue{Key: config.KeyLatestTotalUsers, Value: "0"}, nil)
	keyValueServiceMock.On("GetString", config.KeyNewsbox).Return(&models.KeyStringValue{Key: config.KeyNewsbox, Value: ""}, nil)

	homeHandler := NewHomeHandler(userServiceMock, keyValueServiceMock)
	homeHandler.RegisterRoutes(router)

	t.Run("when requesting frontpage", func(t *testing.T) {
		t.Run("should display it without authentication", func(t *testing.T) {
			rec := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, "/", nil)

			router.ServeHTTP(rec, req)
			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, http.StatusOK, res.StatusCode)

			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("unextected error. Error: %s", err)
			}

			assert.Contains(t, string(data), "<a href=\"login\" class=\"btn-primary\">")
			keyValueServiceMock.AssertNumberOfCalls(t, "GetString", 3)
		})
	})
}

func TestHomeHandler_Get_LoggedIn(t *testing.T) {
	config.Set(config.Empty())

	router := chi.NewRouter()
	router.Use(middlewares.NewSharedDataMiddleware())

	userServiceMock := new(mocks.UserServiceMock)
	userServiceMock.On("GetUserByKey", user1.ApiKey, false).Return(&user1, nil)
	userServiceMock.On("GetUserById", user1.ID).Return(&user1, nil)

	keyValueServiceMock := new(mocks.KeyValueServiceMock)

	homeHandler := NewHomeHandler(userServiceMock, keyValueServiceMock)
	homeHandler.RegisterRoutes(router)

	t.Run("when requesting frontpage", func(t *testing.T) {
		t.Run("should redirect in case of api key auth", func(t *testing.T) {
			rec := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			q := req.URL.Query()
			q.Set("api_key", user1.ApiKey)
			req.URL.RawQuery = q.Encode()

			router.ServeHTTP(rec, req)
			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, http.StatusFound, res.StatusCode)
		})

		t.Run("should redirect in case of trusted header auth", func(t *testing.T) {
			c := config.Get()
			c.Security.TrustedHeaderAuth = true
			c.Security.TrustedHeaderAuthKey = "Remote-User"
			c.Security.TrustReverseProxyIps = "127.0.0.1"
			c.Security.ParseTrustReverseProxyIPs()

			rec := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Remote-User", user1.ID)
			req.RemoteAddr = "127.0.0.1:12345"

			router.ServeHTTP(rec, req)
			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, http.StatusFound, res.StatusCode)
		})
	})
}
