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

// Updated DurationBlock to include fields for all potential slice dimensions
type DurationBlock struct {
	Time     float64 `json:"time"`
	Project  string  `json:"project,omitempty"`
	Language string  `json:"language,omitempty"`
	Entity   string  `json:"entity,omitempty"`
	// Dependencies string `json:"dependencies,omitempty"` // Not in Heartbeat model
	Os       string  `json:"os,omitempty"`
	Editor   string  `json:"editor,omitempty"`
	Category string  `json:"category,omitempty"`
	Machine  string  `json:"machine,omitempty"`
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
	GrandTotal *wakatime.SummariesGrandTotal `json:"grand_total"`
}

// Renamed defaultProject to UnknownValue for broader use
const UnknownValue = "Unknown"
const floatPrecision = 4 // Number of decimal places for rounding durations/times

// Define allowed slice_by values
const (
	SliceByProject  string = "project"
	SliceByEntity   string = "entity"
	SliceByLanguage string = "language"
	// SliceByDependencies string = "dependencies" // Not in Heartbeat model
	SliceByOS       string = "os"
	SliceByEditor   string = "editor"
	SliceByCategory string = "category"
	SliceByMachine  string = "machine"
)

var AllowedSliceBy = map[string]bool{
	SliceByProject:  true,
	SliceByEntity:   true,
	SliceByLanguage: true,
	SliceByOS:       true,
	SliceByEditor:   true,
	SliceByCategory: true,
	SliceByMachine:  true,
}

// round rounds a float64 to a specified number of decimal places.
func round(number float64, precision int) float64 {
	factor := math.Pow(10, float64(precision))
	return math.Round(number*factor) / factor
}

