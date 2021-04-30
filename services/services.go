package services

import (
	"github.com/muety/wakapi/models"
	"time"
)

type IAggregationService interface {
	Schedule()
	Run(map[string]bool) error
}

type IMiscService interface {
	ScheduleCountTotalTime()
}

type IAliasService interface {
	Create(*models.Alias) (*models.Alias, error)
	Delete(*models.Alias) error
	DeleteMulti([]*models.Alias) error
	IsInitialized(string) bool
	InitializeUser(string) error
	GetByUser(string) ([]*models.Alias, error)
	GetByUserAndKeyAndType(string, string, uint8) ([]*models.Alias, error)
	GetAliasOrDefault(string, uint8, string) (string, error)
}

type IHeartbeatService interface {
	Insert(*models.Heartbeat) error
	InsertBatch([]*models.Heartbeat) error
	Count() (int64, error)
	CountByUser(*models.User) (int64, error)
	CountByUsers([]*models.User) ([]*models.CountByUser, error)
	GetAllWithin(time.Time, time.Time, *models.User) ([]*models.Heartbeat, error)
	GetFirstByUsers() ([]*models.TimeByUser, error)
	GetLatestByUser(*models.User) (*models.Heartbeat, error)
	GetLatestByOriginAndUser(string, *models.User) (*models.Heartbeat, error)
	DeleteBefore(time.Time) error
}

type IKeyValueService interface {
	GetString(string) (*models.KeyStringValue, error)
	MustGetString(string) *models.KeyStringValue
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

type IMailService interface {
	SendPasswordReset(*models.User, string) error
	SendImportNotification(*models.User, time.Duration, int) error
	SendReport(*models.User, *models.Report) error
}

type ISummaryService interface {
	Aliased(time.Time, time.Time, *models.User, SummaryRetriever, bool) (*models.Summary, error)
	Retrieve(time.Time, time.Time, *models.User) (*models.Summary, error)
	Summarize(time.Time, time.Time, *models.User) (*models.Summary, error)
	GetLatestByUser() ([]*models.TimeByUser, error)
	DeleteByUser(string) error
	Insert(*models.Summary) error
}

type IReportService interface {
	Schedule()
	SyncSchedule(user *models.User) bool
	Run(*models.User, time.Duration) error
}

type IUserService interface {
	GetUserById(string) (*models.User, error)
	GetUserByKey(string) (*models.User, error)
	GetUserByEmail(string) (*models.User, error)
	GetUserByResetToken(string) (*models.User, error)
	GetAll() ([]*models.User, error)
	GetAllByReports(bool) ([]*models.User, error)
	GetActive() ([]*models.User, error)
	Count() (int64, error)
	CreateOrGet(*models.Signup, bool) (*models.User, bool, error)
	Update(*models.User) (*models.User, error)
	Delete(*models.User) error
	ResetApiKey(*models.User) (*models.User, error)
	SetWakatimeApiKey(*models.User, string) (*models.User, error)
	MigrateMd5Password(*models.User, *models.Login) (*models.User, error)
	GenerateResetToken(*models.User) (*models.User, error)
	FlushCache()
}
