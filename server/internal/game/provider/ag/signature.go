package ag

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

func sign(secretKey string, merchantCode string, timestamp string, nonce string, body []byte) string {
	payload := strings.TrimSpace(merchantCode) + timestamp + nonce + signType + string(body)
	mac := hmac.New(sha256.New, []byte(secretKey))
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func validSignature(secretKey string, merchantCode string, timestamp string, nonce string, body []byte, expected string) bool {
	actual := sign(secretKey, merchantCode, timestamp, nonce, body)
	return hmac.Equal([]byte(strings.ToLower(actual)), []byte(strings.ToLower(strings.TrimSpace(expected))))
}

func timestampMillis(now time.Time) string {
	return strconv.FormatInt(now.UnixMilli(), 10)
}

func nonce() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err == nil {
		return hex.EncodeToString(buffer)
	}
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}
