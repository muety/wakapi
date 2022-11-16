package mail

import (
	"errors"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"io"
)

type SMTPSendingService struct {
	config conf.SMTPMailConfig
	auth   sasl.Client
}

func NewSMTPSendingService(config conf.SMTPMailConfig) *SMTPSendingService {
	return &SMTPSendingService{
		config: config,
		auth: sasl.NewPlainClient(
			"",
			config.Username,
			config.Password,
		),
	}
}

func (s *SMTPSendingService) Send(mail *models.Mail) error {
	dial := smtp.Dial
	if s.config.TLS {
		dial = func(addr string) (*smtp.Client, error) {
			return smtp.DialTLS(addr, nil)
		}
	}

	c, err := dial(s.config.ConnStr())
	if err != nil {
		return err
	}

	defer c.Close()

	if ok, _ := c.Extension("STARTTLS"); ok {
		if err = c.StartTLS(nil); err != nil {
			errCode := err.(*smtp.SMTPError).Code
			if errCode == 503 {
				// TLS already active
			} else {
				return err
			}
		}
	}
	if s.auth != nil {
		if ok, _ := c.Extension("AUTH"); !ok {
			return errors.New("smtp: server doesn't support AUTH")
		}

		if len(s.config.Username) == 0 || len(s.config.Password) == 0 {
			return errors.New("smtp: server requires authentication, but no authentication is provided")
		}

		if err = c.Auth(s.auth); err != nil {
			return err
		}
	}
	if err = c.Mail(mail.From.Raw(), nil); err != nil {
		return err
	}
	for _, addr := range mail.To.RawStrings() {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = io.Copy(w, mail.Reader())
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}
