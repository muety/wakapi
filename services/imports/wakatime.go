package imports

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/emvi/logbuch"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	wakatime "github.com/muety/wakapi/models/compat/wakatime/v1"
	"github.com/muety/wakapi/utils"
	"go.uber.org/atomic"
	"golang.org/x/sync/semaphore"
)

const OriginWakatime = "wakatime"
const (
	// wakatime api permits a max. rate of 10 req / sec
	// https://github.com/wakatime/wakatime/issues/261
	// with 5 workers, each sleeping slightly over 1/2 sec after every req., we should stay well below that limit
	maxWorkers    = 5
	throttleDelay = 550 * time.Millisecond
)

type WakatimeHeartbeatImporter struct {
	ApiKey string
}

func NewWakatimeHeartbeatImporter(apiKey string) *WakatimeHeartbeatImporter {
	return &WakatimeHeartbeatImporter{
		ApiKey: apiKey,
	}
}

func (w *WakatimeHeartbeatImporter) Import(user *models.User, minFrom time.Time, maxTo time.Time) <-chan *models.Heartbeat {
	out := make(chan *models.Heartbeat)

	go func(user *models.User, out chan *models.Heartbeat) {
		baseUrl := user.WakaTimeURL(config.WakatimeApiUrl)

		startDate, endDate, err := w.fetchRange(baseUrl)
		if err != nil {
			config.Log().Error("failed to fetch date range while importing wakatime heartbeats for user '%s' - %v", user.ID, err)
			return
		}

		if startDate.Before(minFrom) {
			startDate = minFrom
		}
		if endDate.After(maxTo) {
			endDate = maxTo
		}

		userAgents, err := w.fetchUserAgents(baseUrl)
		if err != nil {
			config.Log().Error("failed to fetch user agents while importing wakatime heartbeats for user '%s' - %v", user.ID, err)
			return
		}

		machinesNames, err := w.fetchMachineNames(baseUrl)
		if err != nil {
			config.Log().Error("failed to fetch machine names while importing wakatime heartbeats for user '%s' - %v", user.ID, err)
			return
		}

		days := generateDays(startDate, endDate)

		c := atomic.NewUint32(uint32(len(days)))
		ctx := context.TODO()
		sem := semaphore.NewWeighted(maxWorkers)

		for _, d := range days {
			if err := sem.Acquire(ctx, 1); err != nil {
				logbuch.Error("failed to acquire semaphore - %v", err)
				break
			}

			go func(day time.Time) {
				defer sem.Release(1)
				defer time.Sleep(throttleDelay)

				d := day.Format(config.SimpleDateFormat)
				heartbeats, err := w.fetchHeartbeats(d, baseUrl)
				if err != nil {
					config.Log().Error("failed to fetch heartbeats for day '%s' and user '%s' - &v", d, user.ID, err)
				}

				for _, h := range heartbeats {
					out <- mapHeartbeat(h, userAgents, machinesNames, user)
				}

				if c.Dec() == 0 {
					close(out)
				}
			}(d)
		}
	}(user, out)

	return out
}

func (w *WakatimeHeartbeatImporter) ImportAll(user *models.User) <-chan *models.Heartbeat {
	return w.Import(user, time.Time{}, time.Now())
}

