package mail

import (
	"errors"
	"fmt"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"io"
	"time"
)

type SMTPMailService struct {
	publicUrl string
	config    *conf.SMTPMailConfig
	auth      sasl.Client
}

func NewSMTPMailService(config *conf.SMTPMailConfig, publicUrl string) *SMTPMailService {
	return &SMTPMailService{
		publicUrl: publicUrl,
		config:    config,
		auth: sasl.NewPlainClient(
			"",
			config.Username,
			config.Password,
		),
	}
}

func (s *SMTPMailService) SendPasswordReset(recipient *models.User, resetLink string) error {
	template, err := getPasswordResetTemplate(PasswordResetTplData{ResetLink: resetLink})
	if err != nil {
		return err
	}

	mail := &models.Mail{
		From:    models.MailAddress(s.config.Sender),
		To:      models.MailAddresses([]models.MailAddress{models.MailAddress(recipient.Email)}),
		Subject: subjectPasswordReset,
	}
	mail.WithHTML(template.String())

	return s.send(s.config.ConnStr(), s.config.TLS, s.auth, mail.From.Raw(), mail.To.RawStrings(), mail.Reader())
}

func (s *SMTPMailService) SendImportNotification(recipient *models.User, duration time.Duration, numHeartbeats int) error {
	template, err := getImportNotificationTemplate(ImportNotificationTplData{
		PublicUrl:     s.publicUrl,
		Duration:      fmt.Sprintf("%.0f seconds", duration.Seconds()),
		NumHeartbeats: numHeartbeats,
	})
	if err != nil {
		return err
	}

	mail := &models.Mail{
		From:    models.MailAddress(s.config.Sender),
		To:      models.MailAddresses([]models.MailAddress{models.MailAddress(recipient.Email)}),
		Subject: subjectImportNotification,
	}
	mail.WithHTML(template.String())

	return s.send(s.config.ConnStr(), s.config.TLS, s.auth, mail.From.Raw(), mail.To.RawStrings(), mail.Reader())
}

func (s *SMTPMailService) send(addr string, tls bool, a sasl.Client, from string, to []string, r io.Reader) error {
	dial := smtp.Dial
	if tls {
		dial = func(addr string) (*smtp.Client, error) {
			return smtp.DialTLS(addr, nil)
		}
	}

	c, err := dial(addr)
	if err != nil {
		return err
	}

	defer c.Close()

	if ok, _ := c.Extension("STARTTLS"); ok {
		if err = c.StartTLS(nil); err != nil {
			return err
		}
	}
	if a != nil {
		if ok, _ := c.Extension("AUTH"); !ok {
			return errors.New("smtp: server doesn't support AUTH")
		}
		if err = c.Auth(a); err != nil {
			return err
		}
	}
	if err = c.Mail(from, nil); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = io.Copy(w, r)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}
