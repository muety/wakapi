package repositories

import (
	"errors"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

type AliasRepository struct {
	db *gorm.DB
}

func NewAliasRepository(db *gorm.DB) *AliasRepository {
	return &AliasRepository{db: db}
}

func (r *AliasRepository) GetByUser(userId string) ([]*models.Alias, error) {
	var aliases []*models.Alias
	if err := r.db.
		Where(&models.Alias{UserID: userId}).
		Find(&aliases).Error; err != nil {
		return nil, err
	}
	return aliases, nil
}

func (r *AliasRepository) GetByUserAndKey(userId, key string) ([]*models.Alias, error) {
	var aliases []*models.Alias
	if err := r.db.
		Where(&models.Alias{
			UserID: userId,
			Key:    key,
		}).
		Find(&aliases).Error; err != nil {
		return nil, err
	}
	return aliases, nil
}

func (r *AliasRepository) GetByUserAndKeyAndType(userId, key string, summaryType uint8) ([]*models.Alias, error) {
	var aliases []*models.Alias
	if err := r.db.
		Where(&models.Alias{
			UserID: userId,
			Key:    key,
			Type:   summaryType,
		}).
		Find(&aliases).Error; err != nil {
		return nil, err
	}
	return aliases, nil
}

func (r *AliasRepository) GetByUserAndTypeAndValue(userId string, summaryType uint8, value string) (*models.Alias, error) {
	alias := &models.Alias{}
	if err := r.db.
		Where(&models.Alias{
			UserID: userId,
			Type:   summaryType,
			Value:  value,
		}).
		First(alias).Error; err != nil {
		return nil, err
	}
	return alias, nil
}

func (r *AliasRepository) Insert(alias *models.Alias) (*models.Alias, error) {
	if !alias.IsValid() {
		return nil, errors.New("invalid alias")
	}
	result := r.db.Create(alias)
	if err := result.Error; err != nil {
		return nil, err
	}
	return alias, nil
}

func (r *AliasRepository) Delete(id uint) error {
	return r.db.
		Where("id = ?", id).
		Delete(models.Alias{}).Error
}

func (r *AliasRepository) DeleteBatch(ids []uint) error {
	return r.db.
		Where("id IN ?", ids).
		Delete(models.Alias{}).Error
}
