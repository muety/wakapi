package v1

import (
	"github.com/muety/wakapi/helpers"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	v1 "github.com/muety/wakapi/models/compat/wakatime/v1"
	routeutils "github.com/muety/wakapi/routes/utils"
	"github.com/muety/wakapi/services"
)

type StatusBarViewModel struct {
	CachedAt time.Time        `json:"cached_at"`
	Data     v1.SummariesData `json:"data"`
}

type StatusBarHandler struct {
	config      *conf.Config
	userSrvc    services.IUserService
	summarySrvc services.ISummaryService
}

func NewStatusBarHandler(userService services.IUserService, summaryService services.ISummaryService) *StatusBarHandler {
	return &StatusBarHandler{
		userSrvc:    userService,
		summarySrvc: summaryService,
		config:      conf.Get(),
	}
}

func (h *StatusBarHandler) RegisterRoutes(router *mux.Router) {
	r := router.PathPrefix("").Subrouter()

	r.Use(
		middlewares.NewAuthenticateMiddleware(h.userSrvc).Handler,
	)
	r.Path("/users/{user}/statusbar/{range}").Methods(http.MethodGet).HandlerFunc(h.Get)
	r.Path("/v1/users/{user}/statusbar/{range}").Methods(http.MethodGet).HandlerFunc(h.Get)
	r.Path("/compat/wakatime/v1/users/{user}/statusbar/{range}").Methods(http.MethodGet).HandlerFunc(h.Get)
}

// @Summary Retrieve summary for statusbar
// @Description Mimics https://wakatime.com/api/v1/users/current/statusbar/today. Have no official documentation
// @ID statusbar
// @Tags wakatime
// @Produce json
// @Param user path string true "User ID to fetch data for (or 'current')"
// @Security ApiKeyAuth
// @Success 200 {object} StatusBarViewModel
// @Router /users/{user}/statusbar/today [get]
func (h *StatusBarHandler) Get(w http.ResponseWriter, r *http.Request) {
	user, err := routeutils.CheckEffectiveUser(w, r, h.userSrvc, "current")
	if err != nil {
		return // response was already sent by util function
	}
	var vars = mux.Vars(r)

	rangeParam := vars["range"]
	if rangeParam == "" {
		rangeParam = (*models.IntervalToday)[0]
	}

	err, rangeFrom, rangeTo := helpers.ResolveIntervalRawTZ(rangeParam, user.TZ())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid range"))
		return
	}

	summary, status, err := h.loadUserSummary(user, rangeFrom, rangeTo)
	if err != nil {
		w.WriteHeader(status)
		w.Write([]byte(err.Error()))
		return
	}
	summariesView := v1.NewSummariesFrom([]*models.Summary{summary})
	helpers.RespondJSON(w, r, http.StatusOK, StatusBarViewModel{
		CachedAt: time.Now(),
		Data:     *summariesView.Data[0],
	})
}

func (h *StatusBarHandler) loadUserSummary(user *models.User, start, end time.Time) (*models.Summary, int, error) {
	summaryParams := &models.SummaryParams{
		From:      start,
		To:        end,
		User:      user,
		Recompute: false,
	}

	var retrieveSummary services.SummaryRetriever = h.summarySrvc.Retrieve
	if summaryParams.Recompute {
		retrieveSummary = h.summarySrvc.Summarize
	}

	summary, err := h.summarySrvc.Aliased(summaryParams.From, summaryParams.To, summaryParams.User, retrieveSummary, nil, summaryParams.Recompute)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return summary, http.StatusOK, nil
}
