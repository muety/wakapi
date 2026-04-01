package relay

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MockRoundTripper struct {
	mock.Mock
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

type WakatimeRelayMiddlewareTestSuite struct {
	suite.Suite
	sut              *WakatimeRelayMiddleware
	conf             *config.Config
	mockRoundTripper *MockRoundTripper
}

func TestWakatimeRelayMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(WakatimeRelayMiddlewareTestSuite))
}

func (suite *WakatimeRelayMiddlewareTestSuite) SetupTest() {
	suite.conf = config.Empty()
	suite.conf.InstanceId = "test-instance"
	suite.conf.Version = "1.0.0"
	suite.conf.Server.PublicNetUrl, _ = url.Parse("https://wakapi.dev")
	config.Set(suite.conf)

	suite.mockRoundTripper = new(MockRoundTripper)
	suite.sut = NewWakatimeRelayMiddleware()
	suite.sut.httpClient.Transport = suite.mockRoundTripper
}

func (suite *WakatimeRelayMiddlewareTestSuite) TestServeHTTP_SkipNonPost() {
	req, _ := http.NewRequest(http.MethodGet, "/api/heartbeats", nil)
	rr := httptest.NewRecorder()

	calledNext := false
	next := func(w http.ResponseWriter, r *http.Request) {
		calledNext = true
	}

	suite.sut.ServeHTTP(rr, req, next)
	suite.True(calledNext)
	suite.mockRoundTripper.AssertNotCalled(suite.T(), "RoundTrip", mock.Anything)
}

func (suite *WakatimeRelayMiddlewareTestSuite) TestServeHTTP_SkipOwnInstance() {
	req, _ := http.NewRequest(http.MethodPost, "/api/heartbeats", nil)
	req.Header.Set("X-Origin-Instance", "test-instance")
	rr := httptest.NewRecorder()

	calledNext := false
	next := func(w http.ResponseWriter, r *http.Request) {
		calledNext = true
	}

	suite.sut.ServeHTTP(rr, req, next)
	suite.True(calledNext)
	suite.mockRoundTripper.AssertNotCalled(suite.T(), "RoundTrip", mock.Anything)
}

func (suite *WakatimeRelayMiddlewareTestSuite) TestServeHTTP_SkipNoUser() {
	req, _ := http.NewRequest(http.MethodPost, "/api/heartbeats", nil)
	rr := httptest.NewRecorder()

	calledNext := false
	next := func(w http.ResponseWriter, r *http.Request) {
		calledNext = true
	}

	suite.sut.ServeHTTP(rr, req, next)
	suite.True(calledNext)
	suite.mockRoundTripper.AssertNotCalled(suite.T(), "RoundTrip", mock.Anything)
}

func (suite *WakatimeRelayMiddlewareTestSuite) TestServeHTTP_SkipNoApiKey() {
	user := &models.User{ID: "test-user", WakatimeApiKey: ""}
	req, _ := http.NewRequest(http.MethodPost, "/api/heartbeats", nil)
	req = suite.withUser(req, user)
	rr := httptest.NewRecorder()

	calledNext := false
	next := func(w http.ResponseWriter, r *http.Request) {
		calledNext = true
	}

	suite.sut.ServeHTTP(rr, req, next)
	suite.True(calledNext)
	suite.mockRoundTripper.AssertNotCalled(suite.T(), "RoundTrip", mock.Anything)
}

func (suite *WakatimeRelayMiddlewareTestSuite) TestServeHTTP_InvalidWakatimeUrl() {
	// URL that refers back to own instance should be invalid according to ValidateWakatimeUrl
	user := &models.User{
		ID:             "test-user",
		WakatimeApiKey: "waka_123",
		WakatimeApiUrl: "https://wakapi.dev/api",
	}
	req, _ := http.NewRequest(http.MethodPost, "/api/heartbeats", nil)
	req = suite.withUser(req, user)
	rr := httptest.NewRecorder()

	calledNext := false
	next := func(w http.ResponseWriter, r *http.Request) {
		calledNext = true
	}

	suite.sut.ServeHTTP(rr, req, next)
	suite.True(calledNext)
	suite.mockRoundTripper.AssertNotCalled(suite.T(), "RoundTrip", mock.Anything)
}

