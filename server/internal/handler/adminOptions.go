package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"gorm.io/gorm"

	"casino/internal/database"
	"casino/internal/tenancy"
)

type AdminOptionsHandler interface {
	List(*gin.Context)
}

type adminOptionsHandler struct{}

func NewAdminOptionsHandler() AdminOptionsHandler {
	return &adminOptionsHandler{}
}

type adminOption struct {
	Label           string `json:"label" gorm:"column:label"`
	Value           string `json:"value" gorm:"column:value"`
	LabelZh         string `json:"labelZh,omitempty" gorm:"column:label_zh"`
	LabelEn         string `json:"labelEn,omitempty" gorm:"column:label_en"`
	CurrencyCode    string `json:"currencyCode,omitempty" gorm:"column:currency_code"`
	DefaultLanguage string `json:"defaultLanguage,omitempty" gorm:"column:default_language"`
	Timezone        string `json:"timezone,omitempty" gorm:"column:timezone"`
}

func (h *adminOptionsHandler) List(c *gin.Context) {
	db := database.GetDB()
	scope := adminTenancyScope(c)
	data := gin.H{
		"providers":            adminOptionRows(db, scope, "game_game_providers", "provider_code", "name", "deleted_at IS NULL", "sort ASC, provider_code ASC"),
		"sites":                adminOptionRows(db, scope, "game_saas_sites", "id", "site_name", "deleted_at IS NULL", "site_name ASC, id ASC"),
		"games":                adminOptionRows(db, scope, "game_games", "id", "name", "deleted_at IS NULL", "sort ASC, id ASC"),
		"gameCategories":       adminOptionRows(db, scope, "game_game_categories", "category_key", "name_zh", "deleted_at IS NULL", "sort ASC, category_key ASC"),
		"gameTags":             adminOptionRows(db, scope, "game_game_tags", "code", "name_zh", "deleted_at IS NULL", "sort_order ASC, code ASC"),
		"gameCollections":      adminOptionRows(db, scope, "game_game_collections", "code", "name_zh", "deleted_at IS NULL", "sort_order ASC, code ASC"),
		"countries":            adminCountryOptionRows(db),
		"currencies":           adminGlobalDistinctOptionRows(db, "game_country_configs", "currency_code", "currency_code <> '' AND status=1 AND deleted_at IS NULL", "currency_code ASC"),
		"phoneCountryCodes":    adminGlobalDistinctOptionRows(db, "game_country_configs", "phone_country_code", "phone_country_code <> '' AND status=1 AND deleted_at IS NULL", "phone_country_code ASC"),
		"languages":            adminGlobalDistinctOptionRows(db, "game_country_configs", "default_language", "default_language <> '' AND status=1 AND deleted_at IS NULL", "default_language ASC"),
		"timezones":            adminGlobalDistinctOptionRows(db, "game_country_configs", "timezone", "timezone <> '' AND status=1 AND deleted_at IS NULL", "timezone ASC"),
		"gameTypes":            adminDistinctOptionRows(db, scope, "game_games", "game_type", "game_type <> '' AND deleted_at IS NULL", "game_type ASC"),
		"providerCategories":   adminDistinctOptionRows(db, scope, "game_games", "category", "category <> '' AND deleted_at IS NULL", "category ASC"),
		"devices":              adminDistinctOptionRows(db, scope, "game_game_launch_logs", "device", "device <> '' AND deleted_at IS NULL", "device ASC"),
		"registrationChannels": adminDistinctOptionRows(db, scope, "game_users", "registration_channel", "registration_channel <> '' AND deleted_at IS NULL", "registration_channel ASC"),
		"vipLevels":            adminNumberOptionRows(db, scope, "game_vip_levels", "level_no", "deleted_at IS NULL", "level_no ASC"),
		"agentLevels":          adminNumberOptionRows(db, scope, "game_agent_commission_rules", "level_no", "deleted_at IS NULL", "level_no ASC"),
		"commissionPlans":      adminOptionRows(db, scope, "game_commission_plans", "code", "name_zh", "deleted_at IS NULL", "code ASC"),
		"paymentMethods":       adminOptionRows(db, scope, "game_payment_methods", "method_code", "name", "deleted_at IS NULL", "sort ASC, method_code ASC"),
		"paymentChannels":      adminOptionRows(db, scope, "game_payment_channels", "channel_code", "name", "deleted_at IS NULL", "sort ASC, channel_code ASC"),
		"withdrawAccountTypes": []adminOption{{Label: "bank", Value: "bank"}, {Label: "wallet", Value: "wallet"}},
		"booleanValues":        []adminOption{{Label: "true", Value: "true"}, {Label: "false", Value: "false"}},
		"providerEnvironments": []adminOption{{Label: "production", Value: "production"}, {Label: "sandbox", Value: "sandbox"}},
		"providerStatuses":     []adminOption{{Label: "ENABLED", Value: "ENABLED"}, {Label: "DISABLED", Value: "DISABLED"}, {Label: "MAINTENANCE", Value: "MAINTENANCE"}},
		"launchTypes":          []adminOption{{Label: "REAL", Value: "REAL"}, {Label: "DEMO", Value: "DEMO"}},
		"launchStatuses":       []adminOption{{Label: "LAUNCH_REQUESTED", Value: "LAUNCH_REQUESTED"}, {Label: "LAUNCH_SUCCESS", Value: "LAUNCH_SUCCESS"}, {Label: "LAUNCH_FAILED", Value: "LAUNCH_FAILED"}, {Label: "PLAYING", Value: "PLAYING"}, {Label: "CLOSED", Value: "CLOSED"}},
	}
	if scope.Platform {
		data["tenants"] = adminOptionRows(db, scope, "game_saas_tenants", "id", "tenant_name", "deleted_at IS NULL", "tenant_name ASC,id ASC")
	}
	response.Success(c, data)
}

