package services

import (
	"time"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
	"github.com/patrickmn/go-cache"
)

type UserOauthService struct {
	config     *config.Config
	cache      *cache.Cache
	repository repositories.IOauthUserRepository
}

func NewUserOauthService(userOauthRepo repositories.IOauthUserRepository) *UserOauthService {
	return &UserOauthService{
		config:     config.Get(),
		cache:      cache.New(1*time.Hour, 2*time.Hour),
		repository: userOauthRepo,
	}
}

func (srv *UserOauthService) Create(newUserOauth *models.UserOauth) (*models.UserOauth, error) {
	return srv.repository.Create(newUserOauth)
}

func (srv *UserOauthService) GetUserOauth(id string) (*models.UserOauth, error) {
	return srv.repository.GetById(id)
}

func (srv *UserOauthService) GetOne(params models.UserOauth) (*models.UserOauth, error) {
	return srv.repository.FindOne(params)
}

func (srv *UserOauthService) DeleteUserOauth(id string) error {
	return srv.repository.DeleteById(id)
}

type IUserOauthService interface {
	Create(newUserOauth *models.UserOauth) (*models.UserOauth, error)
	GetUserOauth(id string) (*models.UserOauth, error)
	DeleteUserOauth(id string) error
	GetOne(params models.UserOauth) (*models.UserOauth, error)
}
