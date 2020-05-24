package routes

import (
	"errors"
	"net/http"
	"time"

	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
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
	cummarySrvc *services.SummaryService
	config      *models.Config
}

func NewSummaryHandler(summaryService *services.SummaryService) *SummaryHandler {
	return &SummaryHandler{
		cummarySrvc: summaryService,
		config:      models.GetConfig(),
	}
}

func (h *SummaryHandler) ApiGet(w http.ResponseWriter, r *http.Request) {
	summary, err, status := loadUserSummary(r, h.cummarySrvc)
	if err != nil {
		w.WriteHeader(status)
		w.Write([]byte(err.Error()))
		return
	}

	utils.RespondJSON(w, http.StatusOK, summary)
}

func (h *SummaryHandler) Index(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	q := r.URL.Query()
	if q.Get("interval") == "" && q.Get("from") == "" {
		q.Set("interval", "today")
		r.URL.RawQuery = q.Encode()
	}

	summary, err, status := loadUserSummary(r, h.cummarySrvc)
	if err != nil {
		w.WriteHeader(status)
		w.Write([]byte(err.Error()))
		return
	}

	vm := models.SummaryViewModel{
		Summary:        summary,
		LanguageColors: utils.FilterLanguageColors(h.config.LanguageColors, summary),
	}

	templates["summary.tpl.html"].Execute(w, vm)
}

func loadUserSummary(r *http.Request, summaryService *services.SummaryService) (*models.Summary, error, int) {
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
			return nil, errors.New("missing 'from' parameter"), http.StatusBadRequest
		}
	}

	live := (params.Get("live") != "" && params.Get("live") != "false") || interval == IntervalToday
	recompute := params.Get("recompute") != "" && params.Get("recompute") != "false"
	to := utils.StartOfDay()
	if live {
		to = time.Now()
	}

	var summary *models.Summary
	summary, err = summaryService.Construct(from, to, user, recompute) // 'to' is always constant
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}

	return summary, nil, http.StatusOK
}
