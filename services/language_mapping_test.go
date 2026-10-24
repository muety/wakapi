package services

import (
	"testing"
	"time"

	"github.com/jellydator/ttlcache/v3"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type languageMappingRepoMock struct {
	mocks.BaseRepositoryMock
	mock.Mock
}

func (m *languageMappingRepoMock) GetAll() ([]*models.LanguageMapping, error) {
	args := m.Called()
	return args.Get(0).([]*models.LanguageMapping), args.Error(1)
}

func (m *languageMappingRepoMock) GetById(id uint) (*models.LanguageMapping, error) {
	args := m.Called(id)
	return args.Get(0).(*models.LanguageMapping), args.Error(1)
}

func (m *languageMappingRepoMock) GetByUser(userId string) ([]*models.LanguageMapping, error) {
	args := m.Called(userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.LanguageMapping), args.Error(1)
}

func (m *languageMappingRepoMock) Insert(mapping *models.LanguageMapping) (*models.LanguageMapping, error) {
	args := m.Called(mapping)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LanguageMapping), args.Error(1)
}

func (m *languageMappingRepoMock) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

type LanguageMappingServiceTestSuite struct {
	suite.Suite
	repo *languageMappingRepoMock
}

func (suite *LanguageMappingServiceTestSuite) SetupSuite() {
	config.Set(config.Empty())
}

func (suite *LanguageMappingServiceTestSuite) BeforeTest(suiteName, testName string) {
	suite.repo = new(languageMappingRepoMock)
}

func TestLanguageMappingServiceTestSuite(t *testing.T) {
	suite.Run(t, new(LanguageMappingServiceTestSuite))
}

func (suite *LanguageMappingServiceTestSuite) TestLanguageMappingService_GetByUser_CacheRetrieval() {
	sut := NewLanguageMappingService(suite.repo)

	testMappings := []*models.LanguageMapping{
		{ID: 1, UserID: "user1", Extension: "m", Language: "Objective-C"},
	}

	suite.repo.On("GetByUser", "user1").Return(testMappings, nil).Once()

	// 1. Initial retrieval - cache miss, repo called once
	result1, err := sut.GetByUser("user1")
	suite.Nil(err)
	suite.Equal(testMappings, result1)
	suite.repo.AssertNumberOfCalls(suite.T(), "GetByUser", 1)

	// 2. Second retrieval - cache hit, repo not called again
	result2, err := sut.GetByUser("user1")
	suite.Nil(err)
	suite.Equal(testMappings, result2)
	suite.repo.AssertNumberOfCalls(suite.T(), "GetByUser", 1)
}

func (suite *LanguageMappingServiceTestSuite) TestLanguageMappingService_Create_CacheEviction() {
	sut := NewLanguageMappingService(suite.repo)

	initialMappings := []*models.LanguageMapping{
		{ID: 1, UserID: "user1", Extension: "m", Language: "Objective-C"},
	}
	newMapping := &models.LanguageMapping{
		UserID:    "user1",
		Extension: "h",
		Language:  "C",
	}
	createdMapping := &models.LanguageMapping{
		ID:        2,
		UserID:    "user1",
		Extension: "h",
		Language:  "C",
	}

	suite.repo.On("GetByUser", "user1").Return(initialMappings, nil).Once()
	suite.repo.On("Insert", newMapping).Return(createdMapping, nil).Once()

	// 1. Populate cache
	_, err := sut.GetByUser("user1")
	suite.Nil(err)
	suite.repo.AssertNumberOfCalls(suite.T(), "GetByUser", 1)

	// 2. Create evicts user cache
	_, err = sut.Create(newMapping)
	suite.Nil(err)

	// 3. Setup updated repo return
	updatedMappings := append(initialMappings, createdMapping)
	suite.repo.On("GetByUser", "user1").Return(updatedMappings, nil).Once()

	// 4. Retrieval after create - cache miss, repo called again
	result, err := sut.GetByUser("user1")
	suite.Nil(err)
	suite.Equal(updatedMappings, result)
	suite.repo.AssertNumberOfCalls(suite.T(), "GetByUser", 2)
}

func (suite *LanguageMappingServiceTestSuite) TestLanguageMappingService_Delete_CacheEviction() {
	sut := NewLanguageMappingService(suite.repo)

	initialMappings := []*models.LanguageMapping{
		{ID: 1, UserID: "user1", Extension: "m", Language: "Objective-C"},
	}
	toDelete := &models.LanguageMapping{
		ID:     1,
		UserID: "user1",
	}

	suite.repo.On("GetByUser", "user1").Return(initialMappings, nil).Once()
	suite.repo.On("Delete", uint(1)).Return(nil).Once()

	// 1. Populate cache
	_, err := sut.GetByUser("user1")
	suite.Nil(err)
	suite.repo.AssertNumberOfCalls(suite.T(), "GetByUser", 1)

	// 2. Delete evicts user cache
	err = sut.Delete(toDelete)
	suite.Nil(err)

	// 3. Setup repo return after deletion
	suite.repo.On("GetByUser", "user1").Return([]*models.LanguageMapping{}, nil).Once()

	// 4. Retrieval after delete - cache miss, repo called again
	result, err := sut.GetByUser("user1")
	suite.Nil(err)
	suite.Empty(result)
	suite.repo.AssertNumberOfCalls(suite.T(), "GetByUser", 2)
}

func (suite *LanguageMappingServiceTestSuite) TestLanguageMappingService_CacheTTLEviction() {
	shortCache := ttlcache.New[string, []*models.LanguageMapping](
		ttlcache.WithTTL[string, []*models.LanguageMapping](50*time.Millisecond),
		ttlcache.WithDisableTouchOnHit[string, []*models.LanguageMapping](),
	)
	go shortCache.Start()
	defer shortCache.Stop()

	sut := &LanguageMappingService{
		cache:      shortCache,
		repository: suite.repo,
	}

	testMappings := []*models.LanguageMapping{
		{ID: 1, UserID: "user1", Extension: "m", Language: "Objective-C"},
	}

	suite.repo.On("GetByUser", "user1").Return(testMappings, nil).Twice()

	// 1. Initial retrieval - cache miss
	r1, err := sut.GetByUser("user1")
	suite.Nil(err)
	suite.Equal(testMappings, r1)
	suite.repo.AssertNumberOfCalls(suite.T(), "GetByUser", 1)

	// 2. Immediate retrieval - cache hit
	r2, err := sut.GetByUser("user1")
	suite.Nil(err)
	suite.Equal(testMappings, r2)
	suite.repo.AssertNumberOfCalls(suite.T(), "GetByUser", 1)

	// 3. Wait for TTL to expire
	time.Sleep(100 * time.Millisecond)

	// 4. Retrieval after TTL - cache expired, repo called again
	r3, err := sut.GetByUser("user1")
	suite.Nil(err)
	suite.Equal(testMappings, r3)
	suite.repo.AssertNumberOfCalls(suite.T(), "GetByUser", 2)
}
