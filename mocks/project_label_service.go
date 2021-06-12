package mocks

import (
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/mock"
)

type ProjectLabelServiceMock struct {
	mock.Mock
}

func (p *ProjectLabelServiceMock) GetById(u uint) (*models.ProjectLabel, error) {
	args := p.Called(u)
	return args.Get(0).(*models.ProjectLabel), args.Error(1)
}

func (p *ProjectLabelServiceMock) GetByUser(s string) ([]*models.ProjectLabel, error) {
	args := p.Called(s)
	return args.Get(0).([]*models.ProjectLabel), args.Error(1)
}

func (p *ProjectLabelServiceMock) GetByUserGrouped(s string) (map[string][]*models.ProjectLabel, error) {
	args := p.Called(s)
	return args.Get(0).(map[string][]*models.ProjectLabel), args.Error(1)
}

func (p *ProjectLabelServiceMock) Create(l *models.ProjectLabel) (*models.ProjectLabel, error) {
	args := p.Called(l)
	return args.Get(0).(*models.ProjectLabel), args.Error(1)
}

func (p *ProjectLabelServiceMock) Delete(l *models.ProjectLabel) error {
	args := p.Called(l)
	return args.Error(0)
}
