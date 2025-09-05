package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	v1 "github.com/muety/wakapi/models/compat/wakatime/v1"
	routeutils "github.com/muety/wakapi/routes/utils"
	"github.com/muety/wakapi/services"
)

type UserAgentsHandler struct {
	config         *conf.Config
	userSrvc       services.IUserService
	heartbeatsSrvc services.IHeartbeatService
}

func NewUserAgentsHandler(userService services.IUserService, heartbeatsService services.IHeartbeatService) *UserAgentsHandler {
	return &UserAgentsHandler{
		userSrvc:       userService,
		heartbeatsSrvc: heartbeatsService,
		config:         conf.Get(),
	}
}

func (h *UserAgentsHandler) RegisterRoutes(router chi.Router) {
	router.Group(func(r chi.Router) {
		r.Use(middlewares.NewAuthenticateMiddleware(h.userSrvc).Handler)
		r.Get("/compat/wakatime/v1/users/{user}/user_agents", h.Get)
	})
}

// @Summary List of unique user agents for given user.
// @Description Mimics https://wakatime.com/developers#user_agents
// @ID get-wakatime-useragents
// @Tags wakatime
// @Produce json
// @Param user path string true "User ID to fetch data for (or 'current')"
// @Security ApiKeyAuth
// @Success 200 {object} v1.UserAgentsViewModel
// @Router /compat/wakatime/v1/users/{user}/user_agents [get]
func (h *UserAgentsHandler) Get(w http.ResponseWriter, r *http.Request) {
	user, err := routeutils.CheckEffectiveUser(w, r, h.userSrvc, "current")
	if err != nil {
		return // response was already sent by util function
	}

	userAgents, err := h.heartbeatsSrvc.GetUserAgentsByUser(user)
	if err != nil {
		conf.Log().Request(r).Error("failed to get user agents for user", "user", user.ID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("something went wrong"))
		return
	}

	vm := h.buildViewModel(userAgents)
	helpers.RespondJSON(w, r, http.StatusOK, vm)
}

func (h *UserAgentsHandler) buildViewModel(userAgents []*models.UserAgent) *v1.UserAgentsViewModel {
	vm := &v1.UserAgentsViewModel{
		Data:       make([]*v1.UserAgentEntry, len(userAgents)),
		TotalPages: 1,
	}

	for i, ua := range userAgents {
		vm.Data[i] = (&v1.UserAgentEntry{}).FromModel(ua)
	}

	return vm
}
