package mocks

import (
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/mock"
)

type AliasServiceMock struct {
	mock.Mock
}

func (m *AliasServiceMock) IsInitialized(s string) bool {
	args := m.Called(s)
	return args.Bool(0)
}

func (m *AliasServiceMock) InitializeUser(s string) error {
	args := m.Called(s)
	return args.Error(0)
}

func (m *AliasServiceMock) GetAliasOrDefault(s string, u uint8, s2 string) (string, error) {
	args := m.Called(s, u, s2)
	return args.String(0), args.Error(1)
}

func (m *AliasServiceMock) GetByUser(s string) ([]*models.Alias, error) {
	args := m.Called(s)
	return args.Get(0).([]*models.Alias), args.Error(1)
}

func (m *AliasServiceMock) GetByUserAndKeyAndType(s string, s2 string, u uint8) ([]*models.Alias, error) {
	args := m.Called(s, s2, u)
	return args.Get(0).([]*models.Alias), args.Error(1)
}

func (m *AliasServiceMock) Create(a *models.Alias) (*models.Alias, error) {
	args := m.Called(a)
	return args.Get(0).(*models.Alias), args.Error(1)
}

func (m *AliasServiceMock) Delete(s *models.Alias) error {
	args := m.Called(s)
	return args.Error(0)
}

func (m *AliasServiceMock) DeleteMulti(a []*models.Alias) error {
	args := m.Called(a)
	return args.Error(0)
}
