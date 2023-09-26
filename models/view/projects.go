package view

import (
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
)

type ProjectsViewModel struct {
	Messages
	User       *models.User
	Projects   []*models.ProjectStats
	ApiKey     string
	PageParams *utils.PageParams
}

func (s *ProjectsViewModel) LangIcon(lang string) string {
	return GetLanguageIcon(lang)
}

func (s *ProjectsViewModel) WithSuccess(m string) *ProjectsViewModel {
	s.SetSuccess(m)
	return s
}

func (s *ProjectsViewModel) WithError(m string) *ProjectsViewModel {
	s.SetError(m)
	return s
}
