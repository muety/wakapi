package mocks

import (
	"github.com/stretchr/testify/mock"
)

type AliasServiceMock struct {
	mock.Mock
}

func (m *AliasServiceMock) LoadUserAliases(s string) error {
	args := m.Called(s)
	return args.Error(0)
}

func (m *AliasServiceMock) GetAliasOrDefault(s string, u uint8, s2 string) (string, error) {
	args := m.Called(s, u, s2)
	return args.String(0), args.Error(1)
}

func (m *AliasServiceMock) IsInitialized(s string) bool {
	args := m.Called(s)
	return args.Bool(0)
}
