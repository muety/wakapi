package v1

type UserAgentsViewModel struct {
	Data       []*UserAgentEntry    `json:"data"`
	TotalPages int                  `json:"total_pages"`
}

type UserAgentEntry struct {
	Id     string `json:"id"`
	Editor string `json:"editor"`
	Os     string `json:"os"`
	Value  string `json:"value"`
}
