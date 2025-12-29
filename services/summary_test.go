package services

import (
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	TestProjectLabel1 = "private"
	TestProjectLabel2 = "work"
	TestProjectLabel3 = "non-existing"
)

type SummaryServiceTestSuite struct {
	suite.Suite
	TestUser            *models.User
	TestStartTime       time.Time
	TestDurations       []*models.Duration
	TestLabels          []*models.ProjectLabel
	SummaryRepository   *mocks.SummaryRepositoryMock
	HeartbeatService    *mocks.HeartbeatServiceMock
	DurationService     *mocks.DurationServiceMock
	AliasService        *mocks.AliasServiceMock
	ProjectLabelService *mocks.ProjectLabelServiceMock
}

func (suite *SummaryServiceTestSuite) SetupSuite() {
	suite.TestUser = &models.User{ID: TestUserId}

	suite.TestStartTime = time.Unix(0, MinUnixTime1)
	suite.TestDurations = []*models.Duration{
		{
			UserID:          TestUserId,
			Project:         TestProject1,
			Language:        TestLanguageGo,
			Editor:          TestEditorGoland,
			OperatingSystem: TestOsLinux,
			Machine:         TestMachine1,
			Branch:          TestBranchMaster,
			Entity:          TestEntity1,
			Category:        TestCategoryCoding,
			Time:            models.CustomTime(suite.TestStartTime),
			Duration:        150 * time.Second,
			NumHeartbeats:   2,
		},
		{
			UserID:          TestUserId,
			Project:         TestProject1,
			Language:        TestLanguageGo,
			Editor:          TestEditorGoland,
			OperatingSystem: TestOsLinux,
			Machine:         TestMachine1,
			Branch:          TestBranchMaster,
			Entity:          TestEntity1,
			Category:        TestCategoryCoding,
			Time:            models.CustomTime(suite.TestStartTime.Add((30 + 130) * time.Second)),
			Duration:        20 * time.Second,
			NumHeartbeats:   1,
		},
		{
			UserID:          TestUserId,
			Project:         TestProject1,
			Language:        TestLanguageGo,
			Editor:          TestEditorVscode,
			OperatingSystem: TestOsLinux,
			Machine:         TestMachine1,
			Branch:          TestBranchDev,
			Entity:          TestEntity1,
			Category:        TestCategoryBrowsing,
			Time:            models.CustomTime(suite.TestStartTime.Add(3 * time.Minute)),
			Duration:        15 * time.Second,
			NumHeartbeats:   3,
		},
	}
	suite.TestLabels = []*models.ProjectLabel{
		{
			ID:         uint(rand.Uint32()),
			UserID:     TestUserId,
			ProjectKey: TestProject1,
			Label:      TestProjectLabel1,
		},
		{
			ID:         uint(rand.Uint32()),
			UserID:     TestUserId,
			ProjectKey: TestProject2,
			Label:      TestProjectLabel1,
		},
		{
			ID:         uint(rand.Uint32()),
			UserID:     TestUserId,
			ProjectKey: TestProject3,
			Label:      TestProjectLabel3,
		},
	}
}

func (suite *SummaryServiceTestSuite) BeforeTest(suiteName, testName string) {
	suite.SummaryRepository = new(mocks.SummaryRepositoryMock)
	suite.HeartbeatService = new(mocks.HeartbeatServiceMock)
	suite.DurationService = new(mocks.DurationServiceMock)
	suite.AliasService = new(mocks.AliasServiceMock)
	suite.ProjectLabelService = new(mocks.ProjectLabelServiceMock)
}

func TestSummaryServiceTestSuite(t *testing.T) {
	suite.Run(t, new(SummaryServiceTestSuite))
}

