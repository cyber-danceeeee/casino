package operationadmin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/model"
	"casino/internal/storage"
	"casino/internal/tenancy"
)

const (
	H5NavigationPublishDraft     = "DRAFT"
	H5NavigationPublishPublished = "PUBLISHED"
)

var (
	h5NavigationCodePattern    = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)
	allowedH5NavigationActions = map[string]bool{
		"OPEN_INTERNAL_PAGE": true,
		"OPEN_EXTERNAL_URL":  true,
		"OPEN_DEPOSIT":       true,
		"OPEN_WITHDRAW":      true,
		"OPEN_SUPPORT":       true,
		"OPEN_PROMOTION":     true,
		"OPEN_GAME":          true,
		"OPEN_DOWNLOAD":      true,
	}
	allowedH5NavigationPositions = map[string]bool{
		"BOTTOM_NAV":   true,
		"HOME_QUICK":   true,
		"PROFILE_MENU": true,
	}
	allowedH5NavigationBadges     = map[string]bool{"NONE": true, "NEW": true, "HOT": true}
	allowedH5NavigationVisibility = map[string]bool{
		"ALL": true, "GUEST": true, "LOGGED_IN": true,
	}
)

type H5NavigationDraft struct {
	MenuCode         string            `json:"menuCode"`
	Labels           map[string]string `json:"labels"`
	IconAssetID      uint64            `json:"iconAssetID"`
	LottieAssetID    uint64            `json:"lottieAssetID"`
	SoundAssetID     uint64            `json:"soundAssetID"`
	DisplayPosition  string            `json:"displayPosition"`
	SortOrder        int               `json:"sortOrder"`
	BadgeType        string            `json:"badgeType"`
	VisibilityScope  string            `json:"visibilityScope"`
	CountryCodes     []string          `json:"countryCodes"`
	LanguageCodes    []string          `json:"languageCodes"`
	StartAt          *time.Time        `json:"startAt"`
	EndAt            *time.Time        `json:"endAt"`
	Enabled          bool              `json:"enabled"`
	ActionType       string            `json:"actionType"`
	ActionTarget     string            `json:"actionTarget"`
	ProviderCode     string            `json:"providerCode"`
	GameID           uint64            `json:"gameID"`
	ProviderGameCode string            `json:"providerGameCode"`
	GameName         string            `json:"gameName"`
	Deleted          bool              `json:"deleted,omitempty"`
}

type H5NavigationInput struct {
	MenuCode        string            `json:"menuCode" binding:"max=64"`
	LabelZh         string            `json:"labelZh" binding:"max=64"`
	LabelEn         string            `json:"labelEn" binding:"max=64"`
	Labels          map[string]string `json:"labels"`
	IconAssetID     uint64            `json:"iconAssetID"`
	LottieAssetID   uint64            `json:"lottieAssetID"`
	SoundAssetID    uint64            `json:"soundAssetID"`
	DisplayPosition string            `json:"displayPosition" binding:"required,max=32"`
	SortOrder       int               `json:"sortOrder"`
	BadgeType       string            `json:"badgeType" binding:"required,max=16"`
	VisibilityScope string            `json:"visibilityScope" binding:"required,max=16"`
	CountryCodes    []string          `json:"countryCodes"`
	LanguageCodes   []string          `json:"languageCodes"`
	StartAt         *time.Time        `json:"startAt"`
	EndAt           *time.Time        `json:"endAt"`
	Enabled         bool              `json:"enabled"`
	ActionType      string            `json:"actionType" binding:"required,max=32"`
	ActionTarget    string            `json:"actionTarget" binding:"max=500"`
	ProviderCode    string            `json:"providerCode" binding:"max=64"`
	GameID          uint64            `json:"gameID"`
	Reason          string            `json:"reason" binding:"max=500"`
}

type H5NavigationSortInput struct {
	DisplayPosition string   `json:"displayPosition" binding:"required,max=32"`
	IDs             []uint64 `json:"ids" binding:"required,min=1"`
	Reason          string   `json:"reason" binding:"max=500"`
}

type H5NavigationRecord struct {
	model.GameH5NavigationItems
	IsFallback             bool                `gorm:"-" json:"isFallback"`
	CanEdit                bool                `gorm:"-" json:"canEdit"`
	CanDelete              bool                `gorm:"-" json:"canDelete"`
	Visible                bool                `gorm:"-" json:"visible"`
	DraftConfiguration     H5NavigationDraft   `gorm:"-" json:"draftConfiguration"`
	PublishedConfiguration H5NavigationDraft   `gorm:"-" json:"publishedConfiguration"`
	IconAsset              *assetLibraryRecord `gorm:"-" json:"iconAsset,omitempty"`
	LottieAsset            *assetLibraryRecord `gorm:"-" json:"lottieAsset,omitempty"`
	SoundAsset             *assetLibraryRecord `gorm:"-" json:"soundAsset,omitempty"`
}

