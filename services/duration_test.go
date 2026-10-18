package services

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/leandro-lugaresi/hub"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

const (
	TestUserId         = "muety"
	TestProject1       = "test-project-1"
	TestProject2       = "test-project-2"
	TestProject3       = "test-project-3"
	TestProject4       = "something-completely-different-4"
	TestLanguageGo     = "Go"
	TestLanguageJava   = "Java"
	TestLanguagePython = "Python"
	TestEditorGoland   = "GoLand"
	TestEditorIntellij = "idea"
	TestEditorVscode   = "vscode"
	// TODO: add test cases involving "ai coding" category and a editor (harness) + model combination
	TestAiModelClaude    = "claude-3-5-sonnet"
	TestAiModelGpt       = "gpt-4o"
	TestOsLinux          = "Linux"
	TestOsWin            = "Windows"
	TestMachine1         = "muety-desktop"
	TestMachine2         = "muety-work"
	TestEntity1          = "/home/bob/dev/wakapi.go"
	TestEntity2          = "/home/bob/dev/config.go"
	TestEntity3          = "/home/bob/dev/SomethingElse.java"
	TestBranchMaster     = "master"
	TestBranchDev        = "dev"
	TestCategoryCoding   = "coding"
	TestCategoryBrowsing = "browsing"
	MinUnixTime1         = 1601510400000 * 1e6
)

type DurationServiceTestSuite struct {
	suite.Suite
	TestUser               *models.User
	TestStartTime          time.Time
	TestHeartbeats         []*models.Heartbeat
	TestLabels             []*models.ProjectLabel
	DurationRepository     *mocks.DurationRepositoryMock
	HeartbeatService       *mocks.HeartbeatServiceMock
	UserService            *mocks.UserServiceMock
	LanguageMappingService *mocks.LanguageMappingServiceMock
}

func (suite *DurationServiceTestSuite) SetupSuite() {
	suite.TestUser = &models.User{ID: TestUserId, HeartbeatsTimeoutSec: int(models.DefaultHeartbeatsTimeoutLegacy / time.Second)}

	// https:// anchr.io/i/F0HEK.jpg
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
			Entity:          TestEntity1,
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
			Entity:          TestEntity2,
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
			Entity:          TestEntity2,
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
			Entity:          TestEntity2,
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
			Entity:          TestEntity2,
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
			Entity:          TestEntity2,
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
			Entity:          TestEntity2,
			Time:            models.CustomTime(suite.TestStartTime.Add(3*time.Minute + 15*time.Second)), // 3:15
		},
	}
}

func (suite *DurationServiceTestSuite) BeforeTest(suiteName, testName string) {
	suite.DurationRepository = new(mocks.DurationRepositoryMock)
	suite.HeartbeatService = new(mocks.HeartbeatServiceMock)
	suite.UserService = new(mocks.UserServiceMock)
	suite.LanguageMappingService = new(mocks.LanguageMappingServiceMock)

	suite.LanguageMappingService.On("ResolveByUser", suite.TestUser.ID).Return(make(map[string]string), nil)
}

func TestDurationServiceTestSuite(t *testing.T) {
	suite.Run(t, new(DurationServiceTestSuite))
}

func (suite *DurationServiceTestSuite) TestDurationService_Get() {
	// https:// anchr.io/i/F0HEK.jpg
	sut := NewDurationService(suite.DurationRepository, suite.HeartbeatService, suite.UserService, suite.LanguageMappingService)

	var (
		from      time.Time
		to        time.Time
		durations models.Durations
		err       error
	)

	/* TEST 1 */
	from, to = suite.TestStartTime.Add(-1*time.Hour), suite.TestStartTime.Add(-1*time.Minute)
	suite.HeartbeatService.On("StreamAllWithinExcludingHeartbeats", from, to, suite.TestUser, models.ExcludeFromDurations).Return(streamSlice(filterHeartbeats(from, to, suite.TestHeartbeats)), nil)

	durations, err = sut.Get(from, to, suite.TestUser, nil, nil, true)

	assert.Nil(suite.T(), err)
	assert.Empty(suite.T(), durations)

	/* TEST 2 */
	from, to = suite.TestStartTime.Add(-1*time.Hour), suite.TestStartTime.Add(1*time.Second)
	suite.HeartbeatService.On("StreamAllWithinExcludingHeartbeats", from, to, suite.TestUser, models.ExcludeFromDurations).Return(streamSlice(filterHeartbeats(from, to, suite.TestHeartbeats)), nil)

	durations, err = sut.Get(from, to, suite.TestUser, nil, nil, true)

	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), durations, 1)
	assert.Equal(suite.T(), time.Duration(0), durations.First().Duration)
	assert.Equal(suite.T(), 1, durations.First().NumHeartbeats)

	/* TEST 3 */
	from, to = suite.TestStartTime, suite.TestStartTime.Add(1*time.Hour)
	suite.HeartbeatService.On("StreamAllWithinExcludingHeartbeats", from, to, suite.TestUser, models.ExcludeFromDurations).Return(streamSlice(filterHeartbeats(from, to, suite.TestHeartbeats)), nil)

	durations, err = sut.Get(from, to, suite.TestUser, nil, nil, true)

	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), durations, 3)
	assert.Equal(suite.T(), 30*time.Second, durations[0].Duration)
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
	sut := NewDurationService(suite.DurationRepository, suite.HeartbeatService, suite.UserService, suite.LanguageMappingService)

	var (
		from      time.Time
		to        time.Time
		durations models.Durations
		err       error
	)

	from, to = suite.TestStartTime.Add(-1*time.Hour), suite.TestStartTime.Add(1*time.Hour)
	suite.HeartbeatService.On("StreamAllWithinExcludingHeartbeats", from, to, suite.TestUser, models.ExcludeFromDurations).Return(streamSlice(filterHeartbeats(from, to, suite.TestHeartbeats)), nil)

	durations, err = sut.Get(from, to, suite.TestUser, models.NewFiltersWith(models.SummaryEditor, TestEditorGoland), nil, true)
	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), durations, 2)
	assert.Equal(suite.T(), 30*time.Second, durations[0].Duration)
	assert.Equal(suite.T(), 20*time.Second, durations[1].Duration)
	for _, d := range durations {
		assert.Equal(suite.T(), TestEditorGoland, d.Editor)
	}
}

