package services

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/jinzhu/gorm"
	"github.com/muety/wakapi/models"
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
	pw := md5.Sum([]byte(signup.Password))
	pwString := hex.EncodeToString(pw[:])
	apiKey := uuid.NewV4().String()

	u := &models.User{
		ID:       signup.Username,
		ApiKey:   apiKey,
		Password: pwString,
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
