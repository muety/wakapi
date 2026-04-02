package imports

import (
	"net/url"
	"testing"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/suite"
)

type WakatimeImporterTestSuite struct {
	suite.Suite
	conf *config.Config
}

func TestWakatimeImporterTestSuite(t *testing.T) {
	suite.Run(t, new(WakatimeImporterTestSuite))
}

func (suite *WakatimeImporterTestSuite) SetupTest() {
	suite.conf = config.Empty()
	suite.conf.Server.PublicNetUrl, _ = url.Parse("https://wakapi.dev")
	config.Set(suite.conf)
}

func (suite *WakatimeImporterTestSuite) TestCheckUrl() {
	importer := NewWakatimeImporter("test-key", false)

	testCases := []struct {
		name      string
		whitelist []string
		url       string
		wantErr   bool
		errText   string
	}{
		{
			name:      "no whitelist - allowed",
			whitelist: []string{},
			url:       "https://api.wakatime.com/api/v1",
			wantErr:   false,
			errText:   "",
		},
		{
			name:      "on whitelist - allowed",
			whitelist: []string{"wakatime.com"},
			url:       "https://wakatime.com/api/v1",
			wantErr:   false,
			errText:   "",
		},
		{
			name:      "on whitelist wildcard - allowed",
			whitelist: []string{"*.wakatime.com"},
			url:       "https://api.wakatime.com/api/v1",
			wantErr:   false,
			errText:   "",
		},
		{
			name:      "not on whitelist - denied",
			whitelist: []string{"wakatime.com"},
			url:       "https://evil.com/api/v1",
			wantErr:   true,
			errText:   "not allowed",
		},
		{
			name:      "on whitelist, but private - not allowed",
			whitelist: []string{"wakatime.com"},
			url:       "https://localhost:3000/api/v1",
			wantErr:   true,
			errText:   "cannot use private ip",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.conf.App.ImportHostsWhitelist = tc.whitelist

			err := importer.Validate(&models.User{WakatimeApiUrl: tc.url})
			if tc.wantErr {
				suite.Error(err)
				suite.Contains(err.Error(), tc.errText)
			} else {
				suite.NoError(err)
			}
		})
	}
}
