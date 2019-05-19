package routes

import (
	"net/http"
	"time"

	"github.com/n1try/wakapi/models"
	"github.com/n1try/wakapi/services"
	"github.com/n1try/wakapi/utils"
)

type SummaryHandler struct {
	SummarySrvc *services.SummaryService
}

func (h *SummaryHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	user := r.Context().Value(models.UserKey).(*models.User)
	params := r.URL.Query()
	from, err := utils.ParseDate(params.Get("from"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing 'from' parameter"))
		return
	}

	now := time.Now()
	to := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()) // Start of current day

	summary, err := h.SummarySrvc.GetSummary(from, to, user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, http.StatusOK, summary)
}
