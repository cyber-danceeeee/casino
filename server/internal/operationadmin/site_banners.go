package operationadmin

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/operationdefaults"
	"casino/internal/storage"
	"casino/internal/tenancy"
)

const (
	BannerPublishDraft     = "DRAFT"
	BannerPublishPublished = "PUBLISHED"
)

var allowedBannerActions = map[string]bool{
	"NONE": true, "OPEN_INTERNAL_PAGE": true, "OPEN_EXTERNAL_URL": true,
	"OPEN_GAME": true, "OPEN_PROMOTION": true, "OPEN_DEPOSIT": true, "OPEN_SUPPORT": true,
}

type SiteBannerDraft struct {
	Name            string     `json:"name"`
	Title           string     `json:"title"`
	Subtitle        string     `json:"subtitle"`
	DesktopAssetID  uint64     `json:"desktopAssetID"`
	MobileAssetID   uint64     `json:"mobileAssetID"`
	DesktopImageURL string     `json:"desktopImageUrl,omitempty"`
	MobileImageURL  string     `json:"mobileImageUrl,omitempty"`
	ActionType      string     `json:"actionType"`
	ActionTarget    string     `json:"actionTarget"`
	SortOrder       int        `json:"sortOrder"`
	LanguageCode    string     `json:"languageCode"`
	CountryCode     string     `json:"countryCode"`
	CurrencyCode    string     `json:"currencyCode"`
	VisibilityScope string     `json:"visibilityScope"`
	StartAt         *time.Time `json:"startAt"`
	EndAt           *time.Time `json:"endAt"`
	Enabled         bool       `json:"enabled"`
	Deleted         bool       `json:"deleted,omitempty"`
}

type SiteBannerInput struct {
	Name            string     `json:"name" binding:"required,max=128"`
	Title           string     `json:"title" binding:"required,max=256"`
	Subtitle        string     `json:"subtitle" binding:"max=500"`
	DesktopAssetID  uint64     `json:"desktopAssetID"`
	MobileAssetID   uint64     `json:"mobileAssetID"`
	DesktopImageURL string     `json:"desktopImageUrl" binding:"max=500"`
	MobileImageURL  string     `json:"mobileImageUrl" binding:"max=500"`
	ActionType      string     `json:"actionType" binding:"required"`
	ActionTarget    string     `json:"actionTarget" binding:"max=500"`
	SortOrder       int        `json:"sortOrder"`
	LanguageCode    string     `json:"languageCode" binding:"max=16"`
	CountryCode     string     `json:"countryCode" binding:"max=16"`
	CurrencyCode    string     `json:"currencyCode" binding:"max=16"`
	VisibilityScope string     `json:"visibilityScope" binding:"required"`
	StartAt         *time.Time `json:"startAt"`
	EndAt           *time.Time `json:"endAt"`
	Enabled         bool       `json:"enabled"`
	Reason          string     `json:"reason" binding:"max=500"`
}

type SiteBannerSortInput struct {
	IDs    []uint64 `json:"ids" binding:"required,min=1"`
	Reason string   `json:"reason" binding:"max=500"`
}

