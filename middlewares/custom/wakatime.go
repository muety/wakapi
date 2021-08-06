package relay

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/emvi/logbuch"
	"github.com/leandro-lugaresi/hub"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	"github.com/patrickmn/go-cache"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const maxFailuresPerDay = 100

/* Middleware to conditionally relay heartbeats to Wakatime */
type WakatimeRelayMiddleware struct {
	httpClient   *http.Client
	failureCache *cache.Cache
	eventBus     *hub.Hub
}

func NewWakatimeRelayMiddleware() *WakatimeRelayMiddleware {
	return &WakatimeRelayMiddleware{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
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

	if r.Method != http.MethodPost {
		return
	}

	user := middlewares.GetPrincipal(r)
	if user == nil || user.WakatimeApiKey == "" {
		return
	}

	body, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	headers := http.Header{
		"X-Machine-Name": r.Header.Values("X-Machine-Name"),
		"Content-Type":   r.Header.Values("Content-Type"),
		"Accept":         r.Header.Values("Accept"),
		"User-Agent":     r.Header.Values("User-Agent"),
		"X-Origin": []string{
			fmt.Sprintf("wakapi v%s", config.Get().Version),
		},
		"Authorization": []string{
			fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(user.WakatimeApiKey))),
		},
	}

	go m.send(
		http.MethodPost,
		config.WakatimeApiUrl+config.WakatimeApiHeartbeatsBulkUrl,
		bytes.NewReader(body),
		headers,
		user,
	)
}

func (m *WakatimeRelayMiddleware) send(method, url string, body io.Reader, headers http.Header, forUser *models.User) {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		logbuch.Warn("error constructing relayed request – %v", err)
		return
	}

	for k, v := range headers {
		for _, h := range v {
			request.Header.Set(k, h)
		}
	}

	response, err := m.httpClient.Do(request)
	if err != nil {
		logbuch.Warn("error executing relayed request – %v", err)
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
