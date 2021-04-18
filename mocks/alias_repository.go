package mocks

import (
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/mock"
)

type AliasRepositoryMock struct {
	mock.Mock
}

func (m *AliasRepositoryMock) GetAll() ([]*models.Alias, error) {
	args := m.Called()
	return args.Get(0).([]*models.Alias), args.Error(1)
}

func (m *AliasRepositoryMock) GetByUser(s string) ([]*models.Alias, error) {
	args := m.Called(s)
	return args.Get(0).([]*models.Alias), args.Error(1)
}

func (m *AliasRepositoryMock) GetByUserAndKey(s string, s2 string) ([]*models.Alias, error) {
	args := m.Called(s, s2)
	return args.Get(0).([]*models.Alias), args.Error(1)
}

func (m *AliasRepositoryMock) GetByUserAndKeyAndType(s string, s2 string, u uint8) ([]*models.Alias, error) {
	args := m.Called(s, s2, u)
	return args.Get(0).([]*models.Alias), args.Error(1)
}

func (m *AliasRepositoryMock) GetByUserAndTypeAndValue(s string, u uint8, s2 string) (*models.Alias, error) {
	args := m.Called(s, u, s2)
	return args.Get(0).(*models.Alias), args.Error(1)
}

func (m *AliasRepositoryMock) Insert(s *models.Alias) (*models.Alias, error) {
	args := m.Called(s)
	return args.Get(0).(*models.Alias), args.Error(1)
}

func (m *AliasRepositoryMock) Delete(u uint) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *AliasRepositoryMock) DeleteBatch(u []uint) error {
	args := m.Called(u)
	return args.Error(0)
}
