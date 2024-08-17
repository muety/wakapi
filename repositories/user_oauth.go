package repositories

import (
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

type UserOauthRepository struct {
	db *gorm.DB
}

func NewUserOauthRepository(db *gorm.DB) *UserOauthRepository {
	return &UserOauthRepository{db: db}
}

func (r *UserOauthRepository) DeleteById(userOauthId string) error {
	if err := r.db.
		Where("id = ?", userOauthId).
		Delete(models.UserOauth{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *UserOauthRepository) FindOne(attributes models.UserOauth) (*models.UserOauth, error) {
	u := &models.UserOauth{}
	result := r.db.Where(&attributes).First(u)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // No record found
		}
		return nil, result.Error
	}
	return u, nil
}

func (r *UserOauthRepository) GetById(userOauthID string) (*models.UserOauth, error) {
	g := &models.UserOauth{}

	err := r.db.Where(models.UserOauth{ID: userOauthID}).First(g).Error
	if err != nil {
		return g, err
	}

	if g.ID != "" {
		return g, nil
	}
	return nil, err
}

func (r *UserOauthRepository) Create(userOauth *models.UserOauth) (*models.UserOauth, error) {
	result := r.db.Create(userOauth)
	if err := result.Error; err != nil {
		return nil, err
	}
	return userOauth, nil
}
