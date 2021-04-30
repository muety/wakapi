package v1

import (
	"github.com/gorilla/mux"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	v1 "github.com/muety/wakapi/models/compat/wakatime/v1"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"net/http"
)

type UsersHandler struct {
	config        *conf.Config
	userSrvc      services.IUserService
	heartbeatSrvc services.IHeartbeatService
}

func NewUsersHandler(userService services.IUserService, heartbeatService services.IHeartbeatService) *UsersHandler {
	return &UsersHandler{
		userSrvc:      userService,
		heartbeatSrvc: heartbeatService,
		config:        conf.Get(),
	}
}

func (h *UsersHandler) RegisterRoutes(router *mux.Router) {
	r := router.PathPrefix("/compat/wakatime/v1/users/{user}").Subrouter()
	r.Use(
		middlewares.NewAuthenticateMiddleware(h.userSrvc).Handler,
	)
	r.Path("").Methods(http.MethodGet).HandlerFunc(h.Get)
}

// @Summary Retrieve the given user
// @Description Mimics https://wakatime.com/developers#users
// @ID get-wakatime-user
// @Tags wakatime
// @Produce json
// @Param user path string true "User ID to fetch (or 'current')"
// @Security ApiKeyAuth
// @Success 200 {object} v1.UserViewModel
// @Router /compat/wakatime/v1/users/{user} [get]
func (h *UsersHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	requestedUser := vars["user"]
	authorizedUser := middlewares.GetPrincipal(r)

	if requestedUser != authorizedUser.ID && requestedUser != "current" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	user := v1.NewFromUser(authorizedUser)
	if hb, err := h.heartbeatSrvc.GetLatestByUser(authorizedUser); err == nil {
		user = user.WithLatestHeartbeat(hb)
	} else {
		conf.Log().Request(r).Error("%v", err)
	}

	utils.RespondJSON(w, r, http.StatusOK, v1.UserViewModel{Data: user})
}