func (suite *DurationServiceTestSuite) TestDurationService_Get_Filtered_AiModel() {
	sut := NewDurationService(suite.DurationRepository, suite.HeartbeatService, suite.UserService, suite.LanguageMappingService)

	h1 := &models.Heartbeat{
		UserID:   TestUserId,
		Project:  TestProject1,
		Language: TestLanguageGo,
		AIModel:  TestAiModelClaude,
		Time:     models.CustomTime(suite.TestStartTime),
	}
	h2 := &models.Heartbeat{
		UserID:   TestUserId,
		Project:  TestProject1,
		Language: TestLanguageGo,
		AIModel:  TestAiModelGpt,
		Time:     models.CustomTime(suite.TestStartTime.Add(5 * time.Minute)),
	}

	from, to := suite.TestStartTime.Add(-1*time.Hour), suite.TestStartTime.Add(1*time.Hour)
	suite.HeartbeatService.On("StreamAllWithinExcludingHeartbeats", from, to, suite.TestUser, models.ExcludeFromDurations).Return(streamSlice([]*models.Heartbeat{h1, h2}), nil)

	durations, err := sut.Get(from, to, suite.TestUser, models.NewFiltersWith(models.SummaryAiModel, TestAiModelClaude), nil, true)
	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), durations, 1)
	assert.Equal(suite.T(), TestAiModelClaude, durations[0].AIModel)
}

func (suite *DurationServiceTestSuite) TestDurationService_Get_ProjectDetails() {
	// https:// github.com/muety/wakapi/issues/876
	sut := NewDurationService(suite.DurationRepository, suite.HeartbeatService, suite.UserService, suite.LanguageMappingService)

	var (
		from      time.Time
		to        time.Time
		durations models.Durations
		err       error
	)

	from, to = suite.TestStartTime.Add(-1*time.Hour), suite.TestStartTime.Add(1*time.Hour)
	suite.HeartbeatService.On("StreamAllWithinExcludingHeartbeats", from, to, suite.TestUser, models.ExcludeFromDurations).Return(streamSlice(filterHeartbeats(from, to, suite.TestHeartbeats)), nil)

	testFilters := models.NewFiltersWith(models.SummaryEditor, TestEditorGoland).With(models.SummaryProject, TestProject1)
	durations, err = sut.Get(from, to, suite.TestUser, testFilters, nil, true)
	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), durations, 3)
	assert.Equal(suite.T(), TestEntity1, durations[0].Entity) // first duration is split up into two parts, because of different filenames when requesting project details
	assert.Equal(suite.T(), TestEntity2, durations[1].Entity)
	assert.Equal(suite.T(), TestEntity2, durations[2].Entity)
}

func (suite *DurationServiceTestSuite) TestDurationService_Get_CustomTimeout() {
	sut := NewDurationService(suite.DurationRepository, suite.HeartbeatService, suite.UserService, suite.LanguageMappingService)

	var (
		from      time.Time
		to        time.Time
		durations models.Durations
	)

	defer func() {
		suite.TestUser.HeartbeatsTimeoutSec = int(models.DefaultHeartbeatsTimeoutLegacy / time.Second) // revert to defaults
	}()

	from, to = suite.TestStartTime, suite.TestStartTime.Add(1*time.Hour)

	/* Test 1 */
	call1 := suite.HeartbeatService.On("StreamAllWithinExcludingHeartbeats", from, to, suite.TestUser, models.ExcludeFromDurations).Return(streamSlice(filterHeartbeats(from, to, suite.TestHeartbeats)), nil)
	suite.TestUser.HeartbeatsTimeoutSec = 60
	durations, _ = sut.Get(from, to, suite.TestUser, nil, nil, true)

	assert.Len(suite.T(), durations, 3)
	assert.Equal(suite.T(), 30*time.Second, durations[0].Duration)
	assert.Equal(suite.T(), 20*time.Second, durations[1].Duration)
	assert.Equal(suite.T(), 15*time.Second, durations[2].Duration)
	assert.Equal(suite.T(), 3, durations[0].NumHeartbeats)
	assert.Equal(suite.T(), 1, durations[1].NumHeartbeats)
	assert.Equal(suite.T(), 3, durations[2].NumHeartbeats)
	call1.Unset()

	/* Test 2 */
	call2 := suite.HeartbeatService.On("StreamAllWithinExcludingHeartbeats", from, to, suite.TestUser, models.ExcludeFromDurations).Return(streamSlice(filterHeartbeats(from, to, suite.TestHeartbeats)), nil)
	suite.TestUser.HeartbeatsTimeoutSec = 130
	durations, _ = sut.Get(from, to, suite.TestUser, nil, nil, true)

	assert.Len(suite.T(), durations, 3)
	assert.Equal(suite.T(), 30*time.Second, durations[0].Duration)
	assert.Equal(suite.T(), 20*time.Second, durations[1].Duration)
	assert.Equal(suite.T(), 15*time.Second, durations[2].Duration)
	assert.Equal(suite.T(), 3, durations[0].NumHeartbeats)
	assert.Equal(suite.T(), 1, durations[1].NumHeartbeats)
	assert.Equal(suite.T(), 3, durations[2].NumHeartbeats)
	call2.Unset()

	/* Test 3 */
	call3 := suite.HeartbeatService.On("StreamAllWithinExcludingHeartbeats", from, to, suite.TestUser, models.ExcludeFromDurations).Return(streamSlice(filterHeartbeats(from, to, suite.TestHeartbeats)), nil)
	suite.TestUser.HeartbeatsTimeoutSec = 140
	durations, _ = sut.Get(from, to, suite.TestUser, nil, nil, true)

	assert.Len(suite.T(), durations, 2)
	assert.Equal(suite.T(), 180*time.Second, durations[0].Duration)
	assert.Equal(suite.T(), 15*time.Second, durations[1].Duration)
	assert.Equal(suite.T(), 4, durations[0].NumHeartbeats)
	assert.Equal(suite.T(), 3, durations[1].NumHeartbeats)
	call3.Unset()
}

