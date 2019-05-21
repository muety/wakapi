package middlewares

import (
	"context"
	"encoding/base64"
	"net/http"
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

	authHeader := strings.Split(r.Header.Get("Authorization"), " ")
	if len(authHeader) != 2 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	key, err := base64.StdEncoding.DecodeString(authHeader[1])
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var user *models.User
	userKey := strings.TrimSpace(string(key))
	cachedUser, ok := m.Cache.Get(userKey)
	if !ok {
		user, err = m.UserSrvc.GetUserByKey(userKey)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	} else {
		user = cachedUser.(*models.User)
	}

	m.Cache.Set(userKey, user, cache.DefaultExpiration)

	ctx := context.WithValue(r.Context(), models.UserKey, user)
	next(w, r.WithContext(ctx))
}