func (suite *SummaryServiceTestSuite) TestSummaryService_Summarize() {
	sut := NewSummaryService(suite.SummaryRepository, suite.HeartbeatService, suite.DurationService, suite.AliasService, suite.ProjectLabelService)

	var (
		from   time.Time
		to     time.Time
		result *models.Summary
		err    error
	)

	/* TEST 1 */
	from, to = suite.TestStartTime.Add(-1*time.Hour), suite.TestStartTime.Add(-1*time.Minute)
	suite.DurationService.On("Get", from, to, suite.TestUser, mock.Anything, mock.Anything, false).Return(filterDurations(from, to, suite.TestDurations), nil)

	result, err = sut.Summarize(from, to, suite.TestUser, nil, nil)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), from, result.FromTime.T())
	assert.Equal(suite.T(), to, result.ToTime.T())
	assert.Zero(suite.T(), result.TotalTime())
	assert.Zero(suite.T(), result.NumHeartbeats)
	assert.Empty(suite.T(), result.Projects)

	/* TEST 2 */
	from, to = suite.TestStartTime.Add(-1*time.Hour), suite.TestStartTime.Add(1*time.Second)
	suite.DurationService.On("Get", from, to, suite.TestUser, mock.Anything, mock.Anything, false).Return(filterDurations(from, to, suite.TestDurations), nil)

	result, err = sut.Summarize(from, to, suite.TestUser, nil, nil)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), suite.TestDurations[0].Time.T(), result.FromTime.T())
	assert.Equal(suite.T(), suite.TestDurations[0].Time.T(), result.ToTime.T())
	assert.Equal(suite.T(), 150*time.Second, result.TotalTime())
	assert.Equal(suite.T(), 2, result.NumHeartbeats)
	assertNumAllItems(suite.T(), 1, result, "")

	/* TEST 3 */
	from, to = suite.TestStartTime, suite.TestStartTime.Add(1*time.Hour)
	suite.DurationService.On("Get", from, to, suite.TestUser, mock.Anything, mock.Anything, false).Return(filterDurations(from, to, suite.TestDurations), nil)

	result, err = sut.Summarize(from, to, suite.TestUser, nil, nil)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), suite.TestDurations[0].Time.T(), result.FromTime.T())
	assert.Equal(suite.T(), suite.TestDurations[len(suite.TestDurations)-1].Time.T(), result.ToTime.T())
	assert.Equal(suite.T(), 185*time.Second, result.TotalTime())
	assert.Equal(suite.T(), 185*time.Second, result.TotalTimeBy(models.SummaryProject))
	assert.Equal(suite.T(), 185*time.Second, result.TotalTimeBy(models.SummaryOS))
	assert.Equal(suite.T(), 185*time.Second, result.TotalTimeBy(models.SummaryMachine))
	assert.Equal(suite.T(), 185*time.Second, result.TotalTimeBy(models.SummaryLanguage))
	assert.Equal(suite.T(), 185*time.Second, result.TotalTimeBy(models.SummaryEditor))
	assert.Zero(suite.T(), result.TotalTimeBy(models.SummaryBranch)) // no filters -> no branches contained
	assert.Zero(suite.T(), result.TotalTimeBy(models.SummaryEntity)) // no filters -> no entities contained
	assert.Zero(suite.T(), result.TotalTimeBy(models.SummaryLabel))
	assert.Equal(suite.T(), 170*time.Second, result.TotalTimeByKey(models.SummaryEditor, TestEditorGoland))
	assert.Equal(suite.T(), 15*time.Second, result.TotalTimeByKey(models.SummaryEditor, TestEditorVscode))
	assert.Equal(suite.T(), 170*time.Second, result.TotalTimeByKey(models.SummaryCategory, TestCategoryCoding))
	assert.Equal(suite.T(), 15*time.Second, result.TotalTimeByKey(models.SummaryCategory, TestCategoryBrowsing))
	assert.Equal(suite.T(), 6, result.NumHeartbeats)
	assert.Len(suite.T(), result.Editors, 2)
	assertNumAllItems(suite.T(), 1, result, "e")
}

