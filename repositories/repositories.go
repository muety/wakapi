package repositories

import (
	"github.com/muety/wakapi/models"
	"time"
)

type IAliasRepository interface {
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
	InsertBatch([]*models.Heartbeat) error
	GetAll() ([]*models.Heartbeat, error)
	GetAllWithin(time.Time, time.Time, *models.User) ([]*models.Heartbeat, error)
	GetAllWithinByFilters(time.Time, time.Time, *models.User, map[string][]string) ([]*models.Heartbeat, error)
	GetFirstByUsers() ([]*models.TimeByUser, error)
	GetLastByUsers() ([]*models.TimeByUser, error)
	GetLatestByUser(*models.User) (*models.Heartbeat, error)
	GetLatestByOriginAndUser(string, *models.User) (*models.Heartbeat, error)
	Count(bool) (int64, error)
	CountByUser(*models.User) (int64, error)
	CountByUsers([]*models.User) ([]*models.CountByUser, error)
	GetEntitySetByUser(uint8, *models.User) ([]string, error)
	DeleteBefore(time.Time) error
	DeleteByUser(*models.User) error
	DeleteByUserBefore(*models.User, time.Time) error
}

type IDiagnosticsRepository interface {
	Insert(diagnostics *models.Diagnostics) (*models.Diagnostics, error)
}

type IKeyValueRepository interface {
	GetAll() ([]*models.KeyStringValue, error)
	GetString(string) (*models.KeyStringValue, error)
	PutString(*models.KeyStringValue) error
	DeleteString(string) error
	Search(string) ([]*models.KeyStringValue, error)
}

type ILanguageMappingRepository interface {
	GetAll() ([]*models.LanguageMapping, error)
	GetById(uint) (*models.LanguageMapping, error)
	GetByUser(string) ([]*models.LanguageMapping, error)
	Insert(*models.LanguageMapping) (*models.LanguageMapping, error)
	Delete(uint) error
}

type IProjectLabelRepository interface {
	GetAll() ([]*models.ProjectLabel, error)
	GetById(uint) (*models.ProjectLabel, error)
	GetByUser(string) ([]*models.ProjectLabel, error)
	Insert(*models.ProjectLabel) (*models.ProjectLabel, error)
	Delete(uint) error
}

type ISummaryRepository interface {
	Insert(*models.Summary) error
	GetAll() ([]*models.Summary, error)
	GetByUserWithin(*models.User, time.Time, time.Time) ([]*models.Summary, error)
	GetLastByUser() ([]*models.TimeByUser, error)
	DeleteByUser(string) error
	DeleteByUserBefore(string, time.Time) error
}

type IUserRepository interface {
	GetById(string) (*models.User, error)
	GetByIds([]string) ([]*models.User, error)
	GetByApiKey(string) (*models.User, error)
	GetByEmail(string) (*models.User, error)
	GetByResetToken(string) (*models.User, error)
	GetAll() ([]*models.User, error)
	GetMany([]string) ([]*models.User, error)
	GetAllByReports(bool) ([]*models.User, error)
	GetAllByLeaderboard(bool) ([]*models.User, error)
	GetByLoggedInAfter(time.Time) ([]*models.User, error)
	GetByLastActiveAfter(time.Time) ([]*models.User, error)
	Count() (int64, error)
	InsertOrGet(*models.User) (*models.User, bool, error)
	Update(*models.User) (*models.User, error)
	UpdateField(*models.User, string, interface{}) (*models.User, error)
	Delete(*models.User) error
}

type ILeaderboardRepository interface {
	InsertBatch([]*models.LeaderboardItem) error
	CountAllByUser(string) (int64, error)
	CountUsers() (int64, error)
	DeleteByUser(string) error
	DeleteByUserAndInterval(string, *models.IntervalKey) error
	GetAllAggregatedByInterval(*models.IntervalKey, *uint8, int, int) ([]*models.LeaderboardItemRanked, error)
	GetAggregatedByUserAndInterval(string, *models.IntervalKey, *uint8, int, int) ([]*models.LeaderboardItemRanked, error)
}
