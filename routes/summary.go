package routes

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path"
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
	SummarySrvc *services.SummaryService
	Initialized bool
	templates   map[string]*template.Template
}

func (m *SummaryHandler) Init() {
	m.loadTemplates()
	m.Initialized = true
}

func (m *SummaryHandler) loadTemplates() {
	tplPath := "views"
	templates := template.New("").Funcs(template.FuncMap{
		"json": utils.Json,
		"date": utils.FormatDateHuman,
	})
	m.templates = make(map[string]*template.Template)

	files, err := ioutil.ReadDir(tplPath)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		tplName := file.Name()
		if file.IsDir() || path.Ext(tplName) != ".html" {
			continue
		}

		tpl, err := templates.New(tplName).ParseFiles(fmt.Sprintf("%s/%s", tplPath, tplName))
		if err != nil {
			panic(err)
		}

		m.templates[tplName] = tpl
	}
}

func (h *SummaryHandler) ApiGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if !h.Initialized {
		h.Init()
	}

	summary, err, status := loadUserSummary(r, h.SummarySrvc)
	if err != nil {
		w.WriteHeader(status)
		w.Write([]byte(err.Error()))
		return
	}

	utils.RespondJSON(w, http.StatusOK, summary)
}

func (h *SummaryHandler) Index(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if !h.Initialized {
		h.Init()
	}

	if h.SummarySrvc.Config.IsDev() {
		h.loadTemplates()
	}

	q := r.URL.Query()
	if q.Get("interval") == "" && q.Get("from") == "" {
		q.Set("interval", "today")
		r.URL.RawQuery = q.Encode()
	}

	summary, err, status := loadUserSummary(r, h.SummarySrvc)
	if err != nil {
		w.WriteHeader(status)
		w.Write([]byte(err.Error()))
		return
	}

	vm := models.SummaryViewModel{
		Summary:        summary,
		LanguageColors: utils.FilterLanguageColors(h.SummarySrvc.Config.LanguageColors, summary),
	}

	h.templates["index.tpl.html"].Execute(w, vm)
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
