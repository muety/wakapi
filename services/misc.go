package services

import (
	"github.com/emvi/logbuch"
	"github.com/muety/wakapi/config"
	"go.uber.org/atomic"
	"runtime"
	"strconv"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/muety/wakapi/models"
)

type MiscService struct {
	config          *config.Config
	userService     IUserService
	summaryService  ISummaryService
	keyValueService IKeyValueService
	jobCount        atomic.Uint32
}

func NewMiscService(userService IUserService, summaryService ISummaryService, keyValueService IKeyValueService) *MiscService {
	return &MiscService{
		config:          config.Get(),
		userService:     userService,
		summaryService:  summaryService,
		keyValueService: keyValueService,
	}
}

type CountTotalTimeJob struct {
	UserID  string
	NumJobs int
}

type CountTotalTimeResult struct {
	UserId string
	Total  time.Duration
}

func (srv *MiscService) ScheduleCountTotalTime() {
	// Run once initially
	if err := srv.runCountTotalTime(); err != nil {
		logbuch.Error("failed to run CountTotalTimeJob: %v", err)
	}

	s := gocron.NewScheduler(time.Local)
	s.Every(1).Hour().Do(srv.runCountTotalTime)
	s.StartBlocking()
}

func (srv *MiscService) runCountTotalTime() error {
	jobs := make(chan *CountTotalTimeJob)
	results := make(chan *CountTotalTimeResult)

	defer close(jobs)

	for i := 0; i < runtime.NumCPU(); i++ {
		go srv.countTotalTimeWorker(jobs, results)
	}

	go srv.persistTotalTimeWorker(results)

	// generate the jobs
	if users, err := srv.userService.GetAll(); err == nil {
		for _, u := range users {
			jobs <- &CountTotalTimeJob{
				UserID:  u.ID,
				NumJobs: len(users),
			}
		}
	} else {
		return err
	}

	return nil
}

func (srv *MiscService) countTotalTimeWorker(jobs <-chan *CountTotalTimeJob, results chan<- *CountTotalTimeResult) {
	for job := range jobs {
		if result, err := srv.summaryService.Aliased(time.Time{}, time.Now(), &models.User{ID: job.UserID}, srv.summaryService.Retrieve); err != nil {
			logbuch.Error("failed to count total for user %s: %v", job.UserID, err)
		} else {
			logbuch.Info("successfully counted total for user %s", job.UserID)
			results <- &CountTotalTimeResult{
				UserId: job.UserID,
				Total:  result.TotalTime(),
			}
		}
		if srv.jobCount.Inc() == uint32(job.NumJobs) {
			srv.jobCount.Store(0)
			close(results)
		}
	}
}

func (srv *MiscService) persistTotalTimeWorker(results <-chan *CountTotalTimeResult) {
	var c int
	var total time.Duration
	for result := range results {
		total += result.Total
		c++
	}

	if err := srv.keyValueService.PutString(&models.KeyStringValue{
		Key:   config.KeyLatestTotalTime,
		Value: total.String(),
	}); err != nil {
		logbuch.Error("failed to save total time count: %v", err)
	}

	if err := srv.keyValueService.PutString(&models.KeyStringValue{
		Key:   config.KeyLatestTotalUsers,
		Value: strconv.Itoa(c),
	}); err != nil {
		logbuch.Error("failed to save total users count: %v", err)
	}
}
