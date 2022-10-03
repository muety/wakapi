package view

import "github.com/muety/wakapi/models"

type LeaderboardViewModel struct {
	User    *models.User
	By      string
	Key     string
	Items   []*models.LeaderboardItem
	TopKeys []string
	ApiKey  string
	Success string
	Error   string
}

func (s *LeaderboardViewModel) WithSuccess(m string) *LeaderboardViewModel {
	s.Success = m
	return s
}

func (s *LeaderboardViewModel) WithError(m string) *LeaderboardViewModel {
	s.Error = m
	return s
}

func (s *LeaderboardViewModel) ColorModifier(item *models.LeaderboardItem, principal *models.User) string {
	if principal != nil && item.UserID == principal.ID {
		return "self"
	}
	if item.Rank == 1 {
		return "gold"
	}
	if item.Rank == 2 {
		return "silver"
	}
	if item.Rank == 3 {
		return "bronze"
	}
	return "default"
}
