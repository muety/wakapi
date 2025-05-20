package mocks

import (
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/mock"
	"time"
)

type DurationServiceMock struct {
	mock.Mock
}

func (m *DurationServiceMock) Get(time time.Time, time2 time.Time, user *models.User, f *models.Filters, d *time.Duration, b bool) (models.Durations, error) {
	args := m.Called(time, time2, user, f, d, b)
	return args.Get(0).(models.Durations), args.Error(1)
}

func (m *DurationServiceMock) Regenerate(u *models.User, b bool) {
	m.Called(u, b)
}

func (m *DurationServiceMock) RegenerateAll() {
}

func (m *DurationServiceMock) DeleteByUser(u *models.User) error {
	args := m.Called(u)
	return args.Error(0)
}
