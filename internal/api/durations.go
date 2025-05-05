package api

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/datetime"
	"github.com/duke-git/lancet/v2/mathutil"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/internal/utilities"
	"github.com/muety/wakapi/models"
	wakatime "github.com/muety/wakapi/models/compat/wakatime/v1"
	"github.com/muety/wakapi/services"
)

func NewSummariesGrandTotal(originalDuration time.Duration) *wakatime.SummariesGrandTotal {
	d := originalDuration.Round(time.Minute)

	totalSeconds := originalDuration.Seconds()
	hours := int(d / time.Hour)
	minutes := int((d % time.Hour) / time.Minute)

	digital := fmt.Sprintf("%02d:%02d", hours, minutes)
	text := fmt.Sprintf("%d hrs %d mins", hours, minutes)

	return &wakatime.SummariesGrandTotal{
		Digital:      digital,
		Hours:        hours,
		Minutes:      minutes,
		Text:         text,
		TotalSeconds: totalSeconds,
	}
}

type MiniDurationHeartbeat struct {
	models.Heartbeat                 // Embed original heartbeat fields
	CalculatedDuration time.Duration // Duration calculated in heartbeatsToMiniDurations step (seconds)
}

// Updated DurationBlock with omitempty tags.
type DurationBlock struct {
	Time         float64       `json:"time"`
	Project      string        `json:"project,omitempty"`
	Language     string        `json:"language,omitempty"`
	Entity       string        `json:"entity,omitempty"`
	Os           string        `json:"os,omitempty"`
	Editor       string        `json:"editor,omitempty"`
	Category     string        `json:"category,omitempty"`
	Machine      string        `json:"machine,omitempty"`
	DurationSecs float64       `json:"duration"`
	Duration     time.Duration `json:"-"`
	Color        *string       `json:"color"`
}

type ProcessHeartbeatsArgs struct {
	Heartbeats             []*models.Heartbeat
	Start                  time.Time
	End                    time.Time
	TimeoutMinutes         int
	Timezone               *time.Location
	LastHeartbeatYesterday *models.Heartbeat
	FirstHeartbeatTomorrow *models.Heartbeat
	SliceBy                string
}

func (d *DurationBlock) AddDurationSecs() {
	d.DurationSecs = d.Duration.Seconds()
}

func (d *DurationResult) ComputeTotalTimeFromDurations() time.Duration {
	var total time.Duration
	for _, item := range d.Data {
		total += item.Duration
	}
	return total
}

func (d *DurationResult) TotalTime() time.Duration {
	return time.Duration(d.ComputeTotalTimeFromDurations()) //* time.Second
}

// Overall result structure matching Wakatime API response
type DurationResult struct {
	Data       []DurationBlock               `json:"data"`
	Start      time.Time                     `json:"start"`
	End        time.Time                     `json:"end"`
	Timezone   string                        `json:"timezone"`
	GrandTotal *wakatime.SummariesGrandTotal `json:"grand_total"`
}

const UnknownValue = "Unknown"

