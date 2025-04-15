package api

import (
	_ "embed"
	"encoding/json"
	"testing"
	"time"

	"github.com/duke-git/lancet/v2/datetime"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
)

//go:embed testData/heartbeats.json
var testDataJSON []byte

// HeartbeatsData represents the structure of our JSON test data
type HeartbeatsData struct {
	Heartbeats []models.Heartbeat `json:"heartbeats"`
}

// loadTestData parses the embedded JSON test data
func loadTestData(t *testing.T) []*models.Heartbeat {
	var heartbeatsData []*models.Heartbeat
	if err := json.Unmarshal(testDataJSON, &heartbeatsData); err != nil {
		t.Fatalf("Failed to parse test data: %v", err)
	}
	return heartbeatsData
}

func TestHeartbeatsLength(t *testing.T) {
	data := loadTestData(t)

	if len(data) == 0 {
		t.Errorf("Expected heartbeats length to be greater than 0, got 0")
	}

	start, _ := time.Parse(time.RFC3339Nano, "2025-04-13T00:04:05Z")
	timezone := start.Location()

	rangeFrom, rangeTo := datetime.BeginOfDay(start.In(timezone)), datetime.EndOfDay(start.In(timezone))

	result, err := ProcessHeartbeats(data, rangeFrom, rangeTo, 15, timezone, nil, nil)

	assert.Equal(t, err, nil, "Expected process heartbeat to not return an error")
	assert.Equal(t, result.GrandTotal.TotalSeconds, float64(7674), "Total seconds computation must equal 7674")
	assert.Equal(t, result.Start, rangeFrom, "start date check")
	assert.Equal(t, result.End, rangeTo, "end date check")
}
