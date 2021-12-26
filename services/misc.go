package services

import (
	"github.com/emvi/logbuch"
	"github.com/muety/wakapi/config"
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
		logbuch.Fatal("failed to run CountTotalTimeJob: %v", err)
	}

	s := gocron.NewScheduler(time.Local)
	s.Every(1).Hour().Do(srv.runCountTotalTime)
	s.StartBlocking()
}

func (srv *MiscService) runCountTotalTime() error {
	users, err := srv.userService.GetAll()
	if err != nil {
		return err
	}

	jobs := make(chan *CountTotalTimeJob, len(users))
	results := make(chan *CountTotalTimeResult, len(users))

	for _, u := range users {
		jobs <- &CountTotalTimeJob{
			UserID:  u.ID,
			NumJobs: len(users),
		}
	}
	close(jobs)

	for i := 0; i < runtime.NumCPU(); i++ {
		go srv.countTotalTimeWorker(jobs, results)
	}

	// persist
	var i int
	var total time.Duration
	for i = 0; i < len(users); i++ {
		result := <-results
		total += result.Total
	}
	close(results)

	if err := srv.keyValueService.PutString(&models.KeyStringValue{
		Key:   config.KeyLatestTotalTime,
		Value: total.String(),
	}); err != nil {
		logbuch.Error("failed to save total time count: %v", err)
	}

	if err := srv.keyValueService.PutString(&models.KeyStringValue{
		Key:   config.KeyLatestTotalUsers,
		Value: strconv.Itoa(i),
	}); err != nil {
		logbuch.Error("failed to save total users count: %v", err)
	}

	return nil
}

func (srv *MiscService) countTotalTimeWorker(jobs <-chan *CountTotalTimeJob, results chan<- *CountTotalTimeResult) {
	for job := range jobs {
		if result, err := srv.summaryService.Aliased(time.Time{}, time.Now(), &models.User{ID: job.UserID}, srv.summaryService.Retrieve, nil, false); err != nil {
			config.Log().Error("failed to count total for user %s: %v", job.UserID, err)
		} else {
			results <- &CountTotalTimeResult{
				UserId: job.UserID,
				Total:  result.TotalTime(),
			}
		}
	}
}
