package v1

import (
	"strconv"
	"time"

	"github.com/muety/wakapi/models"
)

type HeartbeatsViewModel struct {
	Data []*HeartbeatEntry `json:"data"`
}

type HeartbeatResponseViewModel struct {
	Responses [][]interface{} `json:"responses"`
}

type HeartbeatResponseData struct {
	// response data actually looks like this: https://pastr.de/p/nyf6kj2e6843fbw4xkj4h4pj
	// however, for simplicity, we only implement the top-level fields (and the status code at index 1) and leave them empty
	// neither cli, now the browser plugin require these fields to hold any actual content, see:
	// - https://github.com/wakatime/wakatime-cli/blob/660b2b9702f3bc598e96cb8b8e9f459a6c2928c3/pkg/api/heartbeat.go#L152
	// - https://github.com/wakatime/wakatime-cli/blob/660b2b9702f3bc598e96cb8b8e9f459a6c2928c3/cmd/heartbeat/heartbeat.go#L151
	// - https://github.com/wakatime/browser-wakatime/blob/77d0963d93552756f52e72ed57f0d2f6f7d6239f/src/core/WakaTimeCore.ts#L208
	Data  interface{} `json:"data"`
	Error interface{} `json:"error"`
}

// Incomplete, for now, only the subset of fields is implemented
// that is actually required for the import

type HeartbeatEntry struct {
	Id            string    `json:"id"`
	Branch        string    `json:"branch"`
	Category      string    `json:"category"`
	Entity        string    `json:"entity"`
	IsWrite       bool      `json:"is_write"`
	Language      string    `json:"language"`
	Project       string    `json:"project"`
	Time          float64   `json:"time"`
	Type          string    `json:"type"`
	UserId        string    `json:"user_id"`
	MachineNameId string    `json:"machine_name_id"`
	UserAgentId   string    `json:"user_agent_id"`
	CreatedAt     time.Time `json:"created_at"`
}

func HeartbeatsToCompat(entries []*models.Heartbeat) []*HeartbeatEntry {
	out := make([]*HeartbeatEntry, len(entries))
	for i := 0; i < len(entries); i++ {
		entry := entries[i]
		out[i] = &HeartbeatEntry{
			Id:            strconv.FormatUint(entry.ID, 10),
			Branch:        entry.Branch,
			Category:      entry.Category,
			Entity:        entry.Entity,
			IsWrite:       entry.IsWrite,
			Language:      entry.Language,
			Project:       entry.Project,
			Time:          float64(entry.Time.T().Unix()),
			Type:          entry.Type,
			UserId:        entry.UserID,
			MachineNameId: entry.Machine,
			UserAgentId:   entry.UserAgent,
			CreatedAt:     entry.CreatedAt.T(),
		}
	}
	return out
}
