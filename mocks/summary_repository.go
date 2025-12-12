package mocks

import (
	"time"

	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/mock"
)

type SummaryRepositoryMock struct {
	BaseRepositoryMock
	mock.Mock
}

func (m *SummaryRepositoryMock) Insert(s *models.Summary) error {
	args := m.Called(s)
	return args.Error(0)
}

func (m *SummaryRepositoryMock) InsertWithRetry(s *models.Summary) error {
	args := m.Called(s)
	return args.Error(0)
}

func (m *SummaryRepositoryMock) GetAll() ([]*models.Summary, error) {
	args := m.Called()
	return args.Get(0).([]*models.Summary), args.Error(1)
}

func (m *SummaryRepositoryMock) GetByUserWithin(u *models.User, t1 time.Time, t2 time.Time) ([]*models.Summary, error) {
	args := m.Called(u, t1, t2)
	return args.Get(0).([]*models.Summary), args.Error(1)
}

func (m *SummaryRepositoryMock) GetLastByUser() ([]*models.TimeByUser, error) {
	args := m.Called()
	return args.Get(0).([]*models.TimeByUser), args.Error(1)
}

func (m *SummaryRepositoryMock) DeleteByUser(s string) error {
	args := m.Called(s)
	return args.Error(0)
}

func (m *SummaryRepositoryMock) DeleteByUserBefore(s string, t time.Time) error {
	args := m.Called(s, t)
	return args.Error(0)
}
