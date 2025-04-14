package mail

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/smtp"

	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
)

type SMTPSendingService struct {
	config conf.SMTPMailConfig
}

func NewSMTPSendingService(config conf.SMTPMailConfig) *SMTPSendingService {
	return &SMTPSendingService{
		config: config,
	}
}

func (s *SMTPSendingService) Send(mail *models.Mail) error {
	mail = mail.Sanitized()

	address := s.config.ConnStr()

	// Configure authentication if credentials are provided
	var auth smtp.Auth
	if len(s.config.Username) > 0 && len(s.config.Password) > 0 {
		auth = smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	}

	// Establish a connection with the appropriate TLS settings
	var client *smtp.Client
	var err error

	if s.config.TLS {
		// TLS connection
		conn, err := tls.Dial("tcp", address, &tls.Config{
			InsecureSkipVerify: s.config.SkipVerify,
			ServerName:         s.config.Host,
		})
		if err != nil {
			return err
		}
		client, err = smtp.NewClient(conn, s.config.Host)
		fmt.Println("err", err)
	} else {
		// Plain connection with potential STARTTLS
		client, err = smtp.Dial(address)
	}
	if err != nil {
		return err
	}
	defer client.Close()

	// StartTLS if needed
	if !s.config.TLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err = client.StartTLS(&tls.Config{
				InsecureSkipVerify: s.config.SkipVerify,
				ServerName:         s.config.Host,
			}); err != nil {
				return err
			}
		}
	}

	// Authenticate if applicable
	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return err
		}
	}

	// Set sender
	if err = client.Mail(mail.From.Raw()); err != nil {
		return err
	}

	// Set recipients
	for _, addr := range mail.To.RawStrings() {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}

	// Send mail content
	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = io.Copy(w, mail.Reader())
	if err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}

	return client.Quit()
}
