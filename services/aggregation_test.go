package services

import (
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	datastructure "github.com/duke-git/lancet/v2/datastructure/set"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/models/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type AggregationServiceTestSuite struct {
	suite.Suite
	UserService      *mocks.UserServiceMock
	SummaryService   *mocks.SummaryServiceMock
	HeartbeatService *mocks.HeartbeatServiceMock
	DurationService  *mocks.DurationServiceMock
}

func (suite *AggregationServiceTestSuite) SetupSuite() {
	config.Set(config.Empty())
}

func (suite *AggregationServiceTestSuite) BeforeTest(suiteName, testName string) {
	suite.UserService = new(mocks.UserServiceMock)
	suite.SummaryService = new(mocks.SummaryServiceMock)
	suite.HeartbeatService = new(mocks.HeartbeatServiceMock)
	suite.DurationService = new(mocks.DurationServiceMock)
}

func TestAggregationServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AggregationServiceTestSuite))
}

func (suite *AggregationServiceTestSuite) TestAggregationService_GenerateUserJobs_Timezones() {
	nyTz, err := time.LoadLocation("America/New_York")
	suite.Require().NoError(err)
	tokyoTz, err := time.LoadLocation("Asia/Tokyo")
	suite.Require().NoError(err)

	now := time.Now()

	fromUTC := now.AddDate(0, 0, -5).UTC() // pick reference time 5 days ago to ensure multiple jobs are generated

	tests := []struct {
		name     string
		user     *models.User
		tz       *time.Location
		zoneName string
	}{
		{
			name:     "UTC user",
			user:     &models.User{ID: "u_utc", Location: "UTC"},
			tz:       time.UTC,
			zoneName: "UTC",
		},
		{
			name:     "New York user",
			user:     &models.User{ID: "u_ny", Location: "America/New_York"},
			tz:       nyTz,
			zoneName: "America/New_York",
		},
		{
			name:     "Tokyo user",
			user:     &models.User{ID: "u_tokyo", Location: "Asia/Tokyo"},
			tz:       tokyoTz,
			zoneName: "Asia/Tokyo",
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			jobs := slices.Collect(generateUserJobs(tt.user, fromUTC))
			assert.NotEmpty(t, jobs)

			endOfYesterday := getStartOfToday(tt.tz).Add(-1 * time.Second)

			for _, job := range jobs {
				assert.Equal(t, tt.user, job.User)

				// boundaries must be in user's timezone
				assert.Equal(t, tt.tz.String(), job.From.Location().String())
				assert.Equal(t, tt.tz.String(), job.To.Location().String())

				// boundaries must be exact midnights (00:00:00) in user's timezone
				assert.Equal(t, 0, job.From.Hour())
				assert.Equal(t, 0, job.From.Minute())
				assert.Equal(t, 0, job.From.Second())
				assert.Equal(t, 0, job.From.Nanosecond())

				assert.Equal(t, 0, job.To.Hour())
				assert.Equal(t, 0, job.To.Minute())
				assert.Equal(t, 0, job.To.Second())
				assert.Equal(t, 0, job.To.Nanosecond())

				// each job must span 1 day (allowing for dst transitions of 23-25 hours)
				diffHours := job.To.Sub(job.From).Hours()
				assert.True(t, diffHours >= 23 && diffHours <= 25)

				// cannot exceed end of yesterday in user's timezone
				assert.True(t, job.From.Before(endOfYesterday))
				assert.True(t, job.To.Before(getStartOfToday(tt.tz)))
			}

			// sequential jobs must be contiguous
			for i := 0; i < len(jobs)-1; i++ {
				assert.True(t, jobs[i].To.Equal(jobs[i+1].From))
			}
		})
	}
}

