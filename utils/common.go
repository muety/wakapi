package utils

import (
	"errors"
	"github.com/muety/wakapi/config"
	"regexp"
	"time"
)

func ParseDate(date string) (time.Time, error) {
	return time.Parse(config.SimpleDateFormat, date)
}

func ParseDateTime(date string) (time.Time, error) {
	return time.Parse(config.SimpleDateTimeFormat, date)
}

func FormatDate(date time.Time) string {
	return date.Format(config.SimpleDateFormat)
}

func FormatDateTime(date time.Time) string {
	return date.Format(config.SimpleDateTimeFormat)
}

func FormatDateHuman(date time.Time) string {
	return date.Format("Mon, 02 Jan 2006 15:04")
}

func Add(i, j int) int {
	return i + j
}

func ParseUserAgent(ua string) (string, string, error) {
	re := regexp.MustCompile(`(?iU)^wakatime\/[\d+.]+\s\((\w+)-.*\)\s.+\s([^\/\s]+)-wakatime\/.+$`)
	groups := re.FindAllStringSubmatch(ua, -1)
	if len(groups) == 0 || len(groups[0]) != 3 {
		return "", "", errors.New("failed to parse user agent string")
	}
	return groups[0][1], groups[0][2], nil
}