func (suite *DurationServiceTestSuite) TestDurationService_Get_Cached() {
	sut := NewDurationService(suite.DurationRepository, suite.HeartbeatService, suite.UserService, suite.LanguageMappingService)

	var (
		from      time.Time
		to        time.Time
		toCached  time.Time
		durations models.Durations
		err       error
	)

	testDurations := []*models.Duration{
		models.NewDurationFromHeartbeat(suite.TestHeartbeats[0]),
		models.NewDurationFromHeartbeat(suite.TestHeartbeats[3]),
		models.NewDurationFromHeartbeat(suite.TestHeartbeats[4]),
	}
	testDurations[0].Duration = 30 * time.Second
	testDurations[0].NumHeartbeats = 3
	testDurations[0].WithEntityIgnored().Hashed()
	testDurations[1].Duration = 20 * time.Second
	testDurations[1].NumHeartbeats = 1
	testDurations[1].WithEntityIgnored().Hashed()
	testDurations[2].Duration = 10 * time.Second
	testDurations[2].NumHeartbeats = 2
	testDurations[2].WithEntityIgnored().Hashed()

	from, to, toCached = suite.TestStartTime, suite.TestStartTime.Add(1*time.Hour), testDurations[2].TimeEnd().Add(time.Second)
	suite.DurationRepository.On("GetAllWithinByFilters", from, to, suite.TestUser, mock.Anything).Return(testDurations, nil)
	suite.HeartbeatService.On("StreamAllWithinExcludingHeartbeats", toCached, to, suite.TestUser, models.ExcludeFromDurations).Return(streamSlice(filterHeartbeats(toCached, to, suite.TestHeartbeats)), nil)

	durations, err = sut.Get(from, to, suite.TestUser, nil, nil, false)

	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), durations, 3)
	assert.Equal(suite.T(), 30*time.Second, durations[0].Duration)
	assert.Equal(suite.T(), 20*time.Second, durations[1].Duration)
	assert.Equal(suite.T(), 15*time.Second, durations[2].Duration)
	assert.Equal(suite.T(), 3, durations[0].NumHeartbeats)
	assert.Equal(suite.T(), 1, durations[1].NumHeartbeats)
	assert.Equal(suite.T(), 3, durations[2].NumHeartbeats)
}

func (suite *DurationServiceTestSuite) TestDurationService_Get_Cached_EmptyCache_ZeroFromTime_ComputesLive() {
	sut := NewDurationService(suite.DurationRepository, suite.HeartbeatService, suite.UserService, suite.LanguageMappingService)

	var (
		from      = time.Time{}
		to        = suite.TestStartTime.Add(1 * time.Hour)
		durations models.Durations
		err       error
	)

	suite.DurationRepository.On("GetAllWithinByFilters", from, to, suite.TestUser, mock.Anything).Return([]*models.Duration{}, nil)
	suite.HeartbeatService.On("StreamAllWithinExcludingHeartbeats", from, to, suite.TestUser, models.ExcludeFromDurations).Return(streamSlice(filterHeartbeats(from, to, suite.TestHeartbeats)), nil)

	durations, err = sut.Get(from, to, suite.TestUser, nil, nil, false)

	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), durations, 3)
	assert.Equal(suite.T(), 30*time.Second, durations[0].Duration)
	assert.Equal(suite.T(), 20*time.Second, durations[1].Duration)
	assert.Equal(suite.T(), 15*time.Second, durations[2].Duration)
	assert.Equal(suite.T(), 3, durations[0].NumHeartbeats)
	assert.Equal(suite.T(), 1, durations[1].NumHeartbeats)
	assert.Equal(suite.T(), 3, durations[2].NumHeartbeats)
	suite.HeartbeatService.AssertNumberOfCalls(suite.T(), "StreamAllWithinExcludingHeartbeats", 1)
}

