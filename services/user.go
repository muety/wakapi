package services

import (
	"github.com/jinzhu/gorm"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	uuid "github.com/satori/go.uuid"
)

type UserService struct {
	Config *models.Config
	Db     *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		Config: models.GetConfig(),
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
