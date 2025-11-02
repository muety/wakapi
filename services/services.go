package services

import (
	"time"

	datastructure "github.com/duke-git/lancet/v2/datastructure/set"
	"gorm.io/gorm"

	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/models/types"
	"github.com/muety/wakapi/utils"
)

type IAggregationService interface {
	Schedule()
	AggregateSummaries(set datastructure.Set[string]) error
	AggregateDurations(set datastructure.Set[string]) error
}

type IMiscService interface {
	Schedule()
	CountTotalTime()
}

type IAliasService interface {
	Create(*models.Alias) (*models.Alias, error)
	Delete(*models.Alias) error
	DeleteMulti([]*models.Alias) error
	IsInitialized(string) bool
	InitializeUser(string) error
	GetByUser(string) ([]*models.Alias, error)
	GetByUserAndType(string, uint8) ([]*models.Alias, error)
	GetByUserAndKeyAndType(string, string, uint8) ([]*models.Alias, error)
	GetAliasOrDefault(string, uint8, string) (string, error)
}

type IHeartbeatService interface {
	Insert(*models.Heartbeat) error
	InsertBatch([]*models.Heartbeat) error
	Count(bool) (int64, error)
	CountByUser(*models.User) (int64, error)
	CountByUsers([]*models.User) ([]*models.CountByUser, error)
	GetAllWithin(time.Time, time.Time, *models.User) ([]*models.Heartbeat, error)
	GetAllWithinByFilters(time.Time, time.Time, *models.User, *models.Filters) ([]*models.Heartbeat, error)
	GetFirstAll() ([]*models.TimeByUser, error)
	GetLastAll() ([]*models.TimeByUser, error)
	GetRangeByUser(*models.User) (*models.RangeByUser, error)
	GetFirstByUser(*models.User) (time.Time, error)
	GetLastByUser(*models.User) (time.Time, error)
	GetLatestByUser(*models.User) (*models.Heartbeat, error)
	GetLatestByOriginAndUser(string, *models.User) (*models.Heartbeat, error)
	GetLatestByFilters(*models.User, *models.Filters) (*models.Heartbeat, error)
	GetEntitySetByUser(uint8, string) ([]string, error)
	StreamAllWithin(time.Time, time.Time, *models.User) (chan *models.Heartbeat, error)
	StreamAllWithinByFilters(time.Time, time.Time, *models.User, *models.Filters) (chan *models.Heartbeat, error)
	DeleteBefore(time.Time) error
	DeleteByUser(*models.User) error
	DeleteByUserBefore(*models.User, time.Time) error
	GetUserProjectStats(*models.User, time.Time, time.Time, *utils.PageParams, bool) ([]*models.ProjectStats, error)
	GetUserAgentsByUser(*models.User) ([]*models.UserAgent, error)
}

type IDiagnosticsService interface {
	Create(*models.Diagnostics) (*models.Diagnostics, error)
}

type IKeyValueService interface {
	GetString(string) (*models.KeyStringValue, error)
	MustGetString(string) *models.KeyStringValue
	GetByPrefix(string) ([]*models.KeyStringValue, error)
	PutString(*models.KeyStringValue) error
	DeleteString(string) error
	DeleteStringTx(string, *gorm.DB) error
	DeleteWildcard(string) error
	DeleteWildcardTx(string, *gorm.DB) error
	ReplaceKeySuffix(string, string) error
}

type ILanguageMappingService interface {
	GetById(uint) (*models.LanguageMapping, error)
	GetByUser(string) ([]*models.LanguageMapping, error)
	ResolveByUser(string) (map[string]string, error)
	Create(*models.LanguageMapping) (*models.LanguageMapping, error)
	Delete(mapping *models.LanguageMapping) error
}

type IProjectLabelService interface {
	GetById(uint) (*models.ProjectLabel, error)
	GetByUser(string) ([]*models.ProjectLabel, error)
	GetByUserGrouped(string) (map[string][]*models.ProjectLabel, error)
	GetByUserGroupedInverted(string) (map[string][]*models.ProjectLabel, error)
	Create(*models.ProjectLabel) (*models.ProjectLabel, error)
	Delete(*models.ProjectLabel) error
}

