package routes

import (
	"github.com/go-chi/chi/v5"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/models/view"
	routeutils "github.com/muety/wakapi/routes/utils"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"net/http"
	"time"
)

type ProjectsHandler struct {
	config           *conf.Config
	userService      services.IUserService
	heartbeatService services.IHeartbeatService
}

func NewProjectsHandler(userService services.IUserService, heartbeatService services.IHeartbeatService) *ProjectsHandler {
	return &ProjectsHandler{
		config:           conf.Get(),
		userService:      userService,
		heartbeatService: heartbeatService,
	}
}

func (h *ProjectsHandler) RegisterRoutes(router chi.Router) {
	r := chi.NewRouter()
	r.Use(
		middlewares.NewAuthenticateMiddleware(h.userService).
			WithRedirectTarget(defaultErrorRedirectTarget()).
			WithRedirectErrorMessage("unauthorized").Handler,
	)
	r.Get("/", h.GetIndex)

	router.Mount("/projects", r)
}

func (h *ProjectsHandler) GetIndex(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}
	if err := templates[conf.ProjectsTemplate].Execute(w, h.buildViewModel(r, w)); err != nil {
		conf.Log().Request(r).Error("failed to get projects page - %v", err)
	}
}

func (h *ProjectsHandler) buildViewModel(r *http.Request, w http.ResponseWriter) *view.ProjectsViewModel {
	user := middlewares.GetPrincipal(r)
	if user == nil { // this should actually never occur, because of auth middleware
		w.WriteHeader(http.StatusUnauthorized)
		return h.buildViewModel(r, w).WithError("unauthorized")
	}

	pageParams := utils.ParsePageParamsWithDefault(r, 1, 24)
	// note: pagination is not fully implemented, yet
	// count function to get total item / total pages is missing
	// and according ui (+ optionally search bar) is missing, too

	var err error
	var projects []*models.ProjectStats

	projects, err = h.heartbeatService.GetUserProjectStats(user, time.Time{}, utils.BeginOfToday(time.Local), pageParams, false)
	if err != nil {
		conf.Log().Request(r).Error("error while fetching project stats for '%s' - %v", user.ID, err)
		return &view.ProjectsViewModel{
			Messages:           view.Messages{Error: criticalError},
			LeaderboardEnabled: h.config.App.LeaderboardEnabled,
		}
	}

	vm := &view.ProjectsViewModel{
		User:               user,
		Projects:           projects,
		LeaderboardEnabled: h.config.App.LeaderboardEnabled,
		ApiKey:             user.ApiKey,
		PageParams:         pageParams,
	}
	return routeutils.WithSessionMessages(vm, r, w)
}
