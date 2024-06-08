package models

import (
	"fmt"
	"strings"
)

type StringArray []string

type Goal struct {
	ID               string      `json:"id" gorm:"primary_key"`
	UserID           string      `json:"user_id"`
	CreatedAt        CustomTime  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP" swaggertype:"string" format:"date" example:"2006-01-02 15:04:05.000"`
	UpdatedAt        CustomTime  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP" swaggertype:"string" format:"date" example:"2006-01-02 15:04:05.000"`
	SnoozeUntil      int64       `json:"snooze_until"`
	Seconds          int64       `json:"seconds"`
	ImproveByPercent int64       `json:"improve_by_percent"`
	Delta            string      `json:"delta" gorm:"size:1055"`
	Type             string      `json:"type" gorm:"size:1055"`
	Title            string      `json:"title" gorm:"size:1055"`
	CustomTitle      *string     `json:"custom_title" gorm:"size:1025"`
	CumulativeStatus string      `json:"cumulative_status" gorm:"size:1055"`
	Status           string      `json:"status" gorm:"size:1055"`
	IsSnoozed        bool        `json:"is_snoozed"`
	IsEnabled        bool        `json:"is_enabled"`
	Languages        StringArray `json:"languages" gorm:"serializer:json"`
	Projects         StringArray `json:"projects" gorm:"serializer:json"`
	Editors          StringArray `json:"editors" gorm:"serializer:json"`
	Categories       StringArray `json:"categories" gorm:"serializer:json"`
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
	}
	return fmt.Sprintf("Code %.2f hrs per %s %s", hours, g.Delta, strings.Join(suffixes, ""))
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
