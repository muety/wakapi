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
	AvailableFilters    AvailableFilters
	AvatarURL           string
	EditorColors        map[string]string
	LanguageColors      map[string]string
	OSColors            map[string]string
	Timeline            []*TimelineViewModel
	HourlyBreakdown     []*HourlyBreakdownViewModel
	HourlyBreakdownFrom time.Time
	RawQuery            string
	UserFirstData       time.Time
	DataRetentionMonths int
}

type AvailableFilters struct {
	ProjectNames  []string
	LanguageNames []string
	MachineNames  []string
	LabelNames    []string
	CategoryNames []string
}

type TimelineViewModel struct {
	Date     time.Time       `json:"date"`
	Projects []*TimelineItem `json:"projects"`
}

type TimelineItem struct {
	Name     string        `json:"name"`
	Duration time.Duration `json:"duration"`
}

type HourlyBreakdownsViewModel []*HourlyBreakdownViewModel

type HourlyBreakdownViewModel struct {
	Items   []*HourlyBreakdownItem `json:"items"`
	Project string                 `json:"project"`
}

type HourlyBreakdownItems []*HourlyBreakdownItem

type HourlyBreakdownItem struct {
	FromTime time.Time     `json:"from_time"`
	Duration time.Duration `json:"duration"`
	Entity   string        `json:"entity"`
	Project  string        `json:"-"`
}

func NewTimelineViewModel(summaries []*models.Summary) []*TimelineViewModel {
	vm := make([]*TimelineViewModel, 0)
	for _, summary := range summaries {
		vm = append(vm, &TimelineViewModel{
			Date: summary.FromTime.T(),
			Projects: slice.Map(summary.Projects, func(_ int, curProject *models.SummaryItem) *TimelineItem {
				return &TimelineItem{
					Name:     curProject.Key,
					Duration: curProject.Total,
				}
			}),
		})
	}
	return vm
}

func NewHourlyBreakdownItems(durations models.Durations, resolve models.AliasResolver) HourlyBreakdownItems {
	hourlyBreakdowns := slice.
		Map(durations, func(_ int, duration *models.Duration) *HourlyBreakdownItem {
			return &HourlyBreakdownItem{
				FromTime: duration.Time.T(),
				Duration: duration.Duration,
				Entity:   duration.Entity,
				Project:  duration.Project,
			}
		})

	hourlyBreakdowns = slice.Map(hourlyBreakdowns, func(_ int, item *HourlyBreakdownItem) *HourlyBreakdownItem {
		item.Project = resolve(models.SummaryProject, item.Project)
		item.Entity = resolve(models.SummaryEntity, item.Entity)
		return item
	})

	return hourlyBreakdowns
}

func NewHourlyBreakdownViewModel(items HourlyBreakdownItems) HourlyBreakdownsViewModel {
	hourlyBreakdownMap := slice.GroupWith(items, func(item *HourlyBreakdownItem) string { return item.Project })

	hourlyBreakdown := make([]*HourlyBreakdownViewModel, 0)
	for project, items := range hourlyBreakdownMap {
		hourlyBreakdown = append(hourlyBreakdown, &HourlyBreakdownViewModel{
			Items:   items,
			Project: project,
		})
	}

	hourlyBreakdownSorted := slice.Map(hourlyBreakdown, func(_ int, item *HourlyBreakdownViewModel) *HourlyBreakdownViewModel {
		slice.SortBy(item.Items, func(i, j *HourlyBreakdownItem) bool {
			return i.FromTime.Before(j.FromTime)
		})
		return item
	})

	return hourlyBreakdownSorted
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
