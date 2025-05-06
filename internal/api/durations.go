package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/datetime"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/internal/utilities"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/services"
)

func (a *APIv1) GetDurations(w http.ResponseWriter, r *http.Request) {
	user, err := utilities.CheckEffectiveUser(w, r, a.services.Users(), "current")
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusUnauthorized, map[string]interface{}{
			"message": http.StatusText(http.StatusUnauthorized),
		})
		return
	}

	params := r.URL.Query()
	dateParam := params.Get("date")
	date, err := time.Parse(conf.SimpleDateFormat, dateParam)
	if err != nil {
		helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid date",
			"error":   err.Error(),
		})
		return
	}

	sliceBy := params.Get("slice_by")
	if sliceBy == "" {
		sliceBy = services.SliceByProject
	} else {
		if _, ok := services.AllowedSliceBy[strings.ToLower(sliceBy)]; !ok {
			helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
				"message": fmt.Sprintf("Invalid slice_by value '%s'. Allowed values are: %v", sliceBy, utilities.MapKeys(services.AllowedSliceBy)),
			})
			return
		}
		sliceBy = strings.ToLower(sliceBy)
	}

	timezone := user.TZ()
	rangeFrom, rangeTo := datetime.BeginOfDay(date.In(timezone)), datetime.EndOfDay(date.In(timezone))

	durations, err := a.services.Duration().Get(rangeFrom, rangeTo, user, &models.Filters{}, sliceBy)

	if err != nil {
		conf.Log().Request(r).Error("Error computing durations", "error", err)
		helpers.RespondJSON(w, r, http.StatusOK, map[string]any{
			"message": "Error computing durations",
			"error":   err.Error(),
		})
		return
	}

	response := models.DurationResult{
		Data:       durations,
		Start:      rangeFrom,
		End:        rangeTo,
		GrandTotal: durations.GrandTotal(),
	}

	helpers.RespondJSON(w, r, http.StatusOK, response)
}
