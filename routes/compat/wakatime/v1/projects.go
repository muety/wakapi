package v1

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/muety/wakapi/helpers"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	v1 "github.com/muety/wakapi/models/compat/wakatime/v1"
	routeutils "github.com/muety/wakapi/routes/utils"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
)

type ProjectsHandler struct {
	config        *conf.Config
	userSrvc      services.IUserService
	heartbeatSrvc services.IHeartbeatService
}

func NewProjectsHandler(userService services.IUserService, heartbeatsService services.IHeartbeatService) *ProjectsHandler {
	return &ProjectsHandler{
		userSrvc:      userService,
		heartbeatSrvc: heartbeatsService,
		config:        conf.Get(),
	}
}

func (h *ProjectsHandler) RegisterRoutes(router chi.Router) {
	router.Group(func(r chi.Router) {
		r.Use(middlewares.NewAuthenticateMiddleware(h.userSrvc).Handler)
		r.Get("/compat/wakatime/v1/users/{user}/projects", h.Get)
	})
}

// @Summary Retrieve and fitler the user's projects
// @Description Mimics https://wakatime.com/developers#projects
// @ID get-wakatime-projects
// @Tags wakatime
// @Produce json
// @Param user path string true "User ID to fetch data for (or 'current')"
// @Param q query string true "Query to filter projects by"
// @Security ApiKeyAuth
// @Success 200 {object} v1.ProjectsViewModel
// @Router /compat/wakatime/v1/users/{user}/projects [get]
func (h *ProjectsHandler) Get(w http.ResponseWriter, r *http.Request) {
	user, err := routeutils.CheckEffectiveUser(w, r, h.userSrvc, "current")
	if err != nil {
		return // response was already sent by util function
	}

	results, err := h.heartbeatSrvc.GetUserProjectStats(user, time.Time{}, utils.BeginOfToday(time.Local), nil, false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("something went wrong"))
		conf.Log().Request(r).Error(err.Error())
		return
	}

	q := r.URL.Query().Get("q")

	projects := make([]*v1.Project, 0, len(results))
	for _, p := range results {
		if strings.HasPrefix(p.Project, q) {
			projects = append(projects, &v1.Project{
				ID:                           p.Project,
				Name:                         p.Project,
				LastHeartbeatAt:              p.Last.T(),
				HumanReadableLastHeartbeatAt: helpers.FormatDateTimeHuman(p.Last.T()),
				UrlencodedName:               url.QueryEscape(p.Project),
				CreatedAt:                    p.First.T(),
			})
		}
	}

	vm := &v1.ProjectsViewModel{Data: projects}
	helpers.RespondJSON(w, r, http.StatusOK, vm)
}
