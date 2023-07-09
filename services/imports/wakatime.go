package imports

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"strings"
	"time"
)

type WakatimeImporter struct {
	apiKey string
}

func NewWakatimeImporter(apiKey string) *WakatimeImporter {
	return &WakatimeImporter{apiKey: apiKey}
}

func (w *WakatimeImporter) Import(user *models.User, minFrom time.Time, maxTo time.Time) (<-chan *models.Heartbeat, error) {
	if strings.Contains(user.WakaTimeURL(config.WakatimeApiUrl), "wakatime.com") {
		return NewWakatimeDumpImporter(w.apiKey).Import(user, minFrom, maxTo)
	}
	return NewWakatimeHeartbeatImporter(w.apiKey).Import(user, minFrom, maxTo)
}

func (w *WakatimeImporter) ImportAll(user *models.User) (<-chan *models.Heartbeat, error) {
	if strings.Contains(user.WakaTimeURL(config.WakatimeApiUrl), "wakatime.com") {
		return NewWakatimeDumpImporter(w.apiKey).ImportAll(user)
	}
	return NewWakatimeHeartbeatImporter(w.apiKey).ImportAll(user)
}
