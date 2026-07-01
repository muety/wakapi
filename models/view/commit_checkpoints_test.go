package view

import (
	"testing"
	"time"

	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCommitCheckpointProjectBuildsAccumulatingTimelineWithCommitResets(t *testing.T) {
	from := time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC)
	to := time.Date(2026, 6, 10, 13, 0, 0, 0, time.UTC)
	firstWork := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	secondWork := time.Date(2026, 6, 10, 11, 0, 0, 0, time.UTC)
	firstCommit := time.Date(2026, 6, 10, 10, 30, 0, 0, time.UTC)
	secondCommit := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	repo := &models.ScmRepository{
		FullName: "igbenic/wakapi",
		HTMLURL:  "https://github.com/igbenic/wakapi",
	}
	durations := models.Durations{
		{
			Time:     models.CustomTime(firstWork),
			Duration: time.Hour,
			Project:  "wakapi",
			Branch:   "master",
		},
		{
			Time:     models.CustomTime(secondWork),
			Duration: 30 * time.Minute,
			Project:  "wakapi",
			Branch:   "master",
		},
	}
	commits := map[string]*models.ScmCommit{
		"earlier": {
			Hash:          "earlier",
			TruncatedHash: "earlier",
			Message:       "Add commit sync",
			CommitterDate: models.CustomTime(firstCommit),
		},
		"later": {
			Hash:          "later",
			TruncatedHash: "later12",
			Message:       "Improve summaries\n\nBody is not shown in the overview.",
			HTMLURL:       "https://example.test/later",
			CommitterDate: models.CustomTime(secondCommit),
		},
		"zero": {
			Hash:          "zero",
			TruncatedHash: "zero",
			Message:       "No tracked time",
			CommitterDate: models.CustomTime(to.Add(time.Hour)),
		},
	}
	stats := []*models.CommitStat{
		{CommitHash: "later", TotalSeconds: 7200, HumanReadableTotal: "2 hrs"},
		{CommitHash: "zero", TotalSeconds: 0, HumanReadableTotal: "0 secs"},
		{CommitHash: "earlier", TotalSeconds: 3600, HumanReadableTotal: "1 hr"},
	}

	result := NewCommitCheckpointProject("wakapi", "master", repo, durations, stats, commits, from, to)

	require.NotNil(t, result)
	assert.Equal(t, "wakapi", result.Project)
	assert.Equal(t, "igbenic/wakapi", result.Repository)
	assert.Equal(t, "master", result.Branch)
	assert.Equal(t, 5400.0, result.TotalSeconds)
	assert.Equal(t, 3600.0, result.MaxAccumulatedSeconds)
	require.Len(t, result.Commits, 2)
	assert.Equal(t, "Add commit sync", result.Commits[0].MessageTitle)
	assert.Equal(t, "Improve summaries", result.Commits[1].MessageTitle)

	points := result.Points
	require.Len(t, points, 10)
	assert.Equal(t, from, points[0].Time)
	assert.Equal(t, 0.0, points[0].TotalSeconds)
	assert.Equal(t, firstWork.Add(time.Hour), points[2].Time)
	assert.Equal(t, 3600.0, points[2].TotalSeconds)
	assert.Equal(t, firstCommit, points[3].Time)
	assert.Equal(t, 3600.0, points[3].TotalSeconds)
	assert.Equal(t, "commit", points[3].Kind)
	assert.Equal(t, firstCommit, points[4].Time)
	assert.Equal(t, 0.0, points[4].TotalSeconds)
	assert.Equal(t, "reset", points[4].Kind)
	assert.Equal(t, secondWork.Add(30*time.Minute), points[6].Time)
	assert.Equal(t, 1800.0, points[6].TotalSeconds)
	assert.Equal(t, secondCommit, points[7].Time)
	assert.Equal(t, 1800.0, points[7].TotalSeconds)
	assert.Equal(t, secondCommit, points[8].Time)
	assert.Equal(t, 0.0, points[8].TotalSeconds)
	assert.Equal(t, to, points[9].Time)
	assert.Equal(t, 0.0, points[9].TotalSeconds)
}