func (suite *WakatimeRelayMiddlewareTestSuite) TestServeHTTP_Success() {
	user := &models.User{
		ID:             "test-user",
		WakatimeApiKey: "waka_123",
		WakatimeApiUrl: "https://api.wakatime.com/api/v1",
	}

	hb := &models.Heartbeat{
		Entity: "/tmp/test.go",
		Type:   "file",
		Time:   models.CustomTime(time.Now()),
	}
	hb.UserID = user.ID
	hb.User = user

	body, _ := json.Marshal(map[string]interface{}{
		"entity": hb.Entity,
		"type":   hb.Type,
		"time":   float64(hb.Time.T().UnixNano()) / 1e9,
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/heartbeats", bytes.NewBuffer(body))
	req = suite.withUser(req, user)
	rr := httptest.NewRecorder()

	suite.mockRoundTripper.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "https://api.wakatime.com/api/v1/users/current/heartbeats.bulk" &&
			req.Header.Get("Authorization") != ""
	})).Return(&http.Response{
		StatusCode: http.StatusCreated,
		Body:       io.NopCloser(bytes.NewBufferString("{}")),
	}, nil).Once()

	suite.sut.ServeHTTP(rr, req, func(w http.ResponseWriter, r *http.Request) {})

	time.Sleep(50 * time.Millisecond)
	suite.mockRoundTripper.AssertExpectations(suite.T())
}

func (suite *WakatimeRelayMiddlewareTestSuite) TestSend_RedirectForbidden() {
	// Re-initialize httpClient for this test to use real Transport but our CheckRedirect
	suite.sut.httpClient.Transport = http.DefaultTransport

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://google.com", http.StatusFound)
	}))
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPost, ts.URL, nil)
	resp, err := suite.sut.httpClient.Do(req)

	suite.NoError(err)
	suite.Equal(http.StatusFound, resp.StatusCode)
	suite.Equal("https://google.com", resp.Header.Get("Location")) // not a 200 from google.com
}

func (suite *WakatimeRelayMiddlewareTestSuite) TestFilterByCache_Bulk() {
	user := &models.User{
		ID:             "test-user",
		WakatimeApiKey: "waka_123",
	}

	now := time.Now()
	hb1 := map[string]interface{}{
		"entity": "/tmp/1.go",
		"type":   "file",
		"time":   float64(now.UnixNano()) / 1e9,
	}
	hb2 := map[string]interface{}{
		"entity": "/tmp/2.go",
		"type":   "file",
		"time":   float64(now.Add(time.Second).UnixNano()) / 1e9,
	}

	body, _ := json.Marshal([]interface{}{hb1, hb2})
	req, _ := http.NewRequest(http.MethodPost, "/api/heartbeats", bytes.NewBuffer(body))
	req = suite.withUser(req, user)

	err := suite.sut.filterByCache(req)
	suite.NoError(err)

	// Now hb1 and hb2 are in cache. Try again with hb2 and hb3
	hb3 := map[string]interface{}{
		"entity": "/tmp/3.go",
		"type":   "file",
		"time":   float64(now.Add(2*time.Second).UnixNano()) / 1e9,
	}
	body2, _ := json.Marshal([]interface{}{hb2, hb3})
	req2, _ := http.NewRequest(http.MethodPost, "/api/heartbeats", bytes.NewBuffer(body2))
	req2 = suite.withUser(req2, user)

	err = suite.sut.filterByCache(req2)
	suite.NoError(err)

	var filtered []interface{}
	json.NewDecoder(req2.Body).Decode(&filtered)
	suite.Len(filtered, 1)
	suite.Equal("/tmp/3.go", filtered[0].(map[string]interface{})["entity"])
}

func (suite *WakatimeRelayMiddlewareTestSuite) withUser(r *http.Request, user *models.User) *http.Request {
	sd := config.NewSharedData()
	sd.Set(config.MiddlewareKeyPrincipal, user)
	return r.WithContext(context.WithValue(r.Context(), config.KeySharedData, sd))
}
