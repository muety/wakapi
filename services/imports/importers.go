package imports

import (
	"github.com/muety/wakapi/models"
	"time"
)

type HeartbeatImporter interface {
	Import(*models.User, time.Time, time.Time) <-chan *models.Heartbeat
	ImportAll(*models.User) <-chan *models.Heartbeat
}
