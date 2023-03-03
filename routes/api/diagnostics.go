package api

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/muety/wakapi/helpers"
	"net/http"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/services"
)

type DiagnosticsApiHandler struct {
	config          *conf.Config
	userSrvc        services.IUserService
	diagnosticsSrvc services.IDiagnosticsService
}

func NewDiagnosticsApiHandler(userService services.IUserService, diagnosticsService services.IDiagnosticsService) *DiagnosticsApiHandler {
	return &DiagnosticsApiHandler{
		config:          conf.Get(),
		userSrvc:        userService,
		diagnosticsSrvc: diagnosticsService,
	}
}

func (h *DiagnosticsApiHandler) RegisterRoutes(router chi.Router) {
	router.Post("/plugins/errors", h.Post)
}

// @Summary Push a new diagnostics object
// @ID post-diagnostics
// @Tags diagnostics
// @Accept json
// @Param diagnostics body models.Diagnostics true "A single diagnostics object sent by WakaTime CLI"
// @Success 201
// @Router /plugins/errors [post]
func (h *DiagnosticsApiHandler) Post(w http.ResponseWriter, r *http.Request) {
	var diagnostics models.Diagnostics

	if err := json.NewDecoder(r.Body).Decode(&diagnostics); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(conf.ErrBadRequest))
		conf.Log().Request(r).Error("failed to parse diagnostics for user %s - %v", err)
		return
	}

	if _, err := h.diagnosticsSrvc.Create(&diagnostics); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(conf.ErrInternalServerError))
		conf.Log().Request(r).Error("failed to insert diagnostics for user %s - %v", err)
		return
	}

	helpers.RespondJSON(w, r, http.StatusCreated, struct{}{})
}
