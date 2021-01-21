package mocks

import (
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/mock"
)

type UserServiceMock struct {
	mock.Mock
}

func (m *UserServiceMock) GetUserById(s string) (*models.User, error) {
	args := m.Called(s)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *UserServiceMock) GetUserByKey(s string) (*models.User, error) {
	args := m.Called(s)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *UserServiceMock) GetAll() ([]*models.User, error) {
	args := m.Called()
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *UserServiceMock) CreateOrGet(signup *models.Signup) (*models.User, bool, error) {
	args := m.Called(signup)
	return args.Get(0).(*models.User), args.Bool(1), args.Error(2)
}

func (m *UserServiceMock) Update(user *models.User) (*models.User, error) {
	args := m.Called(user)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *UserServiceMock) ResetApiKey(user *models.User) (*models.User, error) {
	args := m.Called(user)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *UserServiceMock) ToggleBadges(user *models.User) (*models.User, error) {
	args := m.Called(user)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *UserServiceMock) SetWakatimeApiKey(user *models.User, s string) (*models.User, error) {
	args := m.Called(user, s)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *UserServiceMock) MigrateMd5Password(user *models.User, login *models.Login) (*models.User, error) {
	args := m.Called(user, login)
	return args.Get(0).(*models.User), args.Error(1)
}
