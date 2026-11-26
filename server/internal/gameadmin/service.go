package gameadmin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	gameservice "casino/internal/game/service"
	"casino/internal/model"
)

type Service struct {
	db      *gorm.DB
	runtime *gameservice.Runtime
}

func New(db *gorm.DB) *Service {
	service := &Service{db: db, runtime: gameservice.NewRuntime(db)}
	_ = service.ensureRegisteredProviders(context.Background())
	return service
}

type providerRecord struct {
	ID                  uint64     `gorm:"column:id" json:"id"`
	ProviderCode        string     `gorm:"column:provider_code" json:"providerCode"`
	Name                string     `gorm:"column:name" json:"providerName"`
	LogoURL             string     `gorm:"column:logo_url" json:"logoUrl"`
	APIBaseURL          string     `gorm:"column:api_base_url" json:"apiBaseUrl"`
	Environment         string     `gorm:"column:environment" json:"environment"`
	APIStatus           string     `gorm:"column:api_status" json:"apiStatus"`
	Status              int        `gorm:"column:status" json:"status"`
	WalletMode          string     `gorm:"column:wallet_mode" json:"walletMode"`
	CallbackURL         string     `gorm:"column:callback_url" json:"callbackUrl"`
	SupportedCountries  string     `gorm:"column:supported_countries" json:"supportedCountries"`
	SupportedCurrencies string     `gorm:"column:supported_currencies" json:"supportedCurrencies"`
	MaintenanceMessage  string     `gorm:"column:maintenance_message" json:"maintenanceMessage"`
	Sort                int        `gorm:"column:sort" json:"sortOrder"`
	LastTestedAt        *time.Time `gorm:"column:last_tested_at" json:"lastTestedAt"`
	LastSyncAt          *time.Time `gorm:"column:last_sync_at" json:"lastSyncAt"`
	CreatedAt           time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt           time.Time  `gorm:"column:updated_at" json:"updatedAt"`
	GameCount           int64      `gorm:"column:game_count" json:"gameCount"`
	LatestSyncID        uint64     `gorm:"column:latest_sync_id" json:"latestSyncID"`
	LatestSyncStatus    string     `gorm:"column:latest_sync_status" json:"latestSyncStatus"`
	LatestSyncStartedAt *time.Time `gorm:"column:latest_sync_started_at" json:"latestSyncStartedAt"`
}

type gameRecord struct {
	ID                      uint64     `gorm:"column:id" json:"id"`
	GameCode                string     `gorm:"column:game_code" json:"gameCode"`
	ProviderID              uint64     `gorm:"column:provider_id" json:"providerID"`
	ProviderCode            string     `gorm:"column:provider_code" json:"providerCode"`
	ProviderGameID          string     `gorm:"column:provider_game_code" json:"providerGameId"`
	Name                    string     `gorm:"column:name" json:"name"`
	NameI18n                string     `gorm:"column:name_i18n" json:"nameI18n"`
	Slug                    string     `gorm:"column:slug" json:"slug"`
	GameType                string     `gorm:"column:game_type" json:"gameType"`
	RTP                     *float64   `gorm:"column:rtp" json:"rtp"`
	Volatility              string     `gorm:"column:volatility" json:"volatility"`
	ProviderCategoryCode    string     `gorm:"column:category" json:"providerCategoryCode"`
	ThumbnailURL            string     `gorm:"column:cover_url" json:"thumbnailUrl"`
	LocalThumbnailURL       string     `gorm:"column:local_thumbnail_url" json:"localThumbnailUrl"`
	BackgroundURL           string     `gorm:"column:background_url" json:"backgroundUrl"`
	LaunchImageURL          string     `gorm:"column:launch_image_url" json:"launchImageUrl"`
	MaintenanceMessage      string     `gorm:"column:maintenance_message" json:"maintenanceMessage"`
	SupportsDemo            bool       `gorm:"column:supports_demo" json:"supportsDemo"`
	SupportsFreeSpins       bool       `gorm:"column:supports_free_spins" json:"supportsFreeSpins"`
	ProviderEnabled         bool       `gorm:"column:provider_enabled" json:"providerEnabled"`
	PlatformEnabled         bool       `gorm:"column:platform_enabled" json:"platformEnabled"`
	PlatformEnabledDraft    *bool      `gorm:"column:platform_enabled_draft" json:"platformEnabledDraft"`
	IsHot                   bool       `gorm:"column:is_hot" json:"isHot"`
	IsHotDraft              *bool      `gorm:"column:is_hot_draft" json:"isHotDraft"`
	IsFeatured              bool       `gorm:"column:is_featured" json:"isFeatured"`
	IsFeaturedDraft         *bool      `gorm:"column:is_featured_draft" json:"isFeaturedDraft"`
	IsNew                   bool       `gorm:"column:is_new" json:"isNew"`
	IsNewDraft              *bool      `gorm:"column:is_new_draft" json:"isNewDraft"`
	SortOrder               int        `gorm:"column:sort" json:"sortOrder"`
	SortDraft               *int       `gorm:"column:sort_draft" json:"sortDraft"`
	SupportedCountries      string     `gorm:"column:supported_countries" json:"supportedCountries"`
	SupportedCurrencies     string     `gorm:"column:supported_currencies" json:"supportedCurrencies"`
	SupportedDevices        string     `gorm:"column:supported_devices" json:"supportedDevices"`
	RawMetadata             string     `gorm:"column:raw_metadata" json:"-"`
	LastSyncedAt            *time.Time `gorm:"column:last_synced_at" json:"lastSyncedAt"`
	CreatedAt               time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt               time.Time  `gorm:"column:updated_at" json:"updatedAt"`
	PlatformCategoryCodes   string     `gorm:"column:platform_category_codes" json:"platformCategories"`
	TagCodes                string     `gorm:"column:tag_codes" json:"tags"`
	CategorySortOrder       int        `gorm:"column:category_sort_order" json:"categorySortOrder"`
	Status                  int        `gorm:"column:status" json:"status"`
	StatusDraft             *int       `gorm:"column:status_draft" json:"statusDraft"`
	MaintenanceMessageDraft *string    `gorm:"column:maintenance_message_draft" json:"maintenanceMessageDraft"`
	CategoryIDsDraft        string     `gorm:"column:category_ids_draft" json:"categoryIDsDraft"`
	ConfigStatus            string     `gorm:"column:config_status" json:"configStatus"`
	ConfigVersion           uint64     `gorm:"column:config_version" json:"configVersion"`
	ConfigPublishedAt       *time.Time `gorm:"column:config_published_at" json:"configPublishedAt"`
	ConfigPublishedBy       uint64     `gorm:"column:config_published_by" json:"configPublishedBy"`
	DraftUpdatedAt          *time.Time `gorm:"column:draft_updated_at" json:"draftUpdatedAt"`
	DraftUpdatedBy          uint64     `gorm:"column:draft_updated_by" json:"draftUpdatedBy"`
}

