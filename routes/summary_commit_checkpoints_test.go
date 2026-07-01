package routes

import (
	"testing"
	"time"

	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type summaryCheckpointCommitService struct {
	links            []*services.ProjectLinkInfo
	resultsByProject map[string]*services.CommitsResult
	projectsCalled   []string
	dateFrom         *time.Time
	dateTo           *time.Time
}

func (s *summaryCheckpointCommitService) LinkProject(*models.User, string, string, string, string) (*models.ProjectRepositoryLink, error) {
	return nil, nil
}

func (s *summaryCheckpointCommitService) LinkProjectWithRepo(*models.User, string, string, string) (*models.ProjectRepositoryLink, error) {
	return nil, nil
}

func (s *summaryCheckpointCommitService) GetCommits(_ *models.User, project, _ string, _ string, _ int, _ int, dateFrom, dateTo *time.Time) (*services.CommitsResult, error) {
	s.projectsCalled = append(s.projectsCalled, project)
	s.dateFrom = dateFrom
	s.dateTo = dateTo
	return s.resultsByProject[project], nil
}

func (s *summaryCheckpointCommitService) GetCommit(*models.User, string, string, string, string) (*services.CommitResult, error) {
	return nil, nil
}

func (s *summaryCheckpointCommitService) ListLinks(*models.User) ([]*services.ProjectLinkInfo, error) {
	return s.links, nil
}

func (s *summaryCheckpointCommitService) ListRepos(*models.User, string, int, int) ([]*models.ScmRepository, error) {
	return nil, nil
}

func (s *summaryCheckpointCommitService) UpdateLink(*models.User, string, string, string) error {
	return nil
}

func (s *summaryCheckpointCommitService) UpdateLinkByID(*models.User, string, string, string) error {
	return nil
}

func (s *summaryCheckpointCommitService) UnlinkProject(*models.User, string, bool) error {
	return nil
}

func (s *summaryCheckpointCommitService) UnlinkByID(*models.User, string, bool) error {
	return nil
}

func (s *summaryCheckpointCommitService) UpdateToken(*models.User, string) error {
	return nil
}

func (s *summaryCheckpointCommitService) DeleteToken(*models.User) error {
	return nil
}

func (s *summaryCheckpointCommitService) HasToken(*models.User) (bool, error) {
	return true, nil
}

func (s *summaryCheckpointCommitService) SyncNow(*models.User, string) error {
	return nil
}

func (s *summaryCheckpointCommitService) SyncByID(*models.User, string) error {
	return nil
}

func (s *summaryCheckpointCommitService) Schedule() {}

type summaryCheckpointDurationService struct {
	durations     models.Durations
	filters       *models.Filters
	projectsCalls int
}

func (s *summaryCheckpointDurationService) Get(_ time.Time, _ time.Time, _ *models.User, filters *models.Filters, _ *time.Duration, _ bool) (models.Durations, error) {
	s.filters = filters
	s.projectsCalls++
	return s.durations, nil
}

func (s *summaryCheckpointDurationService) Regenerate(*models.User, bool) {}

func (s *summaryCheckpointDurationService) RegenerateAll() {}

func (s *summaryCheckpointDurationService) DeleteByUser(*models.User) error {
	return nil
}

func TestSummaryHandlerFetchCommitCheckpointsKeepsRelevantLinkedProjects(t *testing.T) {
	from := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	user := &models.User{ID: "user1"}
	repo := &models.ScmRepository{
		ID:            "repo1",
		FullName:      "igbenic/wakapi",
		DefaultBranch: "master",
	}

	commitService := &summaryCheckpointCommitService{
		links: []*services.ProjectLinkInfo{
			{
				Link: &models.ProjectRepositoryLink{Project: "wakapi"},
				Repo: repo,
			},
			{
				Link: &models.ProjectRepositoryLink{Project: "unrelated"},
				Repo: &models.ScmRepository{ID: "repo2", FullName: "igbenic/unrelated"},
			},
		},
		resultsByProject: map[string]*services.CommitsResult{
			"wakapi": {
				Repo:   repo,
				Branch: "master",
				Stats: []*models.CommitStat{
					{CommitHash: "hash1", TotalSeconds: 5400, HumanReadableTotal: "1 hr 30 mins"},
				},
				Commits: map[string]*models.ScmCommit{
					"hash1": {
						Hash:          "hash1",
						TruncatedHash: "hash1",
						Message:       "Add checkpoints",
						CommitterDate: models.CustomTime(from.Add(4 * time.Hour)),
					},
				},
			},
		},
	}
	durationService := &summaryCheckpointDurationService{
		durations: models.Durations{
			{
				Time:     models.CustomTime(from.Add(2 * time.Hour)),
				Duration: 90 * time.Minute,
				Project:  "wakapi",
				Branch:   "master",
			},
			{
				Time:     models.CustomTime(from.Add(3 * time.Hour)),
				Duration: 30 * time.Minute,
				Project:  "unrelated",
				Branch:   "master",
			},
		},
	}
	handler := &SummaryHandler{commitSrvc: commitService, durationSrvc: durationService}
	summary := &models.Summary{
		Projects: []*models.SummaryItem{
			{Type: models.SummaryProject, Key: "wakapi", Total: 5400},
		},
	}
	params := &models.SummaryParams{From: from, To: to, User: user}

	result := handler.fetchCommitCheckpoints(user, summary, params)

	require.Len(t, result, 1)
	assert.Equal(t, "wakapi", result[0].Project)
	assert.Equal(t, float64(5400), result[0].TotalSeconds)
	assert.GreaterOrEqual(t, len(result[0].Points), 4)
	assert.Equal(t, []string{"wakapi"}, commitService.projectsCalled)
	assert.Equal(t, 1, durationService.projectsCalls)
	assert.Nil(t, durationService.filters.Project)
	assert.Equal(t, models.OrFilter{"master"}, durationService.filters.Branch)
	require.NotNil(t, commitService.dateFrom)
	require.NotNil(t, commitService.dateTo)
	assert.Equal(t, from, *commitService.dateFrom)
	assert.Equal(t, to, *commitService.dateTo)
}
