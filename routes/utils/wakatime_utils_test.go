package utils

import (
	"net/url"
	"testing"

	"github.com/muety/wakapi/config"
	"github.com/stretchr/testify/suite"
)

type WakatimeUtilsTestSuite struct {
	suite.Suite
	Cfg *config.Config
}

func TestWakatimeUtils(t *testing.T) {
	suite.Run(t, new(WakatimeUtilsTestSuite))
}

func (suite *WakatimeUtilsTestSuite) SetupSuite() {
}

func (suite *WakatimeUtilsTestSuite) TearDownSuite() {
}

func (suite *WakatimeUtilsTestSuite) BeforeTest(suiteName, testName string) {
	config.Set(config.Empty())
	suite.Cfg = config.Get()
	suite.Cfg.Env = "production"

	publicUrl, _ := url.Parse("https://wakapi.dev")
	suite.Cfg.Server.PublicNetUrl = publicUrl
}

func (suite *WakatimeUtilsTestSuite) TestValidateWakatimeUrl_ValidExternalUrl() {
	suite.Nil(ValidateWakatimeUrl("https://wakatime.com/api/v1/"))
}

func (suite *WakatimeUtilsTestSuite) TestValidateWakatimeUrl_EmptyUrlUsesDefault() {
	suite.Nil(ValidateWakatimeUrl(""))
}

func (suite *WakatimeUtilsTestSuite) TestValidateWakatimeUrl_InvalidUrl() {
	suite.Error(ValidateWakatimeUrl("https://192.168.0.%31/")) // an invalid url containing control character
}

func (suite *WakatimeUtilsTestSuite) TestValidateWakatimeUrl_SelfReferencingHost() {
	suite.Error(ValidateWakatimeUrl("https://wakapi.dev/api"))
}

func (suite *WakatimeUtilsTestSuite) TestValidateWakatimeUrl_PrivateIp() {
	suite.Error(ValidateWakatimeUrl("https://192.168.1.1/api"))
	suite.Error(ValidateWakatimeUrl("https://localhost:3000/api"))
	suite.Error(ValidateWakatimeUrl("https://127.0.0.1/api"))
	suite.Error(ValidateWakatimeUrl("https://169.254.10.20/api"))
	suite.Error(ValidateWakatimeUrl("https://0.0.0.0/api"))
}

func (suite *WakatimeUtilsTestSuite) TestValidateWakatimeUrl_PrivateIp_Dev() {
	suite.Cfg.Env = "dev"
	suite.Cfg.Server.PublicNetUrl, _ = url.Parse("http://mydev.local:3000")

	suite.Nil(ValidateWakatimeUrl("https://192.168.1.1/api"))
	suite.Nil(ValidateWakatimeUrl("https://localhost:3000/api"))
	suite.Nil(ValidateWakatimeUrl("https://127.0.0.1/api"))
	suite.Nil(ValidateWakatimeUrl("https://169.254.10.20/api"))
	suite.Nil(ValidateWakatimeUrl("https://0.0.0.0/api"))
	suite.Error(ValidateWakatimeUrl("https://mydev.local:3000/api")) // self-referencing is still blocked
}

func (suite *WakatimeUtilsTestSuite) TestValidateWakatimeUrl_NoHttps() {
	suite.Error(ValidateWakatimeUrl("http://wakatime.com/api"))
}

func (suite *WakatimeUtilsTestSuite) TestValidateWakatimeUrl_NoHttps_Dev() {
	suite.Cfg.Env = "dev"
	suite.Nil(ValidateWakatimeUrl("http://wakatime.com/api"))
}

func (suite *WakatimeUtilsTestSuite) TestValidateWakatimeUrl_RawIpBlockInProduction() {
	suite.Error(ValidateWakatimeUrl("https://8.8.8.8/api"))
}