type providerGameTypeOption struct {
	ProviderCode string `gorm:"column:provider_code" json:"providerCode"`
	Source       string `gorm:"column:source" json:"source"`
	Code         string `gorm:"column:code" json:"code"`
	GameCount    int64  `gorm:"column:game_count" json:"gameCount"`
}

type categoryRecord struct {
	TenantID   uint64    `gorm:"column:tenant_id" json:"tenantID"`
	SiteID     uint64    `gorm:"column:site_id" json:"siteID"`
	ID         uint64    `json:"id"`
	Code       string    `gorm:"column:category_key" json:"code"`
	Label      string    `json:"-"`
	ShortLabel string    `gorm:"column:short_label" json:"-"`
	NameZh     string    `gorm:"column:name_zh" json:"nameZh"`
	NameEn     string    `gorm:"column:name_en" json:"nameEn"`
	Icon       string    `gorm:"column:icon_url" json:"icon"`
	BannerURL  string    `gorm:"column:banner_url" json:"bannerUrl"`
	SortOrder  int       `gorm:"column:sort" json:"sortOrder"`
	Status     int       `json:"-"`
	Enabled    bool      `gorm:"-" json:"enabled"`
	GameCount  int64     `gorm:"->;column:game_count" json:"gameCount"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (categoryRecord) TableName() string { return "game_game_categories" }

type tagRecord struct {
	TenantID  uint64    `gorm:"column:tenant_id" json:"tenantID"`
	SiteID    uint64    `gorm:"column:site_id" json:"siteID"`
	ID        uint64    `json:"id"`
	Code      string    `json:"code"`
	NameZh    string    `json:"nameZh"`
	NameEn    string    `json:"nameEn"`
	Icon      string    `gorm:"column:icon_url" json:"icon"`
	SortOrder int       `json:"sortOrder"`
	Enabled   bool      `json:"enabled"`
	GameCount int64     `gorm:"->;column:game_count" json:"gameCount"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (tagRecord) TableName() string { return "game_game_tags" }

type collectionRecord struct {
	TenantID      uint64     `gorm:"column:tenant_id" json:"tenantID"`
	SiteID        uint64     `gorm:"column:site_id" json:"siteID"`
	ID            uint64     `json:"id"`
	Code          string     `json:"code"`
	NameZh        string     `json:"nameZh"`
	NameEn        string     `json:"nameEn"`
	DescriptionZh string     `json:"descriptionZh"`
	DescriptionEn string     `json:"descriptionEn"`
	BannerURL     string     `json:"bannerUrl"`
	SortOrder     int        `json:"sortOrder"`
	Enabled       bool       `json:"enabled"`
	StartAt       *time.Time `json:"startTime"`
	EndAt         *time.Time `json:"endTime"`
	GameCount     int64      `gorm:"->;column:game_count" json:"gameCount"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

func (collectionRecord) TableName() string { return "game_game_collections" }

func normalizePage(q ListQuery) (int, int, int) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = 20
	}
	if q.PageSize > 100 {
		q.PageSize = 100
	}
	return q.Page, q.PageSize, (q.Page - 1) * q.PageSize
}

func paginate(db *gorm.DB, q ListQuery, out any) (PageResult, error) {
	p, ps, offset := normalizePage(q)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return PageResult{}, err
	}
	if err := db.Offset(offset).Limit(ps).Find(out).Error; err != nil {
		return PageResult{}, err
	}
	return PageResult{List: out, Total: total, Page: p, PageSize: ps}, nil
}

func contains(db *gorm.DB, column, value string) *gorm.DB {
	if value = strings.TrimSpace(value); value != "" {
		return db.Where(column+" LIKE ?", "%"+value+"%")
	}
	return db
}
func equal(db *gorm.DB, column, value string) *gorm.DB {
	if value = strings.TrimSpace(value); value != "" {
		return db.Where(column+" = ?", value)
	}
	return db
}
func boolFilter(db *gorm.DB, column, value string) *gorm.DB {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1":
		return db.Where(column + " = 1")
	case "false", "0":
		return db.Where(column + " = 0")
	default:
		return db
	}
}
func timeRange(db *gorm.DB, column, from, to string) *gorm.DB {
	if from != "" {
		db = db.Where(column+" >= ?", from)
	}
	if to != "" {
		db = db.Where(column+" <= ?", to)
	}
	return db
}
func parseID(value string) uint64 {
	id, _ := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	return id
}

var codePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,63}$`)

func normalizeCode(code string) string {
	code = strings.ToLower(strings.TrimSpace(code))
	code = strings.ReplaceAll(code, "_", "-")
	code = strings.Join(strings.Fields(code), "-")
	for strings.Contains(code, "--") {
		code = strings.ReplaceAll(code, "--", "-")
	}
	return strings.Trim(code, "-")
}

func validateCode(code string) error {
	if !codePattern.MatchString(code) {
		return newError("INVALID_CODE", "invalid code")
	}
	return nil
}
func validateTimeRange(start, end *time.Time) error {
	if start != nil && end != nil && end.Before(*start) {
		return newError("INVALID_TIME_RANGE", "end time must be after start time")
	}
	return nil
}
func notFound(err error, code string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return newError(code, "record not found")
	}
	return err
}

