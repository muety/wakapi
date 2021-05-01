package v1

type ProjectsViewModel struct {
	Data []*Project `json:"data"`
}

type Project struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Repository string `json:"repository"`
}
