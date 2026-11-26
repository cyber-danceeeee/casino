package systemadmin

import (
	"encoding/json"

	"gorm.io/gorm"
)

type contentPublishRow struct {
	ID           uint64 `gorm:"column:id"`
	DraftPayload string `gorm:"column:draft_payload"`
}

func previewAnnouncementDrafts(db *gorm.DB, siteID uint64) ([]map[string]any, error) {
	return previewContentDrafts(db, "game_site_announcements", siteID)
}

func previewPopupDrafts(db *gorm.DB, siteID uint64) ([]map[string]any, error) {
	return previewContentDrafts(db, "game_popup_campaigns", siteID)
}

func previewContentDrafts(db *gorm.DB, table string, siteID uint64) ([]map[string]any, error) {
	var rows []contentPublishRow
	if err := db.Table(table).Select("id,COALESCE(CAST(draft_payload AS CHAR),'') draft_payload").
		Where("site_id = ? AND deleted_at IS NULL", siteID).Order("id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		payload := map[string]any{}
		if err := json.Unmarshal([]byte(row.DraftPayload), &payload); err != nil {
			return nil, NewError("INVALID_CONTENT_DRAFT", "invalid site content draft")
		}
		payload["id"] = row.ID
		out = append(out, payload)
	}
	return out, nil
}

func publishAnnouncementDrafts(tx *gorm.DB, siteID uint64, actorID uint64, configVersion uint64) (int, error) {
	var rows []contentPublishRow
	if err := tx.Table("game_site_announcements").Select("id,COALESCE(CAST(draft_payload AS CHAR),'') draft_payload").
		Where("site_id = ? AND publish_status = ? AND deleted_at IS NULL", siteID, siteContentDraft).Order("id ASC").Find(&rows).Error; err != nil {
		return 0, err
	}
	nowValue := now()
	for _, row := range rows {
		var raw AnnouncementDraft
		if err := json.Unmarshal([]byte(row.DraftPayload), &raw); err != nil {
			return 0, NewError("INVALID_ANNOUNCEMENT_DRAFT", "invalid announcement draft")
		}
		if raw.Deleted {
			if err := tx.Table("game_site_announcements").Where("id = ? AND site_id = ?", row.ID, siteID).
				Updates(map[string]any{"deleted_at": nowValue, "updated_at": nowValue, "updated_by": actorID}).Error; err != nil {
				return 0, err
			}
			continue
		}
		draft, err := normalizeAnnouncementDraft(raw)
		if err != nil {
			return 0, err
		}
		payload := marshal(draft)
		updates := map[string]any{
			"title": draft.Title, "summary": draft.Summary, "content": draft.Content, "announcement_type": draft.AnnouncementType,
			"priority": draft.Priority, "sort_order": draft.SortOrder, "is_pinned": boolValue(draft.IsPinned), "language_code": draft.LanguageCode,
			"country_code": draft.CountryCode, "currency_code": draft.CurrencyCode, "visibility_scope": draft.VisibilityScope,
			"start_at": draft.StartAt, "end_at": draft.EndAt, "enabled": boolValue(draft.Enabled), "draft_payload": payload,
			"published_payload": payload, "publish_status": siteContentPublished, "config_version": configVersion,
			"published_at": nowValue, "published_by": actorID, "updated_by": actorID, "updated_at": nowValue,
		}
		if err := tx.Table("game_site_announcements").Where("id = ? AND site_id = ? AND deleted_at IS NULL", row.ID, siteID).Updates(updates).Error; err != nil {
			return 0, err
		}
	}
	return len(rows), nil
}

func publishPopupDrafts(tx *gorm.DB, siteID uint64, actorID uint64, configVersion uint64) (int, error) {
	var rows []contentPublishRow
	if err := tx.Table("game_popup_campaigns").Select("id,COALESCE(CAST(draft_payload AS CHAR),'') draft_payload").
		Where("site_id = ? AND publish_status = ? AND deleted_at IS NULL", siteID, siteContentDraft).Order("id ASC").Find(&rows).Error; err != nil {
		return 0, err
	}
	nowValue := now()
	service := New(tx)
	for _, row := range rows {
		var raw PopupDraft
		if err := json.Unmarshal([]byte(row.DraftPayload), &raw); err != nil {
			return 0, NewError("INVALID_POPUP_DRAFT", "invalid popup draft")
		}
		if raw.Deleted {
			if err := tx.Table("game_popup_campaigns").Where("id = ? AND site_id = ?", row.ID, siteID).
				Updates(map[string]any{"deleted_at": nowValue, "updated_at": nowValue, "updated_by": actorID}).Error; err != nil {
				return 0, err
			}
			continue
		}
		draft, err := service.normalizePopupDraft(raw)
		if err != nil {
			return 0, err
		}
		payload := marshal(draft)
		updates := map[string]any{
			"popup_component_id": draft.PopupComponentID, "popup_component_id_draft": draft.PopupComponentID,
			"campaign_name": draft.CampaignName, "action_type": draft.ActionType, "action_target": draft.ActionTarget,
			"trigger_type": draft.TriggerType, "trigger_value": draft.TriggerValue,
			"frequency_type": draft.FrequencyType, "display_limit": draft.DisplayLimit, "cooldown_minutes": draft.CooldownMinutes,
			"visibility_scope": draft.VisibilityScope, "start_at": draft.StartAt, "end_at": draft.EndAt,
			"priority": draft.Priority, "enabled": boolValue(draft.Enabled), "draft_payload": payload,
			"published_payload": payload, "publish_status": siteContentPublished, "config_version": configVersion,
			"published_at": nowValue, "published_by": actorID, "updated_by": actorID, "updated_at": nowValue,
		}
		if err := tx.Table("game_popup_campaigns").Where("id = ? AND site_id = ? AND deleted_at IS NULL", row.ID, siteID).Updates(updates).Error; err != nil {
			return 0, err
		}
	}
	return len(rows), nil
}
