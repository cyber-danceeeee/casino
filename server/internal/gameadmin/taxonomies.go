package gameadmin

import (
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/content"
	"casino/internal/model"
)

func (s *Service) ListCategories(q ListQuery) (PageResult, error) {
	var rows []categoryRecord
	db := s.db.Table("game_game_categories c").Select(`c.*,COUNT(r.id) game_count`).Joins(categoryRelationCountJoin(q.Scope)).Where("c.deleted_at IS NULL").Group("c.id")
	db = applyCategoryReadableScope(db, "c", q.Scope)
	db = contains(db, "c.category_key", q.Values["code"])
	db = contains(db, "c.name_zh", q.Values["nameZh"])
	db = contains(db, "c.name_en", q.Values["nameEn"])
	if enabled := q.Values["enabled"]; enabled != "" {
		if enabled == "true" {
			db = db.Where("c.status=1")
		} else {
			db = db.Where("c.status<>1")
		}
	}
	result, err := paginate(db.Order("c.sort ASC,c.id ASC"), q, &rows)
	for i := range rows {
		rows[i].Enabled = rows[i].Status == 1
		signCategoryAsset(&rows[i])
	}
	if q.Scope.SiteScoped() && q.Scope.SiteID != 1 {
		result.List = h5CategoryList(rows)
	} else {
		result.List = rows
	}
	return result, err
}

func applyCategoryReadableScope(db *gorm.DB, alias string, scope content.Scope) *gorm.DB {
	prefix := ""
	if strings.TrimSpace(alias) != "" {
		prefix = alias + "."
	}
	if scope.Platform {
		return db.Where(prefix + "site_id=0")
	}
	if !scope.SiteScoped() {
		return db.Where("1=0")
	}
	if scope.SiteID == 1 {
		return content.ApplySite(db, alias, scope)
	}
	return content.ApplySiteStrict(db, alias, scope)
}

func categoryRelationCountJoin(scope content.Scope) string {
	if scope.Platform {
		return "LEFT JOIN game_game_category_relations AS r ON r.category_id = c.id AND r.site_id = 0 AND r.deleted_at IS NULL"
	}
	if !scope.SiteScoped() {
		return "LEFT JOIN game_game_category_relations r ON 1=0"
	}
	join := "LEFT JOIN game_game_category_relations r ON r.category_id=c.id AND r.deleted_at IS NULL"
	if scope.SiteID > 0 {
		if scope.SiteID == 1 {
			join += " AND r.site_id IN (0,1)"
		} else {
			join += " AND r.site_id=" + uintString(scope.SiteID)
		}
	}
	return join
}

func uintString(value uint64) string {
	return strconv.FormatUint(value, 10)
}

func applyCategoryResourceScope(db *gorm.DB, alias string, scope content.Scope) *gorm.DB {
	if scope.Platform {
		prefix := ""
		if strings.TrimSpace(alias) != "" {
			prefix = alias + "."
		}
		return db.Where(prefix + "site_id=0")
	}
	if scope.SiteID == 1 {
		return content.ApplySite(db, alias, scope)
	}
	return content.ApplySiteStrict(db, alias, scope)
}

func applyCategoryWriteScope(db *gorm.DB, alias string, scope content.Scope) (*gorm.DB, error) {
	prefix := ""
	if strings.TrimSpace(alias) != "" {
		prefix = alias + "."
	}
	if scope.Platform {
		return db.Where(prefix + "site_id=0"), nil
	}
	if scope.SiteID == 0 {
		return nil, newError("SITE_REQUIRED", "H5 scope is required")
	}
	if scope.SiteID == 1 {
		return content.ApplySite(db, alias, scope), nil
	}
	return content.ApplySiteStrict(db, alias, scope), nil
}

func categoryWriteScopeValues(scope content.Scope) (uint64, error) {
	if scope.Platform {
		return 0, nil
	}
	if scope.SiteID == 0 {
		return 0, newError("SITE_REQUIRED", "H5 configuration is required")
	}
	return scope.SiteID, nil
}

func ensureCategoryWritable(tx *gorm.DB, id uint64, scope content.Scope) error {
	categoryDB, err := applyCategoryWriteScope(tx.Table("game_game_categories"), "", scope)
	if err != nil {
		return err
	}
	var count int64
	if err := categoryDB.Where("id=? AND deleted_at IS NULL", id).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return notFound(gorm.ErrRecordNotFound, "CATEGORY_NOT_FOUND")
	}
	return nil
}

