package utils

import (
	"fmt"
	"strings"
)

func Capitalize(s string) string {
	return fmt.Sprintf("%s%s", strings.ToUpper(s[:1]), s[1:])
}

func FindString(needle string, haystack []string, defaultVal string) string {
	for _, s := range haystack {
		if s == needle {
			return s
		}
	}
	return defaultVal
}
