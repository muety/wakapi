package mail

import (
	"github.com/muety/wakapi/models"
	"log/slog"
)

type NoopSendingService struct{}

func (n *NoopSendingService) Send(mail *models.Mail) error {
	slog.Info("noop mail service doing nothing instead of sending password reset mail", "to", mail.To.Strings())
	return nil
}
