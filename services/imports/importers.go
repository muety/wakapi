package imports

import "github.com/muety/wakapi/models"

type HeartbeatImporter interface {
	Import(*models.User) <-chan *models.Heartbeat
}
