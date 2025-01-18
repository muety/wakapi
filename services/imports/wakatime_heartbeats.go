package imports

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alitto/pond/v2"
	"github.com/duke-git/lancet/v2/datetime"
	"github.com/muety/artifex/v2"
	"github.com/muety/wakapi/utils"
	"net/http"
	"strings"
	"time"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	wakatime "github.com/muety/wakapi/models/compat/wakatime/v1"
	"go.uber.org/atomic"
	"log/slog"
)

const OriginWakatime = "wakatime"
const (
	// wakatime api permits a max. rate of 10 req / sec
	// https://github.com/wakatime/wakatime/issues/261
	// with 5 workers, each sleeping slightly over 1/2 sec after every req., we should stay well below that limit
	maxWorkers    = 5
	throttleDelay = 550 * time.Millisecond
)

type WakatimeHeartbeatsImporter struct {
	apiKey     string
	httpClient *http.Client
	queue      *artifex.Dispatcher
}

func NewWakatimeHeartbeatImporter(apiKey string) *WakatimeHeartbeatsImporter {
	return &WakatimeHeartbeatsImporter{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		queue:      config.GetQueue(config.QueueImports),
	}
}

func (w *WakatimeHeartbeatsImporter) Import(user *models.User, minFrom time.Time, maxTo time.Time) (<-chan *models.Heartbeat, error) {
	out := make(chan *models.Heartbeat)

	process := func(user *models.User, minFrom time.Time, maxTo time.Time, out chan *models.Heartbeat) {
		slog.Info("running wakatime import for user", "userID", user.ID)

		baseUrl := user.WakaTimeURL(config.WakatimeApiUrl)

		startDate, endDate, err := w.fetchRange(baseUrl)
		if err != nil {
			config.Log().Error("failed to fetch date range while importing wakatime heartbeats", "userID", user.ID, "error", err)
			return
		}

		if startDate.Before(minFrom) {
			startDate = minFrom
		}
		if endDate.After(maxTo) {
			endDate = maxTo
		}

		userAgents := map[string]*wakatime.UserAgentEntry{}
		if data, err := fetchUserAgents(baseUrl, w.apiKey); err == nil {
			userAgents = data
		} else if strings.Contains(baseUrl, "wakatime.com") {
			// when importing from wakatime, resolving user agents is mandatorily required
			config.Log().Error("failed to fetch user agents while importing wakatime heartbeats", "userID", user.ID, "error", err)
			return
		}

		machinesNames := map[string]*wakatime.MachineEntry{}
		if data, err := fetchMachineNames(baseUrl, w.apiKey); err == nil {
			machinesNames = data
		} else if strings.Contains(baseUrl, "wakatime.com") {
			// when importing from wakatime, resolving machine names is mandatorily required
			config.Log().Error("failed to fetch machine names while importing wakatime heartbeats", "userID", user.ID, "error", err)
			return
		}

		days := generateDays(startDate, endDate)

		c := atomic.NewUint32(uint32(len(days)))
		wp := pond.NewPool(maxWorkers)

		for _, d := range days {
			d := d // https://github.com/golang/go/wiki/CommonMistakes#using-reference-to-loop-iterator-variable

			wp.Submit(func() {
				defer time.Sleep(throttleDelay)

				d := d.Format(config.SimpleDateFormat)
				heartbeats, err := w.fetchHeartbeats(d, baseUrl)
				if err != nil {
					config.Log().Error("failed to fetch heartbeats for day and user", "day", d, "userID", user.ID, "error", err)
				}

				for _, h := range heartbeats {
					hb := mapHeartbeat(h, userAgents, machinesNames, user)
					if hb.Time.T().Before(minFrom) || hb.Time.T().After(maxTo) {
						continue
					}
					out <- hb
				}

				if c.Dec() == 0 {
					close(out)
				}
			})
		}

		wp.StopAndWait()
	}

	if minDataAge := user.MinDataAge(); minFrom.Before(minDataAge) {
		slog.Info("wakatime data import for user capped", "userID", user.ID, "cappedTo", fmt.Sprintf("[%v, %v]", minDataAge, maxTo))
	}

	slog.Info("scheduling wakatime import for user", "userID", user.ID, "interval", fmt.Sprintf("[%v, %v]", minFrom, maxTo))
	if err := w.queue.Dispatch(func() {
		process(user, minFrom, maxTo, out)
	}); err != nil {
		config.Log().Error("failed to dispatch wakatime import job for user", "userID", user.ID, "error", err)
	}

	return out, nil
}

func (w *WakatimeHeartbeatsImporter) ImportAll(user *models.User) (<-chan *models.Heartbeat, error) {
	return w.Import(user, config.BeginningOfWakatime(), time.Now())
}

// https://wakatime.com/api/v1/users/current/heartbeats?date=2021-02-05
// https://pastr.de/p/b5p4od5s8w0pfntmwoi117jy
func (w *WakatimeHeartbeatsImporter) fetchHeartbeats(day string, baseUrl string) ([]*wakatime.HeartbeatEntry, error) {
	req, err := http.NewRequest(http.MethodGet, baseUrl+config.WakatimeApiHeartbeatsUrl, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("date", day)
	req.URL.RawQuery = q.Encode()

	var empty []*wakatime.HeartbeatEntry

	res, err := w.httpClient.Do(w.withHeaders(req))
	if err != nil {
		return empty, err
	} else if res.StatusCode == 402 {
		return empty, nil // date outside free plan range -> return empty data, but do not throw error
	} else if res.StatusCode >= 400 {
		return empty, errors.New(fmt.Sprintf("got status %d from wakatime api", res.StatusCode))
	}
	defer res.Body.Close()

	var heartbeatsData wakatime.HeartbeatsViewModel
	if err := json.NewDecoder(res.Body).Decode(&heartbeatsData); err != nil {
		return empty, err
	}

	return heartbeatsData.Data, nil
}

// https://wakatime.com/api/v1/users/current/all_time_since_today
// https://pastr.de/p/w8xb4biv575pu32pox7jj2gr
func (w *WakatimeHeartbeatsImporter) fetchRange(baseUrl string) (time.Time, time.Time, error) {
	notime := config.BeginningOfWakatime()

	req, err := http.NewRequest(http.MethodGet, baseUrl+config.WakatimeApiAllTimeUrl, nil)
	if err != nil {
		return notime, notime, err
	}

	res, err := w.httpClient.Do(w.withHeaders(req))
	if err != nil {
		return notime, notime, err
	}

	// see https://github.com/muety/wakapi/issues/370
	allTimeData, err := utils.ParseJsonDropKeys[wakatime.AllTimeViewModel](res.Body, "text")
	if err != nil {
		return notime, notime, err
	}

	startDate, err := time.Parse(config.SimpleDateFormat, allTimeData.Data.Range.StartDate)
	if err != nil {
		return notime, notime, err
	}

	endDate, err := time.Parse(config.SimpleDateFormat, allTimeData.Data.Range.EndDate)
	if err != nil {
		return notime, notime, err
	}

	return startDate, endDate, nil
}

func (w *WakatimeHeartbeatsImporter) withHeaders(req *http.Request) *http.Request {
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(w.apiKey))))
	return req
}

func generateDays(from, to time.Time) []time.Time {
	days := make([]time.Time, 0)

	from = datetime.BeginOfDay(from)
	to = datetime.BeginOfDay(to.AddDate(0, 0, 1))

	for d := from; d.Before(to); d = d.AddDate(0, 0, 1) {
		days = append(days, d)
	}

	return days
}
