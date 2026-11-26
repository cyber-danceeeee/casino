package operationdefaults

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/content"
	"casino/internal/model"
)

const (
	ResourcePromotion  = "PROMOTION"
	ResourceVIPLevel   = "VIP_LEVEL"
	ResourceMusic      = "MUSIC"
	ResourceSocialLink = "SOCIAL_LINK"
	ResourceVoucher    = "VOUCHER"
	ResourceBanner     = "BANNER"
)

var resourceTables = map[string]string{
	ResourcePromotion:  "game_activity_events",
	ResourceVIPLevel:   "game_vip_levels",
	ResourceMusic:      "game_operation_music",
	ResourceSocialLink: "game_social_links",
	ResourceVoucher:    "game_deposit_vouchers",
	ResourceBanner:     "game_banners",
}

type Metadata struct {
	IsFallback bool `gorm:"-" json:"isFallback"`
	CanEdit    bool `gorm:"-" json:"canEdit"`
	CanDelete  bool `gorm:"-" json:"canDelete"`
	Visible    bool `gorm:"-" json:"visible"`
}

func NormalizeResource(resource string) (string, error) {
	resource = strings.ToUpper(strings.TrimSpace(resource))
	if _, ok := resourceTables[resource]; !ok {
		return "", fmt.Errorf("unsupported operation default resource")
	}
	return resource, nil
}

func PlatformRows[T any](db *gorm.DB, resource string, query func(*gorm.DB) *gorm.DB) ([]T, error) {
	resource, err := NormalizeResource(resource)
	if err != nil {
		return nil, err
	}
	rows := []T{}
	base := db.Table(resourceTables[resource]).Where("site_id = 0 AND deleted_at IS NULL")
	if query != nil {
		base = query(base)
	}
	return rows, base.Find(&rows).Error
}

func VisibilityMap(db *gorm.DB, scope content.Scope, resource string, ids []uint64) (map[uint64]bool, error) {
	values := make(map[uint64]bool, len(ids))
	if !scope.SiteScoped() || len(ids) == 0 {
		return values, nil
	}
	resource, err := NormalizeResource(resource)
	if err != nil {
		return nil, err
	}
	var rows []model.GameContentVisibility
	err = db.Model(&model.GameContentVisibility{}).
		Where("site_id = ? AND resource_type = ? AND resource_id IN (?) AND deleted_at IS NULL", scope.SiteID, resource, ids).
		Find(&rows).Error
	for _, row := range rows {
		values[row.ResourceID] = row.Enabled
	}
	return values, err
}

func Visible(overrides map[uint64]bool, id uint64, sourceEnabled bool) bool {
	if !sourceEnabled {
		return false
	}
	if value, ok := overrides[id]; ok {
		return value
	}
	return true
}

func SetVisibility(tx *gorm.DB, scope content.Scope, resource string, resourceID uint64, enabled bool, operatorID uint64) error {
	if !scope.SiteScoped() {
		return fmt.Errorf("H5 scope is required")
	}
	resource, err := NormalizeResource(resource)
	if err != nil {
		return err
	}
	var count int64
	if err := tx.Table(resourceTables[resource]).
		Where("id = ? AND site_id = 0 AND deleted_at IS NULL", resourceID).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return gorm.ErrRecordNotFound
	}
	now := time.Now()
	row := map[string]any{
		"site_id": scope.SiteID, "resource_type": resource, "resource_id": resourceID,
		"enabled": enabled, "created_by": operatorID, "updated_by": operatorID, "created_at": now, "updated_at": now,
	}
	return tx.Table("game_content_visibility").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "site_id"}, {Name: "resource_type"}, {Name: "resource_id"}},
		DoUpdates: clause.Assignments(map[string]any{"enabled": enabled, "updated_by": operatorID, "updated_at": now, "deleted_at": nil}),
	}).Create(&row).Error
}

func IDs[T any](rows []T, id func(T) uint64) []uint64 {
	values := make([]uint64, 0, len(rows))
	for _, row := range rows {
		values = append(values, id(row))
	}
	return values
}

func Page[T any](rows []T, pageNumber int, pageSize int) ([]T, int64, int, int) {
	if pageNumber < 1 {
		pageNumber = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	total := int64(len(rows))
	start := (pageNumber - 1) * pageSize
	if start >= len(rows) {
		return []T{}, total, pageNumber, pageSize
	}
	end := start + pageSize
	if end > len(rows) {
		end = len(rows)
	}
	return rows[start:end], total, pageNumber, pageSize
}