type PublicH5NavigationItem struct {
	ID               uint64            `json:"id"`
	MenuCode         string            `json:"menuCode"`
	Label            string            `json:"label"`
	LabelZh          string            `json:"labelZh"`
	LabelEn          string            `json:"labelEn"`
	Labels           map[string]string `json:"labels"`
	IconURL          string            `json:"iconURL"`
	LottieURL        string            `json:"lottieURL"`
	SoundURL         string            `json:"soundURL"`
	DisplayPosition  string            `json:"displayPosition"`
	SortOrder        int               `json:"sortOrder"`
	BadgeType        string            `json:"badgeType"`
	ActionType       string            `json:"actionType"`
	ActionTarget     string            `json:"actionTarget"`
	ProviderCode     string            `json:"providerCode"`
	GameID           uint64            `json:"gameID"`
	ProviderGameCode string            `json:"providerGameCode"`
	GameName         string            `json:"gameName"`
	ConfigVersion    uint64            `json:"configVersion"`
	Enabled          bool              `json:"enabled"`
}

func newH5NavigationCode() (string, error) {
	value := make([]byte, 8)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return "nav-" + hex.EncodeToString(value), nil
}

func (s *Service) EnsureH5NavigationSchema() error {
	return s.db.AutoMigrate(&model.GameH5NavigationItems{})
}

