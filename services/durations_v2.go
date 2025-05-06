package services

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/muety/wakapi/models"
)

func NewSummariesGrandTotal(originalDuration time.Duration) *models.SummariesGrandTotal {
	d := originalDuration.Round(time.Minute)

	totalSeconds := originalDuration.Round(time.Second).Seconds()
	hours := int(d / time.Hour)
	minutes := int((d % time.Hour) / time.Minute)

	digital := fmt.Sprintf("%02d:%02d", hours, minutes)
	text := fmt.Sprintf("%d hrs %d mins", hours, minutes)

	return &models.SummariesGrandTotal{
		Digital:      digital,
		Hours:        hours,
		Minutes:      minutes,
		Text:         text,
		TotalSeconds: totalSeconds,
	}
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
	SliceByBranch   string = "branch"
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

func HeartbeatsToMiniDurations(heartbeats []*models.Heartbeat, timeoutDuration time.Duration) []models.MiniDurationHeartbeat {
	miniDurations := make([]models.MiniDurationHeartbeat, 0, len(heartbeats))

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

		miniDurations = append(miniDurations, models.MiniDurationHeartbeat{
			Heartbeat: *hb,
			Duration:  duration,
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

func populateDurationFields(block *models.Duration, item models.MiniDurationHeartbeat, sliceBy string, respectSliceBy bool) {
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
	setField(&block.OperatingSystem, item.OperatingSystem, "os")
	setField(&block.Editor, item.Editor, "editor")
	setField(&block.Category, item.Category, "category")
	setField(&block.Machine, item.Machine, "machine")
	setField(&block.Branch, item.Branch, "branch")
}

func shouldJoinDuration(current models.MiniDurationHeartbeat, last models.MiniDurationHeartbeat, timeoutDuration time.Duration, sliceBy string) bool {

	currentSliceValue := getValueForSlice(&current.Heartbeat, sliceBy)
	lastSliceValue := getValueForSlice(&last.Heartbeat, sliceBy)

	if !strings.EqualFold(lastSliceValue, currentSliceValue) {
		return false
	}

	lastDuration := last.Duration
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

func CombineMiniDurations(miniDurations []models.MiniDurationHeartbeat, timeoutMinutes time.Duration, sliceBy string) []*models.Duration {
	if len(miniDurations) == 0 {
		return []*models.Duration{}
	}

	finalDurations := make([]*models.Duration, 0)
	var lastProcessed models.MiniDurationHeartbeat

	firstHB := miniDurations[0]
	firstBlock := models.Duration{
		Time:         firstHB.Time,
		Duration:     firstHB.Duration,
		DurationSecs: firstHB.Duration.Seconds(),
		Color:        nil,
	}

	populateDurationFields(&firstBlock, firstHB, sliceBy, sliceBy != SliceByEntity)

	if firstBlock.Duration < 0 {
		firstBlock.Duration = 0
	}
	finalDurations = append(finalDurations, &firstBlock)
	lastProcessed = firstHB

	for i := 1; i < len(miniDurations); i++ {
		currentItem := miniDurations[i]
		currentBlock := finalDurations[len(finalDurations)-1]

		if shouldJoinDuration(currentItem, lastProcessed, timeoutMinutes, sliceBy) {
			itemDuration := currentItem.Duration
			endTime := currentItem.Time.T().Add(itemDuration)
			startTime := currentBlock.Time.T()
			newDuration := endTime.Sub(startTime)

			currentBlock.Duration = max(newDuration, time.Duration(0))
			currentBlock.DurationSecs = currentBlock.Duration.Seconds()

			populateDurationFields(currentBlock, currentItem, sliceBy, sliceBy != SliceByEntity)

		} else {
			newBlock := models.Duration{
				Time:     currentItem.Time,
				Duration: currentItem.Duration,
				Color:    nil,
			}
			populateDurationFields(&newBlock, currentItem, sliceBy, sliceBy != SliceByEntity)

			if newBlock.Duration < 0 {
				newBlock.Duration = 0
			}
			newBlock.DurationSecs = newBlock.Duration.Seconds()
			finalDurations = append(finalDurations, &newBlock)
		}
		lastProcessed = currentItem
	}

	return finalDurations
}

func handleBoundaryHeartbeats(args models.ProcessHeartbeatsArgs) []*models.Heartbeat {
	timeoutDuration := args.User.HeartbeatsTimeout()
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

func MakeHeartbeatDurations(args models.ProcessHeartbeatsArgs) []*models.Duration {
	tempHeartbeats := handleBoundaryHeartbeats(args)

	// 2. Sort Heartbeats
	sort.SliceStable(tempHeartbeats, func(i, j int) bool {
		return tempHeartbeats[i].Time.T().Before(tempHeartbeats[j].Time.T())
	})

	heartbeats := tempHeartbeats

	miniDurations := HeartbeatsToMiniDurations(heartbeats, args.User.HeartbeatsTimeout())

	finalDurations := CombineMiniDurations(miniDurations, args.User.HeartbeatsTimeout(), args.SliceBy)

	return finalDurations
}

func ProcessHeartbeats(args models.ProcessHeartbeatsArgs) (models.DurationResult, error) {
	durations := MakeHeartbeatDurations(args)

	result := models.DurationResult{
		Data:     models.Durations(durations),
		Start:    args.Start,
		End:      args.End,
		Timezone: args.User.TZ().String(),
	}

	total := result.TotalTime()
	result.GrandTotal = NewSummariesGrandTotal(total)

	return result, nil
}
