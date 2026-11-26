package systemadmin

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SiteMaintenanceInput struct {
	SiteID              uint64     `json:"siteID"`
	ScopeType           string     `json:"scopeType"`
	Enabled             bool       `json:"enabled"`
	StartAt             *time.Time `json:"startAt"`
	EndAt               *time.Time `json:"endAt"`
	Reason              string     `json:"reason"`
	Title               string     `json:"title"`
	MessageZh           string     `json:"messageZh"`
	MessageEn           string     `json:"messageEn"`
	EstimatedRecoveryAt *time.Time `json:"estimatedRecoveryAt"`
	SupportURL          string     `json:"supportURL"`
	SupportEmail        string     `json:"supportEmail"`
	SupportChat         string     `json:"supportChat"`
	AllowlistedIPs      []string   `json:"allowlistedIPs"`
	AllowAdminBypass    bool       `json:"allowAdminBypass"`
}

type ProviderMaintenanceInput struct {
	SiteID                uint64     `json:"siteID"`
	ProviderCode          string     `json:"providerCode"`
	Enabled               bool       `json:"enabled"`
	StartAt               *time.Time `json:"startAt"`
	EndAt                 *time.Time `json:"endAt"`
	Reason                string     `json:"reason"`
	MessageZh             string     `json:"messageZh"`
	MessageEn             string     `json:"messageEn"`
	AffectedCountries     []string   `json:"affectedCountries"`
	AffectedCurrencies    []string   `json:"affectedCurrencies"`
	AllowExistingSessions bool       `json:"allowExistingSessions"`
	HideFromLobby         bool       `json:"hideFromLobby"`
	BlockNewLaunches      bool       `json:"blockNewLaunches"`
}

type GameMaintenanceInput struct {
	SiteID                uint64     `json:"siteID"`
	GameID                uint64     `json:"gameID"`
	Enabled               bool       `json:"enabled"`
	StartAt               *time.Time `json:"startAt"`
	EndAt                 *time.Time `json:"endAt"`
	Reason                string     `json:"reason"`
	MessageZh             string     `json:"messageZh"`
	MessageEn             string     `json:"messageEn"`
	AffectedCountries     []string   `json:"affectedCountries"`
	AffectedCurrencies    []string   `json:"affectedCurrencies"`
	AllowExistingSessions bool       `json:"allowExistingSessions"`
	HideFromLobby         bool       `json:"hideFromLobby"`
	BlockNewLaunches      bool       `json:"blockNewLaunches"`
}

