package mocks

import (
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/models/types"
	"github.com/stretchr/testify/mock"
	"time"
)

type SummaryServiceMock struct {
	mock.Mock
}

func (m *SummaryServiceMock) Aliased(t time.Time, t2 time.Time, u *models.User, r types.SummaryRetriever, f *models.Filters, d *time.Duration, b bool) (*models.Summary, error) {
	args := m.Called(t, t2, u, r, f, d, b)
	return args.Get(0).(*models.Summary), args.Error(1)
}

func (m *SummaryServiceMock) Retrieve(t time.Time, t2 time.Time, u *models.User, f *models.Filters, d *time.Duration) (*models.Summary, error) {
	args := m.Called(t, t2, u, d, f)
	return args.Get(0).(*models.Summary), args.Error(1)
}

func (m *SummaryServiceMock) Summarize(t time.Time, t2 time.Time, u *models.User, f *models.Filters, d *time.Duration) (*models.Summary, error) {
	args := m.Called(t, t2, u, d, f)
	return args.Get(0).(*models.Summary), args.Error(1)
}

func (m *SummaryServiceMock) GetLatestByUser() ([]*models.TimeByUser, error) {
	args := m.Called()
	return args.Get(0).([]*models.TimeByUser), args.Error(1)
}

func (m *SummaryServiceMock) DeleteByUser(s string) error {
	args := m.Called(s)
	return args.Error(0)
}

func (m *SummaryServiceMock) DeleteByUserBefore(s string, t time.Time) error {
	args := m.Called(s, t)
	return args.Error(0)
}

func (m *SummaryServiceMock) Insert(s *models.Summary) error {
	args := m.Called(s)
	return args.Error(0)
}
