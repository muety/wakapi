package mocks

import (
	datastructure "github.com/duke-git/lancet/v2/datastructure/set"
	"github.com/muety/wakapi/models/types"
	"github.com/stretchr/testify/mock"
)

type AggregationServiceMock struct {
	mock.Mock
}

func (m *AggregationServiceMock) Schedule() {
	m.Called()
}

func (m *AggregationServiceMock) AggregateSummaries(userIds datastructure.Set[string], durations types.DurationsRegenerationMode) error {
	args := m.Called(userIds, durations)
	return args.Error(0)
}

func (m *AggregationServiceMock) AggregateDurations(userIds datastructure.Set[string]) error {
	args := m.Called(userIds)
	return args.Error(0)
}

func (m *AggregationServiceMock) IsLocked(userId string) bool {
	args := m.Called(userId)
	return args.Bool(0)
}
