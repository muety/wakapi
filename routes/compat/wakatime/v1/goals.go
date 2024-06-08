package v1

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/services"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type GoalsApiHandler struct {
	db          *gorm.DB
	config      *conf.Config
	goalService services.IGoalService
	userSrvc    services.IUserService
}

func NewGoalsApiHandler(db *gorm.DB, goalService services.IGoalService, userSrvc services.IUserService) *GoalsApiHandler {
	return &GoalsApiHandler{db: db, goalService: goalService, userSrvc: userSrvc, config: conf.Get()}
}

func (h *GoalsApiHandler) RegisterRoutes(router chi.Router) {
	router.Group(func(r chi.Router) {
		r.Use(middlewares.NewAuthenticateMiddleware(h.userSrvc).Handler)
		r.Post("/compat/wakatime/v1/users/{user}/goals", h.CreateGoal)
		r.Get("/compat/wakatime/v1/users/{user}/goals", h.FetchUserGoals)
		r.Get("/compat/wakatime/v1/users/{user}/goals/{id}", h.GetGoal)
		r.Delete("/compat/wakatime/v1/users/{user}/goals/{id}", h.DeleteGoal)
	})
}

func extractUser(r *http.Request) *models.User {
	type principalGetter interface {
		GetPrincipal() *models.User
	}
	if p := r.Context().Value("principal"); p != nil {
		return p.(principalGetter).GetPrincipal()
	}
	return nil
}

func (h *GoalsApiHandler) GetGoal(w http.ResponseWriter, r *http.Request) {
	user := extractUser(r)
	goalID := chi.URLParam(r, "id")

	if goalID == "" {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Bad Request",
			"status":  http.StatusBadRequest,
		})
		return
	}

	goal, err := h.goalService.GetGoalForUser(goalID, user.ID)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Goal Cannot Be Found",
			"status":  http.StatusBadRequest,
		})
		return
	}
	response := map[string]interface{}{
		"data": goal,
	}
	helpers.RespondJSON(w, r, http.StatusCreated, response)
}

func (h *GoalsApiHandler) DeleteGoal(w http.ResponseWriter, r *http.Request) {
	user := extractUser(r)
	fmt.Println("User", user)
	goalID := chi.URLParam(r, "id")

	if goalID == "" {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Bad Request",
			"status":  http.StatusBadRequest,
		})
		return
	}

	err := h.goalService.DeleteGoal(goalID, user.ID)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Goal Cannot Be Deleted",
			"status":  http.StatusBadRequest,
		})
		return
	}
	response := map[string]interface{}{
		"message": "Goal deleted successfully",
	}
	helpers.RespondJSON(w, r, http.StatusAccepted, response)
}

func (h *GoalsApiHandler) CreateGoal(w http.ResponseWriter, r *http.Request) {
	user := extractUser(r)

	var params = &models.Goal{}

	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(params)

	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid Input",
			"status":  http.StatusBadRequest,
		})
		return
	}

	params.UserID = user.ID
	params.ID = uuid.NewV4().String()
	params.Title = params.GetTitle()

	_, err = h.goalService.Create(params)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message":       "An unexpected error occurred. Try again later",
			"error_message": err.Error(),
		})
		return
	}
	response := map[string]interface{}{
		"data": params,
	}
	helpers.RespondJSON(w, r, http.StatusCreated, response)
}

func (h *GoalsApiHandler) FetchUserGoals(w http.ResponseWriter, r *http.Request) {
	user := extractUser(r)

	goals, err := h.goalService.FetchUserGoals(user.ID)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Error fetching goals. Try later.",
		})
		return
	}
	response := map[string]interface{}{
		"data": goals,
	}
	helpers.RespondJSON(w, r, http.StatusCreated, response)
}
