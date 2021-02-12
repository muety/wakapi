package repositories

import (
	"github.com/muety/wakapi/models"
	"time"
)

type IAliasRepository interface {
	Insert(*models.Alias) (*models.Alias, error)
	Delete(uint) error
	DeleteBatch([]uint) error
	GetByUser(string) ([]*models.Alias, error)
	GetByUserAndKey(string, string) ([]*models.Alias, error)
	GetByUserAndKeyAndType(string, string, uint8) ([]*models.Alias, error)
	GetByUserAndTypeAndValue(string, uint8, string) (*models.Alias, error)
}

type IHeartbeatRepository interface {
	InsertBatch([]*models.Heartbeat) error
	Count() (int64, error)
	CountByUser(*models.User) (int64, error)
	GetAllWithin(time.Time, time.Time, *models.User) ([]*models.Heartbeat, error)
	GetFirstByUsers() ([]*models.TimeByUser, error)
	GetLatestByOriginAndUser(string, *models.User) (*models.Heartbeat, error)
	DeleteBefore(time.Time) error
}

type IKeyValueRepository interface {
	GetString(string) (*models.KeyStringValue, error)
	PutString(*models.KeyStringValue) error
	DeleteString(string) error
}

type ILanguageMappingRepository interface {
	GetById(uint) (*models.LanguageMapping, error)
	GetByUser(string) ([]*models.LanguageMapping, error)
	Insert(*models.LanguageMapping) (*models.LanguageMapping, error)
	Delete(uint) error
}

type ISummaryRepository interface {
	Insert(*models.Summary) error
	GetByUserWithin(*models.User, time.Time, time.Time) ([]*models.Summary, error)
	GetLastByUser() ([]*models.TimeByUser, error)
	DeleteByUser(string) error
}

type IUserRepository interface {
	GetById(string) (*models.User, error)
	GetByApiKey(string) (*models.User, error)
	GetAll() ([]*models.User, error)
	Count() (int64, error)
	InsertOrGet(*models.User) (*models.User, bool, error)
	Update(*models.User) (*models.User, error)
	UpdateField(*models.User, string, interface{}) (*models.User, error)
	Delete(*models.User) error
}