type siteBannerRecord struct {
	ID                     uint64              `gorm:"primaryKey" json:"id"`
	TenantID               uint64              `gorm:"column:tenant_id" json:"tenantID"`
	SiteID                 uint64              `gorm:"column:site_id" json:"siteID"`
	Name                   string              `gorm:"column:name" json:"name"`
	Title                  string              `gorm:"column:title" json:"title"`
	Subtitle               string              `gorm:"column:subtitle" json:"subtitle"`
	DesktopAssetID         uint64              `gorm:"column:desktop_asset_id" json:"desktopAssetID"`
	MobileAssetID          uint64              `gorm:"column:mobile_asset_id" json:"mobileAssetID"`
	DesktopAssetIDDraft    uint64              `gorm:"column:desktop_asset_id_draft" json:"desktopAssetIDDraft"`
	MobileAssetIDDraft     uint64              `gorm:"column:mobile_asset_id_draft" json:"mobileAssetIDDraft"`
	ActionType             string              `gorm:"column:action_type" json:"actionType"`
	ActionTarget           string              `gorm:"column:action_target" json:"actionTarget"`
	SortOrder              int                 `gorm:"column:sort_order" json:"sortOrder"`
	LanguageCode           string              `gorm:"column:language_code" json:"languageCode"`
	CountryCode            string              `gorm:"column:country_code" json:"countryCode"`
	CurrencyCode           string              `gorm:"column:currency_code" json:"currencyCode"`
	VisibilityScope        string              `gorm:"column:visibility_scope" json:"visibilityScope"`
	StartAt                *time.Time          `gorm:"column:start_at" json:"startAt"`
	EndAt                  *time.Time          `gorm:"column:end_at" json:"endAt"`
	Enabled                bool                `gorm:"column:enabled" json:"enabled"`
	DraftPayload           string              `gorm:"column:draft_payload" json:"-"`
	PublishedPayload       string              `gorm:"column:published_payload" json:"-"`
	PublishStatus          string              `gorm:"column:publish_status" json:"publishStatus"`
	ConfigVersion          uint64              `gorm:"column:config_version" json:"configVersion"`
	CreatedBy              uint64              `gorm:"column:created_by" json:"createdBy"`
	UpdatedBy              uint64              `gorm:"column:updated_by" json:"updatedBy"`
	DraftUpdatedAt         *time.Time          `gorm:"column:draft_updated_at" json:"draftUpdatedAt"`
	PublishedAt            *time.Time          `gorm:"column:published_at" json:"publishedAt"`
	CreatedAt              time.Time           `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt              time.Time           `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt              *time.Time          `gorm:"column:deleted_at" json:"-"`
	DraftConfiguration     SiteBannerDraft     `gorm:"-" json:"draftConfiguration"`
	PublishedConfiguration SiteBannerDraft     `gorm:"-" json:"publishedConfiguration"`
	DesktopAsset           *assetLibraryRecord `gorm:"-" json:"desktopAsset,omitempty"`
	MobileAsset            *assetLibraryRecord `gorm:"-" json:"mobileAsset,omitempty"`
	IsFallback             bool                `gorm:"-" json:"isFallback"`
	CanEdit                bool                `gorm:"-" json:"canEdit"`
	CanDelete              bool                `gorm:"-" json:"canDelete"`
	Visible                bool                `gorm:"-" json:"visible"`
	FallbackSourceID       uint64              `gorm:"-" json:"fallbackSourceID"`
	FallbackImageURL       string              `gorm:"-" json:"fallbackImageURL"`
	RowKey                 string              `gorm:"-" json:"rowKey"`
}

func (siteBannerRecord) TableName() string { return "game_site_banners" }

type PublicBanner struct {
	ID              uint64     `json:"id"`
	Name            string     `json:"name"`
	Title           string     `json:"title"`
	Subtitle        string     `json:"subtitle"`
	DesktopImageURL string     `json:"desktopImageUrl"`
	MobileImageURL  string     `json:"mobileImageUrl"`
	ActionType      string     `json:"actionType"`
	ActionTarget    string     `json:"actionTarget"`
	SortOrder       int        `json:"sortOrder"`
	StartAt         *time.Time `json:"startAt"`
	EndAt           *time.Time `json:"endAt"`
	ConfigVersion   uint64     `json:"configVersion"`
}

