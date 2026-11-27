package operationadmin

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/content"
	"casino/internal/storage"
)

const (
	AssetScopePlatform = "PLATFORM"
	AssetScopeSite     = "SITE"

	AssetStatusActive  = "ACTIVE"
	AssetStatusDeleted = "DELETED"
)

type assetLibraryRecord struct {
	ID uint64 `gorm:"primaryKey" json:"id"`

	SiteID         uint64     `gorm:"column:site_id" json:"siteID"`
	AssetScope     string     `gorm:"column:asset_scope" json:"assetScope"`
	AssetType      string     `gorm:"column:asset_type" json:"assetType"`
	UsageType      string     `gorm:"column:usage_type" json:"usageType"`
	DisplayName    string     `gorm:"column:display_name" json:"displayName"`
	OriginalName   string     `gorm:"column:original_name" json:"originalName"`
	ObjectKey      string     `gorm:"column:object_key" json:"objectKey"`
	PublicURL      string     `gorm:"column:public_url" json:"publicUrl"`
	MimeType       string     `gorm:"column:mime_type" json:"mimeType"`
	FileExt        string     `gorm:"column:file_ext" json:"fileExt"`
	FileSize       int64      `gorm:"column:file_size" json:"fileSize"`
	ChecksumSHA256 string     `gorm:"column:checksum_sha256" json:"checksumSha256"`
	Width          int        `gorm:"column:width" json:"width"`
	Height         int        `gorm:"column:height" json:"height"`
	DurationMs     int64      `gorm:"column:duration_ms" json:"durationMs"`
	ReferenceCount int64      `gorm:"column:reference_count" json:"referenceCount"`
	IsPublic       bool       `gorm:"column:is_public" json:"isPublic"`
	Tags           string     `gorm:"column:tags" json:"tags"`
	AltText        string     `gorm:"column:alt_text" json:"altText"`
	MetadataJSON   string     `gorm:"column:metadata_json" json:"metadataJson"`
	Status         string     `gorm:"column:status" json:"status"`
	CreatedBy      uint64     `gorm:"column:created_by" json:"createdBy"`
	CreatedByName  string     `gorm:"column:created_by_name" json:"createdByName"`
	UpdatedBy      uint64     `gorm:"column:updated_by" json:"updatedBy"`
	CreatedAt      time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt      time.Time  `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt      *time.Time `gorm:"column:deleted_at" json:"-"`
	References     []AssetRef `gorm:"-" json:"references,omitempty"`
}

func (assetLibraryRecord) TableName() string { return "game_asset_library" }

type AssetUploadInput struct {
	AssetType   string `json:"assetType"`
	UsageType   string `json:"usageType"`
	AssetScope  string `json:"assetScope"`
	DisplayName string `json:"displayName"`
	AltText     string `json:"altText"`
	Tags        string `json:"tags"`
}

type AssetUpdateInput struct {
	UsageType   string `json:"usageType"`
	DisplayName string `json:"displayName"`
	AltText     string `json:"altText"`
	Tags        string `json:"tags"`
	Reason      string `json:"reason"`
}

type AssetRef struct {
	Table    string `json:"table"`
	Column   string `json:"column"`
	RecordID uint64 `json:"recordID"`

	SiteID     uint64 `json:"siteID"`
	TargetType string `json:"targetType"`
}

func (s *Service) ListAssets(q ListQuery) (PageResult, error) {
	var rows []assetLibraryRecord
	db := s.assetVisibleQuery(s.db.Model(&assetLibraryRecord{}), q.Scope).Where("deleted_at IS NULL")
	db = contains(db, "display_name", q.Values["keyword"])
	db = equal(db, "asset_type", strings.ToUpper(strings.TrimSpace(q.Values["assetType"])))
	db = equal(db, "usage_type", strings.ToUpper(strings.TrimSpace(q.Values["usageType"])))
	db = equal(db, "asset_scope", strings.ToUpper(strings.TrimSpace(q.Values["assetScope"])))
	db = equal(db, "status", strings.ToUpper(strings.TrimSpace(q.Values["status"])))
	result, err := page(db.Order("created_at DESC,id DESC"), q, &rows)
	if err != nil {
		return result, err
	}
	for i := range rows {
		s.signAsset(&rows[i])
	}
	result.List = rows
	return result, nil
}

func (s *Service) GetAsset(id uint64, scope content.Scope) (map[string]any, error) {
	var row assetLibraryRecord
	if err := s.assetVisibleQuery(s.db.Model(&assetLibraryRecord{}), scope).
		Where("id=? AND deleted_at IS NULL", id).
		First(&row).Error; err != nil {
		return nil, notFound(err, "ASSET_NOT_FOUND")
	}
	row.References = s.AssetReferences(row)
	row.ReferenceCount = int64(len(row.References))
	s.signAsset(&row)
	logs, err := s.operationLogs("asset_library", id, scope)
	if err != nil {
		return nil, err
	}
	return map[string]any{"overview": row, "references": row.References, "operationLogs": logs}, nil
}

func (s *Service) UploadAsset(ctx context.Context, file *multipart.FileHeader, input AssetUploadInput, actor Actor) (assetLibraryRecord, error) {
	if file == nil {
		return assetLibraryRecord{}, newError("FILE_REQUIRED", "file is required")
	}
	normalized, err := normalizeAssetUploadInput(input, file, actor)
	if err != nil {
		return assetLibraryRecord{}, err
	}
	data, err := readAssetFile(file, normalized.AssetType)
	if err != nil {
		return assetLibraryRecord{}, err
	}
	if normalized.AssetType == "LOTTIE" && normalized.FileExt == ".json" && !json.Valid(data) {
		return assetLibraryRecord{}, newError("INVALID_ASSET", "invalid lottie json")
	}
	width, height := imageDimensions(data, normalized.FileExt)
	durationMs := audioDurationMs(data, normalized.FileExt)
	hash := sha256.Sum256(data)
	name, err := storage.RandomName(normalized.FileExt)
	if err != nil {
		return assetLibraryRecord{}, err
	}
	objectKey := path.Join("admin/assets", strings.ToLower(normalized.AssetScope), scopePath(actor.Scope, normalized.AssetScope), strings.ToLower(normalized.AssetType), name)
	stored, err := storage.PutBytes(ctx, objectKey, data, normalized.MimeType)
	if err != nil {
		return assetLibraryRecord{}, err
	}
	row := assetLibraryRecord{

		SiteID:         normalized.SiteID,
		AssetScope:     normalized.AssetScope,
		AssetType:      normalized.AssetType,
		UsageType:      normalized.UsageType,
		DisplayName:    normalized.DisplayName,
		OriginalName:   strings.TrimSpace(file.Filename),
		ObjectKey:      stored.Key,
		PublicURL:      stored.URL,
		MimeType:       normalized.MimeType,
		FileExt:        normalized.FileExt,
		FileSize:       file.Size,
		ChecksumSHA256: hex.EncodeToString(hash[:]),
		Width:          width,
		Height:         height,
		DurationMs:     durationMs,
		ReferenceCount: 0,
		IsPublic:       normalized.AssetScope == AssetScopePlatform,
		Tags:           normalized.Tags,
		AltText:        normalized.AltText,
		MetadataJSON:   marshalAssetMetadata(width, height, durationMs),
		Status:         AssetStatusActive,
		CreatedBy:      actor.ID,
		CreatedByName:  actor.Username,
		UpdatedBy:      actor.ID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		return audit(tx, actor, "operation:asset:upload", "upload", "asset_library", row.ID, nil, row, "")
	}); err != nil {
		return assetLibraryRecord{}, err
	}
	s.signAsset(&row)
	return row, nil
}

func (s *Service) UpdateAsset(id uint64, input AssetUpdateInput, actor Actor) error {
	input.UsageType = normalizeAssetUsage(input.UsageType)
	if strings.TrimSpace(input.Reason) == "" {
		return newError("REASON_REQUIRED", "reason is required")
	}
	var before assetLibraryRecord
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.assetVisibleQuery(tx.Model(&assetLibraryRecord{}).Clauses(clause.Locking{Strength: "UPDATE"}), actor.Scope).
			Where("id=? AND deleted_at IS NULL", id).
			First(&before).Error; err != nil {
			return notFound(err, "ASSET_NOT_FOUND")
		}
		updates := map[string]any{
			"usage_type":   input.UsageType,
			"display_name": strings.TrimSpace(input.DisplayName),
			"alt_text":     strings.TrimSpace(input.AltText),
			"tags":         strings.TrimSpace(input.Tags),
			"updated_by":   actor.ID,
			"updated_at":   time.Now(),
		}
		if updates["display_name"] == "" {
			updates["display_name"] = before.DisplayName
		}
		if err := tx.Model(&assetLibraryRecord{}).Where("id=?", id).Updates(updates).Error; err != nil {
			return err
		}
		return audit(tx, actor, "operation:asset:update", "update", "asset_library", id, before, updates, input.Reason)
	})
}

func (s *Service) DeleteAsset(id uint64, reason string, actor Actor) error {
	if strings.TrimSpace(reason) == "" {
		return newError("REASON_REQUIRED", "reason is required")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var before assetLibraryRecord
		if err := s.assetVisibleQuery(tx.Model(&assetLibraryRecord{}).Clauses(clause.Locking{Strength: "UPDATE"}), actor.Scope).
			Where("id=? AND deleted_at IS NULL", id).
			First(&before).Error; err != nil {
			return notFound(err, "ASSET_NOT_FOUND")
		}
		refs := s.assetReferences(tx, before)
		if len(refs) > 0 {
			_ = tx.Model(&assetLibraryRecord{}).Where("id=?", id).Updates(map[string]any{"reference_count": len(refs), "updated_at": time.Now()}).Error
			return newError("ASSET_IN_USE", "asset is referenced and cannot be deleted")
		}
		now := time.Now()
		updates := map[string]any{"status": AssetStatusDeleted, "deleted_at": now, "updated_by": actor.ID, "updated_at": now}
		if err := tx.Model(&assetLibraryRecord{}).Where("id=?", id).Updates(updates).Error; err != nil {
			return err
		}
		return audit(tx, actor, "operation:asset:delete", "delete", "asset_library", id, before, updates, reason)
	})
}

func (s *Service) GetAssetReferences(id uint64, scope content.Scope) ([]AssetRef, error) {
	var row assetLibraryRecord
	if err := s.assetVisibleQuery(s.db.Model(&assetLibraryRecord{}), scope).Where("id=? AND deleted_at IS NULL", id).First(&row).Error; err != nil {
		return nil, notFound(err, "ASSET_NOT_FOUND")
	}
	refs := s.AssetReferences(row)
	_ = s.db.Model(&assetLibraryRecord{}).Where("id=?", id).Updates(map[string]any{"reference_count": len(refs), "updated_at": time.Now()}).Error
	return refs, nil
}

func (s *Service) AssetReferences(row assetLibraryRecord) []AssetRef {
	return s.assetReferences(s.db, row)
}

func (s *Service) assetReferences(db *gorm.DB, row assetLibraryRecord) []AssetRef {
	values := assetReferenceValues(row)
	if len(values) == 0 {
		return nil
	}
	sources := []struct {
		table      string
		column     string
		targetType string
	}{
		{"game_banners", "image_url", "banner"},
		{"game_banners", "mobile_image_url", "banner"},
		{"game_operation_music", "audio_url", "music"},
		{"game_game_categories", "icon_url", "game_category"},
		{"game_game_categories", "banner_url", "game_category"},
		{"game_game_collections", "banner_url", "game_collection"},
		{"game_games", "local_thumbnail_url", "game"},
		{"game_games", "background_url", "game"},
		{"game_games", "launch_image_url", "game"},
		{"game_social_links", "icon_url", "social_link"},
		{"game_vip_levels", "icon_url", "vip_level"},
		{"game_payment_methods", "icon_url", "payment_method"},
		{"game_deposit_vouchers", "icon_url", "voucher"},
	}
	refs := make([]AssetRef, 0)
	for _, source := range sources {
		var rows []struct {
			ID uint64 `gorm:"column:id"`

			SiteID uint64 `gorm:"column:site_id"`
		}
		query := db.Table(source.table).
			Select("id, COALESCE(site_id, 0) AS site_id").
			Where(source.column+" IN ?", values).
			Where("deleted_at IS NULL")
		if row.AssetScope == AssetScopeSite {
			query = query
			if row.SiteID > 0 {
				query = query.Where("site_id IN (0, ?)", row.SiteID)
			}
		}
		if err := query.Find(&rows).Error; err != nil {
			continue
		}
		for _, matched := range rows {
			refs = append(refs, AssetRef{
				Table:    source.table,
				Column:   source.column,
				RecordID: matched.ID,

				SiteID:     matched.SiteID,
				TargetType: source.targetType,
			})
		}
	}
	for _, column := range []string{"desktop_asset_id", "mobile_asset_id"} {
		var rows []struct {
			ID uint64 `gorm:"column:id"`

			SiteID uint64 `gorm:"column:site_id"`
		}
		query := db.Table("game_site_banners").Select("id, site_id").Where(column+"=? AND deleted_at IS NULL", row.ID)
		if row.AssetScope == AssetScopeSite {
			query = query
			if row.SiteID > 0 {
				query = query.Where("site_id IN (0, ?)", row.SiteID)
			}
		}
		if err := query.Find(&rows).Error; err != nil {
			continue
		}
		for _, matched := range rows {
			refs = append(refs, AssetRef{Table: "game_site_banners", Column: column, RecordID: matched.ID,
				SiteID: matched.SiteID, TargetType: "banner"})
		}
	}
	for _, column := range []string{"icon_asset_id", "lottie_asset_id", "sound_asset_id"} {
		var rows []struct {
			ID uint64 `gorm:"column:id"`

			SiteID uint64 `gorm:"column:site_id"`
		}
		query := db.Table("game_h5_navigation_items").Select("id, site_id").Where(column+"=? AND deleted_at IS NULL", row.ID)
		if row.AssetScope == AssetScopeSite {
			query = query
			if row.SiteID > 0 {
				query = query.Where("site_id IN (0, ?)", row.SiteID)
			}
		}
		if err := query.Find(&rows).Error; err != nil {
			continue
		}
		for _, matched := range rows {
			refs = append(refs, AssetRef{Table: "game_h5_navigation_items", Column: column, RecordID: matched.ID,
				SiteID: matched.SiteID, TargetType: "h5_navigation"})
		}
	}
	return refs
}

func assetReferenceValues(row assetLibraryRecord) []string {
	values := make([]string, 0, 4)
	seen := map[string]bool{}
	for _, value := range []string{
		row.ObjectKey,
		row.PublicURL,
		storage.NormalizeReference(row.ObjectKey),
		storage.NormalizeReference(row.PublicURL),
	} {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		values = append(values, value)
	}
	return values
}

func (s *Service) assetVisibleQuery(db *gorm.DB, scope content.Scope) *gorm.DB {
	if !scope.SiteScoped() {
		return db
	}
	query := db
	if scope.SiteID > 0 {
		query = query.Where("site_id IN (0, ?)", scope.SiteID)
	}
	return query
}

func (s *Service) signAsset(row *assetLibraryRecord) {
	if row == nil || row.ObjectKey == "" {
		return
	}
	if url, err := storage.PublicURL(context.Background(), row.ObjectKey); err == nil && strings.TrimSpace(url) != "" {
		row.PublicURL = url
	}
}

type normalizedAssetUpload struct {
	SiteID      uint64
	AssetScope  string
	AssetType   string
	UsageType   string
	DisplayName string
	AltText     string
	Tags        string
	FileExt     string
	MimeType    string
}

func normalizeAssetUploadInput(input AssetUploadInput, file *multipart.FileHeader, actor Actor) (normalizedAssetUpload, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	assetType := normalizeAssetType(input.AssetType, ext)
	if assetType == "" {
		return normalizedAssetUpload{}, newError("INVALID_ASSET_TYPE", "unsupported asset type")
	}
	if !assetExtensionAllowed(assetType, ext) || file.Size <= 0 || file.Size > assetSizeLimit(assetType) {
		return normalizedAssetUpload{}, newError("INVALID_ASSET", "unsupported file type or size")
	}
	scope := strings.ToUpper(strings.TrimSpace(input.AssetScope))
	if scope == "" {
		scope = AssetScopeSite
		if actor.Scope.Platform {
			scope = AssetScopePlatform
		}
	}
	if scope != AssetScopePlatform && scope != AssetScopeSite {
		return normalizedAssetUpload{}, newError("INVALID_ASSET_SCOPE", "invalid asset scope")
	}
	if !actor.Scope.Platform {
		scope = AssetScopeSite
	}
	siteID := actor.Scope.SiteID
	if scope == AssetScopePlatform {
		siteID = 0
	}
	displayName := strings.TrimSpace(input.DisplayName)
	if displayName == "" {
		displayName = strings.TrimSuffix(path.Base(file.Filename), ext)
	}
	return normalizedAssetUpload{
		SiteID: siteID, AssetScope: scope, AssetType: assetType,
		UsageType: normalizeAssetUsage(input.UsageType), DisplayName: displayName,
		AltText: strings.TrimSpace(input.AltText), Tags: strings.TrimSpace(input.Tags),
		FileExt: ext, MimeType: storageContentType(file, ext),
	}, nil
}

func normalizeAssetType(value string, ext string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "IMAGE" || value == "AUDIO" || value == "VIDEO" || value == "LOTTIE" {
		return value
	}
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".avif", ".gif", ".svg", ".ico":
		return "IMAGE"
	case ".mp3", ".wav", ".ogg", ".m4a", ".aac":
		return "AUDIO"
	case ".mp4", ".webm", ".mov", ".m4v":
		return "VIDEO"
	case ".json", ".lottie":
		return "LOTTIE"
	default:
		return ""
	}
}

func normalizeAssetUsage(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	allowed := map[string]bool{
		"LOGO": true, "FAVICON": true, "BANNER": true, "BACKGROUND": true, "BUTTON_ICON": true,
		"GAME_PLACEHOLDER": true, "POPUP": true, "LOADING": true, "SOUND": true, "GENERAL": true,
	}
	if !allowed[value] {
		return "GENERAL"
	}
	return value
}

func assetExtensionAllowed(assetType string, ext string) bool {
	allowed := map[string]map[string]bool{
		"IMAGE":  {".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".avif": true, ".gif": true, ".svg": true, ".ico": true},
		"AUDIO":  {".mp3": true, ".wav": true, ".ogg": true, ".m4a": true, ".aac": true},
		"VIDEO":  {".mp4": true, ".webm": true, ".mov": true, ".m4v": true},
		"LOTTIE": {".json": true, ".lottie": true},
	}
	return allowed[assetType][ext]
}

func assetSizeLimit(assetType string) int64 {
	switch assetType {
	case "IMAGE":
		return 10 << 20
	case "LOTTIE":
		return 5 << 20
	case "AUDIO":
		return 30 << 20
	case "VIDEO":
		return 100 << 20
	default:
		return 0
	}
}

func readAssetFile(file *multipart.FileHeader, assetType string) ([]byte, error) {
	if file.Size > assetSizeLimit(assetType) {
		return nil, newError("INVALID_ASSET", "unsupported file type or size")
	}
	source, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer source.Close()
	limited := io.LimitReader(source, assetSizeLimit(assetType)+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) != file.Size || int64(len(data)) > assetSizeLimit(assetType) {
		return nil, newError("INVALID_ASSET", "unsupported file type or size")
	}
	return data, nil
}

func imageDimensions(data []byte, ext string) (int, int) {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif":
		config, _, err := image.DecodeConfig(bytes.NewReader(data))
		if err == nil {
			return config.Width, config.Height
		}
	}
	return 0, 0
}

func audioDurationMs(data []byte, ext string) int64 {
	if ext != ".wav" || len(data) < 44 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return 0
	}
	channels := int64(data[22]) | int64(data[23])<<8
	sampleRate := int64(data[24]) | int64(data[25])<<8 | int64(data[26])<<16 | int64(data[27])<<24
	bitsPerSample := int64(data[34]) | int64(data[35])<<8
	if channels <= 0 || sampleRate <= 0 || bitsPerSample <= 0 {
		return 0
	}
	dataSize := int64(data[40]) | int64(data[41])<<8 | int64(data[42])<<16 | int64(data[43])<<24
	bytesPerSecond := sampleRate * channels * bitsPerSample / 8
	if bytesPerSecond <= 0 {
		return 0
	}
	return dataSize * 1000 / bytesPerSecond
}

func storageContentType(file *multipart.FileHeader, ext string) string {
	contentType := strings.TrimSpace(file.Header.Get("Content-Type"))
	switch ext {
	case ".svg":
		return "image/svg+xml"
	case ".json":
		return "application/json"
	case ".lottie":
		return "application/zip"
	}
	if contentType != "" && contentType != "application/octet-stream" {
		return contentType
	}
	return ""
}

func marshalAssetMetadata(width int, height int, durationMs int64) string {
	data, _ := json.Marshal(map[string]any{"width": width, "height": height, "durationMs": durationMs})
	return string(data)
}

func scopePath(scope content.Scope, assetScope string) string {
	if assetScope == AssetScopePlatform {
		return "platform"
	}
	return "site-" + uintString(scope.SiteID)
}

func uintString(value uint64) string {
	return strconv.FormatUint(value, 10)
}
