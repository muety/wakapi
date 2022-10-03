package view

import "github.com/muety/wakapi/models"

type LeaderboardViewModel struct {
	User            *models.User
	Items           []*models.LeaderboardItem
	ItemsByLanguage []*models.LeaderboardItem
	ApiKey          string
	Success         string
	Error           string
}

func (s *LeaderboardViewModel) WithSuccess(m string) *LeaderboardViewModel {
	s.Success = m
	return s
}

func (s *LeaderboardViewModel) WithError(m string) *LeaderboardViewModel {
	s.Error = m
	return s
}