// heartbeatsToMiniDurations calculates the duration from each heartbeat to the next one.
// No changes needed here based on slice_by, as this is per-heartbeat duration.
func heartbeatsToMiniDurations(heartbeats []*models.Heartbeat, timeoutMinutes int) []MiniDurationHeartbeat {
	timeoutDuration := time.Duration(timeoutMinutes) * time.Minute
	miniDurations := make([]MiniDurationHeartbeat, 0, len(heartbeats))

	for i := range heartbeats {
		hb := heartbeats[i] // Use value semantics (copies struct)
		var durationSecs float64 = 0
		var nextHeartbeat *models.Heartbeat = nil

		if i < len(heartbeats)-1 {
			nextHeartbeat = heartbeats[i+1]
		}

		// No need to set defaultProject here anymore, handled when creating DurationBlock

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

// getValueForSlice returns the string value of the Heartbeat field based on the sliceBy parameter.
// This function now also ensures "Unknown" is returned for empty values.
func getValueForSlice(hb *models.Heartbeat, sliceBy string) string {
	var value string
	switch sliceBy {
	case SliceByProject:
		value = hb.Project
	case SliceByEntity:
		value = hb.Entity
	case SliceByLanguage:
		value = hb.Language
	// case SliceByDependencies:
	// value = "" // Not available in Heartbeat
	case SliceByOS:
		value = hb.OperatingSystem
	case SliceByEditor:
		value = hb.Editor
	case SliceByCategory:
		value = hb.Category
	case SliceByMachine:
		value = hb.Machine
	default:
		// Should not happen if sliceBy is validated, but as a fallback
		value = hb.Project // Defaulting value lookup to project
	}

	if value == "" {
		return UnknownValue
	}
	return value
}

// populateDurationBlockFields populates all potential fields in a DurationBlock
// from a MiniDurationHeartbeat, setting "Unknown" for empty values.
func populateDurationBlockFields(block *DurationBlock, item MiniDurationHeartbeat) {
	block.Project = item.Project
	if block.Project == "" {
		block.Project = UnknownValue
	}

	block.Language = item.Language
	if block.Language == "" {
		block.Language = UnknownValue
	}

	block.Entity = item.Entity
	if block.Entity == "" {
		block.Entity = UnknownValue
	}

	block.Os = item.OperatingSystem
	if block.Os == "" {
		block.Os = UnknownValue
	}

	block.Editor = item.Editor
	if block.Editor == "" {
		block.Editor = UnknownValue
	}

	block.Category = item.Category
	if block.Category == "" {
		block.Category = UnknownValue
	}

	block.Machine = item.Machine
	if block.Machine == "" {
		block.Machine = UnknownValue
	}

	// Other fields like Time, Duration, Color are set separately
}

// shouldJoinDuration determines if the current mini-duration heartbeat should be joined
// with the previous one, based on the sliceBy dimension.
func shouldJoinDuration(current MiniDurationHeartbeat, last MiniDurationHeartbeat, timeoutMinutes int, sliceBy string) bool {
	timeoutDuration := time.Duration(timeoutMinutes) * time.Minute

	// Check if the value of the field being sliced by differs (case-insensitive)
	// Use getValueForSlice which now handles the "Unknown" default
	currentSliceValue := getValueForSlice(&current.Heartbeat, sliceBy)
	lastSliceValue := getValueForSlice(&last.Heartbeat, sliceBy)

	if !strings.EqualFold(lastSliceValue, currentSliceValue) {
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

// combineMiniDurations merges consecutive mini-duration heartbeats into final DurationBlocks,
// grouping by the specified sliceBy dimension.
func combineMiniDurations(miniDurations []MiniDurationHeartbeat, timeoutMinutes int, sliceBy string) []DurationBlock {
	if len(miniDurations) == 0 {
		return []DurationBlock{}
	}

	finalDurations := make([]DurationBlock, 0)
	var lastProcessed MiniDurationHeartbeat

	// Initialize with the first block
	if len(miniDurations) > 0 {
		firstHB := miniDurations[0]
		firstBlock := DurationBlock{
			Time:     round(float64(firstHB.Time.T().UnixNano())/1e9, 6), // Use float timestamp
			Duration: round(firstHB.CalculatedDuration, floatPrecision),
			Color:    nil, // Color logic might be needed here or later
		}
		// Populate all fields, using "Unknown" for empty ones
		populateDurationBlockFields(&firstBlock, firstHB)

		if firstBlock.Duration < 0 {
			firstBlock.Duration = 0
		}
		finalDurations = append(finalDurations, firstBlock)
		lastProcessed = firstHB
	}

	for i := 1; i < len(miniDurations); i++ {
		currentItem := miniDurations[i]
		currentBlock := &finalDurations[len(finalDurations)-1] // Pointer to the last block

		// Check if the current item should be joined with the last block based on sliceBy
		if shouldJoinDuration(currentItem, lastProcessed, timeoutMinutes, sliceBy) {
			// Combine: Update the end time of the currentBlock
			itemDuration := time.Duration(round(currentItem.CalculatedDuration, floatPrecision) * float64(time.Second))
			endTime := currentItem.Time.T().Add(itemDuration)

			// Calculate new total duration for the block in seconds (float)
			// The start time of the block is fixed at the time of the first heartbeat in the block.
			startTime := time.Unix(0, int64(round(currentBlock.Time, 9)*1e9)) // Convert block start time back
			newDurationSecs := endTime.Sub(startTime).Seconds()

			currentBlock.Duration = round(newDurationSecs, floatPrecision)
			if currentBlock.Duration < 0 {
				currentBlock.Duration = 0
			}

			// Update all fields in the block to reflect the most recent heartbeat's values,
			// applying the "Unknown" default for empty strings.
			populateDurationBlockFields(currentBlock, currentItem)

		} else {
			// Start a new block
			newBlock := DurationBlock{
				Time:     round(float64(currentItem.Time.T().UnixNano())/1e9, 6), // Use float timestamp
				Duration: round(currentItem.CalculatedDuration, floatPrecision),
				Color:    nil, // Color logic
			}
			// Populate all fields, using "Unknown" for empty ones
			populateDurationBlockFields(&newBlock, currentItem)

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

func ProcessHeartbeats(heartbeats []*models.Heartbeat, start time.Time, end time.Time, timeoutMinutes int, timezone *time.Location, yesterday *models.Heartbeat, tomorrow *models.Heartbeat, sliceBy string) (DurationResult, error) {
	timeoutDuration := time.Duration(timeoutMinutes) * time.Minute

	sort.SliceStable(heartbeats, func(i, j int) bool {
		return heartbeats[i].Time.T().Before(heartbeats[j].Time.T())
	})

	// 3. Handle Boundary Heartbeats
	tempHeartbeats := make([]*models.Heartbeat, 0, len(heartbeats)+2) // Preallocate slice

	// Check if yesterday heartbeat exists and is within timeout of the first heartbeat in rangeFrom/rangeTo
	if yesterday != nil && len(heartbeats) > 0 {
		diff := heartbeats[0].Time.T().Sub(yesterday.Time.T())
		// Important: Check if the slice value matches for yesterday's heartbeat and the first heartbeat of the day
		// Use getValueForSlice which now handles the "Unknown" default
		yesterdaySliceValue := getValueForSlice(yesterday, sliceBy)
		firstDaySliceValue := getValueForSlice(heartbeats[0], sliceBy)

		if diff > 0 && diff < timeoutDuration && strings.EqualFold(yesterdaySliceValue, firstDaySliceValue) {
			// Make a copy and adjust time if it bridges the gap
			yesterdayCopy := *yesterday                   // Create a copy
			yesterdayCopy.Time = models.CustomTime(start) // Set its time to the start of the period
			tempHeartbeats = append(tempHeartbeats, &yesterdayCopy)
		}
	} else if yesterday != nil && len(heartbeats) == 0 && yesterday.Time.T().After(start) && yesterday.Time.T().Before(end) {
		yesterdayCopy := *yesterday // Create a copy
		tempHeartbeats = append(tempHeartbeats, &yesterdayCopy)
	}

	tempHeartbeats = append(tempHeartbeats, heartbeats...) // Add the main heartbeats

	// Check if tomorrow heartbeat exists and is within timeout of the last heartbeat in tempHeartbeats
	if tomorrow != nil && len(tempHeartbeats) > 0 {
		lastHeartbeat := tempHeartbeats[len(tempHeartbeats)-1]
		diff := tomorrow.Time.T().Sub(lastHeartbeat.Time.T())
		// Important: Check if the slice value matches for tomorrow's heartbeat and the last heartbeat of the day
		// Use getValueForSlice which now handles the "Unknown" default
		tomorrowSliceValue := getValueForSlice(tomorrow, sliceBy)
		lastDaySliceValue := getValueForSlice(lastHeartbeat, sliceBy)

		if diff > 0 && diff < timeoutDuration && strings.EqualFold(tomorrowSliceValue, lastDaySliceValue) {
			// Make a copy and adjust time if it bridges the gap
			tomorrowCopy := *tomorrow                  // Create a copy
			tomorrowCopy.Time = models.CustomTime(end) // Set its time to the end of the period
			tempHeartbeats = append(tempHeartbeats, &tomorrowCopy)
		}
	} else if tomorrow != nil && len(tempHeartbeats) == 0 && tomorrow.Time.T().After(start) && tomorrow.Time.T().Before(end) {
		tomorrowCopy := *tomorrow // Create a copy
		tempHeartbeats = append(tempHeartbeats, &tomorrowCopy)
	}

	// Sort again after potentially adding boundary heartbeats
	sort.SliceStable(tempHeartbeats, func(i, j int) bool {
		return tempHeartbeats[i].Time.T().Before(tempHeartbeats[j].Time.T())
	})

	heartbeats = tempHeartbeats // Use the potentially expanded list

	// 4. Run Wakatime Processing Steps
	miniDurations := heartbeatsToMiniDurations(heartbeats, timeoutMinutes)
	// Skipping external durations step
	finalDurations := combineMiniDurations(miniDurations, timeoutMinutes, sliceBy) // Pass sliceBy here

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

	// --- Handle slice_by parameter ---
	sliceBy := params.Get("slice_by")
	if sliceBy == "" {
		sliceBy = SliceByProject // Default value
	} else {
		// Validate the slice_by value
		if _, ok := AllowedSliceBy[strings.ToLower(sliceBy)]; !ok {
			helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
				"message": fmt.Sprintf("Invalid slice_by value '%s'. Allowed values are: %v", sliceBy, utilities.MapKeys(AllowedSliceBy)),
			})
			return
		}
		// Use lowercase for consistent lookup
		sliceBy = strings.ToLower(sliceBy)
	}
	// --- End slice_by handling ---

	timezone := user.TZ()
	rangeFrom, rangeTo := datetime.BeginOfDay(date.In(timezone)), datetime.EndOfDay(date.In(timezone))

	var lastHeartbeatFromYesterday models.Heartbeat

	// We need yesterday's last heartbeat and tomorrow's first one for gap calculation,
	// REGARDLESS of their project/language/etc. because they determine the boundary
	// duration *before* we apply the slice_by logic for joining. The joining logic
	// (`shouldJoinDuration`) will then check if the slice_by value matches across the boundary.

	result := a.db.
		Where("user_id = ? AND time < ?", user.ID, rangeFrom).
		Order("time DESC").
		Limit(1).
		Find(&lastHeartbeatFromYesterday)

	var yesterdayHB *models.Heartbeat = nil
	if result.Error == nil && lastHeartbeatFromYesterday.ID != 0 { // Check if a record was actually found
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
	if result.Error == nil && firstHeartbeatFromTomorrow.ID != 0 { // Check if a record was actually found
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

	durations, err := ProcessHeartbeats(
		heartbeats,
		rangeFrom,
		rangeTo,
		15, // Assuming a default timeout of 15 minutes as in your code
		timezone,
		yesterdayHB, // Pass nil if no heartbeat was found
		tomorrowHB,  // Pass nil if no heartbeat was found
		sliceBy,     // Pass the determined slice_by parameter
	)

	if err != nil {
		conf.Log().Request(r).Error("Error computing durations", "error", err)
		helpers.RespondJSON(w, r, http.StatusOK, map[string]interface{}{ // StatusOK might be intentional here, or should it be 500? Sticking to original code.
			"message": "Error computing durations",
			"error":   err.Error(),
		})
		return
	}

	helpers.RespondJSON(w, r, http.StatusOK, durations)
}
