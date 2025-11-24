package mocks

import (
	"time"

	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type UserRepositoryMock struct {
	BaseRepositoryMock
	mock.Mock
}

func (m *UserRepositoryMock) FindOne(user models.User) (*models.User, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *UserRepositoryMock) GetByIds(userIds []string) ([]*models.User, error) {
	args := m.Called(userIds)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *UserRepositoryMock) GetAll() ([]*models.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *UserRepositoryMock) GetMany(ids []string) ([]*models.User, error) {
	args := m.Called(ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *UserRepositoryMock) GetAllByReports(reportsEnabled bool) ([]*models.User, error) {
	args := m.Called(reportsEnabled)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *UserRepositoryMock) GetAllByLeaderboard(leaderboardEnabled bool) ([]*models.User, error) {
	args := m.Called(leaderboardEnabled)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *UserRepositoryMock) GetByLoggedInBefore(t time.Time) ([]*models.User, error) {
	args := m.Called(t)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *UserRepositoryMock) GetByLoggedInAfter(t time.Time) ([]*models.User, error) {
	args := m.Called(t)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *UserRepositoryMock) GetByLastActiveAfter(t time.Time) ([]*models.User, error) {
	args := m.Called(t)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *UserRepositoryMock) Count() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *UserRepositoryMock) InsertOrGet(user *models.User) (*models.User, bool, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Bool(1), args.Error(2)
	}
	return args.Get(0).(*models.User), args.Bool(1), args.Error(2)
}

func (m *UserRepositoryMock) Update(user *models.User) (*models.User, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *UserRepositoryMock) UpdateField(user *models.User, key string, value interface{}) (*models.User, error) {
	args := m.Called(user, key, value)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *UserRepositoryMock) Delete(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *UserRepositoryMock) DeleteTx(user *models.User, tx *gorm.DB) error {
	args := m.Called(user, tx)
	return args.Error(0)
}
