package v1

import "time"

type ProjectsViewModel struct {
	Data []*Project `json:"data"`
}

type Project struct {
	ID                           string    `json:"id"`
	Name                         string    `json:"name"`
	LastHeartbeatAt              time.Time `json:"last_heartbeat_at"`
	HumanReadableLastHeartbeatAt string    `json:"human_readable_last_heartbeat_at"`
	UrlencodedName               string    `json:"urlencoded_name"`
	CreatedAt                    time.Time `json:"created_at"`
}
