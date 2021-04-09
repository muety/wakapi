package mail

import (
	"github.com/emvi/logbuch"
	"github.com/muety/wakapi/models"
)

type NoopMailService struct{}

func (n *NoopMailService) SendPasswordResetMail(recipient *models.User, resetLink string) error {
	logbuch.Info("noop mail service doing nothing instead of sending password reset mail to %s", recipient.ID)
	return nil
}
