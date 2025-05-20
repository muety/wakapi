package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/muety/wakapi/config"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var testUserAgentsMatch = []string{
	"Mozilla/5.0 (X11; Linux x86_64; rv:138.0) Gecko/20100101 Firefox/138.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.10 Safari/605.1.1",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.3",
	"Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Mobile Safari/537.3",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 18_3_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.3.1 Mobile/15E148 Safari/604.",
	"Mozilla/5.0 (Linux; Android 4.3; GT-I9300 Build/JSS15J) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.91 Mobile Safari/537.36 OPR/42.9.2246.119956",
}

var testUserAgentsNoMatch = []string{
	"",
	"curl/7.81.0",
	"wakatime/13.0.7 (Linux-4.15.0-96-generic-x86_64-with-glibc2.4) Python3.8.0.final.0 GoLand/2019.3.4 GoLand-wakatime/11.0.1",
	"Chrome/114.0.0.0 linux_x86-64 chrome-wakatime/3.0.17",
	"Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; Googlebot/2.1; +http://www.google.com/bot.html) Chrome/W.X.Y.Z Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36 Edg/116.0.1938.62 win_x86-64 edge-wakatime/3.0.18",
	"Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko); compatible; GPTBot/1.1; +https://openai.com/gptbot",
}

func TestApiRootHandler_Get(t *testing.T) {
	config.Set(config.Empty())

	router := chi.NewRouter()
	apiRouter := chi.NewRouter()
	router.Mount("/api", apiRouter)

	apiRootHandler := NewApiRootHandler()
	apiRootHandler.RegisterRoutes(apiRouter)

	t.Run("when calling root route from a browser", func(t *testing.T) {
		t.Run("should redirect to front page", func(t *testing.T) {
			for _, ua := range testUserAgentsMatch {
				rec := httptest.NewRecorder()

				req := httptest.NewRequest(http.MethodGet, "/api", nil)
				req.Header.Set("User-Agent", ua)

				router.ServeHTTP(rec, req)
				res := rec.Result()
				defer res.Body.Close()

				assert.Equal(t, http.StatusFound, res.StatusCode)
				assert.Equal(t, "/", res.Header.Get("Location"))
			}
		})
	})

	t.Run("when calling root route from elsewhere", func(t *testing.T) {
		t.Run("should return not found", func(t *testing.T) {
			for _, ua := range testUserAgentsNoMatch {
				rec := httptest.NewRecorder()

				req := httptest.NewRequest(http.MethodGet, "/api", nil)
				req.Header.Set("User-Agent", ua)

				router.ServeHTTP(rec, req)
				res := rec.Result()
				defer res.Body.Close()

				assert.Equal(t, http.StatusNotFound, res.StatusCode)
			}
		})
	})
}