// https://wakatime.com/api/v1/users/current/heartbeats?date=2021-02-05
// https://pastr.de/p/b5p4od5s8w0pfntmwoi117jy
func (w *WakatimeHeartbeatImporter) fetchHeartbeats(day string, baseUrl string) ([]*wakatime.HeartbeatEntry, error) {
	httpClient := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest(http.MethodGet, baseUrl+config.WakatimeApiHeartbeatsUrl, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("date", day)
	req.URL.RawQuery = q.Encode()

	res, err := httpClient.Do(w.withHeaders(req))
	if err != nil {
		return nil, err
	} else if res.StatusCode >= 400 {
		return nil, errors.New(fmt.Sprintf("got status %d from wakatime api", res.StatusCode))
	}

	var heartbeatsData wakatime.HeartbeatsViewModel
	if err := json.NewDecoder(res.Body).Decode(&heartbeatsData); err != nil {
		return nil, err
	}

	return heartbeatsData.Data, nil
}

// https://wakatime.com/api/v1/users/current/all_time_since_today
// https://pastr.de/p/w8xb4biv575pu32pox7jj2gr
func (w *WakatimeHeartbeatImporter) fetchRange(baseUrl string) (time.Time, time.Time, error) {
	httpClient := &http.Client{Timeout: 10 * time.Second}

	notime := time.Time{}

	req, err := http.NewRequest(http.MethodGet, baseUrl+config.WakatimeApiAllTimeUrl, nil)
	if err != nil {
		return notime, notime, err
	}

	res, err := httpClient.Do(w.withHeaders(req))
	if err != nil {
		return notime, notime, err
	}

	var allTimeData wakatime.AllTimeViewModel
	if err := json.NewDecoder(res.Body).Decode(&allTimeData); err != nil {
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

// https://wakatime.com/api/v1/users/current/user_agents
// https://pastr.de/p/05k5do8q108k94lic4lfl3pc
func (w *WakatimeHeartbeatImporter) fetchUserAgents(baseUrl string) (map[string]*wakatime.UserAgentEntry, error) {
	httpClient := &http.Client{Timeout: 10 * time.Second}

	userAgents := make(map[string]*wakatime.UserAgentEntry)

	for page := 1; ; page++ {
		url := fmt.Sprintf("%s%s?page=%d", baseUrl, config.WakatimeApiUserAgentsUrl, page)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		res, err := httpClient.Do(w.withHeaders(req))
		if err != nil {
			return nil, err
		}

		var userAgentsData wakatime.UserAgentsViewModel
		if err := json.NewDecoder(res.Body).Decode(&userAgentsData); err != nil {
			return nil, err
		}

		for _, ua := range userAgentsData.Data {
			userAgents[ua.Id] = ua
		}

		if page == userAgentsData.TotalPages {
			break
		}
	}

	return userAgents, nil
}

// https://wakatime.com/api/v1/users/current/machine_names
// https://pastr.de/p/v58cv0xrupp3zvyyv8o6973j
func (w *WakatimeHeartbeatImporter) fetchMachineNames(baseUrl string) (map[string]*wakatime.MachineEntry, error) {
	httpClient := &http.Client{Timeout: 10 * time.Second}

	machines := make(map[string]*wakatime.MachineEntry)

	for page := 1; ; page++ {
		url := fmt.Sprintf("%s%s?page=%d", baseUrl, config.WakatimeApiMachineNamesUrl, page)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		res, err := httpClient.Do(w.withHeaders(req))
		if err != nil {
			return nil, err
		}

		var machineData wakatime.MachineViewModel
		if err := json.NewDecoder(res.Body).Decode(&machineData); err != nil {
			return nil, err
		}

		for _, ma := range machineData.Data {
			machines[ma.Id] = ma
		}

		if page == machineData.TotalPages {
			break
		}
	}

	return machines, nil
}

func (w *WakatimeHeartbeatImporter) withHeaders(req *http.Request) *http.Request {
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(w.ApiKey))))
	return req
}

func mapHeartbeat(
	entry *wakatime.HeartbeatEntry,
	userAgents map[string]*wakatime.UserAgentEntry,
	machineNames map[string]*wakatime.MachineEntry,
	user *models.User,
) *models.Heartbeat {
	ua := userAgents[entry.UserAgentId]
	if ua == nil {
		ua = &wakatime.UserAgentEntry{
			Editor: "unknown",
			Os:     "unknown",
		}
	}

	ma := machineNames[entry.MachineNameId]
	if ma == nil {
		ma = &wakatime.MachineEntry{
			Id:    entry.MachineNameId,
			Value: entry.MachineNameId,
		}
	}

	return (&models.Heartbeat{
		User:            user,
		UserID:          user.ID,
		Entity:          entry.Entity,
		Type:            entry.Type,
		Category:        entry.Category,
		Project:         entry.Project,
		Branch:          entry.Branch,
		Language:        entry.Language,
		IsWrite:         entry.IsWrite,
		Editor:          ua.Editor,
		OperatingSystem: ua.Os,
		Machine:         ma.Value,
		UserAgent:       ua.Value,
		Time:            models.CustomTime(time.Unix(0, int64(entry.Time*1e9))),
		Origin:          OriginWakatime,
		OriginId:        entry.Id,
		CreatedAt:       models.CustomTime(entry.CreatedAt),
	}).Hashed()
}

func generateDays(from, to time.Time) []time.Time {
	days := make([]time.Time, 0)

	from = utils.StartOfDay(from)
	to = utils.StartOfDay(to.AddDate(0, 0, 1))

	for d := from; d.Before(to); d = d.AddDate(0, 0, 1) {
		days = append(days, d)
	}

	return days
}
