package routes

import (
	"github.com/gorilla/mux"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/models/view"
	su "github.com/muety/wakapi/routes/utils"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"net/http"
)

type SummaryHandler struct {
	config      *conf.Config
	userSrvc    services.IUserService
	summarySrvc services.ISummaryService
}

func NewSummaryHandler(summaryService services.ISummaryService, userService services.IUserService) *SummaryHandler {
	return &SummaryHandler{
		summarySrvc: summaryService,
		userSrvc:    userService,
		config:      conf.Get(),
	}
}

func (h *SummaryHandler) RegisterRoutes(router *mux.Router) {
	r := router.PathPrefix("/summary").Subrouter()
	r.Use(
		middlewares.NewAuthenticateMiddleware(h.userSrvc).WithRedirectTarget(defaultErrorRedirectTarget()).Handler,
	)
	r.Methods(http.MethodGet).HandlerFunc(h.GetIndex)
}

func (h *SummaryHandler) GetIndex(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	rawQuery := r.URL.RawQuery
	q := r.URL.Query()
	if q.Get("interval") == "" && q.Get("from") == "" {
		q.Set("interval", "today")
		r.URL.RawQuery = q.Encode()
	}

	summary, err, status := su.LoadUserSummary(h.summarySrvc, r)
	if err != nil {
		w.WriteHeader(status)
		templates[conf.SummaryTemplate].Execute(w, h.buildViewModel(r).WithError(err.Error()))
		return
	}

	user := r.Context().Value(models.UserKey).(*models.User)
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		templates[conf.SummaryTemplate].Execute(w, h.buildViewModel(r).WithError("unauthorized"))
		return
	}

	vm := models.SummaryViewModel{
		Summary:        summary,
		User:           user,
		LanguageColors: utils.FilterColors(h.config.App.GetLanguageColors(), summary.Languages),
		EditorColors:   utils.FilterColors(h.config.App.GetEditorColors(), summary.Editors),
		OSColors:       utils.FilterColors(h.config.App.GetOSColors(), summary.OperatingSystems),
		ApiKey:         user.ApiKey,
		RawQuery:       rawQuery,
	}

	templates[conf.SummaryTemplate].Execute(w, vm)
}

func (h *SummaryHandler) buildViewModel(r *http.Request) *view.SummaryViewModel {
	return &view.SummaryViewModel{
		Success: r.URL.Query().Get("success"),
		Error:   r.URL.Query().Get("error"),
	}
}
