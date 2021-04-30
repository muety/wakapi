package services

import (
	"github.com/emvi/logbuch"
	"github.com/go-co-op/gocron"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"sync"
	"time"
)

var reportLock = sync.Mutex{}

type ReportService struct {
	config           *config.Config
	summaryService   ISummaryService
	userService      IUserService
	mailService      IMailService
	schedulersWeekly map[string]*gocron.Scheduler // user id -> scheduler
}

func NewReportService(summaryService ISummaryService, userService IUserService, mailService IMailService) *ReportService {
	return &ReportService{
		config:           config.Get(),
		summaryService:   summaryService,
		userService:      userService,
		mailService:      mailService,
		schedulersWeekly: map[string]*gocron.Scheduler{},
	}
}

func (srv *ReportService) Schedule() {
	logbuch.Info("initializing report service")

	users, err := srv.userService.GetAllByReports(true)
	if err != nil {
		config.Log().Fatal("%v", err)
	}

	logbuch.Info("scheduling reports for %d users", len(users))
	for _, u := range users {
		srv.UpdateUserSchedule(u)
	}
}

func (srv *ReportService) UpdateUserSchedule(u *models.User) {
	reportLock.Lock()
	defer reportLock.Unlock()

	// unschedule
	if s, ok := srv.schedulersWeekly[u.ID]; ok && !u.ReportsWeekly {
		s.Stop()
		s.Clear()
		delete(srv.schedulersWeekly, u.ID)
		return
	}

	// schedule
	if _, ok := srv.schedulersWeekly[u.ID]; !ok && u.ReportsWeekly {
		s := gocron.NewScheduler(u.TZ())
		s.
			Every(1).
			Week().
			Weekday(srv.config.App.GetWeeklyReportDay()).
			At(srv.config.App.GetWeeklyReportTime()).
			Do(srv.Run, u, 7*24*time.Hour)
		s.StartAsync()
		srv.schedulersWeekly[u.ID] = s
	}
}

func (srv *ReportService) Run(user *models.User, duration time.Duration) error {
	end := time.Now().In(user.TZ())
	start := time.Now().Add(-1 * duration)

	summary, err := srv.summaryService.Aliased(start, end, user, srv.summaryService.Retrieve, false)
	if err != nil {
		config.Log().Error("failed to generate report for '%s' – %v", user.ID, err)
		return err
	}

	report := &models.Report{
		From:    start,
		To:      end,
		User:    user,
		Summary: summary,
	}

	if err := srv.mailService.SendReport(user, report); err != nil {
		config.Log().Error("failed to send report for '%s' – %v", user.ID, err)
		return err
	}

	logbuch.Info("sent report to user '%s'", user.ID)
	return nil
}
