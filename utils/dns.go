package utils

import (
	"net"
	"strings"
)

// CheckEmailMX takes an e-mail address and verifies that an MX DNS record exists for its domain
func CheckEmailMX(email string) bool {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	records, err := net.LookupMX(parts[1])
	return len(records) > 0 && err == nil
}
