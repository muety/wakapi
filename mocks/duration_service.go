package mocks

import (
	"time"

	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/mock"
)

type DurationServiceMock struct {
	mock.Mock
}

func (m *DurationServiceMock) Get(time time.Time, time2 time.Time, user *models.User, f *models.Filters) (models.Durations, error) {
	args := m.Called(time, time2, user, f)
	return args.Get(0).(models.Durations), args.Error(1)
}

func (m *DurationServiceMock) MakeDurationsFromHeartbeats(options models.ProcessHeartbeatsArgs, f *models.Filters) (models.Durations, error) {
	args := m.Called(options, f)
	return args.Get(0).(models.Durations), args.Error(1)
}
