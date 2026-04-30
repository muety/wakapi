package routes

import (
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

	suite.Sut = &SettingsHandler{
		config: suite.Cfg,
	}
}

// nothing here yet