type siteMaintenanceRow struct {
	ID uint64 `json:"id"`

	SiteID              uint64     `json:"siteID"`
	ScopeType           string     `json:"scopeType"`
	Enabled             bool       `json:"enabled"`
	StartAt             *time.Time `json:"startAt"`
	EndAt               *time.Time `json:"endAt"`
	Reason              string     `json:"reason"`
	Title               string     `json:"title"`
	MessageZh           string     `json:"messageZh"`
	MessageEn           string     `json:"messageEn"`
	EstimatedRecoveryAt *time.Time `json:"estimatedRecoveryAt"`
	SupportURL          string     `json:"supportURL"`
	SupportEmail        string     `json:"supportEmail"`
	SupportChat         string     `json:"supportChat"`
	AllowlistedIPs      string     `json:"allowlistedIPs"`
	AllowAdminBypass    bool       `json:"allowAdminBypass"`
	CreatedBy           uint64     `json:"createdBy"`
	UpdatedBy           uint64     `json:"updatedBy"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

func (s *Service) ListSiteMaintenance(q Query, actor Actor) (Result, error) {
	rows := []siteMaintenanceRow{}
	db := s.db.Table("game_site_maintenance").Where("deleted_at IS NULL")
	db, err := operationalScope(db, actor, value(q, "siteID"))
	if err != nil {
		return Result{}, err
	}
	db = exact(db, "scope_type", strings.ToUpper(value(q, "scopeType")))
	db = exact(db, "enabled", value(q, "enabled"))
	return paginate(db.Order("site_id, scope_type"), q, &rows)
}

func (s *Service) SaveSiteMaintenance(input SiteMaintenanceInput, actor Actor) (uint64, error) {
	input.ScopeType = strings.ToUpper(strings.TrimSpace(input.ScopeType))
	if input.ScopeType != "SITE" {
		return 0, NewError("INVALID_MAINTENANCE_SCOPE", "scope must be SITE")
	}
	siteID, err := operationalTargetScope(s.db, actor, input.SiteID)
	if err != nil {
		return 0, err
	}
	if err := validateMaintenanceWindow(input.StartAt, input.EndAt, input.Reason); err != nil {
		return 0, err
	}
	if err := validateSupportURL(input.SupportURL); err != nil {
		return 0, err
	}
	if err := validateIPList(input.AllowlistedIPs); err != nil {
		return 0, err
	}
	var id uint64
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var before siteMaintenanceRow
		lookup := tx.Table("game_site_maintenance").Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("site_id = ? AND scope_type = ? AND deleted_at IS NULL", siteID, input.ScopeType)
		found := lookup.Take(&before).Error == nil
		row := map[string]any{
			"site_id": siteID, "scope_type": input.ScopeType, "enabled": boolValue(input.Enabled),
			"start_at": input.StartAt, "end_at": input.EndAt, "reason": strings.TrimSpace(input.Reason), "title": strings.TrimSpace(input.Title),
			"message_zh": strings.TrimSpace(input.MessageZh), "message_en": strings.TrimSpace(input.MessageEn),
			"estimated_recovery_at": input.EstimatedRecoveryAt, "support_url": strings.TrimSpace(input.SupportURL),
			"support_email": strings.TrimSpace(input.SupportEmail), "support_chat": strings.TrimSpace(input.SupportChat),
			"allowlisted_ips":    strings.Join(normalizeList(input.AllowlistedIPs), ","),
			"allow_admin_bypass": boolValue(input.AllowAdminBypass),
			"updated_by":         actor.ID, "updated_at": now(),
		}
		if found {
			id = before.ID
			if err := tx.Table("game_site_maintenance").Where("id=?", id).Updates(row).Error; err != nil {
				return err
			}
		} else {
			row["created_by"], row["created_at"] = actor.ID, now()
			if err := tx.Table("game_site_maintenance").Create(row).Error; err != nil {
				return err
			}
			if err := tx.Table("game_site_maintenance").Select("id").
				Where("site_id = ? AND scope_type = ? AND deleted_at IS NULL", siteID, input.ScopeType).
				Take(&id).Error; err != nil {
				return err
			}
		}
		return audit(tx, targetAuditActor(actor, siteID), "maintenance:site:write", "maintenance.site.save", "site_maintenance", id, before, row, input.Reason)
	})
	return id, err
}

func (s *Service) ListProviderMaintenance(q Query, actor Actor) (Result, error) {
	rows := []map[string]any{}
	db := s.db.Table("game_provider_maintenance").Where("deleted_at IS NULL")
	db = db.Where("site_id IN (0,1)")
	db = like(db, "provider_code", value(q, "providerCode"))
	db = exact(db, "enabled", value(q, "enabled"))
	return paginate(db.Order("site_id, provider_code"), q, &rows)
}

func (s *Service) SaveProviderMaintenance(input ProviderMaintenanceInput, actor Actor) (uint64, error) {
	providerCode := strings.ToUpper(strings.TrimSpace(input.ProviderCode))
	if providerCode == "" {
		return 0, NewError("PROVIDER_REQUIRED", "provider is required")
	}
	siteID := input.SiteID
	if actor.Platform && siteID == 0 {
		siteID = 0
	} else {
		var err error
		siteID, err = operationalTargetScope(s.db, actor, siteID)
		if err != nil {
			return 0, err
		}
	}
	if err := validateMaintenanceWindow(input.StartAt, input.EndAt, input.Reason); err != nil {
		return 0, err
	}
	var providerCount int64
	if err := s.db.Table("game_game_providers").Where("provider_code=? AND deleted_at IS NULL", providerCode).Count(&providerCount).Error; err != nil || providerCount == 0 {
		return 0, NewError("PROVIDER_NOT_FOUND", "provider not found")
	}
	row := map[string]any{
		"site_id": siteID, "scope_type": map[bool]string{true: "GLOBAL", false: "SITE"}[siteID == 0],
		"provider_code": providerCode, "enabled": boolValue(input.Enabled), "start_at": input.StartAt, "end_at": input.EndAt,
		"reason": strings.TrimSpace(input.Reason), "message_zh": strings.TrimSpace(input.MessageZh), "message_en": strings.TrimSpace(input.MessageEn),
		"affected_countries": strings.Join(normalizeList(input.AffectedCountries), ","), "affected_currencies": strings.Join(normalizeList(input.AffectedCurrencies), ","),
		"allow_existing_sessions": boolValue(input.AllowExistingSessions), "hide_from_lobby": boolValue(input.HideFromLobby), "block_new_launches": boolValue(input.BlockNewLaunches),
		"updated_by": actor.ID, "updated_at": now(),
	}
	return s.upsertMaintenance("game_provider_maintenance", "provider_code", providerCode, siteID, row, input.Reason, "maintenance:provider:write", actor)
}

func (s *Service) ListGameMaintenance(q Query, actor Actor) (Result, error) {
	rows := []map[string]any{}
	db := s.db.Table("game_game_maintenance m").Select("m.*,g.name game_name,g.game_code,p.provider_code").
		Joins("JOIN game_games AS g ON g.id = m.game_id AND g.site_id = m.site_id AND g.deleted_at IS NULL").
		Joins("JOIN game_game_providers p ON p.id=g.provider_id AND p.deleted_at IS NULL").Where("m.deleted_at IS NULL")
	var err error
	db, err = operationalScopeAlias(db, "m", actor, value(q, "siteID"))
	if err != nil {
		return Result{}, err
	}
	db = exact(db, "m.enabled", value(q, "enabled"))
	db = exact(db, "p.provider_code", value(q, "providerCode"))
	return paginate(db.Order("m.site_id, m.game_id"), q, &rows)
}

func (s *Service) SaveGameMaintenance(input GameMaintenanceInput, actor Actor) (uint64, error) {
	siteID, err := operationalTargetScope(s.db, actor, input.SiteID)
	if err != nil {
		return 0, err
	}
	if input.GameID == 0 || siteID == 0 {
		return 0, NewError("GAME_SITE_REQUIRED", "game and site are required")
	}
	if err := validateMaintenanceWindow(input.StartAt, input.EndAt, input.Reason); err != nil {
		return 0, err
	}
	var gameCount int64
	if err := s.db.Table("game_games").Where("site_id = ? AND id = ? AND deleted_at IS NULL", siteID, input.GameID).Count(&gameCount).Error; err != nil || gameCount == 0 {
		return 0, NewError("GAME_NOT_AUTHORIZED", "game is not available to this site")
	}
	row := map[string]any{
		"site_id": siteID, "game_id": input.GameID, "enabled": boolValue(input.Enabled), "start_at": input.StartAt, "end_at": input.EndAt,
		"reason": strings.TrimSpace(input.Reason), "message_zh": strings.TrimSpace(input.MessageZh), "message_en": strings.TrimSpace(input.MessageEn),
		"affected_countries": strings.Join(normalizeList(input.AffectedCountries), ","), "affected_currencies": strings.Join(normalizeList(input.AffectedCurrencies), ","),
		"allow_existing_sessions": boolValue(input.AllowExistingSessions), "hide_from_lobby": boolValue(input.HideFromLobby), "block_new_launches": boolValue(input.BlockNewLaunches),
		"updated_by": actor.ID, "updated_at": now(),
	}
	return s.upsertMaintenance("game_game_maintenance", "game_id", input.GameID, siteID, row, input.Reason, "maintenance:game:write", actor)
}

func (s *Service) upsertMaintenance(table string, key string, keyValue any, siteID uint64, row map[string]any, reason string, permission string, actor Actor) (uint64, error) {
	var id uint64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var before map[string]any
		found := tx.Table(table).Clauses(clause.Locking{Strength: "UPDATE"}).Where("site_id=? AND "+key+"=? AND deleted_at IS NULL", siteID, keyValue).Take(&before).Error == nil
		if found {
			id, _ = anyUint64(before["id"])
			if err := tx.Table(table).Where("id=?", id).Updates(row).Error; err != nil {
				return err
			}
		} else {
			row["created_by"], row["created_at"] = actor.ID, now()
			if err := tx.Table(table).Create(row).Error; err != nil {
				return err
			}
			if err := tx.Table(table).Select("id").
				Where("site_id=? AND "+key+"=? AND deleted_at IS NULL", siteID, keyValue).
				Take(&id).Error; err != nil {
				return err
			}
		}
		return audit(tx, targetAuditActor(actor, siteID), permission, "maintenance.save", table, id, before, row, reason)
	})
	return id, err
}

func operationalTargetScope(db *gorm.DB, actor Actor, requestedSiteID uint64) (uint64, error) {
	return 1, nil
}
func operationalScope(db *gorm.DB, actor Actor, siteValue string) (*gorm.DB, error) {
	return operationalScopeAlias(db, "", actor, siteValue)
}
func operationalScopeAlias(db *gorm.DB, alias string, actor Actor, siteValue string) (*gorm.DB, error) {
	prefix := ""
	if alias != "" {
		prefix = alias + "."
	}
	return db.Where(prefix + "site_id IN (0,1)"), nil
}

func targetAuditActor(actor Actor, siteID uint64) Actor {
	actor.SiteID = siteID
	return actor
}

func validateMaintenanceWindow(startAt *time.Time, endAt *time.Time, reason string) error {
	if strings.TrimSpace(reason) == "" {
		return NewError("REASON_REQUIRED", "reason is required")
	}
	if startAt != nil && endAt != nil && !endAt.After(*startAt) {
		return NewError("INVALID_TIME_RANGE", "end time must be after start time")
	}
	return nil
}

func validateSupportURL(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return NewError("INVALID_SUPPORT_URL", "support URL must use https")
	}
	return nil
}

func validateIPList(values []string) error {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if net.ParseIP(value) == nil {
			if _, _, err := net.ParseCIDR(value); err != nil {
				return NewError("INVALID_IP_ALLOWLIST", "allowlist contains an invalid IP or CIDR")
			}
		}
	}
	return nil
}

func normalizeList(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.ToUpper(strings.TrimSpace(value))
		if value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}

func anyUint64(value any) (uint64, bool) {
	switch typed := value.(type) {
	case uint64:
		return typed, true
	case int64:
		return uint64(typed), typed >= 0
	case int:
		return uint64(typed), typed >= 0
	case []byte:
		var result uint64
		_, err := fmt.Sscan(string(typed), &result)
		return result, err == nil
	}
	return 0, false
}
