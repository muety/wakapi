package relay

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/leandro-lugaresi/hub"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	routeutils "github.com/muety/wakapi/routes/utils"
	"github.com/patrickmn/go-cache"
)

const maxFailuresPerDay = 100

// not really an error, it's the "seen all of these already, go back to sleep" signal
var errNoNewHeartbeats = errors.New("no new heartbeats to relay")

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
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse // forbid following redirects
			},
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

	if err := routeutils.ValidateWakatimeUrl(user.WakatimeApiUrl); err != nil {
		config.Log().Request(r).Error("failed to validate wakatime url while relaying", "url", user.WakatimeApiUrl, "error", err)
		return
	}

	// relayBody goes downstream, r.Body still has the full thing for the local store handler
	relayBody, err := m.filterByCache(r)
	if errors.Is(err, errNoNewHeartbeats) {
		return
	}
	if err != nil {
		slog.Warn("filter cache error", "error", err)
		return
	}

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
		bytes.NewReader(relayBody),
		headers,
		user,
	)
}

func (m *WakatimeRelayMiddleware) send(method, url string, body io.Reader, headers http.Header, forUser *models.User) {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		slog.Warn("error constructing relayed request", "error", err)
		return
	}

	for k, v := range headers {
		for _, h := range v {
			request.Header.Set(k, h)
		}
	}

	response, err := m.httpClient.Do(request)
	if err != nil {
		slog.Warn("error executing relayed request", "error", err)
		return
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		slog.Warn("failed to relay request for user", "userID", forUser.ID, "statusCode", response.StatusCode)

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
			slog.Warn("failed wakatime heartbeat relaying attempts for user", "failedCount", n, "maxFailures", maxFailuresPerDay, "userID", forUser.ID)
		}
	}
}

// filterByCache returns the JSON body for the relay request, minus any heartbeats we've already forwarded.
// Works on the raw decoded form (interface{}) since models.Heartbeat doesn't round-trip 1:1 with what the CLI ships.
// Point of all this: stop two linked instances from playing heartbeat ping-pong forever.
// Original request body is left untouched so whoever runs after us still sees the full list.
func (m *WakatimeRelayMiddleware) filterByCache(r *http.Request) ([]byte, error) {
	heartbeats, err := routeutils.ParseHeartbeats(r)
	if err != nil {
		return nil, err
	}

	// ParseHeartbeats already drained r.Body and put it back.
	// Reading it again here, we need the raw form, because models.Heartbeat drops fields the CLI sends that we'd rather forward as-is though.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	var rawData interface{}
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&rawData); err != nil {
		return nil, err
	}

	newData := make([]interface{}, 0, len(heartbeats))

	process := func(heartbeat *models.Heartbeat, rawData interface{}) {
		if heartbeat == nil {
			return // just in case client send [null] or sth., shouldn't happen though
		}
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
		return nil, errNoNewHeartbeats
	}

	if len(newData) != len(heartbeats) {
		user := middlewares.GetPrincipal(r)
		slog.Warn("only relaying partial heartbeats for user", "relayedCount", len(newData), "totalCount", len(heartbeats), "userID", user.ID)
	}

	buf := bytes.Buffer{}
	if err := json.NewEncoder(&buf).Encode(newData); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