func (suite *AggregationServiceTestSuite) TestAggregationService_GenerateUserJobs_MidnightDSTTransition() {
	pyTz, err := time.LoadLocation("America/Asuncion")
	suite.Require().NoError(err)

	user := &models.User{ID: "u_py", Location: "America/Asuncion"}

	// Start 2 days before the Oct 5-6 2024 midnight DST transition in America/Asuncion
	from := time.Date(2024, 10, 4, 12, 0, 0, 0, pyTz)

	jobs := slices.Collect(generateUserJobs(user, from))
	suite.NotEmpty(jobs)

	// Ensure jobs advance monotonically and sequentially without getting stuck
	for i, job := range jobs {
		suite.True(job.To.After(job.From), "job %d did not advance: from=%s, to=%s", i, job.From, job.To)
		if i < len(jobs)-1 {
			suite.True(job.To.Equal(jobs[i+1].From), "job %d not contiguous with job %d", i, i+1)
		}
	}
}

func (suite *AggregationServiceTestSuite) TestAggregationService_GetStartOfToday() {
	tokyoTz, err := time.LoadLocation("Asia/Tokyo")
	suite.Require().NoError(err)

	nyTz, err := time.LoadLocation("America/New_York")
	suite.Require().NoError(err)

	todayTokyo := getStartOfToday(tokyoTz)
	todayNY := getStartOfToday(nyTz)

	suite.Equal(0, todayTokyo.Hour())
	suite.Equal(0, todayTokyo.Minute())
	suite.Equal(0, todayTokyo.Second())
	suite.Equal("Asia/Tokyo", todayTokyo.Location().String())

	suite.Equal(0, todayNY.Hour())
	suite.Equal(0, todayNY.Minute())
	suite.Equal(0, todayNY.Second())
	suite.Equal("America/New_York", todayNY.Location().String())
}

func (suite *AggregationServiceTestSuite) TestAggregationService_LockUsers_EmptySetLocksAll() {
	sut := NewAggregationService(suite.UserService, suite.SummaryService, suite.HeartbeatService, suite.DurationService)

	suite.False(sut.IsLocked("u1"))
	suite.False(sut.IsLocked("u2"))

	err := sut.lockUsers(datastructure.New[string]())
	suite.NoError(err)
	suite.True(sut.IsLocked("u1"))
	suite.True(sut.IsLocked("u2"))

	err = sut.lockUsers(datastructure.New[string]())
	suite.Error(err)

	err = sut.lockUsers(datastructure.New("u1"))
	suite.Error(err)

	sut.unlockUsers(datastructure.New[string]())
	suite.False(sut.IsLocked("u1"))
	suite.False(sut.IsLocked("u2"))

	err = sut.lockUsers(nil)
	suite.NoError(err)
	suite.True(sut.IsLocked("u1"))
	suite.True(sut.IsLocked("u2"))

	sut.unlockUsers(nil)
	suite.False(sut.IsLocked("u1"))
	suite.False(sut.IsLocked("u2"))
}

func (suite *AggregationServiceTestSuite) TestAggregationService_LockUsers_SpecificUsers() {
	sut := NewAggregationService(suite.UserService, suite.SummaryService, suite.HeartbeatService, suite.DurationService)

	err := sut.lockUsers(datastructure.New("u1"))
	suite.NoError(err)
	suite.True(sut.IsLocked("u1"))
	suite.False(sut.IsLocked("u2"))

	err = sut.lockUsers(datastructure.New("u1"))
	suite.Error(err)

	err = sut.lockUsers(datastructure.New[string]())
	suite.Error(err)
	err = sut.lockUsers(nil)
	suite.Error(err)

	err = sut.lockUsers(datastructure.New("u2"))
	suite.NoError(err)
	suite.True(sut.IsLocked("u1"))
	suite.True(sut.IsLocked("u2"))

	sut.unlockUsers(datastructure.New("u1"))
	suite.False(sut.IsLocked("u1"))
	suite.True(sut.IsLocked("u2"))

	err = sut.lockUsers(datastructure.New[string]())
	suite.Error(err)

	sut.unlockUsers(datastructure.New("u2"))
	suite.False(sut.IsLocked("u2"))

	err = sut.lockUsers(datastructure.New[string]())
	suite.NoError(err)
	suite.True(sut.IsLocked("u1"))
	suite.True(sut.IsLocked("u2"))
	sut.unlockUsers(datastructure.New[string]())
}

