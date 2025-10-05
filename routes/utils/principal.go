package utils

import (
	"net/http"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
)

func SetPrincipal(r *http.Request, user *models.User) {
	if p := r.Context().Value(config.KeySharedData).(*config.SharedData); p != nil {
		p.Set(config.MiddlewareKeyPrincipal, user)
		p.Set(config.MiddlewareKeyPrincipalId, user.Identity())
	}
}

func GetPrincipal(r *http.Request) *models.User {
	sharedData := r.Context().Value(config.KeySharedData)
	if sharedData == nil {
		return nil
	}
	if p := sharedData.(*config.SharedData); p != nil {
		val := p.MustGet(config.MiddlewareKeyPrincipal)
		if val == nil {
			return nil
		}
		return val.(*models.User)
	}
	return nil
}
