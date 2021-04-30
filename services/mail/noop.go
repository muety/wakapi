package mail

import (
	"github.com/emvi/logbuch"
	"github.com/muety/wakapi/models"
)

type NoopSendingService struct{}

func (n *NoopSendingService) Send(mail *models.Mail) error {
	logbuch.Info("noop mail service doing nothing instead of sending password reset mail to [%v]", mail.To.Strings())
	return nil
}
