package services

import (
	"github.com/duke-git/lancet/v2/datetime"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/leandro-lugaresi/hub"
	"github.com/muety/artifex/v2"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"log/slog"
	"math/rand"
	"time"
)

// delay between evey report generation task (to throttle email sending frequency)
const reportDelay = 10 * time.Second

// past time range to cover in the report
const reportRange = 7 * 24 * time.Hour

type ReportService struct {
	config         *config.Config
	eventBus       *hub.Hub
	summaryService ISummaryService
	userService    IUserService
	mailService    IMailService
	rand           *rand.Rand
	queueDefault   *artifex.Dispatcher
	queueWorkers   *artifex.Dispatcher
}

func NewReportService(summaryService ISummaryService, userService IUserService, mailService IMailService) *ReportService {
	srv := &ReportService{
		config:         config.Get(),
		eventBus:       config.EventBus(),
		summaryService: summaryService,
		userService:    userService,
		mailService:    mailService,
		rand:           rand.New(rand.NewSource(time.Now().Unix())),
		queueDefault:   config.GetDefaultQueue(),
		queueWorkers:   config.GetQueue(config.QueueReports),
	}

	return srv
}

func (srv *ReportService) Schedule() {
	slog.Info("scheduling report generation")

	scheduleUserReport := func(u *models.User) {
		if err := srv.queueWorkers.Dispatch(func() {
			t0 := time.Now()

			if err := srv.SendReport(u, reportRange); err != nil {
				config.Log().Error("failed to generate report for '%s', %v", u.ID, err)
			}

			// make the job take at least reportDelay seconds
			if diff := reportDelay - time.Now().Sub(t0); diff > 0 {
				slog.Debug("waiting for %v before sending next report", diff)
				time.Sleep(diff)
			}
		}); err != nil {
			config.Log().Error("failed to dispatch report generation job for user '%s', %v", u.ID, err)
		}
	}

	_, err := srv.queueDefault.DispatchCron(func() {
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

		// schedule jobs, throttled by one job per x seconds
		slog.Info("scheduling report generation for %d users", len(users))
		for _, u := range users {
			scheduleUserReport(u)
		}
	}, srv.config.App.GetWeeklyReportCron())

	if err != nil {
		config.Log().Error("failed to dispatch report generation jobs, %v", err)
	}
}

func (srv *ReportService) SendReport(user *models.User, duration time.Duration) error {
	if user.Email == "" {
		slog.Warn("not generating report for '%s' as no e-mail address is set")
		return nil
	}

	slog.Info("generating report for '%s'", user.ID)

	end := time.Now().In(user.TZ())
	start := time.Now().Add(-1 * duration)

	fullSummary, err := srv.summaryService.Aliased(start, end, user, srv.summaryService.Retrieve, nil, false)
	if err != nil {
		config.Log().Error("failed to generate report for '%s' - %v", user.ID, err)
		return err
	}

	// generate per-day summaries
	dayIntervals := utils.SplitRangeByDays(start, end)
	dailySummaries := make([]*models.Summary, len(dayIntervals))

	for i, interval := range dayIntervals {
		from, to := datetime.BeginOfDay(interval[0]), interval[1]
		summary, err := srv.summaryService.Aliased(from, to, user, srv.summaryService.Retrieve, nil, false)
		if err != nil {
			config.Log().Error("failed to generate day summary (%v to %v) for report for '%s' - %v", from, to, user.ID, err)
			break
		}
		summary.FromTime = models.CustomTime(from)
		summary.ToTime = models.CustomTime(to.Add(-1 * time.Second))
		dailySummaries[i] = summary
	}

	report := &models.Report{
		From:           start,
		To:             end,
		User:           user,
		Summary:        fullSummary,
		DailySummaries: dailySummaries,
	}

	if err := srv.mailService.SendReport(user, report); err != nil {
		config.Log().Error("failed to send report for '%s', %v", user.ID, err)
		return err
	}

	slog.Info("sent report to user '%s'", user.ID)
	return nil
}
