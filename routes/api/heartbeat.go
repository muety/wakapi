package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
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

type heartbeatResponseVm struct {
	Responses [][]interface{} `json:"responses"`
}

func (h *HeartbeatApiHandler) RegisterRoutes(router *mux.Router) {
	r := router.PathPrefix("").Subrouter()
	r.Use(
		middlewares.NewAuthenticateMiddleware(h.userSrvc).Handler,
		customMiddleware.NewWakatimeRelayMiddleware().Handler,
	)
	// see https://github.com/muety/wakapi/issues/203
	r.Path("/heartbeat").Methods(http.MethodPost).HandlerFunc(h.Post)
	r.Path("/heartbeats").Methods(http.MethodPost).HandlerFunc(h.Post)
	r.Path("/users/{user}/heartbeats").Methods(http.MethodPost).HandlerFunc(h.Post)
	r.Path("/users/{user}/heartbeats.bulk").Methods(http.MethodPost).HandlerFunc(h.Post)
	r.Path("/v1/users/{user}/heartbeats").Methods(http.MethodPost).HandlerFunc(h.Post)
	r.Path("/v1/users/{user}/heartbeats.bulk").Methods(http.MethodPost).HandlerFunc(h.Post)
	r.Path("/compat/wakatime/v1/users/{user}/heartbeats").Methods(http.MethodPost).HandlerFunc(h.Post)
	r.Path("/compat/wakatime/v1/users/{user}/heartbeats.bulk").Methods(http.MethodPost).HandlerFunc(h.Post)
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
	heartbeats, err = h.tryParseBulk(r)
	if err != nil {
		heartbeats, err = h.tryParseSingle(r)
		if err != nil {
			conf.Log().Request(r).Error(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
	}

	userAgent := r.Header.Get("User-Agent")
	opSys, editor, _ := utils.ParseUserAgent(userAgent)
	machineName := r.Header.Get("X-Machine-Name")

	for _, hb := range heartbeats {
		hb.OperatingSystem = opSys
		hb.Editor = editor
		hb.Machine = machineName
		hb.User = user
		hb.UserID = user.ID
		hb.UserAgent = userAgent

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
		conf.Log().Request(r).Error("failed to batch-insert heartbeats – %v", err)
		return
	}

	if !user.HasData {
		user.HasData = true
		if _, err := h.userSrvc.Update(user); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(conf.ErrInternalServerError))
			conf.Log().Request(r).Error("failed to update user – %v", err)
			return
		}
	}

	defer func() {}()

	utils.RespondJSON(w, r, http.StatusCreated, constructSuccessResponse(len(heartbeats)))
}

func (h *HeartbeatApiHandler) tryParseBulk(r *http.Request) ([]*models.Heartbeat, error) {
	var heartbeats []*models.Heartbeat

	body, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	dec := json.NewDecoder(ioutil.NopCloser(bytes.NewBuffer(body)))
	if err := dec.Decode(&heartbeats); err != nil {
		return nil, err
	}

	return heartbeats, nil
}

func (h *HeartbeatApiHandler) tryParseSingle(r *http.Request) ([]*models.Heartbeat, error) {
	var heartbeat models.Heartbeat

	body, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	dec := json.NewDecoder(ioutil.NopCloser(bytes.NewBuffer(body)))
	if err := dec.Decode(&heartbeat); err != nil {
		return nil, err
	}

	return []*models.Heartbeat{&heartbeat}, nil
}

// construct weird response format (see https://github.com/wakatime/wakatime/blob/2e636d389bf5da4e998e05d5285a96ce2c181e3d/wakatime/api.py#L288)
// to make the cli consider all heartbeats to having been successfully saved
// response looks like: { "responses": [ [ null, 201 ], ... ] }
// this was probably a temporary bug at wakatime, responses actually looks like so: https://pastr.de/p/nyf6kj2e6843fbw4xkj4h4pj
// TODO: adapt response format some time
// however, wakatime-cli is still able to parse the response (see https://github.com/wakatime/wakatime-cli/blob/c2076c0e1abc1449baf5b7ac7db391b06041c719/pkg/api/heartbeat.go#L127), so no urgent need for action
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

// Only for Swagger

// @Summary Push a new heartbeat
// @ID post-heartbeat-2
// @Tags heartbeat
// @Accept json
// @Param heartbeat body models.Heartbeat true "A single heartbeat"
// @Security ApiKeyAuth
// @Success 201
// @Router /v1/users/{user}/heartbeats [post]
func (h *HeartbeatApiHandler) postAlias1() {}

// @Summary Push a new heartbeat
// @ID post-heartbeat-3
// @Tags heartbeat
// @Accept json
// @Param heartbeat body models.Heartbeat true "A single heartbeat"
// @Security ApiKeyAuth
// @Success 201
// @Router /compat/wakatime/v1/users/{user}/heartbeats [post]
func (h *HeartbeatApiHandler) postAlias2() {}

// @Summary Push a new heartbeat
// @ID post-heartbeat-4
// @Tags heartbeat
// @Accept json
// @Param heartbeat body models.Heartbeat true "A single heartbeat"
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
// @Security ApiKeyAuth
// @Success 201
// @Router /v1/users/{user}/heartbeats.bulk [post]
func (h *HeartbeatApiHandler) postAlias5() {}

// @Summary Push new heartbeats
// @ID post-heartbeat-7
// @Tags heartbeat
// @Accept json
// @Param heartbeat body []models.Heartbeat true "Multiple heartbeats"
// @Security ApiKeyAuth
// @Success 201
// @Router /compat/wakatime/v1/users/{user}/heartbeats.bulk [post]
func (h *HeartbeatApiHandler) postAlias6() {}

// @Summary Push new heartbeats
// @ID post-heartbeat-8
// @Tags heartbeat
// @Accept json
// @Param heartbeat body []models.Heartbeat true "Multiple heartbeats"
// @Security ApiKeyAuth
// @Success 201
// @Router /users/{user}/heartbeats.bulk [post]
func (h *HeartbeatApiHandler) postAlias7() {}
