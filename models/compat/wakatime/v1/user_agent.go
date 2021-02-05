package v1

type UserAgentsViewModel struct {
	Data []*UserAgentEntry `json:"data"`
}

type UserAgentEntry struct {
	Id     string
	Editor string
	Os     string
	Value  string
}
