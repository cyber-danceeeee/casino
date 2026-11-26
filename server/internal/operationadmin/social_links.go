package operationadmin

import (
	"strings"
	"time"

	"gorm.io/gorm"

	"casino/internal/model"
	"casino/internal/operationdefaults"
	"casino/internal/tenancy"
)

type socialLinkRecord struct {
	operationdefaults.Metadata
	ID               uint64     `gorm:"column:id;primaryKey" json:"id"`
	TenantID         uint64     `gorm:"column:tenant_id" json:"tenantID"`
	SiteID           uint64     `gorm:"column:site_id" json:"siteID"`
	PlatformCode     string     `gorm:"column:platform_code" json:"platformCode"`
	Label            string     `gorm:"column:label" json:"label"`
	IconURL          string     `gorm:"column:icon_url" json:"iconURL"`
	ShareURL         string     `gorm:"column:share_url" json:"shareURL"`
	ShareTitle       string     `gorm:"column:share_title" json:"shareTitle"`
	ShareText        string     `gorm:"column:share_text" json:"shareText"`
	ShareDescription string     `gorm:"column:share_description" json:"shareDescription"`
	Sort             int        `gorm:"column:sort" json:"sort"`
	Status           int        `gorm:"column:status" json:"status"`
	Enabled          bool       `gorm:"-" json:"enabled"`
	CreatedAt        time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt        time.Time  `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt        *time.Time `gorm:"column:deleted_at" json:"-"`
}

func (socialLinkRecord) TableName() string { return "game_social_links" }

type SocialLinkInput struct {
	PlatformCode     string `json:"platformCode" binding:"required,max=64"`
	Label            string `json:"label" binding:"max=64"`
	IconURL          string `json:"iconURL" binding:"max=500"`
	ShareURL         string `json:"shareURL" binding:"required,max=500"`
	ShareTitle       string `json:"shareTitle" binding:"required,max=255"`
	ShareText        string `json:"shareText" binding:"required,max=500"`
	ShareDescription string `json:"shareDescription" binding:"max=500"`
	Sort             int    `json:"sort"`
	Enabled          bool   `json:"enabled"`
}

func (s *Service) ListSocialLinks(q ListQuery) (PageResult, error) {
	var rows []socialLinkRecord
	db := tenancy.ApplySite(s.db.Model(&socialLinkRecord{}), "", q.Scope).Where("deleted_at IS NULL")
	db = contains(db, "platform_code", q.Values["platformCode"])
	db = contains(db, "label", q.Values["label"])
	db = socialStatusFilter(db, q.Values["enabled"])
	result, err := page(db.Order("sort ASC, id DESC"), q, &rows)
	for i := range rows {
		rows[i].Enabled = rows[i].Status == 1
		rows[i].Metadata = operationdefaults.Metadata{CanEdit: true, CanDelete: !q.Scope.TenantScoped(), Visible: rows[i].Enabled}
		signSocialAssets(&rows[i])
	}
	result.List = rows
	return result, err
}

func (s *Service) GetSocialLink(id uint64, scope tenancy.Scope) (map[string]any, error) {
	var row socialLinkRecord
	if err := tenancy.ApplySite(s.db.Model(&socialLinkRecord{}), "", scope).Where("id = ? AND deleted_at IS NULL", id).First(&row).Error; err != nil {
		return nil, notFound(err, "SOCIAL_LINK_NOT_FOUND")
	}
	row.Enabled = row.Status == 1
	row.Metadata = operationdefaults.Metadata{CanEdit: true, CanDelete: !scope.TenantScoped(), Visible: row.Enabled}
	signSocialAssets(&row)
	logs, _ := s.operationLogs("social_link", id, scope)
	return map[string]any{"overview": row, "operationLogs": logs}, nil
}

func (s *Service) SaveSocialLink(id uint64, in SocialLinkInput, actor Actor) (uint64, error) {
	if actor.Scope.TenantScoped() && id == 0 {
		return 0, newError("SOCIAL_LINK_CREATE_DISABLED", "tenant social platforms are provisioned by the platform")
	}
	in.PlatformCode = strings.ToLower(strings.TrimSpace(in.PlatformCode))
	in.Label = strings.TrimSpace(in.Label)
	in.IconURL = strings.TrimSpace(in.IconURL)
	in.ShareURL = strings.TrimSpace(in.ShareURL)
	in.ShareTitle = strings.TrimSpace(in.ShareTitle)
	in.ShareText = strings.TrimSpace(in.ShareText)
	in.ShareDescription = strings.TrimSpace(in.ShareDescription)
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var before model.GameSocialLinks
		action, permission := "create", "operation:social-link:create"
		if id > 0 {
			if err := tenancy.ApplySite(tx.Model(&model.GameSocialLinks{}), "", actor.Scope).Where("id = ? AND deleted_at IS NULL", id).First(&before).Error; err != nil {
				return notFound(err, "SOCIAL_LINK_NOT_FOUND")
			}
			if actor.Scope.TenantScoped() {
				in.PlatformCode = before.PlatformCode
			}
			action, permission = "update", "operation:social-link:update"
		}
		label := in.Label
		if label == "" {
			if id > 0 {
				label = before.Label
			}
			if label == "" {
				label = in.PlatformCode
			}
		}
		icon := in.IconURL
		if id > 0 && icon == "" {
			icon = before.IconURL
		} else {
			icon = operationAssetRef(icon)
		}
		row := model.GameSocialLinks{TenantID: actor.Scope.TenantID, SiteID: actor.Scope.SiteID, PlatformCode: in.PlatformCode, Label: label, IconURL: icon, ShareURL: in.ShareURL, ShareTitle: in.ShareTitle, ShareText: in.ShareText, ShareDescription: in.ShareDescription, Sort: in.Sort, Status: operationStatus(in.Enabled)}
		if id == 0 {
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			id = row.ID
		} else if err := tenancy.ApplySite(tx.Model(&model.GameSocialLinks{}), "", actor.Scope).Where("id = ?", id).Select("platform_code", "label", "icon_url", "share_url", "share_title", "share_text", "share_description", "sort", "status").Updates(row).Error; err != nil {
			return err
		}
		if err := tenancy.ApplySite(tx.Model(&model.GameSocialLinks{}), "", actor.Scope).Where("id = ?", id).First(&row).Error; err != nil {
			return err
		}
		return audit(tx, actor, permission, action, "social_link", id, before, row, "")
	})
	return id, err
}

func (s *Service) DeleteSocialLink(id uint64, actor Actor) error {
	if actor.Scope.TenantScoped() {
		return newError("SOCIAL_LINK_DELETE_DISABLED", "tenant social platforms cannot be deleted")
	}
	return s.softDelete("game_social_links", "social_link", id, "operation:social-link:delete", actor)
}

func socialStatusFilter(db *gorm.DB, value string) *gorm.DB {
	if value == "true" || value == "1" {
		return db.Where("status = 1")
	}
	if value == "false" || value == "0" {
		return db.Where("status = 2")
	}
	return db
}
