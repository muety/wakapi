package v1

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/models"
	"math"
	"time"
)

// partially compatible with https://wakatime.com/developers#leaders

type LeaderboardViewModel struct {
	CurrentUser *CurrentUser        `json:"current_user"`
	Data        []*LeaderboardEntry `json:"data"`
}

type CurrentUser struct {
	Rank int   `json:"rank"`
	Page int   `json:"page"`
	User *User `json:"user"`
}

type LeaderboardEntry struct {
	Id           int           `json:"-"`
	Rank         int           `json:"rank"`
	RunningTotal *RunningTotal `json:"running_total"`
	User         *User         `json:"user"`
}

func (l1 *LeaderboardEntry) Equals(l2 *LeaderboardEntry) bool {
	return l1.Id == l2.Id
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

func CalculateNewRunningTotal(entry *RunningTotal, newTotalDuration time.Duration) *RunningTotal {
	var totalSeconds = math.RoundToEven(newTotalDuration.Seconds())
	var dailyAvg = math.RoundToEven(totalSeconds / 7)
	entry.TotalSeconds = totalSeconds
	entry.HumanReadableTotal = helpers.FmtWakatimeDuration(newTotalDuration)
	entry.DailyAverage = dailyAvg
	entry.HumanReadableDailyAverage = helpers.FmtWakatimeDuration(time.Duration(dailyAvg) * time.Second)
	return entry
}

func NewRunningTotal(entry *models.LeaderboardItemRanked) *RunningTotal {
	var runningTotal = &RunningTotal{
		Languages: []Language{},
	}
	return CalculateNewRunningTotal(runningTotal, entry.Total)
}

func NewLeaderboardEntry(entry *models.LeaderboardItemRanked) *LeaderboardEntry {
	return &LeaderboardEntry{
		Rank:         int(entry.Rank),
		RunningTotal: NewRunningTotal(entry),
		User:         NewFromUser(entry.User),
	}
}

func NewLeaderboardViewModel(leaderboard *models.Leaderboard, user *models.User) *LeaderboardViewModel {
	var entries []*LeaderboardEntry
	var currentUserEntry *CurrentUser
	for _, entry := range *leaderboard {
		if user != nil && entry.User.ID == user.ID &&
			(currentUserEntry == nil || currentUserEntry.Rank < int(entry.Rank)) {
			currentUserEntry = &CurrentUser{
				Rank: int(entry.Rank),
				Page: 1,
				User: NewFromUser(user),
			}
		}

		if foundEntry, found := slice.FindBy[*LeaderboardEntry](entries, func(i int, entry2 *LeaderboardEntry) bool {
			return entry2.User.ID == entry.User.ID
		}); found {
			foundEntry.RunningTotal = CalculateNewRunningTotal(foundEntry.RunningTotal,
				time.Duration(foundEntry.RunningTotal.TotalSeconds+entry.Total.Seconds())*time.Second)
			foundEntry.RunningTotal.Languages = append(foundEntry.RunningTotal.Languages, Language{
				Name:         *entry.Key,
				TotalSeconds: entry.Total.Seconds(),
			})
			continue
		}
		entries = append(entries, NewLeaderboardEntry(entry))
	}

	return &LeaderboardViewModel{
		CurrentUser: currentUserEntry,
		Data:        entries,
	}
}
