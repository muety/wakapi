package routes

import (
	"net/url"
	"testing"

	"github.com/muety/wakapi/config"
	"github.com/stretchr/testify/suite"
)

type SettingsHandlerTestSuite struct {
	suite.Suite
	Cfg *config.Config
	Sut *SettingsHandler
}

func TestSettingsHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(SettingsHandlerTestSuite))
}

func (suite *SettingsHandlerTestSuite) SetupSuite() {
}

func (suite *SettingsHandlerTestSuite) TearDownSuite() {
}

func (suite *SettingsHandlerTestSuite) BeforeTest(suiteName, testName string) {
	config.Set(config.Empty())
	suite.Cfg = config.Get()
	suite.Cfg.Env = "production"

	publicUrl, _ := url.Parse("https://wakapi.dev")
	suite.Cfg.Server.PublicNetUrl = publicUrl

	suite.Sut = &SettingsHandler{
		config: suite.Cfg,
	}
}

func (suite *SettingsHandlerTestSuite) TestValidateWakatimeUrl_ValidExternalUrl() {
	suite.True(suite.Sut.validateWakatimeUrl("https://wakatime.com/api/v1/"))
}

func (suite *SettingsHandlerTestSuite) TestValidateWakatimeUrl_EmptyUrlUsesDefault() {
	suite.True(suite.Sut.validateWakatimeUrl(""))
}

func (suite *SettingsHandlerTestSuite) TestValidateWakatimeUrl_InvalidUrl() {
	suite.False(suite.Sut.validateWakatimeUrl("https://192.168.0.%31/")) // an invalid url containing control character
}

func (suite *SettingsHandlerTestSuite) TestValidateWakatimeUrl_SelfReferencingHost() {
	suite.False(suite.Sut.validateWakatimeUrl("https://wakapi.dev/api"))
}

func (suite *SettingsHandlerTestSuite) TestValidateWakatimeUrl_PrivateIp() {
	suite.False(suite.Sut.validateWakatimeUrl("https://192.168.1.1/api"))
	suite.False(suite.Sut.validateWakatimeUrl("https://localhost:3000/api"))
	suite.False(suite.Sut.validateWakatimeUrl("https://127.0.0.1/api"))
	suite.False(suite.Sut.validateWakatimeUrl("https://169.254.10.20/api"))
	suite.False(suite.Sut.validateWakatimeUrl("https://0.0.0.0/api"))
}

func (suite *SettingsHandlerTestSuite) TestValidateWakatimeUrl_PrivateIp_Dev() {
	suite.Cfg.Env = "dev"
	suite.Cfg.Server.PublicNetUrl, _ = url.Parse("http://mydev.local:3000")

	suite.True(suite.Sut.validateWakatimeUrl("https://192.168.1.1/api"))
	suite.True(suite.Sut.validateWakatimeUrl("https://localhost:3000/api"))
	suite.True(suite.Sut.validateWakatimeUrl("https://127.0.0.1/api"))
	suite.True(suite.Sut.validateWakatimeUrl("https://169.254.10.20/api"))
	suite.True(suite.Sut.validateWakatimeUrl("https://0.0.0.0/api"))
	suite.False(suite.Sut.validateWakatimeUrl("https://mydev.local:3000/api")) // self-referencing is still blocked
}

func (suite *SettingsHandlerTestSuite) TestValidateWakatimeUrl_NoHttps() {
	suite.False(suite.Sut.validateWakatimeUrl("http://wakatime.com/api"))
}

func (suite *SettingsHandlerTestSuite) TestValidateWakatimeUrl_NoHttps_Dev() {
	suite.Cfg.Env = "dev"
	suite.True(suite.Sut.validateWakatimeUrl("http://wakatime.com/api"))
}

func (suite *SettingsHandlerTestSuite) TestValidateWakatimeUrl_RawIpBlockInProduction() {
	suite.False(suite.Sut.validateWakatimeUrl("https://8.8.8.8/api"))
}
