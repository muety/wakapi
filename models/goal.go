package models

import (
	"fmt"
	"strings"
	"time"
)

type Goal struct {
	ID               string           `json:"id" gorm:"primary_key"`
	UserID           string           `json:"user_id"`
	CreatedAt        CustomTime       `json:"created_at" gorm:"default:CURRENT_TIMESTAMP" swaggertype:"string" format:"date" example:"2006-01-02 15:04:05.000"`
	UpdatedAt        CustomTime       `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP" swaggertype:"string" format:"date" example:"2006-01-02 15:04:05.000"`
	SnoozeUntil      int64            `json:"snooze_until"`
	Seconds          int64            `json:"seconds"`
	ImproveByPercent int64            `json:"improve_by_percent"`
	Delta            string           `json:"delta" gorm:"size:1055"`
	Type             string           `json:"type" gorm:"size:1055"`
	Title            string           `json:"title" gorm:"size:1055"`
	CustomTitle      *string          `json:"custom_title" gorm:"size:1025"`
	CumulativeStatus string           `json:"cumulative_status" gorm:"size:1055"`
	Status           string           `json:"status" gorm:"size:1055"`
	IsSnoozed        bool             `json:"is_snoozed"`
	IsEnabled        bool             `json:"is_enabled"`
	Languages        []string         `json:"languages" gorm:"serializer:json"`
	Projects         []string         `json:"projects" gorm:"serializer:json"`
	Editors          []string         `json:"editors" gorm:"serializer:json"`
	Categories       []string         `json:"categories" gorm:"serializer:json"`
	ChartData        []*GoalChartData `json:"chart_data" gorm:"-"`
}

type GoalChartRange struct {
	Date     string    `json:"date"`
	End      time.Time `json:"end"`
	Start    time.Time `json:"start"`
	Text     string    `json:"text"`
	Timezone string    `json:"timezone"`
}

type GoalChartData struct {
	ActualSeconds          float64        `json:"actual_seconds"`
	ActualSecondsText      string         `json:"actual_seconds_text"`
	GoalSeconds            float64        `json:"goal_seconds"`
	GoalSecondsText        string         `json:"goal_seconds_text"`
	RangeStatus            string         `json:"range_status"`
	RangeStatusReason      string         `json:"range_status_reason"`
	RangeStatusReasonShort string         `json:"range_status_reason_short"`
	Range                  GoalChartRange `json:"range"`
}

func (r *GoalChartRange) ComputeText() {
	// Calculate the difference in days between start and end
	diff := int(r.End.Sub(r.Start).Hours() / 24)

	// Define a format function for consistent date formatting
	formatDate := func(t time.Time) string {
		return fmt.Sprintf("%s %d", t.Month().String()[:3], t.Day())
	}

	if diff == 0 {
		r.Text = formatDate(r.Start)
	} else {
		r.Text = fmt.Sprintf("%s - %s", formatDate(r.Start), formatDate(r.End))
	}
}

func (g *Goal) GetGoalSuffix(prefix string, items []string) string {
	if len(items) > 0 {
		return fmt.Sprintf("in %s %s", prefix, strings.Join(items, ", "))
	}
	return ""
}

func (g *Goal) GetTitle() string {
	hours := float64(g.Seconds) / float64(3600)
	suffixes := []string{
		g.GetGoalSuffix("languages", g.Languages),
		g.GetGoalSuffix("editors", g.Editors),
		g.GetGoalSuffix("categories", g.Categories),
		g.GetGoalSuffix("projects", g.Projects),
	}
	return fmt.Sprintf("Code %.2f hrs per %s %s", hours, g.Delta, strings.Join(suffixes, ""))
}

func (g *Goal) GetGoalSummaryFilter() *Filters {
	if g.Languages != nil && len(g.Languages) > 0 {
		return &Filters{
			Language: OrFilter(g.Languages),
		}
	}
	if g.Editors != nil && len(g.Editors) > 0 {
		return &Filters{
			Editor: OrFilter(g.Editors),
		}
	}
	if g.Projects != nil && len(g.Projects) > 0 {
		return &Filters{
			Project: OrFilter(g.Projects),
		}
	}
	if g.Categories != nil && len(g.Categories) > 0 {
		return &Filters{
			Category: OrFilter(g.Categories),
		}
	}
	return &Filters{}
}

type NewGoal struct {
	UserID         string   `json:"user_id"`
	Type           string   `json:"type"`
	Seconds        int64    `json:"seconds"`
	Delta          string   `json:"delta"`
	IgnoreDays     []string `json:"ignore_days"`
	IgnoreZeroDays bool     `json:"ignore_zero_days"`
	IsInverse      bool     `json:"is_inverse"`
	Languages      []string `json:"languages"`
	Projects       []string `json:"projects"`
	Editors        []string `json:"editors"`
	Categories     []string `json:"categories"`
	GoalCategory   string   `json:"goal_category"`
}
