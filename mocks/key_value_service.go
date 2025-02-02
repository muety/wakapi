package mocks

import (
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/mock"
)

type KeyValueServiceMock struct {
	mock.Mock
}

func (m *KeyValueServiceMock) GetString(s string) (*models.KeyStringValue, error) {
	args := m.Called(s)
	return args.Get(0).(*models.KeyStringValue), args.Error(1)
}

func (m *KeyValueServiceMock) MustGetString(s string) *models.KeyStringValue {
	args := m.Called(s)
	return args.Get(0).(*models.KeyStringValue)
}

func (m *KeyValueServiceMock) GetByPrefix(s string) ([]*models.KeyStringValue, error) {
	args := m.Called(s)
	return args.Get(0).([]*models.KeyStringValue), args.Error(1)
}

func (m *KeyValueServiceMock) PutString(v *models.KeyStringValue) error {
	args := m.Called(v)
	return args.Error(0)
}

func (m *KeyValueServiceMock) DeleteString(s string) error {
	args := m.Called(s)
	return args.Error(0)
}

func (m *KeyValueServiceMock) ReplaceKeySuffix(s1, s2 string) error {
	args := m.Called(s1, s2)
	return args.Error(0)
}
