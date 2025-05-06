package services

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/muety/wakapi/models"
	wakatime "github.com/muety/wakapi/models/compat/wakatime/v1"
)

func NewSummariesGrandTotal(originalDuration time.Duration) *wakatime.SummariesGrandTotal {
	d := originalDuration.Round(time.Minute)

	totalSeconds := originalDuration.Round(time.Second).Seconds()
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
	CalculatedDuration time.Duration // Duration calculated in HeartbeatsToMiniDurations step (seconds)
}

func (d *DurationResult) ComputeTotalTimeFromDurations() time.Duration {
	var total time.Duration
	for _, item := range d.Data {
		total += item.Duration
	}
	return total
}

func (d *DurationResult) TotalTime() time.Duration {
	return d.ComputeTotalTimeFromDurations()
}

type DurationResult struct {
	Data       []models.DurationBlock        `json:"data"`
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

func round(number float64, precision int) float64 {
	factor := math.Pow(10, float64(precision))
	return math.Round(number*factor) / factor
}

func HeartbeatsToMiniDurations(heartbeats []*models.Heartbeat, timeoutDuration time.Duration) []MiniDurationHeartbeat {
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

func populateDurationBlockFields(block *models.DurationBlock, item MiniDurationHeartbeat, sliceBy string, respectSliceBy bool) {
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
	setField(&block.Branch, item.Branch, "branch")
}

func shouldJoinDuration(current MiniDurationHeartbeat, last MiniDurationHeartbeat, timeoutDuration time.Duration, sliceBy string) bool {

	currentSliceValue := getValueForSlice(&current.Heartbeat, sliceBy)
	lastSliceValue := getValueForSlice(&last.Heartbeat, sliceBy)

	if !strings.EqualFold(lastSliceValue, currentSliceValue) {
		return false
	}

	lastDuration := last.CalculatedDuration
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

	lastDuration := last.CalculatedDuration
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

func CombineMiniDurations(miniDurations []MiniDurationHeartbeat, timeoutMinutes time.Duration, sliceBy string) []models.DurationBlock {
	if len(miniDurations) == 0 {
		return []models.DurationBlock{}
	}

	finalDurations := make([]models.DurationBlock, 0)
	var lastProcessed MiniDurationHeartbeat

	firstHB := miniDurations[0]
	firstBlock := models.DurationBlock{
		Time:     round(float64(firstHB.Time.T().UnixNano())/1e9, 6),
		Duration: firstHB.CalculatedDuration,
		Color:    nil,
	}

	populateDurationBlockFields(&firstBlock, firstHB, sliceBy, sliceBy != SliceByEntity)

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

			populateDurationBlockFields(currentBlock, currentItem, sliceBy, sliceBy != SliceByEntity)

		} else {
			newBlock := models.DurationBlock{
				Time:     round(float64(currentItem.Time.T().UnixNano())/1e9, 6),
				Duration: currentItem.CalculatedDuration,
				Color:    nil,
			}
			populateDurationBlockFields(&newBlock, currentItem, sliceBy, sliceBy != SliceByEntity)

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

func MakeHeartbeatDurationBlocks(args models.ProcessHeartbeatsArgs) []models.DurationBlock {
	tempHeartbeats := handleBoundaryHeartbeats(args)

	// 2. Sort Heartbeats
	sort.SliceStable(tempHeartbeats, func(i, j int) bool {
		return tempHeartbeats[i].Time.T().Before(tempHeartbeats[j].Time.T())
	})

	heartbeats := tempHeartbeats

	miniDurations := HeartbeatsToMiniDurations(heartbeats, args.User.HeartbeatsTimeout())

	finalDurations := CombineMiniDurations(miniDurations, args.User.HeartbeatsTimeout(), args.SliceBy)

	for _, duration := range finalDurations {
		duration.AddDurationSecs()
	}
	return finalDurations
}

func ProcessHeartbeats(args models.ProcessHeartbeatsArgs) (DurationResult, error) {
	durationBlocks := MakeHeartbeatDurationBlocks(args)

	result := DurationResult{
		Data:     durationBlocks,
		Start:    args.Start,
		End:      args.End,
		Timezone: args.User.TZ().String(),
	}

	total := result.TotalTime()
	result.GrandTotal = NewSummariesGrandTotal(total)

	return result, nil
}
