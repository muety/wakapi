package mail

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/markbates/pkger"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"io/ioutil"
	"net/http"
	"text/template"
	"time"
)

const (
	tplPath              = "/views/mail"
	tplNamePasswordReset = "reset_password"
)

type MailWhaleService struct {
	config     *conf.MailwhaleMailConfig
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

func NewMailWhaleService(config *conf.MailwhaleMailConfig) *MailWhaleService {
	return &MailWhaleService{
		config: config,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (m *MailWhaleService) SendPasswordResetMail(recipient *models.User, resetLink string) error {
	tpl, err := m.loadTemplate(tplNamePasswordReset)
	if err != nil {
		return err
	}

	type data struct {
		ResetLink string
	}

	var rendered bytes.Buffer
	if err := tpl.Execute(&rendered, data{ResetLink: resetLink}); err != nil {
		return err
	}

	return m.send(recipient.Email, "Wakapi – Password Reset", rendered.String(), true)
}

func (m *MailWhaleService) send(to, subject, body string, isHtml bool) error {
	if to == "" {
		return errors.New("no recipient mail address set, cannot send password reset link")
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

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/mail", m.config.Url), bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.SetBasicAuth(m.config.ClientId, m.config.ClientSecret)
	req.Header.Set("Content-Type", "application/json")

	res, err := m.httpClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		return errors.New(fmt.Sprintf("failed to send password reset mail to %v, got status %d from mailwhale", to, res.StatusCode))
	}

	return nil
}

func (m *MailWhaleService) loadTemplate(tplName string) (*template.Template, error) {
	tplFile, err := pkger.Open(fmt.Sprintf("%s/%s.tpl.html", tplPath, tplName))
	if err != nil {
		return nil, err
	}
	defer tplFile.Close()

	tplData, err := ioutil.ReadAll(tplFile)
	if err != nil {
		return nil, err
	}

	return template.New(tplName).Parse(string(tplData))
}
