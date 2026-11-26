package domain

import "time"

type Game struct {
	ProviderCode       string
	ProviderGameID     string
	Name               string
	Slug               string
	Category           string
	GameType           string
	ThumbnailURL       string
	LocalThumbnailURL  string
	SupportsDemo       bool
	SupportsFreeSpins  bool
	IsEnabled          bool
	RawMetadata        []byte
	LastProviderUpdate *time.Time
}

type GetGamesRequest struct {
	Currency string
	Language string
}
