package config

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/oauth2-proxy/mockoidc"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
	Cfg *Config
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (suite *ConfigTestSuite) SetupSuite() {}

func (suite *ConfigTestSuite) TearDownSuite() {}

func (suite *ConfigTestSuite) BeforeTest(suiteName, testName string) {
	Set(Empty())
	suite.Cfg = Get()
	suite.Cfg.Env = "production"
}

func (suite *ConfigTestSuite) AfterTest(suiteName, testName string) {
	for _, env := range os.Environ() {
		split := strings.Split(env, "=")
		key := split[0]
		if strings.HasPrefix(key, "WAKAPI_") {
			os.Setenv(strings.Split(env, "=")[0], "")
		}
	}
}

func (suite *ConfigTestSuite) TestLoadOidcProviders() {
	oidcMock1, _ := mockoidc.Run()
	defer oidcMock1.Shutdown()
	oidcMock2, _ := mockoidc.Run()
	defer oidcMock2.Shutdown()

	suite.T().Setenv("WAKAPI_OIDC_PROVIDERS_0_NAME", "testprovider1")
	suite.T().Setenv("WAKAPI_OIDC_PROVIDERS_0_DISPLAY_NAME", "Test Provider 1")
	suite.T().Setenv("WAKAPI_OIDC_PROVIDERS_0_CLIENT_ID", oidcMock1.ClientID)
	suite.T().Setenv("WAKAPI_OIDC_PROVIDERS_0_CLIENT_SECRET", oidcMock1.ClientSecret)
	suite.T().Setenv("WAKAPI_OIDC_PROVIDERS_0_ENDPOINT", oidcMock1.Addr()+"/oidc")
	suite.T().Setenv("WAKAPI_OIDC_PROVIDERS_1_NAME", "testprovider2")
	suite.T().Setenv("WAKAPI_OIDC_PROVIDERS_1_CLIENT_ID", oidcMock2.ClientID)
	suite.T().Setenv("WAKAPI_OIDC_PROVIDERS_1_CLIENT_SECRET", oidcMock2.ClientSecret)
	suite.T().Setenv("WAKAPI_OIDC_PROVIDERS_1_ENDPOINT", oidcMock2.Addr()+"/oidc")

	cfg := Load("", "")
	oidcCfg := cfg.Security.OidcProviders

	suite.Len(oidcCfg, 2)
	suite.Equal("testprovider1", oidcCfg[0].Name)
	suite.Equal("Test Provider 1", oidcCfg[0].DisplayName)
	suite.Equal("Test Provider 1", oidcCfg[0].String())
	suite.Equal(oidcMock1.ClientID, oidcCfg[0].ClientID)
	suite.Equal(oidcMock1.ClientSecret, oidcCfg[0].ClientSecret)
	suite.Equal(oidcMock1.Addr()+"/oidc", oidcCfg[0].Endpoint)
	suite.Equal("testprovider2", oidcCfg[1].Name)
	suite.Equal("", oidcCfg[1].DisplayName)
	suite.Equal("Testprovider2", oidcCfg[1].String())
	suite.Equal(oidcMock2.ClientID, oidcCfg[1].ClientID)
	suite.Equal(oidcMock2.ClientSecret, oidcCfg[1].ClientSecret)
	suite.Equal(oidcMock2.Addr()+"/oidc", oidcCfg[1].Endpoint)

	p1, err1 := GetOidcProvider("testprovider1")
	suite.NoError(err1)
	suite.Equal("Test Provider 1", p1.DisplayName)

	p2, err2 := GetOidcProvider("testprovider2")
	suite.NoError(err2)
	suite.Equal("Testprovider2", p2.DisplayName)
}

func (suite *ConfigTestSuite) TestOidcProviderConfigValidate() {
	testCases := []struct {
		name   string
		config oidcProviderConfig
		err    string
	}{
		{
			name: "valid",
			config: oidcProviderConfig{
				Name:         "test-provider-1",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				Endpoint:     "https://provider.com/oidc",
			},
			err: "",
		},
		{
			name: "valid with http",
			config: oidcProviderConfig{
				Name:         "test-provider-1",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				Endpoint:     "http://provider.com/oidc",
			},
			err: "",
		},
		{
			name: "invalid name with spaces",
			config: oidcProviderConfig{
				Name: "test provider",
			},
			err: "invalid provider name 'test provider', must only contain alphanumeric characters or '-'",
		},
		{
			name: "invalid name with underscore",
			config: oidcProviderConfig{
				Name: "test_provider",
			},
			err: "invalid provider name 'test_provider', must only contain alphanumeric characters or '-'",
		},
		{
			name: "missing client id",
			config: oidcProviderConfig{
				Name:         "test-provider",
				ClientSecret: "client-secret",
				Endpoint:     "https://provider.com/oidc",
			},
			err: "provider 'test-provider' is missing client id",
		},
		{
			name: "missing client secret",
			config: oidcProviderConfig{
				Name:     "test-provider",
				ClientID: "client-id",
				Endpoint: "https://provider.com/oidc",
			},
			err: "provider 'test-provider' is missing client secret",
		},
		{
			name: "missing endpoint",
			config: oidcProviderConfig{
				Name:         "test-provider",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
			},
			err: "provider 'test-provider' is missing endpoint",
		},
		{
			name: "invalid endpoint scheme",
			config: oidcProviderConfig{
				Name:         "test-provider",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				Endpoint:     "ftp://provider.com/oidc",
			},
			err: "provider 'test-provider' is missing endpoint",
		},
		{
			name: "endpoint without scheme",
			config: oidcProviderConfig{
				Name:         "test-provider",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				Endpoint:     "provider.com/oidc",
			},
			err: "provider 'test-provider' is missing endpoint",
		},
	}
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if tc.err == "" {
				suite.NoError(err)
			} else {
				suite.EqualError(err, tc.err)
			}
		})
	}
}

