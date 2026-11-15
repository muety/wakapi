package imports

import (
	"testing"
	"time"

	"github.com/muety/wakapi/models"
	wakatime "github.com/muety/wakapi/models/compat/wakatime/v1"
	"github.com/stretchr/testify/assert"
)

func TestMapHeartbeat_AiFields(t *testing.T) {
	userAgents := map[string]*wakatime.UserAgentEntry{
		"ua1": {Id: "ua1", Editor: "Claude Code", Os: "Linux", Value: "claude-code/1.0 (Linux)"},
	}

	entry := &wakatime.HeartbeatEntry{
		Id:                  "f3647f89-e255-4dd1-8fcd-e20ba8f1709b",
		Entity:              "claude-code",
		Type:                "app",
		Category:            models.HeartbeatCategoryAiCoding,
		Time:                1728422364.044,
		CreatedAt:           time.Unix(1728422364, 0),
		UserAgentId:         "ua1",
		AILineChanges:       12,
		AISession:           "657bbd5c",
		AIInputTokens:       728,
		AIOutputTokens:      436,
		AICachedInputTokens: 51000,
		AIPromptLength:      40,
		HumanLineChanges:    3,
	}

	hb := mapHeartbeat(entry, userAgents, nil, &models.User{ID: "user1"})

	assert.Equal(t, 12, hb.AILineChanges)
	assert.Equal(t, "657bbd5c", hb.AISession)
	assert.Equal(t, 728, hb.AIInputTokens)
	assert.Equal(t, 436, hb.AIOutputTokens)
	assert.Equal(t, 51000, hb.AICachedInputTokens)
	assert.Equal(t, 40, hb.AIPromptLength)
	assert.Equal(t, 3, hb.HumanLineChanges)
}

func TestMapHeartbeat_AIModel(t *testing.T) {
	userAgents := map[string]*wakatime.UserAgentEntry{
		"ua1": {Id: "ua1", AIModel: "gpt-4o"},
	}

	// heartbeat without own ai_model falls back to the one from the user agent
	hb := mapHeartbeat(&wakatime.HeartbeatEntry{UserAgentId: "ua1"}, userAgents, nil, &models.User{ID: "user1"})
	assert.Equal(t, "gpt-4o", hb.AIModel)

	// heartbeat-level ai_model (as returned by wakapi's compat endpoint) takes precedence
	hb = mapHeartbeat(&wakatime.HeartbeatEntry{UserAgentId: "ua1", AIModel: "claude-sonnet-4"}, userAgents, nil, &models.User{ID: "user1"})
	assert.Equal(t, "claude-sonnet-4", hb.AIModel)
}

func TestMapHeartbeat_HashUnaffectedByAiFields(t *testing.T) {
	entry1 := &wakatime.HeartbeatEntry{
		Id:          "a1",
		Entity:      "main.go",
		Type:        "file",
		Time:        1728422364.044,
		CreatedAt:   time.Unix(1728422364, 0),
		UserAgentId: "ua1",
	}
	entry2 := *entry1
	entry2.AIPromptLength = 40
	entry2.AICachedInputTokens = 51000
	entry2.AILineChanges = 12

	hb1 := mapHeartbeat(entry1, nil, nil, &models.User{ID: "user1"})
	hb2 := mapHeartbeat(&entry2, nil, nil, &models.User{ID: "user1"})

	assert.Equal(t, hb1.Hash, hb2.Hash)
}
