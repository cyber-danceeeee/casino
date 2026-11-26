package gameadmin

import (
	"encoding/json"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"casino/internal/tenancy"
)

type gameCategoryDraftItem struct {
	CategoryID uint64 `json:"categoryId"`
	Sort       int    `json:"sort"`
}

func gameDraftUpdate(actor Actor, values map[string]any) map[string]any {
	now := time.Now()
	values["config_status"] = "DRAFT"
	values["draft_updated_at"] = now
	values["draft_updated_by"] = actor.ID
	values["updated_at"] = now
	return values
}

func gamePublishedUpdate(actor Actor, values map[string]any) map[string]any {
	now := time.Now()
	values["config_status"] = "PUBLISHED"
	values["config_version"] = gorm.Expr("config_version + 1")
	values["config_published_at"] = now
	values["config_published_by"] = actor.ID
	values["draft_updated_at"] = nil
	values["draft_updated_by"] = 0
	values["updated_at"] = now
	return values
}

func (s *Service) loadPublishedGameCategories(tx *gorm.DB, actor Actor, gameID uint64) ([]gameCategoryDraftItem, error) {
	var rows []gameCategoryDraftItem
	if err := applyCategoryResourceScope(tx.Table("game_game_category_relations"), "", actor.Scope).
		Select("category_id category_id,sort").
		Where("game_id=? AND deleted_at IS NULL", gameID).
		Order("sort ASC,id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return s.normalizeGameCategoryDraft(tx, actor, rows)
}

func (s *Service) savePublishedGameCategories(tx *gorm.DB, actor Actor, gameID uint64, items []gameCategoryDraftItem) error {
	normalized, err := s.normalizeGameCategoryDraft(tx, actor, items)
	if err != nil {
		return err
	}
	ids := make([]uint64, 0, len(normalized))
	for _, item := range normalized {
		ids = append(ids, item.CategoryID)
	}
	if err := s.batchReplaceRelations(tx, actor, []uint64{gameID}, ids, "category"); err != nil {
		return err
	}
	encoded, err := json.Marshal(normalized)
	if err != nil {
		return newError("INVALID_REQUEST", "invalid category configuration")
	}
	return tenancy.ApplySite(tx.Table("game_games"), "", actor.Scope).
		Where("id=? AND deleted_at IS NULL", gameID).
		Updates(gamePublishedUpdate(actor, map[string]any{"category_ids_draft": string(encoded)})).Error
}

func (s *Service) saveGameCategoryDraft(tx *gorm.DB, actor Actor, gameID uint64, items []gameCategoryDraftItem) error {
	normalized, err := s.normalizeGameCategoryDraft(tx, actor, items)
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(normalized)
	if err != nil {
		return newError("INVALID_REQUEST", "invalid category draft")
	}
	return tenancy.ApplySite(tx.Table("game_games"), "", actor.Scope).
		Where("id=? AND deleted_at IS NULL", gameID).
		Updates(gameDraftUpdate(actor, map[string]any{"category_ids_draft": string(encoded)})).Error
}

func (s *Service) loadGameCategoryDraft(tx *gorm.DB, actor Actor, gameID uint64) ([]gameCategoryDraftItem, error) {
	var raw string
	if err := tenancy.ApplySite(tx.Table("game_games"), "", actor.Scope).
		Select("COALESCE(CAST(category_ids_draft AS CHAR),'')").
		Where("id=? AND deleted_at IS NULL", gameID).
		Scan(&raw).Error; err != nil {
		return nil, err
	}
	if strings.TrimSpace(raw) != "" {
		var items []gameCategoryDraftItem
		if err := json.Unmarshal([]byte(raw), &items); err != nil {
			return nil, newError("INVALID_CATEGORY_DRAFT", "invalid category draft")
		}
		return s.normalizeGameCategoryDraft(tx, actor, items)
	}
	var rows []gameCategoryDraftItem
	if err := applyCategoryResourceScope(tx.Table("game_game_category_relations"), "", actor.Scope).
		Select("category_id category_id,sort").
		Where("game_id=? AND deleted_at IS NULL", gameID).
		Order("sort ASC,id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return s.normalizeGameCategoryDraft(tx, actor, rows)
}

func (s *Service) normalizeGameCategoryDraft(tx *gorm.DB, actor Actor, items []gameCategoryDraftItem) ([]gameCategoryDraftItem, error) {
	seen := map[uint64]bool{}
	out := make([]gameCategoryDraftItem, 0, len(items))
	for _, item := range items {
		if item.CategoryID == 0 || seen[item.CategoryID] {
			continue
		}
		seen[item.CategoryID] = true
		out = append(out, item)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Sort == out[j].Sort {
			return out[i].CategoryID < out[j].CategoryID
		}
		return out[i].Sort < out[j].Sort
	})
	for index := range out {
		out[index].Sort = index
	}
	if len(out) == 0 {
		return out, nil
	}
	ids := make([]uint64, 0, len(out))
	for _, item := range out {
		ids = append(ids, item.CategoryID)
	}
	var count int64
	categoryDB, err := applyCategoryWriteScope(tx.Table("game_game_categories"), "", actor.Scope)
	if err != nil {
		return nil, err
	}
	if err := categoryDB.Where("id IN ? AND deleted_at IS NULL", ids).Count(&count).Error; err != nil {
		return nil, err
	}
	if count != int64(len(ids)) {
		return nil, newError("INVALID_RELATION", "one or more relation targets do not exist")
	}
	return out, nil
}

func categoryDraftFromIDs(ids []uint64) []gameCategoryDraftItem {
	unique := uniqueIDs(ids)
	out := make([]gameCategoryDraftItem, 0, len(unique))
	for index, id := range unique {
		out = append(out, gameCategoryDraftItem{CategoryID: id, Sort: index})
	}
	return out
}

func parseGameCategoryDraftItems(raw string) []gameCategoryDraftItem {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var items []gameCategoryDraftItem
	if json.Unmarshal([]byte(raw), &items) != nil {
		return nil
	}
	return items
}

func draftBool(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func draftInt(value *int, fallback int) int {
	if value == nil {
		return fallback
	}
	return *value
}

func draftString(value *string, fallback string) string {
	if value == nil {
		return fallback
	}
	return *value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}
