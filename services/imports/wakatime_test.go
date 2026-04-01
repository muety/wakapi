package imports

import (
	"testing"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/suite"
)

type WakatimeImporterTestSuite struct {
	suite.Suite
}

func TestWakatimeImporterTestSuite(t *testing.T) {
	suite.Run(t, new(WakatimeImporterTestSuite))
}

func (suite *WakatimeImporterTestSuite) TestCheckUrl() {
	config.Set(config.Empty())
	cfg := config.Get()

	importer := NewWakatimeImporter("test-key", false)

	testCases := []struct {
		name      string
		whitelist []string
		url       string
		wantErr   bool
	}{
		{
			name:      "no whitelist - allowed",
			whitelist: []string{},
			url:       "https://api.wakatime.com/api/v1",
			wantErr:   false,
		},
		{
			name:      "on whitelist - allowed",
			whitelist: []string{"wakatime.com"},
			url:       "https://wakatime.com/api/v1",
			wantErr:   false,
		},
		{
			name:      "on whitelist wildcard - allowed",
			whitelist: []string{"*.wakatime.com"},
			url:       "https://api.wakatime.com/api/v1",
			wantErr:   false,
		},
		{
			name:      "not on whitelist - denied",
			whitelist: []string{"wakatime.com"},
			url:       "https://evil.com/api/v1",
			wantErr:   true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			cfg.App.ImportHostsWhitelist = tc.whitelist
			err := importer.Validate(&models.User{WakatimeApiUrl: tc.url})
			if tc.wantErr {
				suite.Error(err)
				suite.Contains(err.Error(), "not allowed")
			} else {
				suite.NoError(err)
			}
		})
	}
}
