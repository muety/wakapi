package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/services"
	"gorm.io/gorm"
)

type UserAgentsApiHandler struct {
	db               *gorm.DB
	userAgentService services.IPluginUserAgentService
	userSrvc         services.IUserService
}

func NewUserAgentApiHandler(db *gorm.DB, userAgentService services.IPluginUserAgentService, userSrvc services.IUserService) *UserAgentsApiHandler {
	return &UserAgentsApiHandler{
		db:               db,
		userAgentService: userAgentService,
		userSrvc:         userSrvc,
	}
}

func (u *UserAgentsApiHandler) RegisterRoutes(router chi.Router) {
	router.Group(func(r chi.Router) {
		r.Use(middlewares.NewAuthenticateMiddleware(u.userSrvc).Handler)
		r.Get("/compat/wakatime/v1/users/{user}/user-agents", u.FetchUserAgents)
	})
}

func (u *UserAgentsApiHandler) FetchUserAgents(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)

	user_agents, err := u.userAgentService.FetchUserAgents(user.ID)
	if err != nil {
		conf.Log().Request(r).Error("failed to retrieve user agents - %v", err)
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Error fetching user agents. Try later.",
		})
		return
	}

	response := map[string]interface{}{
		"data": user_agents,
	}
	helpers.RespondJSON(w, r, http.StatusCreated, response)
}
