package view

import (
	"time"

	"github.com/duke-git/lancet/v2/slice"
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
	DailyStats          []*DailyProjectsViewModel
	RawQuery            string
	UserFirstData       time.Time
	DataRetentionMonths int
}

type DailyProjectsViewModel struct {
	Date     time.Time                `json:"date"`
	Projects []*DailyProjectViewModel `json:"projects"`
}

type DailyProjectViewModel struct {
	Name     string        `json:"name"`
	Duration time.Duration `json:"duration"`
}

func NewDailyProjectStats(summaries []*models.Summary) []*DailyProjectsViewModel {
	dailyProjects := make([]*DailyProjectsViewModel, 0)
	for _, summary := range summaries {
		dailyProjects = append(dailyProjects, &DailyProjectsViewModel{
			Date: summary.FromTime.T(),
			Projects: slice.Map(summary.Projects, func(_ int, curProject *models.SummaryItem) *DailyProjectViewModel {
				return &DailyProjectViewModel{
					Name:     curProject.Key,
					Duration: curProject.Total,
				}
			}),
		})
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
