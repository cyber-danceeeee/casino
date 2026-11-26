package operationadmin

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/tenancy"
)

const (
	registerRewardEnabledKey         = "operation.register_reward_enabled"
	registerRewardMinAmountKey       = "operation.register_reward_min_amount"
	registerRewardMaxAmountKey       = "operation.register_reward_max_amount"
	registerRewardAuditMultiplierKey = "operation.register_reward_audit_multiplier"
)

type RegisterRewardConfig struct {
	Enabled         bool       `json:"enabled"`
	MinAmount       string     `json:"minAmount"`
	MaxAmount       string     `json:"maxAmount"`
	AuditMultiplier string     `json:"auditMultiplier"`
	UpdatedAt       *time.Time `json:"updatedAt"`
}

type RegisterRewardConfigInput struct {
	Enabled         bool   `json:"enabled"`
	MinAmount       string `json:"minAmount" binding:"required"`
	MaxAmount       string `json:"maxAmount" binding:"required"`
	AuditMultiplier string `json:"auditMultiplier" binding:"required"`
	Reason          string `json:"reason" binding:"max=500"`
}

type systemConfigRow struct {
	TenantID    uint64     `gorm:"column:tenant_id"`
	SiteID      uint64     `gorm:"column:site_id"`
	ConfigKey   string     `gorm:"column:config_key"`
	ConfigValue string     `gorm:"column:config_value"`
	UpdatedAt   *time.Time `gorm:"column:updated_at"`
}

func (s *Service) GetRegisterRewardConfig(scope tenancy.Scope) (RegisterRewardConfig, error) {
	rows, err := s.systemConfigRows([]string{
		registerRewardEnabledKey,
		registerRewardMinAmountKey,
		registerRewardMaxAmountKey,
		registerRewardAuditMultiplierKey,
	}, scope)
	if err != nil {
		return RegisterRewardConfig{}, err
	}
	cfg := RegisterRewardConfig{
		Enabled:         true,
		MinAmount:       "1.00",
		MaxAmount:       "20.00",
		AuditMultiplier: "1.0",
	}
	for _, row := range rows {
		switch row.ConfigKey {
		case registerRewardEnabledKey:
			cfg.Enabled = parseConfigBool(row.ConfigValue, true)
		case registerRewardMinAmountKey:
			cfg.MinAmount = configDecimalString(row.ConfigValue, "1.00", 2)
		case registerRewardMaxAmountKey:
			cfg.MaxAmount = configDecimalString(row.ConfigValue, "20.00", 2)
		case registerRewardAuditMultiplierKey:
			cfg.AuditMultiplier = configDecimalString(row.ConfigValue, "1.0", 1)
		}
		if row.UpdatedAt != nil && (cfg.UpdatedAt == nil || row.UpdatedAt.After(*cfg.UpdatedAt)) {
			cfg.UpdatedAt = row.UpdatedAt
		}
	}
	return cfg, nil
}

func (s *Service) SaveRegisterRewardConfig(in RegisterRewardConfigInput, actor Actor) error {
	minAmount, err := parseDecimal(in.MinAmount, "minAmount")
	if err != nil {
		return err
	}
	maxAmount, err := parseDecimal(in.MaxAmount, "maxAmount")
	if err != nil {
		return err
	}
	auditMultiplier, err := parseDecimal(in.AuditMultiplier, "auditMultiplier")
	if err != nil {
		return err
	}
	if maxAmount.LessThan(minAmount) {
		return newError("INVALID_AMOUNT_RANGE", "max amount must be greater than or equal to min amount")
	}
	before, err := s.GetRegisterRewardConfig(actor.Scope)
	if err != nil {
		return err
	}
	after := RegisterRewardConfig{
		Enabled:         in.Enabled,
		MinAmount:       minAmount.StringFixed(2),
		MaxAmount:       maxAmount.StringFixed(2),
		AuditMultiplier: auditMultiplier.StringFixed(1),
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		values := map[string]string{
			registerRewardEnabledKey:         boolConfigValue(in.Enabled),
			registerRewardMinAmountKey:       after.MinAmount,
			registerRewardMaxAmountKey:       after.MaxAmount,
			registerRewardAuditMultiplierKey: after.AuditMultiplier,
		}
		for key, value := range values {
			if err := upsertRegisterRewardConfig(tx, actor.Scope, key, value); err != nil {
				return err
			}
		}
		return audit(tx, actor, "operation:register-reward:update", "update", "operation_register_reward", 0, before, after, strings.TrimSpace(in.Reason))
	})
}

