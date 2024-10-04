package api

import (
	"github.com/duke-git/lancet/v2/condition"
	"github.com/go-chi/chi/v5"
	"github.com/muety/wakapi/helpers"
	"net/http"
	"strconv"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	customMiddleware "github.com/muety/wakapi/middlewares/custom"
	routeutils "github.com/muety/wakapi/routes/utils"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"

	"github.com/muety/wakapi/models"
)

type HeartbeatApiHandler struct {
	config              *conf.Config
	userSrvc            services.IUserService
	heartbeatSrvc       services.IHeartbeatService
	languageMappingSrvc services.ILanguageMappingService
}

func NewHeartbeatApiHandler(userService services.IUserService, heartbeatService services.IHeartbeatService, languageMappingService services.ILanguageMappingService) *HeartbeatApiHandler {
	return &HeartbeatApiHandler{
		config:              conf.Get(),
		userSrvc:            userService,
		heartbeatSrvc:       heartbeatService,
		languageMappingSrvc: languageMappingService,
	}
}

type heartbeatResponseData struct {
	ID     string            `json:"id"`
	Entity string            `json:"entity"`
	Type   string            `json:"type"`
	Time   models.CustomTime `json:"time"`
}

type heartbeatResponseDataWrapper struct {
	Data *heartbeatResponseData `json:"data"`
}

type heartbeatResponseVm struct {
	Responses [][]interface{} `json:"responses"`
}

func (h *HeartbeatApiHandler) RegisterRoutes(router chi.Router) {
	router.Group(func(r chi.Router) {
		r.Use(
			middlewares.NewAuthenticateMiddleware(h.userSrvc).Handler,
			customMiddleware.NewWakatimeRelayMiddleware().Handler,
		)
		// see https://github.com/muety/wakapi/issues/203
		r.Post("/heartbeat", h.Post)
		r.Post("/heartbeats", h.Post)
		r.Post("/users/{user}/heartbeats", h.Post)
		r.Post("/users/{user}/heartbeats.bulk", h.Post)
		r.Post("/v1/users/{user}/heartbeats", h.Post)
		r.Post("/v1/users/{user}/heartbeats.bulk", h.Post)
		r.Post("/compat/wakatime/v1/users/{user}/heartbeats", h.Post)
		r.Post("/compat/wakatime/v1/users/{user}/heartbeats.bulk", h.Post)
	})
}

// @Summary Push a new heartbeat
// @ID post-heartbeat
// @Tags heartbeat
// @Accept json
// @Param heartbeat body models.Heartbeat true "A single heartbeat"
// @Security ApiKeyAuth
// @Success 201
// @Router /heartbeat [post]
func (h *HeartbeatApiHandler) Post(w http.ResponseWriter, r *http.Request) {
	user, err := routeutils.CheckEffectiveUser(w, r, h.userSrvc, "current")
	if err != nil {
		return // response was already sent by util function
	}

	var heartbeats []*models.Heartbeat
	heartbeats, err = routeutils.ParseHeartbeats(r)
	if err != nil {
		conf.Log().Request(r).Error("error occurred", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	userAgent := r.Header.Get("User-Agent")
	opSys, editor, _ := utils.ParseUserAgent(userAgent)
	machineName := r.Header.Get("X-Machine-Name")

	for _, hb := range heartbeats {
		if hb == nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid heartbeat object"))
			return
		}

		// TODO: unit test this
		if hb.UserAgent != "" {
			userAgent = hb.UserAgent
			localOpSys, localEditor, _ := utils.ParseUserAgent(userAgent)
			opSys = condition.TernaryOperator[bool, string](localOpSys != "", localOpSys, opSys)
			editor = condition.TernaryOperator[bool, string](localEditor != "", localEditor, editor)
		}
		if hb.Machine != "" {
			machineName = hb.Machine
		}

		if hb.Branch == "<<LAST_BRANCH>>" {
			if latest, err := h.heartbeatSrvc.GetLatestByFilters(user, models.NewFiltersWith(models.SummaryProject, hb.Project)); latest != nil && err == nil {
				hb.Branch = latest.Branch
			} else {
				hb.Branch = ""
			}
		}

		hb.User = user
		hb.UserID = user.ID
		hb.Machine = machineName
		hb.OperatingSystem = opSys
		hb.Editor = editor
		hb.UserAgent = userAgent

		if !hb.Valid() || !hb.Timely(h.config.App.HeartbeatsMaxAge()) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid heartbeat object"))
			return
		}

		hb.Hashed()
	}

	if err := h.heartbeatSrvc.InsertBatch(heartbeats); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(conf.ErrInternalServerError))
		conf.Log().Request(r).Error("failed to batch-insert heartbeats", "error", err)
		return
	}

	if !user.HasData {
		user.HasData = true
		if _, err := h.userSrvc.Update(user); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(conf.ErrInternalServerError))
			conf.Log().Request(r).Error("failed to update user", "userID", user.ID, "error", err)
			return
		}
	}

	defer func() {}()

	helpers.RespondJSON(w, r, http.StatusCreated, constructSuccessResponse(&heartbeats))
}

