package api

import (
	"net/http"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/helpers"
)

func (a *APIv1) FetchUserAgents(w http.ResponseWriter, r *http.Request) {
	user := helpers.ExtractUser(r)

	user_agents, err := a.services.UserAgentPlugin().FetchUserAgents(user.ID)
	if err != nil {
		conf.Log().Request(r).Error("failed to retrieve user agents - %v", err.Error(), err)
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
