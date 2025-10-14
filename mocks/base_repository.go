package mocks

import (
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
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

func (m *BaseRepositoryMock) RunInTx(f func(db *gorm.DB) error) error {
	args := m.Called(f)
	return args.Error(0)
}

func (m *BaseRepositoryMock) VacuumOrOptimize() {
}
