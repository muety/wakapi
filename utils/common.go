package utils

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/n1try/wakapi/models"
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
	re := regexp.MustCompile(`^wakatime\/[\d+.]+\s\((\w+).*\)\s.+\s(\w+)\/.+$`)
	groups := re.FindAllStringSubmatch(ua, -1)
	if len(groups) == 0 || len(groups[0]) != 3 {
		return "", "", errors.New("failed to parse user agent string")
	}
	return groups[0][1], groups[0][2], nil
}

func MakeConnectionString(config *models.Config) string {
	location, _ := time.LoadLocation("Local")
	str := strings.Builder{}
	str.WriteString(config.DbUser)
	str.WriteString(":")
	str.WriteString(config.DbPassword)
	str.WriteString("@tcp(")
	str.WriteString(config.DbHost)
	str.WriteString(":")
	str.WriteString(strconv.Itoa(int(config.DbPort)))
	str.WriteString(")/")
	str.WriteString(config.DbName)
	str.WriteString("?charset=utf8&parseTime=true&loc=")
	str.WriteString(location.String())
	return str.String()
}
