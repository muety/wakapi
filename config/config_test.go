package config

import (
	"fmt"
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
	assert.Equal(t, c.Name, sqliteConnectionString(c))
}
