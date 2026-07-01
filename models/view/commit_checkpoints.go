package view

import (
	"sort"
	"strings"
	"time"

	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/models"
)

type CommitCheckpointProject struct {
	Project               string                   `json:"project"`
	Repository            string                   `json:"repository"`
	RepositoryURL         string                   `json:"repository_url"`
	Branch                string                   `json:"branch"`
	From                  time.Time                `json:"from"`
	To                    time.Time                `json:"to"`
	TotalSeconds          float64                  `json:"total_seconds"`
	HumanReadableTotal    string                   `json:"human_readable_total"`
	MaxAccumulatedSeconds float64                  `json:"max_accumulated_seconds"`
	Points                []*CommitCheckpointPoint `json:"points"`
	Commits               []*CommitCheckpoint      `json:"commits"`
}

type CommitCheckpoint struct {
	Hash               string    `json:"hash"`
	TruncatedHash      string    `json:"truncated_hash"`
	Message            string    `json:"message"`
	MessageTitle       string    `json:"message_title"`
	HTMLURL            string    `json:"html_url"`
	CommittedAt        time.Time `json:"committed_at"`
	AssignedSeconds    float64   `json:"assigned_seconds"`
	HumanReadableTotal string    `json:"human_readable_total"`
}

type CommitCheckpointPoint struct {
	Time         time.Time `json:"time"`
	TotalSeconds float64   `json:"total_seconds"`
	Kind         string    `json:"kind"`
	CommitHash   string    `json:"commit_hash,omitempty"`
	CommitTitle  string    `json:"commit_title,omitempty"`
}

func NewCommitCheckpointProject(project, branch string, repo *models.ScmRepository, durations models.Durations, stats []*models.CommitStat, commits map[string]*models.ScmCommit, from, to time.Time) *CommitCheckpointProject {
	if repo == nil {
		repo = &models.ScmRepository{}
	}
	if branch == "" {
		branch = repo.DefaultBranch
	}

	result := &CommitCheckpointProject{
		Project:       project,
		Repository:    repo.FullName,
		RepositoryURL: repo.HTMLURL,
		Branch:        branch,
		From:          from,
		To:            to,
		Points:        []*CommitCheckpointPoint{{Time: from, TotalSeconds: 0, Kind: "start"}},
		Commits:       make([]*CommitCheckpoint, 0, len(stats)),
	}

	for _, stat := range stats {
		if stat == nil {
			continue
		}
		commit := commits[stat.CommitHash]
		if commit == nil {
			continue
		}
		committedAt := commitTime(commit)
		if committedAt.Before(from) || committedAt.After(to) {
			continue
		}

		result.Commits = append(result.Commits, &CommitCheckpoint{
			Hash:               commit.Hash,
			TruncatedHash:      commit.TruncatedHash,
			Message:            commit.Message,
			MessageTitle:       commitMessageTitle(commit),
			HTMLURL:            commit.HTMLURL,
			CommittedAt:        committedAt,
			AssignedSeconds:    stat.TotalSeconds,
			HumanReadableTotal: commitHumanTotal(stat),
		})
	}

	sort.SliceStable(result.Commits, func(i, j int) bool {
		return result.Commits[i].CommittedAt.Before(result.Commits[j].CommittedAt)
	})

	result.Points = append(result.Points, buildCommitTimelinePoints(durations, result.Commits, from, to, &result.TotalSeconds, &result.MaxAccumulatedSeconds)...)
	result.HumanReadableTotal = helpers.FmtWakatimeDuration(secondsDuration(result.TotalSeconds))

	return result
}

func buildCommitTimelinePoints(durations models.Durations, commits []*CommitCheckpoint, from, to time.Time, totalSeconds, maxAccumulatedSeconds *float64) []*CommitCheckpointPoint {
	points := make([]*CommitCheckpointPoint, 0)
	if !to.After(from) {
		return points
	}

	sort.SliceStable(durations, func(i, j int) bool {
		return durations[i].Time.T().Before(durations[j].Time.T())
	})

	commitIndex := 0
	accumulated := 0.0
	lastPointTime := from

	addPoint := func(t time.Time, seconds float64, kind, hash, title string) {
		if t.Before(from) {
			t = from
		}
		if t.After(to) {
			t = to
		}
		points = append(points, &CommitCheckpointPoint{
			Time:         t,
			TotalSeconds: seconds,
			Kind:         kind,
			CommitHash:   hash,
			CommitTitle:  title,
		})
		if seconds > *maxAccumulatedSeconds {
			*maxAccumulatedSeconds = seconds
		}
		lastPointTime = t
	}

	resetAtCommit := func(commit *CommitCheckpoint) {
		if commit == nil {
			return
		}
		addPoint(commit.CommittedAt, accumulated, "commit", commit.Hash, commit.MessageTitle)
		accumulated = 0
		addPoint(commit.CommittedAt, accumulated, "reset", commit.Hash, commit.MessageTitle)
	}

	for _, duration := range durations {
		if duration == nil {
			continue
		}
		start := duration.Time.T()
		end := duration.TimeEnd()
		if !end.After(from) || !start.Before(to) {
			continue
		}
		if start.Before(from) {
			start = from
		}
		if end.After(to) {
			end = to
		}
		if !end.After(start) {
			continue
		}

		for commitIndex < len(commits) && !commits[commitIndex].CommittedAt.After(start) {
			resetAtCommit(commits[commitIndex])
			commitIndex++
		}

		if lastPointTime.Before(start) {
			addPoint(start, accumulated, "hold", "", "")
		}

		cursor := start
		for commitIndex < len(commits) && !commits[commitIndex].CommittedAt.After(end) {
			commit := commits[commitIndex]
			if commit.CommittedAt.After(cursor) {
				seconds := commit.CommittedAt.Sub(cursor).Seconds()
				accumulated += seconds
				*totalSeconds += seconds
				addPoint(commit.CommittedAt, accumulated, "work", "", "")
			}
			resetAtCommit(commit)
			cursor = commit.CommittedAt
			commitIndex++
		}

		if end.After(cursor) {
			seconds := end.Sub(cursor).Seconds()
			accumulated += seconds
			*totalSeconds += seconds
			addPoint(end, accumulated, "work", "", "")
		}
	}

	for commitIndex < len(commits) {
		resetAtCommit(commits[commitIndex])
		commitIndex++
	}

	if lastPointTime.Before(to) {
		addPoint(to, accumulated, "end", "", "")
	}

	return points
}

func commitTime(commit *models.ScmCommit) time.Time {
	if t := commit.CommitterDate.T(); !t.IsZero() {
		return t
	}
	return commit.AuthorDate.T()
}

func commitMessageTitle(commit *models.ScmCommit) string {
	for _, line := range strings.Split(commit.Message, "\n") {
		if title := strings.TrimSpace(line); title != "" {
			return title
		}
	}
	if commit.TruncatedHash != "" {
		return commit.TruncatedHash
	}
	return commit.Hash
}

func commitHumanTotal(stat *models.CommitStat) string {
	if stat.HumanReadableTotal != "" {
		return stat.HumanReadableTotal
	}
	return helpers.FmtWakatimeDuration(secondsDuration(stat.TotalSeconds))
}

func secondsDuration(seconds float64) time.Duration {
	return time.Duration(seconds * float64(time.Second))
}
