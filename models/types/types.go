package types

import (
	"github.com/muety/wakapi/models"
	"time"
)

type SummaryRetriever func(f, t time.Time, u *models.User, filters *models.Filters, duration *time.Duration) (*models.Summary, error)

type DurationsRegenerationMode uint8

const (
	DurationsRegenerationNone DurationsRegenerationMode = iota
	DurationsRegenerationIncremental
	DurationsRegenerationFull
)
