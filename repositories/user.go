package repositories

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/muety/wakapi/models"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetById(userId string) (*models.User, error) {
	u := &models.User{}
	if err := r.db.Where(&models.User{ID: userId}).First(u).Error; err != nil {
		return u, err
	}
	return u, nil
}

func (r *UserRepository) GetByApiKey(key string) (*models.User, error) {
	u := &models.User{}
	if err := r.db.Where(&models.User{ApiKey: key}).First(u).Error; err != nil {
		return u, err
	}
	return u, nil
}

func (r *UserRepository) GetAll() ([]*models.User, error) {
	var users []*models.User
	if err := r.db.
		Table("users").
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) InsertOrGet(user *models.User) (*models.User, bool, error) {
	result := r.db.FirstOrCreate(user, &models.User{ID: user.ID})
	if err := result.Error; err != nil {
		return nil, false, err
	}

	if result.RowsAffected == 1 {
		return user, true, nil
	}

	return user, false, nil
}

func (r *UserRepository) Update(user *models.User) (*models.User, error) {
	result := r.db.Model(&models.User{}).Updates(user)
	if err := result.Error; err != nil {
		return nil, err
	}

	if result.RowsAffected != 1 {
		return nil, errors.New("nothing updated")
	}

	return user, nil
}

func (r *UserRepository) UpdateField(user *models.User, key string, value interface{}) (*models.User, error) {
	result := r.db.Model(user).Update(key, value)
	if err := result.Error; err != nil {
		return nil, err
	}

	if result.RowsAffected != 1 {
		return nil, errors.New("nothing updated")
	}

	return user, nil
}
