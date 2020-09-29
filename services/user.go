package services

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	uuid "github.com/satori/go.uuid"
)

type UserService struct {
	Config *config.Config
	Db     *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		Config: config.Get(),
		Db:     db,
	}
}

func (srv *UserService) GetUserById(userId string) (*models.User, error) {
	u := &models.User{}
	if err := srv.Db.Where(&models.User{ID: userId}).First(u).Error; err != nil {
		return u, err
	}
	return u, nil
}

func (srv *UserService) GetUserByKey(key string) (*models.User, error) {
	u := &models.User{}
	if err := srv.Db.Where(&models.User{ApiKey: key}).First(u).Error; err != nil {
		return u, err
	}
	return u, nil
}

func (srv *UserService) GetAll() ([]*models.User, error) {
	var users []*models.User
	if err := srv.Db.
		Table("users").
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (srv *UserService) CreateOrGet(signup *models.Signup) (*models.User, bool, error) {
	u := &models.User{
		ID:       signup.Username,
		ApiKey:   uuid.NewV4().String(),
		Password: signup.Password,
	}

	if err := utils.HashPassword(u, srv.Config.PasswordSalt); err != nil {
		return nil, false, err
	}

	result := srv.Db.FirstOrCreate(u, &models.User{ID: u.ID})
	if err := result.Error; err != nil {
		return nil, false, err
	}

	if result.RowsAffected == 1 {
		return u, true, nil
	}

	return u, false, nil
}

func (srv *UserService) Update(user *models.User) (*models.User, error) {
	result := srv.Db.Model(&models.User{}).Updates(user)
	if err := result.Error; err != nil {
		return nil, err
	}

	if result.RowsAffected != 1 {
		return nil, errors.New("nothing updated")
	}

	return user, nil
}

func (srv *UserService) ResetApiKey(user *models.User) (*models.User, error) {
	user.ApiKey = uuid.NewV4().String()
	return srv.Update(user)
}

func (srv *UserService) ToggleBadges(user *models.User) (*models.User, error) {
	result := srv.Db.Model(user).Update("badges_enabled", !user.BadgesEnabled)
	if err := result.Error; err != nil {
		return nil, err
	}

	if result.RowsAffected != 1 {
		return nil, errors.New("nothing updated")
	}

	return user, nil
}

func (srv *UserService) MigrateMd5Password(user *models.User, login *models.Login) (*models.User, error) {
	user.Password = login.Password
	if err := utils.HashPassword(user, srv.Config.PasswordSalt); err != nil {
		return nil, err
	}

	result := srv.Db.Model(user).Update("password", user.Password)
	if err := result.Error; err != nil {
		return nil, err
	} else if result.RowsAffected < 1 {
		return nil, errors.New("nothing changes")
	}

	return user, nil
}
