package v1

import (
	"github.com/gorilla/mux"
	"github.com/muety/wakapi/models"
	v1 "github.com/muety/wakapi/models/compat/v1"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"net/http"
	"net/url"
	"time"
)

type CompatV1AllHandler struct {
	summarySrvc *services.SummaryService
	config      *models.Config
}

func NewCompatV1AllHandler(summaryService *services.SummaryService) *CompatV1AllHandler {
	return &CompatV1AllHandler{
		summarySrvc: summaryService,
		config:      models.GetConfig(),
	}
}

func (h *CompatV1AllHandler) ApiGet(w http.ResponseWriter, r *http.Request) {
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

	vm := v1.NewAllTimeFrom(summary, &v1.Filters{Project: values.Get("project")})
	utils.RespondJSON(w, http.StatusOK, vm)
}

func (h *CompatV1AllHandler) loadUserSummary(user *models.User) (*models.Summary, error, int) {
	summaryParams := &models.SummaryParams{
		From:      time.Time{},
		To:        time.Now(),
		User:      user,
		Recompute: false,
	}

	summary, err := h.summarySrvc.Construct(summaryParams.From, summaryParams.To, summaryParams.User, summaryParams.Recompute) // 'to' is always constant
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}

	return summary, nil, http.StatusOK
}
