package models

// ProjectLabelReverseResolver returns all projects for a given label
type ProjectLabelReverseResolver func(l string) []string

type ProjectLabel struct {
	ID         uint   `json:"id" gorm:"primary_key"`
	User       *User  `json:"-" gorm:"not null; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	UserID     string `json:"-" gorm:"not null; index:idx_project_label_user"`
	ProjectKey string `json:"project"`
	Label      string `json:"label" gorm:"type:varchar(64)"`
}

func (l *ProjectLabel) IsValid() bool {
	return l.ProjectKey != "" && l.Label != ""
}
