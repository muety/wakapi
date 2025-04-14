package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/models"
	uuid "github.com/satori/go.uuid"
)

func (a *APIv1) UpdateGoal(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)
	goalID := chi.URLParam(r, "id")

	if goalID == "" {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Bad Request",
			"status":  http.StatusBadRequest,
		})
		return
	}

	var params = &models.Goal{}

	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(params)

	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid Input",
			"status":  http.StatusBadRequest,
		})
	}

	goal, err := a.services.Goal().GetGoalForUser(goalID, user.ID)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Goal Cannot Be Found",
			"status":  http.StatusBadRequest,
		})
		return
	}

	goal.CustomTitle = &params.Title

	_, err = a.services.Goal().Update(goal)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message":       "Error updating goal",
			"error_message": err.Error(),
		})
		return
	}
	response := map[string]interface{}{
		"data": goal,
	}
	helpers.RespondJSON(w, r, http.StatusCreated, response)
}

func (a *APIv1) GetGoal(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)
	goalID := chi.URLParam(r, "id")

	if goalID == "" {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Bad Request",
			"status":  http.StatusBadRequest,
		})
		return
	}

	goal, err := a.services.Goal().GetGoalForUser(goalID, user.ID)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Goal Cannot Be Found",
			"status":  http.StatusBadRequest,
		})
		return
	}
	chartData, err := a.services.Goal().LoadGoalChartData(goal, user, a.services.Summary())
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message":       "Error loading chart data",
			"error_message": err.Error(),
		})
		return
	}
	goal.ChartData = chartData
	response := map[string]interface{}{
		"data": goal,
	}
	helpers.RespondJSON(w, r, http.StatusCreated, response)
}

func (a *APIv1) DeleteGoal(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)
	goalID := chi.URLParam(r, "id")

	if goalID == "" {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Bad Request",
			"status":  http.StatusBadRequest,
		})
		return
	}

	err := a.services.Goal().DeleteGoal(goalID, user.ID)
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

func (a *APIv1) CreateGoal(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)

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

	_, err = a.services.Goal().Create(params)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message":       "An unexpected error occurred. Try again later",
			"error_message": err.Error(),
		})
		return
	}
	chartData, err := a.services.Goal().LoadGoalChartData(params, user, a.services.Summary())
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message":       "Error loading chart data",
			"error_message": err.Error(),
		})
		return
	}
	params.ChartData = chartData
	response := map[string]interface{}{
		"data": params,
	}
	helpers.RespondJSON(w, r, http.StatusCreated, response)
}

func (a *APIv1) FetchUserGoals(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)

	goals, err := a.services.Goal().FetchUserGoals(user.ID)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Error fetching goals. Try later.",
		})
		return
	}
	for _, goal := range goals {
		chartData, err := a.services.Goal().LoadGoalChartData(goal, user, a.services.Summary())
		if err != nil {
			helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
				"message":       "Error loading chart data",
				"error_message": err.Error(),
			})
			return
		}
		goal.ChartData = chartData
	}
	response := map[string]interface{}{
		"data": goals,
	}
	helpers.RespondJSON(w, r, http.StatusCreated, response)
}