// construct wakatime response format https://wakatime.com/developers#heartbeats
func constructSuccessResponse(heartbeats *[]*models.Heartbeat) *heartbeatResponseVm {
	responses := make([][]interface{}, len(*heartbeats))

	for i, heartbeat := range *heartbeats {
		r := make([]interface{}, 2)
		r[0] = &heartbeatResponseDataWrapper{
			Data: &heartbeatResponseData{
				// TODO: must be UUID, but we don't use UUID
				ID:     strconv.FormatUint(heartbeat.ID, 10),
				Entity: heartbeat.Entity,
				Type:   heartbeat.Type,
				Time:   heartbeat.Time,
			},
		}
		r[1] = http.StatusCreated
		responses[i] = r
	}

	return &heartbeatResponseVm{
		Responses: responses,
	}
}

// Only for Swagger

// @Summary Push a new heartbeat
// @ID post-heartbeat-2
// @Tags heartbeat
// @Accept json
// @Param heartbeat body models.Heartbeat true "A single heartbeat"
// @Param user path string true "Username (or current)"
// @Security ApiKeyAuth
// @Success 201
// @Router /v1/users/{user}/heartbeats [post]
func (h *HeartbeatApiHandler) postAlias1() {}

// @Summary Push a new heartbeat
// @ID post-heartbeat-3
// @Tags heartbeat
// @Accept json
// @Param heartbeat body models.Heartbeat true "A single heartbeat"
// @Param user path string true "Username (or current)"
// @Security ApiKeyAuth
// @Success 201
// @Router /compat/wakatime/v1/users/{user}/heartbeats [post]
func (h *HeartbeatApiHandler) postAlias2() {}

// @Summary Push a new heartbeat
// @ID post-heartbeat-4
// @Tags heartbeat
// @Accept json
// @Param heartbeat body models.Heartbeat true "A single heartbeat"
// @Param user path string true "Username (or current)"
// @Security ApiKeyAuth
// @Success 201
// @Router /users/{user}/heartbeats [post]
func (h *HeartbeatApiHandler) postAlias3() {}

// @Summary Push new heartbeats
// @ID post-heartbeat-5
// @Tags heartbeat
// @Accept json
// @Param heartbeat body []models.Heartbeat true "Multiple heartbeats"
// @Security ApiKeyAuth
// @Success 201
// @Router /heartbeats [post]
func (h *HeartbeatApiHandler) postAlias4() {}

// @Summary Push new heartbeats
// @ID post-heartbeat-6
// @Tags heartbeat
// @Accept json
// @Param heartbeat body []models.Heartbeat true "Multiple heartbeats"
// @Param user path string true "Username (or current)"
// @Security ApiKeyAuth
// @Success 201
// @Router /v1/users/{user}/heartbeats.bulk [post]
func (h *HeartbeatApiHandler) postAlias5() {}

// @Summary Push new heartbeats
// @ID post-heartbeat-7
// @Tags heartbeat
// @Accept json
// @Param heartbeat body []models.Heartbeat true "Multiple heartbeats"
// @Param user path string true "Username (or current)"
// @Security ApiKeyAuth
// @Success 201
// @Router /compat/wakatime/v1/users/{user}/heartbeats.bulk [post]
func (h *HeartbeatApiHandler) postAlias6() {}

// @Summary Push new heartbeats
// @ID post-heartbeat-8
// @Tags heartbeat
// @Accept json
// @Param heartbeat body []models.Heartbeat true "Multiple heartbeats"
// @Param user path string true "Username (or current)"
// @Security ApiKeyAuth
// @Success 201
// @Router /users/{user}/heartbeats.bulk [post]
func (h *HeartbeatApiHandler) postAlias7() {}
