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

type projectLabelRepoMock struct {
	mocks.BaseRepositoryMock
	mock.Mock
}

func (m *projectLabelRepoMock) GetAll() ([]*models.ProjectLabel, error) {
	args := m.Called()
	return args.Get(0).([]*models.ProjectLabel), args.Error(1)
}

func (m *projectLabelRepoMock) GetById(id uint) (*models.ProjectLabel, error) {
	args := m.Called(id)
	return args.Get(0).(*models.ProjectLabel), args.Error(1)
}

func (m *projectLabelRepoMock) GetByUser(userId string) ([]*models.ProjectLabel, error) {
	args := m.Called(userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ProjectLabel), args.Error(1)
}

func (m *projectLabelRepoMock) Insert(label *models.ProjectLabel) (*models.ProjectLabel, error) {
	args := m.Called(label)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProjectLabel), args.Error(1)
}

func (m *projectLabelRepoMock) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

type ProjectLabelServiceTestSuite struct {
	suite.Suite
	repo *projectLabelRepoMock
}

func (suite *ProjectLabelServiceTestSuite) SetupSuite() {
	config.Set(config.Empty())
}

func (suite *ProjectLabelServiceTestSuite) BeforeTest(suiteName, testName string) {
	suite.repo = new(projectLabelRepoMock)
}

func TestProjectLabelServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ProjectLabelServiceTestSuite))
}

func (suite *ProjectLabelServiceTestSuite) TestProjectLabelService_GetByUser_CacheRetrieval() {
	sut := NewProjectLabelService(suite.repo)

	testLabels := []*models.ProjectLabel{
		{ID: 1, UserID: "user1", ProjectKey: "wakapi", Label: "backend"},
	}

	suite.repo.On("GetByUser", "user1").Return(testLabels, nil).Once()

	// 1. Initial retrieval - cache miss, repo called once
	result1, err := sut.GetByUser("user1")
	suite.Nil(err)
	suite.Equal(testLabels, result1)
	suite.repo.AssertNumberOfCalls(suite.T(), "GetByUser", 1)

	// 2. Second retrieval - cache hit, repo not called again
	result2, err := sut.GetByUser("user1")
	suite.Nil(err)
	suite.Equal(testLabels, result2)
	suite.repo.AssertNumberOfCalls(suite.T(), "GetByUser", 1)
}

func (suite *ProjectLabelServiceTestSuite) TestProjectLabelService_Create_CacheEviction() {
	sut := NewProjectLabelService(suite.repo)

	initialLabels := []*models.ProjectLabel{
		{ID: 1, UserID: "user1", ProjectKey: "wakapi", Label: "backend"},
	}
	newLabel := &models.ProjectLabel{
		UserID:     "user1",
		ProjectKey: "wakapi-frontend",
		Label:      "frontend",
	}
	createdLabel := &models.ProjectLabel{
		ID:         2,
		UserID:     "user1",
		ProjectKey: "wakapi-frontend",
		Label:      "frontend",
	}

	suite.repo.On("GetByUser", "user1").Return(initialLabels, nil).Once()
	suite.repo.On("Insert", newLabel).Return(createdLabel, nil).Once()

	// 1. Populate cache
	_, err := sut.GetByUser("user1")
	suite.Nil(err)
	suite.repo.AssertNumberOfCalls(suite.T(), "GetByUser", 1)

	// 2. Create evicts user cache
	_, err = sut.Create(newLabel)
	suite.Nil(err)

	// 3. Setup updated repo return
	updatedLabels := append(initialLabels, createdLabel)
	suite.repo.On("GetByUser", "user1").Return(updatedLabels, nil).Once()

	// 4. Retrieval after create - cache miss, repo called again
	result, err := sut.GetByUser("user1")
	suite.Nil(err)
	suite.Equal(updatedLabels, result)
	suite.repo.AssertNumberOfCalls(suite.T(), "GetByUser", 2)
}

func (suite *ProjectLabelServiceTestSuite) TestProjectLabelService_Delete_CacheEviction() {
	sut := NewProjectLabelService(suite.repo)

	initialLabels := []*models.ProjectLabel{
		{ID: 1, UserID: "user1", ProjectKey: "wakapi", Label: "backend"},
	}
	toDelete := &models.ProjectLabel{
		ID:     1,
		UserID: "user1",
	}

	suite.repo.On("GetByUser", "user1").Return(initialLabels, nil).Once()
	suite.repo.On("Delete", uint(1)).Return(nil).Once()

	// 1. Populate cache
	_, err := sut.GetByUser("user1")
	suite.Nil(err)
	suite.repo.AssertNumberOfCalls(suite.T(), "GetByUser", 1)

	// 2. Delete evicts user cache
	err = sut.Delete(toDelete)
	suite.Nil(err)

	// 3. Setup repo return after deletion
	suite.repo.On("GetByUser", "user1").Return([]*models.ProjectLabel{}, nil).Once()

	// 4. Retrieval after delete - cache miss, repo called again
	result, err := sut.GetByUser("user1")
	suite.Nil(err)
	suite.Empty(result)
	suite.repo.AssertNumberOfCalls(suite.T(), "GetByUser", 2)
}

func (suite *ProjectLabelServiceTestSuite) TestProjectLabelService_CacheTTLEviction() {
	shortCache := ttlcache.New[string, []*models.ProjectLabel](
		ttlcache.WithTTL[string, []*models.ProjectLabel](50*time.Millisecond),
		ttlcache.WithDisableTouchOnHit[string, []*models.ProjectLabel](),
	)
	go shortCache.Start()
	defer shortCache.Stop()

	sut := &ProjectLabelService{
		cache:      shortCache,
		repository: suite.repo,
	}

	testLabels := []*models.ProjectLabel{
		{ID: 1, UserID: "user1", ProjectKey: "wakapi", Label: "backend"},
	}

	suite.repo.On("GetByUser", "user1").Return(testLabels, nil).Twice()

	// 1. Initial retrieval - cache miss
	r1, err := sut.GetByUser("user1")
	suite.Nil(err)
	suite.Equal(testLabels, r1)
	suite.repo.AssertNumberOfCalls(suite.T(), "GetByUser", 1)

	// 2. Immediate retrieval - cache hit
	r2, err := sut.GetByUser("user1")
	suite.Nil(err)
	suite.Equal(testLabels, r2)
	suite.repo.AssertNumberOfCalls(suite.T(), "GetByUser", 1)

	// 3. Wait for TTL to expire
	time.Sleep(100 * time.Millisecond)

	// 4. Retrieval after TTL - cache expired, repo called again
	r3, err := sut.GetByUser("user1")
	suite.Nil(err)
	suite.Equal(testLabels, r3)
	suite.repo.AssertNumberOfCalls(suite.T(), "GetByUser", 2)
}
