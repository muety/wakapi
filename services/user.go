package services

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/datetime"
	"github.com/duke-git/lancet/v2/validator"
	"github.com/gofrs/uuid/v5"
	"github.com/leandro-lugaresi/hub"
	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
	"github.com/muety/wakapi/utils"
)

type UserService struct {
	config              *config.Config
	cache               *cache.Cache
	eventBus            *hub.Hub
	keyValueService     IKeyValueService
	mailService         IMailService
	apiKeyService       IApiKeyService
	repository          repositories.IUserRepository
	currentOnlineUsers  *cache.Cache
	countersInitialized atomic.Bool
}

func NewUserService(keyValueService IKeyValueService, mailService IMailService, apiKeyService IApiKeyService, userRepo repositories.IUserRepository) *UserService {
	srv := &UserService{
		config:             config.Get(),
		eventBus:           config.EventBus(),
		cache:              cache.New(1*time.Hour, 2*time.Hour),
		keyValueService:    keyValueService,
		apiKeyService:      apiKeyService,
		mailService:        mailService,
		repository:         userRepo,
		currentOnlineUsers: cache.New(models.DefaultHeartbeatsTimeout, 1*time.Minute),
	}

	sub1 := srv.eventBus.Subscribe(0, config.EventWakatimeFailure)
	go func(sub *hub.Subscription) {
		for m := range sub.Receiver {
			user := m.Fields[config.FieldUser].(*models.User)
			n := m.Fields[config.FieldPayload].(int)

			slog.Warn("resetting wakatime api key for user due to too many failures", "userID", user.ID, "failureCount", n)

			if _, err := srv.SetWakatimeApiCredentials(user, "", ""); err != nil {
				config.Log().Error("failed to set wakatime api key for user", "userID", user.ID)
			}

			if user.Email != "" {
				if err := mailService.SendWakatimeFailureNotification(user, n); err != nil {
					config.Log().Error("failed to send wakatime failure notification mail to user", "userID", user.ID)
				} else {
					slog.Info("sent wakatime connection failure mail", "userID", user.ID)
				}
			}
		}
	}(&sub1)

	sub2 := srv.eventBus.Subscribe(0, config.EventHeartbeatCreate)
	go func(sub *hub.Subscription) {
		for m := range sub.Receiver {
			heartbeat := m.Fields[config.FieldPayload].(*models.Heartbeat)
			if time.Now().Sub(heartbeat.Time.T()) > models.DefaultHeartbeatsTimeout {
				continue
			}
			srv.currentOnlineUsers.SetDefault(heartbeat.UserID, true)
		}
	}(&sub2)

	return srv
}

func (srv *UserService) GetUserById(userId string) (*models.User, error) {
	if userId == "" {
		return nil, errors.New("user id must not be empty")
	}

	if u, ok := srv.cache.Get(userId); ok {
		return u.(*models.User), nil
	}

	u, err := srv.repository.FindOne(models.User{ID: userId})
	if err != nil {
		return nil, err
	}

	srv.cache.SetDefault(u.ID, u)
	return u, nil
}

func (srv *UserService) GetUserByKey(key string, requireFullAccessKey bool) (*models.User, error) {
	if key == "" {
		return nil, errors.New("key must not be empty")
	}

	if u, ok := srv.cache.Get(key); ok {
		return u.(*models.User), nil
	}

	u, err := srv.repository.FindOne(models.User{ApiKey: key})
	if err == nil {
		srv.cache.SetDefault(u.ID, u)
		return u, nil
	}

	apiKey, err := srv.apiKeyService.GetByApiKey(key, requireFullAccessKey)
	if err == nil {
		srv.cache.SetDefault(apiKey.User.ID, apiKey.User)
		return apiKey.User, nil
	}

	return nil, err
}

func (srv *UserService) GetUserByEmail(email string) (*models.User, error) {
	if email == "" {
		return nil, errors.New("email must not be empty")
	}
	if !validator.IsEmail(email) {
		return nil, errors.New("not a valid email")
	}
	return srv.repository.FindOne(models.User{Email: email})
}

func (srv *UserService) GetUserByResetToken(resetToken string) (*models.User, error) {
	if resetToken == "" {
		return nil, errors.New("reset token must not be empty")
	}
	return srv.repository.FindOne(models.User{ResetToken: resetToken})
}

