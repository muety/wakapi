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
	LoadUserAliases(string) error
	GetAliasOrDefault(string, uint8, string) (string, error)
	IsInitialized(string) bool
}

type IHeartbeatService interface {
	InsertBatch([]*models.Heartbeat) error
	GetAllWithin(time.Time, time.Time, *models.User) ([]*models.Heartbeat, error)
	GetFirstByUsers() ([]*models.TimeByUser, error)
	DeleteBefore(time.Time) error
}

type IKeyValueService interface {
	GetString(string) (*models.KeyStringValue, error)
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

type ISummaryService interface {
	Aliased(time.Time, time.Time, *models.User, SummaryRetriever) (*models.Summary, error)
	Retrieve(time.Time, time.Time, *models.User) (*models.Summary, error)
	Summarize(time.Time, time.Time, *models.User) (*models.Summary, error)
	GetLatestByUser() ([]*models.TimeByUser, error)
	DeleteByUser(string) error
	Insert(*models.Summary) error
}

type IUserService interface {
	GetUserById(string) (*models.User, error)
	GetUserByKey(string) (*models.User, error)
	GetAll() ([]*models.User, error)
	CreateOrGet(*models.Signup) (*models.User, bool, error)
	Update(*models.User) (*models.User, error)
	ResetApiKey(*models.User) (*models.User, error)
	ToggleBadges(*models.User) (*models.User, error)
	MigrateMd5Password(*models.User, *models.Login) (*models.User, error)
}
