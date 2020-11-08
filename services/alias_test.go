package services

import (
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

type AliasServiceTestSuite struct {
	suite.Suite
	TestUserId    string
	AliasRepoMock *mocks.AliasRepositoryMock
}

func (suite *AliasServiceTestSuite) SetupSuite() {
	suite.TestUserId = "johndoe@example.org"

	aliases := []*models.Alias{
		{
			Type:   models.SummaryProject,
			UserID: suite.TestUserId,
			Key:    "wakapi",
			Value:  "wakapi-mobile",
		},
	}

	aliasRepoMock := new(mocks.AliasRepositoryMock)
	aliasRepoMock.On("GetByUser", suite.TestUserId).Return(aliases, nil)
	aliasRepoMock.On("GetByUser", mock.AnythingOfType("string")).Return([]*models.Alias{}, assert.AnError)

	suite.AliasRepoMock = aliasRepoMock
}

func TestAliasServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AliasServiceTestSuite))
}

func (suite *AliasServiceTestSuite) TestAliasService_GetAliasOrDefault() {
	sut := NewAliasService(suite.AliasRepoMock)

	result1, err1 := sut.GetAliasOrDefault(suite.TestUserId, models.SummaryProject, "wakapi-mobile")
	result2, err2 := sut.GetAliasOrDefault(suite.TestUserId, models.SummaryProject, "wakapi")
	result3, err3 := sut.GetAliasOrDefault(suite.TestUserId, models.SummaryProject, "anchr")

	assert.Equal(suite.T(), "wakapi", result1)
	assert.Nil(suite.T(), err1)
	assert.Equal(suite.T(), "wakapi", result2)
	assert.Nil(suite.T(), err2)
	assert.Equal(suite.T(), "anchr", result3)
	assert.Nil(suite.T(), err3)
}

func (suite *AliasServiceTestSuite) TestAliasService_GetAliasOrDefault_ErrorOnNonExistingUser() {
	sut := NewAliasService(suite.AliasRepoMock)

	result, err := sut.GetAliasOrDefault("nonexisting", models.SummaryProject, "wakapi-mobile")

	assert.Empty(suite.T(), result)
	assert.Error(suite.T(), err)
}
