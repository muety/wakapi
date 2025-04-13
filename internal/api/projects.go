package api

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/muety/wakapi/internal/utilities"
	"github.com/muety/wakapi/models"

	"github.com/go-chi/chi/v5"
	"github.com/muety/wakapi/helpers"

	conf "github.com/muety/wakapi/config"
	v1 "github.com/muety/wakapi/models/compat/wakatime/v1"
	"github.com/muety/wakapi/utils"
)

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
func (a *APIv1) GetProjects(w http.ResponseWriter, r *http.Request) {
	user, err := utilities.CheckEffectiveUser(w, r, a.services.Users(), "current")
	if err != nil {
		return // response was already sent by util function
	}

	projects, err := a.loadProjects(user, r.URL.Query().Get("q"), false)
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
func (a *APIv1) GetProject(w http.ResponseWriter, r *http.Request) {
	user, err := utilities.CheckEffectiveUser(w, r, a.services.Users(), "current")
	if err != nil {
		return // response was already sent by util function
	}

	projects, err := a.loadProjects(user, chi.URLParam(r, "id"), true)
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

func (a *APIv1) loadProjects(user *models.User, q string, exact bool) ([]*v1.Project, error) {
	results, err := a.services.Heartbeat().GetUserProjectStats(user, time.Time{}, utils.BeginOfToday(time.Local), nil, false)
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
