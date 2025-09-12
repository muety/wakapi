package middlewares

import (
	"github.com/muety/wakapi/models"
	routeutils "github.com/muety/wakapi/routes/utils"

	"net/http"
)

func SetPrincipal(r *http.Request, user *models.User) {
	routeutils.SetPrincipal(r, user)
}

func GetPrincipal(r *http.Request) *models.User {
	return routeutils.GetPrincipal(r)
}