func (s *Service) ListH5Navigation(q ListQuery) (PageResult, error) {
	if err := requireH5NavigationTenantScope(q.Scope); err != nil {
		return PageResult{}, err
	}
	if err := s.EnsureH5NavigationSchema(); err != nil {
		return PageResult{}, err
	}
	var rows []model.GameH5NavigationItems
	db := tenancy.ApplySiteStrict(s.db.Model(&model.GameH5NavigationItems{}), "", q.Scope).Where("deleted_at IS NULL")
	if keyword := strings.TrimSpace(q.Values["keyword"]); keyword != "" {
		db = db.Where("menu_code LIKE ? OR draft_payload LIKE ? OR published_payload LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	db = equal(db, "publish_status", strings.ToUpper(q.Values["publishStatus"]))
	result, err := page(db.Order("display_position ASC,sort_order ASC,id ASC"), q, &rows)
	if err != nil {
		return result, err
	}
	list := make([]H5NavigationRecord, 0, len(rows))
	for _, row := range rows {
		item := H5NavigationRecord{GameH5NavigationItems: row}
		s.hydrateH5Navigation(&item, q.Scope)
		if position := strings.ToUpper(strings.TrimSpace(q.Values["displayPosition"])); position != "" && item.DraftConfiguration.DisplayPosition != position {
			continue
		}
		if action := strings.ToUpper(strings.TrimSpace(q.Values["actionType"])); action != "" && item.DraftConfiguration.ActionType != action {
			continue
		}
		list = append(list, item)
	}
	sort.SliceStable(list, func(i, j int) bool {
		left, right := list[i].DraftConfiguration, list[j].DraftConfiguration
		if left.DisplayPosition != right.DisplayPosition {
			return left.DisplayPosition < right.DisplayPosition
		}
		if left.SortOrder != right.SortOrder {
			return left.SortOrder < right.SortOrder
		}
		return list[i].ID < list[j].ID
	})
	result.List = list
	return result, nil
}

func (s *Service) GetH5Navigation(id uint64, scope tenancy.Scope) (map[string]any, error) {
	if err := requireH5NavigationTenantScope(scope); err != nil {
		return nil, err
	}
	if err := s.EnsureH5NavigationSchema(); err != nil {
		return nil, err
	}
	row, err := s.h5NavigationByID(s.db, id, scope, false)
	if err != nil {
		return nil, err
	}
	item := H5NavigationRecord{GameH5NavigationItems: row}
	s.hydrateH5Navigation(&item, scope)
	logs, _ := s.operationLogs("h5_navigation", id, scope)
	return map[string]any{"overview": item, "operationLogs": logs}, nil
}

func (s *Service) SaveH5Navigation(id uint64, in H5NavigationInput, actor Actor) (uint64, error) {
	if err := requireH5NavigationTenantScope(actor.Scope); err != nil {
		return 0, err
	}
	if err := s.EnsureH5NavigationSchema(); err != nil {
		return 0, err
	}
	returnID := id
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var before model.GameH5NavigationItems
		if id > 0 {
			var err error
			before, err = s.h5NavigationByID(tx, id, actor.Scope, true)
			if err != nil {
				return err
			}
			in.MenuCode = before.MenuCode
		} else if strings.TrimSpace(in.MenuCode) == "" {
			code, err := newH5NavigationCode()
			if err != nil {
				return err
			}
			in.MenuCode = code
		}
		draft, err := normalizeH5NavigationInput(tx, in, actor.Scope)
		if err != nil {
			return err
		}
		dimensions, err := loadTenantSiteDimensions(tx, actor.Scope)
		if err != nil {
			return err
		}
		draft.CountryCodes = []string{strings.ToUpper(strings.TrimSpace(dimensions.CountryCode))}
		draft.LanguageCodes = []string{normalizeH5NavigationLanguage(dimensions.DefaultLanguage)}
		if err := s.validateH5NavigationAssets(tx, actor.Scope, draft); err != nil {
			return err
		}
		payload, _ := json.Marshal(draft)
		nowValue := time.Now()
		updates := map[string]any{
			"icon_asset_id_draft": draft.IconAssetID, "lottie_asset_id_draft": draft.LottieAssetID,
			"sound_asset_id_draft": draft.SoundAssetID, "draft_payload": string(payload),
			"publish_status": H5NavigationPublishDraft, "draft_updated_at": nowValue,
			"updated_by": actor.ID, "updated_at": nowValue,
		}
		if id == 0 {
			row := model.GameH5NavigationItems{
				TenantID: actor.Scope.TenantID, SiteID: actor.Scope.SiteID, MenuCode: draft.MenuCode,
				LabelZh: draft.Labels["zh-CN"], LabelEn: draft.Labels["en-US"],
				IconAssetIDDraft: draft.IconAssetID, LottieAssetIDDraft: draft.LottieAssetID,
				SoundAssetIDDraft: draft.SoundAssetID, DisplayPosition: draft.DisplayPosition,
				ActionType: draft.ActionType, SortOrder: draft.SortOrder, BadgeType: draft.BadgeType,
				VisibilityScope: draft.VisibilityScope, DraftPayload: string(payload),
				PublishStatus: H5NavigationPublishDraft, CreatedBy: actor.ID, UpdatedBy: actor.ID,
				DraftUpdatedAt: &nowValue,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			returnID = row.ID
		} else if err := tenancy.ApplySiteStrict(tx.Model(&model.GameH5NavigationItems{}), "", actor.Scope).Where("id=?", id).Updates(updates).Error; err != nil {
			return err
		}
		if err := markH5NavigationSiteDraft(tx, actor); err != nil {
			return err
		}
		permission, action := "operation:h5-navigation:create", "h5_navigation.create_draft"
		if id > 0 {
			permission, action = "operation:h5-navigation:update", "h5_navigation.update_draft"
		}
		return audit(tx, actor, permission, action, "h5_navigation", returnID, before, draft, in.Reason)
	})
	return returnID, err
}

func (s *Service) CopyH5Navigation(id uint64, actor Actor) (uint64, error) {
	if err := requireH5NavigationTenantScope(actor.Scope); err != nil {
		return 0, err
	}
	row, err := s.h5NavigationByID(s.db, id, actor.Scope, false)
	if err != nil {
		return 0, err
	}
	draft := decodeH5NavigationDraft(row.DraftPayload)
	if strings.TrimSpace(row.DraftPayload) == "" {
		draft = decodeH5NavigationDraft(row.PublishedPayload)
	}
	draft.MenuCode = ""
	draft.Labels["zh-CN"] = strings.TrimSpace(draft.Labels["zh-CN"]) + " 副本"
	draft.Labels["en-US"] = strings.TrimSpace(draft.Labels["en-US"]) + " Copy"
	draft.SortOrder++
	draft.Enabled = false
	return s.SaveH5Navigation(0, inputFromH5NavigationDraft(draft, "copy navigation entry"), actor)
}

func (s *Service) DeleteH5Navigation(id uint64, actor Actor) error {
	if err := requireH5NavigationTenantScope(actor.Scope); err != nil {
		return err
	}
	row, err := s.h5NavigationByID(s.db, id, actor.Scope, false)
	if err != nil {
		return err
	}
	draft := decodeH5NavigationDraft(row.DraftPayload)
	if strings.TrimSpace(row.DraftPayload) == "" {
		draft = decodeH5NavigationDraft(row.PublishedPayload)
	}
	draft.Deleted = true
	payload, _ := json.Marshal(draft)
	return s.db.Transaction(func(tx *gorm.DB) error {
		nowValue := time.Now()
		if err := tenancy.ApplySiteStrict(tx.Model(&model.GameH5NavigationItems{}), "", actor.Scope).Where("id=?", id).Updates(map[string]any{
			"draft_payload": string(payload), "publish_status": H5NavigationPublishDraft,
			"draft_updated_at": nowValue, "updated_by": actor.ID, "updated_at": nowValue,
		}).Error; err != nil {
			return err
		}
		if err := markH5NavigationSiteDraft(tx, actor); err != nil {
			return err
		}
		return audit(tx, actor, "operation:h5-navigation:delete", "h5_navigation.delete_draft", "h5_navigation", id, row, draft, "")
	})
}

func (s *Service) SortH5Navigation(in H5NavigationSortInput, actor Actor) error {
	if err := requireH5NavigationTenantScope(actor.Scope); err != nil {
		return err
	}
	position := strings.ToUpper(strings.TrimSpace(in.DisplayPosition))
	if !allowedH5NavigationPositions[position] {
		return newError("INVALID_H5_NAVIGATION_POSITION", "invalid navigation display position")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		for index, id := range in.IDs {
			row, err := s.h5NavigationByID(tx, id, actor.Scope, true)
			if err != nil {
				return err
			}
			draft := decodeH5NavigationDraft(row.DraftPayload)
			if strings.TrimSpace(row.DraftPayload) == "" {
				draft = decodeH5NavigationDraft(row.PublishedPayload)
			}
			if draft.DisplayPosition != position {
				return newError("H5_NAVIGATION_POSITION_MISMATCH", "all reordered entries must use the same display position")
			}
			draft.SortOrder = (index + 1) * 10
			payload, _ := json.Marshal(draft)
			if err := tenancy.ApplySiteStrict(tx.Model(&model.GameH5NavigationItems{}), "", actor.Scope).Where("id=?", id).Updates(map[string]any{
				"draft_payload": string(payload), "publish_status": H5NavigationPublishDraft,
				"draft_updated_at": time.Now(), "updated_by": actor.ID, "updated_at": time.Now(),
			}).Error; err != nil {
				return err
			}
		}
		if err := markH5NavigationSiteDraft(tx, actor); err != nil {
			return err
		}
		return audit(tx, actor, "operation:h5-navigation:update", "h5_navigation.sort_draft", "h5_navigation", 0, nil, in.IDs, in.Reason)
	})
}

