package api

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/duke-git/lancet/v2/datetime"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
)

/**
 * This is a glorified snapshot test. Nothing more.
**/

//go:embed testData/heartbeats.json
var testDataJSON []byte

//go:embed testData/durations_snapshot.json
var durationSnapshot []byte

//go:embed testData/yesterdays_last_heartbeat.json
var lastHeartbeatFromYesterday []byte

//go:embed testData/tomorrows_first_heartbeat.json
var tomorrowsFirstHeartbeat []byte

// HeartbeatsData represents the structure of our JSON test data
type HeartbeatsData struct {
	Heartbeats []models.Heartbeat `json:"heartbeats"`
}

func loadJSON[T any](t *testing.T, data []byte) T {
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
	return result
}

func TestDurationsWithOverlappingFirstLastHeartbeat(t *testing.T) {
	data := loadJSON[[]*models.Heartbeat](t, testDataJSON)
	snapshot := loadJSON[*DurationResult](t, durationSnapshot)

	if len(data) == 0 {
		t.Errorf("Expected heartbeats length to be greater than 0, got 0")
	}

	start, _ := time.Parse(time.RFC3339Nano, "2025-04-14T00:04:05Z")
	timezone := time.UTC

	rangeFrom, rangeTo := datetime.BeginOfDay(start.In(timezone)), datetime.EndOfDay(start.In(timezone))

	yesterdaysLastHeartbeat := loadJSON[*models.Heartbeat](t, lastHeartbeatFromYesterday)
	firsttHeartbeatFromTomorrow := loadJSON[*models.Heartbeat](t, tomorrowsFirstHeartbeat)

	args := ProcessHeartbeatsArgs{
		Heartbeats:             data,
		Start:                  rangeFrom,
		End:                    rangeTo,
		TimeoutMinutes:         15,
		Timezone:               timezone,
		LastHeartbeatYesterday: yesterdaysLastHeartbeat,
		FirstHeartbeatTomorrow: firsttHeartbeatFromTomorrow,
		SliceBy:                "project",
	}

	result, err := ProcessHeartbeats(args)
	var totalSeconds = snapshot.GrandTotal.TotalSeconds

	assert.Equal(t, err, nil, "Expected process heartbeat to not return an error")
	assert.Equal(t, result.GrandTotal.TotalSeconds, totalSeconds, fmt.Sprintf("Total seconds computation must equal %f", totalSeconds))
	assert.Equal(t, result.Start, rangeFrom, "start date check")
	assert.Equal(t, result.End, rangeTo, "end date check")
}

func TestHeartbeatsLength(t *testing.T) {
	data := loadJSON[[]*models.Heartbeat](t, testDataJSON)

	if len(data) == 0 {
		t.Errorf("Expected heartbeats length to be greater than 0, got 0")
	}

	start, _ := time.Parse(time.RFC3339Nano, "2025-04-14T00:04:05Z")
	timezone := time.UTC

	rangeFrom, rangeTo := datetime.BeginOfDay(start.In(timezone)), datetime.EndOfDay(start.In(timezone))

	args := ProcessHeartbeatsArgs{
		Heartbeats:             data,
		Start:                  rangeFrom,
		End:                    rangeTo,
		TimeoutMinutes:         15,
		Timezone:               timezone,
		LastHeartbeatYesterday: nil,
		FirstHeartbeatTomorrow: nil,
		SliceBy:                "project",
	}

	result, err := ProcessHeartbeats(args)

	var totalSeconds float64 = 38079

	assert.Equal(t, err, nil, "Expected process heartbeat to not return an error")
	assert.Equal(t, result.GrandTotal.TotalSeconds, totalSeconds, fmt.Sprintf("Total seconds computation must equal %f", totalSeconds))
	assert.Equal(t, result.Start, rangeFrom, "start date check")
	assert.Equal(t, result.End, rangeTo, "end date check")
}
