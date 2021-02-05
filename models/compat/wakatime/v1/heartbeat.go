package v1

import "github.com/muety/wakapi/models"

type HeartbeatsViewModel struct {
	Data []*HeartbeatEntry `json:"data"`
}

// Incomplete, for now, only the subset of fields is implemented
// that is actually required for the import

type HeartbeatEntry struct {
	Id            string
	Branch        string
	Category      string
	Entity        string
	IsWrite       bool `json:"is_write"`
	Language      string
	Project       string
	Time          models.CustomTime
	Type          string
	UserId        string `json:"user_id"`
	MachineNameId string `json:"machine_name_id"`
	UserAgentId   string `json:"user_agent_id"`
}