func (s *Service) H5Navigation(scope tenancy.Scope, language, country string, loggedIn bool, position string, at time.Time) ([]PublicH5NavigationItem, error) {
	if scope.TenantID == 0 || scope.SiteID == 0 {
		return []PublicH5NavigationItem{}, nil
	}
	position = strings.ToUpper(strings.TrimSpace(position))
	if position == "" {
		position = "BOTTOM_NAV"
	}
	if !allowedH5NavigationPositions[position] {
		return []PublicH5NavigationItem{}, nil
	}
	var rows []model.GameH5NavigationItems
	err := tenancy.ApplySiteStrict(s.db.Model(&model.GameH5NavigationItems{}), "", scope).
		Where("display_position=? AND config_version>0 AND enabled=1 AND deleted_at IS NULL", position).
		Where("(start_at IS NULL OR start_at<=?) AND (end_at IS NULL OR end_at>=?)", at, at).
		Order("sort_order ASC,id ASC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]PublicH5NavigationItem, 0, len(rows))
	for _, row := range rows {
		draft := decodeH5NavigationDraft(row.PublishedPayload)
		if draft.MenuCode == "" {
			draft = legacyPublishedH5Navigation(row)
		}
		if !h5NavigationVisible(draft, language, country, loggedIn, at) {
			continue
		}
		iconURL, iconOK := s.publicH5NavigationAsset(scope, draft.IconAssetID, "IMAGE")
		lottieURL, lottieOK := s.publicH5NavigationAsset(scope, draft.LottieAssetID, "LOTTIE")
		if (draft.IconAssetID > 0 && !iconOK) || (draft.LottieAssetID > 0 && !lottieOK) {
			continue
		}
		soundURL, _ := s.publicH5NavigationAsset(scope, draft.SoundAssetID, "AUDIO")
		label := h5NavigationLabel(draft.Labels, language)
		out = append(out, PublicH5NavigationItem{
			ID: row.ID, MenuCode: draft.MenuCode, Label: label, LabelZh: draft.Labels["zh-CN"],
			LabelEn: draft.Labels["en-US"], Labels: draft.Labels, IconURL: iconURL, LottieURL: lottieURL,
			SoundURL: soundURL, DisplayPosition: draft.DisplayPosition, SortOrder: draft.SortOrder,
			BadgeType: draft.BadgeType, ActionType: draft.ActionType, ActionTarget: draft.ActionTarget,
			ProviderCode: draft.ProviderCode, GameID: draft.GameID, ProviderGameCode: draft.ProviderGameCode,
			GameName: draft.GameName, ConfigVersion: row.ConfigVersion, Enabled: true,
		})
	}
	return out, nil
}

func requireH5NavigationTenantScope(scope tenancy.Scope) error {
	if !scope.SiteScoped() {
		return newError("H5_NAVIGATION_SCOPE_REQUIRED", "tenant site scope is required")
	}
	return nil
}

func normalizeH5NavigationInput(tx *gorm.DB, in H5NavigationInput, scope tenancy.Scope) (H5NavigationDraft, error) {
	code := strings.ToLower(strings.TrimSpace(in.MenuCode))
	if !h5NavigationCodePattern.MatchString(code) {
		return H5NavigationDraft{}, newError("INVALID_H5_NAVIGATION_CODE", "navigation code must use lowercase letters, numbers, or hyphens")
	}
	labels := normalizeH5NavigationLabels(in.Labels, in.LabelZh, in.LabelEn)
	if labels["zh-CN"] == "" || labels["en-US"] == "" {
		return H5NavigationDraft{}, newError("H5_NAVIGATION_LABELS_REQUIRED", "Chinese and English entry names are required")
	}
	position := strings.ToUpper(strings.TrimSpace(in.DisplayPosition))
	if !allowedH5NavigationPositions[position] {
		return H5NavigationDraft{}, newError("INVALID_H5_NAVIGATION_POSITION", "invalid navigation display position")
	}
	badge := strings.ToUpper(strings.TrimSpace(in.BadgeType))
	if !allowedH5NavigationBadges[badge] {
		return H5NavigationDraft{}, newError("INVALID_H5_NAVIGATION_BADGE", "invalid navigation badge")
	}
	visibility := strings.ToUpper(strings.TrimSpace(in.VisibilityScope))
	if !allowedH5NavigationVisibility[visibility] {
		return H5NavigationDraft{}, newError("INVALID_H5_NAVIGATION_VISIBILITY", "invalid navigation visibility")
	}
	if err := validateTime(in.StartAt, in.EndAt); err != nil {
		return H5NavigationDraft{}, err
	}
	action := strings.ToUpper(strings.TrimSpace(in.ActionType))
	if !allowedH5NavigationActions[action] {
		return H5NavigationDraft{}, newError("INVALID_H5_NAVIGATION_ACTION", "invalid navigation action")
	}
	target := strings.TrimSpace(in.ActionTarget)
	providerCode := strings.ToLower(strings.TrimSpace(in.ProviderCode))
	draft := H5NavigationDraft{
		MenuCode: code, Labels: labels, IconAssetID: in.IconAssetID, LottieAssetID: in.LottieAssetID,
		SoundAssetID: in.SoundAssetID, DisplayPosition: position, SortOrder: in.SortOrder,
		BadgeType: badge, VisibilityScope: visibility, CountryCodes: normalizeH5NavigationCodes(in.CountryCodes, true),
		LanguageCodes: normalizeH5NavigationCodes(in.LanguageCodes, false), StartAt: in.StartAt, EndAt: in.EndAt,
		Enabled: in.Enabled, ActionType: action, ActionTarget: target,
	}
	switch action {
	case "OPEN_INTERNAL_PAGE":
		if !validInternalH5Target(target) {
			return H5NavigationDraft{}, newError("INVALID_H5_NAVIGATION_TARGET", "internal navigation target must be an H5 path")
		}
	case "OPEN_EXTERNAL_URL", "OPEN_DOWNLOAD":
		if !validHTTPSURL(target) {
			return H5NavigationDraft{}, newError("INVALID_H5_NAVIGATION_TARGET", "external navigation target must use HTTPS")
		}
	case "OPEN_PROMOTION":
		if target == "" || strings.Contains(target, "://") || strings.ContainsAny(target, "<>\n\r") {
			return H5NavigationDraft{}, newError("INVALID_H5_NAVIGATION_TARGET", "promotion target is invalid")
		}
	case "OPEN_GAME":
		if in.GameID == 0 {
			return H5NavigationDraft{}, newError("H5_NAVIGATION_GAME_REQUIRED", "a tenant game is required")
		}
		game, err := h5NavigationGame(tx, scope, in.GameID, providerCode)
		if err != nil {
			return H5NavigationDraft{}, err
		}
		draft.ProviderCode = game.ProviderCode
		draft.GameID = game.ID
		draft.ProviderGameCode = game.ProviderGameCode
		draft.GameName = game.Name
		draft.ActionTarget = game.ProviderGameCode
	default:
		draft.ActionTarget = ""
	}
	return draft, nil
}

type h5NavigationGameRow struct {
	ID               uint64
	ProviderCode     string
	ProviderGameCode string
	Name             string
}

func h5NavigationGame(tx *gorm.DB, scope tenancy.Scope, gameID uint64, providerCode string) (h5NavigationGameRow, error) {
	var game h5NavigationGameRow
	db := tenancy.ApplySite(tx.Table("game_games g"), "g", scope).
		Scopes(func(db *gorm.DB) *gorm.DB { return tenancy.ApplySite(db, "p", scope) }).
		Select("g.id,p.provider_code,g.provider_game_code,g.name").
		Joins("JOIN game_game_providers p ON p.id=g.provider_id AND p.deleted_at IS NULL").
		Where("g.id=? AND g.deleted_at IS NULL", gameID)
	if providerCode != "" {
		db = db.Where("p.provider_code=?", providerCode)
	}
	if err := db.Take(&game).Error; err != nil {
		return game, newError("H5_NAVIGATION_GAME_NOT_FOUND", "selected tenant game does not exist")
	}
	return game, nil
}

func (s *Service) validateH5NavigationAssets(db *gorm.DB, scope tenancy.Scope, draft H5NavigationDraft) error {
	if draft.IconAssetID == 0 && draft.LottieAssetID == 0 {
		return newError("H5_NAVIGATION_ASSET_REQUIRED", "an icon or Lottie asset is required")
	}
	for _, expected := range []struct {
		id        uint64
		assetType string
	}{
		{draft.IconAssetID, "IMAGE"}, {draft.LottieAssetID, "LOTTIE"}, {draft.SoundAssetID, "AUDIO"},
	} {
		if expected.id == 0 {
			continue
		}
		var count int64
		if err := s.assetVisibleQuery(db.Model(&assetLibraryRecord{}), scope).
			Where("id=? AND asset_type=? AND status=? AND deleted_at IS NULL", expected.id, expected.assetType, AssetStatusActive).
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return newError("H5_NAVIGATION_ASSET_INVALID", "navigation asset is inactive, missing, or outside the site scope")
		}
	}
	return nil
}

