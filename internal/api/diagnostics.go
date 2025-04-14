package api

import (
	"encoding/json"
	"net/http"

	"github.com/muety/wakapi/helpers"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
)

// @Summary Push a new diagnostics object
// @ID post-diagnostics
// @Tags diagnostics
// @Accept json
// @Param diagnostics body models.Diagnostics true "A single diagnostics object sent by WakaTime CLI"
// @Success 201
// @Router /plugins/errors [post]
func (a *APIv1) PostDiagnostics(w http.ResponseWriter, r *http.Request) {
	var diagnostics models.Diagnostics

	if err := json.NewDecoder(r.Body).Decode(&diagnostics); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(conf.ErrBadRequest))
		conf.Log().Request(r).Error("failed to parse diagnostics for user", "error", err)
		return
	}

	if _, err := a.services.Diagnostics().Create(&diagnostics); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(conf.ErrInternalServerError))
		conf.Log().Request(r).Error("failed to insert diagnostics for user", "error", err)
		return
	}

	helpers.RespondJSON(w, r, http.StatusCreated, struct{}{})
}
