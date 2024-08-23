package routes

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/models/view"
	su "github.com/muety/wakapi/routes/utils"
	"github.com/muety/wakapi/services"
	"net/http"
	"time"
)

type SummaryHandler struct {
	config       *conf.Config
	userSrvc     services.IUserService
	summarySrvc  services.ISummaryService
	keyValueSrvc services.IKeyValueService
}

func NewSummaryHandler(summaryService services.ISummaryService, userService services.IUserService, keyValueService services.IKeyValueService) *SummaryHandler {
	return &SummaryHandler{
		summarySrvc:  summaryService,
		userSrvc:     userService,
		keyValueSrvc: keyValueService,
		config:       conf.Get(),
	}
}

func (h *SummaryHandler) RegisterRoutes(router chi.Router) {
	r := chi.NewRouter()
	r.Use(middlewares.NewAuthenticateMiddleware(h.userSrvc).
		WithRedirectTarget(defaultErrorRedirectTarget()).
		WithRedirectErrorMessage("unauthorized").Handler,
	)
	r.Get("/", h.GetIndex)

	router.Mount("/summary", r)
}

func (h *SummaryHandler) GetIndex(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	rawQuery := r.URL.RawQuery
	q := r.URL.Query()
	if q.Get("interval") == "" && q.Get("from") == "" {
		// If the PersistentIntervalKey cookie is set, redirect to the correct summary page
		if intervalCookie, _ := r.Cookie(models.PersistentIntervalKey); intervalCookie != nil {
			redirectAddress := fmt.Sprintf("%s/summary?interval=%s", h.config.Server.BasePath, intervalCookie.Value)
			http.Redirect(w, r, redirectAddress, http.StatusFound)
		}

		q.Set("interval", "today")
		r.URL.RawQuery = q.Encode()
	} else if q.Get("interval") != "" {
		// Send a Set-Cookie header to persist the interval
		headerValue := fmt.Sprintf("%s=%s", models.PersistentIntervalKey, q.Get("interval"))
		w.Header().Add("Set-Cookie", headerValue)
	}

	summaryParams, _ := helpers.ParseSummaryParams(r)
	summary, err, status := su.LoadUserSummary(h.summarySrvc, r)
	if err != nil {
		w.WriteHeader(status)
		conf.Log().Request(r).Error("failed to load summary", "error", err)
		templates[conf.SummaryTemplate].Execute(w, h.buildViewModel(r, w).WithError(err.Error()))
		return
	}

	user := middlewares.GetPrincipal(r)
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		templates[conf.SummaryTemplate].Execute(w, h.buildViewModel(r, w).WithError("unauthorized"))
		return
	}

	// user first data
	var firstData time.Time
	firstDataKv := h.keyValueSrvc.MustGetString(fmt.Sprintf("%s_%s", conf.KeyFirstHeartbeat, user.ID))
	if firstDataKv.Value != "" {
		firstData, _ = time.Parse(time.RFC822Z, firstDataKv.Value)
	}

	vm := view.SummaryViewModel{
		SharedLoggedInViewModel: view.SharedLoggedInViewModel{
			SharedViewModel: view.NewSharedViewModel(h.config, nil),
			User:            user,
			ApiKey:          user.ApiKey,
		},
		Summary:             summary,
		SummaryParams:       summaryParams,
		EditorColors:        su.FilterColors(h.config.App.GetEditorColors(), summary.Editors),
		LanguageColors:      su.FilterColors(h.config.App.GetLanguageColors(), summary.Languages),
		OSColors:            su.FilterColors(h.config.App.GetOSColors(), summary.OperatingSystems),
		RawQuery:            rawQuery,
		UserFirstData:       firstData,
		DataRetentionMonths: h.config.App.DataRetentionMonths,
	}

	templates[conf.SummaryTemplate].Execute(w, vm)
}

func (h *SummaryHandler) buildViewModel(r *http.Request, w http.ResponseWriter) *view.SummaryViewModel {
	return su.WithSessionMessages(&view.SummaryViewModel{
		SharedLoggedInViewModel: view.SharedLoggedInViewModel{
			SharedViewModel: view.NewSharedViewModel(h.config, nil),
		},
	}, r, w)
}