func (s *Service) ListSiteBanners(q ListQuery) (PageResult, error) {
	var rows []siteBannerRecord
	db := tenancy.ApplySiteStrict(s.db.Model(&siteBannerRecord{}), "", q.Scope).Where("deleted_at IS NULL")
	db = contains(db, "name", q.Values["keyword"])
	db = equal(db, "publish_status", strings.ToUpper(q.Values["publishStatus"]))
	if !q.Scope.TenantScoped() {
		result, err := page(db.Order("sort_order ASC,id ASC"), q, &rows)
		if err != nil {
			return result, err
		}
		for i := range rows {
			s.hydrateSiteBanner(&rows[i], q.Scope)
			rows[i].CanEdit, rows[i].CanDelete, rows[i].Visible = true, true, rows[i].DraftConfiguration.Enabled
			rows[i].RowKey = "banner-" + strconv.FormatUint(rows[i].ID, 10)
		}
		result.List = rows
		return result, nil
	}
	if err := db.Order("sort_order ASC,id ASC").Find(&rows).Error; err != nil {
		return PageResult{}, err
	}
	for i := range rows {
		s.hydrateSiteBanner(&rows[i], q.Scope)
		rows[i].CanEdit, rows[i].CanDelete, rows[i].Visible = true, true, rows[i].DraftConfiguration.Enabled
		rows[i].RowKey = "banner-" + strconv.FormatUint(rows[i].ID, 10)
	}
	list, total, p, ps := operationdefaults.Page(rows, q.Page, q.PageSize)
	return PageResult{List: list, Total: total, Page: p, PageSize: ps}, nil
}

func (s *Service) GetSiteBanner(id uint64, scope tenancy.Scope) (map[string]any, error) {
	row, err := s.siteBannerByID(s.db, id, scope, false)
	if err != nil {
		return nil, err
	}
	s.hydrateSiteBanner(&row, scope)
	logs, _ := s.operationLogs("site_banner", id, scope)
	return map[string]any{"overview": row, "operationLogs": logs}, nil
}

func (s *Service) SaveSiteBanner(id uint64, in SiteBannerInput, actor Actor) (uint64, error) {
	draft, err := normalizeSiteBannerInput(in)
	if err != nil {
		return 0, err
	}
	if !actor.Scope.SiteScoped() {
		return 0, newError("BANNER_SCOPE_REQUIRED", "tenant site scope is required")
	}
	dimensions, err := loadTenantSiteDimensions(s.db, actor.Scope)
	if err != nil {
		return 0, err
	}
	draft.CountryCode = strings.ToUpper(strings.TrimSpace(dimensions.CountryCode))
	draft.CurrencyCode = strings.ToUpper(strings.TrimSpace(dimensions.CurrencyCode))
	draft.LanguageCode = strings.TrimSpace(dimensions.DefaultLanguage)
	draft.DesktopImageURL = strings.TrimSpace(in.DesktopImageURL)
	draft.MobileImageURL = strings.TrimSpace(in.MobileImageURL)
	if draft.DesktopAssetID == 0 && strings.TrimSpace(draft.DesktopImageURL) == "" {
		return 0, newError("BANNER_DESKTOP_ASSET_REQUIRED", "desktop banner asset is required")
	}
	if err := s.validateBannerAssets(s.db, actor.Scope, draft.DesktopAssetID, draft.MobileAssetID); err != nil {
		return 0, err
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var before siteBannerRecord
		if id > 0 {
			var err error
			before, err = s.siteBannerByID(tx, id, actor.Scope, true)
			if err != nil {
				return err
			}
		}
		nowValue := time.Now()
		payload, _ := json.Marshal(draft)
		updates := map[string]any{
			"desktop_asset_id_draft": draft.DesktopAssetID, "mobile_asset_id_draft": draft.MobileAssetID,
			"draft_payload": string(payload), "publish_status": BannerPublishDraft, "draft_updated_at": nowValue,
			"updated_by": actor.ID, "updated_at": nowValue,
		}
		if id == 0 {
			row := siteBannerRecord{TenantID: actor.Scope.TenantID, SiteID: actor.Scope.SiteID, Name: draft.Name,
				DesktopAssetIDDraft: draft.DesktopAssetID, MobileAssetIDDraft: draft.MobileAssetID, DraftPayload: string(payload),
				PublishedPayload: "{}",
				PublishStatus:    BannerPublishDraft, CreatedBy: actor.ID, UpdatedBy: actor.ID, DraftUpdatedAt: &nowValue,
				CreatedAt: nowValue, UpdatedAt: nowValue}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			id = row.ID
		} else if err := tenancy.ApplySiteStrict(tx.Model(&siteBannerRecord{}), "", actor.Scope).Where("id=?", id).Updates(updates).Error; err != nil {
			return err
		}
		if err := markBannerSiteDraft(tx, actor); err != nil {
			return err
		}
		return audit(tx, actor, "operation:banner:update", "save_draft", "site_banner", id, before, draft, in.Reason)
	})
	return id, err
}

