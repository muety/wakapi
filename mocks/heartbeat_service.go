package mocks

import (
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"github.com/stretchr/testify/mock"
	"time"
)

type HeartbeatServiceMock struct {
	mock.Mock
}

func (m *HeartbeatServiceMock) Insert(h *models.Heartbeat) error {
	args := m.Called(h)
	return args.Error(0)
}

func (m *HeartbeatServiceMock) InsertBatch(h []*models.Heartbeat) error {
	args := m.Called(h)
	return args.Error(0)
}

func (m *HeartbeatServiceMock) Count(a bool) (int64, error) {
	args := m.Called(a)
	return int64(args.Int(0)), args.Error(1)
}

func (m *HeartbeatServiceMock) CountByUser(user *models.User) (int64, error) {
	args := m.Called(user)
	return args.Get(0).(int64), args.Error(0)
}

func (m *HeartbeatServiceMock) CountByUsers(users []*models.User) ([]*models.CountByUser, error) {
	args := m.Called(users)
	return args.Get(0).([]*models.CountByUser), args.Error(0)
}

func (m *HeartbeatServiceMock) GetAllWithin(time time.Time, time2 time.Time, user *models.User) ([]*models.Heartbeat, error) {
	args := m.Called(time, time2, user)
	return args.Get(0).([]*models.Heartbeat), args.Error(1)
}

func (m *HeartbeatServiceMock) StreamAllWithin(t time.Time, t2 time.Time, u *models.User) (chan *models.Heartbeat, error) {
	args := m.Called(t, t2, u)
	return args.Get(0).(chan *models.Heartbeat), args.Error(1)
}

func (m *HeartbeatServiceMock) GetAllWithinByFilters(time time.Time, time2 time.Time, user *models.User, filters *models.Filters) ([]*models.Heartbeat, error) {
	args := m.Called(time, time2, user, filters)
	return args.Get(0).([]*models.Heartbeat), args.Error(1)
}

func (m *HeartbeatServiceMock) GetFirstByUsers() ([]*models.TimeByUser, error) {
	args := m.Called()
	return args.Get(0).([]*models.TimeByUser), args.Error(1)
}

func (m *HeartbeatServiceMock) GetLatestByUser(user *models.User) (*models.Heartbeat, error) {
	args := m.Called(user)
	return args.Get(0).(*models.Heartbeat), args.Error(1)
}

func (m *HeartbeatServiceMock) GetLatestByOriginAndUser(s string, user *models.User) (*models.Heartbeat, error) {
	args := m.Called(s, user)
	return args.Get(0).(*models.Heartbeat), args.Error(1)
}

func (m *HeartbeatServiceMock) GetLatestByFilters(u *models.User, f *models.Filters) (*models.Heartbeat, error) {
	args := m.Called(u, f)
	return args.Get(0).(*models.Heartbeat), args.Error(1)
}

func (m *HeartbeatServiceMock) GetEntitySetByUser(u uint8, user string) ([]string, error) {
	args := m.Called(u, user)
	return args.Get(0).([]string), args.Error(1)
}

func (m *HeartbeatServiceMock) DeleteBefore(time time.Time) error {
	args := m.Called(time)
	return args.Error(0)
}

func (m *HeartbeatServiceMock) DeleteByUser(u *models.User) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *HeartbeatServiceMock) DeleteByUserBefore(u *models.User, t time.Time) error {
	args := m.Called(u, t)
	return args.Error(0)
}

func (m *HeartbeatServiceMock) GetUserProjectStats(u *models.User, t, t2 time.Time, p *utils.PageParams, b bool) ([]*models.ProjectStats, error) {
	args := m.Called(u, t, t2, p, b)
	return args.Get(0).([]*models.ProjectStats), args.Error(1)
}