func (s *Service) h5NavigationByID(db *gorm.DB, id uint64, scope tenancy.Scope, lock bool) (model.GameH5NavigationItems, error) {
	var row model.GameH5NavigationItems
	query := tenancy.ApplySiteStrict(db.Model(&model.GameH5NavigationItems{}), "", scope).Where("id=? AND deleted_at IS NULL", id)
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.Take(&row).Error; err != nil {
		return row, notFound(err, "H5_NAVIGATION_NOT_FOUND")
	}
	return row, nil
}

func (s *Service) hydrateH5Navigation(row *H5NavigationRecord, scope tenancy.Scope) {
	row.DraftConfiguration = decodeH5NavigationDraft(row.DraftPayload)
	if row.DraftConfiguration.MenuCode == "" {
		row.DraftConfiguration = legacyPublishedH5Navigation(row.GameH5NavigationItems)
	}
	row.PublishedConfiguration = decodeH5NavigationDraft(row.PublishedPayload)
	if row.PublishedConfiguration.MenuCode == "" && row.ConfigVersion > 0 {
		row.PublishedConfiguration = legacyPublishedH5Navigation(row.GameH5NavigationItems)
	}
	row.IconAsset = s.h5NavigationAsset(row.DraftConfiguration.IconAssetID, scope)
	row.LottieAsset = s.h5NavigationAsset(row.DraftConfiguration.LottieAssetID, scope)
	row.SoundAsset = s.h5NavigationAsset(row.DraftConfiguration.SoundAssetID, scope)
	row.IsFallback = row.PlatformSourceID > 0
	row.CanEdit = !row.IsFallback
	row.CanDelete = !row.IsFallback
	row.Visible = row.DraftConfiguration.Enabled
}

