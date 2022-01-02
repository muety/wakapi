package api

import (
	"net/http"

	"github.com/gorilla/mux"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	su "github.com/muety/wakapi/routes/utils"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
)

type SummaryApiHandler struct {
	config      *conf.Config
	userSrvc    services.IUserService
	summarySrvc services.ISummaryService
}

func NewSummaryApiHandler(userService services.IUserService, summaryService services.ISummaryService) *SummaryApiHandler {
	return &SummaryApiHandler{
		summarySrvc: summaryService,
		userSrvc:    userService,
		config:      conf.Get(),
	}
}

func (h *SummaryApiHandler) RegisterRoutes(router *mux.Router) {
	r := router.PathPrefix("/summary").Subrouter()
	r.Use(
		middlewares.NewAuthenticateMiddleware(h.userSrvc).Handler,
	)
	r.Path("").Methods(http.MethodGet).HandlerFunc(h.Get)
}

// @Summary Retrieve a summary
// @ID get-summary
// @Tags summary
// @Produce json
// @Param interval query string false "Interval identifier" Enums(today, yesterday, week, month, year, 7_days, last_7_days, 30_days, last_30_days, 12_months, last_12_months, any)
// @Param from query string false "Start date (e.g. '2021-02-07')"
// @Param to query string false "End date (e.g. '2021-02-08')"
// @Param recompute query bool false "Whether to recompute the summary from raw heartbeat or use cache"
// @Security ApiKeyAuth
// @Success 200 {object} models.Summary
// @Router /summary [get]
func (h *SummaryApiHandler) Get(w http.ResponseWriter, r *http.Request) {
	summary, err, status := su.LoadUserSummary(h.summarySrvc, r)
	if err != nil {
		w.WriteHeader(status)
		w.Write([]byte(err.Error()))
		return
	}

	utils.RespondJSON(w, r, http.StatusOK, summary)
}
