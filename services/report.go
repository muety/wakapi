package services

import (
	"github.com/emvi/logbuch"
	"github.com/go-co-op/gocron"
	"github.com/leandro-lugaresi/hub"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"math/rand"
	"sync"
	"time"
)

var reportLock = sync.Mutex{}

// range for random offset to add / subtract when scheduling a new job
// to avoid all mails being sent at once, but distributed over 2*offsetIntervalMin minutes
const offsetIntervalMin = 15

type ReportService struct {
	config         *config.Config
	eventBus       *hub.Hub
	summaryService ISummaryService
	userService    IUserService
	mailService    IMailService
	scheduler      *gocron.Scheduler
	rand           *rand.Rand
}

func NewReportService(summaryService ISummaryService, userService IUserService, mailService IMailService) *ReportService {
	srv := &ReportService{
		config:         config.Get(),
		eventBus:       config.EventBus(),
		summaryService: summaryService,
		userService:    userService,
		mailService:    mailService,
		scheduler:      gocron.NewScheduler(time.Local),
		rand:           rand.New(rand.NewSource(time.Now().Unix())),
	}

	srv.scheduler.StartAsync()

	sub := srv.eventBus.Subscribe(0, config.EventUserUpdate)
	go func(sub *hub.Subscription) {
		for m := range sub.Receiver {
			srv.SyncSchedule(m.Fields[config.FieldPayload].(*models.User))
		}
	}(&sub)

	return srv
}

func (srv *ReportService) Schedule() {
	logbuch.Info("initializing report service")

	users, err := srv.userService.GetAllByReports(true)
	if err != nil {
		config.Log().Fatal("%v", err)
	}

	logbuch.Info("scheduling reports for %d users", len(users))
	for _, u := range users {
		srv.SyncSchedule(u)
	}
}

// SyncSchedule syncs the currently active schedulers with the user's wish about whether or not to receive reports.
// Returns whether a scheduler is active after this operation has run.
func (srv *ReportService) SyncSchedule(u *models.User) bool {
	reportLock.Lock()
	defer reportLock.Unlock()

	// unschedule
	if !u.ReportsWeekly {
		_ = srv.scheduler.RemoveByTag(u.ID)
		return false
	}

	// schedule
	if j := srv.getJobByTag(u.ID); j == nil && u.ReportsWeekly {
		t, _ := time.ParseInLocation("15:04", srv.config.App.GetWeeklyReportTime(), u.TZ())
		t = t.Add(time.Duration(srv.rand.Intn(offsetIntervalMin)*srv.rand.Intn(2)) * time.Minute)
		if _, err := srv.scheduler.
			Every(1).
			Week().
			Weekday(srv.config.App.GetWeeklyReportDay()).
			At(t).
			Tag(u.ID).
			Do(srv.Run, u, 7*24*time.Hour); err != nil {
			config.Log().Error("failed to schedule report job for user '%s' – %v", u.ID, err)
		}
	}

	return u.ReportsWeekly
}

func (srv *ReportService) Run(user *models.User, duration time.Duration) error {
	if user.Email == "" {
		logbuch.Warn("not generating report for '%s' as no e-mail address is set")
	}

	if !srv.SyncSchedule(user) {
		logbuch.Info("reports for user '%s' were turned off in the meanwhile since last report job ran")
		return nil
	}

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

func (srv *ReportService) getJobByTag(tag string) *gocron.Job {
	for _, j := range srv.scheduler.Jobs() {
		for _, t := range j.Tags() {
			if t == tag {
				return j
			}
		}
	}
	return nil
}