func (s *Service) h5NavigationAsset(id uint64, scope tenancy.Scope) *assetLibraryRecord {
	if id == 0 {
		return nil
	}
	var row assetLibraryRecord
	if s.assetVisibleQuery(s.db.Model(&assetLibraryRecord{}), scope).Where("id=? AND deleted_at IS NULL", id).Take(&row).Error != nil {
		return nil
	}
	s.signAsset(&row)
	return &row
}

func (s *Service) publicH5NavigationAsset(scope tenancy.Scope, id uint64, assetType string) (string, bool) {
	if id == 0 {
		return "", false
	}
	var row assetLibraryRecord
	if s.assetVisibleQuery(s.db.Model(&assetLibraryRecord{}), scope).
		Where("id=? AND asset_type=? AND status=? AND deleted_at IS NULL", id, assetType, AssetStatusActive).Take(&row).Error != nil {
		return "", false
	}
	value := row.ObjectKey
	if publicURL, err := storage.PublicURL(context.Background(), row.ObjectKey); err == nil && strings.TrimSpace(publicURL) != "" {
		value = publicURL
	}
	return value, value != ""
}

func normalizeH5NavigationLabels(values map[string]string, labelZh, labelEn string) map[string]string {
	out := map[string]string{}
	for key, value := range values {
		key = normalizeH5NavigationLanguage(key)
		value = strings.TrimSpace(value)
		if key != "" && value != "" && len(value) <= 64 {
			out[key] = value
		}
	}
	if value := strings.TrimSpace(labelZh); value != "" {
		out["zh-CN"] = value
	}
	if value := strings.TrimSpace(labelEn); value != "" {
		out["en-US"] = value
	}
	return out
}

