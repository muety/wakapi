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

type MailWhaleMailService struct {
	publicUrl  string
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

func NewMailWhaleService(config conf.MailwhaleMailConfig, publicUrl string) *MailWhaleMailService {
	return &MailWhaleMailService{
		publicUrl: publicUrl,
		config:    config,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *MailWhaleMailService) SendPasswordReset(recipient *models.User, resetLink string) error {
	template, err := getPasswordResetTemplate(PasswordResetTplData{ResetLink: resetLink})
	if err != nil {
		return err
	}
	return s.send(recipient.Email, subjectPasswordReset, template.String(), true)
}

func (s *MailWhaleMailService) SendImportNotification(recipient *models.User, duration time.Duration, numHeartbeats int) error {
	template, err := getImportNotificationTemplate(ImportNotificationTplData{
		PublicUrl:     s.publicUrl,
		Duration:      fmt.Sprintf("%.0f seconds", duration.Seconds()),
		NumHeartbeats: numHeartbeats,
	})
	if err != nil {
		return err
	}
	return s.send(recipient.Email, subjectImportNotification, template.String(), true)
}

func (s *MailWhaleMailService) SendReport(recipient *models.User, report *models.Report) error {
	template, err := getReportTemplate(ReportTplData{report})
	if err != nil {
		return err
	}
	return s.send(recipient.Email, subjectReport, template.String(), true)
}

func (s *MailWhaleMailService) send(to, subject, body string, isHtml bool) error {
	if to == "" {
		return errors.New("not sending mail as recipient mail address seems to be invalid")
	}

	sendRequest := &MailWhaleSendRequest{
		To:      []string{to},
		Subject: subject,
	}
	if isHtml {
		sendRequest.Html = body
	} else {
		sendRequest.Text = body
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
