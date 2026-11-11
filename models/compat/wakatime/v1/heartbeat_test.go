package v1

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeartbeatEntry_AICachedInputTokens_JSON(t *testing.T) {
	raw := `{
		"id": "12345",
		"branch": "main",
		"category": "ai coding",
		"entity": "main.go",
		"is_write": true,
		"language": "Go",
		"project": "wakapi",
		"time": 1728422364,
		"type": "file",
		"user_id": "test_user",
		"ai_input_tokens": 200,
		"ai_output_tokens": 100,
		"ai_cached_input_tokens": 4500,
		"ai_prompt_length": 50,
		"human_line_changes": 2
	}`

	var entry HeartbeatEntry
	err := json.Unmarshal([]byte(raw), &entry)
	require.NoError(t, err)
	assert.Equal(t, "12345", entry.Id)
	assert.Equal(t, 200, entry.AIInputTokens)
	assert.Equal(t, 100, entry.AIOutputTokens)
	assert.Equal(t, 4500, entry.AICachedInputTokens)

	// JSON serialization must include ai_cached_input_tokens even when zero (compat spec)
	entry.AICachedInputTokens = 0
	marshaled, err := json.Marshal(&entry)
	require.NoError(t, err)
	assert.Contains(t, string(marshaled), `"ai_cached_input_tokens":0`)

	entry.AICachedInputTokens = 8800
	marshaledWithVal, err := json.Marshal(&entry)
	require.NoError(t, err)
	assert.Contains(t, string(marshaledWithVal), `"ai_cached_input_tokens":8800`)
}

func TestHeartbeatsToCompat_AICachedInputTokens(t *testing.T) {
	now := time.Now()
	entries := []*models.Heartbeat{
		{
			ID:                  42,
			UserID:              "user1",
			Entity:              "server.go",
			Project:             "wakapi",
			Time:                models.CustomTime(now),
			CreatedAt:           models.CustomTime(now),
			AIInputTokens:       500,
			AIOutputTokens:      250,
			AICachedInputTokens: 9500,
		},
		{
			ID:                  43,
			UserID:              "user1",
			Entity:              "client.go",
			Project:             "wakapi",
			Time:                models.CustomTime(now.Add(time.Second)),
			CreatedAt:           models.CustomTime(now),
			AIInputTokens:       100,
			AIOutputTokens:      50,
			AICachedInputTokens: 0,
		},
	}

	compatEntries := HeartbeatsToCompat(entries)
	require.Len(t, compatEntries, 2)
	assert.Equal(t, "42", compatEntries[0].Id)
	assert.Equal(t, 500, compatEntries[0].AIInputTokens)
	assert.Equal(t, 250, compatEntries[0].AIOutputTokens)
	assert.Equal(t, 9500, compatEntries[0].AICachedInputTokens)

	assert.Equal(t, "43", compatEntries[1].Id)
	assert.Equal(t, 100, compatEntries[1].AIInputTokens)
	assert.Equal(t, 50, compatEntries[1].AIOutputTokens)
	assert.Equal(t, 0, compatEntries[1].AICachedInputTokens)
}
