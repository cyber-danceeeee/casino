package softaggregator

import (
	"crypto/md5" //nolint:gosec // SoftAggregator callback documentation mandates MD5(timestamp + salt_key).
	"crypto/subtle"
	"encoding/hex"
	"strings"
)

func CallbackKey(timestamp string, saltKey string) string {
	sum := md5.Sum([]byte(timestamp + saltKey)) //nolint:gosec
	return hex.EncodeToString(sum[:])
}

func ValidCallbackKey(timestamp string, saltKey string, got string) bool {
	expected := strings.ToLower(CallbackKey(timestamp, saltKey))
	actual := strings.ToLower(strings.TrimSpace(got))
	if len(expected) != len(actual) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(actual)) == 1
}
