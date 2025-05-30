package services

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

type AliasServiceTestSuite struct {
	suite.Suite
	TestUserId      string
	AliasRepository *mocks.AliasRepositoryMock
}

func (suite *AliasServiceTestSuite) SetupSuite() {
	config.Set(config.Empty())

	suite.TestUserId = "johndoe@example.org"

	aliases := []*models.Alias{
		{
			Type:   models.SummaryProject,
			UserID: suite.TestUserId,
			Key:    "wakapi",
			Value:  "wakapi-mobile",
		},
		{
			Type:   models.SummaryProject,
			UserID: suite.TestUserId,
			Key:    "telepush",
			Value:  "telepush-*",
		},
	}

	aliasRepoMock := new(mocks.AliasRepositoryMock)
	aliasRepoMock.On("GetByUser", suite.TestUserId).Return(aliases, nil)
	aliasRepoMock.On("GetByUser", mock.AnythingOfType("string")).Return([]*models.Alias{}, assert.AnError)

	suite.AliasRepository = aliasRepoMock
}

func TestAliasServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AliasServiceTestSuite))
}

func (suite *AliasServiceTestSuite) TestAliasService_GetAliasOrDefault() {
	sut := NewAliasService(suite.AliasRepository)

	result1, err1 := sut.GetAliasOrDefault(suite.TestUserId, models.SummaryProject, "wakapi-mobile")
	result2, err2 := sut.GetAliasOrDefault(suite.TestUserId, models.SummaryProject, "wakapi")
	result3, err3 := sut.GetAliasOrDefault(suite.TestUserId, models.SummaryProject, "anchr")
	result4, err4 := sut.GetAliasOrDefault(suite.TestUserId, models.SummaryProject, "telepush-mobile")
	result5, err5 := sut.GetAliasOrDefault(suite.TestUserId, models.SummaryEntity, "telepush-mobile")
	result6, err6 := sut.GetAliasOrDefault(suite.TestUserId, models.SummaryLanguage, "telepush-mobile")

	assert.Equal(suite.T(), "wakapi", result1)
	assert.Nil(suite.T(), err1)
	assert.Equal(suite.T(), "wakapi", result2)
	assert.Nil(suite.T(), err2)
	assert.Equal(suite.T(), "anchr", result3)
	assert.Nil(suite.T(), err3)
	assert.Equal(suite.T(), "telepush", result4)
	assert.Nil(suite.T(), err4)
	assert.Equal(suite.T(), "telepush-mobile", result5)
	assert.Nil(suite.T(), err5)
	assert.Equal(suite.T(), "Telepush-mobile", result6) // not really scope of this test, but nevertheless: language shall always be capitaliized
	assert.Nil(suite.T(), err6)
}
