package mail

import (
	"github.com/emvi/logbuch"
	"github.com/muety/wakapi/models"
	"time"
)

type NoopMailService struct{}

const notImplemented = "noop mail service doing nothing instead of sending password reset mail to %s"

func (n *NoopMailService) SendReport(recipient *models.User, report *models.Report) error {
	logbuch.Info(notImplemented, recipient.ID)
	return nil
}

func (n *NoopMailService) SendPasswordReset(recipient *models.User, resetLink string) error {
	logbuch.Info(notImplemented, recipient.ID)
	return nil
}

func (n *NoopMailService) SendImportNotification(recipient *models.User, duration time.Duration, numHeartbeats int) error {
	logbuch.Info(notImplemented, recipient.ID)
	return nil
}
