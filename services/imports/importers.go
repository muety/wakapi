package imports

import (
	"github.com/muety/wakapi/models"
	"time"
)

type DataImporter interface {
	Import(*models.User, time.Time, time.Time) (<-chan *models.Heartbeat, error)
	ImportAll(*models.User) (<-chan *models.Heartbeat, error)
}
