package mocks

import (
	"time"

	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/mock"
)

type MailServiceMock struct {
	mock.Mock
}

func (m *MailServiceMock) SendPasswordReset(user *models.User, resetLink string) error {
	args := m.Called(user, resetLink)
	return args.Error(0)
}

func (m *MailServiceMock) SendWakatimeFailureNotification(user *models.User, numFailures int) error {
	args := m.Called(user, numFailures)
	return args.Error(0)
}

func (m *MailServiceMock) SendImportNotification(user *models.User, duration time.Duration, numHeartbeats int) error {
	args := m.Called(user, duration, numHeartbeats)
	return args.Error(0)
}

func (m *MailServiceMock) SendReport(user *models.User, report *models.Report) error {
	args := m.Called(user, report)
	return args.Error(0)
}

func (m *MailServiceMock) SendSubscriptionNotification(user *models.User, hasExpired bool) error {
	args := m.Called(user, hasExpired)
	return args.Error(0)
}