func (suite *SummaryServiceTestSuite) TestSummaryService_Retrieve() {
	sut := NewSummaryService(suite.SummaryRepository, suite.HeartbeatService, suite.DurationService, suite.AliasService, suite.ProjectLabelService)

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
			NumHeartbeats:    100,
		},
	}

	suite.SummaryRepository.On("GetByUserWithin", suite.TestUser, from, to).Return(summaries, nil)
	suite.DurationService.On("Get", from, summaries[0].FromTime.T(), suite.TestUser, mock.Anything, mock.Anything, false).Return(models.Durations{}, nil)
	suite.DurationService.On("Get", summaries[0].ToTime.T(), to, suite.TestUser, mock.Anything, mock.Anything, false).Return(models.Durations{}, nil)

	result, err = sut.Retrieve(from, to, suite.TestUser, nil, nil)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Projects, 1)
	assert.Equal(suite.T(), summaries[0].Projects[0].Total*time.Second, result.TotalTime())
	assert.Equal(suite.T(), 100, result.NumHeartbeats)
	suite.DurationService.AssertNumberOfCalls(suite.T(), "Get", 2)

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
			NumHeartbeats:    100,
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
			NumHeartbeats:    100,
		},
	}

	suite.SummaryRepository.On("GetByUserWithin", suite.TestUser, from, to).Return(summaries, nil)
	suite.DurationService.On("Get", from, summaries[0].FromTime.T(), suite.TestUser, mock.Anything, mock.Anything, false).Return(filterDurations(from, summaries[0].FromTime.T(), suite.TestDurations), nil)

	result, err = sut.Retrieve(from, to, suite.TestUser, nil, nil)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Projects, 2)
	assert.Equal(suite.T(), 185*time.Second+90*time.Minute, result.TotalTime())
	assert.Equal(suite.T(), 185*time.Second+45*time.Minute, result.TotalTimeByKey(models.SummaryProject, TestProject1))
	assert.Equal(suite.T(), 45*time.Minute, result.TotalTimeByKey(models.SummaryProject, TestProject2))
	assert.Equal(suite.T(), 206, result.NumHeartbeats)
	suite.DurationService.AssertNumberOfCalls(suite.T(), "Get", 2+1)

	/* TEST 3 */
	from = time.Date(suite.TestStartTime.Year(), suite.TestStartTime.Month(), suite.TestStartTime.Day()+1, 0, 0, 0, 0, suite.TestStartTime.Location()) // start of next day
	to = time.Date(from.Year(), from.Month(), from.Day()+2, 13, 30, 0, 0, from.Location())                                                             // noon of third-next day
	summaries = []*models.Summary{
		{
			ID:       uint(rand.Uint32()),
			UserID:   TestUserId,
			FromTime: models.CustomTime(from),
			ToTime:   models.CustomTime(from.Add(24 * time.Hour)),
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
			NumHeartbeats:    100,
		},
		{
			ID:       uint(rand.Uint32()),
			UserID:   TestUserId,
			FromTime: models.CustomTime(to.Add(-2 * time.Hour)),
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
			NumHeartbeats:    100,
		},
	}

	suite.SummaryRepository.On("GetByUserWithin", suite.TestUser, from, to).Return(summaries, nil)
	suite.DurationService.On("Get", summaries[0].ToTime.T(), summaries[1].FromTime.T(), suite.TestUser, mock.Anything, mock.Anything, false).Return(filterDurations(summaries[0].ToTime.T(), summaries[1].FromTime.T(), suite.TestDurations), nil)

	result, err = sut.Retrieve(from, to, suite.TestUser, nil, nil)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Projects, 2)
	assert.Equal(suite.T(), 90*time.Minute, result.TotalTime())
	assert.Equal(suite.T(), 45*time.Minute, result.TotalTimeByKey(models.SummaryProject, TestProject1))
	assert.Equal(suite.T(), 45*time.Minute, result.TotalTimeByKey(models.SummaryProject, TestProject2))
	assert.Equal(suite.T(), 200, result.NumHeartbeats)
	suite.DurationService.AssertNumberOfCalls(suite.T(), "Get", 2+1)
}

func (suite *SummaryServiceTestSuite) TestSummaryService_Retrieve_DuplicateSummaries() {
	sut := NewSummaryService(suite.SummaryRepository, suite.HeartbeatService, suite.DurationService, suite.AliasService, suite.ProjectLabelService)

	suite.ProjectLabelService.On("GetByUser", suite.TestUser.ID).Return([]*models.ProjectLabel{}, nil)

	var (
		summaries []*models.Summary
		from      time.Time
		to        time.Time
		result    *models.Summary
		err       error
	)

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
	summaries = append(summaries, &(*summaries[0])) // add same summary again -> mustn't be counted twice!

	suite.SummaryRepository.On("GetByUserWithin", suite.TestUser, from, to).Return(summaries, nil)
	suite.DurationService.On("Get", from, summaries[0].FromTime.T(), suite.TestUser, mock.Anything, mock.Anything, false).Return(models.Durations{}, nil)
	suite.DurationService.On("Get", summaries[0].ToTime.T(), to, suite.TestUser, mock.Anything, mock.Anything, false).Return(models.Durations{}, nil)

	result, err = sut.Retrieve(from, to, suite.TestUser, nil, nil)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Projects, 1)
	assert.Equal(suite.T(), summaries[0].Projects[0].Total*time.Second, result.TotalTime())
	suite.DurationService.AssertNumberOfCalls(suite.T(), "Get", 2)
}

