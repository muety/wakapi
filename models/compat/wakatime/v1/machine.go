package v1

// https://wakatime.com/api/v1/users/current/machine_names

type MachineViewModel struct {
	Data       []*MachineEntry `json:"data"`
	TotalPages int             `json:"total_pages"`
}

type MachineEntry struct {
	Id    string `json:"id"`
	Value string `json:"value"`
}
