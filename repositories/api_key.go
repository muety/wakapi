package repositories

import (
	"errors"

	"gorm.io/gorm"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
)

type ApiKeyRepository struct {
	BaseRepository
	config *config.Config
}

func NewApiKeyRepository(db *gorm.DB) *ApiKeyRepository {
	return &ApiKeyRepository{BaseRepository: NewBaseRepository(db), config: config.Get()}
}

func (r *ApiKeyRepository) GetAll() ([]*models.ApiKey, error) {
	var keys []*models.ApiKey
	if err := r.db.Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}

func (r *ApiKeyRepository) GetById(id uint) (*models.ApiKey, error) {
	key := &models.ApiKey{}
	if err := r.db.Where(&models.ApiKey{ID: id}).First(key).Error; err != nil {
		return key, err
	}
	return key, nil
}

func (r *ApiKeyRepository) GetByApiKey(apiKey string) (*models.ApiKey, error) {
	key := &models.ApiKey{}
	if err := r.db.Where(&models.ApiKey{ApiKey: apiKey}).First(key).Error; err != nil {
		return key, err
	}
	return key, nil
}

func (r *ApiKeyRepository) GetByRWApiKey(apiKey string) (*models.ApiKey, error) {
	key := &models.ApiKey{}
	if err := r.db.Where(&models.ApiKey{ApiKey: apiKey, ReadOnly: false}).First(key).Error; err != nil {
		return key, err
	}
	return key, nil
}

func (r *ApiKeyRepository) GetByUser(userId string) ([]*models.ApiKey, error) {
	if userId == "" {
		return []*models.ApiKey{}, nil
	}
	var keys []*models.ApiKey
	if err := r.db.
		Where(&models.ApiKey{UserID: userId}).
		Find(&keys).Error; err != nil {
		return keys, err
	}
	return keys, nil
}

func (r *ApiKeyRepository) Insert(key *models.ApiKey) (*models.ApiKey, error) {
	if !key.IsValid() {
		return nil, errors.New("invalid API key")
	}
	result := r.db.Create(key)
	if err := result.Error; err != nil {
		return nil, err
	}
	return key, nil
}

func (r *ApiKeyRepository) Delete(id uint) error {
	return r.db.
		Where("id = ?", id).
		Delete(models.ApiKey{}).Error
}