func (suite *DurationServiceTestSuite) TestDurationService_Regenerate_SkipsUsersWithoutDurationsAndHeartbeats() {
	sut := NewDurationService(suite.DurationRepository, suite.HeartbeatService, suite.UserService, suite.LanguageMappingService)

	ghostUser := &models.User{ID: "ghost"}

	suite.DurationRepository.On("GetLatestByUser", suite.TestUser).Return(nil, gorm.ErrRecordNotFound)
	suite.HeartbeatService.On("GetLatestByUser", suite.TestUser, false).Return(&models.Heartbeat{UserID: TestUserId}, nil)
	suite.DurationRepository.On("GetLatestByUser", ghostUser).Return(nil, gorm.ErrRecordNotFound)
	suite.HeartbeatService.On("GetLatestByUser", ghostUser, false).Return((*models.Heartbeat)(nil), nil)

	sut.Regenerate(suite.TestUser, false)
	sut.Regenerate(ghostUser, false)

	suite.DurationRepository.AssertNotCalled(suite.T(), "GetAllWithinByFilters", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	suite.HeartbeatService.AssertNotCalled(suite.T(), "StreamAllWithinExcludingHeartbeats", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	suite.DurationRepository.AssertNotCalled(suite.T(), "InsertBatch", mock.Anything)
	suite.DurationRepository.AssertNotCalled(suite.T(), "DeleteByUser", mock.Anything)
}

func (suite *DurationServiceTestSuite) TestDurationService_Regenerate_PersistsInitialDurationsForUserWithoutDurations() {
	sut := NewDurationService(suite.DurationRepository, suite.HeartbeatService, suite.UserService, suite.LanguageMappingService)

	from := time.Time{} // no durations in DB yet -> regeneration starts from zero time

	suite.DurationRepository.On("GetLatestByUser", suite.TestUser).Return(nil, gorm.ErrRecordNotFound)
	suite.HeartbeatService.On("GetLatestByUser", suite.TestUser, false).Return(&models.Heartbeat{UserID: TestUserId, Time: models.CustomTime(suite.TestStartTime)}, nil)
	suite.DurationRepository.On("GetAllWithinByFilters", from, mock.Anything, suite.TestUser, mock.Anything).Return([]*models.Duration{}, nil)
	suite.HeartbeatService.On("StreamAllWithinExcludingHeartbeats", from, mock.Anything, suite.TestUser, models.ExcludeFromDurations).Return(streamSlice(filterHeartbeats(from, time.Now(), suite.TestHeartbeats)), nil)
	suite.DurationRepository.On("InsertBatch", mock.Anything).Return(nil)

	sut.Regenerate(suite.TestUser, false)

	suite.HeartbeatService.AssertNumberOfCalls(suite.T(), "StreamAllWithinExcludingHeartbeats", 1)
	suite.DurationRepository.AssertNumberOfCalls(suite.T(), "InsertBatch", 1)

	var inserted []*models.Duration
	for _, c := range suite.DurationRepository.Calls {
		if c.Method == "InsertBatch" {
			inserted = c.Arguments.Get(0).([]*models.Duration)
		}
	}
	assert.Len(suite.T(), inserted, 3)
	assert.Equal(suite.T(), 30*time.Second, inserted[0].Duration)
	assert.Equal(suite.T(), 20*time.Second, inserted[1].Duration)
	assert.Equal(suite.T(), 15*time.Second, inserted[2].Duration)
}

func (suite *DurationServiceTestSuite) TestDurationService_Get_CustomInterval() {
	sut := NewDurationService(suite.DurationRepository, suite.HeartbeatService, suite.UserService, suite.LanguageMappingService)

	var (
		from      time.Time
		to        time.Time
		durations models.Durations
		err       error
	)

	from, to = suite.TestStartTime.Add(-1*time.Hour), suite.TestStartTime.Add(1*time.Hour)
	suite.HeartbeatService.On("StreamAllWithinExcludingHeartbeats", from, to, suite.TestUser, models.ExcludeFromDurations).Return(streamSlice(filterHeartbeats(from, to, suite.TestHeartbeats)), nil)

	customInterval := 15 * time.Minute
	durations, err = sut.Get(from, to, suite.TestUser, nil, &customInterval, false)

	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), durations, 2)
	assert.Empty(suite.T(), suite.DurationRepository.Calls)
}

func (suite *DurationServiceTestSuite) TestDurationService_Get_WithLanguageMapping() {
	suite.LanguageMappingService.ExpectedCalls[0].Unset()
	suite.LanguageMappingService.On("ResolveByUser", suite.TestUser.ID).Return(map[string]string{"go": "Golang"}, nil)

	sut := NewDurationService(suite.DurationRepository, suite.HeartbeatService, suite.UserService, suite.LanguageMappingService)

	var (
		from      time.Time
		to        time.Time
		durations models.Durations
		err       error
	)

	testDurations := []*models.Duration{
		models.NewDurationFromHeartbeat(suite.TestHeartbeats[0]),
	}

	from, to = suite.TestStartTime.Add(-1*time.Hour), suite.TestStartTime.Add(1*time.Hour)
	suite.DurationRepository.On("GetAllWithinByFilters", from, to, suite.TestUser, mock.Anything).Return(testDurations, nil)
	suite.HeartbeatService.On("StreamAllWithinExcludingHeartbeats", mock.Anything, mock.Anything, suite.TestUser, models.ExcludeFromDurations).Return(streamSlice([]*models.Heartbeat{}), nil)

	durations, err = sut.Get(from, to, suite.TestUser, nil, nil, false)

	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), durations, 1)
	assert.Equal(suite.T(), "Golang", durations.First().Language)
}

func (suite *DurationServiceTestSuite) TestDurationService_Get_WithLanguageMappingsForExistingLanguages() {
	// https:// github.com/muety/wakapi/issues/928
	sut := NewDurationService(suite.DurationRepository, suite.HeartbeatService, suite.UserService, suite.LanguageMappingService)

	// Setup heartbeats
	h1 := &models.Heartbeat{
		UserID:   TestUserId,
		Entity:   "/home/bob/dev/app.ts",
		Language: "TypeScript",
		Time:     models.CustomTime(suite.TestStartTime),
	}
	h2 := &models.Heartbeat{
		UserID:   TestUserId,
		Entity:   "/home/bob/dev/app.tsx",
		Language: "TypeScript",
		Time:     models.CustomTime(suite.TestStartTime.Add(2 * time.Minute)),
	}
	h3 := &models.Heartbeat{
		UserID:   TestUserId,
		Entity:   "/home/bob/dev/app.test.js",
		Language: "JavaScript",
		Time:     models.CustomTime(suite.TestStartTime.Add(4 * time.Minute)),
	}

	// Setup mappings
	// 1. .ts -> Svelte (should not affect .tsx)
	// 2. .js -> PlainJS
	// 3. .test.js -> Jest (should take precedence over .js for app.test.js)
	mappings := map[string]string{
		"ts":      "Svelte",
		"js":      "PlainJS",
		"test.js": "Jest",
	}

	suite.LanguageMappingService.ExpectedCalls[0].Unset()
	suite.LanguageMappingService.On("ResolveByUser", suite.TestUser.ID).Return(mappings, nil)

	from, to := suite.TestStartTime.Add(-1*time.Hour), suite.TestStartTime.Add(1*time.Hour)
	suite.HeartbeatService.On("StreamAllWithinExcludingHeartbeats", from, to, suite.TestUser, models.ExcludeFromDurations).Return(streamSlice([]*models.Heartbeat{h1, h2, h3}), nil)

	durations, err := sut.Get(from, to, suite.TestUser, nil, nil, true)

	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), durations, 3)

	var d1, d2, d3 *models.Duration
	for _, d := range durations {
		if strings.HasSuffix(d.Entity, ".test.js") {
			d3 = d
		} else if strings.HasSuffix(d.Entity, ".tsx") {
			d2 = d
		} else if strings.HasSuffix(d.Entity, ".ts") {
			d1 = d
		}
	}

	assert.NotNil(suite.T(), d1)
	assert.NotNil(suite.T(), d2)
	assert.NotNil(suite.T(), d3)

	assert.Equal(suite.T(), "Svelte", d1.Language)     // exact match for .ts
	assert.Equal(suite.T(), "TypeScript", d2.Language) // no "collateral damage" (did not match .ts)
	assert.Equal(suite.T(), "Jest", d3.Language)       // more precise match (.test.js over .js)
}

