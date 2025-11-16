package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models/view"
	"github.com/muety/wakapi/services"
)

type SetupHandler struct {
	config             *conf.Config
	userService        services.IUserService
	leaderboardService services.ILeaderboardService
}

func NewSetupHandler(userService services.IUserService) *SetupHandler {
	return &SetupHandler{
		config:      conf.Get(),
		userService: userService,
	}
}

func (h *SetupHandler) RegisterRoutes(router chi.Router) {
	r := chi.NewRouter()

	authMiddleware := middlewares.NewAuthenticateMiddleware(h.userService).
		WithRedirectTarget(defaultErrorRedirectTarget()).
		WithRedirectErrorMessage("unauthorized").
		WithOptionalFor("/")

	r.Use(authMiddleware.Handler)
	r.Get("/", h.GetIndex)

	router.Mount("/setup", r)
}

func (h *SetupHandler) GetIndex(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	templates[conf.SetupTemplate].Execute(w, h.buildViewModel(r, w))
}

func (h *SetupHandler) buildViewModel(r *http.Request, w http.ResponseWriter) *view.SetupViewModel {
	return &view.SetupViewModel{
		SharedLoggedInViewModel: view.SharedLoggedInViewModel{
			SharedViewModel: view.NewSharedViewModel(h.config, nil),
			User:            middlewares.GetPrincipal(r),
		},
	}
}
