package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSummary_FillMissing(t *testing.T) {
	testDuration := 10 * time.Minute

	sut := &Summary{
		Projects: []*SummaryItem{
			{
				Type: SummaryProject,
				Key:  "wakapi",
				// hack to work around the issue that the total time of a summary item is mistakenly represented in seconds
				Total: testDuration / time.Second,
			},
		},
	}

	sut.FillMissing()

	itemLists := [][]*SummaryItem{
		sut.Machines,
		sut.OperatingSystems,
		sut.Languages,
		sut.Editors,
	}
	for _, l := range itemLists {
		assert.Len(t, l, 1)
		assert.Equal(t, UnknownSummaryKey, l[0].Key)
		assert.Equal(t, testDuration, l[0].TotalFixed())
	}

	assert.Len(t, sut.Labels, 1)
	assert.Equal(t, DefaultProjectLabel, sut.Labels[0].Key)
	assert.Equal(t, testDuration, sut.Labels[0].TotalFixed())
}

func TestSummary_TotalTimeBy(t *testing.T) {
	testDuration1, testDuration2, testDuration3 := 10*time.Minute, 5*time.Minute, 20*time.Minute

	sut := &Summary{
		Projects: []*SummaryItem{
			{
				Type: SummaryProject,
				Key:  "wakapi",
				// hack to work around the issue that the total time of a summary item is mistakenly represented in seconds
				Total: testDuration1 / time.Second,
			},
			{
				Type:  SummaryProject,
				Key:   "anchr",
				Total: testDuration2 / time.Second,
			},
		},
		Languages: []*SummaryItem{
			{
				Type:  SummaryLanguage,
				Key:   "Go",
				Total: testDuration3 / time.Second,
			},
		},
	}

	assert.Equal(t, testDuration1+testDuration2, sut.TotalTimeBy(SummaryProject))
	assert.Equal(t, testDuration3, sut.TotalTimeBy(SummaryLanguage))
	assert.Zero(t, sut.TotalTimeBy(SummaryEditor))
	assert.Zero(t, sut.TotalTimeBy(SummaryMachine))
	assert.Zero(t, sut.TotalTimeBy(SummaryOS))
}

func TestSummary_TotalTimeByFilters(t *testing.T) {
	testDuration1, testDuration2, testDuration3 := 10*time.Minute, 5*time.Minute, 20*time.Minute

	sut := &Summary{
		Projects: []*SummaryItem{
			{
				Type: SummaryProject,
				Key:  "wakapi",
				// hack to work around the issue that the total time of a summary item is mistakenly represented in seconds
				Total: testDuration1 / time.Second,
			},
			{
				Type:  SummaryProject,
				Key:   "anchr",
				Total: testDuration2 / time.Second,
			},
		},
		Languages: []*SummaryItem{
			{
				Type:  SummaryLanguage,
				Key:   "Go",
				Total: testDuration3 / time.Second,
			},
		},
	}

	filters1 := NewFiltersWith(SummaryProject, "wakapi").OneOrEmpty()
	filters2 := NewFiltersWith(SummaryLanguage, "Go").OneOrEmpty()
	filters3 := FilterElement{}

	assert.Equal(t, testDuration1, sut.TotalTimeByFilter(filters1))
	assert.Equal(t, testDuration3, sut.TotalTimeByFilter(filters2))
	assert.Zero(t, sut.TotalTimeByFilter(filters3))
}

func TestSummary_WithResolvedAliases(t *testing.T) {
	testDuration1, testDuration2, testDuration3, testDuration4 := 10*time.Minute, 5*time.Minute, 1*time.Minute, 20*time.Minute

	var resolver AliasResolver = func(t uint8, k string) string {
		switch t {
		case SummaryProject:
			switch k {
			case "wakapi-mobile":
				return "wakapi"
			}
		case SummaryLanguage:
			switch k {
			case "Java 8":
				return "Java"
			}
		}
		return k
	}

	sut := &Summary{
		Projects: []*SummaryItem{
			{
				Type:  SummaryProject,
				Key:   "wakapi",
				Total: testDuration1 / time.Second,
			},
			{
				Type:  SummaryProject,
				Key:   "wakapi-mobile",
				Total: testDuration2 / time.Second,
			},
			{
				Type:  SummaryProject,
				Key:   "anchr",
				Total: testDuration3 / time.Second,
			},
		},
		Languages: []*SummaryItem{
			{
				Type:  SummaryLanguage,
				Key:   "Java 8",
				Total: testDuration4 / time.Second,
			},
		},
	}

	sut = sut.WithResolvedAliases(resolver)

	assert.Equal(t, testDuration1+testDuration2, sut.TotalTimeByKey(SummaryProject, "wakapi"))
	assert.Zero(t, sut.TotalTimeByKey(SummaryProject, "wakapi-mobile"))
	assert.Equal(t, testDuration3, sut.TotalTimeByKey(SummaryProject, "anchr"))
	assert.Equal(t, testDuration4, sut.TotalTimeByKey(SummaryLanguage, "Java"))
	assert.Zero(t, sut.TotalTimeByKey(SummaryLanguage, "wakapi"))
	assert.Zero(t, sut.TotalTimeByKey(SummaryProject, "Java 8"))
	assert.Len(t, sut.Projects, 2)
	assert.Len(t, sut.Languages, 1)
	assert.Empty(t, sut.Editors)
	assert.Empty(t, sut.OperatingSystems)
	assert.Empty(t, sut.Machines)
}