func (suite *DurationServiceTestSuite) TestDuration_Hashed() {
	now := time.Unix(1600000000, 0)
	d1 := &models.Duration{
		UserID:          "user1",
		Project:         "project1",
		Language:        "lang1",
		Editor:          "editor1",
		OperatingSystem: "os1",
		Machine:         "machine1",
		Category:        "cat1",
		Branch:          "branch1",
		Entity:          "entity1",
		Extension:       "ext1",
		AIModel:         "aimodel1",
		Time:            models.CustomTime(now),
		Duration:        10 * time.Second,
		NumHeartbeats:   1,
		Timeout:         2 * time.Minute,
	}
	d1.Hashed()

	// same values -> same hash
	d2 := *d1
	d2.Hashed()
	assert.Equal(suite.T(), d1.GroupHash, d2.GroupHash)

	// different project -> different hash
	d3 := *d1
	d3.Project = "project2"
	d3.Hashed()
	assert.NotEqual(suite.T(), d1.GroupHash, d3.GroupHash)

	// ...

	// different branch -> different hash
	d9 := *d1
	d9.Branch = "branch2"
	d9.Hashed()
	assert.NotEqual(suite.T(), d1.GroupHash, d9.GroupHash)

	// different user ID -> different hash
	d10 := *d1
	d10.UserID = "user2"
	d10.Hashed()
	assert.NotEqual(suite.T(), d1.GroupHash, d10.GroupHash)

	// different extension -> different hash
	d11 := *d1
	d11.Extension = "ext2"
	d11.Hashed()
	assert.NotEqual(suite.T(), d1.GroupHash, d11.GroupHash)

	// different entity -> different hash (when not excluded)
	d12 := *d1
	d12.Entity = "entity2"
	d12.Hashed()
	assert.NotEqual(suite.T(), d1.GroupHash, d12.GroupHash)

	// different entity but entity ignored -> same hash
	d13 := *d1
	d13.WithEntityIgnored().Hashed()
	d14 := *d1
	d14.Entity = "entity2"
	d14.WithEntityIgnored().Hashed()
	assert.Equal(suite.T(), d13.GroupHash, d14.GroupHash)

	// different time -> same hash
	d15 := *d1
	d15.Time = models.CustomTime(now.Add(1 * time.Hour))
	d15.Hashed()
	assert.Equal(suite.T(), d1.GroupHash, d15.GroupHash)

	// different duration -> same hash
	d16 := *d1
	d16.Duration = 20 * time.Second
	d16.Hashed()
	assert.Equal(suite.T(), d1.GroupHash, d16.GroupHash)

	// different num heartbeats -> same hash
	d17 := *d1
	d17.NumHeartbeats = 5
	d17.Hashed()
	assert.Equal(suite.T(), d1.GroupHash, d17.GroupHash)

	// different ID -> same hash
	d18 := *d1
	d18.ID = 123
	d18.Hashed()
	assert.Equal(suite.T(), d1.GroupHash, d18.GroupHash)

	// different timeout -> same hash
	d19 := *d1
	d19.Timeout = 10 * time.Minute
	d19.Hashed()
	assert.Equal(suite.T(), d1.GroupHash, d19.GroupHash)

	// different ai model -> different hash
	d20 := *d1
	d20.AIModel = "aimodel2"
	d20.Hashed()
	assert.NotEqual(suite.T(), d1.GroupHash, d20.GroupHash)
}

func (suite *DurationServiceTestSuite) TestDurationService_HeartbeatCreateEvent_DeletesDurations() {
	sut, eventBus := suite.createSut()
	sut.lastUserJob[suite.TestUser.ID] = time.Now()

	now := time.Now().Truncate(time.Second)
	t1 := now.Add(-1 * time.Hour)
	t2 := now.Add(-30 * time.Minute)
	t3 := now.Add(-2 * time.Hour)

	// the listener does its work in a separate goroutine, so signal deleted timestamps via a channel instead of polling mock.Calls from the test, which races with the mock's internal bookkeeping
	deleted := make(chan time.Time, 10)
	onDeleted := func(call *mock.Call) *mock.Call {
		return call.Run(func(args mock.Arguments) { deleted <- args.Get(1).(time.Time) })
	}

	// 1. Initial heartbeat: triggers DeleteByUserAfter on cache miss and initializes dirtyFrom
	onDeleted(suite.DurationRepository.On("DeleteByUserAfter", suite.TestUser, t1).Return(nil).Once())

	eventBus.Publish(hub.Message{
		Name: config.EventHeartbeatCreate,
		Fields: map[string]interface{}{
			config.FieldPayload: &models.Heartbeat{
				User:   suite.TestUser,
				UserID: suite.TestUser.ID,
				Time:   models.CustomTime(t1),
			},
		},
	})

	suite.awaitDeleted(deleted, t1)
	suite.awaitDirtyFrom(sut, t1)

	// 2. Subsequent heartbeat with timestamp after dirtyFrom: skips DeleteByUserAfter
	eventBus.Publish(hub.Message{
		Name: config.EventHeartbeatCreate,
		Fields: map[string]interface{}{
			config.FieldPayload: &models.Heartbeat{
				User:   suite.TestUser,
				UserID: suite.TestUser.ID,
				Time:   models.CustomTime(t2),
			},
		},
	})

	time.Sleep(100 * time.Millisecond)
	suite.DurationRepository.AssertNumberOfCalls(suite.T(), "DeleteByUserAfter", 1)

	// 3. Out-of-order heartbeat with timestamp before dirtyFrom: triggers DeleteByUserAfter and updates dirtyFrom
	onDeleted(suite.DurationRepository.On("DeleteByUserAfter", suite.TestUser, t3).Return(nil).Once())

	eventBus.Publish(hub.Message{
		Name: config.EventHeartbeatCreate,
		Fields: map[string]interface{}{
			config.FieldPayload: &models.Heartbeat{
				User:   suite.TestUser,
				UserID: suite.TestUser.ID,
				Time:   models.CustomTime(t3),
			},
		},
	})

	suite.awaitDeleted(deleted, t3)
	suite.awaitDirtyFrom(sut, t3)

	suite.DurationRepository.AssertExpectations(suite.T())
	_ = sut
}

