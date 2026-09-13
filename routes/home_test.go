package routes

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/securecookie"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	routeutils "github.com/muety/wakapi/routes/utils"
	"github.com/stretchr/testify/assert"
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
				t.Errorf("unexpected error. Error: %s", err)
			}

			assert.Contains(t, string(data), "<a href=\"login\" class=\"btn-primary\">")
			keyValueServiceMock.AssertNumberOfCalls(t, "GetString", 3)
		})
	})
}

func TestHomeHandler_Get_LoggedIn(t *testing.T) {
	cfg := config.Empty()
	cfg.Security.CookieKeyBytes = securecookie.GenerateRandomKey(128)
	config.Set(cfg)
	config.InitializeCookies()

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
		t.Run("should redirect in case of cookie auth", func(t *testing.T) {
			authCookie, err := routeutils.CreateAuthCookie(user1.ID)
			assert.NoError(t, err)

			rec := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.AddCookie(authCookie)

			router.ServeHTTP(rec, req)
			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, http.StatusFound, res.StatusCode)
			assert.Equal(t, "/summary", res.Header.Get("Location"))
		})

		t.Run("should not authenticate via api key", func(t *testing.T) {
			rec := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			q := req.URL.Query()
			q.Set("api_key", user1.ApiKey)
			req.URL.RawQuery = q.Encode()

			router.ServeHTTP(rec, req)
			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, http.StatusOK, res.StatusCode)

			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("unexpected error. Error: %s", err)
			}

			assert.Contains(t, string(data), "<a href=\"login\" class=\"btn-primary\">")
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
