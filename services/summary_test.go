package services

import (
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"strings"
	"testing"
	"time"
)

const (
	TestUserId         = "muety"
	TestProject1       = "test-project-1"
	TestProject2       = "test-project-2"
	TestLanguageGo     = "Go"
	TestLanguageJava   = "Java"
	TestLanguagePython = "Python"
	TestEditorGoland   = "GoLand"
	TestEditorIntellij = "idea"
	TestEditorVscode   = "vscode"
	TestOsLinux        = "Linux"
	TestOsWin          = "Windows"
	TestMachine1       = "muety-desktop"
	TestMachine2       = "muety-work"
	MinUnixTime1       = 1601510400000 * 1e6
)

type SummaryServiceTestSuite struct {
	suite.Suite
	TestUser          *models.User
	TestStartTime     time.Time
	TestHeartbeats    []*models.Heartbeat
	SummaryRepository *mocks.SummaryRepositoryMock
	HeartbeatService  *mocks.HeartbeatServiceMock
	AliasService      *mocks.AliasServiceMock
}

func (suite *SummaryServiceTestSuite) SetupSuite() {
	suite.TestUser = &models.User{ID: TestUserId}

	suite.TestStartTime = time.Unix(0, MinUnixTime1)
	suite.TestHeartbeats = []*models.Heartbeat{
		{
			ID:              uint(rand.Uint32()),
			UserID:          TestUserId,
			Project:         TestProject1,
			Language:        TestLanguageGo,
			Editor:          TestEditorGoland,
			OperatingSystem: TestOsLinux,
			Machine:         TestMachine1,
			Time:            models.CustomTime(suite.TestStartTime),
		},
		{
			ID:              uint(rand.Uint32()),
			UserID:          TestUserId,
			Project:         TestProject1,
			Language:        TestLanguageGo,
			Editor:          TestEditorGoland,
			OperatingSystem: TestOsLinux,
			Machine:         TestMachine1,
			Time:            models.CustomTime(suite.TestStartTime.Add(30 * time.Second)),
		},
		{
			ID:              uint(rand.Uint32()),
			UserID:          TestUserId,
			Project:         TestProject1,
			Language:        TestLanguageGo,
			Editor:          TestEditorVscode,
			OperatingSystem: TestOsLinux,
			Machine:         TestMachine1,
			Time:            models.CustomTime(suite.TestStartTime.Add(3 * time.Minute)),
		},
	}
}

func (suite *SummaryServiceTestSuite) BeforeTest(suiteName, testName string) {
	suite.SummaryRepository = new(mocks.SummaryRepositoryMock)
	suite.HeartbeatService = new(mocks.HeartbeatServiceMock)
	suite.AliasService = new(mocks.AliasServiceMock)
}

func TestSummaryServiceTestSuite(t *testing.T) {
	suite.Run(t, new(SummaryServiceTestSuite))
}