func (suite *DurationServiceTestSuite) awaitDeleted(deleted chan time.Time, expected time.Time) {
	suite.T().Helper()
	select {
	case actual := <-deleted:
		assert.Equal(suite.T(), expected, actual)
	case <-time.After(2 * time.Second):
		suite.T().Fatalf("timed out waiting for durations to be deleted after %v", expected)
	}
}

func (suite *DurationServiceTestSuite) awaitDirtyFrom(sut *DurationService, expected time.Time) {
	suite.T().Helper()
	// go-cache is internally synchronized, so reading it from the test goroutine is safe
	assert.Eventually(suite.T(), func() bool {
		dirtyFrom, ok := sut.dirtyFrom.Get("dirty_from_" + suite.TestUser.ID)
		return ok && dirtyFrom.(time.Time).Equal(expected)
	}, 2*time.Second, 20*time.Millisecond)
}

func (suite *DurationServiceTestSuite) TestDurationService_Pending_RegenerationsAreDeduplicated() {
	// regeneration for a single user can be triggered concurrently from different queues (heartbeat event
	// listener, aggregation service, RegenerateAll), so acquiring the pending marker has to be atomic
	sut, _ := suite.createSut()

	const racers = 50
	acquired := make(chan bool, racers)
	for i := 0; i < racers; i++ {
		go func() { acquired <- sut.acquirePending(suite.TestUser.ID) }()
	}

	wins := 0
	for i := 0; i < racers; i++ {
		if <-acquired {
			wins++
		}
	}
	assert.Equal(suite.T(), 1, wins, "only one regeneration per user should proceed")

	sut.releasePending(suite.TestUser.ID)
	assert.True(suite.T(), sut.acquirePending(suite.TestUser.ID), "regeneration should be acquirable again after release")

	// distinct users should not interfere with each other
	others := make(chan bool, racers)
	for i := 0; i < racers; i++ {
		userId := fmt.Sprintf("user_%d", i)
		go func() { others <- sut.acquirePending(userId) }()
	}
	for i := 0; i < racers; i++ {
		assert.True(suite.T(), <-others)
	}
}

func (suite *DurationServiceTestSuite) TestDurationService_Get_SameDayTimezone() {
	// See https://github.com/muety/wakapi/issues/975
	sut := NewDurationService(suite.DurationRepository, suite.HeartbeatService, suite.UserService, suite.LanguageMappingService)

	nyUser := &models.User{
		ID:                   "user_ny",
		Location:             "America/New_York",
		HeartbeatsTimeoutSec: int(models.DefaultHeartbeatsTimeoutLegacy / time.Second),
	}

	suite.LanguageMappingService.On("ResolveByUser", nyUser.ID).Return(make(map[string]string), nil)

	// In UTC: 2023-10-01 23:59:30 UTC and 2023-10-02 00:00:30 UTC span across UTC midnight (server-local).
	// In EDT (UTC-4): 2023-10-01 19:59:30 EDT and 2023-10-01 20:00:30 EDT are on the same day (Oct 1).
	hb1Time := time.Date(2023, 10, 1, 23, 59, 30, 0, time.UTC)
	hb2Time := time.Date(2023, 10, 2, 0, 0, 30, 0, time.UTC)

	heartbeats := []*models.Heartbeat{
		{
			UserID:          nyUser.ID,
			Project:         TestProject1,
			Language:        TestLanguageGo,
			Editor:          TestEditorGoland,
			OperatingSystem: TestOsLinux,
			Machine:         TestMachine1,
			Entity:          TestEntity1,
			Time:            models.CustomTime(hb1Time),
		},
		{
			UserID:          nyUser.ID,
			Project:         TestProject1,
			Language:        TestLanguageGo,
			Editor:          TestEditorGoland,
			OperatingSystem: TestOsLinux,
			Machine:         TestMachine1,
			Entity:          TestEntity1,
			Time:            models.CustomTime(hb2Time),
		},
	}

	from := hb1Time.Add(-1 * time.Hour)
	to := hb2Time.Add(1 * time.Hour)

	suite.HeartbeatService.
		On("StreamAllWithinExcludingHeartbeats", from, to, nyUser, models.ExcludeFromDurations).
		Return(streamSlice(heartbeats), nil).
		Once()

	durations, err := sut.Get(from, to, nyUser, nil, nil, true)
	assert.NoError(suite.T(), err)
	// because both heartbeats fall on Oct 1 in America/New_York, sameDay is true and they get merged into 1 duration
	assert.Len(suite.T(), durations, 1)
	assert.Equal(suite.T(), 2, durations[0].NumHeartbeats)
	assert.Equal(suite.T(), 60*time.Second, durations[0].Duration)

	// Now test heartbeats that cross America/New_York midnight (Oct 1 23:59:30 EDT = Oct 2 03:59:30 UTC, and Oct 2 00:00:30 EDT = Oct 2 04:00:30 UTC)
	// In UTC, both are on Oct 2 (same day in server-local), in EDT, they are on different days (Oct 1 and Oct 2), so sameDay must be false!
	hb3Time := time.Date(2023, 10, 2, 3, 59, 30, 0, time.UTC) // Oct 1 23:59:30 EDT
	hb4Time := time.Date(2023, 10, 2, 4, 0, 30, 0, time.UTC)  // Oct 2 00:00:30 EDT

	heartbeatsMidnight := []*models.Heartbeat{
		{
			UserID:          nyUser.ID,
			Project:         TestProject1,
			Language:        TestLanguageGo,
			Editor:          TestEditorGoland,
			OperatingSystem: TestOsLinux,
			Machine:         TestMachine1,
			Entity:          TestEntity1,
			Time:            models.CustomTime(hb3Time),
		},
		{
			UserID:          nyUser.ID,
			Project:         TestProject1,
			Language:        TestLanguageGo,
			Editor:          TestEditorGoland,
			OperatingSystem: TestOsLinux,
			Machine:         TestMachine1,
			Entity:          TestEntity1,
			Time:            models.CustomTime(hb4Time),
		},
	}

	from2 := hb3Time.Add(-1 * time.Hour)
	to2 := hb4Time.Add(1 * time.Hour)

	suite.HeartbeatService.On("StreamAllWithinExcludingHeartbeats", from2, to2, nyUser, models.ExcludeFromDurations).
		Return(streamSlice(heartbeatsMidnight), nil).Once()

	durationsMidnight, err := sut.Get(from2, to2, nyUser, nil, nil, true)
	assert.NoError(suite.T(), err)
	// because they cross midnight in America/New_York, sameDay is false, splitting them into 2 durations
	assert.Len(suite.T(), durationsMidnight, 2)
}

