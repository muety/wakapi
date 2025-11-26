package mocks

import (
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/mock"
)

type MockApiKeyService struct {
	mock.Mock
}

func (m *MockApiKeyService) GetByApiKey(apiKey string, requireFullAccessKey bool) (*models.ApiKey, error) {
	args := m.Called(apiKey, requireFullAccessKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ApiKey), args.Error(1)
}

func (m *MockApiKeyService) GetByUser(userID string) ([]*models.ApiKey, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ApiKey), args.Error(1)
}

func (m *MockApiKeyService) Create(apiKey *models.ApiKey) (*models.ApiKey, error) {
	args := m.Called(apiKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ApiKey), args.Error(1)
}

func (m *MockApiKeyService) Delete(apiKey *models.ApiKey) error {
	args := m.Called(apiKey)
	return args.Error(0)
}
