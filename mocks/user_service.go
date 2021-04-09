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

func (m *UserServiceMock) GetUserByEmail(s string) (*models.User, error) {
	args := m.Called(s)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *UserServiceMock) GetUserByResetToken(s string) (*models.User, error) {
	args := m.Called(s)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *UserServiceMock) GetAll() ([]*models.User, error) {
	args := m.Called()
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *UserServiceMock) GetActive() ([]*models.User, error) {
	args := m.Called()
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *UserServiceMock) Count() (int64, error) {
	args := m.Called()
	return int64(args.Int(0)), args.Error(1)
}

func (m *UserServiceMock) CreateOrGet(signup *models.Signup, isAdmin bool) (*models.User, bool, error) {
	args := m.Called(signup, isAdmin)
	return args.Get(0).(*models.User), args.Bool(1), args.Error(2)
}

func (m *UserServiceMock) Update(user *models.User) (*models.User, error) {
	args := m.Called(user)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *UserServiceMock) Delete(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
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

func (m *UserServiceMock) GenerateResetToken(user *models.User) (*models.User, error) {
	args := m.Called(user)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *UserServiceMock) FlushCache() {
	m.Called()
}