func (srv *UserService) GetUserByUnsubscribeToken(unsubscribeToken string) (*models.User, error) {
	if unsubscribeToken == "" {
		return nil, errors.New("unsubscribe token must not be empty")
	}
	return srv.repository.FindOne(models.User{UnsubscribeToken: unsubscribeToken})
}

func (srv *UserService) GetUserByStripeCustomerId(customerId string) (*models.User, error) {
	if customerId == "" {
		return nil, errors.New("customer id must not be empty")
	}
	return srv.repository.FindOne(models.User{StripeCustomerId: customerId})
}

func (srv *UserService) GetUserByOidc(provider, sub string) (*models.User, error) {
	if sub == "" || provider == "" {
		return nil, errors.New("sub and provider must not be empty")
	}
	return srv.repository.FindOne(models.User{
		Sub:      sub,
		AuthType: provider,
	})
}

func (srv *UserService) GetAll() ([]*models.User, error) {
	return srv.repository.GetAll()
}

func (srv *UserService) GetAllMapped() (map[string]*models.User, error) {
	users, err := srv.repository.GetAll()
	if err != nil {
		return nil, err
	}
	return srv.MapUsersById(users), nil
}

func (srv *UserService) GetMany(ids []string) ([]*models.User, error) {
	return srv.repository.GetMany(ids)
}

func (srv *UserService) GetManyMapped(ids []string) (map[string]*models.User, error) {
	users, err := srv.repository.GetMany(ids)
	if err != nil {
		return nil, err
	}
	return srv.MapUsersById(users), nil
}

func (srv *UserService) GetAllByReports(reportsEnabled bool) ([]*models.User, error) {
	return srv.repository.GetAllByReports(reportsEnabled)
}

func (srv *UserService) GetAllByLeaderboard(leaderboardEnabled bool) ([]*models.User, error) {
	return srv.repository.GetAllByLeaderboard(leaderboardEnabled)
}

func (srv *UserService) GetActive(exact bool) ([]*models.User, error) {
	minDate := time.Now().AddDate(0, 0, -1*srv.config.App.InactiveDays)
	if !exact {
		minDate = datetime.BeginOfHour(minDate)
	}

	cacheKey := fmt.Sprintf("%s--active", minDate.String())
	if u, ok := srv.cache.Get(cacheKey); ok {
		return u.([]*models.User), nil
	}

	results, err := srv.repository.GetByLastActiveAfter(minDate)
	if err != nil {
		return nil, err
	}

	srv.cache.SetDefault(cacheKey, results)
	return results, nil
}

func (srv *UserService) Count() (int64, error) {
	return srv.repository.Count()
}

func (srv *UserService) CountCurrentlyOnline() (int, error) {
	if !srv.countersInitialized.Load() {
		minDate := time.Now().Add(-1 * models.DefaultHeartbeatsTimeout)
		result, err := srv.repository.GetByLastActiveAfter(minDate)
		if err != nil {
			return 0, err
		}
		for _, r := range result {
			srv.currentOnlineUsers.SetDefault(r.ID, true)
		}
		srv.countersInitialized.Store(true)
	}

	return srv.currentOnlineUsers.ItemCount(), nil
}

func (srv *UserService) CreateOrGet(signup *models.Signup, isAdmin bool) (*models.User, bool, error) {
	u := &models.User{
		ID:        signup.Username,
		ApiKey:    uuid.Must(uuid.NewV4()).String(),
		Email:     signup.Email,
		Location:  signup.Location,
		Password:  signup.Password,
		IsAdmin:   isAdmin,
		InvitedBy: signup.InvitedBy,
		AuthType:  signup.OidcProvider, // empty for local auth, will fallback to column default value
		Sub:       signup.OidcSubject,
	}

	if hash, err := utils.HashPassword(u.Password, srv.config.Security.PasswordSalt); err != nil {
		return nil, false, err
	} else {
		u.Password = hash
	}

	return srv.repository.InsertOrGet(u)
}

func (srv *UserService) Update(user *models.User) (*models.User, error) {
	srv.FlushUserCache(user.ID)
	srv.notifyUpdate(user)
	return srv.repository.Update(user)
}