func (suite *DurationServiceTestSuite) TestDurationService_Get_Filter_Cached_TailRecovery() {
	sut := NewDurationService(suite.DurationRepository, suite.HeartbeatService, suite.UserService, suite.LanguageMappingService)

	from := suite.TestStartTime.Add(-24 * time.Hour)
	to := suite.TestStartTime.Add(2 * time.Hour)

	cachedDurations := []*models.Duration{
		{
			UserID:   TestUserId,
			Project:  TestProject1,
			Language: TestLanguageGo,
			Time:     models.CustomTime(suite.TestStartTime),
			Duration: 10 * time.Minute,
		},
	}
	latestDuration := &models.Duration{
		UserID:   TestUserId,
		Project:  TestProject2,
		Time:     models.CustomTime(suite.TestStartTime.Add(1 * time.Hour)),
		Duration: 10 * time.Minute,
	}

	testFilters := models.NewFiltersWith(models.SummaryLanguage, TestLanguageGo)

	suite.DurationRepository.On("GetAllWithinByFilters", from, to, suite.TestUser, mock.Anything).Return(cachedDurations, nil).Once()
	suite.DurationRepository.On("GetLatestByUser", suite.TestUser).Return(latestDuration, nil).Once()

	liveHeartbeat := &models.Heartbeat{
		UserID:   TestUserId,
		Project:  TestProject1,
		Language: TestLanguageGo,
		Time:     models.CustomTime(suite.TestStartTime.Add(1*time.Hour + 30*time.Minute)),
	}
	expectedMissingFrom := latestDuration.TimeEnd().Add(time.Second)
	suite.HeartbeatService.On("StreamAllWithinExcludingHeartbeats", expectedMissingFrom, to, suite.TestUser, models.ExcludeFromDurations).
		Return(streamSlice([]*models.Heartbeat{liveHeartbeat}), nil).Once()

	durations, err := sut.Get(from, to, suite.TestUser, testFilters, nil, false)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), durations, 2)
	assert.Equal(suite.T(), TestLanguageGo, durations[0].Language)
	assert.Equal(suite.T(), TestLanguageGo, durations[1].Language)
}

func (suite *DurationServiceTestSuite) TestDurationService_Get_Filter_Cached_NoLiveNeeded() {
	sut := NewDurationService(suite.DurationRepository, suite.HeartbeatService, suite.UserService, suite.LanguageMappingService)

	from := suite.TestStartTime.Add(-24 * time.Hour)
	to := suite.TestStartTime.Add(1 * time.Hour)

	cachedDurations := []*models.Duration{
		{
			UserID:   TestUserId,
			Project:  TestProject1,
			Language: TestLanguageGo,
			Time:     models.CustomTime(suite.TestStartTime),
			Duration: 10 * time.Minute,
		},
	}
	latestDuration := &models.Duration{
		UserID:   TestUserId,
		Project:  TestProject2,
		Time:     models.CustomTime(suite.TestStartTime.Add(1 * time.Hour)),
		Duration: 30 * time.Minute,
	}

	testFilters := models.NewFiltersWith(models.SummaryLanguage, TestLanguageGo)

	suite.DurationRepository.On("GetAllWithinByFilters", from, to, suite.TestUser, mock.Anything).Return(cachedDurations, nil).Once()
	suite.DurationRepository.On("GetLatestByUser", suite.TestUser).Return(latestDuration, nil).Once()

	durations, err := sut.Get(from, to, suite.TestUser, testFilters, nil, false)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), durations, 1)
	assert.Equal(suite.T(), TestLanguageGo, durations[0].Language)
}

