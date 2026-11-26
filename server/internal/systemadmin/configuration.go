package systemadmin

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"casino/internal/storage"
)

const maskedSecret = "******"

type ConfigRow struct {
	Key             string          `json:"key" gorm:"column:config_key"`
	NameZh          string          `json:"nameZh" gorm:"column:name_zh"`
	NameEn          string          `json:"nameEn" gorm:"column:name_en"`
	Category        string          `json:"category"`
	Value           string          `json:"value" gorm:"column:config_value"`
	ValueType       string          `json:"valueType" gorm:"column:value_type"`
	DescriptionZh   string          `json:"descriptionZh" gorm:"column:description_zh"`
	DescriptionEn   string          `json:"descriptionEn" gorm:"column:description_en"`
	Editable        bool            `json:"editable"`
	Sensitive       bool            `json:"sensitive"`
	RestartRequired bool            `json:"restartRequired" gorm:"column:restart_required"`
	AllowedValues   json.RawMessage `json:"allowedValues" gorm:"column:allowed_values"`
	Enabled         bool            `json:"enabled" gorm:"column:status"`
	UpdatedAt       time.Time       `json:"updatedAt" gorm:"column:updated_at"`
}

func (s *Service) ListConfigs(q Query) (Result, error) {
	rows := []ConfigRow{}
	db := s.db.Table("game_system_configs").Select("config_key,name_zh,name_en,category,config_value,value_type,description_zh,description_en,editable,`sensitive`,restart_required,allowed_values,status,updated_at").Where("deleted_at IS NULL")
	if strings.EqualFold(value(q, "adminScope"), "PLATFORM") {
		db = db.Where("config_key NOT LIKE 'h5.%' AND category NOT IN ('OPERATION','USER','GAME')")
	}
	db = like(db, "config_key", value(q, "key"))
	db = exact(db, "category", value(q, "category"))
	db = exact(db, "value_type", value(q, "valueType"))
	db = exact(db, "status", value(q, "status"))
	result, err := paginate(db.Order("category,config_key"), q, &rows)
	if err != nil {
		return Result{}, err
	}
	for index := range rows {
		if rows[index].Sensitive {
			rows[index].Value = maskedSecret
		}
	}
	result.List = rows
	return result, nil
}

func (s *Service) PlatformBranding(ctx context.Context) (map[string]any, error) {
	rows := []struct {
		Key   string `gorm:"column:config_key"`
		Value string `gorm:"column:config_value"`
	}{}
	keys := []string{"basic.platform_name", "basic.platform_logo_url", "basic.platform_contact", "finance.platform_usdt_address"}
	if err := s.db.Table("game_system_configs").Select("config_key,config_value").Where("config_key IN ? AND deleted_at IS NULL AND status=1", keys).Find(&rows).Error; err != nil {
		return nil, err
	}
	values := map[string]string{}
	for _, row := range rows {
		values[row.Key] = strings.TrimSpace(row.Value)
	}
	logoURL, err := storage.PublicURL(ctx, values["basic.platform_logo_url"])
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"platformName":        defaultString(values["basic.platform_name"], "Casino Admin"),
		"platformLogoUrl":     logoURL,
		"platformContact":     values["basic.platform_contact"],
		"platformUSDTAddress": values["finance.platform_usdt_address"],
	}, nil
}
func validateConfigValue(row ConfigRow, value string) error {
	switch row.ValueType {
	case "STRING", "SECRET":
	case "NUMBER":
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return NewError("INVALID_CONFIG_VALUE", "configuration value must be a number")
		}
	case "BOOLEAN":
		if value != "true" && value != "false" {
			return NewError("INVALID_CONFIG_VALUE", "configuration value must be boolean")
		}
	case "JSON":
		if !json.Valid([]byte(value)) {
			return NewError("INVALID_CONFIG_VALUE", "configuration value must be valid JSON")
		}
	case "ENUM":
		var values []string
		if len(row.AllowedValues) == 0 || json.Unmarshal(row.AllowedValues, &values) != nil {
			return NewError("INVALID_CONFIG_SCHEMA", "enum configuration has no allowed values")
		}
		found := false
		for _, candidate := range values {
			if candidate == value {
				found = true
				break
			}
		}
		if !found {
			return NewError("INVALID_CONFIG_VALUE", "configuration value is not allowed")
		}
	default:
		return NewError("INVALID_CONFIG_TYPE", "unsupported configuration type")
	}
	return nil
}
func (s *Service) UpdateConfig(key string, input ConfigInput, secretPermission bool, actor Actor) error {
	var row ConfigRow
	if err := s.db.Table("game_system_configs").Select("config_key,name_zh,name_en,category,config_value,value_type,description_zh,description_en,editable,`sensitive`,restart_required,allowed_values,status,updated_at").Where("config_key=? AND deleted_at IS NULL", key).Take(&row).Error; err != nil {
		return rowNotFound("configuration")
	}
	if !row.Editable {
		return NewError("CONFIG_NOT_EDITABLE", "configuration is not editable")
	}
	if row.Sensitive && !secretPermission {
		return NewError("SECRET_PERMISSION_REQUIRED", "secret update permission is required")
	}
	if row.Sensitive && strings.TrimSpace(input.Value) == maskedSecret {
		return NewError("SECRET_REPLACEMENT_REQUIRED", "provide a replacement secret value")
	}
	if err := validateConfigValue(row, input.Value); err != nil {
		return err
	}
	before := map[string]any{"value": row.Value, "sensitive": row.Sensitive}
	after := map[string]any{"value": input.Value, "sensitive": row.Sensitive}
	if row.Sensitive {
		before["value"], after["value"] = maskedSecret, maskedSecret
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_system_configs").Where("config_key=?", key).Updates(map[string]any{"config_value": input.Value, "updated_at": now()}).Error; err != nil {
			return err
		}
		permission := "system:config:update"
		if row.Sensitive {
			permission = "system:config:secret-update"
		}
		return audit(tx, actor, permission, "system.config.update", "system_config", 0, before, after, input.Reason)
	})
}
