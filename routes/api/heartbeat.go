package api

import (
	"encoding/json"
	"github.com/emvi/logbuch"
	"github.com/gorilla/mux"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	customMiddleware "github.com/muety/wakapi/middlewares/custom"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"net/http"

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

type heartbeatResponseVm struct {
	Responses [][]interface{} `json:"responses"`
}

func (h *HeartbeatApiHandler) RegisterRoutes(router *mux.Router) {
	r := router.PathPrefix("/heartbeat").Subrouter()
	r.Use(
		middlewares.NewAuthenticateMiddleware(h.userSrvc).Handler,
		customMiddleware.NewWakatimeRelayMiddleware().Handler,
	)
	r.Methods(http.MethodPost).HandlerFunc(h.Post)
}

// @Summary Push a new heartbeat
// @ID post-heartbeat
// @Tags heartbeat
// @Accept json
// @Param heartbeat body models.Heartbeat true "A heartbeat"
// @Security ApiKeyAuth
// @Success 201
// @Router /heartbeat [post]
func (h *HeartbeatApiHandler) Post(w http.ResponseWriter, r *http.Request) {
	var heartbeats []*models.Heartbeat
	user := middlewares.GetPrincipal(r)
	opSys, editor, _ := utils.ParseUserAgent(r.Header.Get("User-Agent"))
	machineName := r.Header.Get("X-Machine-Name")

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&heartbeats); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	for _, hb := range heartbeats {
		hb.OperatingSystem = opSys
		hb.Editor = editor
		hb.Machine = machineName
		hb.User = user
		hb.UserID = user.ID

		if !hb.Valid() {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid heartbeat object"))
			return
		}

		hb.Hashed()
	}

	if err := h.heartbeatSrvc.InsertBatch(heartbeats); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(conf.ErrInternalServerError))
		logbuch.Error("failed to batch-insert heartbeats – %v", err)
		return
	}

	if !user.HasData {
		user.HasData = true
		if _, err := h.userSrvc.Update(user); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(conf.ErrInternalServerError))
			logbuch.Error("failed to update user – %v", err)
			return
		}
	}

	utils.RespondJSON(w, http.StatusCreated, constructSuccessResponse(len(heartbeats)))
}

// construct weird response format (see https://github.com/wakatime/wakatime/blob/2e636d389bf5da4e998e05d5285a96ce2c181e3d/wakatime/api.py#L288)
// to make the cli consider all heartbeats to having been successfully saved
// response looks like: { "responses": [ [ { "data": {...} }, 201 ], ... ] }
func constructSuccessResponse(n int) *heartbeatResponseVm {
	responses := make([][]interface{}, n)

	for i := 0; i < n; i++ {
		r := make([]interface{}, 2)
		r[0] = nil
		r[1] = http.StatusCreated
		responses[i] = r
	}

	return &heartbeatResponseVm{
		Responses: responses,
	}
}
