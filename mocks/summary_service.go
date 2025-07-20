package mocks

import (
	"time"

	"github.com/muety/wakapi/models"
	summarytypes "github.com/muety/wakapi/types"
	"github.com/stretchr/testify/mock"
)


type SummaryServiceMock struct {
	mock.Mock
}


func (m *SummaryServiceMock) Generate(request *summarytypes.SummaryRequest, options *summarytypes.ProcessingOptions) (*models.Summary, error) {
	args := m.Called(request, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Summary), args.Error(1)
}

func (m *SummaryServiceMock) QuickSummary(from, to time.Time, user *models.User) (*models.Summary, error) {
	args := m.Called(from, to, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Summary), args.Error(1)
}

func (m *SummaryServiceMock) DetailedSummary(request *summarytypes.SummaryRequest) (*models.Summary, error) {
	args := m.Called(request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Summary), args.Error(1)
}

func (m *SummaryServiceMock) RetrieveFromStorage(request *summarytypes.SummaryRequest) (*models.Summary, error) {
	args := m.Called(request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Summary), args.Error(1)
}

func (m *SummaryServiceMock) ComputeFromDurations(request *summarytypes.SummaryRequest) (*models.Summary, error) {
	args := m.Called(request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Summary), args.Error(1)
}


func (m *SummaryServiceMock) GetHeartbeatsWritePercentage(userID string, start, end time.Time) (float64, error) {
	args := m.Called(userID, start, end)
	return 50, args.Error(0)
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
