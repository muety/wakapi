package routes

import (
	"crypto/md5"
	"net/http"
	"strconv"
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

var summaryCache map[string]*models.Summary

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

	live := params.Get("live") != "" && params.Get("live") != "false"
	to := utils.StartOfDay()
	if live {
		to = time.Now()
	}

	var summary *models.Summary
	cacheKey := getHash([]time.Time{from, to})
	if _, ok := summaryCache[cacheKey]; !ok {
		// Cache Miss
		summary, err = h.SummarySrvc.GetSummary(from, to, user) // 'to' is always constant
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !live {
			summaryCache[cacheKey] = summary
		}
	} else {
		summary, _ = summaryCache[cacheKey]
	}

	utils.RespondJSON(w, http.StatusOK, summary)
}

func tryInitCache() {
	if summaryCache == nil {
		summaryCache = make(map[string]*models.Summary)
	}
}

func getHash(times []time.Time) string {
	digest := md5.New()
	for _, t := range times {
		digest.Write([]byte(strconv.Itoa(int(t.Unix()))))
	}
	return string(digest.Sum(nil))
}
