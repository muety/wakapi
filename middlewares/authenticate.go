package middlewares

import (
	"context"
	"errors"
	"fmt"
	"github.com/muety/wakapi/utils"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/services"
)

type AuthenticateMiddleware struct {
	config         *models.Config
	userSrvc       *services.UserService
	cache          *cache.Cache
	whitelistPaths []string
}

func NewAuthenticateMiddleware(userService *services.UserService, whitelistPaths []string) *AuthenticateMiddleware {
	return &AuthenticateMiddleware{
		config:         models.GetConfig(),
		userSrvc:       userService,
		cache:          cache.New(1*time.Hour, 2*time.Hour),
		whitelistPaths: whitelistPaths,
	}
}

func (m *AuthenticateMiddleware) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.ServeHTTP(w, r, h.ServeHTTP)
	})
}

func (m *AuthenticateMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	for _, p := range m.whitelistPaths {
		if strings.HasPrefix(r.URL.Path, p) || r.URL.Path == p {
			next(w, r)
			return
		}
	}

	var user *models.User
	user, err := m.tryGetUserByCookie(r)

	if err != nil {
		user, err = m.tryGetUserByApiKey(r)
	}

	if err != nil {
		if strings.HasPrefix(r.URL.Path, "/api") {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			utils.ClearCookie(w, models.AuthCookieKey, !m.config.InsecureCookies)
			http.Redirect(w, r, fmt.Sprintf("%s/?error=unauthorized", m.config.BasePath), http.StatusFound)
		}
		return
	}

	m.cache.Set(user.ID, user, cache.DefaultExpiration)

	ctx := context.WithValue(r.Context(), models.UserKey, user)
	next(w, r.WithContext(ctx))
}

func (m *AuthenticateMiddleware) tryGetUserByApiKey(r *http.Request) (*models.User, error) {
	key, err := utils.ExtractBearerAuth(r)
	if err != nil {
		return nil, err
	}

	var user *models.User
	userKey := strings.TrimSpace(key)
	cachedUser, ok := m.cache.Get(userKey)
	if !ok {
		user, err = m.userSrvc.GetUserByKey(userKey)
		if err != nil {
			return nil, err
		}
	} else {
		user = cachedUser.(*models.User)
	}
	return user, nil
}

func (m *AuthenticateMiddleware) tryGetUserByCookie(r *http.Request) (*models.User, error) {
	login, err := utils.ExtractCookieAuth(r, m.config)
	if err != nil {
		return nil, err
	}

	cachedUser, ok := m.cache.Get(login.Username)

	if ok {
		return cachedUser.(*models.User), nil
	}

	user, err := m.userSrvc.GetUserById(login.Username)
	if err != nil {
		return nil, err
	}

	if !CheckAndMigratePassword(user, login, m.config.PasswordSalt, m.userSrvc) {
		return nil, errors.New("invalid password")
	}

	return user, nil
}

// migrate old md5-hashed passwords to new salted bcrypt hashes for backwards compatibility
func CheckAndMigratePassword(user *models.User, login *models.Login, salt string, userServiceRef *services.UserService) bool {
	if utils.IsMd5(user.Password) {
		if utils.CheckPasswordMd5(user, login.Password) {
			log.Printf("migrating old md5 password to new bcrypt format for user '%s'", user.ID)
			userServiceRef.MigrateMd5Password(user, login)
			return true
		}
		return false
	}
	return utils.CheckPasswordBcrypt(user, login.Password, salt)
}
