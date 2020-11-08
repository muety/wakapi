package services

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
	"github.com/muety/wakapi/utils"
	uuid "github.com/satori/go.uuid"
)

type UserService struct {
	Config     *config.Config
	repository repositories.IUserRepository
}

func NewUserService(userRepo repositories.IUserRepository) *UserService {
	return &UserService{
		Config:     config.Get(),
		repository: userRepo,
	}
}

func (srv *UserService) GetUserById(userId string) (*models.User, error) {
	return srv.repository.GetById(userId)
}

func (srv *UserService) GetUserByKey(key string) (*models.User, error) {
	return srv.repository.GetByApiKey(key)
}

func (srv *UserService) GetAll() ([]*models.User, error) {
	return srv.repository.GetAll()
}

func (srv *UserService) CreateOrGet(signup *models.Signup) (*models.User, bool, error) {
	u := &models.User{
		ID:       signup.Username,
		ApiKey:   uuid.NewV4().String(),
		Password: signup.Password,
	}

	if hash, err := utils.HashBcrypt(u.Password, srv.Config.Security.PasswordSalt); err != nil {
		return nil, false, err
	} else {
		u.Password = hash
	}

	return srv.repository.InsertOrGet(u)
}

func (srv *UserService) Update(user *models.User) (*models.User, error) {
	return srv.repository.Update(user)
}

func (srv *UserService) ResetApiKey(user *models.User) (*models.User, error) {
	user.ApiKey = uuid.NewV4().String()
	return srv.Update(user)
}

func (srv *UserService) ToggleBadges(user *models.User) (*models.User, error) {
	return srv.repository.UpdateField(user, "badges_enabled", !user.BadgesEnabled)
}

func (srv *UserService) MigrateMd5Password(user *models.User, login *models.Login) (*models.User, error) {
	user.Password = login.Password
	if hash, err := utils.HashBcrypt(user.Password, srv.Config.Security.PasswordSalt); err != nil {
		return nil, err
	} else {
		user.Password = hash
	}
	return srv.repository.UpdateField(user, "password", user.Password)
}