func (s *Service) CopySiteBanner(id uint64, actor Actor) (uint64, error) {
	row, err := s.siteBannerByID(s.db, id, actor.Scope, false)
	if err != nil {
		return 0, err
	}
	draft := decodeBannerDraft(row.DraftPayload)
	if strings.TrimSpace(row.DraftPayload) == "" {
		draft = decodeBannerDraft(row.PublishedPayload)
	}
	draft.Name = strings.TrimSpace(draft.Name) + " Copy"
	draft.Enabled = false
	return s.SaveSiteBanner(0, SiteBannerInput{Name: draft.Name, Title: draft.Title, Subtitle: draft.Subtitle,
		DesktopAssetID: draft.DesktopAssetID, MobileAssetID: draft.MobileAssetID, ActionType: draft.ActionType,
		ActionTarget: draft.ActionTarget, SortOrder: draft.SortOrder + 1, LanguageCode: draft.LanguageCode,
		CountryCode: draft.CountryCode, CurrencyCode: draft.CurrencyCode, VisibilityScope: draft.VisibilityScope,
		StartAt: draft.StartAt, EndAt: draft.EndAt, Enabled: false, Reason: "copy banner draft"}, actor)
}

func (s *Service) DeleteSiteBanner(id uint64, actor Actor) error {
	row, err := s.siteBannerByID(s.db, id, actor.Scope, false)
	if err != nil {
		return err
	}
	draft := decodeBannerDraft(row.DraftPayload)
	if strings.TrimSpace(row.DraftPayload) == "" {
		draft = decodeBannerDraft(row.PublishedPayload)
	}
	draft.Deleted = true
	payload, _ := json.Marshal(draft)
	return s.db.Transaction(func(tx *gorm.DB) error {
		nowValue := time.Now()
		if err := tenancy.ApplySiteStrict(tx.Model(&siteBannerRecord{}), "", actor.Scope).Where("id=?", id).Updates(map[string]any{
			"draft_payload": string(payload), "publish_status": BannerPublishDraft, "draft_updated_at": nowValue,
			"updated_by": actor.ID, "updated_at": nowValue,
		}).Error; err != nil {
			return err
		}
		if err := markBannerSiteDraft(tx, actor); err != nil {
			return err
		}
		return audit(tx, actor, "operation:banner:delete", "delete_draft", "site_banner", id, row, draft, "")
	})
}

func (s *Service) SortSiteBanners(in SiteBannerSortInput, actor Actor) error {
	if !actor.Scope.SiteScoped() {
		return newError("BANNER_SCOPE_REQUIRED", "tenant site scope is required")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		for index, id := range in.IDs {
			row, err := s.siteBannerByID(tx, id, actor.Scope, true)
			if err != nil {
				return err
			}
			draft := decodeBannerDraft(row.DraftPayload)
			if strings.TrimSpace(row.DraftPayload) == "" {
				draft = decodeBannerDraft(row.PublishedPayload)
			}
			draft.SortOrder = index
			payload, _ := json.Marshal(draft)
			if err := tx.Model(&siteBannerRecord{}).Where("id=? AND tenant_id=? AND site_id=?", id, actor.Scope.TenantID, actor.Scope.SiteID).
				Updates(map[string]any{"draft_payload": string(payload), "publish_status": BannerPublishDraft, "draft_updated_at": time.Now(), "updated_by": actor.ID, "updated_at": time.Now()}).Error; err != nil {
				return err
			}
		}
		if err := markBannerSiteDraft(tx, actor); err != nil {
			return err
		}
		return audit(tx, actor, "operation:banner:update", "sort_draft", "site_banner", 0, nil, in.IDs, in.Reason)
	})
}

