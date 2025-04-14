package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/internal/utilities"

	"github.com/muety/wakapi/models"
	v1 "github.com/muety/wakapi/models/compat/wakatime/v1"
)

type StatusBarViewModel struct {
	CachedAt time.Time        `json:"cached_at"`
	Data     v1.SummariesData `json:"data"`
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
func (a *APIv1) GetStatusBarRange(w http.ResponseWriter, r *http.Request) {
	user, err := utilities.CheckEffectiveUser(w, r, a.services.Users(), "current")
	if err != nil {
		return // response was already sent by util function
	}

	rangeParam := chi.URLParam(r, "range")
	if rangeParam == "" {
		rangeParam = (*models.IntervalToday)[0]
	}

	err, rangeFrom, rangeTo := helpers.ResolveIntervalRawTZ(rangeParam, user.TZ())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid range"))
		return
	}

	summaryParams := &models.SummaryParams{
		From:      rangeFrom,
		To:        rangeTo,
		User:      user,
		Recompute: false,
	}

	summary, err, status := a.loadUserSummary(summaryParams, &models.Filters{})
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
