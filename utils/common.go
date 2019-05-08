package utils

import (
	"errors"
	"regexp"
	"time"
)

func ParseDate(date string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", date)
}

func FormatDate(date time.Time) string {
	return date.Format("2006-01-02 15:04:05")
}

func ParseUserAgent(ua string) (string, string, error) {
	re := regexp.MustCompile(`^wakatime\/[\d+.]+\s\((\w+).*\)\s.+\s(\w+)\/.+$`)
	groups := re.FindAllStringSubmatch(ua, -1)
	if len(groups) == 0 || len(groups[0]) != 3 {
		return "", "", errors.New("Failed to parse user agent string")
	}
	return groups[0][1], groups[0][2], nil
}
