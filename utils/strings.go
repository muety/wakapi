package utils

import (
	"strings"
)

func SplitMulti(s string, delimiters ...string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		for _, d := range delimiters {
			if string(r) == d {
				return true
			}
		}
		return false
	})
}

func FindString(needle string, haystack []string, defaultVal string) string {
	for _, s := range haystack {
		if s == needle {
			return s
		}
	}
	return defaultVal
}
