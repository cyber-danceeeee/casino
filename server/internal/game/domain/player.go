package domain

type CreatePlayerRequest struct {
	ProviderCode           string
	UserID                 string
	ExternalPlayerID       string
	ProviderPlayerPassword string
	Currency               string
}

type ProviderPlayer struct {
	ProviderCode     string
	UserID           string
	ProviderPlayerID string
	Currency         string
}
