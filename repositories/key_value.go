package repositories

import (
	"errors"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

type KeyValueRepository struct {
	db *gorm.DB
}

func NewKeyValueRepository(db *gorm.DB) *KeyValueRepository {
	return &KeyValueRepository{db: db}
}

func (r *KeyValueRepository) GetString(key string) (*models.KeyStringValue, error) {
	kv := &models.KeyStringValue{}
	if err := r.db.
		Where(&models.KeyStringValue{Key: key}).
		First(&kv).Error; err != nil {
		return nil, err
	}

	return kv, nil
}

func (r *KeyValueRepository) PutString(kv *models.KeyStringValue) error {
	result := r.db.
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

func (r *KeyValueRepository) DeleteString(key string) error {
	result := r.db.
		Delete(&models.KeyStringValue{}, &models.KeyStringValue{Key: key})

	if err := result.Error; err != nil {
		return err
	}

	if result.RowsAffected != 1 {
		return errors.New("nothing deleted")
	}

	return nil
}
