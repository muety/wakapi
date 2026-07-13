package services

import (
	"errors"
	"testing"

	"github.com/leandro-lugaresi/hub"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type HeartbeatServiceTestSuite struct {
	suite.Suite
	HeartbeatRepository *mocks.HeartbeatRepositoryMock
}

func (suite *HeartbeatServiceTestSuite) BeforeTest(suiteName, testName string) {
	suite.HeartbeatRepository = new(mocks.HeartbeatRepositoryMock)
}

func (suite *HeartbeatServiceTestSuite) createSut() *HeartbeatService {
	return NewHeartbeatService(suite.HeartbeatRepository, nil)
}

func TestHeartbeatServiceTestSuite(t *testing.T) {
	config.Set(config.Empty())
	config.SetEventBus(hub.New())
	suite.Run(t, new(HeartbeatServiceTestSuite))
}

func (suite *HeartbeatServiceTestSuite) TestHeartbeatService_SearchBranchesByUser_ForwardsArgs() {
	sut := suite.createSut()

	suite.HeartbeatRepository.On("SearchBranchesByUser", TestUserId2, "myproject", "mai", 50).Return([]string{"main"}, nil)

	result, err := sut.SearchBranchesByUser(TestUserId2, "myproject", "mai", 50)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), []string{"main"}, result)
	suite.HeartbeatRepository.AssertExpectations(suite.T())
}

func (suite *HeartbeatServiceTestSuite) TestHeartbeatService_SearchBranchesByUser_PropagatesError() {
	sut := suite.createSut()

	suite.HeartbeatRepository.On("SearchBranchesByUser", TestUserId2, "", "mai", 50).Return([]string(nil), errors.New("boom"))

	result, err := sut.SearchBranchesByUser(TestUserId2, "", "mai", 50)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	suite.HeartbeatRepository.AssertExpectations(suite.T())
}

func (suite *HeartbeatServiceTestSuite) TestHeartbeatService_SearchBranchesByUser_FiltersWhitespaceOnly() {
	sut := suite.createSut()

	suite.HeartbeatRepository.On("SearchBranchesByUser", TestUserId2, "", "mai", 50).Return([]string{"main", "   ", "", "maintenance"}, nil)

	result, err := sut.SearchBranchesByUser(TestUserId2, "", "mai", 50)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), []string{"main", "maintenance"}, result)
}
