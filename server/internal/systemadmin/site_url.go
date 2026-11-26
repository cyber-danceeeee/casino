package systemadmin

import (
	"net"
	"strings"
)

func normalizeDomain(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.TrimPrefix(value, "http://")
	value = strings.TrimPrefix(value, "https://")
	value = strings.Split(value, "/")[0]
	value = strings.Split(value, "?")[0]
	value = strings.Split(value, "#")[0]
	if host, _, err := net.SplitHostPort(value); err == nil {
		value = host
	} else if strings.Count(value, ":") == 1 {
		value = strings.Split(value, ":")[0]
	}
	return strings.Trim(value, ". ")
}
