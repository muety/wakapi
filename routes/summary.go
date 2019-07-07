package routes

import (
	"crypto/md5"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/n1try/wakapi/models"
	"github.com/n1try/wakapi/services"
	"github.com/n1try/wakapi/utils"
	cache "github.com/patrickmn/go-cache"
)

const (
	IntervalToday	  string = "today"
	IntervalLastDay   string = "day"
	IntervalLastWeek  string = "week"
	IntervalLastMonth string = "month"
	IntervalLastYear  string = "year"
)

type SummaryHandler struct {
	SummarySrvc *services.SummaryService
	Cache       *cache.Cache
	Initialized bool
}

func (m *SummaryHandler) Init() {
	if m.Cache == nil {
		m.Cache = cache.New(24*time.Hour, 24*time.Hour)
	}
	m.Initialized = true
}

func (h *SummaryHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if !h.Initialized {
		h.Init()
	}

	user := r.Context().Value(models.UserKey).(*models.User)

	// Initialize aliases for user
	if !h.SummarySrvc.AliasService.IsInitialized(user.ID) {
		log.Printf("Initializing aliases for user '%s'\n", user.ID)
		h.SummarySrvc.AliasService.InitUser(user.ID)
	}

	params := r.URL.Query()
	interval := params.Get("interval")		
	from, err := utils.ParseDate(params.Get("from"))
	if err != nil {
		switch interval {
		case IntervalToday:
			from = utils.StartOfDay()
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

	live := (params.Get("live") != "" && params.Get("live") != "false") || interval == IntervalToday
	to := utils.StartOfDay()
	if live {
		to = time.Now()
	}

	var summary *models.Summary
	cacheKey := getHash([]time.Time{from, to})
	cachedSummary, ok := h.Cache.Get(cacheKey)
	if !ok {
		// Cache Miss
		summary, err = h.SummarySrvc.GetSummary(from, to, user) // 'to' is always constant
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !live {
			h.Cache.Set(cacheKey, summary, cache.DefaultExpiration)
		}
	} else {
		summary = cachedSummary.(*models.Summary)
	}

	utils.RespondJSON(w, http.StatusOK, summary)
}

func getHash(times []time.Time) string {
	digest := md5.New()
	for _, t := range times {
		digest.Write([]byte(strconv.Itoa(int(t.Unix()))))
	}
	return string(digest.Sum(nil))
}
