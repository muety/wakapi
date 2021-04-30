package mail

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"net/http"
	"time"
)

type MailWhaleSendingService struct {
	config     conf.MailwhaleMailConfig
	httpClient *http.Client
}

type MailWhaleSendRequest struct {
	To           []string          `json:"to"`
	Subject      string            `json:"subject"`
	Text         string            `json:"text"`
	Html         string            `json:"html"`
	TemplateId   string            `json:"template_id"`
	TemplateVars map[string]string `json:"template_vars"`
}

func NewMailWhaleSendingService(config conf.MailwhaleMailConfig) *MailWhaleSendingService {
	return &MailWhaleSendingService{
		config: config,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *MailWhaleSendingService) Send(mail *models.Mail) error {
	if len(mail.To) == 0 {
		return errors.New("not sending mail as recipient mail address seems to be invalid")
	}

	sendRequest := &MailWhaleSendRequest{
		To:      mail.To.Strings(),
		Subject: mail.Subject,
	}
	if mail.Type == models.HtmlType {
		sendRequest.Html = mail.Body
	} else {
		sendRequest.Text = mail.Body
	}
	payload, _ := json.Marshal(sendRequest)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/mail", s.config.Url), bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.SetBasicAuth(s.config.ClientId, s.config.ClientSecret)
	req.Header.Set("Content-Type", "application/json")

	res, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		return errors.New(fmt.Sprintf("got status %d from mailwhale", res.StatusCode))
	}

	return nil
}
