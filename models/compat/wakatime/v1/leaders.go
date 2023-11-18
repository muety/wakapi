package v1

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/muety/wakapi/models"
	"time"
)

// partially compatible with https://wakatime.com/developers#leaders

type LeaderboardViewModel struct {
	CurrentUser CurrentUser        `json:"current_user"`
	Data        []LeaderboardEntry `json:"data"`
}

type CurrentUser struct {
	Rank int   `json:"rank"`
	Page int   `json:"page"`
	User *User `json:"user"`
}

type LeaderboardEntry struct {
	Rank         int          `json:"rank"`
	RunningTotal RunningTotal `json:"running_total"`
	User         *User        `json:"user"`
}

type RunningTotal struct {
	TotalSeconds              float64    `json:"total_seconds"`
	HumanReadableTotal        string     `json:"human_readable_total"`
	DailyAverage              float64    `json:"daily_average"`
	HumanReadableDailyAverage string     `json:"human_readable_daily_average"`
	Languages                 []Language `json:"languages"`
}

type Language struct {
	Name         string  `json:"name"`
	TotalSeconds float64 `json:"total_seconds"`
}

func NewRunningTotal(entry *models.LeaderboardItemRanked) RunningTotal {
	return RunningTotal{
		TotalSeconds:              entry.Total.Seconds(),
		HumanReadableTotal:        entry.Total.String(),
		DailyAverage:              entry.Total.Seconds() / 7,
		HumanReadableDailyAverage: (entry.Total / 7).String(),
		Languages:                 []Language{},
	}
}

func NewLeaderboardEntry(entry *models.LeaderboardItemRanked) LeaderboardEntry {
	return LeaderboardEntry{
		Rank:         int(entry.Rank),
		RunningTotal: NewRunningTotal(entry),
		User:         NewFromUser(entry.User),
	}
}

func NewLeaderboardViewModel(leaderboard *models.Leaderboard, user *models.User) *LeaderboardViewModel {
	var entries []LeaderboardEntry
	var currentUserEntry = CurrentUser{}
	for _, entry := range *leaderboard {
		if user != nil && entry.User.ID == user.ID {
			currentUserEntry = CurrentUser{
				Rank: int(entry.Rank),
				Page: 0,
				User: NewFromUser(user),
			}
		}

		if foundEntry, found := slice.Find[LeaderboardEntry](entries, func(i int, entry2 LeaderboardEntry) bool {
			return entry2.User.ID == entry.User.ID
		}); found {
			foundEntry.RunningTotal.Languages = append(foundEntry.RunningTotal.Languages, Language{
				Name:         *entry.Key,
				TotalSeconds: entry.Total.Seconds(),
			})
			foundEntry.RunningTotal.TotalSeconds += entry.Total.Seconds()
			foundEntry.RunningTotal.DailyAverage = foundEntry.RunningTotal.TotalSeconds / 7
			var totalDuration time.Duration = time.Duration(foundEntry.RunningTotal.TotalSeconds) * time.Second
			foundEntry.RunningTotal.HumanReadableTotal = totalDuration.String()
			foundEntry.RunningTotal.HumanReadableDailyAverage = (totalDuration / 7).String()

			entries = slice.Filter[LeaderboardEntry](entries, func(i int, entry2 LeaderboardEntry) bool {
				return entry2.User.ID != entry.User.ID
			})
			entries = append(entries, *foundEntry)
			continue
		}
		entries = append(entries, NewLeaderboardEntry(entry))
	}

	return &LeaderboardViewModel{
		CurrentUser: currentUserEntry,
		Data:        entries,
	}
}
