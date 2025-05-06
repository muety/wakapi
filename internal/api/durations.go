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

	var lastHeartbeatFromYesterday models.Heartbeat
	result := a.db.
		Where("user_id = ? AND time < ?", user.ID, rangeFrom).
		Order("time DESC").
		Limit(1).
		Find(&lastHeartbeatFromYesterday)

	var yesterdayHB *models.Heartbeat = nil
	if result.Error == nil && lastHeartbeatFromYesterday.ID != 0 {
		yesterdayHB = &lastHeartbeatFromYesterday
	} else if result.Error != nil && result.Error.Error() != "record not found" {
		conf.Log().Request(r).Error("Failed to retrieve last heartbeat from yesterday", "error", result.Error)
		helpers.RespondJSON(w, r, http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to retrieve last heartbeat from yesterday",
			"error":   result.Error.Error(),
		})
		return
	}

	startOfTomorrow := rangeTo.Add(time.Second)
	var firstHeartbeatFromTomorrow models.Heartbeat
	result = a.db.
		Where("user_id = ? AND time >= ?", user.ID, startOfTomorrow).
		Order("time ASC").
		Limit(1).
		Find(&firstHeartbeatFromTomorrow)

	var tomorrowHB *models.Heartbeat = nil
	if result.Error == nil && firstHeartbeatFromTomorrow.ID != 0 {
		tomorrowHB = &firstHeartbeatFromTomorrow
	} else if result.Error != nil && result.Error.Error() != "record not found" {
		conf.Log().Request(r).Error("Failed to retrieve first heartbeat from tomorrow", "error", result.Error)
		helpers.RespondJSON(w, r, http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to retrieve first heartbeat from tomorrow",
			"error":   result.Error.Error(),
		})
		return
	}

	heartbeats, err := a.services.Heartbeat().GetAllWithin(rangeFrom, rangeTo, user)
	if err != nil {
		errMessage := "Failed to retrieve heartbeats"
		conf.Log().Request(r).Error(errMessage, "error", err)
		helpers.RespondJSON(w, r, http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to retrieve heartbeats",
			"error":   err.Error(),
		})
		return
	}

	args := models.ProcessHeartbeatsArgs{
		Heartbeats:             heartbeats,
		Start:                  rangeFrom,
		End:                    rangeTo,
		User:                   user,
		LastHeartbeatYesterday: yesterdayHB,
		FirstHeartbeatTomorrow: tomorrowHB,
		SliceBy:                sliceBy,
	}

	durations, err := services.ProcessHeartbeats(args)

	if err != nil {
		conf.Log().Request(r).Error("Error computing durations", "error", err)
		helpers.RespondJSON(w, r, http.StatusOK, map[string]any{
			"message": "Error computing durations",
			"error":   err.Error(),
		})
		return
	}

	helpers.RespondJSON(w, r, http.StatusOK, durations)
}

func (a *APIv1) GetDurationsV2(w http.ResponseWriter, r *http.Request) {
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

	timezone := user.TZ()
	rangeFrom, rangeTo := datetime.BeginOfDay(date.In(timezone)), datetime.EndOfDay(date.In(timezone))

	heartbeats, err := a.services.Heartbeat().GetAllWithin(rangeFrom, rangeTo, user)
	if err != nil {
		errMessage := "Failed to retrieve heartbeats"
		conf.Log().Request(r).Error(errMessage, "error", err)
		helpers.RespondJSON(w, r, http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to retrieve heartbeats",
			"error":   err.Error(),
		})
		return
	}

	userHeartbeatsTimeout := user.HeartbeatsTimeout()

	durations, err := a.services.Duration().Get(rangeFrom, rangeTo, user, &models.Filters{})
	minidurations := services.HeartbeatsToMiniDurations(heartbeats, userHeartbeatsTimeout)
	finalDurations := services.CombineMiniDurations(minidurations, userHeartbeatsTimeout, "entity")

	if err != nil {
		conf.Log().Request(r).Error("Error computing durations", "error", err)
		helpers.RespondJSON(w, r, http.StatusOK, map[string]any{
			"message": "Error computing durations",
			"error":   err.Error(),
		})
		return
	}

	var totalDuration time.Duration
	for _, duration := range durations {
		totalDuration += duration.Duration
	}

	result := services.DurationResult{
		Data:     finalDurations,
		Start:    rangeFrom,
		End:      rangeTo,
		Timezone: timezone.String(),
	}

	total := result.TotalTime()

	helpers.RespondJSON(w, r, http.StatusOK, map[string]any{
		"durations": durations,
		// "mapping":            mapping,
		"finalDurations":     finalDurations,
		"finalDurationsSize": len(finalDurations),
		"durationsSize":      len(durations),
		"myTotal":            total.Seconds(),
		"totalDuration":      totalDuration.Seconds(),
	})
}