func (suite *SummaryServiceTestSuite) TestSummaryService_Summarize() {
	sut := NewSummaryService(suite.SummaryRepository, suite.HeartbeatService, suite.AliasService)

	var (
		from   time.Time
		to     time.Time
		result *models.Summary
		err    error
	)

	/* TEST 1 */
	from, to = suite.TestStartTime.Add(-1*time.Hour), suite.TestStartTime.Add(-1*time.Minute)
	suite.HeartbeatService.On("GetAllWithin", from, to, suite.TestUser).Return(filter(from, to, suite.TestHeartbeats), nil)

	result, err = sut.Summarize(from, to, suite.TestUser)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), from, result.FromTime.T())
	assert.Equal(suite.T(), to, result.ToTime.T())
	assert.Zero(suite.T(), result.TotalTime())
	assert.Empty(suite.T(), result.Projects)

	/* TEST 2 */
	from, to = suite.TestStartTime.Add(-1*time.Hour), suite.TestStartTime.Add(1*time.Second)
	suite.HeartbeatService.On("GetAllWithin", from, to, suite.TestUser).Return(filter(from, to, suite.TestHeartbeats), nil)

	result, err = sut.Summarize(from, to, suite.TestUser)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), suite.TestHeartbeats[0].Time.T(), result.FromTime.T())
	assert.Equal(suite.T(), suite.TestHeartbeats[0].Time.T(), result.ToTime.T())
	assert.Zero(suite.T(), result.TotalTime())
	assertNumAllItems(suite.T(), 1, result, "")

	/* TEST 3 */
	from, to = suite.TestStartTime, suite.TestStartTime.Add(1*time.Hour)
	suite.HeartbeatService.On("GetAllWithin", from, to, suite.TestUser).Return(filter(from, to, suite.TestHeartbeats), nil)

	result, err = sut.Summarize(from, to, suite.TestUser)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), suite.TestHeartbeats[0].Time.T(), result.FromTime.T())
	assert.Equal(suite.T(), suite.TestHeartbeats[len(suite.TestHeartbeats)-1].Time.T(), result.ToTime.T())
	assert.Equal(suite.T(), 150*time.Second, result.TotalTime())
	assert.Equal(suite.T(), 30*time.Second, result.TotalTimeByKey(models.SummaryEditor, TestEditorGoland))
	assert.Equal(suite.T(), 120*time.Second, result.TotalTimeByKey(models.SummaryEditor, TestEditorVscode))
	assert.Len(suite.T(), result.Editors, 2)
	assertNumAllItems(suite.T(), 1, result, "e")
}

func (suite *SummaryServiceTestSuite) TestSummaryService_Retrieve() {
	sut := NewSummaryService(suite.SummaryRepository, suite.HeartbeatService, suite.AliasService)

	var (
		summaries []*models.Summary
		from      time.Time
		to        time.Time
		result    *models.Summary
		err       error
	)

	/* TEST 1 */
	from, to = suite.TestStartTime.Add(-12*time.Hour), suite.TestStartTime.Add(12*time.Hour)
	summaries = []*models.Summary{
		{
			ID:       uint(rand.Uint32()),
			UserID:   TestUserId,
			FromTime: models.CustomTime(from.Add(10 * time.Minute)),
			ToTime:   models.CustomTime(to.Add(-10 * time.Minute)),
			Projects: []*models.SummaryItem{
				{
					Type:  models.SummaryProject,
					Key:   TestProject1,
					Total: 45 * time.Minute / time.Second, // hack
				},
			},
			Languages:        []*models.SummaryItem{},
			Editors:          []*models.SummaryItem{},
			OperatingSystems: []*models.SummaryItem{},
			Machines:         []*models.SummaryItem{},
		},
	}

	suite.SummaryRepository.On("GetByUserWithin", suite.TestUser, from, to).Return(summaries, nil)
	suite.HeartbeatService.On("GetAllWithin", from, summaries[0].FromTime.T(), suite.TestUser).Return([]*models.Heartbeat{}, nil)
	suite.HeartbeatService.On("GetAllWithin", summaries[0].ToTime.T(), to, suite.TestUser).Return([]*models.Heartbeat{}, nil)

	result, err = sut.Retrieve(from, to, suite.TestUser)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Projects, 1)
	assert.Equal(suite.T(), summaries[0].Projects[0].Total*time.Second, result.TotalTime())
	suite.HeartbeatService.AssertNumberOfCalls(suite.T(), "GetAllWithin", 2)

	/* TEST 2 */
	from, to = suite.TestStartTime.Add(-10*time.Minute), suite.TestStartTime.Add(12*time.Hour)
	summaries = []*models.Summary{
		{
			ID:       uint(rand.Uint32()),
			UserID:   TestUserId,
			FromTime: models.CustomTime(from.Add(20 * time.Minute)),
			ToTime:   models.CustomTime(to.Add(-6 * time.Hour)),
			Projects: []*models.SummaryItem{
				{
					Type:  models.SummaryProject,
					Key:   TestProject1,
					Total: 45 * time.Minute / time.Second, // hack
				},
			},
			Languages:        []*models.SummaryItem{},
			Editors:          []*models.SummaryItem{},
			OperatingSystems: []*models.SummaryItem{},
			Machines:         []*models.SummaryItem{},
		},
		{
			ID:       uint(rand.Uint32()),
			UserID:   TestUserId,
			FromTime: models.CustomTime(to.Add(-6 * time.Hour)),
			ToTime:   models.CustomTime(to),
			Projects: []*models.SummaryItem{
				{
					Type:  models.SummaryProject,
					Key:   TestProject2,
					Total: 45 * time.Minute / time.Second, // hack
				},
			},
			Languages:        []*models.SummaryItem{},
			Editors:          []*models.SummaryItem{},
			OperatingSystems: []*models.SummaryItem{},
			Machines:         []*models.SummaryItem{},
		},
	}

	suite.SummaryRepository.On("GetByUserWithin", suite.TestUser, from, to).Return(summaries, nil)
	suite.HeartbeatService.On("GetAllWithin", from, summaries[0].FromTime.T(), suite.TestUser).Return(filter(from, summaries[0].FromTime.T(), suite.TestHeartbeats), nil)

	result, err = sut.Retrieve(from, to, suite.TestUser)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Projects, 2)
	assert.Equal(suite.T(), 150*time.Second+90*time.Minute, result.TotalTime())
	assert.Equal(suite.T(), 150*time.Second+45*time.Minute, result.TotalTimeByKey(models.SummaryProject, TestProject1))
	assert.Equal(suite.T(), 45*time.Minute, result.TotalTimeByKey(models.SummaryProject, TestProject2))
}

