package services

import (
	"errors"
	"testing"

	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/suite"
)

const (
	TestUserID = "muety"
	TestAPIKey = "full-access-key-from-user-model"
)

type UserServiceTestSuite struct {
	suite.Suite
	TestUser        *models.User
	KeyValueService *mocks.KeyValueServiceMock
	MailService     *mocks.MailServiceMock
	ApiKeyService   *mocks.MockApiKeyService
	UserRepo        *mocks.UserRepositoryMock
}

func (suite *UserServiceTestSuite) SetupSuite() {
	suite.TestUser = &models.User{ID: TestUserID, ApiKey: TestAPIKey}
}

func (suite *UserServiceTestSuite) BeforeTest(suiteName, testName string) {
	suite.KeyValueService = new(mocks.KeyValueServiceMock)
	suite.MailService = new(mocks.MailServiceMock)
	suite.ApiKeyService = new(mocks.MockApiKeyService)
	suite.UserRepo = new(mocks.UserRepositoryMock)
}

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

func (suite *UserServiceTestSuite) TestUserService_GetByEmail_Empty() {
	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)

	result, err := sut.GetUserByEmail("")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(err, errors.New("email must not be empty"))
}

func (suite *UserServiceTestSuite) TestUserService_GetByEmail_Invalid() {
	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)

	result, err := sut.GetUserByEmail("notanemailaddress")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(err, errors.New("not a valid email"))
}

func (suite *UserServiceTestSuite) TestUserService_GetByEmail_Valid() {
	const testEmail = "foo@bar.com"

	suite.UserRepo.On("FindOne", models.User{Email: testEmail}).Return(suite.TestUser, nil)

	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)
	result, err := sut.GetUserByEmail(testEmail)

	suite.Equal(suite.TestUser, result)
	suite.Nil(err)
}

func (suite *UserServiceTestSuite) TestUserService_GetByEmptyKey_Failed() {
	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)

	result, err := sut.GetUserByKey("", false)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(err, errors.New("key must not be empty"))
}

func (suite *UserServiceTestSuite) TestUserService_GetByKeyFromCache_Success() {
	userCached := &models.User{ID: TestUserID, ApiKey: "cached-key"}

	userCache := cache.New(cache.NoExpiration, cache.NoExpiration)
	userCache.SetDefault(TestAPIKey, userCached)

	sut := &UserService{cache: userCache}

	result, err := sut.GetUserByKey(TestAPIKey, false)
	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(1, userCache.ItemCount())
	suite.Equal(userCached, result)
	suite.Equal(userCached.ApiKey, result.ApiKey)
}

func (suite *UserServiceTestSuite) TestUserService_GetByKeyFromUserModel_Success() {
	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)

	suite.UserRepo.On("FindOne", models.User{ApiKey: TestAPIKey}).Return(suite.TestUser, nil)

	result, err := sut.GetUserByKey(TestAPIKey, false)
	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(suite.TestUser, result)
	suite.Equal(suite.TestUser.ApiKey, result.ApiKey)
}

func (suite *UserServiceTestSuite) TestUserService_GetByKeyFromAdditionalApiKeys_Success() {
	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)

	suite.UserRepo.On("FindOne", models.User{ApiKey: TestAPIKey}).Return(nil, errors.New("not found"))
	suite.ApiKeyService.On("GetByApiKey", TestAPIKey, true).Return(&models.ApiKey{User: suite.TestUser}, nil)

	result, err := sut.GetUserByKey(TestAPIKey, true)
	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(suite.TestUser, result)
	suite.Equal(suite.TestUser.ApiKey, result.ApiKey)
}

func (suite *UserServiceTestSuite) TestUserService_GetByKeyFromAdditionalApiKeys_Failed() {
	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)

	suite.UserRepo.On("FindOne", models.User{ApiKey: TestAPIKey}).Return(nil, errors.New("not found"))
	suite.ApiKeyService.On("GetByApiKey", TestAPIKey, true).Return(nil, errors.New("not found"))

	result, err := sut.GetUserByKey(TestAPIKey, true)
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(err, errors.New("not found"))
}
