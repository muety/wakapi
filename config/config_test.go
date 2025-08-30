package config

import (
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	}

	assert.Equal(t, fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=%s&sql_mode=ANSI_QUOTES",
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
	}

	assert.Equal(t, fmt.Sprintf(
		"%s:%s@unix(%s)/%s?charset=utf8mb4&parseTime=true&loc=%s&sql_mode=ANSI_QUOTES",
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
	sut, err := url.QueryUnescape(sqliteConnectionString(c))
	assert.Nil(t, err)
	assert.True(t, strings.HasPrefix(sut, fmt.Sprintf("file:%s", c.Name)))
	assert.Contains(t, strings.ToLower(sut), "journal_mode=wal")
	assert.Contains(t, strings.ToLower(sut), "_timefmt=2006-01-02 15:04:05.999-07:00")
}

func Test_mssqlConnectionString(t *testing.T) {
	c := &dbConfig{
		Name:     "dbinstance",
		Host:     "test_host",
		Port:     1433,
		User:     "test_user",
		Password: "test_password",
		Dialect:  "mssql",
		Ssl:      true,
	}

	assert.Equal(t,
		"sqlserver://test_user:test_password@test_host:1433?database=dbinstance&encrypt=true",
		mssqlConnectionString(c))
}
