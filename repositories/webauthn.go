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

func (r *WebAuthnRepository) Insert(credential *models.WebAuthnCredential) (*models.WebAuthnCredential, error) {
	result := r.db.Create(credential)
	if result.Error != nil {
		return nil, result.Error
	}
	return credential, nil
}

func (r *WebAuthnRepository) GetByUser(userID string) ([]*models.WebAuthnCredential, error) {
	var credentials []*models.WebAuthnCredential
	result := r.db.Where("user_id = ?", userID).Find(&credentials)
	if result.Error != nil {
		return nil, result.Error
	}
	return credentials, nil
}

func (r *WebAuthnRepository) GetByUserAndName(userID string, name string) (*models.WebAuthnCredential, error) {
	var credential models.WebAuthnCredential
	result := r.db.Where("user_id = ? AND name = ?", userID, name).First(&credential)
	if result.Error != nil {
		return nil, result.Error
	}
	return &credential, nil
}

func (r *WebAuthnRepository) Delete(credential *models.WebAuthnCredential) error {
	return r.db.Delete(credential).Error
}

func (r *WebAuthnRepository) Update(credential *models.WebAuthnCredential) error {
	updateMap := map[string]interface{}{
		// no need to update user_id or created_at
		"id":               credential.ID,
		"public_key":       credential.PublicKey,
		"name":             credential.Name,
		"attestation_type": credential.AttestationType,
		"transport":        credential.Transport,
		"flags":            credential.Flags,
		"authenticator":    credential.Authenticator,
		"attestation":      credential.Attestation,
		"last_used_at":     time.Now(),
	}
	result := r.db.Model(credential).Updates(updateMap)
	return result.Error
}
