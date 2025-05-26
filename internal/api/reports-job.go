package api

import (
	"context"

	"github.com/riverqueue/river"
)

type WeeklyReportArgs struct{}

func (WeeklyReportArgs) Kind() string { return "weekly_reports" }

func (a *APIv1) weeklyReportWorker(_ context.Context, _ *river.Job[WeeklyReportArgs]) error {
	err := a.services.Report().SendWeeklyReports()
	if err != nil {
		return err
	}
	return nil
}
