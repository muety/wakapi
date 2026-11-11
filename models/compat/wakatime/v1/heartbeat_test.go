package v1

import (
	"encoding/json"
	"testing"

	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
)

func TestHeartbeatsToCompat_AITokens(t *testing.T) {
	out := HeartbeatsToCompat([]*models.Heartbeat{{
		ID:                  1,
		AIInputTokens:       10,
		AIOutputTokens:      20,
		AICachedInputTokens: 3000,
	}})
	assert.Len(t, out, 1)
	assert.Equal(t, 3000, out[0].AICachedInputTokens)

	raw, err := json.Marshal(out[0])
	assert.NoError(t, err)
	assert.Contains(t, string(raw), `"ai_cached_input_tokens":3000`)
}
