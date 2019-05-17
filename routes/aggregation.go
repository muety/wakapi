package routes

import (
	"net/http"
	"time"

	"github.com/n1try/wakapi/models"
	"github.com/n1try/wakapi/services"
	"github.com/n1try/wakapi/utils"
)

type AggregationHandler struct {
	AggregationSrvc *services.AggregationService
}

func (h *AggregationHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(415)
		return
	}

	user := r.Context().Value(models.UserKey).(*models.User)
	params := r.URL.Query()
	from, err := utils.ParseDate(params.Get("from"))
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("Missing 'from' parameter"))
		return
	}

	to, err := utils.ParseDate(params.Get("to"))
	if err != nil {
		to = time.Now()
	}

	aggregations, err := h.AggregationSrvc.FindOrAggregate(from, to, user)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	for i := 0; i < len(aggregations); i++ {
		if err := h.AggregationSrvc.SaveAggregation(aggregations[i]); err != nil {
			w.WriteHeader(500)
			return
		}
	}

	w.WriteHeader(200)
}