func (suite *DurationServiceTestSuite) TestDurationService_Get_Filter_Cached_EmptyCache_NoLiveNeeded() {
	sut := NewDurationService(suite.DurationRepository, suite.HeartbeatService, suite.UserService, suite.LanguageMappingService)

	from := suite.TestStartTime.Add(-24 * time.Hour)
	to := suite.TestStartTime.Add(1 * time.Hour)

	latestDuration := &models.Duration{
		UserID:   TestUserId,
		Project:  TestProject2,
		Time:     models.CustomTime(suite.TestStartTime.Add(1 * time.Hour)),
		Duration: 30 * time.Minute,
	}

	testFilters := models.NewFiltersWith(models.SummaryLanguage, TestLanguageGo)

	suite.DurationRepository.On("GetAllWithinByFilters", from, to, suite.TestUser, mock.Anything).Return([]*models.Duration{}, nil).Once()
	suite.DurationRepository.On("GetLatestByUser", suite.TestUser).Return(latestDuration, nil).Once()

	durations, err := sut.Get(from, to, suite.TestUser, testFilters, nil, false)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), durations)
}

func (suite *DurationServiceTestSuite) TestDurationService_Get_ProjectFilter_SelectFilteredOnly() {
	sut := NewDurationService(suite.DurationRepository, suite.HeartbeatService, suite.UserService, suite.LanguageMappingService)

	from := suite.TestStartTime.Add(-24 * time.Hour)
	to := suite.TestStartTime.Add(1 * time.Hour)

	cachedDurations := []*models.Duration{
		{
			UserID:   TestUserId,
			Project:  TestProject1,
			Language: TestLanguageGo,
			Time:     models.CustomTime(suite.TestStartTime),
			Duration: 10 * time.Minute,
		},
	}
	latestDuration := &models.Duration{
		UserID:   TestUserId,
		Project:  TestProject2,
		Time:     models.CustomTime(suite.TestStartTime.Add(1 * time.Hour)),
		Duration: 30 * time.Minute,
	}

	// SelectFilteredOnly indicates project total only, so skipCache is false and cached durations are used
	testFilters := models.NewFiltersWith(models.SummaryProject, TestProject1).WithSelectFilteredOnly()

	suite.DurationRepository.On("GetAllWithinByFilters", from, to, suite.TestUser, mock.Anything).Return(cachedDurations, nil).Once()
	suite.DurationRepository.On("GetLatestByUser", suite.TestUser).Return(latestDuration, nil).Once()

	durations, err := sut.Get(from, to, suite.TestUser, testFilters, nil, false)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), durations, 1)
	assert.Equal(suite.T(), TestProject1, durations[0].Project)
}

func (suite *DurationServiceTestSuite) TestDurationService_Merge() {
	sut := NewDurationService(suite.DurationRepository, suite.HeartbeatService, suite.UserService, suite.LanguageMappingService)

	t0 := suite.TestStartTime
	d1 := models.Durations{
		(&models.Duration{
			UserID:   TestUserId,
			Project:  TestProject1,
			Time:     models.CustomTime(t0),
			Duration: 5 * time.Minute,
		}).Hashed(),
	}

	// Case 1: multiple items in d2, ensure last item is NOT dropped
	d2Multiple := models.Durations{
		(&models.Duration{
			UserID:   TestUserId,
			Project:  TestProject2,
			Time:     models.CustomTime(t0.Add(1 * time.Hour)),
			Duration: 5 * time.Minute,
		}).Hashed(),
		(&models.Duration{
			UserID:   TestUserId,
			Project:  TestProject2,
			Time:     models.CustomTime(t0.Add(2 * time.Hour)),
			Duration: 5 * time.Minute,
		}).Hashed(),
		(&models.Duration{
			UserID:   TestUserId,
			Project:  TestProject2,
			Time:     models.CustomTime(t0.Add(3 * time.Hour)),
			Duration: 5 * time.Minute,
		}).Hashed(),
	}

	merged, err := sut.merge(d1, d2Multiple, suite.TestUser)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), merged, 4)
	assert.Equal(suite.T(), t0.Add(3*time.Hour), merged[3].Time.T())

	// Case 2: diff < timeout, different hash - ensure middleRight is NOT dropped
	d2CloseDiffHash := models.Durations{
		(&models.Duration{
			UserID:   TestUserId,
			Project:  TestProject2,
			Time:     models.CustomTime(t0.Add(6 * time.Minute)), // 1 min after d1 ends
			Duration: 5 * time.Minute,
		}).Hashed(),
		(&models.Duration{
			UserID:   TestUserId,
			Project:  TestProject2,
			Time:     models.CustomTime(t0.Add(20 * time.Minute)),
			Duration: 5 * time.Minute,
		}).Hashed(),
	}

	merged2, err := sut.merge(d1, d2CloseDiffHash, suite.TestUser)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), merged2, 3)
	assert.Equal(suite.T(), 6*time.Minute, merged2[0].Duration)          // extended by 1 min
	assert.Equal(suite.T(), TestProject2, merged2[1].Project)            // middleRight kept
	assert.Equal(suite.T(), t0.Add(20*time.Minute), merged2[2].Time.T()) // d2[1] kept

	// Case 3: diff < timeout, same hash - merged into single
	d2CloseSameHash := models.Durations{
		(&models.Duration{
			UserID:   TestUserId,
			Project:  TestProject1,
			Time:     models.CustomTime(t0.Add(6 * time.Minute)),
			Duration: 5 * time.Minute,
		}).Hashed(),
	}

	merged3, err := sut.merge(d1, d2CloseSameHash, suite.TestUser)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), merged3, 1)
	assert.Equal(suite.T(), 11*time.Minute, merged3[0].Duration) // 5m + 1m gap + 5m
}

func (suite *DurationServiceTestSuite) createSut() (*DurationService, *hub.Hub) {
	originalEventBus := config.EventBus()
	defer config.SetEventBus(originalEventBus)
	eventBus := hub.New()
	config.SetEventBus(eventBus)
	return NewDurationService(suite.DurationRepository, suite.HeartbeatService, suite.UserService, suite.LanguageMappingService), eventBus
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

func streamSlice[T any](data []T) chan T {
	c := make(chan T)
	go func() {
		defer close(c)
		for _, h := range data {
			c <- h
		}
	}()
	return c
}
