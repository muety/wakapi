package services

import (
	datastructure "github.com/duke-git/lancet/v2/datastructure/set"
	"github.com/muety/wakapi/models"
	"time"
)

type IAggregationService interface {
	Schedule()
	AggregateSummaries(set datastructure.Set[string]) error
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
	GetFirstByUsers() ([]*models.TimeByUser, error)
	GetLatestByUser(*models.User) (*models.Heartbeat, error)
	GetLatestByOriginAndUser(string, *models.User) (*models.Heartbeat, error)
	GetEntitySetByUser(uint8, *models.User) ([]string, error)
	DeleteBefore(time.Time) error
	DeleteByUser(*models.User) error
	DeleteByUserBefore(*models.User, time.Time) error
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
}

type IDurationService interface {
	Get(time.Time, time.Time, *models.User, *models.Filters) (models.Durations, error)
}

type ISummaryService interface {
	Aliased(time.Time, time.Time, *models.User, SummaryRetriever, *models.Filters, bool) (*models.Summary, error)
	Retrieve(time.Time, time.Time, *models.User, *models.Filters) (*models.Summary, error)
	Summarize(time.Time, time.Time, *models.User, *models.Filters) (*models.Summary, error)
	GetLatestByUser() ([]*models.TimeByUser, error)
	DeleteByUser(string) error
	DeleteByUserBefore(string, time.Time) error
	Insert(*models.Summary) error
}

type IReportService interface {
	Schedule()
	SendReport(*models.User, time.Duration) error
}

type IHousekeepingService interface {
	Schedule()
	ClearOldUserData(*models.User, time.Duration) error
}

type ILeaderboardService interface {
	Schedule()
	ComputeLeaderboard([]*models.User, *models.IntervalKey, []uint8) error
	ExistsAnyByUser(string) (bool, error)
	CountUsers() (int64, error)
	GetByInterval(*models.IntervalKey, *models.PageParams, bool) (models.Leaderboard, error)
	GetByIntervalAndUser(*models.IntervalKey, string, bool) (models.Leaderboard, error)
	GetAggregatedByInterval(*models.IntervalKey, *uint8, *models.PageParams, bool) (models.Leaderboard, error)
	GetAggregatedByIntervalAndUser(*models.IntervalKey, string, *uint8, bool) (models.Leaderboard, error)
	GenerateByUser(*models.User, *models.IntervalKey) (*models.LeaderboardItem, error)
	GenerateAggregatedByUser(*models.User, *models.IntervalKey, uint8) ([]*models.LeaderboardItem, error)
}

type IUserService interface {
	GetUserById(string) (*models.User, error)
	GetUserByKey(string) (*models.User, error)
	GetUserByEmail(string) (*models.User, error)
	GetUserByResetToken(string) (*models.User, error)
	GetAll() ([]*models.User, error)
	GetMany([]string) ([]*models.User, error)
	GetManyMapped([]string) (map[string]*models.User, error)
	GetAllByReports(bool) ([]*models.User, error)
	GetAllByLeaderboard(bool) ([]*models.User, error)
	GetActive(bool) ([]*models.User, error)
	Count() (int64, error)
	CreateOrGet(*models.Signup, bool) (*models.User, bool, error)
	Update(*models.User) (*models.User, error)
	Delete(*models.User) error
	ResetApiKey(*models.User) (*models.User, error)
	SetWakatimeApiCredentials(*models.User, string, string) (*models.User, error)
	MigrateMd5Password(*models.User, *models.Login) (*models.User, error)
	GenerateResetToken(*models.User) (*models.User, error)
	FlushCache()
}
