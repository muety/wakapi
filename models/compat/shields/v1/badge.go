package v1

import (
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
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

func NewBadgeDataFrom(summary *models.Summary) *BadgeData {
	return &BadgeData{
		SchemaVersion: 1,
		Label:         defaultLabel,
		Message:       utils.FmtWakatimeDuration(summary.TotalTime()),
		Color:         defaultColor,
	}
}
