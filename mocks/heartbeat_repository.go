package mocks

import (
	"time"

	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/mock"
)

type HeartbeatRepositoryMock struct {
	BaseRepositoryMock
	mock.Mock
}

func (m *HeartbeatRepositoryMock) InsertBatch(h []*models.Heartbeat) error {
	args := m.Called(h)
	return args.Error(0)
}

func (m *HeartbeatRepositoryMock) GetAll() ([]*models.Heartbeat, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]*models.Heartbeat), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *HeartbeatRepositoryMock) GetWithin(from, to time.Time, user *models.User) ([]*models.Heartbeat, error) {
	args := m.Called(from, to, user)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.Heartbeat), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *HeartbeatRepositoryMock) GetAllWithinByFilters(from, to time.Time, user *models.User, filters map[string][]string) ([]*models.Heartbeat, error) {
	args := m.Called(from, to, user, filters)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.Heartbeat), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *HeartbeatRepositoryMock) GetLatestByFilters(user *models.User, filters map[string][]string) (*models.Heartbeat, error) {
	args := m.Called(user, filters)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Heartbeat), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *HeartbeatRepositoryMock) GetFirstAll() ([]*models.TimeByUser, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]*models.TimeByUser), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *HeartbeatRepositoryMock) GetLastAll() ([]*models.TimeByUser, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]*models.TimeByUser), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *HeartbeatRepositoryMock) GetRangeByUser(user *models.User) (*models.RangeByUser, error) {
	args := m.Called(user)
	if args.Get(0) != nil {
		return args.Get(0).(*models.RangeByUser), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *HeartbeatRepositoryMock) GetLatestByUser(user *models.User) (*models.Heartbeat, error) {
	args := m.Called(user)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Heartbeat), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *HeartbeatRepositoryMock) GetLatestByOriginAndUser(origin string, user *models.User) (*models.Heartbeat, error) {
	args := m.Called(origin, user)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Heartbeat), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *HeartbeatRepositoryMock) StreamWithin(from, to time.Time, user *models.User) (chan *models.Heartbeat, error) {
	args := m.Called(from, to, user)
	if args.Get(0) != nil {
		return args.Get(0).(chan *models.Heartbeat), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *HeartbeatRepositoryMock) StreamWithinByFilters(from, to time.Time, user *models.User, filters map[string][]string) (chan *models.Heartbeat, error) {
	args := m.Called(from, to, user, filters)
	if args.Get(0) != nil {
		return args.Get(0).(chan *models.Heartbeat), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *HeartbeatRepositoryMock) StreamWithinBatched(from, to time.Time, user *models.User, batchSize int) (chan []*models.Heartbeat, error) {
	args := m.Called(from, to, user, batchSize)
	if args.Get(0) != nil {
		return args.Get(0).(chan []*models.Heartbeat), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *HeartbeatRepositoryMock) Count(distinct bool) (int64, error) {
	args := m.Called(distinct)
	return args.Get(0).(int64), args.Error(1)
}

func (m *HeartbeatRepositoryMock) CountByUser(user *models.User) (int64, error) {
	args := m.Called(user)
	return args.Get(0).(int64), args.Error(1)
}

func (m *HeartbeatRepositoryMock) CountByUsers(users []*models.User) ([]*models.CountByUser, error) {
	args := m.Called(users)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.CountByUser), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *HeartbeatRepositoryMock) GetEntitySetByUser(entityType uint8, userId string) ([]string, error) {
	args := m.Called(entityType, userId)
	if args.Get(0) != nil {
		return args.Get(0).([]string), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *HeartbeatRepositoryMock) DeleteBefore(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *HeartbeatRepositoryMock) DeleteByUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *HeartbeatRepositoryMock) DeleteByUserBefore(user *models.User, t time.Time) error {
	args := m.Called(user, t)
	return args.Error(0)
}

func (m *HeartbeatRepositoryMock) GetUserProjectStats(user *models.User, from, to time.Time) ([]*models.ProjectStats, error) {
	args := m.Called(user, from, to)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.ProjectStats), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *HeartbeatRepositoryMock) GetUserAgentsByUser(user *models.User) ([]*models.UserAgent, error) {
	args := m.Called(user)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.UserAgent), args.Error(1)
	}
	return nil, args.Error(1)
}
