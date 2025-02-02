package mocks

import (
	"github.com/stretchr/testify/mock"
)

type BaseRepositoryMock struct {
	mock.Mock
}

func (m *BaseRepositoryMock) GetDialector() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *BaseRepositoryMock) GetTableDDLMysql(s string) (string, error) {
	args := m.Called(s)
	return args.Get(0).(string), args.Error(1)
}

func (m *BaseRepositoryMock) GetTableDDLSqlite(s string) (string, error) {
	args := m.Called(s)
	return args.Get(0).(string), args.Error(1)
}
