package imports

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	routeutils "github.com/muety/wakapi/routes/utils"
)

type WakatimeImporter struct {
	apiKey      string
	forceLegacy bool
}

func NewWakatimeImporter(apiKey string, forceLegacy bool) *WakatimeImporter {
	return &WakatimeImporter{apiKey: apiKey, forceLegacy: forceLegacy}
}

func (w *WakatimeImporter) Import(user *models.User, minFrom time.Time, maxTo time.Time) (<-chan *models.Heartbeat, error) {
	if err := w.Validate(user); err != nil {
		return nil, err
	}
	if strings.Contains(user.WakaTimeURL(config.WakatimeApiUrl), "wakatime.com") && !w.forceLegacy {
		return NewWakatimeDumpImporter(w.apiKey).Import(user, minFrom, maxTo)
	}
	return NewWakatimeHeartbeatImporter(w.apiKey).Import(user, minFrom, maxTo)
}

func (w *WakatimeImporter) ImportAll(user *models.User) (<-chan *models.Heartbeat, error) {
	if err := w.Validate(user); err != nil {
		return nil, err
	}
	if strings.Contains(user.WakaTimeURL(config.WakatimeApiUrl), "wakatime.com") && !w.forceLegacy {
		return NewWakatimeDumpImporter(w.apiKey).ImportAll(user)
	}
	return NewWakatimeHeartbeatImporter(w.apiKey).ImportAll(user)
}

func (w *WakatimeImporter) Validate(user *models.User) error {
	return w.checkUrl(user)
}

func (w *WakatimeImporter) checkUrl(user *models.User) error {
	wakatimeUrl := user.WakaTimeURL(config.WakatimeApiUrl)

	wakatimeUrlParsed, err := url.Parse(wakatimeUrl)
	if err != nil {
		return err
	}

	if err := routeutils.ValidateWakatimeUrl(wakatimeUrl); err != nil {
		return err
	}

	if !config.Get().App.IsImportHostWhitelisted(wakatimeUrlParsed.Hostname()) {
		return fmt.Errorf("import from host '%s' is not allowed", wakatimeUrlParsed.Hostname())
	}
	return nil
}
