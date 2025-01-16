package view

import (
	"time"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
)

type SummaryViewModel struct {
	SharedLoggedInViewModel
	*models.Summary
	*models.SummaryParams
	AvatarURL           string
	EditorColors        map[string]string
	LanguageColors      map[string]string
	OSColors            map[string]string
	DailyStats          []*DailyProjectViewModel
	RawQuery            string
	UserFirstData       time.Time
	DataRetentionMonths int
}

type DailyProjectViewModel struct {
	Date     time.Time     `json:"date"`
	Project  string        `json:"project"`
	Duration time.Duration `json:"duration"`
}

func NewDailyProjectStats(summaries []*models.Summary) []*DailyProjectViewModel {
	dailyProjects := make([]*DailyProjectViewModel, 0)
	for _, summary := range summaries {
		for _, project := range summary.Projects {
			dailyProjects = append(dailyProjects, &DailyProjectViewModel{
				Date:     summary.FromTime.T(),
				Project:  project.Key,
				Duration: project.Total,
			})
		}
	}
	return dailyProjects
}

func (s SummaryViewModel) UserDataExpiring() bool {
	cfg := conf.Get()
	return cfg.Subscriptions.Enabled &&
		cfg.App.DataRetentionMonths > 0 &&
		!s.UserFirstData.IsZero() &&
		!s.SharedLoggedInViewModel.User.HasActiveSubscription() &&
		time.Now().AddDate(0, -cfg.App.DataRetentionMonths, 0).After(s.UserFirstData)
}

func (s *SummaryViewModel) WithSuccess(m string) *SummaryViewModel {
	s.SetSuccess(m)
	return s
}

func (s *SummaryViewModel) WithError(m string) *SummaryViewModel {
	s.SetError(m)
	return s
}