func TestSummary_KeepOnly(t *testing.T) {
	newSummary := func() *Summary {
		return &Summary{
			Projects: []*SummaryItem{
				{
					Type: SummaryProject,
					Key:  "wakapi",
					// hack to work around the issue that the total time of a summary item is mistakenly represented in seconds
					Total: 10 * time.Minute / time.Second,
				},
				{
					Type:  SummaryProject,
					Key:   "anchr",
					Total: 10 * time.Minute / time.Second,
				},
			},
			Languages: []*SummaryItem{
				{
					Type:  SummaryLanguage,
					Key:   "Go",
					Total: 10 * time.Minute / time.Second,
				},
			},
			Editors: []*SummaryItem{
				{
					Type:  SummaryEditor,
					Key:   "VSCode",
					Total: 10 * time.Minute / time.Second,
				},
			},
		}
	}

	var sut *Summary

	sut = newSummary().KeepOnly(map[uint8]bool{}) // keep all
	assert.NotZero(t, sut.TotalTimeBy(SummaryProject))
	assert.NotZero(t, sut.TotalTimeBy(SummaryLanguage))
	assert.NotZero(t, sut.TotalTimeBy(SummaryEditor))
	assert.Equal(t, 20*time.Minute, sut.TotalTime())

	sut = newSummary().KeepOnly(map[uint8]bool{SummaryProject: true})
	assert.NotZero(t, sut.TotalTimeBy(SummaryProject))
	assert.Zero(t, sut.TotalTimeBy(SummaryLanguage))
	assert.Zero(t, sut.TotalTimeBy(SummaryEditor))
	assert.Equal(t, 20*time.Minute, sut.TotalTime())

	sut = newSummary().KeepOnly(map[uint8]bool{SummaryEditor: true, SummaryLanguage: true})
	assert.Zero(t, sut.TotalTimeBy(SummaryProject))
	assert.NotZero(t, sut.TotalTimeBy(SummaryLanguage))
	assert.NotZero(t, sut.TotalTimeBy(SummaryEditor))
	assert.Equal(t, 10*time.Minute, sut.TotalTime())

	sut = newSummary().KeepOnly(map[uint8]bool{SummaryProject: true})
	sut.FillMissing()
	assert.NotZero(t, sut.TotalTimeBy(SummaryProject))
	assert.NotZero(t, sut.TotalTimeBy(SummaryLanguage))
	assert.NotZero(t, sut.TotalTimeBy(SummaryEditor))
}

func TestSummaryItems_Sorted(t *testing.T) {
	testDuration1, testDuration2, testDuration3 := 10*time.Minute, 5*time.Minute, 20*time.Minute

	sut := &Summary{
		Projects: []*SummaryItem{
			{
				Type:  SummaryProject,
				Key:   "wakapi",
				Total: testDuration1,
			},
			{
				Type:  SummaryProject,
				Key:   "anchr",
				Total: testDuration2,
			},
			{
				Type:  SummaryProject,
				Key:   "anchr-mobile",
				Total: testDuration3,
			},
		},
	}

	sut = sut.Sorted()

	assert.Equal(t, testDuration3, sut.Projects[0].Total)
	assert.Equal(t, testDuration1, sut.Projects[1].Total)
	assert.Equal(t, testDuration2, sut.Projects[2].Total)
}

func TestSummary_ApplyFilter(t *testing.T) {
	var (
		key1 = "A Banana"
		key2 = "Bananas are delicious"
	)

	sut := Summary{
		Projects: []*SummaryItem{
			{Type: SummaryProject, Key: key1, Total: 10 * time.Minute / time.Second},
			{Type: SummaryProject, Key: key2, Total: 20 * time.Minute / time.Second},
		},
		Languages: []*SummaryItem{
			{Type: SummaryLanguage, Key: "Italian", Total: 30 * time.Minute / time.Second},
		},
	}

	sut.ApplyFilter(FilterElement{
		Entity: SummaryProject,
		Filter: OrFilter{key2},
	})

	assert.Len(t, *sut.GetByType(SummaryProject), 1)
	assert.Len(t, *sut.GetByType(SummaryLanguage), 1)
	assert.Equal(t, key2, sut.Projects[0].Key)
	assert.Equal(t, 20*time.Minute, sut.TotalTimeBy(SummaryProject))
}
