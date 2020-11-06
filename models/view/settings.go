package view

import "github.com/muety/wakapi/models"

type SettingsViewModel struct {
	User             *models.User
	LanguageMappings []*models.LanguageMapping
	Success          string
	Error            string
}

func (s *SettingsViewModel) WithSuccess(m string) *SettingsViewModel {
	s.Success = m
	return s
}

func (s *SettingsViewModel) WithError(m string) *SettingsViewModel {
	s.Error = m
	return s
}
