package routes

import (
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"net/http"
)

type SummaryHandler struct {
	summarySrvc *services.SummaryService
	config      *conf.Config
}

func NewSummaryHandler(summaryService *services.SummaryService) *SummaryHandler {
	return &SummaryHandler{
		summarySrvc: summaryService,
		config:      conf.Get(),
	}
}

func (h *SummaryHandler) ApiGet(w http.ResponseWriter, r *http.Request) {
	summary, err, status := h.loadUserSummary(r)
	if err != nil {
		w.WriteHeader(status)
		w.Write([]byte(err.Error()))
		return
	}

	utils.RespondJSON(w, http.StatusOK, summary)
}

func (h *SummaryHandler) GetIndex(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	q := r.URL.Query()
	if q.Get("interval") == "" && q.Get("from") == "" {
		q.Set("interval", "today")
		r.URL.RawQuery = q.Encode()
	}

	summary, err, status := h.loadUserSummary(r)
	if err != nil {
		respondAlert(w, err.Error(), "", conf.SummaryTemplate, status)
		return
	}

	user := r.Context().Value(models.UserKey).(*models.User)
	if user == nil {
		respondAlert(w, "unauthorized", "", conf.SummaryTemplate, http.StatusUnauthorized)
		return
	}

	vm := models.SummaryViewModel{
		Summary:        summary,
		LanguageColors: utils.FilterLanguageColors(h.config.App.LanguageColors, summary),
		ApiKey:         user.ApiKey,
	}

	templates[conf.SummaryTemplate].Execute(w, vm)
}

func (h *SummaryHandler) loadUserSummary(r *http.Request) (*models.Summary, error, int) {
	summaryParams, err := utils.ParseSummaryParams(r)
	if err != nil {
		return nil, err, http.StatusBadRequest
	}

	summary, err := h.summarySrvc.PostProcessWrapped(
		h.summarySrvc.Construct(summaryParams.From, summaryParams.To, summaryParams.User, summaryParams.Recompute), // 'to' is always constant
	)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}

	return summary, nil, http.StatusOK
}
