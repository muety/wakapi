package v1

import (
	"github.com/gorilla/mux"
	config2 "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	v1 "github.com/muety/wakapi/models/compat/wakatime/v1"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"net/http"
	"net/url"
	"time"
)

type AllTimeHandler struct {
	summarySrvc *services.SummaryService
	config      *config2.Config
}

func NewAllTimeHandler(summaryService *services.SummaryService) *AllTimeHandler {
	return &AllTimeHandler{
		summarySrvc: summaryService,
		config:      config2.Get(),
	}
}

func (h *AllTimeHandler) ApiGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	values, _ := url.ParseQuery(r.URL.RawQuery)

	requestedUser := vars["user"]
	authorizedUser := r.Context().Value(models.UserKey).(*models.User)

	if requestedUser != authorizedUser.ID && requestedUser != "current" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	summary, err, status := h.loadUserSummary(authorizedUser)
	if err != nil {
		w.WriteHeader(status)
		w.Write([]byte(err.Error()))
		return
	}

	vm := v1.NewAllTimeFrom(summary, models.NewFiltersWith(models.SummaryProject, values.Get("project")))
	utils.RespondJSON(w, http.StatusOK, vm)
}

func (h *AllTimeHandler) loadUserSummary(user *models.User) (*models.Summary, error, int) {
	summaryParams := &models.SummaryParams{
		From:      time.Time{},
		To:        time.Now(),
		User:      user,
		Recompute: false,
	}

	summary, err := h.summarySrvc.PostProcessWrapped(
		h.summarySrvc.Construct(summaryParams.From, summaryParams.To, summaryParams.User, summaryParams.Recompute), // 'to' is always constant
	)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}

	return summary, nil, http.StatusOK
}
