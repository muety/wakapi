package view

import (
	"github.com/muety/wakapi/models"
	"time"
)

type SettingsViewModel struct {
	Messages
	User                *models.User
	LanguageMappings    []*models.LanguageMapping
	Aliases             []*SettingsVMCombinedAlias
	Labels              []*SettingsVMCombinedLabel
	Projects            []string
	SubscriptionPrice   string
	DataRetentionMonths int
	UserFirstData       time.Time
	SupportContact      string
	ApiKey              string
}

type SettingsVMCombinedAlias struct {
	Key    string
	Type   uint8
	Values []string
}

type SettingsVMCombinedLabel struct {
	Key    string
	Values []string
}

func (s *SettingsViewModel) SubscriptionsEnabled() bool {
	return s.SubscriptionPrice != ""
}

func (s *SettingsViewModel) WithSuccess(m string) *SettingsViewModel {
	s.SetSuccess(m)
	return s
}

func (s *SettingsViewModel) WithError(m string) *SettingsViewModel {
	s.SetError(m)
	return s
}
