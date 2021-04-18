package mocks

import (
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/mock"
	"time"
)

type SummaryRepositoryMock struct {
	mock.Mock
}

func (m *SummaryRepositoryMock) Insert(summary *models.Summary) error {
	args := m.Called(summary)
	return args.Error(0)
}

func (m *SummaryRepositoryMock) GetAll() ([]*models.Summary, error) {
	args := m.Called()
	return args.Get(0).([]*models.Summary), args.Error(1)
}

func (m *SummaryRepositoryMock) GetByUserWithin(user *models.User, time time.Time, time2 time.Time) ([]*models.Summary, error) {
	args := m.Called(user, time, time2)
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