func (suite *AggregationServiceTestSuite) TestAggregationService_LockUsers_ConcurrentMutualExclusion() {
	sut := NewAggregationService(suite.UserService, suite.SummaryService, suite.HeartbeatService, suite.DurationService)

	const goroutines = 50
	var successCount atomic.Int32
	var wg sync.WaitGroup

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			if err := sut.lockUsers(datastructure.New("u1")); err == nil {
				successCount.Add(1)
			}
		}()
	}
	wg.Wait()

	suite.Equal(int32(1), successCount.Load())
	suite.True(sut.IsLocked("u1"))

	sut.unlockUsers(datastructure.New("u1"))
	suite.False(sut.IsLocked("u1"))
}

func (suite *AggregationServiceTestSuite) TestAggregationService_LockUsers_ConcurrentGlobalVersusIndividual() {
	sut := NewAggregationService(suite.UserService, suite.SummaryService, suite.HeartbeatService, suite.DurationService)

	const total = 50
	var globalSuccess atomic.Int32
	var individualSuccess atomic.Int32
	var wg sync.WaitGroup

	wg.Add(total)
	for i := 0; i < total; i++ {
		if i%2 == 0 {
			go func() {
				defer wg.Done()
				if err := sut.lockUsers(datastructure.New[string]()); err == nil {
					globalSuccess.Add(1)
				}
			}()
		} else {
			go func() {
				defer wg.Done()
				if err := sut.lockUsers(datastructure.New("u1")); err == nil {
					individualSuccess.Add(1)
				}
			}()
		}
	}
	wg.Wait()

	if globalSuccess.Load() > 0 {
		suite.Equal(int32(1), globalSuccess.Load())
		suite.Equal(int32(0), individualSuccess.Load())
		sut.unlockUsers(datastructure.New[string]())
	} else {
		suite.Equal(int32(1), individualSuccess.Load())
		suite.Equal(int32(0), globalSuccess.Load())
		sut.unlockUsers(datastructure.New("u1"))
	}
}

func (suite *AggregationServiceTestSuite) TestAggregationService_AggregateSummaries_WaitAndLock() {
	yesterdayMidnight := getStartOfToday(time.UTC).AddDate(0, 0, -1)
	twoDaysAgoMidnight := yesterdayMidnight.AddDate(0, 0, -1)

	testUser := &models.User{
		ID:       "u_lock_test",
		Location: "UTC",
	}

	suite.UserService.On("GetManyMapped").Return(map[string]*models.User{
		"u_lock_test": testUser,
	}, nil)

	suite.SummaryService.On("GetLatestByUser").Return([]*models.TimeByUser{
		{User: "u_lock_test", Time: models.CustomTime(twoDaysAgoMidnight)},
	}, nil)

	suite.HeartbeatService.On("GetFirstAll").Return([]*models.TimeByUser{}, nil)

	inJobChan := make(chan struct{}, 1)
	finishJobChan := make(chan struct{})

	suite.SummaryService.On("Summarize", mock.Anything, mock.Anything, testUser, (*models.Filters)(nil), (*time.Duration)(nil)).
		Run(func(args mock.Arguments) {
			select {
			case inJobChan <- struct{}{}:
			default:
			}
			<-finishJobChan
		}).
		Return(&models.Summary{UserID: "u_lock_test"}, nil)

	suite.SummaryService.On("Insert", mock.Anything).Return(nil)

	sut := NewAggregationService(suite.UserService, suite.SummaryService, suite.HeartbeatService, suite.DurationService)

	aggDone := make(chan error, 1)
	go func() {
		aggDone <- sut.AggregateSummaries(datastructure.New("u_lock_test"), types.DurationsRegenerationNone)
	}()

	<-inJobChan
	suite.True(sut.IsLocked("u_lock_test"))

	err := sut.lockUsers(datastructure.New("u_lock_test"))
	suite.Error(err)

	err = sut.AggregateSummaries(datastructure.New("u_lock_test"), types.DurationsRegenerationNone)
	suite.Error(err)

	close(finishJobChan)

	err = <-aggDone
	suite.NoError(err)
	suite.False(sut.IsLocked("u_lock_test"))

	suite.SummaryService.AssertExpectations(suite.T())
	suite.UserService.AssertExpectations(suite.T())
}