func (suite *SummaryServiceTestSuite) TestSummaryService_Retrieve_ZeroLengthSummaries() {
	sut := NewSummaryService(suite.SummaryRepository, suite.HeartbeatService, suite.DurationService, suite.AliasService, suite.ProjectLabelService)

	suite.ProjectLabelService.On("GetByUser", suite.TestUser.ID).Return([]*models.ProjectLabel{}, nil)

	var (
		summaries []*models.Summary
		from      time.Time
		to        time.Time
		result    *models.Summary
		err       error
	)

	from, to = suite.TestStartTime.Add(-12*time.Hour), suite.TestStartTime.Add(12*time.Hour)
	summaries = []*models.Summary{
		{
			ID:               uint(rand.Uint32()),
			UserID:           TestUserId,
			FromTime:         models.CustomTime(from.Add(10 * time.Minute)),
			ToTime:           models.CustomTime(from.Add(10 * time.Minute)),
			Projects:         []*models.SummaryItem{},
			Languages:        []*models.SummaryItem{},
			Editors:          []*models.SummaryItem{},
			OperatingSystems: []*models.SummaryItem{},
			Machines:         []*models.SummaryItem{},
		},
	}

	suite.SummaryRepository.On("GetByUserWithin", suite.TestUser, from, to).Return(summaries, nil)
	suite.DurationService.On("Get", from, summaries[0].FromTime.T(), suite.TestUser, mock.Anything, mock.Anything, false).Return(models.Durations{}, nil)
	suite.DurationService.On("Get", summaries[0].ToTime.T().Add(1*time.Second), to, suite.TestUser, mock.Anything, mock.Anything, false).Return(models.Durations{}, nil)

	result, err = sut.Retrieve(from, to, suite.TestUser, nil, nil)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.DurationService.AssertNumberOfCalls(suite.T(), "Get", 2)
}

func (suite *SummaryServiceTestSuite) TestSummaryService_Retrieve_DateRange() {
	sut := NewSummaryService(suite.SummaryRepository, suite.HeartbeatService, suite.DurationService, suite.AliasService, suite.ProjectLabelService)

	suite.ProjectLabelService.On("GetByUser", suite.TestUser.ID).Return([]*models.ProjectLabel{}, nil)

	var (
		summaries []*models.Summary
		from      time.Time
		to        time.Time
		result    *models.Summary
		err       error
	)

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
	suite.DurationService.On("Get", from, summaries[0].FromTime.T(), suite.TestUser, mock.Anything, mock.Anything, false).Return(models.Durations{}, nil)
	suite.DurationService.On("Get", summaries[0].ToTime.T(), to, suite.TestUser, mock.Anything, mock.Anything, false).Return(models.Durations{}, nil)

	result, err = sut.Retrieve(from, to, suite.TestUser, nil, nil)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), summaries[0].FromTime.T(), result.FromTime.T()) // actual first data date
	assert.Equal(suite.T(), summaries[0].ToTime.T(), result.ToTime.T())     // actual last data date
	suite.DurationService.AssertNumberOfCalls(suite.T(), "Get", 2)
}

func (suite *SummaryServiceTestSuite) TestSummaryService_Retrieve_DateRange_NoData() {
	sut := NewSummaryService(suite.SummaryRepository, suite.HeartbeatService, suite.DurationService, suite.AliasService, suite.ProjectLabelService)

	suite.ProjectLabelService.On("GetByUser", suite.TestUser.ID).Return([]*models.ProjectLabel{}, nil)

	var (
		from   time.Time
		to     time.Time
		result *models.Summary
		err    error
	)

	from, to = suite.TestStartTime.Add(-12*time.Hour), suite.TestStartTime.Add(12*time.Hour)

	suite.SummaryRepository.On("GetByUserWithin", suite.TestUser, from, to).Return([]*models.Summary{}, nil)
	suite.DurationService.On("Get", from, to, suite.TestUser, mock.Anything, mock.Anything, false).Return(models.Durations{}, nil)

	result, err = sut.Retrieve(from, to, suite.TestUser, nil, nil)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), from, result.FromTime.T())
	assert.Equal(suite.T(), to, result.ToTime.T())
	suite.DurationService.AssertNumberOfCalls(suite.T(), "Get", 1)
}

