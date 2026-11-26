package softaggregator

import "encoding/json"

type apiRequest map[string]any

type apiEnvelope struct {
	Error    int             `json:"error"`
	Response json.RawMessage `json:"response"`
	Message  string          `json:"message,omitempty"`
}

type apiCallTrace struct {
	Request  map[string]any `json:"request,omitempty"`
	Response map[string]any `json:"response,omitempty"`
}

type gameListItem struct {
	ID                  int    `json:"id"`
	Name                string `json:"name"`
	Type                string `json:"type"`
	Category            string `json:"category"`
	IDHash              string `json:"id_hash"`
	FreeroundsSupported bool   `json:"freerounds_supported"`
	PlayForFunSupported bool   `json:"play_for_fun_supported"`
	Image               string `json:"image"`
}

type providerError struct {
	Code          int
	Message       string
	StatusCode    int
	RequestTrace  string
	ResponseTrace string
}

func (e *providerError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	return "softaggregator provider error"
}

func (e *providerError) ProviderRequestTrace() string {
	if e == nil {
		return ""
	}
	return e.RequestTrace
}

func (e *providerError) ProviderResponseTrace() string {
	if e == nil {
		return ""
	}
	return e.ResponseTrace
}

type callbackResponse struct {
	Error   int   `json:"error"`
	Balance int64 `json:"balance"`
}
