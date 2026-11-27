package gameadmin

import (
	"time"

	"casino/internal/content"
)

type ListQuery struct {
	Page     int
	PageSize int
	Values   map[string]string
	Scope    content.Scope
}

type PageResult struct {
	List     any   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

type Actor struct {
	ID            uint64
	Username      string
	Scope         content.Scope
	IP            string
	UserAgent     string
	CorrelationID string
}

type ReasonCommand struct {
	Enabled bool   `json:"enabled"`
	Reason  string `json:"reason" binding:"required,max=500"`
}

type ProviderMaintenanceCommand struct {
	Maintenance        bool   `json:"maintenance"`
	MaintenanceMessage string `json:"maintenanceMessage" binding:"max=500"`
	Reason             string `json:"reason" binding:"required,max=500"`
}

type GameInput struct {
	Name                string         `json:"name" binding:"required,max=128"`
	NameI18n            map[string]any `json:"nameI18n"`
	LocalThumbnailURL   string         `json:"localThumbnailUrl" binding:"max=500"`
	BackgroundURL       string         `json:"backgroundUrl" binding:"max=500"`
	LaunchImageURL      string         `json:"launchImageUrl" binding:"max=500"`
	PlatformEnabled     bool           `json:"platformEnabled"`
	IsHot               bool           `json:"isHot"`
	IsFeatured          bool           `json:"isFeatured"`
	IsNew               bool           `json:"isNew"`
	RTP                 *float64       `json:"rtp"`
	Volatility          string         `json:"volatility" binding:"max=32"`
	SupportsDemo        *bool          `json:"supportsDemo"`
	SupportsFreeSpins   *bool          `json:"supportsFreeSpins"`
	MaintenanceMessage  string         `json:"maintenanceMessage" binding:"max=500"`
	SortOrder           int            `json:"sortOrder"`
	SupportedCountries  string         `json:"supportedCountries" binding:"max=500"`
	SupportedCurrencies string         `json:"supportedCurrencies" binding:"max=500"`
	SupportedDevices    string         `json:"supportedDevices" binding:"max=255"`
	Reason              string         `json:"reason" binding:"required,max=500"`
}

type GameBatchInput struct {
	IDs                []uint64        `json:"ids"`
	PlatformEnabled    *bool           `json:"platformEnabled"`
	IsHot              *bool           `json:"isHot"`
	IsFeatured         *bool           `json:"isFeatured"`
	IsNew              *bool           `json:"isNew"`
	Maintenance        *bool           `json:"maintenance"`
	MaintenanceMessage *string         `json:"maintenanceMessage" binding:"omitempty,max=500"`
	CategoryIDs        *[]uint64       `json:"categoryIds"`
	AddCategoryIDs     *[]uint64       `json:"addCategoryIds"`
	TagIDs             *[]uint64       `json:"tagIds"`
	SortOrders         []GameSortInput `json:"sortOrders"`
	Reason             string          `json:"reason" binding:"max=500"`
}

type GameImportInput struct {
	IDs    []uint64 `json:"ids" binding:"required,min=1,max=500"`
	Reason string   `json:"reason" binding:"max=500"`
}

type GameSortInput struct {
	ID        uint64 `json:"id" binding:"required"`
	SortOrder int    `json:"sortOrder"`
}

type RelationInput struct {
	IDs    []uint64 `json:"ids"`
	Reason string   `json:"reason" binding:"max=500"`
}

type CategoryGamesAddInput struct {
	IDs    []uint64 `json:"ids" binding:"required,min=1,max=100"`
	Reason string   `json:"reason" binding:"max=500"`
}

type CategoryGameSortInput struct {
	SortOrder int    `json:"sortOrder" binding:"min=0"`
	Reason    string `json:"reason" binding:"max=500"`
}

type ProviderCategoryMappingInput struct {
	ProviderCode string   `json:"providerCode" binding:"required,max=64"`
	Source       string   `json:"source" binding:"required,oneof=providerCategory gameType"`
	Code         string   `json:"code" binding:"required,max=128"`
	CategoryIDs  []uint64 `json:"categoryIds" binding:"required"`
	Reason       string   `json:"reason" binding:"max=500"`
}

type CategoryInput struct {
	Code      string `json:"code" binding:"required,max=32"`
	NameZh    string `json:"nameZh" binding:"required,max=64"`
	NameEn    string `json:"nameEn" binding:"required,max=64"`
	Icon      string `json:"icon" binding:"max=500"`
	BannerURL string `json:"bannerUrl" binding:"max=500"`
	SortOrder int    `json:"sortOrder"`
	Enabled   bool   `json:"enabled"`
	Reason    string `json:"reason" binding:"max=500"`
}

type TagInput struct {
	Code      string `json:"code" binding:"required,max=64"`
	NameZh    string `json:"nameZh" binding:"required,max=64"`
	NameEn    string `json:"nameEn" binding:"required,max=64"`
	Icon      string `json:"icon" binding:"max=500"`
	SortOrder int    `json:"sortOrder"`
	Enabled   bool   `json:"enabled"`
	Reason    string `json:"reason" binding:"required,max=500"`
}

type CollectionGameInput struct {
	GameID    uint64 `json:"gameId" binding:"required"`
	SortOrder int    `json:"sortOrder"`
}

type CollectionInput struct {
	Code          string                 `json:"code" binding:"required,max=64"`
	NameZh        string                 `json:"nameZh" binding:"required,max=128"`
	NameEn        string                 `json:"nameEn" binding:"required,max=128"`
	DescriptionZh string                 `json:"descriptionZh"`
	DescriptionEn string                 `json:"descriptionEn"`
	BannerURL     string                 `json:"bannerUrl" binding:"max=500"`
	SortOrder     int                    `json:"sortOrder"`
	Enabled       bool                   `json:"enabled"`
	StartTime     *time.Time             `json:"startTime"`
	EndTime       *time.Time             `json:"endTime"`
	Games         *[]CollectionGameInput `json:"games"`
	Reason        string                 `json:"reason" binding:"required,max=500"`
}

type SyncCommand struct {
	ProviderCode string `json:"providerCode"`
	Reason       string `json:"reason" binding:"required,max=500"`
}

type Error struct{ Code, Message string }

func (e *Error) Error() string                   { return e.Message }
func newError(code string, message string) error { return &Error{Code: code, Message: message} }
