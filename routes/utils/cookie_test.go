package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/stretchr/testify/assert"

	"github.com/muety/wakapi/config"
)

func setupCookieTestConfig(t *testing.T, maxAgeSec int) *config.Config {
	cfg := config.Empty()
	cfg.Security.CookieKeyBytes = securecookie.GenerateRandomKey(128)
	cfg.Security.CookieMaxAgeSec = maxAgeSec
	config.Set(cfg)
	config.InitializeCookies()
	return cfg
}

func TestGetAuthCookieDuration(t *testing.T) {
	t.Run("default timeout when CookieMaxAgeSec is zero", func(t *testing.T) {
		setupCookieTestConfig(t, 0)
		assert.Equal(t, DefaultSessionTimeout, GetAuthCookieDuration())
	})

	t.Run("configured timeout when CookieMaxAgeSec is positive", func(t *testing.T) {
		setupCookieTestConfig(t, 3600)
		assert.Equal(t, 3600*time.Second, GetAuthCookieDuration())
	})
}

func TestCreateAuthCookie(t *testing.T) {
	setupCookieTestConfig(t, 1800)

	cookie, err := CreateAuthCookie("testuser")
	assert.NoError(t, err)
	assert.NotNil(t, cookie)
	assert.Equal(t, config.CookieKeyAuth, cookie.Name)
	assert.NotEmpty(t, cookie.Value)
	assert.Equal(t, 1800, cookie.MaxAge)
	assert.True(t, cookie.Expires.After(time.Now()))
}

func TestExtractCookieAuthData_Valid(t *testing.T) {
	setupCookieTestConfig(t, 3600)

	cookie, err := CreateAuthCookie("john_doe")
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)

	data, err := ExtractCookieAuthData(req)
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, "john_doe", data.Username)
	assert.True(t, data.Expiry > time.Now().Unix())

	username, err := ExtractCookieAuth(req)
	assert.NoError(t, err)
	assert.NotNil(t, username)
	assert.Equal(t, "john_doe", *username)
}

func TestExtractCookieAuthData_Missing(t *testing.T) {
	setupCookieTestConfig(t, 3600)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	data, err := ExtractCookieAuthData(req)
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "missing authentication")
}

func TestExtractCookieAuthData_InvalidValue(t *testing.T) {
	setupCookieTestConfig(t, 3600)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  config.CookieKeyAuth,
		Value: "invalid-garbage-value",
	})

	data, err := ExtractCookieAuthData(req)
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "cookie is invalid")
}

func TestExtractCookieAuthData_Expired(t *testing.T) {
	setupCookieTestConfig(t, 3600)

	cookie, err := CreateAuthCookieWithExpiry("john_doe", time.Now().Add(-10*time.Minute))
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)

	data, err := ExtractCookieAuthData(req)
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "cookie is expired")
}

func TestShouldRefreshCookie(t *testing.T) {
	setupCookieTestConfig(t, 7200) // 2 hours TTL, half threshold is 1 hour

	t.Run("fresh cookie with more than half lifetime remaining should not refresh", func(t *testing.T) {
		// 1.5 hours remaining > 1 hour threshold
		expiry := time.Now().Add(90 * time.Minute).Unix()
		assert.False(t, ShouldRefreshCookie(expiry))
	})

	t.Run("cookie at exactly half threshold should refresh", func(t *testing.T) {
		// 1 hour remaining == 1 hour threshold
		expiry := time.Now().Add(60 * time.Minute).Unix()
		assert.True(t, ShouldRefreshCookie(expiry))
	})

	t.Run("cookie with less than half lifetime remaining should refresh", func(t *testing.T) {
		// 30 minutes remaining < 1 hour threshold
		expiry := time.Now().Add(30 * time.Minute).Unix()
		assert.True(t, ShouldRefreshCookie(expiry))
	})
}