const (
	SliceByProject  string = "project"
	SliceByEntity   string = "entity"
	SliceByLanguage string = "language"
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

func round(number float64, precision int) float64 {
	factor := math.Pow(10, float64(precision))
	return math.Round(number*factor) / factor
}

func heartbeatsToMiniDurations(heartbeats []*models.Heartbeat, timeoutMinutes int) []MiniDurationHeartbeat {
	timeoutDuration := time.Duration(timeoutMinutes) * time.Minute
	miniDurations := make([]MiniDurationHeartbeat, 0, len(heartbeats))

	for i := range heartbeats {
		hb := heartbeats[i]
		var duration time.Duration
		var nextHeartbeat *models.Heartbeat = nil

		if i < len(heartbeats)-1 {
			nextHeartbeat = heartbeats[i+1]
		}

		if nextHeartbeat != nil {
			diff := nextHeartbeat.Time.T().Sub(hb.Time.T())
			if diff > 0 && diff < timeoutDuration {
				duration += diff
			}
		}

		miniDurations = append(miniDurations, MiniDurationHeartbeat{
			Heartbeat:          *hb,
			CalculatedDuration: duration,
		})
	}
	return miniDurations
}

func getValueForSlice(hb *models.Heartbeat, sliceBy string) string {
	var value string
	switch sliceBy {
	case SliceByProject:
		value = hb.Project
	case SliceByEntity:
		value = hb.Entity
	case SliceByLanguage:
		value = hb.Language
	case SliceByOS:
		value = hb.OperatingSystem
	case SliceByEditor:
		value = hb.Editor
	case SliceByCategory:
		value = hb.Category
	case SliceByMachine:
		value = hb.Machine
	default:
		value = hb.Project
	}

	if value == "" {
		return UnknownValue
	}
	return value
}

func populateDurationBlockFields(block *DurationBlock, item MiniDurationHeartbeat, sliceBy string, respectSliceBy bool) {
	// Helper function to assign value or default to UnknownValue if empty
	setField := func(field *string, value string, fieldName string) {
		if fieldName != "project" && respectSliceBy {
			if !strings.EqualFold(sliceBy, fieldName) {
				return
			}
		}

		// Assign the value to the field or UnknownValue if it's empty
		if value == "" {
			*field = UnknownValue
		} else {
			*field = value
		}
	}

	// Set fields using the helper function with lowercase field names
	setField(&block.Project, item.Project, "project")
	setField(&block.Language, item.Language, "language")
	setField(&block.Entity, item.Entity, "entity")
	setField(&block.Os, item.OperatingSystem, "os")
	setField(&block.Editor, item.Editor, "editor")
	setField(&block.Category, item.Category, "category")
	setField(&block.Machine, item.Machine, "machine")
}

func shouldJoinDuration(current MiniDurationHeartbeat, last MiniDurationHeartbeat, timeoutMinutes int, sliceBy string) bool {
	timeoutDuration := time.Duration(timeoutMinutes) * time.Minute

	currentSliceValue := getValueForSlice(&current.Heartbeat, sliceBy)
	lastSliceValue := getValueForSlice(&last.Heartbeat, sliceBy)

	if !strings.EqualFold(lastSliceValue, currentSliceValue) {
		return false
	}

	lastDuration := last.CalculatedDuration //time.Duration(round(last.CalculatedDuration, floatPrecision) * float64(time.Second))
	lastTimeEnd := last.Time.T().Add(lastDuration)
	gap := current.Time.T().Sub(lastTimeEnd)

	if gap >= 0 && gap <= timeoutDuration {
		return true
	}

	if gap < 0 {
		return true
	}

	return false
}

func ShouldJoinDurationBasedOnHash(current MiniDurationHeartbeat, last MiniDurationHeartbeat, timeoutMinutes int, sliceBy string) bool {
	timeoutDuration := time.Duration(timeoutMinutes) * time.Minute

	currentEntityHash := models.NewDurationFromHeartbeat(&current.Heartbeat).WithEntityIgnored().Hashed().GroupHash
	lastEntityHash := models.NewDurationFromHeartbeat(&last.Heartbeat).WithEntityIgnored().Hashed().GroupHash

	if lastEntityHash != currentEntityHash {
		return false
	}

	lastDuration := last.CalculatedDuration //time.Duration(round(last.CalculatedDuration, floatPrecision) * float64(time.Second))
	lastTimeEnd := last.Time.T().Add(lastDuration)
	gap := current.Time.T().Sub(lastTimeEnd)

	if gap >= 0 && gap <= timeoutDuration {
		return true
	}

	if gap < 0 {
		return true
	}

	return false
}

func combineMiniDurations(miniDurations []MiniDurationHeartbeat, timeoutMinutes int, sliceBy string) []DurationBlock {
	if len(miniDurations) == 0 {
		return []DurationBlock{}
	}

	finalDurations := make([]DurationBlock, 0)
	var lastProcessed MiniDurationHeartbeat

	firstHB := miniDurations[0]
	firstBlock := DurationBlock{
		Time:     round(float64(firstHB.Time.T().UnixNano())/1e9, 6),
		Duration: firstHB.CalculatedDuration,
		Color:    nil,
	}

	populateDurationBlockFields(&firstBlock, firstHB, sliceBy, true)

	if firstBlock.Duration < 0 {
		firstBlock.Duration = 0
	}
	finalDurations = append(finalDurations, firstBlock)
	lastProcessed = firstHB

	for i := 1; i < len(miniDurations); i++ {
		currentItem := miniDurations[i]
		currentBlock := &finalDurations[len(finalDurations)-1]

		if shouldJoinDuration(currentItem, lastProcessed, timeoutMinutes, sliceBy) {
			itemDuration := currentItem.CalculatedDuration
			endTime := currentItem.Time.T().Add(itemDuration)
			startTime := time.Unix(0, int64(round(currentBlock.Time, 9)*1e9))
			newDuration := endTime.Sub(startTime)

			currentBlock.Duration = max(newDuration, time.Duration(0))
			currentBlock.DurationSecs = currentBlock.Duration.Seconds()

			populateDurationBlockFields(currentBlock, currentItem, sliceBy, true)

		} else {
			newBlock := DurationBlock{
				Time:     round(float64(currentItem.Time.T().UnixNano())/1e9, 6),
				Duration: currentItem.CalculatedDuration,
				Color:    nil,
			}
			populateDurationBlockFields(&newBlock, currentItem, sliceBy, true)

			if newBlock.Duration < 0 {
				newBlock.Duration = 0
			}
			newBlock.DurationSecs = newBlock.Duration.Seconds()
			finalDurations = append(finalDurations, newBlock)
		}
		lastProcessed = currentItem
	}

	return finalDurations
}

func handleBoundaryHeartbeats(args ProcessHeartbeatsArgs) []*models.Heartbeat {
	timeoutDuration := time.Duration(args.TimeoutMinutes) * time.Minute
	tempHeartbeats := make([]*models.Heartbeat, 0, len(args.Heartbeats)+2)

	if args.LastHeartbeatYesterday != nil && len(args.Heartbeats) > 0 {
		diff := args.Heartbeats[0].Time.T().Sub(args.LastHeartbeatYesterday.Time.T())
		yesterdaySliceValue := getValueForSlice(args.LastHeartbeatYesterday, args.SliceBy)
		firstDaySliceValue := getValueForSlice(args.Heartbeats[0], args.SliceBy)

		if diff > 0 && diff < timeoutDuration && strings.EqualFold(yesterdaySliceValue, firstDaySliceValue) {
			yesterdayCopy := *args.LastHeartbeatYesterday
			yesterdayCopy.Time = models.CustomTime(args.Start)
			tempHeartbeats = append(tempHeartbeats, &yesterdayCopy)
		}
	} else if args.LastHeartbeatYesterday != nil && len(args.Heartbeats) == 0 && args.LastHeartbeatYesterday.Time.T().After(args.Start) && args.LastHeartbeatYesterday.Time.T().Before(args.End) {
		yesterdayCopy := *args.LastHeartbeatYesterday
		tempHeartbeats = append(tempHeartbeats, &yesterdayCopy)
	}

	tempHeartbeats = append(tempHeartbeats, args.Heartbeats...)

	if args.FirstHeartbeatTomorrow != nil && len(tempHeartbeats) > 0 {
		lastHeartbeat := tempHeartbeats[len(tempHeartbeats)-1]
		diff := args.FirstHeartbeatTomorrow.Time.T().Sub(lastHeartbeat.Time.T())
		tomorrowSliceValue := getValueForSlice(args.FirstHeartbeatTomorrow, args.SliceBy)
		lastDaySliceValue := getValueForSlice(lastHeartbeat, args.SliceBy)

		if diff > 0 && diff < timeoutDuration && strings.EqualFold(tomorrowSliceValue, lastDaySliceValue) {
			tomorrowCopy := *args.FirstHeartbeatTomorrow
			tomorrowCopy.Time = models.CustomTime(args.End)
			tempHeartbeats = append(tempHeartbeats, &tomorrowCopy)
		}
	} else if args.FirstHeartbeatTomorrow != nil && len(tempHeartbeats) == 0 && args.FirstHeartbeatTomorrow.Time.T().After(args.Start) && args.FirstHeartbeatTomorrow.Time.T().Before(args.End) {
		tomorrowCopy := *args.FirstHeartbeatTomorrow
		tempHeartbeats = append(tempHeartbeats, &tomorrowCopy)
	}

	return tempHeartbeats
}

func ProcessHeartbeats(args ProcessHeartbeatsArgs) (DurationResult, error) {
	tempHeartbeats := handleBoundaryHeartbeats(args)

	// 2. Sort Heartbeats
	sort.SliceStable(tempHeartbeats, func(i, j int) bool {
		return tempHeartbeats[i].Time.T().Before(tempHeartbeats[j].Time.T())
	})

	heartbeats := tempHeartbeats

	miniDurations := heartbeatsToMiniDurations(heartbeats, args.TimeoutMinutes)

	finalDurations := combineMiniDurations(miniDurations, args.TimeoutMinutes, args.SliceBy)

	for _, duration := range finalDurations {
		duration.AddDurationSecs()
	}

	result := DurationResult{
		Data:     finalDurations,
		Start:    args.Start,
		End:      args.End,
		Timezone: args.Timezone.String(),
	}

	total := result.TotalTime()
	result.GrandTotal = NewSummariesGrandTotal(total)

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

	sliceBy := params.Get("slice_by")
	if sliceBy == "" {
		sliceBy = SliceByProject
	} else {
		if _, ok := AllowedSliceBy[strings.ToLower(sliceBy)]; !ok {
			helpers.RespondJSON(w, r, http.StatusBadRequest, map[string]interface{}{
				"message": fmt.Sprintf("Invalid slice_by value '%s'. Allowed values are: %v", sliceBy, utilities.MapKeys(AllowedSliceBy)),
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

	args := ProcessHeartbeatsArgs{
		Heartbeats:             heartbeats,
		Start:                  rangeFrom,
		End:                    rangeTo,
		TimeoutMinutes:         15,
		Timezone:               timezone,
		LastHeartbeatYesterday: yesterdayHB,
		FirstHeartbeatTomorrow: tomorrowHB,
		SliceBy:                sliceBy,
	}

	durations, err := ProcessHeartbeats(args)

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

	durations, mapping, err := MakeDurationsFromHeartbeats(heartbeats, user, &models.Filters{})
	minidurations := heartbeatsToMiniDurations(heartbeats, 15)
	finalDurations := combineMiniDurations(minidurations, 15, "entity")
	reconciled, _ := services.MakeDurationsFromHeartbeatsReconciled(heartbeats, user, &models.Filters{})

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

	var totalReconciledDuration time.Duration
	for _, duration := range reconciled {
		totalReconciledDuration += duration.Duration
	}

	result := DurationResult{
		Data:     finalDurations,
		Start:    rangeFrom,
		End:      rangeTo,
		Timezone: timezone.String(),
	}

	total := result.TotalTime()

	helpers.RespondJSON(w, r, http.StatusOK, map[string]any{
		"durations":          durations,
		"mapping":            mapping,
		"finalDurations":     finalDurations,
		"finalDurationsSize": len(finalDurations),
		"durationsSize":      len(durations),
		"myTotal":            total,
		"totalDuration":      totalDuration,
	})
}

func MakeDurationsFromHeartbeats(heartbeats []*models.Heartbeat, user *models.User, filters *models.Filters) (models.Durations, map[string][]*models.Duration, error) {
	heartbeatsTimeout := user.HeartbeatsTimeout()
	// Aggregation
	// the below logic is approximately equivalent to the SQL query at scripts/aggregate_durations_mysql.sql
	// a postgres-compatible script was contributed by @cwilby and is available at scripts/aggregate_durations_postgres.sql
	// i'm hesitant to replicate that logic for sqlite and mssql too (because probably painful to impossible), but we could
	// think about adding a distrinctio here to use pure-sql aggregation for mysql and postgres, and traditional, programmatic
	// aggregation for all other databases
	var count int
	var latest *models.Duration

	mapping := make(map[string][]*models.Duration)

	for _, h := range heartbeats {
		d1 := models.NewDurationFromHeartbeat(h).WithEntityIgnored().Hashed()

		if list, ok := mapping[d1.GroupHash]; !ok || len(list) < 1 {
			mapping[d1.GroupHash] = []*models.Duration{d1}
		}

		if latest == nil {
			latest = d1
			continue
		}

		sameDay := datetime.BeginOfDay(d1.Time.T()) == datetime.BeginOfDay(latest.Time.T())
		dur := time.Duration(mathutil.Min(
			int64(d1.Time.T().Sub(latest.Time.T().Add(latest.Duration))),
			int64(heartbeatsTimeout),
		))

		// skip heartbeats that span across two adjacent summaries (assuming there are no more than 1 summary per day)
		// this is relevant to prevent the time difference between generating summaries from raw heartbeats and aggregating pre-generated summaries
		// for the latter case, the very last heartbeat of a day won't be counted, so we don't want to count it here either
		// another option would be to adapt the Summarize() method to always append up to DefaultHeartbeatsTimeout seconds to a day's very last duration
		if !sameDay {
			dur = 0
		}
		latest.Duration += dur

		// start new "group" if:
		// (a) heartbeats were too far apart each other,
		// (b) if they are of a different entity or,
		// (c) if they span across two days
		if dur >= heartbeatsTimeout || latest.GroupHash != d1.GroupHash || !sameDay {
			list := mapping[d1.GroupHash]
			if d0 := list[len(list)-1]; d0 != d1 {
				mapping[d1.GroupHash] = append(mapping[d1.GroupHash], d1)
			}
			latest = d1
		} else {
			latest.NumHeartbeats++
		}

		count++
	}

	values := make([][]*models.Duration, 0)
	for _, list := range mapping {
		values = append(values, list)
	}

	fmt.Println("VALUES", len(values), values[0])

	// s, _ := json.Marshal(values)
	// fmt.Println("VALUES", string(s))

	// s, _ = json.Marshal(mapping)
	// fmt.Println("MAPPING", string(s))

	durations := make(models.Durations, 0)

	for _, list := range mapping {
		for _, d := range list {
			// even when filters are applied, we'll still have to compute the whole summary first and then filter out non-matching durations
			// if we fetched only matching heartbeats in the first place, there will be false positive gaps (see DefaultHeartbeatsTimeout)
			// in case the user worked on different projects in parallel
			// see https://github.com/muety/wakapi/issues/535
			if filters != nil && !filters.MatchDuration(d) {
				continue
			}

			if user.ExcludeUnknownProjects && d.Project == "" {
				continue
			}

			// will only happen if two heartbeats with different hashes (e.g. different project) have the same timestamp
			// that, in turn, will most likely only happen for mysql, where `time` column's precision was set to second for a while
			// assume that two non-identical heartbeats with identical time are sub-second apart from each other, so round up to expectancy value
			// also see https://github.com/muety/wakapi/issues/340
			if d.Duration == 0 {
				d.Duration = 500 * time.Millisecond
			}
			durations = append(durations, d)
		}
	}

	if len(heartbeats) == 1 && len(durations) == 1 {
		durations[0].Duration = heartbeatsTimeout
	}

	return durations.Sorted(), mapping, nil
}
