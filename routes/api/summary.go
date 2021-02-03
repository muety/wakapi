package api

import (
	"github.com/gorilla/mux"
	conf "github.com/muety/wakapi/config"
	su "github.com/muety/wakapi/routes/utils"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"net/http"
)

type SummaryApiHandler struct {
	config      *conf.Config
	summarySrvc services.ISummaryService
}

func NewSummaryApiHandler(summaryService services.ISummaryService) *SummaryApiHandler {
	return &SummaryApiHandler{
		summarySrvc: summaryService,
		config:      conf.Get(),
	}
}

func (h *SummaryApiHandler) RegisterRoutes(router *mux.Router) {
	router.Methods(http.MethodGet).HandlerFunc(h.Get)
}

func (h *SummaryApiHandler) Get(w http.ResponseWriter, r *http.Request) {
	summary, err, status := su.LoadUserSummary(h.summarySrvc, r)
	if err != nil {
		w.WriteHeader(status)
		w.Write([]byte(err.Error()))
		return
	}

	utils.RespondJSON(w, http.StatusOK, summary)
}
