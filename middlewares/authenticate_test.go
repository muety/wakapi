package middlewares

import (
	"encoding/base64"
	"fmt"
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

// TODO: somehow test cookie auth function
