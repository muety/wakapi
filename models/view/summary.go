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
	// Timeout is the heartbeat gap threshold used to decide whether two
	// adjacent segments belong to the same activity block during coalescing.
	// It mirrors the per-user models.Duration.Timeout value.
	Timeout time.Duration `json:"-"`
}

// TimeEnd returns the exclusive end time of this item.
func (h *HourlyBreakdownItem) TimeEnd() time.Time {
	return h.FromTime.Add(h.Duration)
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
				Timeout:  duration.Timeout,
			}
		})

	hourlyBreakdowns = slice.Map(hourlyBreakdowns, func(_ int, item *HourlyBreakdownItem) *HourlyBreakdownItem {
		item.Project = resolve(models.SummaryProject, item.Project)
		item.Entity = resolve(models.SummaryEntity, item.Entity)
		return item
	})

	return hourlyBreakdowns
}

// coalesceHourlyBreakdownItems merges chronologically adjacent items whose gap
// is no larger than the user's heartbeat timeout. The input slice must already
// be sorted ascending by FromTime (as guaranteed by NewHourlyBreakdownViewModel).
// Coalescing reduces the number of chart bars and prevents the chart from
// becoming unreadably fragmented on AI-assisted or fast-switching sessions.
func coalesceHourlyBreakdownItems(items HourlyBreakdownItems) HourlyBreakdownItems {
	if len(items) == 0 {
		return items
	}

	merged := make(HourlyBreakdownItems, 0, len(items))
	current := &HourlyBreakdownItem{
		FromTime: items[0].FromTime,
		Duration: items[0].Duration,
		Entity:   items[0].Entity,
		Project:  items[0].Project,
		Timeout:  items[0].Timeout,
	}

	for _, next := range items[1:] {
		gap := next.FromTime.Sub(current.TimeEnd())
		if gap <= current.Timeout {
			// extend the current block to absorb this segment
			current.Duration = next.TimeEnd().Sub(current.FromTime)
			// keep the timeout of the earliest segment in the block
		} else {
			merged = append(merged, current)
			current = &HourlyBreakdownItem{
				FromTime: next.FromTime,
				Duration: next.Duration,
				Entity:   next.Entity,
				Project:  next.Project,
				Timeout:  next.Timeout,
			}
		}
	}
	merged = append(merged, current)

	return merged
}

func NewHourlyBreakdownViewModel(items HourlyBreakdownItems) HourlyBreakdownsViewModel {
	hourlyBreakdownMap := slice.GroupWith(items, func(item *HourlyBreakdownItem) string { return item.Project })

	hourlyBreakdown := make([]*HourlyBreakdownViewModel, 0)
	for project, projectItems := range hourlyBreakdownMap {
		// sort ascending by start time before coalescing
		slice.SortBy(projectItems, func(i, j *HourlyBreakdownItem) bool {
			return i.FromTime.Before(j.FromTime)
		})
		// coalesce contiguous segments so the chart stays readable regardless
		// of how fragmented the raw duration data is (fixes #952)
		coalesced := coalesceHourlyBreakdownItems(projectItems)
		hourlyBreakdown = append(hourlyBreakdown, &HourlyBreakdownViewModel{
			Items:   coalesced,
			Project: project,
		})
	}

	return hourlyBreakdown
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
