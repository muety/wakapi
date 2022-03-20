package utils

import (
	"fmt"
	"strings"
)

func Capitalize(s string) string {
	return fmt.Sprintf("%s%s", strings.ToUpper(s[:1]), s[1:])
}
