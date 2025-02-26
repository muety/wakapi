package types

import (
	"github.com/muety/wakapi/models"
	"time"
)

type SummaryRetriever func(f, t time.Time, u *models.User, filters *models.Filters, duration *time.Duration) (*models.Summary, error)