func (s *Service) PublicSiteBanners(scope tenancy.Scope, language, country string, loggedIn bool, at time.Time) ([]PublicBanner, error) {
	if scope.TenantID == 0 || scope.SiteID == 0 {
		return []PublicBanner{}, nil
	}
	visibility := []string{"ALL"}
	languages := bannerLanguageCandidates(language)
	if loggedIn {
		visibility = append(visibility, "LOGGED_IN")
	} else {
		visibility = append(visibility, "LOGGED_OUT")
	}
	var rows []struct {
		ID               uint64     `gorm:"column:id"`
		Name             string     `gorm:"column:name"`
		Title            string     `gorm:"column:title"`
		Subtitle         string     `gorm:"column:subtitle"`
		ActionType       string     `gorm:"column:action_type"`
		ActionTarget     string     `gorm:"column:action_target"`
		SortOrder        int        `gorm:"column:sort_order"`
		StartAt          *time.Time `gorm:"column:start_at"`
		EndAt            *time.Time `gorm:"column:end_at"`
		ConfigVersion    uint64     `gorm:"column:config_version"`
		DesktopObjectKey string     `gorm:"column:desktop_object_key"`
		MobileObjectKey  string     `gorm:"column:mobile_object_key"`
		PublishedPayload string     `gorm:"column:published_payload"`
	}
	err := s.db.Table("game_site_banners b").Select("b.*,da.object_key desktop_object_key,ma.object_key mobile_object_key").
		Joins("LEFT JOIN game_asset_library da ON da.id=b.desktop_asset_id AND da.status=? AND da.deleted_at IS NULL", AssetStatusActive).
		Joins("LEFT JOIN game_asset_library ma ON ma.id=b.mobile_asset_id AND ma.status=? AND ma.deleted_at IS NULL", AssetStatusActive).
		Where("b.tenant_id=? AND b.site_id=? AND b.config_version>0 AND b.enabled=1 AND b.deleted_at IS NULL", scope.TenantID, scope.SiteID).
		Where("b.visibility_scope IN ?", visibility).
		Where("(b.language_code='' OR b.language_code='ALL' OR LOWER(b.language_code) IN ?)", languages).
		Where("(b.country_code='' OR b.country_code='ALL' OR UPPER(b.country_code)=?)", strings.ToUpper(strings.TrimSpace(country))).
		Where("(b.start_at IS NULL OR b.start_at<=?) AND (b.end_at IS NULL OR b.end_at>=?)", at, at).
		Order("b.sort_order ASC,b.id ASC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]PublicBanner, 0, len(rows))
	for _, row := range rows {
		draft := decodeBannerDraft(row.PublishedPayload)
		desktop := publicAssetURL(row.DesktopObjectKey)
		if desktop == "" {
			desktop = operationAssetURL(draft.DesktopImageURL)
		}
		if strings.TrimSpace(desktop) == "" {
			continue
		}
		mobile := publicAssetURL(row.MobileObjectKey)
		if mobile == "" {
			mobile = operationAssetURL(draft.MobileImageURL)
		}
		if strings.TrimSpace(mobile) == "" {
			mobile = desktop
		}
		out = append(out, PublicBanner{ID: row.ID, Name: row.Name, Title: row.Title, Subtitle: row.Subtitle,
			DesktopImageURL: desktop, MobileImageURL: mobile, ActionType: row.ActionType, ActionTarget: row.ActionTarget,
			SortOrder: row.SortOrder, StartAt: row.StartAt, EndAt: row.EndAt, ConfigVersion: row.ConfigVersion})
	}
	return out, nil
}

func legacyBannerAction(row bannerRecord) (string, string) {
	target := strings.TrimSpace(firstNonEmpty(row.TargetValue, row.LinkURL))
	kind := strings.ToUpper(strings.TrimSpace(row.TargetType))
	if target == "" {
		return "NONE", ""
	}
	if kind == "EXTERNAL_URL" || strings.HasPrefix(strings.ToLower(target), "https://") {
		return "OPEN_EXTERNAL_URL", target
	}
	return "OPEN_INTERNAL_PAGE", target
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func bannerLanguageCandidates(language string) []string {
	language = strings.ToLower(strings.TrimSpace(language))
	if language == "" {
		return []string{""}
	}
	values := []string{language}
	if index := strings.IndexAny(language, "-_"); index > 0 {
		values = append(values, language[:index])
	}
	return values
}

func normalizeSiteBannerInput(in SiteBannerInput) (SiteBannerDraft, error) {
	if err := validateTime(in.StartAt, in.EndAt); err != nil {
		return SiteBannerDraft{}, err
	}
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Title) == "" {
		return SiteBannerDraft{}, newError("BANNER_FIELDS_REQUIRED", "banner name and title are required")
	}
	action := strings.ToUpper(strings.TrimSpace(in.ActionType))
	if !allowedBannerActions[action] {
		return SiteBannerDraft{}, newError("INVALID_BANNER_ACTION", "invalid banner action")
	}
	target := strings.TrimSpace(in.ActionTarget)
	if action == "OPEN_EXTERNAL_URL" {
		parsed, err := url.ParseRequestURI(target)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return SiteBannerDraft{}, newError("INVALID_BANNER_TARGET", "external URL must use HTTPS")
		}
	}
	if action != "NONE" && action != "OPEN_DEPOSIT" && action != "OPEN_SUPPORT" && target == "" {
		return SiteBannerDraft{}, newError("BANNER_TARGET_REQUIRED", "banner action target is required")
	}
	visibility := strings.ToUpper(strings.TrimSpace(in.VisibilityScope))
	if visibility != "ALL" && visibility != "LOGGED_OUT" && visibility != "LOGGED_IN" {
		return SiteBannerDraft{}, newError("INVALID_BANNER_VISIBILITY", "invalid banner visibility")
	}
	if strings.TrimSpace(in.DesktopImageURL) != "" && !storage.IsAssetURL(in.DesktopImageURL) {
		return SiteBannerDraft{}, newError("BANNER_IMAGE_INVALID", "desktop banner image is not a valid asset URL")
	}
	if strings.TrimSpace(in.MobileImageURL) != "" && !storage.IsAssetURL(in.MobileImageURL) {
		return SiteBannerDraft{}, newError("BANNER_IMAGE_INVALID", "mobile banner image is not a valid asset URL")
	}
	return SiteBannerDraft{Name: strings.TrimSpace(in.Name), Title: strings.TrimSpace(in.Title), Subtitle: strings.TrimSpace(in.Subtitle),
		DesktopAssetID: in.DesktopAssetID, MobileAssetID: in.MobileAssetID, DesktopImageURL: strings.TrimSpace(in.DesktopImageURL), MobileImageURL: strings.TrimSpace(in.MobileImageURL), ActionType: action, ActionTarget: target,
		SortOrder: in.SortOrder, LanguageCode: normalizeBannerDimension(in.LanguageCode), CountryCode: normalizeBannerDimension(in.CountryCode),
		CurrencyCode: normalizeBannerDimension(in.CurrencyCode), VisibilityScope: visibility, StartAt: in.StartAt, EndAt: in.EndAt, Enabled: in.Enabled}, nil
}

