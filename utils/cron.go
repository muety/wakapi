package utils

import "strings"

func CronPadToSecondly(expr string) string {
	parts := strings.Split(expr, " ")
	if len(parts) == 6 {
		return expr
	}
	return "0 " + expr
}