func duplicateCode(err error, code, message string) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return newError(code, message)
	}
	return err
}
func audit(tx *gorm.DB, actor Actor, permission, action, target string, targetID uint64, before, after any, reason string) error {
	b, _ := json.Marshal(before)
	a, _ := json.Marshal(after)
	return tx.Create(&model.GameAdminOperationLogs{TenantID: actor.Scope.TenantID, SiteID: actor.Scope.SiteID, OperatorID: actor.ID, OperatorUsername: actor.Username, Permission: permission, Action: action, TargetType: target, TargetID: targetID, BeforeState: string(b), AfterState: string(a), Reason: reason, IPAddress: actor.IP, UserAgent: actor.UserAgent, CorrelationID: actor.CorrelationID}).Error
}
func (s *Service) RecordAssetUpload(kind, assetURL string, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return audit(tx, actor, "game:asset:upload", "upload", "game_asset", 0, nil, map[string]any{"kind": kind, "url": assetURL}, "")
	})
}
func (s *Service) operationLogs(target string, id uint64) []model.GameAdminOperationLogs {
	var rows []model.GameAdminOperationLogs
	_ = s.db.Where("target_type=? AND target_id=?", target, id).Order("id DESC").Limit(100).Find(&rows).Error
	return rows
}
func generatedCode(prefix string) string { return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano()) }

func redactJSON(raw string) any {
	var value any
	if json.Unmarshal([]byte(raw), &value) != nil {
		return map[string]any{}
	}
	redactValue(value)
	return value
}
func redactValue(value any) {
	switch typed := value.(type) {
	case map[string]any:
		for key, item := range typed {
			lower := strings.ToLower(key)
			if strings.Contains(lower, "password") || strings.Contains(lower, "secret") || strings.Contains(lower, "salt") || strings.Contains(lower, "token") || strings.Contains(lower, "signature") || strings.Contains(lower, "url") && strings.Contains(lower, "launch") {
				typed[key] = "***"
				continue
			}
			redactValue(item)
		}
	case []any:
		for _, item := range typed {
			redactValue(item)
		}
	}
}
