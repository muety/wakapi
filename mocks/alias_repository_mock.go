package mocks

import (
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/mock"
)

type AliasRepositoryMock struct {
	mock.Mock
}

func (m *AliasRepositoryMock) GetByUser(s string) ([]*models.Alias, error) {
	args := m.Called(s)
	return args.Get(0).([]*models.Alias), args.Error(1)
}
