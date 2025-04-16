package api

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/datetime"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/internal/utilities"
	"github.com/muety/wakapi/models"
	wakatime "github.com/muety/wakapi/models/compat/wakatime/v1"
)

type MiniDurationHeartbeat struct {
	models.Heartbeat           // Embed original heartbeat fields
	CalculatedDuration float64 // Duration calculated in heartbeatsToMiniDurations step (seconds)
}

type DurationBlock struct {
	Time     float64 `json:"time"`
	Project  string  `json:"project"`
	Duration float64 `json:"duration"`
	Color    *string `json:"color"`
}

func (d *DurationResult) ComputeTotalTimeFromDurations() float64 {
	total := 0.0
	for _, item := range d.Data {
		total += item.Duration
	}
	return total
}

func (d *DurationResult) TotalTime() time.Duration {
	return time.Duration(d.ComputeTotalTimeFromDurations()) * time.Second
}

// Overall result structure matching Wakatime API response
type DurationResult struct {
	Data       []DurationBlock               `json:"data"`
	Start      time.Time                     `json:"start"`    // ISO 8601 format string
	End        time.Time                     `json:"end"`      // ISO 8601 format string
	Timezone   string                        `json:"timezone"` // Timezone name (e.g., "Africa/Accra")
	GrandTotal *wakatime.SummariesGrandTotal `json:"grand_totel"`
}

const defaultProject = "Unknown Project"
const floatPrecision = 4 // Number of decimal places for rounding durations/times

// round rounds a float64 to a specified number of decimal places.
func round(number float64, precision int) float64 {
	factor := math.Pow(10, float64(precision))
	return math.Round(number*factor) / factor
}

// heartbeatsToMiniDurations calculates the duration from each heartbeat to the next one.
func heartbeatsToMiniDurations(heartbeats []*models.Heartbeat, timeoutMinutes int) []MiniDurationHeartbeat {
	timeoutDuration := time.Duration(timeoutMinutes) * time.Minute
	miniDurations := make([]MiniDurationHeartbeat, 0, len(heartbeats))

	for i := 0; i < len(heartbeats); i++ {
		hb := heartbeats[i] // Use value semantics (copies struct)
		var durationSecs float64 = 0
		var nextHeartbeat *models.Heartbeat = nil

		if i < len(heartbeats)-1 {
			nextHeartbeat = heartbeats[i+1]
		}

		if hb.Project == "" {
			hb.Project = defaultProject
		}

		if nextHeartbeat != nil {
			// Use time.Time subtraction for accurate duration calculation
			diff := nextHeartbeat.Time.T().Sub(hb.Time.T())
			if diff > 0 && diff < timeoutDuration { // Check positive diff and within timeout
				durationSecs = diff.Seconds()
			}
		}

		miniDurations = append(miniDurations, MiniDurationHeartbeat{
			Heartbeat:          *hb,
			CalculatedDuration: round(durationSecs, floatPrecision),
		})
	}
	return miniDurations
}

// shouldJoinDuration determines if the current mini-duration heartbeat should be joined.
func shouldJoinDuration(current MiniDurationHeartbeat, last MiniDurationHeartbeat, timeoutMinutes int) bool {
	timeoutDuration := time.Duration(timeoutMinutes) * time.Minute

	// Check if projects differ (case-insensitive)
	if !strings.EqualFold(last.Project, current.Project) {
		return false
	}

	// Calculate the effective end time of the last heartbeat's calculated duration
	lastDuration := time.Duration(round(last.CalculatedDuration, floatPrecision) * float64(time.Second))
	lastTimeEnd := last.Time.T().Add(lastDuration)

	// Calculate the gap between the end of the last and the start of the current
	gap := current.Time.T().Sub(lastTimeEnd)

	// Join if gap is non-negative (current starts at or after last ended) AND within timeout
	if gap >= 0 && gap <= timeoutDuration {
		return true
	}

	// Handle overlapping case (current starts *before* last one ended) - always join
	if gap < 0 {
		return true
	}

	return false
}

// combineMiniDurations merges consecutive mini-duration heartbeats into final DurationBlocks.
func combineMiniDurations(miniDurations []MiniDurationHeartbeat, timeoutMinutes int) []DurationBlock {
	if len(miniDurations) == 0 {
		return []DurationBlock{}
	}

	finalDurations := make([]DurationBlock, 0)
	var lastProcessed MiniDurationHeartbeat

	// Initialize with the first block
	if len(miniDurations) > 0 {
		firstHB := miniDurations[0]
		firstBlock := DurationBlock{
			Time:     round(float64(firstHB.Time.T().UnixNano())/1e9, 6),
			Project:  firstHB.Project,
			Duration: round(firstHB.CalculatedDuration, floatPrecision),
			Color:    nil,
		}
		if firstBlock.Duration < 0 {
			firstBlock.Duration = 0
		}
		finalDurations = append(finalDurations, firstBlock)
		lastProcessed = firstHB
	}

	for i := 1; i < len(miniDurations); i++ {
		currentItem := miniDurations[i]
		currentBlock := &finalDurations[len(finalDurations)-1] // Pointer to the last block

		if shouldJoinDuration(currentItem, lastProcessed, timeoutMinutes) {
			// Combine: Update the end time of the currentBlock
			itemDuration := time.Duration(round(currentItem.CalculatedDuration, floatPrecision) * float64(time.Second))
			endTime := currentItem.Time.T().Add(itemDuration)

			// Calculate new total duration for the block in seconds (float)
			startTime := time.Unix(0, int64(round(currentBlock.Time, 9)*1e9)) // Convert block start time back
			newDurationSecs := endTime.Sub(startTime).Seconds()

			currentBlock.Duration = round(newDurationSecs, floatPrecision)
			if currentBlock.Duration < 0 {
				currentBlock.Duration = 0
			}

		} else {
			// Start a new block
			newBlock := DurationBlock{
				Time:     round(float64(currentItem.Time.T().UnixNano())/1e9, 6), // Use float timestamp
				Project:  currentItem.Project,
				Duration: round(currentItem.CalculatedDuration, floatPrecision),
				Color:    nil,
			}
			if newBlock.Duration < 0 {
				newBlock.Duration = 0
			}
			finalDurations = append(finalDurations, newBlock)
		}
		// Update lastProcessed to the item from this iteration for the next check
		lastProcessed = currentItem
	}

	return finalDurations
}