func normalizeBannerDimension(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "ALL"
	}
	return strings.ToUpper(value)
}

func (s *Service) siteBannerByID(db *gorm.DB, id uint64, scope tenancy.Scope, lock bool) (siteBannerRecord, error) {
	var row siteBannerRecord
	query := tenancy.ApplySiteStrict(db.Model(&siteBannerRecord{}), "", scope).Where("id=? AND deleted_at IS NULL", id)
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.First(&row).Error; err != nil {
		return row, notFound(err, "BANNER_NOT_FOUND")
	}
	return row, nil
}

func (s *Service) validateBannerAssets(db *gorm.DB, scope tenancy.Scope, ids ...uint64) error {
	for _, id := range ids {
		if id == 0 {
			continue
		}
		var asset assetLibraryRecord
		if err := s.assetVisibleQuery(db.Model(&assetLibraryRecord{}), scope).Where("id=? AND asset_type=? AND status=? AND deleted_at IS NULL", id, "IMAGE", AssetStatusActive).First(&asset).Error; err != nil {
			return newError("BANNER_ASSET_INVALID", "banner asset is missing, inactive, or not visible")
		}
	}
	return nil
}

func (s *Service) hydrateSiteBanner(row *siteBannerRecord, scope tenancy.Scope) {
	row.DraftConfiguration = decodeBannerDraft(row.DraftPayload)
	row.PublishedConfiguration = decodeBannerDraft(row.PublishedPayload)
	s.hydrateLegacyBannerImages(row)
	if row.DesktopAssetIDDraft > 0 {
		row.DesktopAsset = s.bannerAsset(row.DesktopAssetIDDraft, scope)
	}
	if row.MobileAssetIDDraft > 0 {
		row.MobileAsset = s.bannerAsset(row.MobileAssetIDDraft, scope)
	}
}

