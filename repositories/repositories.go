package repositories

import (
	"time"

	"gorm.io/gorm"

	"github.com/muety/wakapi/models"
)

type IBaseRepository interface {
	GetDialector() string
	GetTableDDLMysql(string) (string, error)
	GetTableDDLSqlite(string) (string, error)
	RunInTx(func(*gorm.DB) error) error
	VacuumOrOptimize()
}

type IAliasRepository interface {
	IBaseRepository
	Insert(*models.Alias) (*models.Alias, error)
	Delete(uint) error
	DeleteBatch([]uint) error
	GetAll() ([]*models.Alias, error)
	GetByUser(string) ([]*models.Alias, error)
	GetByUserAndKey(string, string) ([]*models.Alias, error)
	GetByUserAndKeyAndType(string, string, uint8) ([]*models.Alias, error)
	GetByUserAndTypeAndValue(string, uint8, string) (*models.Alias, error)
}

type IHeartbeatRepository interface {
	IBaseRepository
	InsertBatch([]*models.Heartbeat) error
	GetAll() ([]*models.Heartbeat, error)
	GetWithin(time.Time, time.Time, *models.User) ([]*models.Heartbeat, error)
	GetAllWithinByFilters(time.Time, time.Time, *models.User, map[string][]string) ([]*models.Heartbeat, error)
	GetLatestByFilters(*models.User, map[string][]string) (*models.Heartbeat, error)
	GetFirstAll() ([]*models.TimeByUser, error)
	GetLastAll() ([]*models.TimeByUser, error)
	GetRangeByUser(*models.User) (*models.RangeByUser, error)
	GetLatestByUser(*models.User) (*models.Heartbeat, error)
	GetLatestByOriginAndUser(string, *models.User) (*models.Heartbeat, error)
	StreamWithin(time.Time, time.Time, *models.User) (chan *models.Heartbeat, error)
	StreamWithinByFilters(time.Time, time.Time, *models.User, map[string][]string) (chan *models.Heartbeat, error)
	StreamWithinBatched(time.Time, time.Time, *models.User, int) (chan []*models.Heartbeat, error)
	Count(bool) (int64, error)
	CountByUser(*models.User) (int64, error)
	CountByUsers([]*models.User) ([]*models.CountByUser, error)
	GetEntitySetByUser(uint8, string) ([]string, error)
	DeleteBefore(time.Time) error
	DeleteByUser(*models.User) error
	DeleteByUserBefore(*models.User, time.Time) error
	GetUserProjectStats(*models.User, time.Time, time.Time, int, int) ([]*models.ProjectStats, error)
	GetUserAgentsByUser(user *models.User) ([]*models.UserAgent, error)
}

type IDurationRepository interface {
	IBaseRepository
	InsertBatch([]*models.Duration) error
	GetAll() ([]*models.Duration, error)
	GetAllWithin(time.Time, time.Time, *models.User) ([]*models.Duration, error)
	GetAllWithinByFilters(time.Time, time.Time, *models.User, map[string][]string) ([]*models.Duration, error)
	StreamAllBatched(int) (chan []*models.Duration, error)
	StreamByUserBatched(*models.User, int) (chan []*models.Duration, error)
	GetLatestByUser(*models.User) (*models.Duration, error)
	DeleteByUser(*models.User) error
	DeleteByUserBefore(*models.User, time.Time) error
}

type IDiagnosticsRepository interface {
	IBaseRepository
	Insert(diagnostics *models.Diagnostics) (*models.Diagnostics, error)
}

type IKeyValueRepository interface {
	IBaseRepository
	GetAll() ([]*models.KeyStringValue, error)
	GetString(string) (*models.KeyStringValue, error)
	PutString(*models.KeyStringValue) error
	DeleteString(string) error
	DeleteStringTx(string, *gorm.DB) error
	DeleteWildcard(string) error
	DeleteWildcardTx(string, *gorm.DB) error
	Search(string) ([]*models.KeyStringValue, error)
	ReplaceKeySuffix(string, string) error
}

type ILanguageMappingRepository interface {
	IBaseRepository
	GetAll() ([]*models.LanguageMapping, error)
	GetById(uint) (*models.LanguageMapping, error)
	GetByUser(string) ([]*models.LanguageMapping, error)
	Insert(*models.LanguageMapping) (*models.LanguageMapping, error)
	Delete(uint) error
}

type IProjectLabelRepository interface {
	IBaseRepository
	GetAll() ([]*models.ProjectLabel, error)
	GetById(uint) (*models.ProjectLabel, error)
	GetByUser(string) ([]*models.ProjectLabel, error)
	Insert(*models.ProjectLabel) (*models.ProjectLabel, error)
	Delete(uint) error
}

type ISummaryRepository interface {
	IBaseRepository
	Insert(*models.Summary) error
	InsertWithRetry(*models.Summary) error
	GetAll() ([]*models.Summary, error)
	GetByUserWithin(*models.User, time.Time, time.Time) ([]*models.Summary, error)
	GetLastByUser() ([]*models.TimeByUser, error)
	DeleteByUser(string) error
	DeleteByUserBefore(string, time.Time) error
}

type IUserRepository interface {
	IBaseRepository
	FindOne(user models.User) (*models.User, error)
	GetByIds([]string) ([]*models.User, error)
	GetAll() ([]*models.User, error)
	GetMany([]string) ([]*models.User, error)
	GetAllByReports(bool) ([]*models.User, error)
	GetAllByLeaderboard(bool) ([]*models.User, error)
	GetByLoggedInBefore(time.Time) ([]*models.User, error)
	GetByLoggedInAfter(time.Time) ([]*models.User, error)
	GetByLastActiveAfter(time.Time) ([]*models.User, error)
	Count() (int64, error)
	InsertOrGet(*models.User) (*models.User, bool, error)
	Update(*models.User) (*models.User, error)
	UpdateField(*models.User, string, interface{}) (*models.User, error)
	Delete(*models.User) error
	DeleteTx(*models.User, *gorm.DB) error
}

type ILeaderboardRepository interface {
	IBaseRepository
	InsertBatch([]*models.LeaderboardItem) error
	CountAllByUser(string) (int64, error)
	CountUsers(bool) (int64, error)
	DeleteByUser(string) error
	DeleteByUserAndInterval(string, *models.IntervalKey) error
	GetAll() ([]*models.LeaderboardItem, error)
	GetAllAggregatedByInterval(*models.IntervalKey, *uint8, int, int) ([]*models.LeaderboardItemRanked, error)
	GetAggregatedByUserAndInterval(string, *models.IntervalKey, *uint8, int, int) ([]*models.LeaderboardItemRanked, error)
}

type IApiKeyRepository interface {
	IBaseRepository
	GetAll() ([]*models.ApiKey, error)
	GetByUser(string) ([]*models.ApiKey, error)
	GetByApiKey(string, bool) (*models.ApiKey, error)
	Insert(*models.ApiKey) (*models.ApiKey, error)
	Delete(string) error
}