func normalizeH5NavigationCodes(values []string, country bool) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if country {
			value = strings.ToUpper(value)
		} else {
			value = normalizeH5NavigationLanguage(value)
		}
		if value == "" || value == "ALL" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func normalizeH5NavigationLanguage(value string) string {
	value = strings.ReplaceAll(strings.TrimSpace(value), "_", "-")
	parts := strings.Split(value, "-")
	if len(parts) == 0 || parts[0] == "" {
		return ""
	}
	parts[0] = strings.ToLower(parts[0])
	if len(parts) > 1 {
		parts[1] = strings.ToUpper(parts[1])
	}
	return strings.Join(parts, "-")
}

func validInternalH5Target(value string) bool {
	return strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "//") &&
		!strings.ContainsAny(value, "\\<>\n\r") && !strings.Contains(strings.ToLower(value), "javascript:")
}

func validHTTPSURL(value string) bool {
	parsed, err := url.ParseRequestURI(value)
	return err == nil && strings.EqualFold(parsed.Scheme, "https") && parsed.Host != ""
}

func decodeH5NavigationDraft(raw string) H5NavigationDraft {
	var out H5NavigationDraft
	_ = json.Unmarshal([]byte(raw), &out)
	if out.Labels == nil {
		out.Labels = map[string]string{}
	}
	return out
}

func legacyPublishedH5Navigation(row model.GameH5NavigationItems) H5NavigationDraft {
	action, target := row.ActionType, row.ActionTarget
	if action == "PAGE" {
		action, target = "OPEN_INTERNAL_PAGE", row.PageRoute
	} else if action == "GAME" {
		action, target = "OPEN_GAME", row.ProviderGameCode
	}
	return H5NavigationDraft{
		MenuCode: row.MenuCode, Labels: map[string]string{"zh-CN": row.LabelZh, "en-US": row.LabelEn},
		IconAssetID: row.IconAssetID, LottieAssetID: row.LottieAssetID, SoundAssetID: row.SoundAssetID,
		DisplayPosition: row.DisplayPosition, SortOrder: row.SortOrder, BadgeType: row.BadgeType,
		VisibilityScope: row.VisibilityScope, StartAt: row.StartAt, EndAt: row.EndAt, Enabled: row.Enabled,
		ActionType: action, ActionTarget: target, ProviderCode: row.ProviderCode, GameID: row.GameID,
		ProviderGameCode: row.ProviderGameCode, GameName: row.GameName,
	}
}

func inputFromH5NavigationDraft(draft H5NavigationDraft, reason string) H5NavigationInput {
	return H5NavigationInput{
		MenuCode: draft.MenuCode, Labels: draft.Labels, IconAssetID: draft.IconAssetID,
		LottieAssetID: draft.LottieAssetID, SoundAssetID: draft.SoundAssetID,
		DisplayPosition: draft.DisplayPosition, SortOrder: draft.SortOrder, BadgeType: draft.BadgeType,
		VisibilityScope: draft.VisibilityScope, CountryCodes: draft.CountryCodes, LanguageCodes: draft.LanguageCodes,
		StartAt: draft.StartAt, EndAt: draft.EndAt, Enabled: draft.Enabled, ActionType: draft.ActionType,
		ActionTarget: draft.ActionTarget, ProviderCode: draft.ProviderCode, GameID: draft.GameID, Reason: reason,
	}
}

func h5NavigationVisible(draft H5NavigationDraft, language, country string, loggedIn bool, at time.Time) bool {
	if !draft.Enabled || draft.Deleted || (draft.StartAt != nil && draft.StartAt.After(at)) || (draft.EndAt != nil && draft.EndAt.Before(at)) {
		return false
	}
	if draft.VisibilityScope == "GUEST" && loggedIn {
		return false
	}
	if draft.VisibilityScope == "LOGGED_IN" && !loggedIn {
		return false
	}
	if len(draft.CountryCodes) > 0 && !containsH5NavigationCode(draft.CountryCodes, strings.ToUpper(strings.TrimSpace(country)), false) {
		return false
	}
	if len(draft.LanguageCodes) > 0 && !containsH5NavigationCode(draft.LanguageCodes, normalizeH5NavigationLanguage(language), true) {
		return false
	}
	return true
}

func containsH5NavigationCode(values []string, requested string, language bool) bool {
	for _, value := range values {
		if value == requested {
			return true
		}
		if language && strings.EqualFold(strings.Split(value, "-")[0], strings.Split(requested, "-")[0]) {
			return true
		}
	}
	return false
}

func h5NavigationLabel(labels map[string]string, language string) string {
	language = normalizeH5NavigationLanguage(language)
	if value := labels[language]; value != "" {
		return value
	}
	base := strings.Split(language, "-")[0]
	for key, value := range labels {
		if strings.EqualFold(strings.Split(key, "-")[0], base) && value != "" {
			return value
		}
	}
	if value := labels["en-US"]; value != "" {
		return value
	}
	return labels["zh-CN"]
}

