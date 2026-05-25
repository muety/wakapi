package services

import (
	"fmt"
	"testing"
	"time"

	datastructure "github.com/duke-git/lancet/v2/datastructure/set"
	"github.com/leandro-lugaresi/hub"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	TestUserId2  = "testuser"
	MinUnixTime2 = 1609459200000 * 1e6 // 2021-01-01 00:00:00 UTC
)

type ProjectServiceTestSuite struct {
	suite.Suite
	TestUser            *models.User
	HeartbeatRepository *mocks.HeartbeatRepositoryMock
	AliasService        *mocks.AliasServiceMock
	HeartbeatService    *mocks.HeartbeatServiceMock
}

func (suite *ProjectServiceTestSuite) SetupSuite() {
	suite.TestUser = &models.User{ID: TestUserId2}
}

func (suite *ProjectServiceTestSuite) BeforeTest(suiteName, testName string) {
	suite.HeartbeatRepository = new(mocks.HeartbeatRepositoryMock)
	suite.AliasService = new(mocks.AliasServiceMock)
	suite.HeartbeatService = new(mocks.HeartbeatServiceMock)
}

func (suite *ProjectServiceTestSuite) createSut() (*ProjectService, *hub.Hub) {
	originalEventBus := config.EventBus()
	eventBus := hub.New()
	config.SetEventBus(eventBus)
	// restore event bus
	sut := NewProjectService(suite.AliasService, suite.HeartbeatRepository, suite.HeartbeatService)
	config.SetEventBus(originalEventBus)
	return sut, eventBus
}

func TestProjectServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ProjectServiceTestSuite))
}

func (suite *ProjectServiceTestSuite) TestProjectService_GetUserProjectStats_Pagination() {
	sut, _ := suite.createSut()
	now := time.Unix(0, MinUnixTime2)

	rawStats := []*models.ProjectStats{
		{Project: "project1", Count: 10, Last: models.CustomTime(now.Add(5 * time.Hour))},
		{Project: "project2", Count: 20, Last: models.CustomTime(now.Add(4 * time.Hour))},
		{Project: "project3", Count: 30, Last: models.CustomTime(now.Add(3 * time.Hour))},
		{Project: "project4", Count: 40, Last: models.CustomTime(now.Add(2 * time.Hour))},
		{Project: "project5", Count: 50, Last: models.CustomTime(now.Add(1 * time.Hour))},
	}

	suite.HeartbeatRepository.On("GetUserProjectStats", suite.TestUser, mock.Anything, mock.Anything).Return(rawStats, nil).Once()
	suite.AliasService.On("GetAliasOrDefault", suite.TestUser.ID, models.SummaryProject, "project1").Return("project1", nil)
	suite.AliasService.On("GetAliasOrDefault", suite.TestUser.ID, models.SummaryProject, "project2").Return("project2", nil)
	suite.AliasService.On("GetAliasOrDefault", suite.TestUser.ID, models.SummaryProject, "project3").Return("project3", nil)
	suite.AliasService.On("GetAliasOrDefault", suite.TestUser.ID, models.SummaryProject, "project4").Return("project4", nil)
	suite.AliasService.On("GetAliasOrDefault", suite.TestUser.ID, models.SummaryProject, "project5").Return("project5", nil)
	suite.AliasService.On("MayInitializeUser", suite.TestUser.ID).Return()
	suite.HeartbeatService.On("GetEntitySetByUser", models.SummaryProject, suite.TestUser.ID).Return([]string{}, nil)

	pageParams := &utils.PageParams{Page: 2, PageSize: 2} // limit 2, offset 2

	results, err := sut.GetUserProjectStats(suite.TestUser, time.Time{}, time.Now(), "", pageParams, false)

	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), results, 2)
	assert.Equal(suite.T(), "project3", results[0].Project)
	assert.Equal(suite.T(), "project4", results[1].Project)
}

func (suite *ProjectServiceTestSuite) TestProjectService_GetUserProjectStats_AliasedAndMerged() {
	sut, _ := suite.createSut()
	now := time.Unix(0, MinUnixTime2)

	rawStats := []*models.ProjectStats{
		{Project: "wakapi-web", Count: 5, First: models.CustomTime(now.Add(1 * time.Hour)), Last: models.CustomTime(now.Add(5 * time.Hour)), TopLanguage: "Javascript"},
		{Project: "wakapi-mobile", Count: 10, First: models.CustomTime(now.Add(2 * time.Hour)), Last: models.CustomTime(now.Add(3 * time.Hour)), TopLanguage: "Dart"},
		{Project: "other-project", Count: 2, First: models.CustomTime(now.Add(2 * time.Hour)), Last: models.CustomTime(now.Add(4 * time.Hour)), TopLanguage: "Go"},
	}

	suite.HeartbeatRepository.On("GetUserProjectStats", suite.TestUser, mock.Anything, mock.Anything).Return(rawStats, nil).Once()
	suite.AliasService.On("MayInitializeUser", suite.TestUser.ID).Return()
	suite.AliasService.On("GetAliasOrDefault", suite.TestUser.ID, models.SummaryProject, "wakapi-web").Return("wakapi", nil)
	suite.AliasService.On("GetAliasOrDefault", suite.TestUser.ID, models.SummaryProject, "wakapi-mobile").Return("wakapi", nil)
	suite.AliasService.On("GetAliasOrDefault", suite.TestUser.ID, models.SummaryProject, "other-project").Return("other-project", nil)
	suite.HeartbeatService.On("GetEntitySetByUser", models.SummaryProject, suite.TestUser.ID).Return([]string{}, nil)

	results, err := sut.GetUserProjectStats(suite.TestUser, time.Time{}, time.Now(), "", nil, false)

	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), results, 2)

	// because sorted by Last desc: wakapi should be first (last was 5*time.Hour), other-project should be second (last was 4*time.Hour)
	assert.Equal(suite.T(), "wakapi", results[0].Project)
	assert.Equal(suite.T(), int64(15), results[0].Count)                               // 5 + 10
	assert.Equal(suite.T(), models.CustomTime(now.Add(1*time.Hour)), results[0].First) // earliest
	assert.Equal(suite.T(), models.CustomTime(now.Add(5*time.Hour)), results[0].Last)  // latest
	assert.Equal(suite.T(), "Dart", results[0].TopLanguage)                            // higher count

	assert.Equal(suite.T(), "other-project", results[1].Project)
}

