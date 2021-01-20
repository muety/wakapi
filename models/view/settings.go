package view

import "github.com/muety/wakapi/models"

type SettingsViewModel struct {
	User             *models.User
	LanguageMappings []*models.LanguageMapping
	Aliases          []*SettingsVMCombinedAlias
	Success          string
	Error            string
}

type SettingsVMCombinedAlias struct {
	Key    string
	Type   uint8
	Values []string
}

func (s *SettingsViewModel) WithSuccess(m string) *SettingsViewModel {
	s.Success = m
	return s
}

func (s *SettingsViewModel) WithError(m string) *SettingsViewModel {
	s.Error = m
	return s
}
