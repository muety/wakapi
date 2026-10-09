package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	datastructure "github.com/duke-git/lancet/v2/datastructure/set"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/suite"
)

type stubAggregationService struct {
	lockedUsers map[string]bool
}

func (s *stubAggregationService) Schedule() {}
func (s *stubAggregationService) AggregateSummaries(u datastructure.Set[string], b bool) error { return nil }
func (s *stubAggregationService) AggregateDurations(u datastructure.Set[string]) error { return nil }
func (s *stubAggregationService) IsLocked(userId string) bool { return s.lockedUsers[userId] }

type SettingsHandlerTestSuite struct {
	suite.Suite
	Cfg *config.Config
	Sut *SettingsHandler
}

func TestSettingsHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(SettingsHandlerTestSuite))
}

func (suite *SettingsHandlerTestSuite) SetupSuite() {
}

func (suite *SettingsHandlerTestSuite) TearDownSuite() {
}

func (suite *SettingsHandlerTestSuite) BeforeTest(suiteName, testName string) {
	config.Set(config.Empty())
	suite.Cfg = config.Get()
	suite.Cfg.Env = "production"

	suite.Sut = &SettingsHandler{
		config:           suite.Cfg,
		aggregationLocks: make(map[string]bool),
	}
}

func (suite *SettingsHandlerTestSuite) TestSettingsHandler_TryLockAggregation() {
	user := "user123"

	suite.False(suite.Sut.isAggregationLocked(user))

	suite.True(suite.Sut.tryLockAggregation(user))
	suite.True(suite.Sut.isAggregationLocked(user))

	suite.False(suite.Sut.tryLockAggregation(user))

	suite.True(suite.Sut.tryLockAggregation("otherUser"))

	suite.Sut.toggleAggregationLock(user, false)
	suite.False(suite.Sut.isAggregationLocked(user))
	suite.True(suite.Sut.isAggregationLocked("otherUser"))

	suite.True(suite.Sut.tryLockAggregation(user))
}

func (suite *SettingsHandlerTestSuite) TestSettingsHandler_TryLockAggregation_ServiceLocked() {
	stub := &stubAggregationService{
		lockedUsers: map[string]bool{"serviceUser": true},
	}
	suite.Sut.aggregationSrvc = stub

	suite.False(suite.Sut.tryLockAggregation("serviceUser"))
	suite.True(suite.Sut.isAggregationLocked("serviceUser"))

	suite.True(suite.Sut.tryLockAggregation("freeUser"))
}

func (suite *SettingsHandlerTestSuite) TestSettingsHandler_TryLockAggregation_Concurrent() {
	user := "concurrentUser"
	const goroutines = 50
	var successCount atomic.Int32
	var wg sync.WaitGroup

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			if suite.Sut.tryLockAggregation(user) {
				successCount.Add(1)
			}
		}()
	}
	wg.Wait()

	suite.Equal(int32(1), successCount.Load())
	suite.True(suite.Sut.isAggregationLocked(user))
}

func (suite *SettingsHandlerTestSuite) TestSettingsHandler_ActionRegenerateSummaries_Conflict() {
	suite.Sut.tryLockAggregation("test_user")

	sharedData := config.NewSharedData()
	sharedData.Set(config.MiddlewareKeyPrincipal, &models.User{ID: "test_user"})
	ctx := context.WithValue(context.Background(), config.KeySharedData, sharedData)
	req := httptest.NewRequest(http.MethodPost, "/settings", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	result := suite.Sut.actionRegenerateSummaries(w, req)
	suite.Equal(http.StatusConflict, result.code)
	suite.Contains(result.error, "already in progress")
}