func (suite *SummaryServiceTestSuite) TestSummaryService_Aliased() {
	sut := NewSummaryService(suite.SummaryRepository, suite.HeartbeatService, suite.AliasService)

	var (
		from   time.Time
		to     time.Time
		result *models.Summary
		err    error
	)

	from, to = suite.TestStartTime, suite.TestStartTime.Add(1*time.Hour)
	suite.HeartbeatService.On("GetAllWithin", from, to, suite.TestUser).Return(filter(from, to, suite.TestHeartbeats), nil)
	suite.AliasService.On("InitializeUser", TestUserId).Return(nil)
	suite.AliasService.On("GetAliasOrDefault", TestUserId, models.SummaryProject, TestProject1).Return(TestProject2, nil)
	suite.AliasService.On("GetAliasOrDefault", TestUserId, mock.Anything, mock.Anything).Return("", nil)

	result, err = sut.Aliased(from, to, suite.TestUser, sut.Summarize)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Zero(suite.T(), result.TotalTimeByKey(models.SummaryProject, TestProject1))
	assert.NotZero(suite.T(), result.TotalTimeByKey(models.SummaryProject, TestProject2))
}

func filter(from, to time.Time, heartbeats []*models.Heartbeat) []*models.Heartbeat {
	filtered := make([]*models.Heartbeat, 0, len(heartbeats))
	for _, h := range heartbeats {
		if (h.Time.T().Equal(from) || h.Time.T().After(from)) && h.Time.T().Before(to) {
			filtered = append(filtered, h)
		}
	}
	return filtered
}

func assertNumAllItems(t *testing.T, expected int, summary *models.Summary, except string) {
	if !strings.Contains(except, "p") {
		assert.Len(t, summary.Projects, expected)
	}
	if !strings.Contains(except, "e") {
		assert.Len(t, summary.Editors, expected)
	}
	if !strings.Contains(except, "l") {
		assert.Len(t, summary.Languages, expected)
	}
	if !strings.Contains(except, "o") {
		assert.Len(t, summary.OperatingSystems, expected)
	}
	if !strings.Contains(except, "m") {
		assert.Len(t, summary.Machines, expected)
	}
}
