package view

import (
	"github.com/muety/wakapi/models"
	"strings"
	"time"
)

type LeaderboardViewModel struct {
	User          *models.User
	By            string
	Key           string
	Items         []*models.LeaderboardItem
	TopKeys       []string
	UserLanguages map[string][]string
	ApiKey        string
	Success       string
	Error         string
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

func (s *LeaderboardViewModel) LangIcon(lang string) string {
	// https://icon-sets.iconify.design/mdi/
	langs := map[string]string{
		"c++":        "cpp",
		"cpp":        "cpp",
		"go":         "go",
		"haskell":    "haskell",
		"html":       "html5",
		"java":       "java",
		"javascript": "javascript",
		"kotlin":     "kotlin",
		"lua":        "lua",
		"php":        "php",
		"python":     "python",
		"r":          "r",
		"ruby":       "ruby",
		"rust":       "rust",
		"swift":      "swift",
		"typescript": "typescript",
	}
	if match, ok := langs[strings.ToLower(lang)]; ok {
		return "mdi:language-" + match
	}
	return ""
}

func (s *LeaderboardViewModel) LastUpdate() time.Time {
	return models.Leaderboard(s.Items).LastUpdate()
}
