package api

import (
	"net/http"

	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/internal/utilities"

	conf "github.com/muety/wakapi/config"
	v1 "github.com/muety/wakapi/models/compat/wakatime/v1"
)

// @Summary Retrieve the given user
// @Description Mimics https://wakatime.com/developers#users
// @ID get-wakatime-user
// @Tags wakatime
// @Produce json
// @Param user path string true "User ID to fetch (or 'current')"
// @Security ApiKeyAuth
// @Success 200 {object} v1.UserViewModel
// @Router /compat/wakatime/v1/users/{user} [get]
func (a *APIv1) GetUser(w http.ResponseWriter, r *http.Request) {
	wakapiUser, err := utilities.CheckEffectiveUser(w, r, a.services.Users(), "current")
	if err != nil {
		return // response was already sent by util function
	}

	user := v1.NewFromUser(wakapiUser)
	if hb, err := a.services.Heartbeat().GetLatestByUser(wakapiUser); err == nil {
		user = user.WithLatestHeartbeat(hb)
	} else {
		conf.Log().Request(r).Error("error occurred", "error", err)
	}

	helpers.RespondJSON(w, r, http.StatusOK, v1.UserViewModel{Data: user})
}
