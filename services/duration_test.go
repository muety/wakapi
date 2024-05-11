package services

import (
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"testing"
	"time"
)

const (
	TestUserId           = "muety"
	TestProject1         = "test-project-1"
	TestProject2         = "test-project-2"
	TestProject3         = "test-project-3"
	TestLanguageGo       = "Go"
	TestLanguageJava     = "Java"
	TestLanguagePython   = "Python"
	TestEditorGoland     = "GoLand"
	TestEditorIntellij   = "idea"
	TestEditorVscode     = "vscode"
	TestOsLinux          = "Linux"
	TestOsWin            = "Windows"
	TestMachine1         = "muety-desktop"
	TestMachine2         = "muety-work"
	TestEntity1          = "/home/bob/dev/wakapi.go"
	TestEntity2          = "/home/bob/dev/SomethingElse.java"
	TestBranchMaster     = "master"
	TestBranchDev        = "dev"
	TestCategoryCoding   = "coding"
	TestCategoryBrowsing = "browsing"
	MinUnixTime1         = 1601510400000 * 1e6
)

type DurationServiceTestSuite struct {
	suite.Suite
	TestUser         *models.User
	TestStartTime    time.Time
	TestHeartbeats   []*models.Heartbeat
	TestLabels       []*models.ProjectLabel
	HeartbeatService *mocks.HeartbeatServiceMock
}

func (suite *DurationServiceTestSuite) SetupSuite() {
	suite.TestUser = &models.User{ID: TestUserId}

	suite.TestStartTime = time.Unix(0, MinUnixTime1)
	suite.TestHeartbeats = []*models.Heartbeat{
		{
			ID:              rand.Uint64(),
			UserID:          TestUserId,
			Project:         TestProject1,
			Language:        TestLanguageGo,
			Editor:          TestEditorGoland,
			OperatingSystem: TestOsLinux,
			Machine:         TestMachine1,
			Time:            models.CustomTime(suite.TestStartTime), // 0:00
		},
		{
			ID:              rand.Uint64(),
			UserID:          TestUserId,
			Project:         TestProject1,
			Language:        TestLanguageGo,
			Editor:          TestEditorGoland,
			OperatingSystem: TestOsLinux,
			Machine:         TestMachine1,
			Time:            models.CustomTime(suite.TestStartTime.Add(30 * time.Second)), // 0:30
		},
		// duplicate of previous one
		{
			ID:              rand.Uint64(),
			UserID:          TestUserId,
			Project:         TestProject1,
			Language:        TestLanguageGo,
			Editor:          TestEditorGoland,
			OperatingSystem: TestOsLinux,
			Machine:         TestMachine1,
			Time:            models.CustomTime(suite.TestStartTime.Add(30 * time.Second)), // 0:30
		},
		{
			ID:              rand.Uint64(),
			UserID:          TestUserId,
			Project:         TestProject1,
			Language:        TestLanguageGo,
			Editor:          TestEditorGoland,
			OperatingSystem: TestOsLinux,
			Machine:         TestMachine1,
			Time:            models.CustomTime(suite.TestStartTime.Add((30 + 130) * time.Second)), // 2:40
		},
		{
			ID:              rand.Uint64(),
			UserID:          TestUserId,
			Project:         TestProject1,
			Language:        TestLanguageGo,
			Editor:          TestEditorVscode,
			OperatingSystem: TestOsLinux,
			Machine:         TestMachine1,
			Time:            models.CustomTime(suite.TestStartTime.Add(3 * time.Minute)), // 3:00
		},
		{
			ID:              rand.Uint64(),
			UserID:          TestUserId,
			Project:         TestProject1,
			Language:        TestLanguageGo,
			Editor:          TestEditorVscode,
			OperatingSystem: TestOsLinux,
			Machine:         TestMachine1,
			Time:            models.CustomTime(suite.TestStartTime.Add(3*time.Minute + 10*time.Second)), // 3:10
		},
		{
			ID:              rand.Uint64(),
			UserID:          TestUserId,
			Project:         TestProject1,
			Language:        TestLanguageGo,
			Editor:          TestEditorVscode,
			OperatingSystem: TestOsLinux,
			Machine:         TestMachine1,
			Time:            models.CustomTime(suite.TestStartTime.Add(3*time.Minute + 15*time.Second)), // 3:15
		},
	}
}

