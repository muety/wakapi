package services

import (
	"log/slog"
	"math/rand"
	"time"

	"github.com/duke-git/lancet/v2/datetime"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/leandro-lugaresi/hub"
	"github.com/muety/artifex/v2"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/internal/mail"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"gorm.io/gorm"
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
	mailService    mail.IMailService
	rand           *rand.Rand
	queueDefault   *artifex.Dispatcher
	queueWorkers   *artifex.Dispatcher
	db             *gorm.DB
}

// a special type used to perform business logic that ensures a report is not sent multiple times
// for the same user on the same day. This is non-invasive
type ReportDeduplicator struct {
	db         *gorm.DB
	user       *models.User
	reportDate time.Time
}

func ReportSentTracker(db *gorm.DB, user *models.User) *ReportDeduplicator {
	return &ReportDeduplicator{
		db:         db,
		user:       user,
		reportDate: time.Now().Truncate(24 * time.Hour),
	}
}

func (rst *ReportDeduplicator) IsReportSent() bool {
	var existingReportSent models.UserReportSent
	err := rst.db.Where("user_id = ? AND report_date = ?", rst.user.ID, rst.reportDate).First(&existingReportSent).Error
	return err == nil
}

func (rst *ReportDeduplicator) MarkReportAsSent() error {
	reportSent := &models.UserReportSent{
		UserID:     rst.user.ID,
		ReportDate: rst.reportDate, // Use server's local date
		SentAt:     time.Now(),
	}
	if err := rst.db.Create(reportSent).Error; err != nil {
		slog.Warn("failed to save report sent record (possible race condition or duplicate)",
			"userID", reportSent.UserID,
			"reportDate", reportSent.ReportDate,
			"error", err)
		return nil
	}
	slog.Debug("saved report sent record", "userID", reportSent.UserID, "reportDate", reportSent.ReportDate)
	return nil
}

func NewReportService(db *gorm.DB) *ReportService {
	summaryService := NewSummaryService(db)
	userService := NewUserService(db)

	srv := &ReportService{
		config:         config.Get(),
		eventBus:       config.EventBus(),
		summaryService: summaryService,
		userService:    userService,
		mailService:    mail.NewMailService(),
		rand:           rand.New(rand.NewSource(time.Now().Unix())),
		queueDefault:   config.GetDefaultQueue(),
		queueWorkers:   config.GetQueue(config.QueueReports),
		db:             db,
	}

	return srv
}

func (srv *ReportService) Schedule() {
	slog.Info("scheduling report generation")

	scheduleUserReport := func(u *models.User) {
		if err := srv.queueWorkers.Dispatch(func() {
			t0 := time.Now()

			if err := srv.SendReport(u, reportRange); err != nil {
				config.Log().Error("failed to generate report", "userID", u.ID, "error", err)
			}

			// make the job take at least reportDelay seconds
			if diff := reportDelay - time.Since(t0); diff > 0 {
				slog.Debug("waiting before sending next report", "duration", diff)
				time.Sleep(diff)
			}
		}); err != nil {
			config.Log().Error("failed to dispatch report generation job for user", "userID", u.ID, "error", err)
		}
	}

	_, err := srv.queueDefault.DispatchCron(func() {
		// fetch all users with reports enabled
		users, err := srv.userService.GetAllByReports(true)
		if err != nil {
			config.Log().Error("failed to get users for report generation", "error", err)
			return
		}

		// filter users who have their email set
		users = slice.Filter(users, func(i int, u *models.User) bool {
			return u.Email != ""
		})

		// schedule jobs, throttled by one job per x seconds
		slog.Info("scheduling report generation", "userCount", len(users))
		for _, u := range users {
			scheduleUserReport(u)
		}
	}, "0 0 6 * * 1")

	if err != nil {
		config.Log().Error("failed to dispatch report generation jobs", "error", err)
	}
}

func (srv *ReportService) SendReport(user *models.User, duration time.Duration) error {
	if user.Email == "" {
		slog.Warn("not generating report as no e-mail address is set", "userID", user.ID)
		return nil
	}

	tracker := ReportSentTracker(srv.db, user)
	if tracker.IsReportSent() {
		slog.Debug("report already sent for today, skipping", "userID", user.ID)
		return nil
	}

	slog.Info("generating report for user", "userID", user.ID)
	end := datetime.EndOfDay(time.Now().Add(-24 * time.Hour).In(user.TZ()))
	start := end.Add(-1 * duration).Add(1 * time.Second)

	fullSummary, err := srv.summaryService.Aliased(start, end, user, srv.summaryService.Retrieve, nil, false)
	if err != nil {
		config.Log().Error("failed to generate report", "userID", user.ID, "error", err)
		return err
	}

	// generate per-day summaries
	dayIntervals := utils.SplitRangeByDays(start, end)
	dailySummaries := make([]*models.Summary, len(dayIntervals))

	for i, interval := range dayIntervals {
		from, to := datetime.BeginOfDay(interval[0]), interval[1]
		summary, err := srv.summaryService.Aliased(from, to, user, srv.summaryService.Retrieve, nil, false)
		if err != nil {
			config.Log().Error("failed to generate day summary for report", "from", from, "to", to, "userID", user.ID, "error", err)
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
		config.Log().Error("failed to send report", "userID", user.ID, "error", err)
		return err
	}

	slog.Info("sent report to user", "userID", user.ID)
	_ = tracker.MarkReportAsSent()
	return nil
}