func (suite *ConfigTestSuite) TestIsDev() {
	suite.True(IsDev("dev"))
	suite.True(IsDev("development"))
	suite.False(IsDev("prod"))
	suite.False(IsDev("production"))
	suite.False(IsDev("anything else"))
}

func (suite *ConfigTestSuite) TestMysqlConnectionString() {
	c := &dbConfig{
		Host:     "test_host",
		Port:     9999,
		User:     "test_user",
		Password: "test_password",
		Name:     "test_name",
		Dialect:  "mysql",
		Charset:  "utf8mb4",
		MaxConn:  10,
		Compress: true,
	}
	suite.Equal(fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=%s&compress=true&sql_mode=ANSI_QUOTES",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
		"Local",
	), mysqlConnectionString(c))
}

func (suite *ConfigTestSuite) TestMysqlConnectionStringSocket() {
	c := &dbConfig{
		Socket:   "/var/run/mysql.sock",
		User:     "test_user",
		Password: "test_password",
		Name:     "test_name",
		Dialect:  "mysql",
		Charset:  "utf8mb4",
		MaxConn:  10,
		Compress: true,
	}
	suite.Equal(fmt.Sprintf(
		"%s:%s@unix(%s)/%s?charset=utf8mb4&parseTime=true&loc=%s&compress=true&sql_mode=ANSI_QUOTES",
		c.User,
		c.Password,
		c.Socket,
		c.Name,
		"Local",
	), mysqlConnectionString(c))
}

func (suite *ConfigTestSuite) TestPostgresConnectionString() {
	c := &dbConfig{
		Host:     "test_host",
		Port:     9999,
		User:     "test_user",
		Password: "test_password",
		Name:     "test_name",
		Dialect:  "postgres",
		MaxConn:  10,
	}
	suite.Equal(fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
		c.Host,
		c.Port,
		c.User,
		c.Name,
		c.Password,
	), postgresConnectionString(c))
}

func (suite *ConfigTestSuite) TestSqliteConnectionString() {
	c := &dbConfig{
		Name:    "test_name",
		Dialect: "sqlite3",
	}
	suite.True(strings.HasPrefix(sqliteConnectionString(c), c.Name))
	suite.Contains(strings.ToLower(sqliteConnectionString(c)), "journal_mode=wal")
}

func (suite *ConfigTestSuite) TestPublicNetUrl() {
	suite.T().Setenv("WAKAPI_PUBLIC_URL", "https://wakapi.dev")
	cfg := Load("", "")
	suite.NotNil(cfg.Server.PublicNetUrl)
	suite.Equal("wakapi.dev", cfg.Server.PublicNetUrl.Hostname())
	suite.Equal("https", cfg.Server.PublicNetUrl.Scheme)
}

func (suite *ConfigTestSuite) TestIsImportHostWhitelisted() {
	testCases := []struct {
		name      string
		whitelist []string
		host      string
		expected  bool
	}{
		{
			name:      "empty whitelist",
			whitelist: []string{},
			host:      "google.com",
			expected:  true,
		},
		{
			name:      "exact match",
			whitelist: []string{"google.com"},
			host:      "google.com",
			expected:  true,
		},
		{
			name:      "no match",
			whitelist: []string{"google.com"},
			host:      "bing.com",
			expected:  false,
		},
		{
			name:      "wildcard prefix",
			whitelist: []string{"*.google.com"},
			host:      "api.google.com",
			expected:  true,
		},
		{
			name:      "wildcard suffix",
			whitelist: []string{"google.*"},
			host:      "google.de",
			expected:  true,
		},
		{
			name:      "wildcard both sides",
			whitelist: []string{"*google*"},
			host:      "my-google-app.com",
			expected:  true,
		},
		{
			name:      "multiple entries",
			whitelist: []string{"bing.com", "*.google.com"},
			host:      "api.google.com",
			expected:  true,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			if len(tc.whitelist) > 0 {
				suite.T().Setenv("WAKAPI_IMPORT_HOSTS_WHITELIST", strings.Join(tc.whitelist, ","))
			} else {
				os.Unsetenv("WAKAPI_IMPORT_HOSTS_WHITELIST")
			}

			cfg := Load("", "")
			suite.Equal(tc.expected, cfg.App.IsImportHostWhitelisted(tc.host))
		})
	}
}
