package relay

import (
	"github.com/gorilla/mux"
	conf "github.com/muety/wakapi/config"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
)

const targetUrlHeader = "X-Target-URL"
const pathMatcherPattern = `^/api/(heartbeat|heartbeats|summary|users|v1/users|compat/wakatime)`

type RelayHandler struct {
	config *conf.Config
}

func NewRelayHandler() *RelayHandler {
	return &RelayHandler{
		config: conf.Get(),
	}
}

type filteringMiddleware struct {
	handler     http.Handler
	pathMatcher *regexp.Regexp
}

func newFilteringMiddleware() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return &filteringMiddleware{
			handler:     h,
			pathMatcher: regexp.MustCompile(pathMatcherPattern),
		}
	}
}

func (m *filteringMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	targetUrl, err := url.Parse(r.Header.Get(targetUrlHeader))
	if err != nil || !m.pathMatcher.MatchString(targetUrl.Path) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte{})
		return
	}
	m.handler.ServeHTTP(w, r)
}

func (h *RelayHandler) RegisterRoutes(router *mux.Router) {
	if !h.config.Security.EnableProxy {
		return
	}

	r := router.PathPrefix("/relay").Subrouter()
	r.Use(newFilteringMiddleware())
	r.Path("").HandlerFunc(h.Any)
}

func (h *RelayHandler) Any(w http.ResponseWriter, r *http.Request) {
	targetUrl, err := url.Parse(r.Header.Get(targetUrlHeader))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte{})
		return
	}

	p := httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL = targetUrl
			r.Host = targetUrl.Host
		},
	}

	p.ServeHTTP(w, r)
}

// @Summary Proxy an GET API request to another Wakapi instance
// @ID relay-get
// @Tags relay
// @Param X-Target-URL header string true "Original URL to perform the request to"
// @Failure 403 {string} string "Returned if request path is not whitelisted"
// @Failure 502 {string} string "Returned if upstream host is down"
// @Router /relay [get]
func (h *RelayHandler) alias1() {}

// @Summary Proxy an POST API request to another Wakapi instance
// @ID relay-post
// @Tags relay
// @Param X-Target-URL header string true "Original URL to perform the request to"
// @Failure 403 {string} string "Returned if request path is not whitelisted"
// @Failure 502 {string} string "Returned if upstream host is down"
// @Router /relay [post]
func (h *RelayHandler) alias2() {}

// @Summary Proxy an PUT API request to another Wakapi instance
// @ID relay-put
// @Tags relay
// @Param X-Target-URL header string true "Original URL to perform the request to"
// @Failure 403 {string} string "Returned if request path is not whitelisted"
// @Failure 502 {string} string "Returned if upstream host is down"
// @Router /relay [put]
func (h *RelayHandler) alias3() {}

// @Summary Proxy an PATCH API request to another Wakapi instance
// @ID relay-patch
// @Tags relay
// @Param X-Target-URL header string true "Original URL to perform the request to"
// @Failure 403 {string} string "Returned if request path is not whitelisted"
// @Failure 502 {string} string "Returned if upstream host is down"
// @Router /relay [patch]
func (h *RelayHandler) alias4() {}

// @Summary Proxy an DELETE API request to another Wakapi instance
// @ID relay-delete
// @Tags relay
// @Param X-Target-URL header string true "Original URL to perform the request to"
// @Failure 403 {string} string "Returned if request path is not whitelisted"
// @Failure 502 {string} string "Returned if upstream host is down"
// @Router /relay [delete]
func (h *RelayHandler) alias5() {}
