package view

import "github.com/muety/wakapi/models"

type SummaryViewModel struct {
	Messages
	*models.Summary
	*models.SummaryParams
	User           *models.User
	AvatarURL      string
	EditorColors   map[string]string
	LanguageColors map[string]string
	OSColors       map[string]string
	ApiKey         string
	RawQuery       string
}

func (s *SummaryViewModel) WithSuccess(m string) *SummaryViewModel {
	s.SetSuccess(m)
	return s
}

func (s *SummaryViewModel) WithError(m string) *SummaryViewModel {
	s.SetError(m)
	return s
}
