package relay

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	WakatimeApiUrl                = "https://wakatime.com/api/v1"
	WakatimeApiHeartbeatsEndpoint = "/users/current/heartbeats.bulk"
)

/* Middleware to conditionally relay heartbeats to Wakatime */
type WakatimeRelayMiddleware struct {
	httpClient *http.Client
}

func NewWakatimeRelayMiddleware() *WakatimeRelayMiddleware {
	return &WakatimeRelayMiddleware{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
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

	user := r.Context().Value(models.UserKey).(*models.User)
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
		WakatimeApiUrl+WakatimeApiHeartbeatsEndpoint,
		bytes.NewReader(body),
		headers,
	)
}

func (m *WakatimeRelayMiddleware) send(method, url string, body io.Reader, headers http.Header) {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Printf("error constructing relayed request – %v\n", err)
		return
	}

	for k, v := range headers {
		for _, h := range v {
			request.Header.Set(k, h)
		}
	}

	response, err := m.httpClient.Do(request)
	if err != nil {
		log.Printf("error executing relayed request – %v\n", err)
		return
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		log.Printf("failed to relay request, got status %d\n", response.StatusCode)
	}
}
