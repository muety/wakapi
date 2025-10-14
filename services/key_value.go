package services

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
	"gorm.io/gorm"
)

type KeyValueService struct {
	config     *config.Config
	repository repositories.IKeyValueRepository
}

func NewKeyValueService(keyValueRepo repositories.IKeyValueRepository) *KeyValueService {
	return &KeyValueService{
		config:     config.Get(),
		repository: keyValueRepo,
	}
}

func (srv *KeyValueService) GetString(key string) (*models.KeyStringValue, error) {
	return srv.repository.GetString(key)
}

func (srv *KeyValueService) GetByPrefix(prefix string) ([]*models.KeyStringValue, error) {
	return srv.repository.Search(prefix + "%")
}

func (srv *KeyValueService) MustGetString(key string) *models.KeyStringValue {
	kv, err := srv.repository.GetString(key)
	if err != nil {
		return &models.KeyStringValue{
			Key:   key,
			Value: "",
		}
	}
	return kv
}

func (srv *KeyValueService) PutString(kv *models.KeyStringValue) error {
	return srv.repository.PutString(kv)
}

func (srv *KeyValueService) DeleteString(key string) error {
	return srv.repository.DeleteString(key)
}

func (srv *KeyValueService) DeleteStringTx(key string, tx *gorm.DB) error {
	return srv.repository.DeleteStringTx(key, tx)
}

func (srv *KeyValueService) DeleteWildcard(key string) error {
	return srv.repository.DeleteWildcard(key)
}

func (srv *KeyValueService) DeleteWildcardTx(key string, tx *gorm.DB) error {
	return srv.repository.DeleteWildcardTx(key, tx)
}

func (srv *KeyValueService) ReplaceKeySuffix(suffixOld, suffixNew string) error {
	return srv.repository.ReplaceKeySuffix(suffixOld, suffixNew)
}
