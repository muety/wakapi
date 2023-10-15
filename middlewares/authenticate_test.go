package middlewares

import (
	"encoding/base64"
	"fmt"
	"github.com/muety/wakapi/config"
	"net/http"
	"net/url"
	"testing"

	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticateMiddleware_tryGetUserByApiKeyHeader_Success(t *testing.T) {
	testApiKey := "z5uig69cn9ut93n"
	testToken := base64.StdEncoding.EncodeToString([]byte(testApiKey))
	testUser := &models.User{ApiKey: testApiKey}

	mockRequest := &http.Request{
		Header: http.Header{
			"Authorization": []string{fmt.Sprintf("Basic %s", testToken)},
		},
	}

	userServiceMock := new(mocks.UserServiceMock)
	userServiceMock.On("GetUserByKey", testApiKey).Return(testUser, nil)

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

	result, err := sut.tryGetUserByApiKeyHeader(mockRequest)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAuthenticateMiddleware_tryGetUserByApiKeyQuery_Success(t *testing.T) {
	testApiKey := "z5uig69cn9ut93n"
	testUser := &models.User{ApiKey: testApiKey}

	params := url.Values{}
	params.Add("api_key", testApiKey)
	mockRequest := &http.Request{
		URL: &url.URL{
			RawQuery: params.Encode(),
		},
	}

	userServiceMock := new(mocks.UserServiceMock)
	userServiceMock.On("GetUserByKey", testApiKey).Return(testUser, nil)

	sut := NewAuthenticateMiddleware(userServiceMock)

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
		Header: http.Header{"Remote-User": []string{testUser.ID}},
	}

	userServiceMock := new(mocks.UserServiceMock)
	userServiceMock.On("GetUserById", testUser.ID).Return(testUser, nil)

	sut := NewAuthenticateMiddleware(userServiceMock)

	result, actualErr := sut.tryGetUserByTrustedHeader(mockRequest)
	assert.Error(t, actualErr)
	assert.Nil(t, result)
}

func TestAuthenticateMiddleware_tryGetUserByTrustedHeader_Untrusted(t *testing.T) {
	cfg := config.Empty()
	cfg.Security.TrustedHeaderAuth = true
	cfg.Security.TrustedHeaderAuthKey = "Remote-User"
	cfg.Security.TrustReverseProxyIps = "192.168.0.1"
	cfg.Security.ParseTrustReverseProxyIPs()
	config.Set(cfg)

	testUser := &models.User{ID: "user01"}

	mockRequest := &http.Request{
		Header: http.Header{
			"Remote-User":     []string{testUser.ID},
			"X-Forwarded-For": []string{"192.168.0.1"},
		},
		RemoteAddr: "127.0.0.1:54654",
	}

	userServiceMock := new(mocks.UserServiceMock)
	userServiceMock.On("GetUserById", testUser.ID).Return(testUser, nil)

	sut := NewAuthenticateMiddleware(userServiceMock)

	result, actualErr := sut.tryGetUserByTrustedHeader(mockRequest)
	assert.Error(t, actualErr)
	assert.Nil(t, result)
}

func TestAuthenticateMiddleware_tryGetUserByTrustedHeader_Success(t *testing.T) {
	cfg := config.Empty()
	cfg.Security.TrustedHeaderAuth = true
	cfg.Security.TrustedHeaderAuthKey = "Remote-User"
	cfg.Security.TrustReverseProxyIps = "127.0.0.1,::1"
	cfg.Security.ParseTrustReverseProxyIPs()
	config.Set(cfg)

	testUser := &models.User{ID: "user01"}

	mockRequest := &http.Request{
		Header:     http.Header{"Remote-User": []string{testUser.ID}},
		RemoteAddr: "[::1]:54654",
	}

	userServiceMock := new(mocks.UserServiceMock)
	userServiceMock.On("GetUserById", testUser.ID).Return(testUser, nil)

	sut := NewAuthenticateMiddleware(userServiceMock)

	result, actualErr := sut.tryGetUserByTrustedHeader(mockRequest)
	assert.Equal(t, testUser, result)
	assert.Nil(t, actualErr)
}

// TODO: somehow test cookie auth function
