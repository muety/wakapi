package v1

import (
	"github.com/muety/wakapi/models"
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
		r.Get("/compat/wakatime/v1/users/{user}/projects/{id}", h.GetOne)
	})
}

// @Summary Retrieve and filter the user's projects
// @Description Mimics https://wakatime.com/developers#projects
// @ID get-wakatime-projects
// @Tags wakatime
// @Produce json
// @Param user path string true "User ID to fetch data for (or 'current')"
// @Param q query string false "Query to filter projects by"
// @Security ApiKeyAuth
// @Success 200 {object} v1.ProjectsViewModel
// @Router /compat/wakatime/v1/users/{user}/projects [get]
func (h *ProjectsHandler) Get(w http.ResponseWriter, r *http.Request) {
	user, err := routeutils.CheckEffectiveUser(w, r, h.userSrvc, "current")
	if err != nil {
		return // response was already sent by util function
	}

	projects, err := h.loadProjects(user, r.URL.Query().Get("q"), false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("something went wrong"))
		conf.Log().Request(r).Error("error occurred", "error", err)
		return
	}

	vm := &v1.ProjectsViewModel{Data: projects}
	helpers.RespondJSON(w, r, http.StatusOK, vm)
}

// @Summary Retrieve a single project
// @Description Mimics undocumented endpoint related to https://wakatime.com/developers#projects
// @ID get-wakatime-project
// @Tags wakatime
// @Produce json
// @Param user path string true "User ID to fetch data for (or 'current')"
// @Param id path string true "Project ID to fetch"
// @Security ApiKeyAuth
// @Success 200 {object} v1.ProjectViewModel
// @Router /compat/wakatime/v1/users/{user}/projects/{id} [get]
func (h *ProjectsHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	user, err := routeutils.CheckEffectiveUser(w, r, h.userSrvc, "current")
	if err != nil {
		return // response was already sent by util function
	}

	projects, err := h.loadProjects(user, chi.URLParam(r, "id"), true)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(conf.ErrInternalServerError))
		conf.Log().Request(r).Error("error occurred", "error", err)
		return
	}

	if len(projects) != 1 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(conf.ErrNotFound))
		return
	}

	vm := &v1.ProjectViewModel{Data: projects[0]}
	helpers.RespondJSON(w, r, http.StatusOK, vm)
}

func (h *ProjectsHandler) loadProjects(user *models.User, q string, exact bool) ([]*v1.Project, error) {
	results, err := h.heartbeatSrvc.GetUserProjectStats(user, time.Time{}, utils.BeginOfToday(time.Local), nil, false)
	if err != nil {
		return nil, err
	}

	projects := make([]*v1.Project, 0, len(results))
	for _, p := range results {
		if (exact && p.Project == q) || (!exact && strings.HasPrefix(p.Project, q)) {
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

	return projects, nil
}
