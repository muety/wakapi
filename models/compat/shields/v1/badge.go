package v1

import (
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"time"
)

// https://shields.io/endpoint

const (
	defaultLabel = "coding time"
	defaultColor = "#2D3748" // not working
)

type BadgeData struct {
	SchemaVersion int    `json:"schemaVersion"`
	Label         string `json:"label"`
	Message       string `json:"message"`
	Color         string `json:"color"`
}

func NewBadgeDataFrom(summary *models.Summary, filters *models.Filters) *BadgeData {
	var total time.Duration
	if hasFilter, _, _ := filters.First(); hasFilter {
		total = summary.TotalTimeByFilters(filters)
	} else {
		total = summary.TotalTime()
	}

	return &BadgeData{
		SchemaVersion: 1,
		Label:         defaultLabel,
		Message:       utils.FmtWakatimeDuration(total),
		Color:         defaultColor,
	}
}
