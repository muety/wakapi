package v1

import (
	"github.com/duke-git/lancet/v2/datetime"
	"github.com/go-chi/chi/v5"
	"github.com/muety/wakapi/helpers"
	"net/http"
	"time"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	wakatime "github.com/muety/wakapi/models/compat/wakatime/v1"
	routeutils "github.com/muety/wakapi/routes/utils"
	"github.com/muety/wakapi/services"
)

type HeartbeatsResult struct {
	Data     []*wakatime.HeartbeatEntry `json:"data"`
	End      string                     `json:"end"`
	Start    string                     `json:"start"`
	Timezone string                     `json:"timezone"`
}

type HeartbeatHandler struct {
	userSrvc      services.IUserService
	heartbeatSrvc services.IHeartbeatService
}

func NewHeartbeatHandler(userService services.IUserService, heartbeatService services.IHeartbeatService) *HeartbeatHandler {
	return &HeartbeatHandler{
		userSrvc:      userService,
		heartbeatSrvc: heartbeatService,
	}
}

func (h *HeartbeatHandler) RegisterRoutes(router chi.Router) {
	router.Group(func(r chi.Router) {
		r.Use(middlewares.NewAuthenticateMiddleware(h.userSrvc).Handler)
		r.Get("/compat/wakatime/v1/users/{user}/heartbeats", h.Get)
	})
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
func (h *HeartbeatHandler) Get(w http.ResponseWriter, r *http.Request) {
	user, err := routeutils.CheckEffectiveUser(w, r, h.userSrvc, "current")
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

	heartbeats, err := h.heartbeatSrvc.GetAllWithin(rangeFrom, rangeTo, user)
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