func (suite *SummaryServiceTestSuite) TestSummaryService_Aliased() {
	sut := NewSummaryService(suite.SummaryRepository, suite.HeartbeatService, suite.DurationService, suite.AliasService, suite.ProjectLabelService)

	suite.AliasService.On("InitializeUser", suite.TestUser.ID).Return(nil)
	suite.ProjectLabelService.On("GetByUser", suite.TestUser.ID).Return([]*models.ProjectLabel{}, nil)

	var (
		from   time.Time
		to     time.Time
		result *models.Summary
		err    error
	)

	from, to = suite.TestStartTime, suite.TestStartTime.Add(1*time.Hour)

	durations := filterDurations(from, to, suite.TestDurations)
	durations = append(durations, &models.Duration{
		UserID:          TestUserId,
		Project:         TestProject2,
		Language:        TestLanguageGo,
		Editor:          TestEditorGoland,
		OperatingSystem: TestOsLinux,
		Machine:         TestMachine1,
		Time:            models.CustomTime(durations[len(durations)-1].Time.T().Add(10 * time.Second)),
		Duration:        0, // not relevant here
	})

	suite.DurationService.On("Get", from, to, suite.TestUser, mock.Anything, mock.Anything, false).Return(durations, nil)
	suite.AliasService.On("InitializeUser", TestUserId).Return(nil)
	suite.AliasService.On("GetAliasOrDefault", TestUserId, mock.Anything, TestProject1).Return(TestProject2, nil)
	suite.AliasService.On("GetAliasOrDefault", TestUserId, mock.Anything, TestProject2).Return(TestProject2, nil)
	suite.AliasService.On("GetAliasOrDefault", TestUserId, mock.Anything, mock.Anything).Return("", nil)
	suite.ProjectLabelService.On("GetByUser", suite.TestUser.ID).Return(suite.TestLabels, nil).Once()

	result, err = sut.Aliased(from, to, suite.TestUser, sut.Summarize, nil, nil, false)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Zero(suite.T(), result.TotalTimeByKey(models.SummaryProject, TestProject1))
	assert.NotZero(suite.T(), result.TotalTimeByKey(models.SummaryProject, TestProject2))
	assert.Equal(suite.T(), 6, result.NumHeartbeats)
	assert.Nil(suite.T(), result.Branches)
}

func (suite *SummaryServiceTestSuite) TestSummaryService_Aliased_ProjectLabels() {
	sut := NewSummaryService(suite.SummaryRepository, suite.HeartbeatService, suite.DurationService, suite.AliasService, suite.ProjectLabelService)

	var (
		from   time.Time
		to     time.Time
		result *models.Summary
		err    error
	)

	from, to = suite.TestStartTime, suite.TestStartTime.Add(1*time.Hour)

	durations := filterDurations(from, to, suite.TestDurations)
	durations = append(durations, &models.Duration{
		UserID:          TestUserId,
		Project:         TestProject2,
		Language:        TestLanguageGo,
		Editor:          TestEditorGoland,
		OperatingSystem: TestOsLinux,
		Machine:         TestMachine1,
		Time:            models.CustomTime(durations[len(durations)-1].Time.T().Add(10 * time.Second)),
		Duration:        10 * time.Second,
	})

	suite.ProjectLabelService.On("GetByUser", suite.TestUser.ID).Return(suite.TestLabels, nil).Once()
	suite.DurationService.On("Get", from, to, suite.TestUser, mock.Anything, mock.Anything, false).Return(models.Durations(durations), nil)
	suite.AliasService.On("InitializeUser", TestUserId).Return(nil)
	suite.AliasService.On("GetAliasOrDefault", TestUserId, mock.Anything, TestProject1).Return(TestProject1, nil)
	suite.AliasService.On("GetAliasOrDefault", TestUserId, mock.Anything, TestProject2).Return(TestProject1, nil)
	suite.AliasService.On("GetAliasOrDefault", TestUserId, mock.Anything, mock.Anything).Return("", nil)

	result, err = sut.Aliased(from, to, suite.TestUser, sut.Summarize, nil, nil, false)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), 195*time.Second, result.TotalTimeByKey(models.SummaryLabel, TestProjectLabel1))
	assert.Equal(suite.T(), 6, result.NumHeartbeats)
}

