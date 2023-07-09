package v1

type JsonExportViewModel struct {
	//User  *User            `json:"user"`
	Range *JsonExportRange `json:"range"`
	Days  []*JsonExportDay `json:"days"`
}

type JsonExportRange struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

type JsonExportDay struct {
	Date       string            `json:"date"`
	Heartbeats []*HeartbeatEntry `json:"heartbeats"`
}
