package imports

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/muety/wakapi/utils"
	"net/http"
	"time"

	"github.com/emvi/logbuch"
	"github.com/muety/artifex/v2"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	wakatime "github.com/muety/wakapi/models/compat/wakatime/v1"
)

// data example: https://github.com/muety/wakapi/issues/323#issuecomment-1627467052

type WakatimeDumpImporter struct {
	apiKey     string
	httpClient *http.Client
	queue      *artifex.Dispatcher
}

func NewWakatimeDumpImporter(apiKey string) *WakatimeDumpImporter {
	return &WakatimeDumpImporter{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		queue:      config.GetQueue(config.QueueImports),
	}
}

func (w *WakatimeDumpImporter) Import(user *models.User, minFrom time.Time, maxTo time.Time) (<-chan *models.Heartbeat, error) {
	out := make(chan *models.Heartbeat)
	logbuch.Info("running wakatime dump import for user '%s'", user.ID)

	url := config.WakatimeApiUrl + config.WakatimeApiDataDumpUrl // this importer only works with wakatime currently, so no point in using user's custom wakatime api url
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte(`{ "type": "heartbeats", "email_when_finished": false }`)))
	res, err := utils.RaiseForStatus(w.httpClient.Do(w.withHeaders(req)))

	if err != nil && res != nil && res.StatusCode == http.StatusBadRequest {
		var datadumpError wakatime.DataDumpResultErrorModel
		if err := json.NewDecoder(res.Body).Decode(&datadumpError); err != nil {
			return nil, err
		}
		// in case of this error message, a dump had already been requested before and can simply be downloaded now
		// -> just keep going as usual (kick off poll loop), otherwise yield error
		if datadumpError.Error == "Wait for your current export to expire before creating another." {
			logbuch.Info("failed to request new dump, because other non-expired dump already existing, using that one")
		} else {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var readyPollTimer *artifex.DispatchTicker

	// callbacks
	checkDumpAvailable := func(user *models.User) (bool, *wakatime.DataDumpData, error) {
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		res, err := utils.RaiseForStatus(w.httpClient.Do(w.withHeaders(req)))
		if err != nil {
			return false, nil, err
		}

		var datadumpData wakatime.DataDumpViewModel
		if err := json.NewDecoder(res.Body).Decode(&datadumpData); err != nil {
			return false, nil, err
		}

		if len(datadumpData.Data) < 1 {
			return false, nil, errors.New("no dumps available")
		}

		return datadumpData.Data[0].Status == "Completed", datadumpData.Data[0], nil
	}

	onDumpFailed := func(err error, user *models.User) {
		config.Log().Error("fetching data dump for user '%s' failed - %v", user.ID, err)
		readyPollTimer.Stop()
		close(out)
	}

	onDumpReady := func(dump *wakatime.DataDumpData, user *models.User, out chan *models.Heartbeat) {
		config.Log().Info("data dump for user '%s' is available for download", user.ID)
		readyPollTimer.Stop()
		defer close(out)

		// download
		req, _ := http.NewRequest(http.MethodGet, dump.DownloadUrl, nil)
		res, err := utils.RaiseForStatus((&http.Client{Timeout: 5 * time.Minute}).Do(req))
		if err != nil {
			config.Log().Error("failed to download %s - %v", dump.DownloadUrl, err)
			return
		}
		defer res.Body.Close()

		logbuch.Info("fetched %d bytes data dump for user '%s'", res.ContentLength, user.ID)

		// decode
		var data wakatime.JsonExportViewModel
		if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
			config.Log().Error("failed to decode data dump for user '%s' ('%s') - %v", user.ID, dump.DownloadUrl, err)
			return
		}

		// fetch user agents and machine names
		var userAgents map[string]*wakatime.UserAgentEntry
		if userAgents, err = fetchUserAgents(config.WakatimeApiUrl, w.apiKey); err != nil {
			config.Log().Error("failed to fetch user agents while importing wakatime heartbeats for user '%s' - %v", user.ID, err)
			return
		}
		var machinesNames map[string]*wakatime.MachineEntry
		if machinesNames, err = fetchMachineNames(config.WakatimeApiUrl, w.apiKey); err != nil {
			config.Log().Error("failed to fetch machine names while importing wakatime heartbeats for user '%s' - %v", user.ID, err)
			return
		}

		// stream
		for _, d := range data.Days {
			for _, h := range d.Heartbeats {
				hb := mapHeartbeat(h, userAgents, machinesNames, user)
				if hb.Time.T().Before(minFrom) || hb.Time.T().After(maxTo) {
					continue
				}
				out <- hb
			}
		}
	}

	// start polling for dump to be ready
	readyPollTimer, err = w.queue.DispatchEvery(func() {
		u := *user
		ok, dump, err := checkDumpAvailable(&u)
		if err != nil {
			onDumpFailed(err, &u)
		} else if ok {
			logbuch.Info("waiting for data dump '%s' for user '%s' to become downloadable (%.2f percent complete)", dump.Id, u.ID, dump.PercentComplete)
			onDumpReady(dump, &u, out)
		}
	}, 10*time.Second)

	return out, nil
}

func (w *WakatimeDumpImporter) ImportAll(user *models.User) (<-chan *models.Heartbeat, error) {
	return w.Import(user, time.Time{}, time.Now())
}

func (w *WakatimeDumpImporter) withHeaders(req *http.Request) *http.Request {
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(w.apiKey))))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return req
}