func (suite *AggregationServiceTestSuite) TestAggregationService_AggregateSummaries_UserWithoutPriorSummaries() {
	yesterdayMidnight := getStartOfToday(time.UTC).AddDate(0, 0, -1)
	twoDaysAgoMidnight := yesterdayMidnight.AddDate(0, 0, -1)

	testUser := &models.User{
		ID:       "u_new_user",
		Location: "UTC",
	}

	suite.UserService.On("GetManyMapped").Return(map[string]*models.User{
		"u_new_user": testUser,
	}, nil)

	// User has NO prior summaries (e.g. empty result or not in list)
	suite.SummaryService.On("GetLatestByUser").Return([]*models.TimeByUser{}, nil)

	// User has heartbeats from two days ago
	suite.HeartbeatService.On("GetFirstAll").Return([]*models.TimeByUser{
		{User: "u_new_user", Time: models.CustomTime(twoDaysAgoMidnight)},
	}, nil)

	summarized := atomic.Int32{}
	suite.SummaryService.On("Summarize", mock.Anything, mock.Anything, testUser, (*models.Filters)(nil), (*time.Duration)(nil)).
		Run(func(args mock.Arguments) {
			summarized.Add(1)
		}).
		Return(&models.Summary{UserID: "u_new_user"}, nil)

	suite.SummaryService.On("Insert", mock.Anything).Return(nil)

	sut := NewAggregationService(suite.UserService, suite.SummaryService, suite.HeartbeatService, suite.DurationService)

	err := sut.AggregateSummaries(datastructure.New("u_new_user"), types.DurationsRegenerationNone)
	suite.NoError(err)
	suite.Equal(int32(2), summarized.Load(), "summary aggregation should have generated jobs from first heartbeat")

	suite.SummaryService.AssertExpectations(suite.T())
	suite.HeartbeatService.AssertExpectations(suite.T())
	suite.UserService.AssertExpectations(suite.T())
}

func (suite *AggregationServiceTestSuite) TestAggregationService_AggregateSummaries_IncrementalDurationsRegeneration() {
	suite.assertDurationsRegenBeforeSummarize(types.DurationsRegenerationIncremental, false, "u_dur_regen")
}

func (suite *AggregationServiceTestSuite) TestAggregationService_AggregateSummaries_FullDurationsRegeneration() {
	suite.assertDurationsRegenBeforeSummarize(types.DurationsRegenerationFull, true, "u_dur_regen_full")
}

