package mocks

import (
	"time"

	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"github.com/stretchr/testify/mock"
)

type ProjectServiceMock struct {
	mock.Mock
}

func (m *ProjectServiceMock) GetUserProjectStats(u *models.User, t, t2 time.Time, s string, p *utils.PageParams, b bool) ([]*models.ProjectStats, error) {
	args := m.Called(u, t, t2, s, p, b)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.ProjectStats), args.Error(1)
	}
	return nil, args.Error(1)
}
