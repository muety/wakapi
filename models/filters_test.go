package models

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type FiltersTestSuite struct {
	suite.Suite
	TestAliases                    []*Alias
	TestProjectLabels              []*ProjectLabel
	GetAliasReverseResolver        func(indices []int) AliasReverseResolver
	GetProjectLabelReverseResolver func(indices []int) ProjectLabelReverseResolver
}

func (suite *FiltersTestSuite) SetupSuite() {
	suite.TestAliases = []*Alias{
		{
			Type:  SummaryProject,
			Key:   "wakapi",
			Value: "wakapi-mobile",
		},
		{
			Type:  SummaryProject,
			Key:   "wakapi",
			Value: "wakapi-desktop",
		},
		{
			Type:  SummaryLanguage,
			Key:   "Python",
			Value: "Python 3",
		},
	}

	suite.TestProjectLabels = []*ProjectLabel{
		{
			ProjectKey: "wakapi",
			Label:      "oss",
		},
		{
			ProjectKey: "anchr",
			Label:      "oss",
		},
		{
			ProjectKey: "business-application",
			Label:      "work",
		},
	}

	suite.GetAliasReverseResolver = func(indices []int) AliasReverseResolver {
		return func(t uint8, k string) []string {
			aliases := make([]string, 0, len(indices))
			for _, j := range indices {
				if a := suite.TestAliases[j]; a.Type == t && a.Key == k {
					aliases = append(aliases, a.Value)
				}
			}
			return aliases
		}
	}

	suite.GetProjectLabelReverseResolver = func(indices []int) ProjectLabelReverseResolver {
		return func(k string) []string {
			labels := make([]string, 0, len(indices))
			for _, j := range indices {
				if l := suite.TestProjectLabels[j]; l.Label == k {
					labels = append(labels, l.ProjectKey)
				}
			}
			return labels
		}
	}
}

func TestFiltersTestSuite(t *testing.T) {
	suite.Run(t, new(FiltersTestSuite))
}

func (suite *FiltersTestSuite) TestFilters_IsEmpty() {
	assert.False(suite.T(), NewFiltersWith(SummaryProject, "wakapi").IsEmpty())
	assert.True(suite.T(), (&Filters{}).IsEmpty())
}

func (suite *FiltersTestSuite) TestFilters_Match() {
	heartbeats := []*Heartbeat{
		{Project: "wakapi", Language: "Go"},
		{Project: "anchr", Language: "Javascript"},
	}

	sut1 := NewFiltersWith(SummaryProject, "wakapi")
	assert.True(suite.T(), sut1.Match(heartbeats[0]))
	assert.False(suite.T(), sut1.Match(heartbeats[1]))

	sut2 := NewFiltersWith(SummaryProject, "Go").With(SummaryLanguage, "JavaScript")
	assert.False(suite.T(), sut2.Match(heartbeats[0]))
	assert.False(suite.T(), sut2.Match(heartbeats[1]))

	sut3 := NewFilterWithMultiple(SummaryProject, []string{"wakapi", "anchr"})
	assert.True(suite.T(), sut3.Match(heartbeats[0]))
	assert.True(suite.T(), sut3.Match(heartbeats[1]))

	sut4 := &Filters{}
	assert.True(suite.T(), sut4.Match(heartbeats[0]))
	assert.True(suite.T(), sut4.Match(heartbeats[1]))
}

func (suite *FiltersTestSuite) TestFilters_One() {
	sut1 := NewFiltersWith(SummaryLanguage, "Java")
	ok1, type1, filters1 := sut1.One()
	assert.True(suite.T(), ok1)
	assert.Equal(suite.T(), SummaryLanguage, type1)
	assert.Equal(suite.T(), "Java", filters1[0])

	sut2 := &Filters{}
	ok2, type2, filters2 := sut2.One()
	assert.False(suite.T(), ok2)
	assert.Zero(suite.T(), type2)
	assert.Empty(suite.T(), filters2)
}

func (suite *FiltersTestSuite) TestFilters_WithAliases() {
	sut1 := NewFiltersWith(SummaryProject, "wakapi")
	sut1 = sut1.WithAliases(suite.GetAliasReverseResolver([]int{0, 1, 2}))
	assert.Len(suite.T(), sut1.Project, 3)
	assert.Len(suite.T(), sut1.Language, 0)
	assert.Contains(suite.T(), sut1.Project, "wakapi")
	assert.Contains(suite.T(), sut1.Project, "wakapi-desktop")
	assert.Contains(suite.T(), sut1.Project, "wakapi-mobile")

	sut2 := NewFiltersWith(SummaryProject, "wakapi").With(SummaryLanguage, "Python")
	sut2 = sut2.WithAliases(suite.GetAliasReverseResolver([]int{0, 1, 2}))
	assert.Len(suite.T(), sut2.Project, 3)
	assert.Len(suite.T(), sut2.Language, 2)
	assert.Contains(suite.T(), sut2.Language, "Python")
	assert.Contains(suite.T(), sut2.Language, "Python 3")

	sut3 := NewFiltersWith(SummaryProject, "foo")
	sut3 = sut3.WithAliases(suite.GetAliasReverseResolver([]int{0, 1, 2}))
	assert.Len(suite.T(), sut3.Project, 1)
	assert.Len(suite.T(), sut3.Language, 0)
	assert.Contains(suite.T(), sut3.Project, "foo")
}

func (suite *FiltersTestSuite) TestFilters_WithProjectLabels() {
	sut1 := NewFiltersWith(SummaryProject, "mailwhale").With(SummaryLabel, "oss")
	sut1 = sut1.WithProjectLabels(suite.GetProjectLabelReverseResolver([]int{0, 1, 2}))
	assert.Len(suite.T(), sut1.Project, 3)
	assert.Contains(suite.T(), sut1.Project, "wakapi")
	assert.Contains(suite.T(), sut1.Project, "anchr")
	assert.Contains(suite.T(), sut1.Project, "mailwhale")
	assert.Contains(suite.T(), sut1.Label, "oss")

	sut2 := NewFiltersWith(SummaryLabel, "oss")
	sut2 = sut2.WithProjectLabels(suite.GetProjectLabelReverseResolver([]int{0, 1, 2}))
	assert.Len(suite.T(), sut2.Project, 2)
	assert.Contains(suite.T(), sut2.Project, "wakapi")
	assert.Contains(suite.T(), sut2.Project, "anchr")
	assert.Contains(suite.T(), sut2.Label, "oss")
}
