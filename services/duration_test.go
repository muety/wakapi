package services

import (
	"math/rand"
	"testing"
	"time"

	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	TestUserId           = "muety"
	TestProject1         = "test-project-1"
	TestProject2         = "test-project-2"
	TestProject3         = "test-project-3"
	TestProject4         = "something-completely-different-4"
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

	// https://anchr.io/i/F0HEK.jpg
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
	// https://anchr.io/i/F0HEK.jpg
	sut := NewDurationService(suite.DurationRepository, suite.HeartbeatService, suite.UserService, suite.LanguageMappingService)

	var (
		from      time.Time
		to        time.Time
		durations models.Durations
		err       error
	)

	/* TEST 1 */
	from, to = suite.TestStartTime.Add(-1*time.Hour), suite.TestStartTime.Add(-1*time.Minute)
	suite.HeartbeatService.On("StreamAllWithin", from, to, suite.TestUser).Return(streamSlice(filterHeartbeats(from, to, suite.TestHeartbeats)), nil)

	durations, err = sut.Get(from, to, suite.TestUser, nil, nil, true)

	assert.Nil(suite.T(), err)
	assert.Empty(suite.T(), durations)

	/* TEST 2 */
	from, to = suite.TestStartTime.Add(-1*time.Hour), suite.TestStartTime.Add(1*time.Second)
	suite.HeartbeatService.On("StreamAllWithin", from, to, suite.TestUser).Return(streamSlice(filterHeartbeats(from, to, suite.TestHeartbeats)), nil)

	durations, err = sut.Get(from, to, suite.TestUser, nil, nil, true)

	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), durations, 1)
	assert.Equal(suite.T(), time.Duration(0), durations.First().Duration)
	assert.Equal(suite.T(), 1, durations.First().NumHeartbeats)

	/* TEST 3 */
	from, to = suite.TestStartTime, suite.TestStartTime.Add(1*time.Hour)
	suite.HeartbeatService.On("StreamAllWithin", from, to, suite.TestUser).Return(streamSlice(filterHeartbeats(from, to, suite.TestHeartbeats)), nil)

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
	suite.HeartbeatService.On("StreamAllWithin", from, to, suite.TestUser).Return(streamSlice(filterHeartbeats(from, to, suite.TestHeartbeats)), nil)

	durations, err = sut.Get(from, to, suite.TestUser, models.NewFiltersWith(models.SummaryEditor, TestEditorGoland), nil, true)
	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), durations, 2)
	assert.Equal(suite.T(), 30*time.Second, durations[0].Duration)
	assert.Equal(suite.T(), 20*time.Second, durations[1].Duration)
	for _, d := range durations {
		assert.Equal(suite.T(), TestEditorGoland, d.Editor)
	}
}

func (suite *DurationServiceTestSuite) TestDurationService_Get_ProjectDetails() {
	// https://github.com/muety/wakapi/issues/876
	sut := NewDurationService(suite.DurationRepository, suite.HeartbeatService, suite.UserService, suite.LanguageMappingService)

	var (
		from      time.Time
		to        time.Time
		durations models.Durations
		err       error
	)

	from, to = suite.TestStartTime.Add(-1*time.Hour), suite.TestStartTime.Add(1*time.Hour)
	suite.HeartbeatService.On("StreamAllWithin", from, to, suite.TestUser).Return(streamSlice(filterHeartbeats(from, to, suite.TestHeartbeats)), nil)

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
	call1 := suite.HeartbeatService.On("StreamAllWithin", from, to, suite.TestUser).Return(streamSlice(filterHeartbeats(from, to, suite.TestHeartbeats)), nil)
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
	call2 := suite.HeartbeatService.On("StreamAllWithin", from, to, suite.TestUser).Return(streamSlice(filterHeartbeats(from, to, suite.TestHeartbeats)), nil)
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
	call3 := suite.HeartbeatService.On("StreamAllWithin", from, to, suite.TestUser).Return(streamSlice(filterHeartbeats(from, to, suite.TestHeartbeats)), nil)
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
	suite.HeartbeatService.On("StreamAllWithin", toCached, to, suite.TestUser).Return(streamSlice(filterHeartbeats(toCached, to, suite.TestHeartbeats)), nil)

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

func (suite *DurationServiceTestSuite) TestDurationService_Get_CustomInterval() {
	sut := NewDurationService(suite.DurationRepository, suite.HeartbeatService, suite.UserService, suite.LanguageMappingService)

	var (
		from      time.Time
		to        time.Time
		durations models.Durations
		err       error
	)

	from, to = suite.TestStartTime.Add(-1*time.Hour), suite.TestStartTime.Add(1*time.Hour)
	suite.HeartbeatService.On("StreamAllWithin", from, to, suite.TestUser).Return(streamSlice(filterHeartbeats(from, to, suite.TestHeartbeats)), nil)

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
	suite.HeartbeatService.On("StreamAllWithin", mock.Anything, mock.Anything, suite.TestUser).Return(streamSlice([]*models.Heartbeat{}), nil)

	durations, err = sut.Get(from, to, suite.TestUser, nil, nil, false)

	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), durations, 1)
	assert.Equal(suite.T(), "Golang", durations.First().Language)
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