func ProcessHeartbeats(heartbeats []*models.Heartbeat, start time.Time, end time.Time, timeoutMinutes int, timezone *time.Location, yesterday *models.Heartbeat, tomorrow *models.Heartbeat) (DurationResult, error) {
	timeoutDuration := time.Duration(timeoutMinutes) * time.Minute

	sort.SliceStable(heartbeats, func(i, j int) bool {
		return heartbeats[i].Time.T().Before(heartbeats[j].Time.T())
	})

	// 3. Handle Boundary Heartbeats
	tempHeartbeats := make([]*models.Heartbeat, 0, len(heartbeats)+2) // Preallocate slice

	if yesterday != nil && len(heartbeats) > 0 {
		// Make a copy of yesterday's heartbeat
		yesterdayCopy := yesterday
		diff := heartbeats[0].Time.T().Sub(yesterdayCopy.Time.T())
		if diff > 0 && diff < timeoutDuration { // Check positive diff
			// Set its time to the start of the period
			yesterdayCopy.Time = models.CustomTime(start)
			tempHeartbeats = append(tempHeartbeats, yesterdayCopy)
		}
	}

	tempHeartbeats = append(tempHeartbeats, heartbeats...) // Add the main heartbeats

	if tomorrow != nil && len(tempHeartbeats) > 0 {
		lastHeartbeat := tempHeartbeats[len(tempHeartbeats)-1]
		// Make a copy of tomorrow's heartbeat
		tomorrowCopy := tomorrow
		diff := tomorrowCopy.Time.T().Sub(lastHeartbeat.Time.T())
		if diff > 0 && diff < timeoutDuration { // Check positive diff
			// Set its time to the end of the period
			tomorrowCopy.Time = models.CustomTime(end)
			tempHeartbeats = append(tempHeartbeats, tomorrowCopy)
		}
	}

	heartbeats = tempHeartbeats // Use the potentially expanded list

	// 4. Run Wakatime Processing Steps
	miniDurations := heartbeatsToMiniDurations(heartbeats, timeoutMinutes)
	// Skipping external durations step
	finalDurations := combineMiniDurations(miniDurations, timeoutMinutes)

	// 5. Construct Final Result
	result := DurationResult{
		Data:     finalDurations,
		Start:    start,
		End:      end,
		Timezone: timezone.String(),
	}

	total := result.TotalTime()
	totalHrs, totalMins := int(total.Hours()), int((total - time.Duration(total.Hours())*time.Hour).Minutes())

	result.GrandTotal = &wakatime.SummariesGrandTotal{
		Digital:      fmt.Sprintf("%d:%d", totalHrs, totalMins),
		Hours:        totalHrs,
		Minutes:      totalMins,
		Text:         helpers.FmtWakatimeDuration(total),
		TotalSeconds: total.Seconds(),
	}

	return result, nil
}

func (a *APIv1) GetDurations(w http.ResponseWriter, r *http.Request) {
	user, err := utilities.CheckEffectiveUser(w, r, a.services.Users(), "current")
	if err != nil {
		return // response was already sent by util function
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

	var lastHeartbeatFromYesterday models.Heartbeat

	result := a.db.
		Where("user_id = ? AND time < ?", user.ID, rangeFrom).
		Order("time DESC").
		Limit(1).
		Find(&lastHeartbeatFromYesterday)

	if result.Error != nil {
		helpers.RespondJSON(w, r, http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to retrieve last heartbeat from yesterday",
			"error":   result.Error,
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

	if result.Error != nil {
		helpers.RespondJSON(w, r, http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to retrieve first heartbeat from tomorrow",
			"error":   result.Error,
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

	durations, err := ProcessHeartbeats(
		heartbeats,
		rangeFrom,
		rangeTo,
		15,
		timezone,
		&lastHeartbeatFromYesterday,
		&firstHeartbeatFromTomorrow,
	)

	if err != nil {
		helpers.RespondJSON(w, r, http.StatusOK, map[string]interface{}{
			"message": "Error computing durations",
			"error":   err.Error(),
		})
		return
	}

	helpers.RespondJSON(w, r, http.StatusOK, durations)
}
