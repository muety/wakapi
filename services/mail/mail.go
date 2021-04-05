package mail

import (
	"github.com/emvi/logbuch"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/services"
)

func NewMailService() services.IMailService {
	config := conf.Get()
	if config.Mail.Provider == conf.MailProviderMailWhale {
		return NewMailWhaleService(config.Mail.MailWhale)
	}
	return &NoopMailService{}
}

type NoopMailService struct{}

func (n NoopMailService) SendPasswordResetMail(recipient *models.User, resetLink string) error {
	logbuch.Info("noop mail service doing nothing instead of sending password reset mail to %s", recipient.ID)
	return nil
}
