package notificationchannel

import "strings"

var supported = map[string]struct{}{
	"IN_APP":   {},
	"FCM":      {},
	"WEB_PUSH": {},
}

func Parse(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r'
	})
	seen := make(map[string]struct{}, len(parts))
	channels := make([]string, 0, len(parts))
	for _, part := range parts {
		channel := strings.ToUpper(strings.TrimSpace(part))
		if channel == "" {
			continue
		}
		if _, exists := seen[channel]; exists {
			continue
		}
		seen[channel] = struct{}{}
		channels = append(channels, channel)
	}
	return channels
}

func Normalize(value string) string {
	return strings.Join(Parse(value), ",")
}

func Valid(value string) bool {
	channels := Parse(value)
	if len(channels) == 0 {
		return false
	}
	for _, channel := range channels {
		if _, ok := supported[channel]; !ok {
			return false
		}
	}
	return true
}

func Contains(value string, channel string) bool {
	target := strings.ToUpper(strings.TrimSpace(channel))
	for _, item := range Parse(value) {
		if item == target {
			return true
		}
	}
	return false
}
