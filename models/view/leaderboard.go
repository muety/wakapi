package view

import (
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"strings"
	"time"
)

type LeaderboardViewModel struct {
	Messages
	User          *models.User
	By            string
	Key           string
	Items         []*models.LeaderboardItemRanked
	TopKeys       []string
	UserLanguages map[string][]string
	ApiKey        string
	PageParams    *utils.PageParams
}

func (s *LeaderboardViewModel) WithSuccess(m string) *LeaderboardViewModel {
	s.SetSuccess(m)
	return s
}

func (s *LeaderboardViewModel) WithError(m string) *LeaderboardViewModel {
	s.SetError(m)
	return s
}

func (s *LeaderboardViewModel) ColorModifier(item *models.LeaderboardItemRanked, principal *models.User) string {
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
		"c++":        "language-cpp",
		"cpp":        "language-cpp",
		"go":         "language-go",
		"haskell":    "language-haskell",
		"html":       "language-html5",
		"java":       "language-java",
		"javascript": "language-javascript",
		"jsx":        "language-javascript",
		"kotlin":     "language-kotlin",
		"lua":        "language-lua",
		"php":        "language-php",
		"python":     "language-python",
		"r":          "language-r",
		"ruby":       "language-ruby",
		"rust":       "language-rust",
		"swift":      "language-swift",
		"typescript": "language-typescript",
		"tsx":        "language-typescript",
		"markdown":   "language-markdown",
		"vue":        "vuejs",
		"react":      "react",
		"bash":       "bash",
		"json":       "code-json",
	}
	if match, ok := langs[strings.ToLower(lang)]; ok {
		return "mdi:" + match
	}
	return ""
}

func (s *LeaderboardViewModel) LastUpdate() time.Time {
	return models.Leaderboard(s.Items).LastUpdate()
}
