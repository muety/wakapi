package repositories

import (
	"errors"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
	"time"
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

func (r *UserRepository) GetByIds(userIds []string) ([]*models.User, error) {
	var users []*models.User
	if err := r.db.
		Model(&models.User{}).
		Where("id in ?", userIds).
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
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
		Where(&models.User{}).
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) GetByLoggedInAfter(t time.Time) ([]*models.User, error) {
	var users []*models.User
	if err := r.db.
		Where("last_logged_in_at >= ?", t).
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// Returns a list of user ids, whose last heartbeat is not older than t
// NOTE: Only ID field will be populated
func (r *UserRepository) GetByLastActiveAfter(t time.Time) ([]*models.User, error) {
	subQuery1 := r.db.Model(&models.User{}).
		Select("users.id as user, max(time) as time").
		Joins("left join heartbeats on users.id = heartbeats.user_id").
		Group("user")

	var userIds []string
	if err := r.db.
		Select("user as id").
		Table("(?) as q", subQuery1).
		Where("time >= ?", t).
		Scan(&userIds).Error; err != nil {
		return nil, err
	}

	return r.GetByIds(userIds)
}

func (r *UserRepository) Count() (int64, error) {
	var count int64
	if err := r.db.
		Model(&models.User{}).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
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
	updateMap := map[string]interface{}{
		"api_key":             user.ApiKey,
		"password":            user.Password,
		"last_logged_in_at":   user.LastLoggedInAt,
		"share_data_max_days": user.ShareDataMaxDays,
		"share_editors":       user.ShareEditors,
		"share_languages":     user.ShareLanguages,
		"share_oss":           user.ShareOSs,
		"share_projects":      user.ShareProjects,
		"share_machines":      user.ShareMachines,
		"wakatime_api_key":    user.WakatimeApiKey,
		"has_data":            user.HasData,
	}

	result := r.db.Model(user).Updates(updateMap)
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

func (r *UserRepository) Delete(user *models.User) error {
	return r.db.Delete(user).Error
}