func (suite *DurationServiceTestSuite) BeforeTest(suiteName, testName string) {
	suite.HeartbeatService = new(mocks.HeartbeatServiceMock)
}

func TestDurationServiceTestSuite(t *testing.T) {
	suite.Run(t, new(DurationServiceTestSuite))
}

func (suite *DurationServiceTestSuite) TestDurationService_Get() {
	sut := NewDurationService(suite.HeartbeatService)

	var (
		from      time.Time
		to        time.Time
		durations models.Durations
		err       error
	)

	/* TEST 1 */
	from, to = suite.TestStartTime.Add(-1*time.Hour), suite.TestStartTime.Add(-1*time.Minute)
	suite.HeartbeatService.On("GetAllWithin", from, to, suite.TestUser).Return(filterHeartbeats(from, to, suite.TestHeartbeats), nil)

	durations, err = sut.Get(from, to, suite.TestUser, nil)

	assert.Nil(suite.T(), err)
	assert.Empty(suite.T(), durations)

	/* TEST 2 */
	from, to = suite.TestStartTime.Add(-1*time.Hour), suite.TestStartTime.Add(1*time.Second)
	suite.HeartbeatService.On("GetAllWithin", from, to, suite.TestUser).Return(filterHeartbeats(from, to, suite.TestHeartbeats), nil)

	durations, err = sut.Get(from, to, suite.TestUser, nil)

	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), durations, 1)
	assert.Equal(suite.T(), HeartbeatDiffThreshold, durations.First().Duration)
	assert.Equal(suite.T(), 1, durations.First().NumHeartbeats)

	/* TEST 3 */
	from, to = suite.TestStartTime, suite.TestStartTime.Add(1*time.Hour)
	suite.HeartbeatService.On("GetAllWithin", from, to, suite.TestUser).Return(filterHeartbeats(from, to, suite.TestHeartbeats), nil)

	durations, err = sut.Get(from, to, suite.TestUser, nil)

	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), durations, 3)
	assert.Equal(suite.T(), 150*time.Second, durations[0].Duration)
	assert.Equal(suite.T(), 20*time.Second, durations[1].Duration)
	assert.Equal(suite.T(), 15*time.Second, durations[2].Duration)
	assert.Equal(suite.T(), TestEditorGoland, durations[0].Editor)
	assert.Equal(suite.T(), TestEditorGoland, durations[1].Editor)
	assert.Equal(suite.T(), TestEditorVscode, durations[2].Editor)
	assert.Equal(suite.T(), 3, durations[0].NumHeartbeats)
	assert.Equal(suite.T(), 1, durations[1].NumHeartbeats)
	assert.Equal(suite.T(), 3, durations[2].NumHeartbeats)
}

func (suite *DurationServiceTestSuite) TestDurationService_Get_Filtered() {
	sut := NewDurationService(suite.HeartbeatService)

	var (
		from      time.Time
		to        time.Time
		durations models.Durations
		err       error
	)

	from, to = suite.TestStartTime.Add(-1*time.Hour), suite.TestStartTime.Add(1*time.Hour)
	suite.HeartbeatService.On("GetAllWithin", from, to, suite.TestUser).Return(filterHeartbeats(from, to, suite.TestHeartbeats), nil)

	durations, err = sut.Get(from, to, suite.TestUser, models.NewFiltersWith(models.SummaryEditor, TestEditorGoland))
	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), durations, 2)
	for _, d := range durations {
		assert.Equal(suite.T(), TestEditorGoland, d.Editor)
	}
}

func filterHeartbeats(from, to time.Time, heartbeats []*models.Heartbeat) []*models.Heartbeat {
	filtered := make([]*models.Heartbeat, 0, len(heartbeats))
	for _, h := range heartbeats {
		if (h.Time.T().Equal(from) || h.Time.T().After(from)) && h.Time.T().Before(to) {
			filtered = append(filtered, h)
		}
	}
	return filtered
}