func TestAuthenticateCookie(t *testing.T) {
	setupCookieTestConfig(t, 3600) // 1 hour TTL, threshold is 30 mins

	t.Run("valid session young does not set cookie", func(t *testing.T) {
		cookie, err := CreateAuthCookieWithExpiry("alice", time.Now().Add(50*time.Minute))
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(cookie)
		w := httptest.NewRecorder()

		username, err := AuthenticateCookie(w, req)
		assert.NoError(t, err)
		assert.NotNil(t, username)
		assert.Equal(t, "alice", *username)
		assert.Empty(t, w.Header().Values("Set-Cookie"))
	})

	t.Run("valid session past threshold refreshes auth cookie", func(t *testing.T) {
		cookie, err := CreateAuthCookieWithExpiry("alice", time.Now().Add(20*time.Minute))
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(cookie)
		w := httptest.NewRecorder()

		username, err := AuthenticateCookie(w, req)
		assert.NoError(t, err)
		assert.NotNil(t, username)
		assert.Equal(t, "alice", *username)

		cookies := w.Result().Cookies()
		var authCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == config.CookieKeyAuth {
				authCookie = c
				break
			}
		}
		assert.NotNil(t, authCookie)
		assert.Equal(t, 3600, authCookie.MaxAge)

		// Decode refreshed cookie to verify renewed expiry
		var refreshedData CookieKeyData
		err = config.GetAuthCookie().Decode(config.CookieKeyAuth, authCookie.Value, &refreshedData)
		assert.NoError(t, err)
		assert.Equal(t, "alice", refreshedData.Username)
		assert.True(t, time.Unix(refreshedData.Expiry, 0).After(time.Now().Add(50*time.Minute)))
	})

	t.Run("valid session past threshold refreshes both auth and session cookie when present", func(t *testing.T) {
		cookie, err := CreateAuthCookieWithExpiry("alice", time.Now().Add(20*time.Minute))
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(cookie)

		// Create an existing session cookie
		session, err := config.GetSessionStore().Get(req, config.CookieKeySession)
		assert.NoError(t, err)
		session.Values["test_key"] = "test_val"
		wInit := httptest.NewRecorder()
		err = session.Save(req, wInit)
		assert.NoError(t, err)

		var sessionCookie *http.Cookie
		for _, c := range wInit.Result().Cookies() {
			if c.Name == config.CookieKeySession {
				sessionCookie = c
				break
			}
		}
		assert.NotNil(t, sessionCookie)
		req.AddCookie(sessionCookie)

		w := httptest.NewRecorder()
		username, err := AuthenticateCookie(w, req)
		assert.NoError(t, err)
		assert.Equal(t, "alice", *username)

		cookies := w.Result().Cookies()
		var hasAuth, hasSession bool
		for _, c := range cookies {
			if c.Name == config.CookieKeyAuth {
				hasAuth = true
			}
			if c.Name == config.CookieKeySession {
				hasSession = true
			}
		}
		assert.True(t, hasAuth)
		assert.True(t, hasSession)
	})
}

func TestRefreshSession(t *testing.T) {
	setupCookieTestConfig(t, 3600)

	t.Run("nil parameters return nil safely", func(t *testing.T) {
		assert.NoError(t, RefreshSession(nil, nil))
	})

	t.Run("no session cookie does not write Set-Cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		assert.NoError(t, RefreshSession(req, w))
		assert.Empty(t, w.Result().Cookies())
	})

	t.Run("existing session cookie is refreshed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		session, err := config.GetSessionStore().Get(req, config.CookieKeySession)
		assert.NoError(t, err)
		session.Values["data"] = "123"
		wInit := httptest.NewRecorder()
		assert.NoError(t, session.Save(req, wInit))

		var sessCookie *http.Cookie
		for _, c := range wInit.Result().Cookies() {
			if c.Name == config.CookieKeySession {
				sessCookie = c
				break
			}
		}
		assert.NotNil(t, sessCookie)

		req2 := httptest.NewRequest(http.MethodGet, "/", nil)
		req2.AddCookie(sessCookie)
		w2 := httptest.NewRecorder()

		assert.NoError(t, RefreshSession(req2, w2))
		cookies := w2.Result().Cookies()
		assert.Len(t, cookies, 1)
		assert.Equal(t, config.CookieKeySession, cookies[0].Name)
	})
}
