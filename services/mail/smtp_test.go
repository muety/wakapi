package mail

/*
Uses smtp4dev (https://github.com/rnwood/smtp4dev) mock SMTP server to test against.
To spawn an smtp4dev instance in Docker, run:
$ docker run --rm -it -p 5000:80 -p 2525:25 -p 8080:80 rnwood/smtp4dev
*/

// TODO: test actual message content / title / recipients / etc.
// TODO: run in ci (gh actions)

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/emvi/logbuch"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
	"time"
)

const (
	TestSmtpUser   = "admin"
	TestSmtpPass   = "admin"
	Smtp4DevApiUrl = "http://localhost:8080/api"
	Smtp4DevHost   = "localhost"
	Smtp4DevPort   = 2525
)

type SmtpTestSuite struct {
	suite.Suite
	smtp4dev *Smtp4DevClient
}

func (suite *SmtpTestSuite) SetupSuite() {
	suite.smtp4dev = newSmtp4DevClient()
}

func (suite *SmtpTestSuite) BeforeTest(suiteName, testName string) {
	suite.smtp4dev.Setup()
}

func TestSmtpTestSuite(t *testing.T) {
	if smtp4dev := newSmtp4DevClient(); smtp4dev.Check() != nil {
		t.Skip(fmt.Sprintf("WARNING: smtp4dev not available at %s - skipping smtp tests", smtp4dev.ApiBaseUrl))
		return
	}
	suite.Run(t, new(SmtpTestSuite))
}

func (suite *SmtpTestSuite) TestSMTPSendingService_SendPlain() {
	smtp4Dev := newSmtp4DevClient()
	smtp4Dev.Setup()
	if err := smtp4Dev.SetNoTls(); err != nil {
		suite.Error(err)
	}

	cfg := createDefaultSMTPConfig()

	sut := NewSMTPSendingService(cfg)
	err := sut.Send(createTestMail())

	msgCount, _ := smtp4Dev.CountMessages()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, msgCount)
}

func (suite *SmtpTestSuite) TestSMTPSendingService_SendTLS() {
	smtp4Dev := newSmtp4DevClient()
	smtp4Dev.Setup()
	if err := smtp4Dev.SetForcedTls(); err != nil {
		suite.Error(err)
	}

	cfg := createDefaultSMTPConfig()
	cfg.TLS = true

	sut := NewSMTPSendingService(cfg)
	err := sut.Send(createTestMail())

	msgCount, _ := smtp4Dev.CountMessages()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, msgCount)
}

func (suite *SmtpTestSuite) TestSMTPSendingService_SendStartTLS() {
	smtp4Dev := newSmtp4DevClient()
	smtp4Dev.Setup()
	if err := smtp4Dev.SetStartTls(); err != nil {
		suite.Error(err)
	}

	cfg := createDefaultSMTPConfig()
	cfg.TLS = false

	sut := NewSMTPSendingService(cfg)
	err := sut.Send(createTestMail())

	msgCount, _ := smtp4Dev.CountMessages()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, msgCount)
}

// Private utility methods

func createTestMail() *models.Mail {
	mail := models.Mail{
		From:    "Wakapi <noreply@wakapi.dev>",
		To:      []models.MailAddress{"Ferdinand Mütsch <ferdinand@muetsch.io>"},
		Subject: "Wakapi Test Mail",
		Body:    "This is just a test",
		Type:    models.PlainType,
		Date:    time.Now(),
	}
	return mail.Sanitized()
}

func createDefaultSMTPConfig() config.SMTPMailConfig {
	return config.SMTPMailConfig{
		Host:       Smtp4DevHost,
		Port:       Smtp4DevPort,
		Username:   TestSmtpUser,
		Password:   TestSmtpPass,
		TLS:        false,
		SkipVerify: true,
	}
}

type Smtp4DevClient struct {
	ApiBaseUrl string
}

func newSmtp4DevClient() *Smtp4DevClient {
	return &Smtp4DevClient{
		ApiBaseUrl: Smtp4DevApiUrl,
	}
}

func (c *Smtp4DevClient) Check() error {
	res, err := http.Get(fmt.Sprintf("%s/Version", c.ApiBaseUrl))
	if err != nil {
		return err
	}
	if _, err := utils.RaiseForStatus(res, err); err != nil {
		return err
	}
	return nil
}

func (c *Smtp4DevClient) Setup() error {
	if c.Check() != nil {
		return fmt.Errorf("smtp4dev is unavailable at %s", c.ApiBaseUrl)
	}

	if err := c.CreateTestUsers(); err != nil {
		return err
	}

	if err := c.ClearInboxes(); err != nil {
		return err
	}

	return nil
}

func (c *Smtp4DevClient) GetConfig() (map[string]interface{}, error) {
	var data map[string]interface{}

	res, err := http.Get(fmt.Sprintf("%s/Server", c.ApiBaseUrl))
	if err != nil {
		return nil, err
	}
	if _, err := utils.RaiseForStatus(res, err); err != nil {
		return nil, err
	}

	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func (c *Smtp4DevClient) CountMessages() (int, error) {
	var data map[string]interface{}

	res, err := http.Get(fmt.Sprintf("%s/Messages", c.ApiBaseUrl))
	if err != nil {
		return 0, err
	}
	if _, err := utils.RaiseForStatus(res, err); err != nil {
		return 0, err
	}

	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return 0, err
	}

	return len(data["results"].([]interface{})), nil
}

func (c *Smtp4DevClient) SetNoTls() error {
	logbuch.Info("[smtp4dev] disabling tls encryption")
	err := c.SetConfigValue("tlsMode", "None")
	time.Sleep(1 * time.Second)
	return err
}

func (c *Smtp4DevClient) SetForcedTls() error {
	logbuch.Info("[smtp4dev] enabling forced tls encryption")
	err := c.SetConfigValue("tlsMode", "ImplicitTls")
	time.Sleep(1 * time.Second)
	return err
}

func (c *Smtp4DevClient) SetStartTls() error {
	logbuch.Info("[smtp4dev] enabling tls encryption via starttls")
	err := c.SetConfigValue("tlsMode", "StartTls")
	time.Sleep(1 * time.Second)
	return err
}

func (c *Smtp4DevClient) CreateTestUsers() error {
	logbuch.Info("[smtp4dev] creating test users")
	err := c.SetConfigValue("users", []map[string]interface{}{
		{
			"username":       TestSmtpUser,
			"password":       TestSmtpPass,
			"defaultMailbox": "Default",
		},
	})
	time.Sleep(100 * time.Millisecond)
	return err
}

func (c *Smtp4DevClient) ClearInboxes() error {
	logbuch.Info("[smtp4dev] clearing inboxes")
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/Messages/*", c.ApiBaseUrl), nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if _, err := utils.RaiseForStatus(res, err); err != nil {
		return err
	}

	return nil
}

func (c *Smtp4DevClient) SetConfigValue(key string, val interface{}) error {
	settings, err := c.GetConfig()
	if err != nil {
		return err
	}

	settings[key] = val

	data, _ := json.Marshal(settings)
	res, err := http.Post(fmt.Sprintf("%s/Server", c.ApiBaseUrl), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	if _, err := utils.RaiseForStatus(res, err); err != nil {
		return err
	}

	return nil
}
