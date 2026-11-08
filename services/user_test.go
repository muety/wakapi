package services

import (
	"errors"
	"testing"
	"time"

	"github.com/jellydator/ttlcache/v3"
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
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

	userCache := ttlcache.New[string, any](ttlcache.WithTTL[string, any](ttlcache.NoTTL))
	userCache.Set("key_"+TestAPIKey+"_false", userCached, ttlcache.DefaultTTL)

	sut := &UserService{cache: userCache}

	result, err := sut.GetUserByKey(TestAPIKey, false)
	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(1, userCache.Len())
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

func (suite *UserServiceTestSuite) TestUserService_GetUserByKey_DoesNotHitUserByIdCache() {
	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)

	suite.UserRepo.On("FindOne", models.User{ID: TestUserID}).Return(suite.TestUser, nil)
	suite.UserRepo.On("FindOne", models.User{ApiKey: TestUserID}).Return(nil, errors.New("user not found"))
	suite.ApiKeyService.On("GetByApiKey", TestUserID, false).Return(nil, errors.New("key not found"))

	// 1. Prime cache by looking up user by ID (username)
	userByID, err := sut.GetUserById(TestUserID)
	suite.Nil(err)
	suite.Equal(suite.TestUser, userByID)

	// 2. Attempt to look up user by API key passing username as key
	userByKey, err := sut.GetUserByKey(TestUserID, false)
	suite.Nil(userByKey)
	suite.NotNil(err)
}

func (suite *UserServiceTestSuite) TestUserService_GetUserByKey_DoesNotCacheUnderUsername() {
	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)

	suite.UserRepo.On("FindOne", models.User{ApiKey: TestAPIKey}).Return(suite.TestUser, nil)
	suite.UserRepo.On("FindOne", models.User{ApiKey: TestUserID}).Return(nil, errors.New("user not found"))
	suite.ApiKeyService.On("GetByApiKey", TestUserID, false).Return(nil, errors.New("key not found"))

	// 1. Look up user by valid API key
	userByKeyValid, err := sut.GetUserByKey(TestAPIKey, false)
	suite.Nil(err)
	suite.Equal(suite.TestUser, userByKeyValid)

	// 2. Attempt to look up user by API key using username as key
	userByKeyInvalid, err := sut.GetUserByKey(TestUserID, false)
	suite.Nil(userByKeyInvalid)
	suite.NotNil(err)
}

func (suite *UserServiceTestSuite) TestUserService_GetUserByKey_ReadOnlyKeyCacheDoesNotGrantFullAccess() {
	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)

	suite.UserRepo.On("FindOne", models.User{ApiKey: TestAPIKey}).Return(nil, errors.New("not primary key"))
	suite.ApiKeyService.On("GetByApiKey", TestAPIKey, false).Return(&models.ApiKey{User: suite.TestUser, ReadOnly: true}, nil)
	suite.ApiKeyService.On("GetByApiKey", TestAPIKey, true).Return(nil, errors.New("key is read-only"))

	// 1. Look up with requireFullAccessKey = false (should cache under read-only key)
	userReadOnly, err := sut.GetUserByKey(TestAPIKey, false)
	suite.Nil(err)
	suite.Equal(suite.TestUser, userReadOnly)

	// 2. Look up with requireFullAccessKey = true (must not hit read-only cache)
	userFullAccess, err := sut.GetUserByKey(TestAPIKey, true)
	suite.Nil(userFullAccess)
	suite.NotNil(err)
}

func (suite *UserServiceTestSuite) TestUserService_GetUserById_CacheRetrievalAndFlushUserCache() {
	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)

	suite.UserRepo.On("FindOne", models.User{ID: TestUserID}).Return(suite.TestUser, nil).Once()

	// 1. Initial retrieval - cache miss, repository called once
	user, err := sut.GetUserById(TestUserID)
	suite.Nil(err)
	suite.Equal(suite.TestUser, user)

	// 2. Second retrieval - cache hit, repository FindOne is not called again
	userCached, err := sut.GetUserById(TestUserID)
	suite.Nil(err)
	suite.Equal(suite.TestUser, userCached)
	suite.UserRepo.AssertNumberOfCalls(suite.T(), "FindOne", 1)

	// 3. Evict cache via FlushUserCache
	sut.FlushUserCache(TestUserID)

	// 4. Setup mock expectation for repository call after eviction
	suite.UserRepo.On("FindOne", models.User{ID: TestUserID}).Return(suite.TestUser, nil).Once()

	// 5. Third retrieval after eviction - repository called again
	userReloaded, err := sut.GetUserById(TestUserID)
	suite.Nil(err)
	suite.Equal(suite.TestUser, userReloaded)
	suite.UserRepo.AssertNumberOfCalls(suite.T(), "FindOne", 2)
}

func (suite *UserServiceTestSuite) TestUserService_FlushCache() {
	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)

	suite.UserRepo.On("FindOne", models.User{ID: TestUserID}).Return(suite.TestUser, nil).Once()

	// Populate cache
	_, err := sut.GetUserById(TestUserID)
	suite.Nil(err)
	suite.NotZero(sut.cache.Len())

	// Flush cache
	sut.FlushCache()
	suite.Zero(sut.cache.Len())
}

func (suite *UserServiceTestSuite) TestUserService_CacheTTLEviction() {
	shortCache := ttlcache.New[string, any](
		ttlcache.WithTTL[string, any](50*time.Millisecond),
		ttlcache.WithDisableTouchOnHit[string, any](),
	)
	go shortCache.Start()
	defer shortCache.Stop()

	sut := &UserService{
		cache:      shortCache,
		repository: suite.UserRepo,
	}

	suite.UserRepo.On("FindOne", models.User{ID: TestUserID}).Return(suite.TestUser, nil).Twice()

	// 1. Initial retrieval - cache miss
	user1, err := sut.GetUserById(TestUserID)
	suite.Nil(err)
	suite.Equal(suite.TestUser, user1)
	suite.UserRepo.AssertNumberOfCalls(suite.T(), "FindOne", 1)

	// 2. Immediate retrieval - cache hit
	userCached, err := sut.GetUserById(TestUserID)
	suite.Nil(err)
	suite.Equal(suite.TestUser, userCached)
	suite.UserRepo.AssertNumberOfCalls(suite.T(), "FindOne", 1)

	// 3. Wait for TTL to expire
	time.Sleep(100 * time.Millisecond)

	// 4. Retrieval after TTL - cache expired, repo called again
	userExpired, err := sut.GetUserById(TestUserID)
	suite.Nil(err)
	suite.Equal(suite.TestUser, userExpired)
	suite.UserRepo.AssertNumberOfCalls(suite.T(), "FindOne", 2)
}
