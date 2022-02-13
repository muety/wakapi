package v1

import (
	"strconv"
	"time"

	"github.com/muety/wakapi/models"
)

type HeartbeatsViewModel struct {
	Data []*HeartbeatEntry `json:"data"`
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
