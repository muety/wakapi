package services

import (
	"testing"
	"time"

	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type HousekeepingServiceTestSuite struct {
	suite.Suite
	TestUsers        []*models.User
	UserService      *mocks.UserServiceMock
	HeartbeatService *mocks.HeartbeatServiceMock
	DurationService  *mocks.DurationServiceMock
	ProjectService   *mocks.ProjectServiceMock
	SummaryService   *mocks.SummaryServiceMock
	BaseRepository   *mocks.BaseRepositoryMock
}

func (suite *HousekeepingServiceTestSuite) SetupSuite() {
	suite.TestUsers = []*models.User{
		{ID: "testuser01", LastLoggedInAt: models.CustomTime(time.Now().AddDate(0, -16, 0)), HasData: false},
		{ID: "testuser02", LastLoggedInAt: models.CustomTime(time.Now().AddDate(0, -16, 0)), HasData: true},
		{ID: "testuser03", LastLoggedInAt: models.CustomTime(time.Now().AddDate(0, -1, 0)), HasData: false},
	}
}

func (suite *HousekeepingServiceTestSuite) BeforeTest(suiteName, testName string) {
	suite.UserService = new(mocks.UserServiceMock)
	suite.HeartbeatService = new(mocks.HeartbeatServiceMock)
	suite.DurationService = new(mocks.DurationServiceMock)
	suite.ProjectService = new(mocks.ProjectServiceMock)
	suite.SummaryService = new(mocks.SummaryServiceMock)
	suite.BaseRepository = new(mocks.BaseRepositoryMock)
}

func TestHouseKeepingServiceTestSuite(t *testing.T) {
	suite.Run(t, new(HousekeepingServiceTestSuite))
}

func (suite *HousekeepingServiceTestSuite) TestHousekeepingService_CleanInactiveUsers() {
	sut := NewHousekeepingService(suite.UserService, suite.HeartbeatService, suite.DurationService, suite.ProjectService, suite.SummaryService, suite.BaseRepository)

	suite.UserService.On("GetAll").Return(suite.TestUsers, nil)
	suite.UserService.On("Delete", suite.TestUsers[0]).Return(nil)

	err := sut.CleanInactiveUsers(time.Now().AddDate(0, -12, 0))

	assert.Nil(suite.T(), err)
	suite.UserService.AssertNumberOfCalls(suite.T(), "GetAll", 1)
	suite.UserService.AssertNumberOfCalls(suite.T(), "Delete", 1)
	suite.UserService.AssertCalled(suite.T(), "Delete", suite.TestUsers[0])
}

func (suite *HousekeepingServiceTestSuite) TestHousekeepingService_CleanUserDataBefore() {
	sut := NewHousekeepingService(suite.UserService, suite.HeartbeatService, suite.DurationService, suite.ProjectService, suite.SummaryService, suite.BaseRepository)

	user := suite.TestUsers[0]
	before := time.Now().AddDate(0, -6, 0)

	suite.HeartbeatService.On("DeleteByUserBefore", user, before).Return(nil).Once()
	suite.DurationService.On("DeleteByUserBefore", user, before).Return(nil).Once()
	suite.SummaryService.On("DeleteByUserBefore", user.ID, before).Return(nil).Once()

	err := sut.CleanUserDataBefore(user, before)

	assert.Nil(suite.T(), err)
	suite.HeartbeatService.AssertNumberOfCalls(suite.T(), "DeleteByUserBefore", 1)
	suite.HeartbeatService.AssertCalled(suite.T(), "DeleteByUserBefore", user, before)
	suite.DurationService.AssertNumberOfCalls(suite.T(), "DeleteByUserBefore", 1)
	suite.DurationService.AssertCalled(suite.T(), "DeleteByUserBefore", user, before)
	suite.SummaryService.AssertNumberOfCalls(suite.T(), "DeleteByUserBefore", 1)
	suite.SummaryService.AssertCalled(suite.T(), "DeleteByUserBefore", user.ID, before)
}
