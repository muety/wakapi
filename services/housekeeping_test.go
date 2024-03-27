package services

import (
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type HousekeepingServiceTestSuite struct {
	suite.Suite
	TestUsers        []*models.User
	UserService      *mocks.UserServiceMock
	HeartbeatService *mocks.HeartbeatServiceMock
	SummaryService   *mocks.SummaryServiceMock
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
	suite.SummaryService = new(mocks.SummaryServiceMock)
}

func TestHouseKeepingServiceTestSuite(t *testing.T) {
	suite.Run(t, new(HousekeepingServiceTestSuite))
}

func (suite *HousekeepingServiceTestSuite) TestHousekeepingService_CleanInactiveUsers() {
	sut := NewHousekeepingService(suite.UserService, suite.HeartbeatService, suite.SummaryService)

	suite.UserService.On("GetAll").Return(suite.TestUsers, nil)
	suite.UserService.On("Delete", suite.TestUsers[0]).Return(nil)

	err := sut.CleanInactiveUsers(time.Now().AddDate(0, -12, 0))

	assert.Nil(suite.T(), err)
	suite.UserService.AssertNumberOfCalls(suite.T(), "GetAll", 1)
	suite.UserService.AssertNumberOfCalls(suite.T(), "Delete", 1)
	suite.UserService.AssertCalled(suite.T(), "Delete", suite.TestUsers[0])
}
