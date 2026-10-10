package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type SettingsHandlerTestSuite struct {
	suite.Suite
	Cfg                    *config.Config
	Sut                    *SettingsHandler
	IsLockedMockCall       *mock.Call
	UserService            *mocks.UserServiceMock
	HeartbeatService       *mocks.HeartbeatServiceMock
	DurationService        *mocks.DurationServiceMock
	SummaryService         *mocks.SummaryServiceMock
	AliasService           *mocks.AliasServiceMock
	AggregationService     *mocks.AggregationServiceMock
	LanguageMappingService *mocks.LanguageMappingServiceMock
	ProjectLabelService    *mocks.ProjectLabelServiceMock
	KeyValueService        *mocks.KeyValueServiceMock
	MailService            *mocks.MailServiceMock
	ApiKeyService          *mocks.MockApiKeyService
	WebAuthnService        *mocks.WebAuthnServiceMock
}

func TestSettingsHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(SettingsHandlerTestSuite))
}

func (suite *SettingsHandlerTestSuite) BeforeTest(suiteName, testName string) {
	suite.Cfg = config.Empty()
	suite.Cfg.Env = "production"
	config.Set(suite.Cfg)

	suite.UserService = new(mocks.UserServiceMock)
	suite.HeartbeatService = new(mocks.HeartbeatServiceMock)
	suite.DurationService = new(mocks.DurationServiceMock)
	suite.SummaryService = new(mocks.SummaryServiceMock)
	suite.AliasService = new(mocks.AliasServiceMock)
	suite.AggregationService = new(mocks.AggregationServiceMock)
	suite.LanguageMappingService = new(mocks.LanguageMappingServiceMock)
	suite.ProjectLabelService = new(mocks.ProjectLabelServiceMock)
	suite.KeyValueService = new(mocks.KeyValueServiceMock)
	suite.MailService = new(mocks.MailServiceMock)
	suite.ApiKeyService = new(mocks.MockApiKeyService)
	suite.WebAuthnService = new(mocks.WebAuthnServiceMock)

	suite.Sut = NewSettingsHandler(
		suite.UserService,
		suite.HeartbeatService,
		suite.DurationService,
		suite.SummaryService,
		suite.AliasService,
		suite.AggregationService,
		suite.LanguageMappingService,
		suite.ProjectLabelService,
		suite.KeyValueService,
		suite.MailService,
		suite.ApiKeyService,
		suite.WebAuthnService,
	)

	suite.IsLockedMockCall = suite.AggregationService.On("IsLocked", mock.Anything).Return(false).Maybe()
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
	suite.IsLockedMockCall.Unset()
	suite.AggregationService.On("IsLocked", "serviceUser").Return(true).Maybe()
	suite.AggregationService.On("IsLocked", mock.Anything).Return(false).Maybe()

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
	for range goroutines {
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
