package mail

import (
	"github.com/emvi/logbuch"
	"github.com/muety/wakapi/models"
	"time"
)

type NoopMailService struct{}

func (n *NoopMailService) SendPasswordReset(recipient *models.User, resetLink string) error {
	logbuch.Info("noop mail service doing nothing instead of sending password reset mail to %s", recipient.ID)
	return nil
}

func (n *NoopMailService) SendImportNotification(recipient *models.User, duration time.Duration, numHeartbeats int) error {
	logbuch.Info("noop mail service doing nothing instead of sending import notification mail to %s", recipient.ID)
	return nil
}
