package middlewares

import (
	"encoding/base64"
	"fmt"
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"testing"
)

func TestAuthenticateMiddleware_tryGetUserByApiKey_Success(t *testing.T) {
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

	sut := NewAuthenticateMiddleware(userServiceMock, []string{})

	result, err := sut.tryGetUserByApiKey(mockRequest)

	assert.Nil(t, err)
	assert.Equal(t, testUser, result)
}

func TestAuthenticateMiddleware_tryGetUserByApiKey_GetFromCache(t *testing.T) {
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

	sut := NewAuthenticateMiddleware(userServiceMock, []string{})
	sut.cache.SetDefault(testApiKey, testUser)

	result, err := sut.tryGetUserByApiKey(mockRequest)

	assert.Nil(t, err)
	assert.Equal(t, testUser, result)
	userServiceMock.AssertNotCalled(t, "GetUserByKey", mock.Anything)
}

func TestAuthenticateMiddleware_tryGetUserByApiKey_InvalidHeader(t *testing.T) {
	testApiKey := "z5uig69cn9ut93n"
	testToken := base64.StdEncoding.EncodeToString([]byte(testApiKey))

	mockRequest := &http.Request{
		Header: http.Header{
			// 'Basic' prefix missing here
			"Authorization": []string{fmt.Sprintf("%s", testToken)},
		},
	}

	userServiceMock := new(mocks.UserServiceMock)

	sut := NewAuthenticateMiddleware(userServiceMock, []string{})

	result, err := sut.tryGetUserByApiKey(mockRequest)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// TODO: somehow test cookie auth function
