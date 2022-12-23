package repositories

import (
	"errors"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type KeyValueRepository struct {
	db *gorm.DB
}

func NewKeyValueRepository(db *gorm.DB) *KeyValueRepository {
	return &KeyValueRepository{db: db}
}

func (r *KeyValueRepository) GetAll() ([]*models.KeyStringValue, error) {
	var keyValues []*models.KeyStringValue
	if err := r.db.Find(&keyValues).Error; err != nil {
		return nil, err
	}
	return keyValues, nil
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

func (r *KeyValueRepository) Search(like string) ([]*models.KeyStringValue, error) {
	var keyValues []*models.KeyStringValue
	if err := r.db.Table("key_string_values").
		Where("`key` like ?", like).
		Find(&keyValues).
		Error; err != nil {
		return nil, err
	}
	return keyValues, nil
}

func (r *KeyValueRepository) PutString(kv *models.KeyStringValue) error {
	result := r.db.
		Clauses(clause.OnConflict{
			UpdateAll: true,
		}).
		Where(&models.KeyStringValue{Key: kv.Key}).
		Assign(kv).
		Create(kv)

	if err := result.Error; err != nil {
		return err
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
