package services

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
	"github.com/muety/wakapi/utils"
	"github.com/patrickmn/go-cache"
	uuid "github.com/satori/go.uuid"
	"time"
)

type UserService struct {
	Config     *config.Config
	cache      *cache.Cache
	repository repositories.IUserRepository
}

func NewUserService(userRepo repositories.IUserRepository) *UserService {
	return &UserService{
		Config:     config.Get(),
		repository: userRepo,
		cache:      cache.New(1*time.Hour, 2*time.Hour),
	}
}

func (srv *UserService) GetUserById(userId string) (*models.User, error) {
	if u, ok := srv.cache.Get(userId); ok {
		return u.(*models.User), nil
	}

	u, err := srv.repository.GetById(userId)
	if err != nil {
		return nil, err
	}

	srv.cache.Set(u.ID, u, cache.DefaultExpiration)
	return u, nil
}

func (srv *UserService) GetUserByKey(key string) (*models.User, error) {
	if u, ok := srv.cache.Get(key); ok {
		return u.(*models.User), nil
	}

	u, err := srv.repository.GetByApiKey(key)
	if err != nil {
		return nil, err
	}

	srv.cache.Set(u.ID, u, cache.DefaultExpiration)
	return u, nil
}

func (srv *UserService) GetUserByEmail(email string) (*models.User, error) {
	return srv.repository.GetByEmail(email)
}

func (srv *UserService) GetUserByResetToken(resetToken string) (*models.User, error) {
	return srv.repository.GetByResetToken(resetToken)
}

func (srv *UserService) GetAll() ([]*models.User, error) {
	return srv.repository.GetAll()
}

func (srv *UserService) GetAllByReports(reportsEnabled bool) ([]*models.User, error) {
	return srv.repository.GetAllByReports(reportsEnabled)
}

func (srv *UserService) GetActive() ([]*models.User, error) {
	minDate := time.Now().Add(-24 * time.Hour * time.Duration(srv.Config.App.InactiveDays))
	return srv.repository.GetByLastActiveAfter(minDate)
}

func (srv *UserService) Count() (int64, error) {
	return srv.repository.Count()
}

func (srv *UserService) CreateOrGet(signup *models.Signup, isAdmin bool) (*models.User, bool, error) {
	u := &models.User{
		ID:       signup.Username,
		ApiKey:   uuid.NewV4().String(),
		Email:    signup.Email,
		Location: signup.Location,
		Password: signup.Password,
		IsAdmin:  isAdmin,
	}

	if hash, err := utils.HashBcrypt(u.Password, srv.Config.Security.PasswordSalt); err != nil {
		return nil, false, err
	} else {
		u.Password = hash
	}

	return srv.repository.InsertOrGet(u)
}

func (srv *UserService) Update(user *models.User) (*models.User, error) {
	srv.cache.Flush()
	return srv.repository.Update(user)
}

func (srv *UserService) ResetApiKey(user *models.User) (*models.User, error) {
	srv.cache.Flush()
	user.ApiKey = uuid.NewV4().String()
	return srv.Update(user)
}

func (srv *UserService) SetWakatimeApiKey(user *models.User, apiKey string) (*models.User, error) {
	srv.cache.Flush()
	return srv.repository.UpdateField(user, "wakatime_api_key", apiKey)
}

func (srv *UserService) MigrateMd5Password(user *models.User, login *models.Login) (*models.User, error) {
	srv.cache.Flush()
	user.Password = login.Password
	if hash, err := utils.HashBcrypt(user.Password, srv.Config.Security.PasswordSalt); err != nil {
		return nil, err
	} else {
		user.Password = hash
	}
	return srv.repository.UpdateField(user, "password", user.Password)
}

func (srv *UserService) GenerateResetToken(user *models.User) (*models.User, error) {
	return srv.repository.UpdateField(user, "reset_token", uuid.NewV4())
}

func (srv *UserService) Delete(user *models.User) error {
	srv.cache.Flush()
	return srv.repository.Delete(user)
}

func (srv *UserService) FlushCache() {
	srv.cache.Flush()
}