type h5CategoryListItem struct {
	SiteID    uint64    `json:"siteID"`
	ID        uint64    `json:"id"`
	NameEn    string    `json:"nameEn"`
	Icon      string    `json:"icon"`
	BannerURL string    `json:"bannerUrl"`
	SortOrder int       `json:"sortOrder"`
	Enabled   bool      `json:"enabled"`
	GameCount int64     `json:"gameCount"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func h5CategoryList(rows []categoryRecord) []h5CategoryListItem {
	items := make([]h5CategoryListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, h5CategoryListItem{
			SiteID: row.SiteID, ID: row.ID, NameEn: row.NameEn,
			Icon: row.Icon, BannerURL: row.BannerURL, SortOrder: row.SortOrder, Enabled: row.Enabled,
			GameCount: row.GameCount, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		})
	}
	return items
}
func (s *Service) CategoryDetail(id uint64, scope content.Scope) (map[string]any, error) {
	var row categoryRecord
	db := s.db.Table("game_game_categories c").Select("c.*,COUNT(r.id) game_count").Joins(categoryRelationCountJoin(scope)).Where("c.id=? AND c.deleted_at IS NULL", id).Group("c.id")
	db = applyCategoryReadableScope(db, "c", scope)
	if err := db.Take(&row).Error; err != nil {
		return nil, notFound(err, "CATEGORY_NOT_FOUND")
	}
	row.Enabled = row.Status == 1
	signCategoryAsset(&row)
	return map[string]any{"overview": row}, nil
}

func (s *Service) CategoryGames(id uint64, q ListQuery) (PageResult, error) {
	if err := s.ensureCategoryExists(id, q.Scope); err != nil {
		return PageResult{}, err
	}
	var games []gameRecord
	db := s.db.Table("game_games g").
		Select("g.*,p.provider_code,r.sort AS category_sort_order").
		Joins("JOIN game_game_providers p ON p.id=g.provider_id AND p.deleted_at IS NULL").
		Joins("JOIN game_game_category_relations r ON r.game_id=g.id AND r.deleted_at IS NULL").
		Where("r.category_id=? AND g.deleted_at IS NULL", id)
	db = applyCategoryResourceScope(db, "g", q.Scope)
	db = applyCategoryResourceScope(db, "p", q.Scope)
	db = applyCategoryReadableScope(db, "r", q.Scope)
	db = contains(db, "g.name", q.Values["keyword"])
	db = db.Order("r.sort ASC,r.id ASC")
	result, err := paginate(db, q, &games)
	for i := range games {
		signGameAsset(&games[i])
	}
	result.List = games
	return result, err
}

func (s *Service) CategoryGameCandidates(id uint64, q ListQuery) (PageResult, error) {
	if err := s.ensureCategoryExists(id, q.Scope); err != nil {
		return PageResult{}, err
	}
	providerCode := strings.TrimSpace(q.Values["providerCode"])
	source := strings.TrimSpace(q.Values["source"])
	categoryCode := strings.TrimSpace(q.Values["categoryCode"])
	if providerCode == "" || categoryCode == "" || (source != "providerCategory" && source != "gameType") {
		return PageResult{}, newError("INVALID_REQUEST", "provider and provider category are required")
	}

	var games []gameRecord
	db := s.db.Table("game_games g").
		Select("g.id,g.game_code,g.name,g.slug,g.cover_url,g.local_thumbnail_url,p.provider_code,g.provider_game_code,g.platform_enabled").
		Joins("JOIN game_game_providers p ON p.id=g.provider_id AND p.deleted_at IS NULL").
		Where("g.deleted_at IS NULL AND p.provider_code=?", providerCode).
		Where("NOT EXISTS (SELECT 1 FROM game_game_category_relations r WHERE r.game_id=g.id AND r.category_id=? AND r.deleted_at IS NULL AND "+categoryRelationScopePredicate("r", q.Scope)+")", id)
	db = applyCategoryResourceScope(db, "g", q.Scope)
	db = applyCategoryResourceScope(db, "p", q.Scope)
	if source == "gameType" {
		db = db.Where("g.game_type=?", categoryCode)
	} else {
		db = db.Where("g.category=?", categoryCode)
	}
	if keyword := strings.TrimSpace(q.Values["keyword"]); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("g.name LIKE ? OR g.game_code LIKE ? OR g.provider_game_code LIKE ?", like, like, like)
	}
	result, err := paginate(db.Order("g.name ASC,g.id ASC"), q, &games)
	for i := range games {
		signGameAsset(&games[i])
	}
	result.List = games
	return result, err
}

func (s *Service) CategoryOperationLogs(id uint64, q ListQuery) (PageResult, error) {
	if err := s.ensureCategoryExists(id, q.Scope); err != nil {
		return PageResult{}, err
	}
	var rows []model.GameAdminOperationLogs
	db := content.ApplySite(s.db.Model(&model.GameAdminOperationLogs{}), "", q.Scope).
		Where("target_type=? AND target_id=?", "game_category", id).
		Order("id DESC")
	return paginate(db, q, &rows)
}

func (s *Service) ensureCategoryExists(id uint64, scope content.Scope) error {
	var count int64
	if err := applyCategoryReadableScope(s.db.Table("game_game_categories"), "", scope).Where("id=? AND deleted_at IS NULL", id).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return newError("CATEGORY_NOT_FOUND", "category not found")
	}
	return nil
}

func categoryRelationScopePredicate(alias string, scope content.Scope) string {
	prefix := ""
	if strings.TrimSpace(alias) != "" {
		prefix = alias + "."
	}
	if scope.Platform {
		return prefix + "site_id=0"
	}
	if !scope.SiteScoped() {
		return "1=0"
	}
	predicate := "1=1"
	if scope.SiteID > 0 {
		if scope.SiteID == 1 {
			predicate += " AND " + prefix + "site_id IN (0,1)"
		} else {
			predicate += " AND " + prefix + "site_id=" + uintString(scope.SiteID)
		}
	}
	return predicate
}
func (s *Service) SaveCategory(id uint64, in CategoryInput, actor Actor) (uint64, error) {
	in.Code = normalizeCode(in.Code)
	if err := validateCode(in.Code); err != nil {
		return 0, err
	}
	siteID, err := categoryWriteScopeValues(actor.Scope)
	if err != nil {
		return 0, err
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if actor.Scope.SiteID == 1 {
			var exists int64
			if err := tx.Table("game_game_categories").Where("site_id IN (0,1) AND category_key=? AND id<>? AND deleted_at IS NULL", in.Code, id).Count(&exists).Error; err != nil {
				return err
			}
			if exists > 0 {
				return newError("CATEGORY_CODE_EXISTS", "category code already exists")
			}
		}
		var before categoryRecord
		action, permission := "create", "game:category:create"
		if id > 0 {
			categoryDB, err := applyCategoryWriteScope(tx.Model(&categoryRecord{}), "", actor.Scope)
			if err != nil {
				return err
			}
			if err := categoryDB.Where("id=? AND deleted_at IS NULL", id).Take(&before).Error; err != nil {
				return notFound(err, "CATEGORY_NOT_FOUND")
			}
			action, permission = "update", "game:category:update"
		}
		shortLabel := in.Code
		if len(shortLabel) > 16 {
			shortLabel = shortLabel[:16]
		}
		icon := adminAssetRef(in.Icon)
		bannerURL := adminAssetRef(in.BannerURL)
		row := categoryRecord{SiteID: siteID, ID: id, Code: in.Code, Label: in.NameEn, ShortLabel: shortLabel, NameZh: in.NameZh, NameEn: in.NameEn, Icon: icon, BannerURL: bannerURL, SortOrder: in.SortOrder, Status: statusInt(in.Enabled)}
		if id == 0 {
			if err := tx.Create(&row).Error; err != nil {
				return duplicateCode(err, "CATEGORY_CODE_EXISTS", "category code already exists")
			}
			id = row.ID
		} else {
			categoryDB, err := applyCategoryWriteScope(tx.Model(&categoryRecord{}), "", actor.Scope)
			if err != nil {
				return err
			}
			if err := categoryDB.Where("id=?", id).Updates(map[string]any{"category_key": row.Code, "label": row.Label, "short_label": row.ShortLabel, "name_zh": row.NameZh, "name_en": row.NameEn, "icon_url": row.Icon, "banner_url": row.BannerURL, "sort": row.SortOrder, "status": row.Status}).Error; err != nil {
				return err
			}
		}
		return audit(tx, actor, permission, action, "game_category", id, before, row, in.Reason)
	})
	return id, err
}
func (s *Service) DeleteCategory(id uint64, reason string, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var before categoryRecord
		categoryDB, err := applyCategoryWriteScope(tx.Table("game_game_categories"), "", actor.Scope)
		if err != nil {
			return err
		}
		if err := categoryDB.Where("id=? AND deleted_at IS NULL", id).Take(&before).Error; err != nil {
			return notFound(err, "CATEGORY_NOT_FOUND")
		}

		var affectedGameIDs []uint64
		relationDB, err := applyCategoryWriteScope(tx.Table("game_game_category_relations"), "", actor.Scope)
		if err != nil {
			return err
		}
		if err := relationDB.
			Where("category_id=? AND deleted_at IS NULL", id).
			Pluck("game_id", &affectedGameIDs).Error; err != nil {
			return err
		}

		now := time.Now()
		relationDB, err = applyCategoryWriteScope(tx.Table("game_game_category_relations"), "", actor.Scope)
		if err != nil {
			return err
		}
		if err := relationDB.
			Where("category_id=? AND deleted_at IS NULL", id).
			Updates(map[string]any{"deleted_at": now, "updated_at": now}).Error; err != nil {
			return err
		}
		mappingDB, err := applyCategoryWriteScope(tx.Table("game_provider_category_mapping"), "", actor.Scope)
		if err != nil {
			return err
		}
		if err := mappingDB.
			Where("platform_category_id=? AND deleted_at IS NULL", id).
			Updates(map[string]any{"deleted_at": now, "updated_at": now}).Error; err != nil {
			return err
		}

		if len(affectedGameIDs) > 0 {
			if err := tx.Exec(`
				UPDATE game_games AS g
				SET
					platform_category_id = COALESCE((
						SELECT r.category_id
						FROM game_game_category_relations AS r
						JOIN game_game_categories AS c ON c.id = r.category_id AND c.deleted_at IS NULL
						WHERE r.game_id = g.id AND r.deleted_at IS NULL
						ORDER BY r.sort ASC, r.id ASC
						LIMIT 1
					), 0),
					platform_category_code = COALESCE((
						SELECT c.category_key
						FROM game_game_category_relations AS r
						JOIN game_game_categories AS c ON c.id = r.category_id AND c.deleted_at IS NULL
						WHERE r.game_id = g.id AND r.deleted_at IS NULL
						ORDER BY r.sort ASC, r.id ASC
						LIMIT 1
					), ''),
					updated_at = ?
				WHERE g.id IN ? AND g.deleted_at IS NULL
			`, now, affectedGameIDs).Error; err != nil {
				return err
			}
		}

		categoryDB, err = applyCategoryWriteScope(tx.Table("game_game_categories"), "", actor.Scope)
		if err != nil {
			return err
		}
		if err := categoryDB.Where("id=?", id).
			Updates(map[string]any{"deleted_at": now, "updated_at": now}).Error; err != nil {
			return err
		}
		return audit(tx, actor, "game:category:delete", "delete", "game_category", id, before, map[string]any{
			"deleted":           true,
			"detachedGameCount": len(affectedGameIDs),
		}, reason)
	})
}
func (s *Service) AddCategoryGames(id uint64, in CategoryGamesAddInput, actor Actor) error {
	gameIDs := uniqueIDs(in.IDs)
	if len(gameIDs) == 0 {
		return newError("INVALID_REQUEST", "game ids are required")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var category categoryRecord
		categoryDB, err := applyCategoryWriteScope(tx.Table("game_game_categories"), "", actor.Scope)
		if err != nil {
			return err
		}
		if err := categoryDB.Where("id=? AND deleted_at IS NULL", id).Take(&category).Error; err != nil {
			return notFound(err, "CATEGORY_NOT_FOUND")
		}
		var count int64
		if err := content.ApplySite(tx.Table("game_games"), "", actor.Scope).Where("id IN ? AND deleted_at IS NULL", gameIDs).Count(&count).Error; err != nil {
			return err
		}
		if count != int64(len(gameIDs)) {
			return newError("INVALID_RELATION", "one or more games do not exist")
		}
		before := map[uint64][]gameCategoryItem{}
		after := map[uint64][]gameCategoryItem{}
		for index, gameID := range gameIDs {
			items, err := s.loadGameCategories(tx, actor, gameID)

			if err != nil {
				return err
			}
			before[gameID] = items
			items = append(items, gameCategoryItem{CategoryID: id, Sort: len(items) + index})

			err = s.saveGameCategories(tx, actor, gameID, items)

			if err != nil {
				return err
			}

			after[gameID], _ = s.loadGameCategories(tx, actor, gameID)

		}
		action := "add_games"
		resultKey := "configuration"

		action = "add_games"
		resultKey = "configuration"

		return audit(tx, actor, "game:category:assign", action, "game_category", id, before, map[string]any{"addedGameIDs": gameIDs, "categoryCode": category.Code, resultKey: after}, fallbackReason(in.Reason, "add_category_games"))
	})
}

func (s *Service) UpdateCategoryGameSort(id uint64, gameID uint64, in CategoryGameSortInput, actor Actor) error {
	if in.SortOrder < 0 {
		return newError("INVALID_REQUEST", "sort order cannot be negative")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := ensureCategoryWritable(tx, id, actor.Scope); err != nil {
			return err
		}
		var before struct {
			ID   uint64
			Sort int
		}
		query := applyCategoryResourceScope(tx.Table("game_game_category_relations"), "", actor.Scope).Where("category_id=? AND game_id=? AND deleted_at IS NULL", id, gameID)
		if err := query.Take(&before).Error; err != nil {
			return notFound(err, "CATEGORY_GAME_NOT_FOUND")
		}
		if err := tx.Table("game_game_category_relations").Where("id=?", before.ID).Updates(map[string]any{"sort": in.SortOrder, "updated_at": time.Now()}).Error; err != nil {
			return err
		}
		return audit(tx, actor, "game:category:assign", "sort_game", "game_category", id, before, map[string]any{"gameID": gameID, "sortOrder": in.SortOrder}, fallbackReason(in.Reason, "sort_category_game"))
	})
}

func (s *Service) RemoveCategoryGame(id uint64, gameID uint64, reason string, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := ensureCategoryWritable(tx, id, actor.Scope); err != nil {
			return err
		}
		items, err := s.loadGameCategories(tx, actor, gameID)

		if err != nil {
			return err
		}
		before := append([]gameCategoryItem(nil), items...)
		next := make([]gameCategoryItem, 0, len(items))
		found := false
		for _, item := range items {
			if item.CategoryID == id {
				found = true
				continue
			}
			next = append(next, item)
		}
		if !found {
			return notFound(gorm.ErrRecordNotFound, "CATEGORY_GAME_NOT_FOUND")
		}

		err = s.saveGameCategories(tx, actor, gameID, next)

		if err != nil {
			return err
		}
		after, _ := s.loadGameCategories(tx, actor, gameID)
		action := "remove_game"

		after, _ = s.loadGameCategories(tx, actor, gameID)
		action = "remove_game"

		return audit(tx, actor, "game:category:assign", action, "game_category", id, before, after, fallbackReason(reason, "remove_category_game"))
	})
}

func syncGamePrimaryCategory(tx *gorm.DB, gameID uint64) error {
	var primary struct {
		ID   uint64 `gorm:"column:id"`
		Code string `gorm:"column:code"`
	}
	result := tx.Table("game_game_category_relations r").
		Select("c.id,c.category_key code").
		Joins("JOIN game_game_categories c ON c.id=r.category_id AND c.deleted_at IS NULL").
		Where("r.game_id=? AND r.deleted_at IS NULL", gameID).
		Order("r.category_priority ASC,r.id ASC").Limit(1).Find(&primary)
	if result.Error != nil {
		return result.Error
	}
	return tx.Table("game_games").Where("id=? AND deleted_at IS NULL", gameID).Updates(map[string]any{
		"platform_category_id": primary.ID, "platform_category_code": primary.Code, "updated_at": time.Now(),
	}).Error
}

func fallbackReason(reason string, fallback string) string {
	if reason = strings.TrimSpace(reason); reason != "" {
		return reason
	}
	return fallback
}

func (s *Service) ListTags(q ListQuery) (PageResult, error) {
	var rows []tagRecord
	db := s.db.Table("game_game_tags t").Select("t.*,COUNT(r.id) game_count").Joins("LEFT JOIN game_game_tag_relations r ON r.tag_id=t.id AND r.deleted_at IS NULL").Where("t.deleted_at IS NULL").Group("t.id")
	db = content.ApplySite(db, "t", q.Scope)
	db = contains(db, "t.code", q.Values["code"])
	db = contains(db, "t.name_zh", q.Values["nameZh"])
	db = contains(db, "t.name_en", q.Values["nameEn"])
	db = boolFilter(db, "t.enabled", q.Values["enabled"])
	result, err := paginate(db.Order("t.sort_order,t.id"), q, &rows)
	for i := range rows {
		signTagAsset(&rows[i])
	}
	result.List = rows
	return result, err
}
func (s *Service) TagDetail(id uint64, scope content.Scope) (map[string]any, error) {
	var row tagRecord
	db := s.db.Table("game_game_tags t").Select("t.*,COUNT(r.id) game_count").Joins("LEFT JOIN game_game_tag_relations r ON r.tag_id=t.id AND r.deleted_at IS NULL").Where("t.id=? AND t.deleted_at IS NULL", id).Group("t.id")
	db = content.ApplySite(db, "t", scope)
	if err := db.Take(&row).Error; err != nil {
		return nil, notFound(err, "TAG_NOT_FOUND")
	}
	signTagAsset(&row)
	var games []gameRecord
	gameDB := s.db.Table("game_games g").Select("g.id,g.game_code,g.name,p.provider_code,g.provider_game_code,g.platform_enabled,g.sort").Joins("JOIN game_game_providers p ON p.id=g.provider_id").Joins("JOIN game_game_tag_relations r ON r.game_id=g.id AND r.deleted_at IS NULL").Where("r.tag_id=? AND g.deleted_at IS NULL", id)
	gameDB = content.ApplySite(gameDB, "g", scope)
	gameDB = content.ApplySite(gameDB, "p", scope)
	gameDB = content.ApplySite(gameDB, "r", scope)
	_ = gameDB.Order("g.sort,g.id").Find(&games).Error
	return map[string]any{"overview": row, "games": games, "operationLogs": s.operationLogs("game_tag", id)}, nil
}
func (s *Service) SaveTag(id uint64, in TagInput, actor Actor) (uint64, error) {
	in.Code = normalizeCode(in.Code)
	if err := validateCode(in.Code); err != nil {
		return 0, err
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var before tagRecord
		action, permission := "create", "game:tag:create"
		if id > 0 {
			if err := content.ApplySite(tx.Model(&tagRecord{}), "", actor.Scope).Where("id=? AND deleted_at IS NULL", id).Take(&before).Error; err != nil {
				return notFound(err, "TAG_NOT_FOUND")
			}
			action, permission = "update", "game:tag:update"
		}
		icon := adminAssetRef(in.Icon)
		row := tagRecord{SiteID: actor.Scope.SiteID, ID: id, Code: in.Code, NameZh: in.NameZh, NameEn: in.NameEn, Icon: icon, SortOrder: in.SortOrder, Enabled: in.Enabled}
		if id == 0 {
			if err := tx.Create(&row).Error; err != nil {
				return duplicateCode(err, "TAG_CODE_EXISTS", "tag code already exists")
			}
			id = row.ID
		} else if err := content.ApplySite(tx.Model(&tagRecord{}), "", actor.Scope).Where("id=?", id).Updates(map[string]any{
			"code": in.Code, "name_zh": in.NameZh, "name_en": in.NameEn, "icon_url": row.Icon,
			"sort_order": in.SortOrder, "enabled": in.Enabled,
		}).Error; err != nil {
			return err
		}
		return audit(tx, actor, permission, action, "game_tag", id, before, row, in.Reason)
	})
	return id, err
}
func (s *Service) DeleteTag(id uint64, reason string, actor Actor) error {
	return s.safeDelete(id, reason, actor, "game_game_tags", "game_game_tag_relations", "tag_id", "game_tag", "game:tag:delete", "TAG_NOT_FOUND", "TAG_IN_USE")
}
func (s *Service) SetTagGames(id uint64, in RelationInput, actor Actor) error {
	return s.replaceTaxonomyGames(id, in, actor, "tag")
}

func (s *Service) replaceTaxonomyGames(targetID uint64, in RelationInput, actor Actor, kind string) error {
	config := map[string]struct {
		targetTable, relationTable, foreign, target, permission, notFoundCode string
	}{
		"tag": {"game_game_tags", "game_game_tag_relations", "tag_id", "game_tag", "game:tag:assign", "TAG_NOT_FOUND"},
	}[kind]
	gameIDs := uniqueIDs(in.IDs)
	return s.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := content.ApplySite(tx.Table(config.targetTable), "", actor.Scope).Where("id=? AND deleted_at IS NULL", targetID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return newError(config.notFoundCode, "record not found")
		}
		if len(gameIDs) > 0 {
			if err := content.ApplySite(tx.Table("game_games"), "", actor.Scope).Where("id IN ? AND deleted_at IS NULL", gameIDs).Count(&count).Error; err != nil {
				return err
			}
			if count != int64(len(gameIDs)) {
				return newError("INVALID_RELATION", "one or more games do not exist")
			}
		}
		var before []uint64
		if err := content.ApplySite(tx.Table(config.relationTable), "", actor.Scope).Where(config.foreign+"=? AND deleted_at IS NULL", targetID).Order("id").Pluck("game_id", &before).Error; err != nil {
			return err
		}
		now := time.Now()
		if err := content.ApplySite(tx.Table(config.relationTable), "", actor.Scope).Where(config.foreign+"=? AND deleted_at IS NULL", targetID).Update("deleted_at", now).Error; err != nil {
			return err
		}
		for index, gameID := range gameIDs {
			row := map[string]any{"site_id": actor.Scope.SiteID, "game_id": gameID, config.foreign: targetID, "created_at": now, "updated_at": now, "deleted_at": nil}
			if kind == "category" {
				row["sort"] = index
			}
			if err := tx.Table(config.relationTable).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "game_id"}, {Name: config.foreign}}, DoUpdates: clause.Assignments(row)}).Create(row).Error; err != nil {
				return err
			}
		}
		return audit(tx, actor, config.permission, "assign_games", config.target, targetID, before, gameIDs, in.Reason)
	})
}

func (s *Service) ListCollections(q ListQuery) (PageResult, error) {
	var rows []collectionRecord
	db := s.db.Table("game_game_collections c").Select("c.*,COUNT(r.id) game_count").Joins("LEFT JOIN game_game_collection_relations r ON r.collection_id=c.id AND r.deleted_at IS NULL").Where("c.deleted_at IS NULL").Group("c.id")
	db = content.ApplySite(db, "c", q.Scope)
	db = contains(db, "c.code", q.Values["code"])
	db = contains(db, "c.name_zh", q.Values["nameZh"])
	db = contains(db, "c.name_en", q.Values["nameEn"])
	db = boolFilter(db, "c.enabled", q.Values["enabled"])
	db = timeRange(db, "c.start_at", q.Values["startFrom"], q.Values["startTo"])
	result, err := paginate(db.Order("c.sort_order,c.id"), q, &rows)
	for i := range rows {
		signCollectionAsset(&rows[i])
	}
	result.List = rows
	return result, err
}
func (s *Service) CollectionDetail(id uint64, scope content.Scope) (map[string]any, error) {
	var row collectionRecord
	db := s.db.Table("game_game_collections c").Select("c.*,COUNT(r.id) game_count").Joins("LEFT JOIN game_game_collection_relations r ON r.collection_id=c.id AND r.deleted_at IS NULL").Where("c.id=? AND c.deleted_at IS NULL", id).Group("c.id")
	db = content.ApplySite(db, "c", scope)
	if err := db.Take(&row).Error; err != nil {
		return nil, notFound(err, "COLLECTION_NOT_FOUND")
	}
	signCollectionAsset(&row)
	var games []gameRecord
	gameDB := s.db.Table("game_games g").Select("g.id,g.game_code,g.name,p.provider_code,g.provider_game_code,g.platform_enabled,r.sort_order sort").Joins("JOIN game_game_providers p ON p.id=g.provider_id").Joins("JOIN game_game_collection_relations r ON r.game_id=g.id AND r.deleted_at IS NULL").Where("r.collection_id=? AND g.deleted_at IS NULL", id)
	gameDB = content.ApplySite(gameDB, "g", scope)
	gameDB = content.ApplySite(gameDB, "p", scope)
	gameDB = content.ApplySite(gameDB, "r", scope)
	_ = gameDB.Order("r.sort_order,r.id").Find(&games).Error
	return map[string]any{"overview": row, "games": games, "displayRules": map[string]any{"enabled": row.Enabled, "startTime": row.StartAt, "endTime": row.EndAt, "sortOrder": row.SortOrder}, "operationLogs": s.operationLogs("game_collection", id)}, nil
}
func (s *Service) SaveCollection(id uint64, in CollectionInput, actor Actor) (uint64, error) {
	in.Code = normalizeCode(in.Code)
	if err := validateCode(in.Code); err != nil {
		return 0, err
	}
	if err := validateTimeRange(in.StartTime, in.EndTime); err != nil {
		return 0, err
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var before collectionRecord
		action, permission := "create", "game:collection:create"
		if id > 0 {
			if err := content.ApplySite(tx.Model(&collectionRecord{}), "", actor.Scope).Where("id=? AND deleted_at IS NULL", id).Take(&before).Error; err != nil {
				return notFound(err, "COLLECTION_NOT_FOUND")
			}
			action, permission = "update", "game:collection:update"
		}
		bannerURL := adminAssetRef(in.BannerURL)
		row := collectionRecord{SiteID: actor.Scope.SiteID, ID: id, Code: in.Code, NameZh: in.NameZh, NameEn: in.NameEn, DescriptionZh: in.DescriptionZh, DescriptionEn: in.DescriptionEn, BannerURL: bannerURL, SortOrder: in.SortOrder, Enabled: in.Enabled, StartAt: in.StartTime, EndAt: in.EndTime}
		if id == 0 {
			if err := tx.Create(&row).Error; err != nil {
				return duplicateCode(err, "COLLECTION_CODE_EXISTS", "collection code already exists")
			}
			id = row.ID
		} else if err := content.ApplySite(tx.Model(&collectionRecord{}), "", actor.Scope).Where("id=?", id).Updates(map[string]any{
			"code": in.Code, "name_zh": in.NameZh, "name_en": in.NameEn,
			"description_zh": in.DescriptionZh, "description_en": in.DescriptionEn,
			"banner_url": row.BannerURL, "sort_order": in.SortOrder, "enabled": in.Enabled,
			"start_at": in.StartTime, "end_at": in.EndTime,
		}).Error; err != nil {
			return err
		}
		if in.Games != nil {
			if err := content.ApplySite(tx.Table("game_game_collection_relations"), "", actor.Scope).Where("collection_id=? AND deleted_at IS NULL", id).Update("deleted_at", time.Now()).Error; err != nil {
				return err
			}
			for _, item := range *in.Games {
				relation := map[string]any{"site_id": actor.Scope.SiteID, "collection_id": id, "game_id": item.GameID, "sort_order": item.SortOrder, "deleted_at": nil, "created_at": time.Now(), "updated_at": time.Now()}
				if err := tx.Table("game_game_collection_relations").Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "collection_id"}, {Name: "game_id"}}, DoUpdates: clause.Assignments(relation)}).Create(relation).Error; err != nil {
					return err
				}
			}
		}
		return audit(tx, actor, permission, action, "game_collection", id, before, row, in.Reason)
	})
	return id, err
}
func (s *Service) DeleteCollection(id uint64, reason string, actor Actor) error {
	return s.safeDelete(id, reason, actor, "game_game_collections", "game_game_collection_relations", "collection_id", "game_collection", "game:collection:delete", "COLLECTION_NOT_FOUND", "COLLECTION_IN_USE")
}

func (s *Service) safeDelete(id uint64, reason string, actor Actor, table string, relation string, foreign string, target string, permission string, notFoundCode string, inUseCode string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := content.ApplySite(tx.Table(table), "", actor.Scope).Where("id=? AND deleted_at IS NULL", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return newError(notFoundCode, "record not found")
		}
		if err := content.ApplySite(tx.Table(relation), "", actor.Scope).Where(foreign+"=? AND deleted_at IS NULL", id).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return newError(inUseCode, "record is assigned to games")
		}
		if err := content.ApplySite(tx.Table(table), "", actor.Scope).Where("id=?", id).Update("deleted_at", time.Now()).Error; err != nil {
			return err
		}
		return audit(tx, actor, permission, "delete", target, id, nil, map[string]any{"deleted": true}, reason)
	})
}
func statusInt(enabled bool) int {
	if enabled {
		return 1
	}
	return 2
}
