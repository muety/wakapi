package v1

import (
	"github.com/muety/wakapi/models"
)

type HeartbeatsViewModel struct {
	Data []*HeartbeatEntry `json:"data"`
}

// Incomplete, for now, only the subset of fields is implemented
// that is actually required for the import

type HeartbeatEntry struct {
	Id            string            `json:"id"`
	Branch        string            `json:"branch"`
	Category      string            `json:"category"`
	Entity        string            `json:"entity"`
	IsWrite       bool              `json:"is_write"`
	Language      string            `json:"language"`
	Project       string            `json:"project"`
	Time          models.CustomTime `json:"time"`
	Type          string            `json:"type"`
	UserId        string            `json:"user_id"`
	MachineNameId string            `json:"machine_name_id"`
	UserAgentId   string            `json:"user_agent_id"`
	CreatedAt     models.CustomTime `json:"created_at"`
	ModifiedAt    models.CustomTime `json:"created_at"`
}