func (srv *UserService) ChangeUserId(user *models.User, newUserId string) (*models.User, error) {
	if !srv.checkUpdateCascade() {
		return nil, errors.New("sqlite database too old to perform user id change consistently")
	}

	// https://github.com/muety/wakapi/issues/739
	oldUserId := user.ID
	defer srv.FlushUserCache(oldUserId)

	// TODO: make this transactional somehow
	userNew, err := srv.repository.UpdateField(user, "id", newUserId)
	if err != nil {
		return nil, err
	}

	err = srv.keyValueService.ReplaceKeySuffix(fmt.Sprintf("_%s", oldUserId), fmt.Sprintf("_%s", newUserId))
	if err != nil {
		// try roll back "manually"
		config.Log().Error("failed to update key string values during user id change, trying to roll back manually", "userID", oldUserId, "newUserID", newUserId)
		if _, err := srv.repository.UpdateField(userNew, "id", oldUserId); err != nil {
			config.Log().Error("manual user id rollback failed", "userID", oldUserId, "newUserID", newUserId)
		}
		return nil, err
	}

	config.Log().Info("user changed their user id", "userID", oldUserId, "newUserID", newUserId)

	return userNew, err
}

func (srv *UserService) ResetApiKey(user *models.User) (*models.User, error) {
	srv.FlushUserCache(user.ID)
	user.ApiKey = uuid.Must(uuid.NewV4()).String()
	return srv.Update(user)
}

func (srv *UserService) SetWakatimeApiCredentials(user *models.User, apiKey string, apiUrl string) (*models.User, error) {
	srv.FlushUserCache(user.ID)

	if apiKey != user.WakatimeApiKey {
		if u, err := srv.repository.UpdateField(user, "wakatime_api_key", apiKey); err != nil {
			return u, err
		}
	}

	if apiUrl != user.WakatimeApiUrl {
		return srv.repository.UpdateField(user, "wakatime_api_url", apiUrl)
	}

	return user, nil
}

func (srv *UserService) GenerateResetToken(user *models.User) (*models.User, error) {
	return srv.repository.UpdateField(user, "reset_token", uuid.Must(uuid.NewV4()))
}

func (srv *UserService) GenerateUnsubscribeToken(user *models.User) (*models.User, error) {
	return srv.repository.UpdateField(user, "unsubscribe_token", uuid.Must(uuid.NewV4()))
}

func (srv *UserService) Delete(user *models.User) error {
	srv.FlushUserCache(user.ID)

	user.ReportsWeekly = false
	srv.notifyUpdate(user)

	return srv.repository.RunInTx(func(tx *gorm.DB) error {
		if err := srv.repository.DeleteTx(user, tx); err != nil {
			return err
		}
		if err := srv.keyValueService.DeleteWildcardTx(fmt.Sprintf("*_%s", user.ID), tx); err != nil {
			return err
		}

		srv.notifyDelete(user)
		return nil
	})
}

func (srv *UserService) MapUsersById(users []*models.User) map[string]*models.User {
	return convertor.ToMap[*models.User, string, *models.User](users, func(u *models.User) (string, *models.User) {
		return u.ID, u
	})
}

func (srv *UserService) FlushCache() {
	srv.cache.Flush()
}

func (srv *UserService) FlushUserCache(userId string) {
	srv.cache.Delete(userId)
}

func (srv *UserService) notifyUpdate(user *models.User) {
	srv.eventBus.Publish(hub.Message{
		Name:   config.EventUserUpdate,
		Fields: map[string]interface{}{config.FieldPayload: user},
	})
}

func (srv *UserService) notifyDelete(user *models.User) {
	srv.eventBus.Publish(hub.Message{
		Name:   config.EventUserDelete,
		Fields: map[string]interface{}{config.FieldPayload: user},
	})
}

func (srv *UserService) checkUpdateCascade() bool {
	if dialector := srv.repository.GetDialector(); dialector == "sqlite" || dialector == "sqlite3" {
		ddl, _ := srv.repository.GetTableDDLSqlite("heartbeats")
		return strings.Contains(ddl, "ON UPDATE CASCADE")
	}
	return true
}