// Older tenant banner rows were created before image URLs were kept in the
// site-banner payload. Resolve those rows from the platform catalog so the
// admin can preview and edit them without inventing a zero asset selection.
func (s *Service) hydrateLegacyBannerImages(row *siteBannerRecord) {
	if row.DraftConfiguration.DesktopImageURL != "" && row.PublishedConfiguration.DesktopImageURL != "" {
		return
	}
	var legacy bannerRecord
	query := s.db.Table("game_banners").
		Where("tenant_id=0 AND site_id=0 AND deleted_at IS NULL").
		Where("banner_code=? OR title_en=? OR title_zh=? OR title=?", row.Name, row.Name, row.Name, row.Name).
		Order("id ASC")
	if query.First(&legacy).Error != nil {
		return
	}
	if row.DraftConfiguration.DesktopImageURL == "" {
		row.DraftConfiguration.DesktopImageURL = operationAssetURL(legacy.ImageURL)
	}
	if row.DraftConfiguration.MobileImageURL == "" {
		row.DraftConfiguration.MobileImageURL = operationAssetURL(firstNonEmpty(legacy.MobileImageURL, legacy.ImageURL))
	}
	if row.PublishedConfiguration.DesktopImageURL == "" {
		row.PublishedConfiguration.DesktopImageURL = operationAssetURL(legacy.ImageURL)
	}
	if row.PublishedConfiguration.MobileImageURL == "" {
		row.PublishedConfiguration.MobileImageURL = operationAssetURL(firstNonEmpty(legacy.MobileImageURL, legacy.ImageURL))
	}
}

func (s *Service) bannerAsset(id uint64, scope tenancy.Scope) *assetLibraryRecord {
	var row assetLibraryRecord
	if s.assetVisibleQuery(s.db.Model(&assetLibraryRecord{}), scope).Where("id=? AND deleted_at IS NULL", id).First(&row).Error != nil {
		return nil
	}
	s.signAsset(&row)
	return &row
}

func decodeBannerDraft(raw string) SiteBannerDraft {
	var out SiteBannerDraft
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func publicAssetURL(objectKey string) string {
	if strings.TrimSpace(objectKey) == "" {
		return ""
	}
	value, err := storage.PublicURL(context.Background(), objectKey)
	if err != nil {
		return objectKey
	}
	return value
}

func markBannerSiteDraft(tx *gorm.DB, actor Actor) error {
	return tx.Table("game_saas_sites").Where("id=? AND tenant_id=? AND deleted_at IS NULL", actor.Scope.SiteID, actor.Scope.TenantID).
		Updates(map[string]any{"config_status": BannerPublishDraft, "config_draft_updated_at": time.Now(), "updated_by": actor.ID, "updated_at": time.Now()}).Error
}
