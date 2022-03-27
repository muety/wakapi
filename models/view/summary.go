package view

import "github.com/muety/wakapi/models"

type SummaryViewModel struct {
	*models.Summary
	*models.SummaryParams
	User           *models.User
	AvatarURL      string
	VibrantColor   bool
	EditorColors   map[string]string
	LanguageColors map[string]string
	OSColors       map[string]string
	Error          string
	Success        string
	ApiKey         string
	RawQuery       string
}

func (s *SummaryViewModel) WithSuccess(m string) *SummaryViewModel {
	s.Success = m
	return s
}

func (s *SummaryViewModel) WithError(m string) *SummaryViewModel {
	s.Error = m
	return s
}
