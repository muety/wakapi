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
	args := m.Called(u, t, t2)
	return args.Get(0).([]*models.Duration), args.Error(1)
}

func (m *DurationRepositoryMock) DeleteByUser(u *models.User) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *DurationRepositoryMock) DeleteByUserBefore(u *models.User, t time.Time) error {
	args := m.Called(u, t)
	return args.Error(0)
}
