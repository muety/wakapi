package utils

import (
	"net/http"
	"strings"
)

func IPv4HandledByDualStack(ip4, ip6 string) bool {
	return strings.HasPrefix(ip6, "[::]") && strings.HasPrefix(ip4, "0.0.0.0")
}

func IPv4HandledByDualStackHttp(server4, server6 *http.Server) bool {
	return server4 == nil || (server6 != nil && IPv4HandledByDualStack(server4.Addr, server6.Addr))
}
