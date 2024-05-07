package relay

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/emvi/logbuch"
	"github.com/leandro-lugaresi/hub"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	routeutils "github.com/muety/wakapi/routes/utils"
	"github.com/patrickmn/go-cache"
	"io"
	"net/http"
	"time"
)

const maxFailuresPerDay = 100

// WakatimeRelayMiddleware is a middleware to conditionally relay heartbeats to Wakatime (and other compatible services)
type WakatimeRelayMiddleware struct {
	httpClient   *http.Client
	hashCache    *cache.Cache
	failureCache *cache.Cache
	eventBus     *hub.Hub
}

func NewWakatimeRelayMiddleware() *WakatimeRelayMiddleware {
	return &WakatimeRelayMiddleware{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		hashCache:    cache.New(10*time.Minute, 10*time.Minute),
		failureCache: cache.New(24*time.Hour, 1*time.Hour),
		eventBus:     config.EventBus(),
	}
}

func (m *WakatimeRelayMiddleware) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.ServeHTTP(w, r, h.ServeHTTP)
	})
}

func (m *WakatimeRelayMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	defer next(w, r)

	ownInstanceId := config.Get().InstanceId
	originInstanceId := r.Header.Get("X-Origin-Instance")

	if r.Method != http.MethodPost || originInstanceId == ownInstanceId {
		return
	}

	user := middlewares.GetPrincipal(r)
	if user == nil || user.WakatimeApiKey == "" {
		return
	}

	err := m.filterByCache(r)
	if err != nil {
		logbuch.Warn("%v", err)
		return
	}

	body, _ := io.ReadAll(r.Body)
	r.Body.Close()
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	// prevent cycles
	downstreamInstanceId := ownInstanceId
	if originInstanceId != "" {
		downstreamInstanceId = originInstanceId
	}

	headers := http.Header{
		"X-Machine-Name": r.Header.Values("X-Machine-Name"),
		"Content-Type":   r.Header.Values("Content-Type"),
		"Accept":         r.Header.Values("Accept"),
		"User-Agent":     r.Header.Values("User-Agent"),
		"X-Origin": []string{
			fmt.Sprintf("wakapi v%s", config.Get().Version),
		},
		"X-Origin-Instance": []string{downstreamInstanceId},
		"Authorization": []string{
			fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(user.WakatimeApiKey))),
		},
	}

	url := user.WakaTimeURL(config.WakatimeApiUrl) + config.WakatimeApiHeartbeatsBulkUrl

	go m.send(
		http.MethodPost,
		url,
		bytes.NewReader(body),
		headers,
		user,
	)
}

func (m *WakatimeRelayMiddleware) send(method, url string, body io.Reader, headers http.Header, forUser *models.User) {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		logbuch.Warn("error constructing relayed request - %v", err)
		return
	}

	for k, v := range headers {
		for _, h := range v {
			request.Header.Set(k, h)
		}
	}

	response, err := m.httpClient.Do(request)
	if err != nil {
		logbuch.Warn("error executing relayed request - %v", err)
		return
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		logbuch.Warn("failed to relay request for user %s, got status %d", forUser.ID, response.StatusCode)

		// TODO: use leaky bucket instead of expiring cache?
		if _, found := m.failureCache.Get(forUser.ID); !found {
			m.failureCache.SetDefault(forUser.ID, 0)
		}
		if n, _ := m.failureCache.IncrementInt(forUser.ID, 1); n == maxFailuresPerDay {
			m.eventBus.Publish(hub.Message{
				Name:   config.EventWakatimeFailure,
				Fields: map[string]interface{}{config.FieldUser: forUser, config.FieldPayload: n},
			})
		} else if n%10 == 0 {
			logbuch.Warn("%d / %d failed wakatime heartbeat relaying attempts for user %s within last 24 hours", n, maxFailuresPerDay, forUser.ID)
		}
	}
}

// filterByCache takes an HTTP request, tries to parse the body contents as heartbeats, checks against a local cache for whether a heartbeat has already been relayed before according to its hash and in-place filters these from the request's raw json body.
// This method operates on the raw body data (interface{}), because serialization of models.Heartbeat is not necessarily identical to what the CLI has actually sent.
// Purpose of this mechanism is mainly to prevent cyclic relays / loops.
// Caution: this method does in-place changes to the request.
func (m *WakatimeRelayMiddleware) filterByCache(r *http.Request) error {
	heartbeats, err := routeutils.ParseHeartbeats(r)
	if err != nil {
		return err
	}

	body, _ := io.ReadAll(r.Body)
	r.Body.Close()
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	var rawData interface{}
	if err := json.NewDecoder(io.NopCloser(bytes.NewBuffer(body))).Decode(&rawData); err != nil {
		return err
	}

	newData := make([]interface{}, 0, len(heartbeats))

	process := func(heartbeat *models.Heartbeat, rawData interface{}) {
		heartbeat = heartbeat.Hashed()
		// we didn't see this particular heartbeat before
		if _, found := m.hashCache.Get(heartbeat.Hash); !found {
			m.hashCache.SetDefault(heartbeat.Hash, true)
			newData = append(newData, rawData)
		}
	}

	if _, isList := rawData.([]interface{}); isList {
		for i, hb := range heartbeats {
			process(hb, rawData.([]interface{})[i])
		}
	} else if len(heartbeats) > 0 {
		process(heartbeats[0], rawData.(interface{}))
	}

	if len(newData) == 0 {
		return errors.New("no new heartbeats to relay")
	}

	if len(newData) != len(heartbeats) {
		user := middlewares.GetPrincipal(r)
		logbuch.Warn("only relaying %d of %d heartbeats for user %s", len(newData), len(heartbeats), user.ID)
	}

	buf := bytes.Buffer{}
	if err := json.NewEncoder(&buf).Encode(newData); err != nil {
		return err
	}
	r.Body = io.NopCloser(&buf)

	return nil
}
