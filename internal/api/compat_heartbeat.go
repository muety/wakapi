package api

import (
	"net/http"
	"time"

	"github.com/duke-git/lancet/v2/datetime"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/internal/utilities"

	conf "github.com/muety/wakapi/config"
	wakatime "github.com/muety/wakapi/models/compat/wakatime/v1"
)

type HeartbeatsResult struct {
	Data     []*wakatime.HeartbeatEntry `json:"data"`
	End      string                     `json:"end"`
	Start    string                     `json:"start"`
	Timezone string                     `json:"timezone"`
}

// @Summary Get heartbeats of user for specified date
// @ID get-heartbeats
// @Tags heartbeat
// @Param date query string true "Date"
// @Param user path string true "Username (or current)"
// @Security ApiKeyAuth
// @Success 200 {object} HeartbeatsResult
// @Failure 400 {string} string "bad date"
// @Router /compat/wakatime/v1/users/{user}/heartbeats [get]
func (a *APIv1) GetHeartBeats(w http.ResponseWriter, r *http.Request) {
	user, err := utilities.CheckEffectiveUser(w, r, a.services.Users(), "current")
	if err != nil {
		return // response was already sent by util function
	}

	params := r.URL.Query()
	dateParam := params.Get("date")
	date, err := time.Parse(conf.SimpleDateFormat, dateParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad date"))
		return
	}

	timezone := user.TZ()
	rangeFrom, rangeTo := datetime.BeginOfDay(date.In(timezone)), datetime.EndOfDay(date.In(timezone))

	heartbeats, err := a.services.Heartbeat().GetAllWithin(rangeFrom, rangeTo, user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(conf.ErrInternalServerError))
		conf.Log().Request(r).Error("failed to retrieve heartbeats", "error", err)
		return
	}

	res := HeartbeatsResult{
		Data:     wakatime.HeartbeatsToCompat(heartbeats),
		Start:    rangeFrom.UTC().Format(time.RFC3339),
		End:      rangeTo.UTC().Format(time.RFC3339),
		Timezone: timezone.String(),
	}
	helpers.RespondJSON(w, r, http.StatusOK, res)
}
