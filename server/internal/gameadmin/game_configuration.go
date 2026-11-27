package gameadmin

import (
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"casino/internal/content"
)

type gameCategoryItem struct {
	CategoryID uint64 `json:"categoryId"`
	Sort       int    `json:"sort"`
}

func gameConfigurationUpdate(actor Actor, values map[string]any) map[string]any {
	values["config_version"] = gorm.Expr("config_version+1")
	values["updated_at"] = time.Now()
	return values
}

func (s *Service) loadGameCategories(tx *gorm.DB, actor Actor, gameID uint64) ([]gameCategoryItem, error) {
	var rows []gameCategoryItem
	if err := applyCategoryResourceScope(tx.Table("game_game_category_relations"), "", actor.Scope).
		Select("category_id,category_priority AS sort").
		Where("game_id=? AND deleted_at IS NULL", gameID).
		Order("category_priority ASC,id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return s.normalizeGameCategories(tx, actor, rows)
}

func (s *Service) saveGameCategories(tx *gorm.DB, actor Actor, gameID uint64, items []gameCategoryItem) error {
	normalized, err := s.normalizeGameCategories(tx, actor, items)
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
	return content.ApplySite(tx.Table("game_games"), "", actor.Scope).Where("id=? AND deleted_at IS NULL", gameID).Updates(gameConfigurationUpdate(actor, map[string]any{})).Error
}

func (s *Service) normalizeGameCategories(tx *gorm.DB, actor Actor, items []gameCategoryItem) ([]gameCategoryItem, error) {
	seen := map[uint64]bool{}
	out := make([]gameCategoryItem, 0, len(items))
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

func categoriesFromIDs(ids []uint64) []gameCategoryItem {
	unique := uniqueIDs(ids)
	out := make([]gameCategoryItem, 0, len(unique))
	for index, id := range unique {
		out = append(out, gameCategoryItem{CategoryID: id, Sort: index})
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}
