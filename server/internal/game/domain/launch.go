package domain

import "time"

type LaunchGameRequest struct {
	ProviderCode   string
	ProviderGameID string
	UserID         string
	Currency       string
	Language       string
	Country        string
	Device         string
	ReturnURL      string
	SessionID      string
	IPAddress      string

	ProviderPlayerID       string
	ProviderPlayerPassword string
}

type LaunchDemoRequest struct {
	ProviderCode   string
	ProviderGameID string
	Currency       string
	Language       string
	Country        string
	Device         string
	ReturnURL      string
	SessionID      string
	IPAddress      string
}

type LaunchGameResult struct {
	LaunchURL         string     `json:"launchUrl"`
	SessionID         string     `json:"sessionId,omitempty"`
	ProviderCode      string     `json:"providerCode"`
	ProviderPlayerID  string     `json:"providerPlayerId,omitempty"`
	ProviderSessionID string     `json:"providerSessionId,omitempty"`
	ExpiresAt         *time.Time `json:"expiresAt,omitempty"`

	ProviderRequestTrace  string `json:"-"`
	ProviderResponseTrace string `json:"-"`
}
