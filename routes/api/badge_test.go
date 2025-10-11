package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/routes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	user1 = models.User{
		ID:               "user1",
		ShareDataMaxDays: 30,
		ShareLanguages:   true,
	}

	summary1 = models.Summary{
		User:     &user1,
		UserID:   "user1",
		FromTime: models.CustomTime(time.Date(2023, 3, 14, 0, 0, 0, 0, time.Local)),
		ToTime:   models.CustomTime(time.Date(2023, 3, 14, 23, 59, 59, 0, time.Local)),
		Languages: []*models.SummaryItem{
			{
				Type:  models.SummaryLanguage,
				Key:   "go",
				Total: 12 * time.Minute / time.Second,
			},
		},
	}
)

func TestBadgeHandler_Get(t *testing.T) {
	config.Set(config.Empty())

	router := chi.NewRouter()
	apiRouter := chi.NewRouter()
	apiRouter.Use(middlewares.NewSharedDataMiddleware())
	router.Mount("/api", apiRouter)

	userServiceMock := new(mocks.UserServiceMock)
	userServiceMock.On("GetUserById", "user1").Return(&user1, nil)

	summaryServiceMock := new(mocks.SummaryServiceMock)
	summaryServiceMock.On("Aliased", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), &user1, mock.AnythingOfType("types.SummaryRetriever"), mock.AnythingOfType("*models.Filters"), mock.AnythingOfType("*time.Duration"), mock.Anything).Return(&summary1, nil)

	badgeHandler := NewBadgeHandler(userServiceMock, summaryServiceMock)
	badgeHandler.RegisterRoutes(apiRouter)

	t.Run("when requesting badge", func(t *testing.T) {
		t.Run("should return badge", func(t *testing.T) {
			rec := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, "/api/badge/{user}/interval:week/language:go", nil)
			req = routes.WithUrlParam(req, "user", "user1")

			router.ServeHTTP(rec, req)
			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, http.StatusOK, res.StatusCode)

			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("unextected error. Error: %s", err)
			}

			assert.True(t, strings.HasPrefix(string(data), "<svg")) // alternatively, use assert.HTTPBodyContains() ?
			assert.Contains(t, string(data), "0 hrs 12 mins")
		})

		t.Run("should not return badge if shared interval exceeded", func(t *testing.T) {
			rec := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, "/api/badge/{user}/interval:last_year/language:go", nil)
			req = routes.WithUrlParam(req, "user", "user1")

			router.ServeHTTP(rec, req)
			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, http.StatusForbidden, res.StatusCode)

			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("unextected error. Error: %s", err)
			}

			assert.False(t, strings.HasPrefix(string(data), "<svg"))
		})

		t.Run("should not return badge if entity type not shared", func(t *testing.T) {
			rec := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, "/api/badge/{user}/interval:year/project:foo", nil)
			req = routes.WithUrlParam(req, "user", "user1")

			router.ServeHTTP(rec, req)
			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, http.StatusForbidden, res.StatusCode)

			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("unextected error. Error: %s", err)
			}

			assert.False(t, strings.HasPrefix(string(data), "<svg"))
		})
	})
}

func TestBadgeHandler_EntityPattern(t *testing.T) {
	type test struct {
		test string
		key  string
		val  string
	}

	pathPrefix := "/compat/shields/v1/current/today/"

	tests := []test{
		{test: pathPrefix + "project:wakapi", key: "project", val: "wakapi"},
		{test: pathPrefix + "os:Linux", key: "os", val: "Linux"},
		{test: pathPrefix + "editor:VSCode", key: "editor", val: "VSCode"},
		{test: pathPrefix + "language:Java", key: "language", val: "Java"},
		{test: pathPrefix + "machine:devmachine", key: "machine", val: "devmachine"},
		{test: pathPrefix + "label:work", key: "label", val: "work"},
		{test: pathPrefix + "foo:bar", key: "", val: ""},                                   // invalid entity
		{test: pathPrefix + "project:01234", key: "project", val: "01234"},                 // digits only
		{test: pathPrefix + "project:anchr-web-ext", key: "project", val: "anchr-web-ext"}, // with dashes
		{test: pathPrefix + "project:wakapi v2", key: "project", val: "wakapi v2"},         // with blank space
		{test: pathPrefix + "project:project", key: "project", val: "project"},
		{test: pathPrefix + "project:Anchr-Android_v2.0", key: "project", val: "Anchr-Android_v2.0"}, // all the way
	}

	sut := regexp.MustCompile(`(project|os|editor|language|machine|label):([^:?&/]+)`) // see entityFilterPattern in badge_utils.go

	for _, tc := range tests {
		var key, val string
		if groups := sut.FindStringSubmatch(tc.test); len(groups) > 2 {
			key, val = groups[1], groups[2]
		}
		assert.Equal(t, tc.key, key)
		assert.Equal(t, tc.val, val)
	}
}
