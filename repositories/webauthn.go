package repositories

import (
	"time"

	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

type WebAuthnRepository struct {
	BaseRepository
}

func NewWebAuthnRepository(db *gorm.DB) *WebAuthnRepository {
	return &WebAuthnRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *WebAuthnRepository) Insert(credential *models.Credential) (*models.Credential, error) {
	result := r.db.Create(credential)
	if result.Error != nil {
		return nil, result.Error
	}
	return credential, nil
}

func (r *WebAuthnRepository) GetByUser(userID string) ([]*models.Credential, error) {
	var credentials []*models.Credential
	result := r.db.Where("user_id = ?", userID).Find(&credentials)
	if result.Error != nil {
		return nil, result.Error
	}
	return credentials, nil
}

func (r *WebAuthnRepository) GetByUserAndName(userID string, name string) (*models.Credential, error) {
	var credential models.Credential
	result := r.db.Where("user_id = ? AND name = ?", userID, name).First(&credential)
	if result.Error != nil {
		return nil, result.Error
	}
	return &credential, nil
}

func (r *WebAuthnRepository) Delete(credential *models.Credential) error {
	result := r.db.Delete(credential)
	return result.Error
}

func (r *WebAuthnRepository) Update(credential *models.Credential) error {
	credential.LastUsedAt = models.CustomTime(time.Now())
	result := r.db.Model(&credential).Updates(credential)
	return result.Error
}