func (suite *SummaryServiceTestSuite) TestSummaryService_Aliased_CustomTimeout() {
	sut := NewSummaryService(suite.SummaryRepository, suite.HeartbeatService, suite.DurationService, suite.AliasService, suite.ProjectLabelService)

	var (
		from time.Time
		to   time.Time
		err  error
	)

	from, to = suite.TestStartTime, suite.TestStartTime.Add(1*time.Hour)
	testSummary := &models.Summary{
		FromTime: models.CustomTime(from),
		ToTime:   models.CustomTime(to),
	}

	suite.SummaryRepository.On("GetByUserWithin", suite.TestUser, from, to).Return([]*models.Summary{testSummary}, nil)

	_, err = sut.Retrieve(from, to, suite.TestUser, nil, nil)
	assert.Nil(suite.T(), err)

	suite.DurationService.On("Get", from, to, suite.TestUser, mock.Anything, mock.Anything, false).Return(models.Durations([]*models.Duration{}), nil)

	customTimeout := 1337 * time.Minute
	_, err = sut.Retrieve(from, to, suite.TestUser, nil, &customTimeout)
	assert.Nil(suite.T(), err)
	suite.DurationService.AssertExpectations(suite.T())
}

func (suite *SummaryServiceTestSuite) TestSummaryService_Filters() {
	sut := NewSummaryService(suite.SummaryRepository, suite.HeartbeatService, suite.DurationService, suite.AliasService, suite.ProjectLabelService)

	suite.HeartbeatService.On("GetEntitySetByUser", models.SummaryProject, suite.TestUser.ID).Return([]string{TestProject1, TestProject2, TestProject3, TestProject4}, nil)
	suite.AliasService.On("InitializeUser", suite.TestUser.ID).Return(nil)
	suite.ProjectLabelService.On("GetByUser", suite.TestUser.ID).Return([]*models.ProjectLabel{}, nil)

	from, to := suite.TestStartTime, suite.TestStartTime.Add(1*time.Hour)
	filtersWithProject := models.NewFiltersWith(models.SummaryProject, TestProject1).With(models.SummaryLabel, TestProjectLabel1)
	filtersWithoutProject := models.NewFiltersWith(models.SummaryLabel, TestProjectLabel1)

	suite.DurationService.On("Get", from, to, suite.TestUser, mock.Anything, mock.Anything, false).Return(models.Durations{}, nil)
	suite.AliasService.On("GetByUserAndKeyAndType", TestUserId, TestProject1, models.SummaryProject).Return([]*models.Alias{
		// map any project starting with "som" to project 1
		{
			Type:  models.SummaryProject,
			Key:   TestProject1,
			Value: TestProject4[0:3] + "*", // wildcard alias
		},
	}, nil)
	suite.AliasService.On("GetByUserAndKeyAndType", TestUserId, TestProject3, models.SummaryProject).Return([]*models.Alias{
		// map project 2 to project 3
		{
			Type:  models.SummaryProject,
			Key:   TestProject3,
			Value: TestProject2,
		},
	}, nil)
	suite.AliasService.On("GetByUserAndKeyAndType", TestUserId, TestProject2, models.SummaryProject).Return([]*models.Alias{}, nil)
	suite.ProjectLabelService.On("GetByUserGroupedInverted", suite.TestUser.ID).Return(map[string][]*models.ProjectLabel{
		suite.TestLabels[0].Label: suite.TestLabels[0:2], // "private" -> ["test-project-1", "test-project-2"]
		suite.TestLabels[2].Label: suite.TestLabels[2:3], // "non-existing" -> ["test-project-3"]
	}, nil).Once()

	// first request: project details with filtering by project and label
	// when requesting project details, don't ignore data from other projects even though they'd match the filter's label (see https://github.com/muety/wakapi/issues/883)
	result1, _ := sut.Aliased(from, to, suite.TestUser, sut.Summarize, filtersWithProject, nil, false)
	effectiveFilters1 := suite.DurationService.Calls[0].Arguments[3].(*models.Filters)
	assert.NotNil(suite.T(), result1.Branches) // project filters were applied -> include branches
	assert.NotNil(suite.T(), result1.Entities) // project filters were applied -> include entities
	assert.Len(suite.T(), effectiveFilters1.Project, 2)
	assert.Contains(suite.T(), effectiveFilters1.Project, TestProject1) // because actually requested
	assert.Contains(suite.T(), effectiveFilters1.Project, TestProject4) // because of wildcard alias (whenever project 1 is requested, anything mapping to it must be included as well)
	assert.Contains(suite.T(), effectiveFilters1.Label, TestProjectLabel1)

	// second request: summary with filtering by label
	result2, _ := sut.Aliased(from, to, suite.TestUser, sut.Summarize, filtersWithoutProject, nil, false)
	effectiveFilters2 := suite.DurationService.Calls[1].Arguments[3].(*models.Filters)
	assert.NotNil(suite.T(), result2.Branches) // project filters were applied -> include branches
	assert.NotNil(suite.T(), result2.Entities) // project filters were applied -> include entities
	assert.Len(suite.T(), effectiveFilters2.Project, 3)
	assert.Contains(suite.T(), effectiveFilters2.Project, TestProject1) // because actually requested
	assert.Contains(suite.T(), effectiveFilters2.Project, TestProject2) // because covered by label
	assert.Contains(suite.T(), effectiveFilters2.Project, TestProject4) // because of wildcard alias (whenever project 1 is requested, anything mapping to it must be included as well)
	assert.Contains(suite.T(), effectiveFilters2.Label, TestProjectLabel1)
}

