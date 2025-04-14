package services

import (
	"time"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"
)

type UserOauthService struct {
	config *config.Config
	cache  *cache.Cache
	db     *gorm.DB
}

func NewUserOauthService(db *gorm.DB) *UserOauthService {
	return &UserOauthService{
		config: config.Get(),
		cache:  cache.New(1*time.Hour, 2*time.Hour),
		db:     db,
	}
}

func (srv *UserOauthService) Create(newUserOauth *models.UserOauth) (*models.UserOauth, error) {
	result := srv.db.Create(newUserOauth)
	if err := result.Error; err != nil {
		return nil, err
	}
	return newUserOauth, nil
}

func (srv *UserOauthService) GetUserOauth(userOauthID string) (*models.UserOauth, error) {
	g := &models.UserOauth{}

	err := srv.db.Where(models.UserOauth{ID: userOauthID}).First(g).Error
	if err != nil {
		return g, err
	}

	if g.ID != "" {
		return g, nil
	}
	return nil, err
}

func (srv *UserOauthService) GetOne(params models.UserOauth) (*models.UserOauth, error) {
	u := &models.UserOauth{}
	result := srv.db.Where(&params).First(u)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // No record found
		}
		return nil, result.Error
	}
	return u, nil
}

func (srv *UserOauthService) DeleteUserOauth(userOauthId string) error {
	if err := srv.db.
		Where("id = ?", userOauthId).
		Delete(models.UserOauth{}).Error; err != nil {
		return err
	}
	return nil
}

type IUserOauthService interface {
	Create(newUserOauth *models.UserOauth) (*models.UserOauth, error)
	GetUserOauth(id string) (*models.UserOauth, error)
	DeleteUserOauth(id string) error
	GetOne(params models.UserOauth) (*models.UserOauth, error)
}
