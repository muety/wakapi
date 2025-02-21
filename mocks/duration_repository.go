package mocks

import (
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/mock"
	"time"
)

type DurationRepositoryMock struct {
	BaseRepositoryMock
	mock.Mock
}

func (m *DurationRepositoryMock) InsertBatch(d []*models.Duration) error {
	args := m.Called(d)
	return args.Error(0)
}

func (m *DurationRepositoryMock) GetAllWithin(t time.Time, t2 time.Time, u *models.User) ([]*models.Duration, error) {
	args := m.Called(t, t2, u)
	return args.Get(0).([]*models.Duration), args.Error(1)
}

func (m *DurationRepositoryMock) GetAllWithinByFilters(t time.Time, t2 time.Time, u *models.User, m2 map[string][]string) ([]*models.Duration, error) {
	args := m.Called(t, t2, u, m2)
	return args.Get(0).([]*models.Duration), args.Error(1)
}

func (m *DurationRepositoryMock) GetLatestByUser(u *models.User) (*models.Duration, error) {
	args := m.Called(u)
	return args.Get(0).(*models.Duration), args.Error(1)
}

func (m *DurationRepositoryMock) StreamAllWithin(t time.Time, t2 time.Time, u *models.User) (chan *models.Duration, error) {
	args := m.Called(t, t2, u)
	return args.Get(0).(chan *models.Duration), args.Error(1)
}

func (m *DurationRepositoryMock) StreamAllWithinByFilters(t time.Time, t2 time.Time, u *models.User, m2 map[string][]string) (chan *models.Duration, error) {
	args := m.Called(t, t2, u, m2)
	return args.Get(0).(chan *models.Duration), args.Error(1)
}

func (m *DurationRepositoryMock) DeleteByUser(u *models.User) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *DurationRepositoryMock) DeleteByUserBefore(u *models.User, t time.Time) error {
	args := m.Called(u, t)
	return args.Error(0)
}
