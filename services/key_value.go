package services

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
)

type KeyValueService struct {
	Config *config.Config
	Db     *gorm.DB
}

func NewKeyValueService(db *gorm.DB) *KeyValueService {
	return &KeyValueService{
		Config: config.Get(),
		Db:     db,
	}
}

func (srv *KeyValueService) GetString(key string) (*models.KeyStringValue, error) {
	kv := &models.KeyStringValue{}
	if err := srv.Db.
		Where(&models.KeyStringValue{Key: key}).
		First(&kv).Error; err != nil {
		return nil, err
	}

	return kv, nil
}

func (srv *KeyValueService) PutString(kv *models.KeyStringValue) error {
	result := srv.Db.
		Where(&models.KeyStringValue{Key: kv.Key}).
		Assign(kv).
		FirstOrCreate(kv)

	if err := result.Error; err != nil {
		return err
	}

	if result.RowsAffected != 1 {
		return errors.New("nothing updated")
	}

	return nil
}

func (srv *KeyValueService) DeleteString(key string) error {
	result := srv.Db.
		Delete(&models.KeyStringValue{}, &models.KeyStringValue{Key: key})

	if err := result.Error; err != nil {
		return err
	}

	if result.RowsAffected != 1 {
		return errors.New("nothing deleted")
	}

	return nil
}
