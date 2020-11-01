package repositories

import (
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
