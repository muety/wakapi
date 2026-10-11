package services

import (
	"testing"
	"time"

	"github.com/leandro-lugaresi/hub"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type HeartbeatServiceTestSuite struct {
	suite.Suite
	HeartbeatRepository *mocks.HeartbeatRepositoryMock
	eventBus            *hub.Hub
	sut                 *HeartbeatService
	user                *models.User
}

func (suite *HeartbeatServiceTestSuite) SetupSuite() {
	config.Set(config.Empty())
}

func (suite *HeartbeatServiceTestSuite) BeforeTest(suiteName, testName string) {
	originalEventBus := config.EventBus()
	suite.eventBus = hub.New()
	config.SetEventBus(suite.eventBus)

	suite.HeartbeatRepository = new(mocks.HeartbeatRepositoryMock)
	suite.sut = NewHeartbeatService(suite.HeartbeatRepository, new(mocks.LanguageMappingServiceMock))

	config.SetEventBus(originalEventBus)
	suite.user = &models.User{ID: "u1"}
}

func TestHeartbeatServiceTestSuite(t *testing.T) {
	suite.Run(t, new(HeartbeatServiceTestSuite))
}

func (suite *HeartbeatServiceTestSuite) publishHeartbeat(heartbeat *models.Heartbeat) {
	suite.eventBus.Publish(hub.Message{
		Name:   config.EventHeartbeatCreate,
		Fields: map[string]interface{}{config.FieldPayload: heartbeat},
	})
}

func createTestHeartbeatAt(id uint64, userId, project string, timestamp time.Time) *models.Heartbeat {
	return &models.Heartbeat{ID: id, UserID: userId, Project: project, Branch: "main", Language: "Go", Time: models.CustomTime(timestamp)}
}

func (suite *HeartbeatServiceTestSuite) TestHeartbeatService_GetLatestByUser_Caches() {
	first := createTestHeartbeatAt(1, suite.user.ID, "p1", time.Unix(100, 0))
	second := createTestHeartbeatAt(2, suite.user.ID, "p2", time.Unix(200, 0))

	suite.HeartbeatRepository.On("GetLatestByUser", suite.user).Return(first, nil).Once()
	suite.HeartbeatRepository.On("GetLatestByUser", suite.user).Return(second, nil).Once()

	result, err := suite.sut.GetLatestByUser(suite.user, false)
	suite.Require().NoError(err)
	suite.Equal(uint64(1), result.ID)

	// served from cache, no second repository call
	result, err = suite.sut.GetLatestByUser(suite.user, false)
	suite.Require().NoError(err)
	suite.Equal(uint64(1), result.ID)

	// skipCache bypasses the cache, but refreshes it
	result, err = suite.sut.GetLatestByUser(suite.user, true)
	suite.Require().NoError(err)
	suite.Equal(uint64(2), result.ID)

	suite.HeartbeatRepository.AssertNumberOfCalls(suite.T(), "GetLatestByUser", 2)

	// the refreshed value is now what subsequent cached reads serve
	result, err = suite.sut.GetLatestByUser(suite.user, false)
	suite.Require().NoError(err)
	suite.Equal(uint64(2), result.ID)

	suite.HeartbeatRepository.AssertNumberOfCalls(suite.T(), "GetLatestByUser", 2)
	suite.HeartbeatRepository.AssertExpectations(suite.T())
}

func (suite *HeartbeatServiceTestSuite) TestHeartbeatService_GetLatestByUser_CachesEmptyResult() {
	suite.HeartbeatRepository.On("GetLatestByUser", suite.user).Return(&models.Heartbeat{}, nil).Once()

	for range 2 {
		result, err := suite.sut.GetLatestByUser(suite.user, false)
		suite.Require().NoError(err)
		suite.Zero(result.ID)
	}

	suite.HeartbeatRepository.AssertNumberOfCalls(suite.T(), "GetLatestByUser", 1)
}

func (suite *HeartbeatServiceTestSuite) TestHeartbeatService_GetLatestByUser_CacheRefreshedByHeartbeatEvents() {
	initial := createTestHeartbeatAt(1, suite.user.ID, "p1", time.Unix(100, 0))
	suite.HeartbeatRepository.On("GetLatestByUser", suite.user).Return(initial, nil).Once()

	result, err := suite.sut.GetLatestByUser(suite.user, false)
	suite.Require().NoError(err)
	suite.Equal(uint64(1), result.ID)

	suite.publishHeartbeat(createTestHeartbeatAt(2, suite.user.ID, "p2", time.Unix(200, 0)))

	suite.Eventually(func() bool {
		result, err := suite.sut.GetLatestByUser(suite.user, false)
		return err == nil && result.ID == 2 && result.Project == "p2"
	}, 2*time.Second, 20*time.Millisecond)

	// an older heartbeat must not overwrite a newer one
	suite.publishHeartbeat(createTestHeartbeatAt(3, suite.user.ID, "p0", time.Unix(50, 0)))
	time.Sleep(200 * time.Millisecond)

	result, err = suite.sut.GetLatestByUser(suite.user, false)
	suite.Require().NoError(err)
	suite.Equal(uint64(2), result.ID)

	suite.HeartbeatRepository.AssertNumberOfCalls(suite.T(), "GetLatestByUser", 1)
}

func (suite *HeartbeatServiceTestSuite) TestHeartbeatService_LatestCacheNotEagerlyPopulatedByEvents() {
	suite.publishHeartbeat(createTestHeartbeatAt(1, suite.user.ID, "p1", time.Unix(100, 0)))

	// the entry for this user was never read before, so the cache should be empty and the repository consulted
	suite.HeartbeatRepository.On("GetLatestByUser", suite.user).Return(&models.Heartbeat{}, nil).Once()

	_, err := suite.sut.GetLatestByUser(suite.user, false)
	suite.Require().NoError(err)

	suite.HeartbeatRepository.AssertExpectations(suite.T())
}

func (suite *HeartbeatServiceTestSuite) TestHeartbeatService_LatestCacheKeepsProjectDimensionInSync() {
	repoResult := createTestHeartbeatAt(1, suite.user.ID, "p1", time.Unix(100, 0))
	suite.HeartbeatRepository.On("GetLatestByFilters", suite.user, map[string][]string{"project": {"p1"}}).Return(repoResult, nil).Once()

	_, err := suite.sut.GetLatestByFilters(suite.user, models.NewFiltersWith(models.SummaryProject, "p1"), false)
	suite.Require().NoError(err)

	suite.publishHeartbeat(createTestHeartbeatAt(2, suite.user.ID, "p1", time.Unix(200, 0)))

	suite.Eventually(func() bool {
		result, err := suite.sut.GetLatestByFilters(suite.user, models.NewFiltersWith(models.SummaryProject, "p1"), false)
		return err == nil && result.ID == 2
	}, 2*time.Second, 20*time.Millisecond)

	suite.HeartbeatRepository.AssertNumberOfCalls(suite.T(), "GetLatestByFilters", 1)
}

func (suite *HeartbeatServiceTestSuite) TestHeartbeatService_GetLatestByFilters_CachesSingleProjectFilter() {
	repoResult := createTestHeartbeatAt(1, suite.user.ID, "p1", time.Unix(100, 0))
	suite.HeartbeatRepository.On("GetLatestByFilters", suite.user, map[string][]string{"project": {"p1"}}).Return(repoResult, nil)

	for range 2 {
		result, err := suite.sut.GetLatestByFilters(suite.user, models.NewFiltersWith(models.SummaryProject, "p1"), false)
		suite.Require().NoError(err)
		suite.Equal(uint64(1), result.ID)
	}

	result, err := suite.sut.GetLatestByFilters(suite.user, models.NewFiltersWith(models.SummaryProject, "p1"), true)
	suite.Require().NoError(err)
	suite.Equal(uint64(1), result.ID)

	suite.HeartbeatRepository.AssertNumberOfCalls(suite.T(), "GetLatestByFilters", 2)
}

func (suite *HeartbeatServiceTestSuite) TestHeartbeatService_GetLatestByFilters_FallsThroughForOtherFilterings() {
	repoResult := createTestHeartbeatAt(1, suite.user.ID, "p1", time.Unix(100, 0))
	suite.HeartbeatRepository.On("GetLatestByFilters", mock.Anything, mock.Anything).Return(repoResult, nil)

	// multiple projects in the filter
	for range 2 {
		_, err := suite.sut.GetLatestByFilters(suite.user, models.NewFilterWithMultiple(models.SummaryProject, []string{"p1", "p2"}), false)
		suite.Require().NoError(err)
	}

	// project combined with another filter type
	for range 2 {
		_, err := suite.sut.GetLatestByFilters(suite.user, models.NewFiltersWith(models.SummaryProject, "p1").With(models.SummaryLanguage, "Go"), false)
		suite.Require().NoError(err)
	}

	// language-only filter
	_, err := suite.sut.GetLatestByFilters(suite.user, models.NewFiltersWith(models.SummaryLanguage, "Go"), false)
	suite.Require().NoError(err)

	suite.HeartbeatRepository.AssertNumberOfCalls(suite.T(), "GetLatestByFilters", 5)
}

func (suite *HeartbeatServiceTestSuite) TestHeartbeatService_GetLatestByFilters_CachesMissingRows() {
	suite.HeartbeatRepository.On("GetLatestByFilters", suite.user, map[string][]string{"project": {"p1"}}).Return((*models.Heartbeat)(nil), nil).Once()

	for range 2 {
		result, err := suite.sut.GetLatestByFilters(suite.user, models.NewFiltersWith(models.SummaryProject, "p1"), false)
		suite.Require().NoError(err)
		suite.Nil(result)
	}

	suite.HeartbeatRepository.AssertNumberOfCalls(suite.T(), "GetLatestByFilters", 1)

	// once a heartbeat for that project arrives, the cached 'missing' sentinel must be upgraded to the real one
	suite.publishHeartbeat(createTestHeartbeatAt(2, suite.user.ID, "p1", time.Unix(200, 0)))

	suite.Eventually(func() bool {
		result, err := suite.sut.GetLatestByFilters(suite.user, models.NewFiltersWith(models.SummaryProject, "p1"), false)
		return err == nil && result != nil && result.ID == 2
	}, 2*time.Second, 20*time.Millisecond)

	suite.HeartbeatRepository.AssertNumberOfCalls(suite.T(), "GetLatestByFilters", 1)
}
