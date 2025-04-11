package mail

/*
Uses smtp4Dev (https://github.com/rnwood/smtp4dev) mock SMTP server to test against.
To spawn an smtp4Dev instance in Docker, run:
$ docker run --rm -it -p 5000:80 -p 2525:25 -p 8080:80 rnwood/smtp4Dev
*/

// TODO: test actual message content / title / recipients / etc.
// TODO: run in ci (gh actions)

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"log/slog"
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
	smtp4Dev *Smtp4DevClient
}

func (suite *SmtpTestSuite) SetupSuite() {
	suite.smtp4Dev = newSmtp4DevClient()
	if err := suite.smtp4Dev.Setup(); err != nil {
		suite.Error(err)
	}
}

func (suite *SmtpTestSuite) BeforeTest(suiteName, testName string) {
	if err := suite.smtp4Dev.ClearInboxes(); err != nil {
		suite.Error(err)
	}
}

func TestSmtpTestSuite(t *testing.T) {
	address := net.JoinHostPort(Smtp4DevHost, fmt.Sprintf("%d", Smtp4DevPort))
	conn, err := net.DialTimeout("tcp", address, time.Second)
	if err != nil {
		t.Skipf("WARNING: smtp4Dev not available at %s - skipping smtp tests", address)
		return
	}
	conn.Close()

	smtp4dev := newSmtp4DevClient()
	for i := 0; i < 5; i++ {
		if smtp4dev.Check() == nil {
			break
		}
		t.Logf("smtp4Dev not ready at %s, retrying...", smtp4dev.ApiBaseUrl)
		time.Sleep(1 * time.Second)
	}

	suite.Run(t, new(SmtpTestSuite))
}

func (suite *SmtpTestSuite) TestSMTPSendingService_SendPlain() {
	if err := suite.smtp4Dev.SetNoTls(); err != nil {
		suite.Error(err)
	}

	cfg := createDefaultSMTPConfig()

	sut := NewSMTPSendingService(cfg)
	err := sut.Send(createTestMail())

	msgCount, _ := suite.smtp4Dev.CountMessages()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, msgCount)
}

func (suite *SmtpTestSuite) TestSMTPSendingService_SendTLS() {
	if err := suite.smtp4Dev.SetForcedTls(); err != nil {
		suite.Error(err)
	}

	cfg := createDefaultSMTPConfig()
	cfg.TLS = true

	sut := NewSMTPSendingService(cfg)
	err := sut.Send(createTestMail())

	msgCount, _ := suite.smtp4Dev.CountMessages()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, msgCount)
}

func (suite *SmtpTestSuite) TestSMTPSendingService_SendStartTLS() {
	if err := suite.smtp4Dev.SetStartTls(); err != nil {
		suite.Error(err)
	}

	cfg := createDefaultSMTPConfig()
	cfg.TLS = false

	sut := NewSMTPSendingService(cfg)
	err := sut.Send(createTestMail())

	msgCount, _ := suite.smtp4Dev.CountMessages()
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
		return fmt.Errorf("smtp4Dev is unavailable at %s", c.ApiBaseUrl)
	}

	if err := c.SetConfigValue("deliverMessagesToUsersDefaultMailbox", false); err != nil {
		return err
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
	slog.Info("[smtp4Dev] disabling tls encryption")
	err := c.SetConfigValue("tlsMode", "None")
	time.Sleep(2 * time.Second)
	return err
}

func (c *Smtp4DevClient) SetForcedTls() error {
	slog.Info("[smtp4Dev] enabling forced tls encryption")
	err := c.SetConfigValue("tlsMode", "ImplicitTls")
	time.Sleep(2 * time.Second)
	return err
}

func (c *Smtp4DevClient) SetStartTls() error {
	slog.Info("[smtp4Dev] enabling tls encryption via starttls")
	err := c.SetConfigValue("tlsMode", "StartTls")
	time.Sleep(2 * time.Second)
	return err
}

func (c *Smtp4DevClient) CreateTestUsers() error {
	slog.Info("[smtp4Dev] creating test users")
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
	slog.Info("[smtp4Dev] clearing inboxes")
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

	time.Sleep(5 * time.Second) // server will restart to load config changes
	return nil
}