// assertDurationsRegenBeforeSummarize asserts that a user's summary jobs only start after user's durations regeneration has finished,
// that the aggregation lock spans the whole run, and that AggregateSummaries returns only once both durations regeneration and summaries have completed.
func (suite *AggregationServiceTestSuite) assertDurationsRegenBeforeSummarize(mode types.DurationsRegenerationMode, expectForceAll bool, userId string) {
	yesterdayMidnight := getStartOfToday(time.UTC).AddDate(0, 0, -1)
	twoDaysAgoMidnight := yesterdayMidnight.AddDate(0, 0, -1)

	testUser := &models.User{
		ID:       userId,
		Location: "UTC",
	}

	suite.UserService.On("GetManyMapped").Return(map[string]*models.User{
		userId: testUser,
	}, nil)

	suite.SummaryService.On("GetLatestByUser").Return([]*models.TimeByUser{
		{User: userId, Time: models.CustomTime(twoDaysAgoMidnight)},
	}, nil)

	suite.HeartbeatService.On("GetFirstAll").Return([]*models.TimeByUser{}, nil)

	durStarted := make(chan struct{}, 1)
	durFinish := make(chan struct{})
	sumStarted := make(chan struct{}, 1)
	var sumCompleted atomic.Int32

	// expect durations regeneration to run with the force flag derived from the mode
	suite.DurationService.On("Regenerate", testUser, expectForceAll).
		Run(func(args mock.Arguments) {
			select {
			case durStarted <- struct{}{}:
			default:
			}
			<-durFinish
		}).
		Return(nil)

	suite.SummaryService.On("Summarize", mock.Anything, mock.Anything, testUser, (*models.Filters)(nil), (*time.Duration)(nil)).
		Run(func(args mock.Arguments) {
			select {
			case sumStarted <- struct{}{}:
			default:
			}
		}).
		Return(&models.Summary{UserID: userId}, nil)

	// insert is called by the summary job right after summarizing, and before its wgGlobal.Done() fires
	suite.SummaryService.On("Insert", mock.Anything).
		Run(func(args mock.Arguments) {
			sumCompleted.Add(1)
		}).
		Return(nil)

	sut := NewAggregationService(suite.UserService, suite.SummaryService, suite.HeartbeatService, suite.DurationService)

	aggDone := make(chan error, 1)
	go func() {
		aggDone <- sut.AggregateSummaries(datastructure.New(userId), mode)
	}()

	<-durStarted

	// summary aggregation for a user must not start before that user's durations regeneration has finished
	time.Sleep(500 * time.Millisecond)
	suite.Zero(len(sumStarted), "summary aggregation started before durations regeneration finished")

	// the aggregation lock must still be held for the entire duration of the run
	suite.True(sut.IsLocked(userId))
	suite.Error(sut.AggregateSummaries(datastructure.New(userId), types.DurationsRegenerationNone))

	close(durFinish)

	select {
	case <-sumStarted:
	case <-time.After(5 * time.Second):
		suite.Fail("summary aggregation did not run after durations regeneration finished")
	}

	select {
	case err := <-aggDone:
		suite.NoError(err)
	case <-time.After(5 * time.Second):
		suite.Fail("aggregate summaries did not return")
	}

	// once AggregateSummaries returned, all summary jobs must have run to completion and the lock be released
	suite.Equal(int32(2), sumCompleted.Load(), "aggregate summaries returned before all summary jobs completed")
	suite.False(sut.IsLocked(userId))

	suite.DurationService.AssertExpectations(suite.T())
	suite.SummaryService.AssertExpectations(suite.T())
	suite.UserService.AssertExpectations(suite.T())
}

func (suite *AggregationServiceTestSuite) TestAggregationService_AggregateDurations_WaitAndLock() {
	testUser := &models.User{
		ID:       "u_dur_test",
		Location: "UTC",
	}

	suite.UserService.On("GetManyMapped").Return(map[string]*models.User{
		"u_dur_test": testUser,
	}, nil)

	inJobChan := make(chan struct{}, 1)
	finishJobChan := make(chan struct{})

	suite.DurationService.On("Regenerate", mock.Anything, true).
		Run(func(args mock.Arguments) {
			select {
			case inJobChan <- struct{}{}:
			default:
			}
			<-finishJobChan
		}).
		Return(nil)

	sut := NewAggregationService(suite.UserService, suite.SummaryService, suite.HeartbeatService, suite.DurationService)

	aggDone := make(chan error, 1)
	go func() {
		aggDone <- sut.AggregateDurations(datastructure.New("u_dur_test"))
	}()

	<-inJobChan
	suite.True(sut.IsLocked("u_dur_test"))

	err := sut.lockUsers(datastructure.New("u_dur_test"))
	suite.Error(err)

	close(finishJobChan)

	err = <-aggDone
	suite.NoError(err)
	suite.False(sut.IsLocked("u_dur_test"))

	suite.DurationService.AssertExpectations(suite.T())
	suite.UserService.AssertExpectations(suite.T())
}
