package utils

import (
	"errors"
	"fmt"
	"github.com/muety/wakapi/config"
	"regexp"
	"time"
)

func ParseDate(date string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", date)
}

func FormatDate(date time.Time) string {
	return date.Format("2006-01-02 15:04:05")
}

func FormatDateHuman(date time.Time) string {
	return date.Format("Mon, 02 Jan 2006 15:04")
}

func ParseUserAgent(ua string) (string, string, error) {
	re := regexp.MustCompile(`(?iU)^wakatime\/[\d+.]+\s\((\w+)-.*\)\s.+\s([^\/\s]+)-wakatime\/.+$`)
	groups := re.FindAllStringSubmatch(ua, -1)
	if len(groups) == 0 || len(groups[0]) != 3 {
		return "", "", errors.New("failed to parse user agent string")
	}
	return groups[0][1], groups[0][2], nil
}

func MakeConnectionString(config *config.Config) string {
	switch config.Db.Dialect {
	case "mysql":
		return mysqlConnectionString(config)
	case "postgres":
		return postgresConnectionString(config)
	case "sqlite3":
		return sqliteConnectionString(config)
	}
	return ""
}

func mysqlConnectionString(config *config.Config) string {
	//location, _ := time.LoadLocation("Local")
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=true&loc=%s&sql_mode=ANSI_QUOTES",
		config.Db.User,
		config.Db.Password,
		config.Db.Host,
		config.Db.Port,
		config.Db.Name,
		"Local",
	)
}

func postgresConnectionString(config *config.Config) string {
	return fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
		config.Db.Host,
		config.Db.Port,
		config.Db.User,
		config.Db.Name,
		config.Db.Password,
	)
}

func sqliteConnectionString(config *config.Config) string {
	return config.Db.Name
}