func (s *Service) systemConfigRows(keys []string, scope tenancy.Scope) ([]systemConfigRow, error) {
	var rows []systemConfigRow
	err := tenancy.ApplySite(s.db.Table("game_system_configs"), "", scope).
		Select("tenant_id,site_id,config_key,config_value,updated_at").
		Where("config_key IN ? AND deleted_at IS NULL", keys).
		Order("site_id ASC,id ASC").
		Find(&rows).Error
	return rows, err
}

func upsertRegisterRewardConfig(tx *gorm.DB, scope tenancy.Scope, key, value string) error {
	meta := registerRewardConfigMeta(key)
	now := time.Now()
	values := map[string]any{
		"config_value":   value,
		"name_zh":        meta.nameZh,
		"name_en":        meta.nameEn,
		"category":       "OPERATION",
		"value_type":     meta.valueType,
		"description":    meta.descriptionEn,
		"description_zh": meta.descriptionZh,
		"description_en": meta.descriptionEn,
		"editable":       1,
		"sensitive":      0,
		"status":         1,
		"updated_at":     now,
	}
	result := tenancy.ApplySiteStrict(tx.Table("game_system_configs"), "", scope).
		Where("config_key = ? AND deleted_at IS NULL", key).
		Updates(values)
	if result.Error != nil || result.RowsAffected > 0 {
		return result.Error
	}
	insert := map[string]any{
		"tenant_id":        scope.TenantID,
		"site_id":          scope.SiteID,
		"config_key":       key,
		"name_zh":          meta.nameZh,
		"name_en":          meta.nameEn,
		"category":         "OPERATION",
		"config_value":     value,
		"value_type":       meta.valueType,
		"description":      meta.descriptionEn,
		"description_zh":   meta.descriptionZh,
		"description_en":   meta.descriptionEn,
		"editable":         1,
		"sensitive":        0,
		"restart_required": 0,
		"status":           1,
		"created_at":       now,
		"updated_at":       now,
	}
	return tx.Table("game_system_configs").
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(insert).Error
}

type registerRewardMeta struct {
	nameZh        string
	nameEn        string
	valueType     string
	descriptionZh string
	descriptionEn string
}

func registerRewardConfigMeta(key string) registerRewardMeta {
	switch key {
	case registerRewardEnabledKey:
		return registerRewardMeta{"注册奖励开关", "Register Reward Enabled", "BOOLEAN", "是否启用 H5 注册奖励", "Enable H5 register reward"}
	case registerRewardMinAmountKey:
		return registerRewardMeta{"注册奖励最小金额", "Register Reward Min Amount", "NUMBER", "注册后随机现金奖励最小金额", "Minimum random cash reward after registration"}
	case registerRewardMaxAmountKey:
		return registerRewardMeta{"注册奖励最大金额", "Register Reward Max Amount", "NUMBER", "注册后随机现金奖励最大金额", "Maximum random cash reward after registration"}
	case registerRewardAuditMultiplierKey:
		return registerRewardMeta{"注册奖励稽核倍数", "Register Reward Audit Multiplier", "NUMBER", "注册奖励提现前需要完成的有效投注倍数", "Valid-bet multiplier required before withdrawing register reward"}
	default:
		return registerRewardMeta{"注册奖励配置", "Register Reward Config", "STRING", "注册奖励配置", "Register reward configuration"}
	}
}

func parseConfigBool(value string, fallback bool) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value == "1" || strings.EqualFold(value, "true") || strings.EqualFold(value, "yes")
}

func boolConfigValue(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func configDecimalString(value, fallback string, places int32) string {
	parsed, err := decimal.NewFromString(strings.TrimSpace(value))
	if err != nil || parsed.IsNegative() {
		parsed, _ = decimal.NewFromString(fallback)
	}
	return parsed.StringFixed(places)
}
