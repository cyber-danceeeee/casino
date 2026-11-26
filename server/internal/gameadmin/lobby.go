package gameadmin

import (
	"encoding/json"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/model"
	"casino/internal/tenancy"
)

func (s *Service) ListLobbySections(q ListQuery) (PageResult, error) {
	var rows []model.GameLobbySections
	db := tenancy.ApplySite(s.db.Model(&model.GameLobbySections{}), "", q.Scope).
		Where("deleted_at IS NULL")
	db = contains(db, "section_code", q.Values["sectionCode"])
	db = equal(db, "section_type", q.Values["sectionType"])
	if enabled := strings.ToLower(strings.TrimSpace(q.Values["enabled"])); enabled != "" {
		if enabled == "true" || enabled == "1" {
			db = db.Where("enabled = 1")
		} else if enabled == "false" || enabled == "0" {
			db = db.Where("enabled = 0")
		}
	}
	return paginate(db.Order("sort_order ASC,id ASC"), q, &rows)
}

func (s *Service) LobbySectionDetail(id uint64, scope tenancy.Scope) (model.GameLobbySections, error) {
	var row model.GameLobbySections
	err := tenancy.ApplySite(s.db.Model(&model.GameLobbySections{}), "", scope).
		Where("id = ? AND deleted_at IS NULL", id).
		Take(&row).Error
	if err != nil {
		return row, notFound(err, "LOBBY_SECTION_NOT_FOUND")
	}
	return row, nil
}

func (s *Service) SaveLobbySection(id uint64, in LobbySectionInput, actor Actor) (uint64, error) {
	code := normalizeCode(in.SectionCode)
	if err := validateCode(code); err != nil {
		return 0, err
	}
	if id > 0 {
		var existing model.GameLobbySections
		if err := tenancy.ApplySite(s.db.Model(&model.GameLobbySections{}), "", actor.Scope).
			Where("id = ? AND deleted_at IS NULL", id).Take(&existing).Error; err != nil {
			return 0, notFound(err, "LOBBY_SECTION_NOT_FOUND")
		}
	}
	if actor.Scope.SiteScoped() {
		var dimensions struct {
			CountryCode  string `gorm:"column:country_code"`
			CurrencyCode string `gorm:"column:currency_code"`
		}
		if err := s.db.Table("game_saas_sites").Select("country_code,currency_code").
			Where("id=? AND tenant_id=? AND deleted_at IS NULL", actor.Scope.SiteID, actor.Scope.TenantID).
			Take(&dimensions).Error; err != nil {
			return 0, newError("SITE_NOT_FOUND", "tenant site was not found")
		}
		in.SupportedCountries = strings.ToUpper(strings.TrimSpace(dimensions.CountryCode))
		in.SupportedCurrencies = strings.ToUpper(strings.TrimSpace(dimensions.CurrencyCode))
	}
	draftPayload, err := lobbySectionDraftPayload(code, in)
	if err != nil {
		return 0, err
	}
	configJSON := lobbySectionConfigJSON(in)
	if in.MaxItems <= 0 {
		in.MaxItems = 12
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var before model.GameLobbySections
		action := "create"
		if id > 0 {
			action = "update"
			if err := tenancy.ApplySite(tx.Model(&model.GameLobbySections{}).Clauses(clause.Locking{Strength: "UPDATE"}), "", actor.Scope).
				Where("id = ? AND deleted_at IS NULL", id).
				Take(&before).Error; err != nil {
				return notFound(err, "LOBBY_SECTION_NOT_FOUND")
			}
		}
		row := model.GameLobbySections{
			TenantID:            actor.Scope.TenantID,
			SiteID:              actor.Scope.SiteID,
			SectionCode:         code,
			SectionType:         strings.TrimSpace(in.SectionType),
			TitleZh:             strings.TrimSpace(in.TitleZh),
			TitleEn:             strings.TrimSpace(in.TitleEn),
			DisplayStyle:        strings.TrimSpace(in.DisplayStyle),
			SourceType:          strings.TrimSpace(in.SourceType),
			SourceRef:           strings.TrimSpace(in.SourceRef),
			MaxItems:            in.MaxItems,
			SortOrder:           in.SortOrder,
			Enabled:             in.Enabled,
			SupportedCountries:  strings.TrimSpace(in.SupportedCountries),
			SupportedCurrencies: strings.TrimSpace(in.SupportedCurrencies),
			ConfigJSON:          configJSON,
			DraftPayload:        draftPayload,
			PublishStatus:       "DRAFT",
		}
		if id == 0 {
			if err := tx.Create(&row).Error; err != nil {
				return duplicateCode(err, "LOBBY_SECTION_CODE_EXISTS", "lobby section code already exists")
			}
			id = row.ID
		} else {
			if err := tenancy.ApplySite(tx.Model(&model.GameLobbySections{}), "", actor.Scope).Where("id = ?", id).Updates(map[string]any{
				"draft_payload":    draftPayload,
				"publish_status":   "DRAFT",
				"draft_updated_at": time.Now(),
				"draft_updated_by": actor.ID,
				"updated_at":       time.Now(),
			}).Error; err != nil {
				return duplicateCode(err, "LOBBY_SECTION_CODE_EXISTS", "lobby section code already exists")
			}
		}
		return audit(tx, actor, "lobby:update", action+"_draft", "game_lobby_section", id, before, row, in.Reason)
	})
	return id, err
}

func lobbySectionConfigJSON(in LobbySectionInput) string {
	if in.Config == nil {
		return "{}"
	}
	encoded, err := json.Marshal(in.Config)
	if err != nil {
		return "{}"
	}
	return string(encoded)
}

func lobbySectionDraftPayload(code string, in LobbySectionInput) (string, error) {
	if in.Config == nil {
		in.Config = map[string]any{}
	}
	if in.MaxItems <= 0 {
		in.MaxItems = 12
	}
	payload := map[string]any{
		"sectionCode":         code,
		"sectionType":         strings.TrimSpace(in.SectionType),
		"titleZh":             strings.TrimSpace(in.TitleZh),
		"titleEn":             strings.TrimSpace(in.TitleEn),
		"displayStyle":        strings.TrimSpace(in.DisplayStyle),
		"sourceType":          strings.TrimSpace(in.SourceType),
		"sourceRef":           strings.TrimSpace(in.SourceRef),
		"maxItems":            in.MaxItems,
		"sortOrder":           in.SortOrder,
		"enabled":             in.Enabled,
		"supportedCountries":  strings.TrimSpace(in.SupportedCountries),
		"supportedCurrencies": strings.TrimSpace(in.SupportedCurrencies),
		"config":              in.Config,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", newError("INVALID_REQUEST", "invalid section config")
	}
	return string(encoded), nil
}

func (s *Service) DeleteLobbySection(id uint64, reason string, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var before model.GameLobbySections
		if err := tenancy.ApplySite(tx.Model(&model.GameLobbySections{}).Clauses(clause.Locking{Strength: "UPDATE"}), "", actor.Scope).
			Where("id = ? AND deleted_at IS NULL", id).
			Take(&before).Error; err != nil {
			return notFound(err, "LOBBY_SECTION_NOT_FOUND")
		}
		now := time.Now()
		if err := tenancy.ApplySite(tx.Model(&model.GameLobbySections{}), "", actor.Scope).
			Where("id = ?", id).
			Updates(map[string]any{"deleted_at": now, "updated_at": now}).Error; err != nil {
			return err
		}
		return audit(tx, actor, "lobby:update", "delete", "game_lobby_section", id, before, map[string]any{"deleted": true}, reason)
	})
}
