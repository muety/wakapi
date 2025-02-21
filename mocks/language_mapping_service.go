package mocks

import (
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/mock"
)

type LanguageMappingServiceMock struct {
	mock.Mock
}

func (l *LanguageMappingServiceMock) GetById(u uint) (*models.LanguageMapping, error) {
	args := l.Called(u)
	return args.Get(0).(*models.LanguageMapping), args.Error(1)
}

func (l *LanguageMappingServiceMock) GetByUser(s string) ([]*models.LanguageMapping, error) {
	args := l.Called(s)
	return args.Get(0).([]*models.LanguageMapping), args.Error(1)
}

func (l *LanguageMappingServiceMock) ResolveByUser(s string) (map[string]string, error) {
	args := l.Called(s)
	return args.Get(0).(map[string]string), args.Error(1)
}

func (l *LanguageMappingServiceMock) Create(m *models.LanguageMapping) (*models.LanguageMapping, error) {
	args := l.Called(m)
	return args.Get(0).(*models.LanguageMapping), args.Error(1)
}

func (l *LanguageMappingServiceMock) Delete(m *models.LanguageMapping) error {
	args := l.Called(m)
	return args.Error(0)
}
