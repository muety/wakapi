package routes

import (
	"crypto/md5"
	"net/http"
	"strconv"
	"time"

	"github.com/n1try/wakapi/models"
	"github.com/n1try/wakapi/services"
	"github.com/n1try/wakapi/utils"
	cache "github.com/patrickmn/go-cache"
	uuid "github.com/satori/go.uuid"
)

const (
	IntervalToday     string = "today"
	IntervalLastDay   string = "day"
	IntervalLastWeek  string = "week"
	IntervalLastMonth string = "month"
	IntervalLastYear  string = "year"
	IntervalAny       string = "any"
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
		case IntervalAny:
			from = time.Time{}
		default:
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing 'from' parameter"))
			return
		}
	}

	live := (params.Get("live") != "" && params.Get("live") != "false") || interval == IntervalToday
	recompute := params.Get("recompute") != "" && params.Get("recompute") != "false"
	to := utils.StartOfDay()
	if live {
		to = time.Now()
	}

	var summary *models.Summary
	var cacheKey string
	if !recompute {
		cacheKey = getHash([]time.Time{from, to}, user)
	} else {
		cacheKey = uuid.NewV4().String()
	}
	if cachedSummary, ok := h.Cache.Get(cacheKey); !ok {
		// Cache Miss
		summary, err = h.SummarySrvc.Construct(from, to, user, recompute) // 'to' is always constant
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !live && !recompute {
			h.Cache.Set(cacheKey, summary, cache.DefaultExpiration)
		}
	} else {
		summary = cachedSummary.(*models.Summary)
	}

	utils.RespondJSON(w, http.StatusOK, summary)
}

func getHash(times []time.Time, user *models.User) string {
	digest := md5.New()
	for _, t := range times {
		digest.Write([]byte(strconv.Itoa(int(t.Unix()))))
	}
	digest.Write([]byte(user.ID))
	return string(digest.Sum(nil))
}
