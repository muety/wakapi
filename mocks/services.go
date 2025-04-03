package mocks

// WORK IN PROGRESS

import (
	"github.com/stretchr/testify/mock"
)

type ServicesMock struct {
	mock.Mock
}

func (m *ServicesMock) Alias() *AliasServiceMock {
	return new(AliasServiceMock)
}

func (m *ServicesMock) User() *UserServiceMock {
	return new(UserServiceMock)
}

func (m *ServicesMock) ProjectLabel() *ProjectLabelServiceMock {
	return new(ProjectLabelServiceMock)
}

func (m *ServicesMock) Duration() *DurationServiceMock {
	return new(DurationServiceMock)
}

func (m *ServicesMock) Summary() *SummaryServiceMock {
	return new(SummaryServiceMock)
}

func (s *ServicesMock) KeyValue() *KeyValueServiceMock {
	return new(KeyValueServiceMock)
}