func markH5NavigationSiteDraft(tx *gorm.DB, actor Actor) error {
	return tx.Table("game_saas_sites").Where("id=? AND tenant_id=? AND deleted_at IS NULL", actor.Scope.SiteID, actor.Scope.TenantID).
		Updates(map[string]any{"config_status": H5NavigationPublishDraft, "config_draft_updated_at": time.Now(), "updated_by": actor.ID, "updated_at": time.Now()}).Error
}

func PreviewH5NavigationDrafts(db *gorm.DB, tenantID, siteID uint64) ([]map[string]any, error) {
	var rows []model.GameH5NavigationItems
	if err := db.Where("tenant_id=? AND site_id=? AND deleted_at IS NULL", tenantID, siteID).Order("id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		draft := decodeH5NavigationDraft(row.DraftPayload)
		if draft.MenuCode == "" {
			draft = decodeH5NavigationDraft(row.PublishedPayload)
		}
		payload, _ := json.Marshal(draft)
		item := map[string]any{}
		if err := json.Unmarshal(payload, &item); err != nil {
			return nil, err
		}
		item["id"] = row.ID
		out = append(out, item)
	}
	return out, nil
}

func PublishH5NavigationDrafts(tx *gorm.DB, tenantID, siteID, actorID, configVersion uint64) (int, error) {
	var rows []model.GameH5NavigationItems
	if err := tx.Where("tenant_id=? AND site_id=? AND publish_status=? AND deleted_at IS NULL", tenantID, siteID, H5NavigationPublishDraft).
		Order("id ASC").Find(&rows).Error; err != nil {
		return 0, err
	}
	service := New(tx)
	scope := tenancy.Scope{TenantID: tenantID, SiteID: siteID}
	nowValue := time.Now()
	for _, row := range rows {
		draft := decodeH5NavigationDraft(row.DraftPayload)
		if draft.MenuCode == "" {
			return 0, newError("INVALID_H5_NAVIGATION_DRAFT", "navigation draft payload is invalid")
		}
		if draft.Deleted {
			if err := tx.Model(&model.GameH5NavigationItems{}).Where("id=? AND tenant_id=? AND site_id=?", row.ID, tenantID, siteID).
				Updates(map[string]any{"deleted_at": nowValue, "updated_by": actorID, "updated_at": nowValue}).Error; err != nil {
				return 0, err
			}
			continue
		}
		normalized, err := normalizeH5NavigationInput(tx, inputFromH5NavigationDraft(draft, "publish navigation entry"), scope)
		if err != nil {
			return 0, err
		}
		if err := service.validateH5NavigationAssets(tx, scope, normalized); err != nil {
			return 0, err
		}
		payload, _ := json.Marshal(normalized)
		labels, _ := json.Marshal(normalized.Labels)
		countries, _ := json.Marshal(normalized.CountryCodes)
		languages, _ := json.Marshal(normalized.LanguageCodes)
		updates := map[string]any{
			"label_zh": normalized.Labels["zh-CN"], "label_en": normalized.Labels["en-US"], "labels_json": string(labels),
			"icon_asset_id": normalized.IconAssetID, "lottie_asset_id": normalized.LottieAssetID, "sound_asset_id": normalized.SoundAssetID,
			"icon_asset_id_draft": normalized.IconAssetID, "lottie_asset_id_draft": normalized.LottieAssetID, "sound_asset_id_draft": normalized.SoundAssetID,
			"display_position": normalized.DisplayPosition, "sort_order": normalized.SortOrder, "badge_type": normalized.BadgeType,
			"visibility_scope": normalized.VisibilityScope, "country_codes": string(countries), "language_codes": string(languages),
			"start_at": normalized.StartAt, "end_at": normalized.EndAt, "enabled": normalized.Enabled,
			"action_type": normalized.ActionType, "action_target": normalized.ActionTarget, "page_route": "",
			"provider_code": normalized.ProviderCode, "game_id": normalized.GameID,
			"provider_game_code": normalized.ProviderGameCode, "game_name": normalized.GameName,
			"draft_payload": string(payload), "published_payload": string(payload), "publish_status": H5NavigationPublishPublished,
			"config_version": configVersion, "published_at": nowValue, "published_by": actorID,
			"updated_by": actorID, "updated_at": nowValue,
		}
		if err := tx.Model(&model.GameH5NavigationItems{}).Where("id=? AND tenant_id=? AND site_id=? AND deleted_at IS NULL", row.ID, tenantID, siteID).
			Updates(updates).Error; err != nil {
			return 0, err
		}
	}
	return len(rows), nil
}
