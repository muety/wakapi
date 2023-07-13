package v1

type DataDumpViewModel struct {
	Data       []*DataDumpData `json:"data"`
	Total      int             `json:"total"`
	TotalPages int             `json:"total_pages"`
}

type DataDumpResultErrorModel struct {
	Error string `json:"error"`
}

type DataDumpResultViewModel struct {
	Data *DataDumpData `json:"data"`
}

type DataDumpData struct {
	Id              string  `json:"id"`
	Type            string  `json:"type"`
	DownloadUrl     string  `json:"download_url"`
	Status          string  `json:"status"`
	PercentComplete float32 `json:"percent_complete"`
	Expires         string  `json:"expires"`
	CreatedAt       string  `json:"created_at"`
	HasFailed       bool    `json:"has_failed"`
	IsStuck         bool    `json:"is_stuck"`
	IsProcessing    bool    `json:"is_processing"`
}
