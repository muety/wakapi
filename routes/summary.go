package routes

import (
	"net/http"
	"time"

	"github.com/n1try/wakapi/models"
	"github.com/n1try/wakapi/services"
	"github.com/n1try/wakapi/utils"
)

const (
	IntervalLastDay   string = "day"
	IntervalLastWeek  string = "week"
	IntervalLastMonth string = "month"
	IntervalLastYear  string = "year"
)

var summaryCache map[time.Time]*models.Summary

type SummaryHandler struct {
	SummarySrvc *services.SummaryService
}

func (h *SummaryHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	tryInitCache()

	user := r.Context().Value(models.UserKey).(*models.User)
	params := r.URL.Query()
	from, err := utils.ParseDate(params.Get("from"))
	if err != nil {
		interval := params.Get("interval")
		switch interval {
		case IntervalLastDay:
			from = utils.StartOfDay().Add(-24 * time.Hour)
		case IntervalLastWeek:
			from = utils.StartOfWeek()
		case IntervalLastMonth:
			from = utils.StartOfMonth()
		case IntervalLastYear:
			from = utils.StartOfYear()
		default:
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Missing 'from' parameter"))
			return
		}
	}

	if _, ok := summaryCache[from]; !ok {
		// Cache Miss
		summary, err := h.SummarySrvc.GetSummary(from, utils.StartOfDay(), user) // 'to' is always constant
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		summaryCache[from] = summary
	}

	summary, _ := summaryCache[from]
	utils.RespondJSON(w, http.StatusOK, summary)
}

func tryInitCache() {
	if summaryCache == nil {
		summaryCache = make(map[time.Time]*models.Summary)
	}
}