func adminCountryOptionRows(db *gorm.DB) []adminOption {
	rows := []adminOption{}
	_ = db.Table("game_country_configs").
		Select(`country_code AS value,
			COALESCE(NULLIF(country_name_zh,''),NULLIF(country_name_en,''),country_code) AS label,
			country_name_zh AS label_zh,country_name_en AS label_en,currency_code,default_language,timezone`).
		Where("status=1 AND deleted_at IS NULL").
		Order("is_default DESC,country_code ASC").
		Scan(&rows).Error
	return adminCleanOptions(rows)
}

func adminGlobalDistinctOptionRows(db *gorm.DB, table, column, where, order string) []adminOption {
	rows := []adminOption{}
	query := db.Table(table).Select("DISTINCT " + column + " AS value, " + column + " AS label")
	if strings.TrimSpace(where) != "" {
		query = query.Where(where)
	}
	if strings.TrimSpace(order) != "" {
		query = query.Order(order)
	}
	_ = query.Scan(&rows).Error
	return adminCleanOptions(rows)
}

func adminOptionRows(db *gorm.DB, scope tenancy.Scope, table, valueColumn, labelColumn, where, order string) []adminOption {
	rows := []adminOption{}
	query := db.Table(table).Select(valueColumn + " AS value, COALESCE(NULLIF(" + labelColumn + ", ''), " + valueColumn + ") AS label")
	query = tenancy.ApplySite(query, "", scope)
	if strings.TrimSpace(where) != "" {
		query = query.Where(where)
	}
	if strings.TrimSpace(order) != "" {
		query = query.Order(order)
	}
	_ = query.Scan(&rows).Error
	return adminCleanOptions(rows)
}

func adminDistinctOptionRows(db *gorm.DB, scope tenancy.Scope, table, column, where, order string) []adminOption {
	rows := []adminOption{}
	query := db.Table(table).Select("DISTINCT " + column + " AS value, " + column + " AS label")
	query = tenancy.ApplySite(query, "", scope)
	if strings.TrimSpace(where) != "" {
		query = query.Where(where)
	}
	if strings.TrimSpace(order) != "" {
		query = query.Order(order)
	}
	_ = query.Scan(&rows).Error
	return adminCleanOptions(rows)
}

func adminNumberOptionRows(db *gorm.DB, scope tenancy.Scope, table, column, where, order string) []adminOption {
	rows := []adminOption{}
	query := db.Table(table).Select("DISTINCT CAST(" + column + " AS CHAR) AS value, CAST(" + column + " AS CHAR) AS label")
	query = tenancy.ApplySite(query, "", scope)
	if strings.TrimSpace(where) != "" {
		query = query.Where(where)
	}
	if strings.TrimSpace(order) != "" {
		query = query.Order(order)
	}
	_ = query.Scan(&rows).Error
	return adminCleanOptions(rows)
}

func adminCleanOptions(rows []adminOption) []adminOption {
	seen := map[string]bool{}
	out := make([]adminOption, 0, len(rows))
	for _, row := range rows {
		row.Value = strings.TrimSpace(row.Value)
		row.Label = strings.TrimSpace(row.Label)
		row.LabelZh = strings.TrimSpace(row.LabelZh)
		row.LabelEn = strings.TrimSpace(row.LabelEn)
		row.CurrencyCode = strings.ToUpper(strings.TrimSpace(row.CurrencyCode))
		row.DefaultLanguage = strings.TrimSpace(row.DefaultLanguage)
		row.Timezone = strings.TrimSpace(row.Timezone)
		if row.Value == "" || seen[row.Value] {
			continue
		}
		if row.Label == "" {
			row.Label = row.Value
		}
		seen[row.Value] = true
		out = append(out, row)
	}
	return out
}