func (suite *SummaryServiceTestSuite) TestSummaryService_getMissingIntervals() {
	sut := NewSummaryService(suite.SummaryRepository, suite.HeartbeatService, suite.DurationService, suite.AliasService, suite.ProjectLabelService)

	from1, _ := time.Parse(time.RFC822, "25 Mar 22 11:00 UTC")
	to1, _ := time.Parse(time.RFC822, "25 Mar 22 13:00 UTC")
	from2, _ := time.Parse(time.RFC822, "25 Mar 22 15:00 UTC")
	to2, _ := time.Parse(time.RFC822, "26 Mar 22 00:00 UTC")

	summaries := []*models.Summary{
		{FromTime: models.CustomTime(from1), ToTime: models.CustomTime(to1)},
		{FromTime: models.CustomTime(from2), ToTime: models.CustomTime(to2)},
	}

	r1 := sut.getMissingIntervals(from1, to1, summaries, true)
	assert.Empty(suite.T(), r1)

	r2 := sut.getMissingIntervals(from1, from1, summaries, true)
	assert.Empty(suite.T(), r2)

	// non-precise mode will not return intra-day intervals
	// we might want to change this ...
	r3 := sut.getMissingIntervals(from1, to2, summaries, false)
	assert.Len(suite.T(), r3, 0)

	r4 := sut.getMissingIntervals(from1, to2, summaries, true)
	assert.Len(suite.T(), r4, 1)
	assert.Equal(suite.T(), to1, r4[0].Start)
	assert.Equal(suite.T(), from2, r4[0].End)

	r5 := sut.getMissingIntervals(from1.Add(-time.Hour), to2.Add(time.Hour), summaries, true)
	assert.Len(suite.T(), r5, 3)
	assert.Equal(suite.T(), from1.Add(-time.Hour), r5[0].Start)
	assert.Equal(suite.T(), from1, r5[0].End)
	assert.Equal(suite.T(), to1, r5[1].Start)
	assert.Equal(suite.T(), from2, r5[1].End)
	assert.Equal(suite.T(), to2, r5[2].Start)
	assert.Equal(suite.T(), to2.Add(time.Hour), r5[2].End)
}

func filterDurations(from, to time.Time, durations models.Durations) models.Durations {
	filtered := make([]*models.Duration, 0, len(durations))
	for _, d := range durations {
		if (d.Time.T().Equal(from) || d.Time.T().After(from)) && d.Time.T().Before(to) {
			filtered = append(filtered, d)
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