func (suite *ProjectServiceTestSuite) TestProjectService_GetUserProjectStats_Search() {
	sut, _ := suite.createSut()
	now := time.Unix(0, MinUnixTime2)

	rawStats := []*models.ProjectStats{
		{Project: "wakapi-web", Count: 5, First: models.CustomTime(now.Add(1 * time.Hour)), Last: models.CustomTime(now.Add(5 * time.Hour)), TopLanguage: "Javascript"},
		{Project: "wakapi-mobile", Count: 10, First: models.CustomTime(now.Add(2 * time.Hour)), Last: models.CustomTime(now.Add(3 * time.Hour)), TopLanguage: "Dart"},
		{Project: "other-project", Count: 2, First: models.CustomTime(now.Add(2 * time.Hour)), Last: models.CustomTime(now.Add(4 * time.Hour)), TopLanguage: "Go"},
	}

	suite.HeartbeatRepository.On("GetUserProjectStats", suite.TestUser, mock.Anything, mock.Anything).Return(rawStats, nil).Once()
	suite.AliasService.On("MayInitializeUser", suite.TestUser.ID).Return()
	suite.AliasService.On("GetAliasOrDefault", suite.TestUser.ID, models.SummaryProject, "wakapi-web").Return("wakapi", nil)
	suite.AliasService.On("GetAliasOrDefault", suite.TestUser.ID, models.SummaryProject, "wakapi-mobile").Return("wakapi", nil)
	suite.AliasService.On("GetAliasOrDefault", suite.TestUser.ID, models.SummaryProject, "other-project").Return("other-project", nil)
	suite.HeartbeatService.On("GetEntitySetByUser", models.SummaryProject, suite.TestUser.ID).Return([]string{}, nil)

	results, err := sut.GetUserProjectStats(suite.TestUser, time.Time{}, time.Now(), "WaKa", nil, false)

	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), results, 1)
	assert.Equal(suite.T(), "wakapi", results[0].Project)
}

func (suite *ProjectServiceTestSuite) TestProjectService_EventHeartbeatCreate_InvalidatesCache() {
	sut, eventBus := suite.createSut()

	cacheKeyStats := fmt.Sprintf("project_stats_%s_some_suffix", suite.TestUser.ID)
	cacheKeyProjects := fmt.Sprintf("unique_projects_%s", suite.TestUser.ID)

	// pre-populate cache
	sut.cache.Set(cacheKeyStats, "some_data", 5*time.Minute)
	// we need to simulate the set being there but not containing the new project
	uniqueSet := datastructure.New[string]("old-project")
	sut.cache.Set(cacheKeyProjects, uniqueSet, 5*time.Minute)

	heartbeat := &models.Heartbeat{
		UserID:  suite.TestUser.ID,
		Project: "new-project",
	}

	eventBus.Publish(hub.Message{
		Name: config.EventHeartbeatCreate,
		Fields: map[string]interface{}{
			config.FieldPayload: heartbeat,
		},
	})

	assert.Eventually(suite.T(), func() bool {
		_, foundStats := sut.cache.Get(cacheKeyStats)
		_, foundProjects := sut.cache.Get(cacheKeyProjects)
		return !foundStats && !foundProjects
	}, 2*time.Second, 20*time.Millisecond)
}

func (suite *ProjectServiceTestSuite) TestProjectService_EventAliasChanged_InvalidatesCache() {
	sut, eventBus := suite.createSut()

	cacheKeyStats := fmt.Sprintf("project_stats_%s_some_suffix", suite.TestUser.ID)
	cacheKeyProjects := fmt.Sprintf("unique_projects_%s", suite.TestUser.ID)

	// pre-populate cache
	sut.cache.Set(cacheKeyStats, "some_data", 5*time.Minute)
	uniqueSet := datastructure.New[string]("some-project")
	sut.cache.Set(cacheKeyProjects, uniqueSet, 5*time.Minute)

	eventBus.Publish(hub.Message{
		Name: config.EventAliasCreate,
		Fields: map[string]interface{}{
			config.FieldUserId: suite.TestUser.ID,
		},
	})

	assert.Eventually(suite.T(), func() bool {
		_, foundStats := sut.cache.Get(cacheKeyStats)
		_, foundProjects := sut.cache.Get(cacheKeyProjects)
		return !foundStats && !foundProjects
	}, 2*time.Second, 20*time.Millisecond)
}
