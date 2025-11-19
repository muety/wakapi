package api

import (
	"net/http"

	"github.com/duke-git/lancet/v2/condition"
	"github.com/go-chi/chi/v5"
	"github.com/rs/cors"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/middlewares"
	customMiddleware "github.com/muety/wakapi/middlewares/custom"
	"github.com/muety/wakapi/models"
	v1 "github.com/muety/wakapi/models/compat/wakatime/v1"
	routeutils "github.com/muety/wakapi/routes/utils"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
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

func (h *HeartbeatApiHandler) RegisterRoutes(router chi.Router) {
	router.Group(func(r chi.Router) {
		r.Use(
			middlewares.NewAuthenticateMiddleware(h.userSrvc).WithOptionalForMethods(http.MethodOptions).WithFullAccessOnly(true).Handler,
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

		// https://github.com/muety/wakapi/issues/690
		for _, route := range r.Routes() {
			r.Options(route.Pattern, cors.AllowAll().HandlerFunc)
		}
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
		conf.Log().Request(r).Error("error occurred", "error", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	userAgent := r.Header.Get("User-Agent")
	opSys, editor, _ := utils.ParseUserAgent(userAgent)
	machineName := r.Header.Get("X-Machine-Name")

	creationResults := make(v1.HeartbeatCreationResults, len(heartbeats))

	for i, hb := range heartbeats {
		if hb == nil {
			creationResults[i] = &v1.HeartbeatCreationResult{
				Status: http.StatusBadRequest,
				Data:   &v1.HeartbeatResponseData{Error: "invalid heartbeat object"},
			}
			continue
		}

		// TODO: unit test this
		if hb.UserAgent != "" {
			userAgent = hb.UserAgent
			localOpSys, localEditor, _ := utils.ParseUserAgent(userAgent)
			opSys = condition.Ternary[bool, string](localOpSys != "", localOpSys, opSys)
			editor = condition.Ternary[bool, string](localEditor != "", localEditor, editor)
		}
		if hb.Machine != "" {
			machineName = hb.Machine
		}

		hb = fillPlaceholders(hb, user, h.heartbeatSrvc)

		hb.User = user
		hb.UserID = user.ID
		hb.Machine = machineName
		hb.OperatingSystem = opSys
		hb.Editor = editor
		hb.UserAgent = userAgent

		if !hb.Valid() || !hb.Timely(h.config.App.HeartbeatsMaxAge()) {
			creationResults[i] = &v1.HeartbeatCreationResult{
				Status: http.StatusBadRequest,
				Data:   &v1.HeartbeatResponseData{Error: "invalid heartbeat object"},
			}
			continue
		}

		hb.Hashed()
		creationResults[i] = v1.HeartbeatSuccess
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

	if creationResults.None() {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no valid heartbeat object given"))
		return
	}

	helpers.RespondJSON(w, r, http.StatusCreated, makeBulkResponse(creationResults))
}

// construct wakatime response format https://wakatime.com/developers#heartbeats (well, not quite...)
func makeBulkResponse(results []*v1.HeartbeatCreationResult) *v1.HeartbeatResponseViewModel {
	vm := &v1.HeartbeatResponseViewModel{
		Responses: make([][]interface{}, len(results)),
	}
	for i, r := range results {
		vm.Responses[i] = []interface{}{r.Data, r.Status}
	}
	return vm
}

// inplace!
func fillPlaceholders(hb *models.Heartbeat, user *models.User, srv services.IHeartbeatService) *models.Heartbeat {
	// wakatime has a special keyword that indicates to use the most recent project for a given heartbeat
	// in chrome, the browser extension sends this keyword for (most?) heartbeats
	// presumably the idea behind this is that if you're coding, your browsing activity will likely also relate to that coding project
	// but i don't really like this, i'd rather have all browsing activity under the "unknown" project (as the case with firefox, for whatever reason)
	// see https://github.com/wakatime/browser-wakatime/pull/206
	if hb.Type == "url" || hb.Type == "domain" {
		hb.ClearPlaceholders() // ignore placeholders for data sent by browser plugin
	}

	if hb.IsPlaceholderProject() {
		// get project of latest heartbeat by user
		if latest, err := srv.GetLatestByUser(user); latest != nil && err == nil {
			hb.Project = latest.Project
		}
	}

	if hb.IsPlaceholderLanguage() {
		// get language of latest heartbeat by user and project
		if latest, err := srv.GetLatestByFilters(user, models.NewFiltersWith(models.SummaryProject, hb.Project)); latest != nil && err == nil {
			hb.Language = latest.Language
		}
	}

	if hb.IsPlaceholderBranch() {
		// get branch of latest heartbeat by user and project
		if latest, err := srv.GetLatestByFilters(user, models.NewFiltersWith(models.SummaryProject, hb.Project)); latest != nil && err == nil {
			hb.Branch = latest.Branch
		}
	}

	hb.ClearPlaceholders()
	return hb
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
