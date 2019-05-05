package middlewares

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/n1try/wakapi/models"
	"github.com/n1try/wakapi/services"
)

type AuthenticateMiddleware struct {
	UserSrvc *services.UserService
}

func (m *AuthenticateMiddleware) Handle(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	authHeader := strings.Split(r.Header.Get("Authorization"), " ")
	if len(authHeader) != 2 {
		w.WriteHeader(401)
		return
	}

	key, err := base64.StdEncoding.DecodeString(authHeader[1])
	if err != nil {
		w.WriteHeader(401)
		return
	}

	user, err := m.UserSrvc.GetUserByKey(strings.TrimSpace(string(key)))
	if err != nil {
		w.WriteHeader(401)
		return
	}

	ctx := context.WithValue(r.Context(), models.UserKey, &user)
	next(w, r.WithContext(ctx))
}
