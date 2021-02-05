package imports

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/emvi/logbuch"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	wakatime "github.com/muety/wakapi/models/compat/wakatime/v1"
	"github.com/muety/wakapi/utils"
	"go.uber.org/atomic"
	"golang.org/x/sync/semaphore"
	"net/http"
	"time"
)

const (
	maxWorkers = 6
)

type WakatimeHeartbeatImporter struct {
	ApiKey string
}

func NewWakatimeHeartbeatImporter(apiKey string) *WakatimeHeartbeatImporter {
	return &WakatimeHeartbeatImporter{
		ApiKey: apiKey,
	}
}

func (w *WakatimeHeartbeatImporter) Import(user *models.User) <-chan *models.Heartbeat {
	out := make(chan *models.Heartbeat)

	go func(user *models.User, out chan *models.Heartbeat) {
		startDate, endDate, err := w.fetchRange()
		if err != nil {
			logbuch.Error("failed to fetch date range while importing wakatime heartbeats for user '%s' – %v", user.ID, err)
			return
		}

		userAgents, err := w.fetchUserAgents()
		if err != nil {
			logbuch.Error("failed to fetch user agents while importing wakatime heartbeats for user '%s' – %v", user.ID, err)
			return
		}

		days := generateDays(startDate, endDate)

		c := atomic.NewUint32(uint32(len(days)))
		ctx := context.TODO()
		sem := semaphore.NewWeighted(maxWorkers)

		for _, d := range days {
			if err := sem.Acquire(ctx, 1); err != nil {
				logbuch.Error("failed to acquire semaphore – %v", err)
				break
			}

			go func(day time.Time) {
				defer sem.Release(1)

				d := day.Format("2006-01-02")
				heartbeats, err := w.fetchHeartbeats(d)
				if err != nil {
					logbuch.Error("failed to fetch heartbeats for day '%s' and user '%s' – &v", day, user.ID, err)
				}

				for _, h := range heartbeats {
					out <- mapHeartbeat(h, userAgents, user)
				}

				if c.Dec() == 0 {
					close(out)
				}
			}(d)
		}
	}(user, out)

	return out
}

// https://wakatime.com/api/v1/users/current/heartbeats?date=2021-02-05
func (w *WakatimeHeartbeatImporter) fetchHeartbeats(day string) ([]*wakatime.HeartbeatEntry, error) {
	httpClient := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest(http.MethodGet, config.WakatimeApiUrl+config.WakatimeApiHeartbeatsUrl, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("date", day)
	req.URL.RawQuery = q.Encode()

	res, err := httpClient.Do(w.withHeaders(req))
	if err != nil {
		return nil, err
	}

	var heartbeatsData wakatime.HeartbeatsViewModel
	if err := json.NewDecoder(res.Body).Decode(&heartbeatsData); err != nil {
		return nil, err
	}

	return heartbeatsData.Data, nil
}

// https://wakatime.com/api/v1/users/current/all_time_since_today
func (w *WakatimeHeartbeatImporter) fetchRange() (time.Time, time.Time, error) {
	httpClient := &http.Client{Timeout: 10 * time.Second}

	notime := time.Time{}

	req, err := http.NewRequest(http.MethodGet, config.WakatimeApiUrl+config.WakatimeApiAllTimeUrl, nil)
	if err != nil {
		return notime, notime, err
	}

	res, err := httpClient.Do(w.withHeaders(req))
	if err != nil {
		return notime, notime, err
	}

	var allTimeData map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&allTimeData); err != nil {
		return notime, notime, err
	}

	data := allTimeData["data"].(map[string]interface{})
	if data == nil {
		return notime, notime, errors.New("invalid response")
	}

	dataRange := data["range"].(map[string]interface{})
	if dataRange == nil {
		return notime, notime, errors.New("invalid response")
	}

	startDate, err := time.Parse("2006-01-02", dataRange["start_date"].(string))
	if err != nil {
		return notime, notime, err
	}

	endDate, err := time.Parse("2006-01-02", dataRange["end_date"].(string))
	if err != nil {
		return notime, notime, err
	}

	return startDate, endDate, nil
}

// https://wakatime.com/api/v1/users/current/user_agents
func (w *WakatimeHeartbeatImporter) fetchUserAgents() (map[string]*wakatime.UserAgentEntry, error) {
	httpClient := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest(http.MethodGet, config.WakatimeApiUrl+config.WakatimeApiUserAgentsUrl, nil)
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

	userAgents := make(map[string]*wakatime.UserAgentEntry)
	for _, ua := range userAgentsData.Data {
		userAgents[ua.Id] = ua
	}

	return userAgents, nil
}

func (w *WakatimeHeartbeatImporter) withHeaders(req *http.Request) *http.Request {
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(w.ApiKey))))
	return req
}

func mapHeartbeat(
	entry *wakatime.HeartbeatEntry,
	userAgents map[string]*wakatime.UserAgentEntry,
	user *models.User,
) *models.Heartbeat {
	ua := userAgents[entry.UserAgentId]
	if ua == nil {
		ua = &wakatime.UserAgentEntry{
			Editor: "unknown",
			Os:     "unknown",
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
		Machine:         entry.MachineNameId, // TODO
		Time:            entry.Time,
		Origin:          fmt.Sprintf("wt@%s", entry.Id),
	}).Hashed()
}

func generateDays(from, to time.Time) []time.Time {
	days := make([]time.Time, 0)

	from = utils.StartOfDay(from)
	to = utils.StartOfDay(to.Add(24 * time.Hour))

	for d := from; d.Before(to); d = d.Add(24 * time.Hour) {
		days = append(days, d)
	}

	return days
}
