package middlewares

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/oauth2-proxy/mockoidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	testutils "github.com/muety/wakapi/utils/test"
)

func TestAuthenticateMiddleware_tryGetUserByApiKeyHeader_Success(t *testing.T) {
	testApiKey := "z5uig69cn9ut93n"
	testToken := base64.StdEncoding.EncodeToString([]byte(testApiKey))
	testUser := &models.User{ApiKey: testApiKey}
	// In the case of the API Key from User Model - it's always full access
	testApiKeyRequireFullAccess := false

	mockRequest := &http.Request{
		Header: http.Header{
			"Authorization": []string{fmt.Sprintf("Basic %s", testToken)},
		},
	}

	userServiceMock := new(mocks.UserServiceMock)
	userServiceMock.On("GetUserByKey", testApiKey, testApiKeyRequireFullAccess).Return(testUser, nil)

	sut := NewAuthenticateMiddleware(userServiceMock)

	result, err := sut.tryGetUserByApiKeyHeader(mockRequest)

	assert.Nil(t, err)
	assert.Equal(t, testUser, result)
}

func TestAuthenticateMiddleware_tryGetUserByApiKeyHeader_Invalid(t *testing.T) {
	testApiKey := "z5uig69cn9ut93n"
	testToken := base64.StdEncoding.EncodeToString([]byte(testApiKey))

	mockRequest := &http.Request{
		Header: http.Header{
			// 'Basic' prefix missing here
			"Authorization": []string{fmt.Sprintf("%s", testToken)},
		},
	}

	userServiceMock := new(mocks.UserServiceMock)

	sut := NewAuthenticateMiddleware(userServiceMock)
	sut.WithFullAccessOnly(false)

	result, err := sut.tryGetUserByApiKeyHeader(mockRequest)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAuthenticateMiddleware_tryGetUserByApiKeyHeaderWithReadOnlyKey_Invalid(t *testing.T) {
	testApiKey := "read-only-additional-key"
	testToken := base64.StdEncoding.EncodeToString([]byte(testApiKey))

	mockRequest := &http.Request{
		Header: http.Header{
			"Authorization": []string{fmt.Sprintf("Basic %s", testToken)},
		},
	}

	userServiceMock := new(mocks.UserServiceMock)
	userServiceMock.On("GetUserByKey", testApiKey, true).Return(nil, errors.New("forbidden: requires full access"))

	sut := NewAuthenticateMiddleware(userServiceMock)
	sut.WithFullAccessOnly(true)

	result, err := sut.tryGetUserByApiKeyHeader(mockRequest)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAuthenticateMiddleware_tryGetUserByApiKeyQuery_Success(t *testing.T) {
	testApiKey := "z5uig69cn9ut93n"
	testUser := &models.User{ApiKey: testApiKey}
	// In the case of the API Key from User Model - it's always full access
	testApiKeyRequireFullAccess := true

	params := url.Values{}
	params.Add("api_key", testApiKey)
	mockRequest := &http.Request{
		URL: &url.URL{
			RawQuery: params.Encode(),
		},
	}

	userServiceMock := new(mocks.UserServiceMock)
	userServiceMock.On("GetUserByKey", testApiKey, testApiKeyRequireFullAccess).Return(testUser, nil)

	sut := NewAuthenticateMiddleware(userServiceMock)
	sut.WithFullAccessOnly(true)

	result, err := sut.tryGetUserByApiKeyQuery(mockRequest)

	assert.Nil(t, err)
	assert.Equal(t, testUser, result)
}

func TestAuthenticateMiddleware_tryGetUserByApiKeyQuery_Invalid(t *testing.T) {
	testApiKey := "z5uig69cn9ut93n"

	params := url.Values{}
	params.Add("token", testApiKey)
	mockRequest := &http.Request{
		URL: &url.URL{
			RawQuery: params.Encode(),
		},
	}

	userServiceMock := new(mocks.UserServiceMock)

	sut := NewAuthenticateMiddleware(userServiceMock)

	result, actualErr := sut.tryGetUserByApiKeyQuery(mockRequest)

	assert.Error(t, actualErr)
	assert.Equal(t, errEmptyKey, actualErr)
	assert.Nil(t, result)
}

func TestAuthenticateMiddleware_tryGetUserByTrustedHeader_Disabled(t *testing.T) {
	cfg := config.Empty()
	cfg.Security.TrustedHeaderAuth = false
	cfg.Security.TrustedHeaderAuthKey = "Remote-User"
	cfg.Security.TrustReverseProxyIps = "127.0.0.1,::1"
	cfg.Security.ParseTrustReverseProxyIPs()
	config.Set(cfg)

	testUser := &models.User{ID: "user01"}

	mockRequest := &http.Request{
		Header:     http.Header{"Remote-User": []string{testUser.ID}},
		RemoteAddr: "127.0.0.1:54654",
	}

	userServiceMock := new(mocks.UserServiceMock)
	userServiceMock.On("GetUserById", testUser.ID).Return(testUser, nil)

	sut := NewAuthenticateMiddleware(userServiceMock)

	result, actualErr := sut.tryGetUserByTrustedHeader(mockRequest, false)
	assert.Error(t, actualErr)
	assert.Nil(t, result)
}

func TestAuthenticateMiddleware_tryGetUserByTrustedHeader_Untrusted(t *testing.T) {
	cfg := config.Empty()
	cfg.Security.TrustedHeaderAuth = true
	cfg.Security.TrustedHeaderAuthKey = "Remote-User"
	cfg.Security.TrustReverseProxyIps = "127.0.0.1,::1,192.168.0.1,192.168.178.0/24,33b7:08d8:c07a:c2ee:0fac:cb95:dadc:dafb,1ddc:e2d6:dcce:ab6c::1/64"
	cfg.Security.ParseTrustReverseProxyIPs()
	config.Set(cfg)

	testIps := []string{"192.168.0.2", "192.168.179.35", "[33b7:08d8:c07a:c2ee:0fac:cb95:dadc:dafa]", "[1ddc:e2d6:dcce:ab6b:1ba7:7aaa:58dc:a42b]"} // none of these should be authorized
	testUser := &models.User{ID: "user01"}

	for _, ip := range testIps {
		mockRequest := &http.Request{
			Header: http.Header{
				"Remote-User":     []string{testUser.ID},
				"X-Forwarded-For": []string{"192.168.0.1"}, // forward for some trusted ip -> header should be ignored for auth. checks, because only actual reverse proxy must be legitimized
			},
			RemoteAddr: fmt.Sprintf("%s:54654", ip),
		}

		userServiceMock := new(mocks.UserServiceMock)
		userServiceMock.On("GetUserById", testUser.ID).Return(testUser, nil)

		sut := NewAuthenticateMiddleware(userServiceMock)

		result, actualErr := sut.tryGetUserByTrustedHeader(mockRequest, false)
		assert.Error(t, actualErr)
		assert.Nil(t, result)
	}
}

func TestAuthenticateMiddleware_tryGetUserByTrustedHeader_Success(t *testing.T) {
	cfg := config.Empty()
	cfg.Security.TrustedHeaderAuth = true
	cfg.Security.TrustedHeaderAuthKey = "Remote-User"
	cfg.Security.TrustReverseProxyIps = "127.0.0.1,::1,192.168.0.1,192.168.178.0/24,33b7:08d8:c07a:c2ee:0fac:cb95:dadc:dafb,1ddc:e2d6:dcce:ab6c::1/64"
	cfg.Security.ParseTrustReverseProxyIPs()
	config.Set(cfg)

	testIps := []string{"127.0.0.1", "[::1]", "192.168.0.1", "192.168.178.1", "192.168.178.35", "[33b7:08d8:c07a:c2ee:0fac:cb95:dadc:dafb]", "[1ddc:e2d6:dcce:ab6c:2ba7:7aaa:58dc:a42b]", "[1ddc:e2d6:dcce:ab6c:1ba7:7aaa:58dc:a42b]"} // all of these should be authorized
	testUser := &models.User{ID: "user01"}

	for _, ip := range testIps {
		mockRequest := &http.Request{
			Header:     http.Header{"Remote-User": []string{testUser.ID}},
			RemoteAddr: fmt.Sprintf("%s:54654", ip),
		}

		userServiceMock := new(mocks.UserServiceMock)
		userServiceMock.On("GetUserById", testUser.ID).Return(testUser, nil)

		sut := NewAuthenticateMiddleware(userServiceMock)

		result, actualErr := sut.tryGetUserByTrustedHeader(mockRequest, false)
		assert.Equal(t, testUser, result)
		assert.Nil(t, actualErr)
	}
}

func TestAuthenticateMiddleware_tryGetUserByTrustedHeader_NoSignup(t *testing.T) {
	cfg := config.Empty()
	cfg.Security.TrustedHeaderAuth = true
	cfg.Security.TrustedHeaderAuthAllowSignup = false
	cfg.Security.TrustedHeaderAuthKey = "Remote-User"
	cfg.Security.TrustReverseProxyIps = "127.0.0.1,::1,192.168.0.1,192.168.178.0/24,33b7:08d8:c07a:c2ee:0fac:cb95:dadc:dafb,1ddc:e2d6:dcce:ab6c::1/64"
	cfg.Security.ParseTrustReverseProxyIPs()
	config.Set(cfg)

	testIps := []string{"127.0.0.1", "[::1]", "192.168.0.1", "192.168.178.1", "192.168.178.35", "[33b7:08d8:c07a:c2ee:0fac:cb95:dadc:dafb]", "[1ddc:e2d6:dcce:ab6c:2ba7:7aaa:58dc:a42b]", "[1ddc:e2d6:dcce:ab6c:1ba7:7aaa:58dc:a42b]"} // all of these should be authorized
	testUser := &models.User{ID: "nonexisting"}

	for _, ip := range testIps {
		mockRequest := &http.Request{
			Header:     http.Header{"Remote-User": []string{testUser.ID}},
			RemoteAddr: fmt.Sprintf("%s:54654", ip),
		}

		userServiceMock := new(mocks.UserServiceMock)
		userServiceMock.On("GetUserById", testUser.ID).Return(nil, errors.New("record not found"))

		sut := NewAuthenticateMiddleware(userServiceMock)

		result, actualErr := sut.tryGetUserByTrustedHeader(mockRequest, false)
		assert.Error(t, actualErr)
		assert.Nil(t, result)
		userServiceMock.AssertNumberOfCalls(t, "GetUserById", 1)
	}
}

func TestAuthenticateMiddleware_tryGetUserByTrustedHeader_Signup(t *testing.T) {
	cfg := config.Empty()
	cfg.Security.TrustedHeaderAuth = true
	cfg.Security.TrustedHeaderAuthAllowSignup = true
	cfg.Security.TrustedHeaderAuthKey = "Remote-User"
	cfg.Security.TrustReverseProxyIps = "127.0.0.1,::1,192.168.0.1,192.168.178.0/24,33b7:08d8:c07a:c2ee:0fac:cb95:dadc:dafb,1ddc:e2d6:dcce:ab6c::1/64"
	cfg.Security.ParseTrustReverseProxyIPs()
	config.Set(cfg)

	testIps := []string{"127.0.0.1", "[::1]", "192.168.0.1", "192.168.178.1", "192.168.178.35", "[33b7:08d8:c07a:c2ee:0fac:cb95:dadc:dafb]", "[1ddc:e2d6:dcce:ab6c:2ba7:7aaa:58dc:a42b]", "[1ddc:e2d6:dcce:ab6c:1ba7:7aaa:58dc:a42b]"} // all of these should be authorized
	testUser := &models.User{ID: "tobecreated"}

	for _, ip := range testIps {
		mockRequest := &http.Request{
			Header:     http.Header{"Remote-User": []string{testUser.ID}},
			RemoteAddr: fmt.Sprintf("%s:54654", ip),
		}

		userServiceMock := new(mocks.UserServiceMock)
		userServiceMock.On("GetUserById", testUser.ID).Return(nil, errors.New("record not found"))
		userServiceMock.On("CreateOrGet", mock.Anything, false).Return(testUser, true, nil)
		userServiceMock.On("GetUserById", testUser.ID).Return(testUser, nil)

		sut := NewAuthenticateMiddleware(userServiceMock)

		result, actualErr := sut.tryGetUserByTrustedHeader(mockRequest, true)
		assert.Error(t, actualErr)
		assert.Nil(t, result)
		userServiceMock.AssertNumberOfCalls(t, "GetUserById", 2)
	}
}

func TestAuthenticateMiddleware_tryGetUserByOidc_NoToken(t *testing.T) {
	config.Set(config.Empty())

	userServiceMock := new(mocks.UserServiceMock)

	r := httptest.NewRequest(http.MethodGet, "/summary", nil)
	w := httptest.NewRecorder()

	sut := NewAuthenticateMiddleware(userServiceMock)

	result, err := sut.tryGetUserByOidc(w, r)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAuthenticateMiddleware_tryGetUserByOidc_ValidToken(t *testing.T) {
	const (
		testProvider = "mock"
		testSub      = "testsub"
		testEmail    = "test@example.com"
	)
	var testUser = &models.User{ID: "testuser"}

	oidcMock, _ := mockoidc.Run()
	defer oidcMock.Shutdown()

	cfg := config.Empty()
	config.Set(cfg)
	config.WithOidcProvider(cfg, testProvider, oidcMock.ClientID, oidcMock.ClientSecret, oidcMock.Addr()+"/oidc", "")

	r := httptest.NewRequest(http.MethodGet, "/summary", nil)
	w := httptest.NewRecorder()

	session, err := oidcMock.SessionStore.NewSession(
		"openid profile email",
		"",
		&mockoidc.MockUser{
			Subject:           testSub,
			Email:             testEmail,
			PreferredUsername: testUser.ID,
		},
		"code",
		"method",
	)
	assert.NoError(t, err)

	idToken, err := session.IDToken(oidcMock.Config(), oidcMock.Keypair, time.Now())
	assert.NoError(t, err)

	r.AddCookie(cfg.CreateCookie(config.CookieKeyOidcProvider, testProvider))
	r.AddCookie(cfg.CreateCookie(config.CookieKeyOidcIdToken, idToken))

	userServiceMock := new(mocks.UserServiceMock)
	userServiceMock.On("GetUserByOidc", testProvider, testSub).Return(testUser, nil)

	sut := NewAuthenticateMiddleware(userServiceMock)

	result, err := sut.tryGetUserByOidc(w, r)
	assert.NoError(t, err)
	assert.Equal(t, testUser, result)
}

func TestAuthenticateMiddleware_tryGetUserByOidc_ExpiredTokenNoRefreshToken(t *testing.T) {
	const (
		testProvider = "mock"
		testSub      = "testsub"
		testEmail    = "test@example.com"
	)
	var testUser = &models.User{ID: "testuser"}

	oidcMock, _ := mockoidc.Run()
	defer oidcMock.Shutdown()

	cfg := config.Empty()
	config.Set(cfg)
	config.WithOidcProvider(cfg, testProvider, oidcMock.ClientID, oidcMock.ClientSecret, oidcMock.Addr()+"/oidc", "")

	r := httptest.NewRequest(http.MethodGet, "/summary", nil)
	w := httptest.NewRecorder()

	session, err := oidcMock.SessionStore.NewSession(
		"openid profile email",
		"",
		&mockoidc.MockUser{
			Subject:           testSub,
			Email:             testEmail,
			PreferredUsername: testUser.ID,
		},
		"code",
		"method",
	)
	assert.NoError(t, err)

	oidcMockConfig := oidcMock.Config()
	// Ensure the token is expired
	tokenIssuedTime := time.Now().Add(-oidcMockConfig.AccessTTL * 2)

	idToken, err := session.IDToken(oidcMockConfig, oidcMock.Keypair, tokenIssuedTime)
	assert.NoError(t, err)

	r.AddCookie(cfg.CreateCookie(config.CookieKeyOidcProvider, testProvider))
	r.AddCookie(cfg.CreateCookie(config.CookieKeyOidcIdToken, idToken))

	userServiceMock := new(mocks.UserServiceMock)
	userServiceMock.On("GetUserByOidc", testProvider, testSub).Return(testUser, nil)

	sut := NewAuthenticateMiddleware(userServiceMock)

	_, err = sut.tryGetUserByOidc(w, r)
	assert.Error(t, err)
}

func TestAuthenticateMiddleware_tryGetUserByOidc_ExpiredTokenWithRefreshToken(t *testing.T) {
	const (
		testProvider = "mock"
		testSub      = "testsub"
		testEmail    = "test@example.com"
	)
	var testUser = &models.User{ID: "testuser"}

	oidcMock, _ := mockoidc.Run()
	defer oidcMock.Shutdown()

	cfg := config.Empty()
	config.Set(cfg)
	config.WithOidcProvider(cfg, testProvider, oidcMock.ClientID, oidcMock.ClientSecret, oidcMock.Addr()+"/oidc", "")

	r := httptest.NewRequest(http.MethodGet, "/summary", nil)
	w := httptest.NewRecorder()

	session, err := oidcMock.SessionStore.NewSession(
		"openid profile email",
		"",
		&mockoidc.MockUser{
			Subject:           testSub,
			Email:             testEmail,
			PreferredUsername: testUser.ID,
		},
		"code",
		"method",
	)
	assert.NoError(t, err)

	oidcMockConfig := oidcMock.Config()
	// Ensure the token is expired
	tokenIssuedTime := time.Now().Add(-oidcMockConfig.AccessTTL * 2)

	idToken, err := session.IDToken(oidcMockConfig, oidcMock.Keypair, tokenIssuedTime)
	assert.NoError(t, err)
	refreshToken, err := session.RefreshToken(oidcMockConfig, oidcMock.Keypair, tokenIssuedTime)
	assert.NoError(t, err)

	r.AddCookie(cfg.CreateCookie(config.CookieKeyOidcProvider, testProvider))
	r.AddCookie(cfg.CreateCookie(config.CookieKeyOidcIdToken, idToken))
	r.AddCookie(cfg.CreateCookie(config.CookieKeyOidcRefreshToken, refreshToken))

	userServiceMock := new(mocks.UserServiceMock)
	userServiceMock.On("GetUserByOidc", testProvider, testSub).Return(testUser, nil)

	sut := NewAuthenticateMiddleware(userServiceMock)

	result, err := sut.tryGetUserByOidc(w, r)
	assert.NoError(t, err)
	assert.Equal(t, testUser, result)

	// Check that refresh token and id token are set
	testutils.AssertContainsHeaderMatching(t, w.Header(), "Set-Cookie", func(value string) bool {
		return strings.Contains(value, "oidc_id_token=")
	}, "OIDC id_token cookie not set in response")
	testutils.AssertContainsHeaderMatching(t, w.Header(), "Set-Cookie", func(value string) bool {
		return strings.Contains(value, "oidc_refresh_token=")
	})
}

// TODO: somehow test cookie auth function
