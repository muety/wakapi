package services

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/emvi/logbuch"
	"github.com/leandro-lugaresi/hub"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"math/rand"
	"time"
)

// delay between evey report generation task (to throttle email sending frequency)
const reportDelay = 15 * time.Second

// past time range to cover in the report
const reportRange = 7 * 24 * time.Hour

type ReportService struct {
	config         *config.Config
	eventBus       *hub.Hub
	summaryService ISummaryService
	userService    IUserService
	mailService    IMailService
	rand           *rand.Rand
}

func NewReportService(summaryService ISummaryService, userService IUserService, mailService IMailService) *ReportService {
	srv := &ReportService{
		config:         config.Get(),
		eventBus:       config.EventBus(),
		summaryService: summaryService,
		userService:    userService,
		mailService:    mailService,
		rand:           rand.New(rand.NewSource(time.Now().Unix())),
	}

	return srv
}

func (srv *ReportService) Schedule() {
	logbuch.Info("initializing report service")

	_, err := config.GetDefaultQueue().DispatchCron(func() {
		// fetch all users with reports enabled
		users, err := srv.userService.GetAllByReports(true)
		if err != nil {
			config.Log().Error("failed to get users for report generation, %v", err)
			return
		}

		// filter users who have their email set
		users = slice.Filter[*models.User](users, func(i int, u *models.User) bool {
			return u.Email != ""
		})

		// schedule jobs, throttled by one job per 15 seconds
		logbuch.Info("scheduling report generation for %d users", len(users))
		for i, u := range users {
			err := config.GetQueue(config.QueueMails).DispatchIn(func() {
				if err := srv.SendReport(u, reportRange); err != nil {
					config.Log().Error("failed to generate report for '%s', %v", u.ID, err)
				}
			}, time.Duration(i)*reportDelay)

			if err != nil {
				config.Log().Error("failed to dispatch report generation job for user '%s', %v", u.ID, err)
			}
		}
	}, srv.config.App.GetWeeklyReportCron())

	if err != nil {
		config.Log().Error("failed to dispatch report generation jobs, %v", err)
	}
}

func (srv *ReportService) SendReport(user *models.User, duration time.Duration) error {
	if user.Email == "" {
		logbuch.Warn("not generating report for '%s' as no e-mail address is set")
		return nil
	}

	logbuch.Info("generating report for '%s'", user.ID)

	end := time.Now().In(user.TZ())
	start := time.Now().Add(-1 * duration)

	summary, err := srv.summaryService.Aliased(start, end, user, srv.summaryService.Retrieve, nil, false)
	if err != nil {
		config.Log().Error("failed to generate report for '%s' - %v", user.ID, err)
		return err
	}

	report := &models.Report{
		From:    start,
		To:      end,
		User:    user,
		Summary: summary,
	}

	if err := srv.mailService.SendReport(user, report); err != nil {
		config.Log().Error("failed to send report for '%s', %v", user.ID, err)
		return err
	}

	logbuch.Info("sent report to user '%s'", user.ID)
	return nil
}