type IMailService interface {
	SendPasswordReset(*models.User, string) error
	SendWakatimeFailureNotification(*models.User, int) error
	SendImportNotification(*models.User, time.Duration, int) error
	SendReport(*models.User, *models.Report) error
	SendSubscriptionNotification(*models.User, bool) error
}

type IDurationService interface {
	Get(time.Time, time.Time, *models.User, *models.Filters, *time.Duration, bool) (models.Durations, error)
	Regenerate(*models.User, bool)
	RegenerateAll()
	DeleteByUser(*models.User) error
}

type ISummaryService interface {
	Aliased(time.Time, time.Time, *models.User, types.SummaryRetriever, *models.Filters, *time.Duration, bool) (*models.Summary, error)
	Retrieve(time.Time, time.Time, *models.User, *models.Filters, *time.Duration) (*models.Summary, error)
	Summarize(time.Time, time.Time, *models.User, *models.Filters, *time.Duration) (*models.Summary, error)
	GetLatestByUser() ([]*models.TimeByUser, error)
	DeleteByUser(string) error
	DeleteByUserBefore(string, time.Time) error
	Insert(*models.Summary) error
}

type IActivityService interface {
	GetChart(*models.User, *models.IntervalKey, bool, bool, bool) (string, error)
}

type IReportService interface {
	Schedule()
	SendReport(*models.User, time.Duration) error
}

type IHousekeepingService interface {
	Schedule()
	CleanUserDataBefore(*models.User, time.Time) error
}

type ILeaderboardService interface {
	GetDefaultScope() *models.IntervalKey
	Schedule()
	ComputeLeaderboard([]*models.User, *models.IntervalKey, []uint8) error
	ExistsAnyByUser(string) (bool, error)
	CountUsers(bool) (int64, error)
	GetByInterval(*models.IntervalKey, *utils.PageParams, bool) (models.Leaderboard, error)
	GetByIntervalAndUser(*models.IntervalKey, string, bool) (models.Leaderboard, error)
	GetAggregatedByInterval(*models.IntervalKey, *uint8, *utils.PageParams, bool) (models.Leaderboard, error)
	GetAggregatedByIntervalAndUser(*models.IntervalKey, string, *uint8, bool) (models.Leaderboard, error)
	GenerateByUser(*models.User, *models.IntervalKey) (*models.LeaderboardItem, error)
	GenerateAggregatedByUser(*models.User, *models.IntervalKey, uint8) ([]*models.LeaderboardItem, error)
}

type IUserService interface {
	GetUserById(string) (*models.User, error)
	GetUserByKey(string, bool) (*models.User, error)
	GetUserByEmail(string) (*models.User, error)
	GetUserByResetToken(string) (*models.User, error)
	GetUserByUnsubscribeToken(string) (*models.User, error)
	GetUserByStripeCustomerId(string) (*models.User, error)
	GetUserByOidc(string, string) (*models.User, error)
	GetAll() ([]*models.User, error)
	GetAllMapped() (map[string]*models.User, error)
	GetMany([]string) ([]*models.User, error)
	GetManyMapped([]string) (map[string]*models.User, error)
	GetAllByReports(bool) ([]*models.User, error)
	GetAllByLeaderboard(bool) ([]*models.User, error)
	GetActive(bool) ([]*models.User, error)
	Count() (int64, error)
	CountCurrentlyOnline() (int, error)
	CreateOrGet(*models.Signup, bool) (*models.User, bool, error)
	Update(*models.User) (*models.User, error)
	Delete(*models.User) error
	ChangeUserId(*models.User, string) (*models.User, error)
	ResetApiKey(*models.User) (*models.User, error)
	SetWakatimeApiCredentials(*models.User, string, string) (*models.User, error)
	GenerateResetToken(*models.User) (*models.User, error)
	GenerateUnsubscribeToken(*models.User) (*models.User, error)
	FlushCache()
	FlushUserCache(string)
}

type IApiKeyService interface {
	GetByApiKey(string, bool) (*models.ApiKey, error)
	GetByUser(string) ([]*models.ApiKey, error)
	Create(*models.ApiKey) (*models.ApiKey, error)
	Delete(*models.ApiKey) error
}
