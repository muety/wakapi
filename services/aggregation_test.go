package services

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	datastructure "github.com/duke-git/lancet/v2/datastructure/set"
	"github.com/muety/artifex/v2"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAggregationService_GenerateUserJobs_Timezones(t *testing.T) {
	nyTz, err := time.LoadLocation("America/New_York")
	assert.NoError(t, err)
	tokyoTz, err := time.LoadLocation("Asia/Tokyo")
	assert.NoError(t, err)

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
		t.Run(tt.name, func(t *testing.T) {
			jobs := generateUserJobs(tt.user, fromUTC)
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

func TestAggregationService_GetStartOfToday(t *testing.T) {
	tokyoTz, err := time.LoadLocation("Asia/Tokyo")
	assert.NoError(t, err)

	nyTz, err := time.LoadLocation("America/New_York")
	assert.NoError(t, err)

	todayTokyo := getStartOfToday(tokyoTz)
	todayNY := getStartOfToday(nyTz)

	assert.Equal(t, 0, todayTokyo.Hour())
	assert.Equal(t, 0, todayTokyo.Minute())
	assert.Equal(t, 0, todayTokyo.Second())
	assert.Equal(t, "Asia/Tokyo", todayTokyo.Location().String())

	assert.Equal(t, 0, todayNY.Hour())
	assert.Equal(t, 0, todayNY.Minute())
	assert.Equal(t, 0, todayNY.Second())
	assert.Equal(t, "America/New_York", todayNY.Location().String())
}

func TestAggregationService_LockUsers_EmptySetLocksAll(t *testing.T) {
	srv := &AggregationService{
		inProgress: datastructure.New[string](),
	}

	assert.False(t, srv.IsLocked("u1"))
	assert.False(t, srv.IsLocked("u2"))

	err := srv.lockUsers(datastructure.New[string]())
	assert.NoError(t, err)
	assert.True(t, srv.IsLocked("u1"))
	assert.True(t, srv.IsLocked("u2"))

	err = srv.lockUsers(datastructure.New[string]())
	assert.Error(t, err)

	err = srv.lockUsers(datastructure.New("u1"))
	assert.Error(t, err)

	srv.unlockUsers(datastructure.New[string]())
	assert.False(t, srv.IsLocked("u1"))
	assert.False(t, srv.IsLocked("u2"))

	err = srv.lockUsers(nil)
	assert.NoError(t, err)
	assert.True(t, srv.IsLocked("u1"))
	assert.True(t, srv.IsLocked("u2"))

	srv.unlockUsers(nil)
	assert.False(t, srv.IsLocked("u1"))
	assert.False(t, srv.IsLocked("u2"))
}

func TestAggregationService_LockUsers_SpecificUsers(t *testing.T) {
	srv := &AggregationService{
		inProgress: datastructure.New[string](),
	}

	err := srv.lockUsers(datastructure.New("u1"))
	assert.NoError(t, err)
	assert.True(t, srv.IsLocked("u1"))
	assert.False(t, srv.IsLocked("u2"))

	err = srv.lockUsers(datastructure.New("u1"))
	assert.Error(t, err)

	err = srv.lockUsers(datastructure.New[string]())
	assert.Error(t, err)
	err = srv.lockUsers(nil)
	assert.Error(t, err)

	err = srv.lockUsers(datastructure.New("u2"))
	assert.NoError(t, err)
	assert.True(t, srv.IsLocked("u1"))
	assert.True(t, srv.IsLocked("u2"))

	srv.unlockUsers(datastructure.New("u1"))
	assert.False(t, srv.IsLocked("u1"))
	assert.True(t, srv.IsLocked("u2"))

	err = srv.lockUsers(datastructure.New[string]())
	assert.Error(t, err)

	srv.unlockUsers(datastructure.New("u2"))
	assert.False(t, srv.IsLocked("u2"))

	err = srv.lockUsers(datastructure.New[string]())
	assert.NoError(t, err)
	assert.True(t, srv.IsLocked("u1"))
	assert.True(t, srv.IsLocked("u2"))
	srv.unlockUsers(datastructure.New[string]())
}

func TestAggregationService_LockUsers_ConcurrentMutualExclusion(t *testing.T) {
	srv := &AggregationService{
		inProgress: datastructure.New[string](),
	}

	const goroutines = 50
	var successCount atomic.Int32
	var wg sync.WaitGroup

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			if err := srv.lockUsers(datastructure.New("u1")); err == nil {
				successCount.Add(1)
			}
		}()
	}
	wg.Wait()

	assert.Equal(t, int32(1), successCount.Load())
	assert.True(t, srv.IsLocked("u1"))

	srv.unlockUsers(datastructure.New("u1"))
	assert.False(t, srv.IsLocked("u1"))
}

func TestAggregationService_LockUsers_ConcurrentGlobalVersusIndividual(t *testing.T) {
	srv := &AggregationService{
		inProgress: datastructure.New[string](),
	}

	const total = 50
	var globalSuccess atomic.Int32
	var individualSuccess atomic.Int32
	var wg sync.WaitGroup

	wg.Add(total)
	for i := 0; i < total; i++ {
		if i%2 == 0 {
			go func() {
				defer wg.Done()
				if err := srv.lockUsers(datastructure.New[string]()); err == nil {
					globalSuccess.Add(1)
				}
			}()
		} else {
			go func() {
				defer wg.Done()
				if err := srv.lockUsers(datastructure.New("u1")); err == nil {
					individualSuccess.Add(1)
				}
			}()
		}
	}
	wg.Wait()

	if globalSuccess.Load() > 0 {
		assert.Equal(t, int32(1), globalSuccess.Load())
		assert.Equal(t, int32(0), individualSuccess.Load())
		srv.unlockUsers(datastructure.New[string]())
	} else {
		assert.Equal(t, int32(1), individualSuccess.Load())
		assert.Equal(t, int32(0), globalSuccess.Load())
		srv.unlockUsers(datastructure.New("u1"))
	}
}

func TestAggregationService_AggregateSummaries_WaitAndLock(t *testing.T) {
	config.Set(config.Empty())
	cfg := config.Get()

	queueSummary := artifex.NewDispatcher(2, 64)
	queueSummary.Start()
	defer queueSummary.Stop()

	queueDuration := artifex.NewDispatcher(2, 64)
	queueDuration.Start()
	defer queueDuration.Stop()

	testUser := &models.User{
		ID:       "u_lock_test",
		Location: "UTC",
	}

	mockUserSrvc := &mocks.UserServiceMock{}
	mockSummarySrvc := &mocks.SummaryServiceMock{}
	mockHeartbeatSrvc := &mocks.HeartbeatServiceMock{}
	mockDurationSrvc := &mocks.DurationServiceMock{}

	yesterdayMidnight := getStartOfToday(time.UTC).AddDate(0, 0, -1)
	twoDaysAgoMidnight := yesterdayMidnight.AddDate(0, 0, -1)

	mockUserSrvc.On("GetManyMapped").Return(map[string]*models.User{
		"u_lock_test": testUser,
	}, nil)

	mockSummarySrvc.On("GetLatestByUser").Return([]*models.TimeByUser{
		{User: "u_lock_test", Time: models.CustomTime(twoDaysAgoMidnight)},
	}, nil)

	mockHeartbeatSrvc.On("GetFirstAll").Return([]*models.TimeByUser{}, nil)

	inJobChan := make(chan struct{}, 1)
	finishJobChan := make(chan struct{})

	mockSummarySrvc.On("Summarize", mock.Anything, mock.Anything, testUser, (*time.Duration)(nil), (*models.Filters)(nil)).
		Run(func(args mock.Arguments) {
			select {
			case inJobChan <- struct{}{}:
			default:
			}
			<-finishJobChan
		}).
		Return(&models.Summary{UserID: "u_lock_test"}, nil)

	mockSummarySrvc.On("Insert", mock.Anything).Return(nil)

	srv := &AggregationService{
		config:                cfg,
		userService:           mockUserSrvc,
		summaryService:        mockSummarySrvc,
		heartbeatService:      mockHeartbeatSrvc,
		durationService:       mockDurationSrvc,
		inProgress:            datastructure.New[string](),
		queueSummaryWorkers:   queueSummary,
		queuedDurationWorkers: queueDuration,
	}

	aggDone := make(chan error, 1)
	go func() {
		aggDone <- srv.AggregateSummaries(datastructure.New("u_lock_test"), false)
	}()

	<-inJobChan
	assert.True(t, srv.IsLocked("u_lock_test"))

	err := srv.lockUsers(datastructure.New("u_lock_test"))
	assert.Error(t, err)

	err = srv.AggregateSummaries(datastructure.New("u_lock_test"), false)
	assert.Error(t, err)

	close(finishJobChan)

	err = <-aggDone
	assert.NoError(t, err)
	assert.False(t, srv.IsLocked("u_lock_test"))

	mockSummarySrvc.AssertExpectations(t)
	mockUserSrvc.AssertExpectations(t)
}

func TestAggregationService_AggregateDurations_WaitAndLock(t *testing.T) {
	config.Set(config.Empty())
	cfg := config.Get()

	queueDuration := artifex.NewDispatcher(2, 64)
	queueDuration.Start()
	defer queueDuration.Stop()

	testUser := &models.User{
		ID:       "u_dur_test",
		Location: "UTC",
	}

	mockUserSrvc := &mocks.UserServiceMock{}
	mockDurationSrvc := &mocks.DurationServiceMock{}

	mockUserSrvc.On("GetManyMapped").Return(map[string]*models.User{
		"u_dur_test": testUser,
	}, nil)

	inJobChan := make(chan struct{}, 1)
	finishJobChan := make(chan struct{})

	mockDurationSrvc.On("Regenerate", mock.Anything, true).
		Run(func(args mock.Arguments) {
			select {
			case inJobChan <- struct{}{}:
			default:
			}
			<-finishJobChan
		}).
		Return(nil)

	srv := &AggregationService{
		config:                cfg,
		userService:           mockUserSrvc,
		durationService:       mockDurationSrvc,
		inProgress:            datastructure.New[string](),
		queuedDurationWorkers: queueDuration,
	}

	aggDone := make(chan error, 1)
	go func() {
		aggDone <- srv.AggregateDurations(datastructure.New("u_dur_test"))
	}()

	<-inJobChan
	assert.True(t, srv.IsLocked("u_dur_test"))

	err := srv.lockUsers(datastructure.New("u_dur_test"))
	assert.Error(t, err)

	close(finishJobChan)

	err = <-aggDone
	assert.NoError(t, err)
	assert.False(t, srv.IsLocked("u_dur_test"))

	mockDurationSrvc.AssertExpectations(t)
	mockUserSrvc.AssertExpectations(t)
}
