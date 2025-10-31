package config

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/oauth2-proxy/mockoidc"
	"github.com/stretchr/testify/assert"
)

// TODO: add more tests, including yaml- and env. parsing, validation, etc.

func Test_Load_OidcProviders(t *testing.T) {
	oidcMock1, _ := mockoidc.Run()
	defer oidcMock1.Shutdown()
	oidcMock2, _ := mockoidc.Run()
	defer oidcMock2.Shutdown()

	os.Setenv("WAKAPI_OIDC_PROVIDERS_0_NAME", "testprovider1")
	os.Setenv("WAKAPI_OIDC_PROVIDERS_0_CLIENT_ID", oidcMock1.ClientID)
	os.Setenv("WAKAPI_OIDC_PROVIDERS_0_CLIENT_SECRET", oidcMock1.ClientSecret)
	os.Setenv("WAKAPI_OIDC_PROVIDERS_0_ENDPOINT", oidcMock1.Addr()+"/oidc")
	os.Setenv("WAKAPI_OIDC_PROVIDERS_1_NAME", "testprovider2")
	os.Setenv("WAKAPI_OIDC_PROVIDERS_1_CLIENT_ID", oidcMock2.ClientID)
	os.Setenv("WAKAPI_OIDC_PROVIDERS_1_CLIENT_SECRET", oidcMock2.ClientSecret)
	os.Setenv("WAKAPI_OIDC_PROVIDERS_1_ENDPOINT", oidcMock2.Addr()+"/oidc")

	cfg := Load("", "")
	oidcCfg := cfg.Security.OidcProviders

	assert.Len(t, oidcCfg, 2)
	assert.Equal(t, "testprovider1", oidcCfg[0].Name)
	assert.Equal(t, oidcMock1.ClientID, oidcCfg[0].ClientID)
	assert.Equal(t, oidcMock1.ClientSecret, oidcCfg[0].ClientSecret)
	assert.Equal(t, oidcMock1.Addr()+"/oidc", oidcCfg[0].Endpoint)
	assert.Equal(t, "testprovider2", oidcCfg[1].Name)
	assert.Equal(t, oidcMock2.ClientID, oidcCfg[1].ClientID)
	assert.Equal(t, oidcMock2.ClientSecret, oidcCfg[1].ClientSecret)
	assert.Equal(t, oidcMock2.Addr()+"/oidc", oidcCfg[1].Endpoint)

	_, err1 := GetOidcProvider("testprovider1")
	_, err2 := GetOidcProvider("testprovider2")

	assert.Nil(t, err1)
	assert.Nil(t, err2)
}

func TestConfig_IsDev(t *testing.T) {
	assert.True(t, IsDev("dev"))
	assert.True(t, IsDev("development"))
	assert.False(t, IsDev("prod"))
	assert.False(t, IsDev("production"))
	assert.False(t, IsDev("anything else"))
}

func Test_mysqlConnectionString(t *testing.T) {
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

	assert.Equal(t, fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=%s&compress=true&sql_mode=ANSI_QUOTES",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
		"Local",
	), mysqlConnectionString(c))
}

func Test_mysqlConnectionStringSocket(t *testing.T) {
	c := &dbConfig{
		Socket:   "/var/run/mysql.sock",
		Port:     9999,
		User:     "test_user",
		Password: "test_password",
		Name:     "test_name",
		Dialect:  "mysql",
		Charset:  "utf8mb4",
		MaxConn:  10,
		Compress: true,
	}

	assert.Equal(t, fmt.Sprintf(
		"%s:%s@unix(%s)/%s?charset=utf8mb4&parseTime=true&loc=%s&compress=true&sql_mode=ANSI_QUOTES",
		c.User,
		c.Password,
		c.Socket,
		c.Name,
		"Local",
	), mysqlConnectionString(c))
}

func Test_postgresConnectionString(t *testing.T) {
	c := &dbConfig{
		Host:     "test_host",
		Port:     9999,
		User:     "test_user",
		Password: "test_password",
		Name:     "test_name",
		Dialect:  "postgres",
		MaxConn:  10,
	}

	assert.Equal(t, fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
		c.Host,
		c.Port,
		c.User,
		c.Name,
		c.Password,
	), postgresConnectionString(c))
}

func Test_sqliteConnectionString(t *testing.T) {
	c := &dbConfig{
		Name:    "test_name",
		Dialect: "sqlite3",
	}
	assert.True(t, strings.HasPrefix(sqliteConnectionString(c), c.Name))
	assert.Contains(t, strings.ToLower(sqliteConnectionString(c)), "journal_mode=wal")
}
