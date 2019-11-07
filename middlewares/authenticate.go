package middlewares

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/n1try/wakapi/models"
	"github.com/n1try/wakapi/services"
)

type AuthenticateMiddleware struct {
	UserSrvc    *services.UserService
	Cache       *cache.Cache
	Initialized bool
}

func (m *AuthenticateMiddleware) Init() {
	if m.Cache == nil {
		m.Cache = cache.New(1*time.Hour, 2*time.Hour)
	}
	m.Initialized = true
}

func (m *AuthenticateMiddleware) Handle(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if !m.Initialized {
		m.Init()
	}

	var user *models.User
	var userKey string
	user, userKey, err := m.tryGetUserByPassword(r)

	if err != nil {
		user, userKey, err = m.tryGetUserByApiKey(r)
	}

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	m.Cache.Set(userKey, user, cache.DefaultExpiration)

	ctx := context.WithValue(r.Context(), models.UserKey, user)
	next(w, r.WithContext(ctx))
}

func (m *AuthenticateMiddleware) tryGetUserByApiKey(r *http.Request) (*models.User, string, error) {
	authHeader := strings.Split(r.Header.Get("Authorization"), " ")
	if len(authHeader) != 2 || authHeader[0] != "Basic" {
		return nil, "", errors.New("failed to extract API key")
	}

	key, err := base64.StdEncoding.DecodeString(authHeader[1])
	if err != nil {
		return nil, "", err
	}

	var user *models.User
	userKey := strings.TrimSpace(string(key))
	cachedUser, ok := m.Cache.Get(userKey)
	if !ok {
		user, err = m.UserSrvc.GetUserByKey(userKey)
		if err != nil {
			return nil, "", err
		}
	} else {
		user = cachedUser.(*models.User)
	}
	return user, userKey, nil
}

func (m *AuthenticateMiddleware) tryGetUserByPassword(r *http.Request) (*models.User, string, error) {
	authHeader := strings.Split(r.Header.Get("Authorization"), " ")
	if len(authHeader) != 2 || authHeader[0] != "Basic" {
		return nil, "", errors.New("failed to extract API key")
	}

	hash, err := base64.StdEncoding.DecodeString(authHeader[1])
	userKey := strings.TrimSpace(string(hash))
	if err != nil {
		return nil, "", err
	}

	var user *models.User
	cachedUser, ok := m.Cache.Get(userKey)
	if !ok {
		re := regexp.MustCompile(`^(.+):(.+)$`)
		groups := re.FindAllStringSubmatch(userKey, -1)
		if len(groups) == 0 || len(groups[0]) != 3 {
			return nil, "", errors.New("failed to parse user agent string")
		}
		userId, password := groups[0][1], groups[0][2]
		user, err = m.UserSrvc.GetUserById(userId)
		if err != nil {
			return nil, "", err
		}
		passwordHash := md5.Sum([]byte(password))
		passwordHashString := hex.EncodeToString(passwordHash[:])
		if passwordHashString != user.Password {
			return nil, "", errors.New("invalid password")
		}
	} else {
		user = cachedUser.(*models.User)
	}
	return user, userKey, nil
}
