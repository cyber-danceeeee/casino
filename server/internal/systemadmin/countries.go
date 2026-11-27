package systemadmin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"

	"casino/internal/storage"
)

type countryConfigRecord struct {
	ID                    uint64    `gorm:"column:id"`
	CountryCode           string    `gorm:"column:country_code"`
	ISOCountryCode        string    `gorm:"column:iso_country_code"`
	CountryNameZh         string    `gorm:"column:country_name_zh"`
	CountryNameEn         string    `gorm:"column:country_name_en"`
	CurrencyCode          string    `gorm:"column:currency_code"`
	CurrencyNameZh        string    `gorm:"column:currency_name_zh"`
	CurrencyNameEn        string    `gorm:"column:currency_name_en"`
	CurrencySymbol        string    `gorm:"column:currency_symbol"`
	CurrencyDecimalPlaces int       `gorm:"column:currency_decimal_places"`
	Timezone              string    `gorm:"column:timezone"`
	DefaultLanguage       string    `gorm:"column:default_language"`
	SupportedLanguages    string    `gorm:"column:supported_languages"`
	PhoneCountryCode      string    `gorm:"column:phone_country_code"`
	PhoneLength           int       `gorm:"column:phone_length"`
	BrandingConfig        string    `gorm:"column:branding_config"`
	ThemeConfig           string    `gorm:"column:theme_config"`
	Enabled               bool      `gorm:"column:status"`
	IsDefault             bool      `gorm:"column:is_default"`
	CreatedAt             time.Time `gorm:"column:created_at"`
	UpdatedAt             time.Time `gorm:"column:updated_at"`
}

type CountryRow struct {
	ID                    uint64    `json:"id"`
	CountryCode           string    `json:"countryCode"`
	ISOCountryCode        string    `json:"isoCountryCode"`
	CountryNameZh         string    `json:"countryNameZh"`
	CountryNameEn         string    `json:"countryNameEn"`
	CountryName           string    `json:"countryName"`
	CurrencyCode          string    `json:"currencyCode"`
	CurrencyNameZh        string    `json:"currencyNameZh"`
	CurrencyNameEn        string    `json:"currencyNameEn"`
	CurrencySymbol        string    `json:"currencySymbol"`
	CurrencyDecimalPlaces int       `json:"currencyDecimalPlaces"`
	Timezone              string    `json:"timezone"`
	DefaultLanguage       string    `json:"defaultLanguage"`
	SupportedLanguages    string    `json:"supportedLanguages"`
	PhoneCountryCode      string    `json:"phoneCountryCode"`
	PhoneLength           int       `json:"phoneLength"`
	Enabled               bool      `json:"enabled"`
	Status                bool      `json:"status"`
	IsDefault             bool      `json:"isDefault"`
	LogoURL               string    `json:"logoUrl"`
	FaviconURL            string    `json:"faviconUrl"`
	LoginLogoURL          string    `json:"loginLogoUrl"`
	HeaderLogoURL         string    `json:"headerLogoUrl"`
	LoadingLogoURL        string    `json:"loadingLogoUrl"`
	AppIconURL            string    `json:"appIconUrl"`
	LoadingAnimationURL   string    `json:"loadingAnimationUrl"`
	LoadingBackgroundURL  string    `json:"loadingBackgroundUrl"`
	PrimaryColor          string    `json:"primaryColor"`
	SecondaryColor        string    `json:"secondaryColor"`
	ButtonColor           string    `json:"buttonColor"`
	BackgroundColor       string    `json:"backgroundColor"`
	TextColor             string    `json:"textColor"`
	CreatedAt             time.Time `json:"createdAt"`
	UpdatedAt             time.Time `json:"updatedAt"`
}

type countryBranding struct {
	LogoURL              string `json:"logoUrl,omitempty"`
	FaviconURL           string `json:"faviconUrl,omitempty"`
	LoginLogoURL         string `json:"loginLogoUrl,omitempty"`
	HeaderLogoURL        string `json:"headerLogoUrl,omitempty"`
	LoadingLogoURL       string `json:"loadingLogoUrl,omitempty"`
	AppIconURL           string `json:"appIconUrl,omitempty"`
	LoadingAnimationURL  string `json:"loadingAnimationUrl,omitempty"`
	LoadingBackgroundURL string `json:"loadingBackgroundUrl,omitempty"`
}

type countryTheme struct {
	PrimaryColor    string `json:"primaryColor,omitempty"`
	SecondaryColor  string `json:"secondaryColor,omitempty"`
	ButtonColor     string `json:"buttonColor,omitempty"`
	BackgroundColor string `json:"backgroundColor,omitempty"`
	TextColor       string `json:"textColor,omitempty"`
}

func (s *Service) ListCountries(q Query) (Result, error) {
	records := []countryConfigRecord{}
	db := s.db.Table("game_country_configs").Where("deleted_at IS NULL")
	db = like(db, "country_code", strings.ToUpper(value(q, "countryCode")))
	name := value(q, "countryName")
	if name != "" {
		db = db.Where("(country_name_zh LIKE ? OR country_name_en LIKE ?)", "%"+name+"%", "%"+name+"%")
	}
	db = exact(db, "currency_code", strings.ToUpper(value(q, "currencyCode")))
	db = exact(db, "status", value(q, "status"))
	db = exact(db, "is_default", value(q, "isDefault"))
	db = timeRange(db, "created_at", value(q, "createdFrom"), value(q, "createdTo"))
	result, err := paginate(db.Order("is_default DESC,country_code"), q, &records)
	if err != nil {
		return Result{}, err
	}
	rows := make([]CountryRow, 0, len(records))
	for _, record := range records {
		rows = append(rows, countryRow(record))
	}
	result.List = rows
	return result, nil
}

func (s *Service) CountryDetail(id uint64) (map[string]any, error) {
	record, err := s.countryByID(id)
	if err != nil {
		return nil, err
	}
	row := countryRow(record)
	return map[string]any{
		"overview": map[string]any{
			"id":                    row.ID,
			"countryCode":           row.CountryCode,
			"isoCountryCode":        row.ISOCountryCode,
			"countryNameZh":         row.CountryNameZh,
			"countryNameEn":         row.CountryNameEn,
			"currencyCode":          row.CurrencyCode,
			"currencyNameZh":        row.CurrencyNameZh,
			"currencyNameEn":        row.CurrencyNameEn,
			"currencySymbol":        row.CurrencySymbol,
			"currencyDecimalPlaces": row.CurrencyDecimalPlaces,
			"timezone":              row.Timezone,
			"defaultLanguage":       row.DefaultLanguage,
			"supportedLanguages":    row.SupportedLanguages,
			"phoneCountryCode":      row.PhoneCountryCode,
			"phoneLength":           row.PhoneLength,
			"enabled":               row.Enabled,
			"isDefault":             row.IsDefault,
			"createdAt":             row.CreatedAt,
			"updatedAt":             row.UpdatedAt,
		},
		"logoAssets": map[string]any{
			"logoUrl":              row.LogoURL,
			"faviconUrl":           row.FaviconURL,
			"loginLogoUrl":         row.LoginLogoURL,
			"headerLogoUrl":        row.HeaderLogoURL,
			"loadingLogoUrl":       row.LoadingLogoURL,
			"appIconUrl":           row.AppIconURL,
			"loadingBackgroundUrl": row.LoadingBackgroundURL,
		},
		"theme": map[string]any{
			"primaryColor":    row.PrimaryColor,
			"secondaryColor":  row.SecondaryColor,
			"buttonColor":     row.ButtonColor,
			"backgroundColor": row.BackgroundColor,
			"textColor":       row.TextColor,
		},
		"loadingAssets": map[string]any{
			"loadingLogoUrl":       row.LoadingLogoURL,
			"loadingAnimationUrl":  row.LoadingAnimationURL,
			"loadingBackgroundUrl": row.LoadingBackgroundURL,
		},
		"operationLogs": []map[string]any{},
	}, nil
}

func (s *Service) CreateCountry(input CountryInput, actor Actor) (uint64, error) {
	input = countryInputDefaults(input)
	if err := validateCountryInput(input, true); err != nil {
		return 0, err
	}
	var id uint64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		data := countryUpdates(input, true)
		data["status"] = 0
		data["created_at"], data["updated_at"] = now(), now()
		var old struct {
			ID        uint64
			DeletedAt *time.Time
		}
		err := tx.Table("game_country_configs").Select("id,deleted_at").Where("country_code=?", input.CountryCode).Take(&old).Error
		if err == nil && old.DeletedAt == nil {
			return NewError("DUPLICATE_COUNTRY_CODE", "country already exists")
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if old.ID > 0 {
			delete(data, "branding_config")
			delete(data, "theme_config")
			delete(data, "created_at")
			data["deleted_at"] = nil
			if err := tx.Table("game_country_configs").Where("id=?", old.ID).Updates(data).Error; err != nil {
				return err
			}
			id = old.ID
		} else {
			if err := tx.Table("game_country_configs").Create(data).Error; err != nil {
				return duplicateError(err)
			}
			if err := tx.Table("game_country_configs").Select("id").Where("country_code=? AND deleted_at IS NULL", input.CountryCode).Scan(&id).Error; err != nil {
				return err
			}
		}
		if boolDefault(input.Enabled, true) {
			if err := activateCountry(tx, id, actor); err != nil {
				return err
			}
			data["status"], data["is_default"] = 1, 1
		}
		return audit(tx, actor, "system:country:create", "system.country.create", "country_config", id, nil, data, input.Reason)
	})
	return id, err
}

func (s *Service) UpdateCountry(id uint64, input CountryInput, actor Actor) error {
	input = countryInputDefaults(input)
	if err := validateCountryInput(input, false); err != nil {
		return err
	}
	before, err := s.countryByID(id)
	if err != nil {
		return err
	}
	if input.Enabled == nil {
		input.Enabled = &before.Enabled
	}
	updates := countryUpdates(input, false)
	if boolDefault(input.Enabled, true) {
		delete(updates, "status")
	}
	if input.CountryCode != before.CountryCode {
		return NewError("COUNTRY_CODE_IMMUTABLE", "create a different country instead of renaming an existing code")
	}
	updates["updated_at"] = now()
	return s.db.Transaction(func(tx *gorm.DB) error {
		if before.Enabled && !boolDefault(input.Enabled, true) {
			if err := countryNotInUse(tx, before.CountryCode); err != nil {
				return err
			}
		}
		if err := tx.Table("game_country_configs").Where("id=? AND deleted_at IS NULL", id).Updates(updates).Error; err != nil {
			return duplicateError(err)
		}
		if boolDefault(input.Enabled, before.Enabled) {
			if err := activateCountry(tx, id, actor); err != nil {
				return err
			}
		}
		return audit(tx, actor, "system:country:update", "system.country.update", "country_config", id, before, updates, input.Reason)
	})
}

func (s *Service) CountryStatus(id uint64, enabled bool, reason string, actor Actor) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	before, err := s.countryByID(id)
	if err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if enabled {
			if err := activateCountry(tx, id, actor); err != nil {
				return err
			}
		} else {
			if err := countryNotInUse(tx, before.CountryCode); err != nil {
				return err
			}
			if err := tx.Table("game_country_configs").Where("id=?", id).Updates(map[string]any{"status": 0, "updated_at": now()}).Error; err != nil {
				return err
			}
		}
		return audit(tx, actor, "system:country:status", "system.country.status", "country_config", id, before, map[string]any{"enabled": enabled}, reason)
	})
}

func (s *Service) DeleteCountry(id uint64, reason string, actor Actor) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	before, err := s.countryByID(id)
	if err != nil {
		return err
	}
	if before.IsDefault {
		return NewError("DEFAULT_COUNTRY_DELETE_DENIED", "default country cannot be deleted")
	}
	updates := map[string]any{"deleted_at": now(), "updated_at": now()}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := countryNotInUse(tx, before.CountryCode); err != nil {
			return err
		}
		if err := tx.Table("game_country_configs").Where("id=? AND deleted_at IS NULL", id).Updates(updates).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:country:delete", "system.country.delete", "country_config", id, before, updates, reason)
	})
}

func (s *Service) SetDefaultCountry(id uint64, reason string, actor Actor) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	before, err := s.countryByID(id)
	if err != nil {
		return err
	}
	if !before.Enabled {
		return NewError("COUNTRY_DISABLED", "disabled country cannot be default")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_country_configs").Where("deleted_at IS NULL").Updates(map[string]any{"is_default": 0, "updated_at": now()}).Error; err != nil {
			return err
		}
		if err := tx.Table("game_country_configs").Where("id=? AND deleted_at IS NULL", id).Updates(map[string]any{"is_default": 1, "updated_at": now()}).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:country:set-default", "system.country.set_default", "country_config", id, before, map[string]any{"is_default": true}, reason)
	})
}

func (s *Service) RecordCountryAssetUpload(kind string, url string, actor Actor) error {
	return audit(s.db, actor, "system:country:branding", "system.country.asset_upload", "country_asset", 0, nil, map[string]any{"kind": kind, "url": url}, "")
}

func (s *Service) PublicAppConfig(countryCode string, language string) (map[string]any, error) {
	record, err := s.countryByCode(countryCode)
	if err != nil {
		record, err = s.defaultCountry()
	}
	if err != nil {
		return nil, err
	}
	row := countryRow(record)
	supported := splitList(row.SupportedLanguages)
	selectedLanguage := strings.TrimSpace(language)
	if selectedLanguage == "" || !containsString(supported, selectedLanguage) {
		selectedLanguage = row.DefaultLanguage
	}
	return map[string]any{
		"resourceVersion": countryResourceVersion(record),
		"country": map[string]any{
			"code":             row.CountryCode,
			"isoCountryCode":   row.ISOCountryCode,
			"name":             row.CountryNameEn,
			"nameZh":           row.CountryNameZh,
			"nameEn":           row.CountryNameEn,
			"timezone":         row.Timezone,
			"phoneCountryCode": row.PhoneCountryCode,
			"phoneLength":      row.PhoneLength,
		},
		"currency": map[string]any{
			"code":          row.CurrencyCode,
			"symbol":        row.CurrencySymbol,
			"decimalPlaces": row.CurrencyDecimalPlaces,
			"nameZh":        row.CurrencyNameZh,
			"nameEn":        row.CurrencyNameEn,
		},
		"language": map[string]any{
			"default":   row.DefaultLanguage,
			"current":   selectedLanguage,
			"supported": supported,
		},
		"branding": map[string]any{
			"logoUrl":              row.LogoURL,
			"faviconUrl":           row.FaviconURL,
			"loginLogoUrl":         row.LoginLogoURL,
			"headerLogoUrl":        row.HeaderLogoURL,
			"loadingLogoUrl":       row.LoadingLogoURL,
			"appIconUrl":           row.AppIconURL,
			"loadingAnimationUrl":  row.LoadingAnimationURL,
			"loadingBackgroundUrl": row.LoadingBackgroundURL,
			"theme": map[string]any{
				"primaryColor":    row.PrimaryColor,
				"secondaryColor":  row.SecondaryColor,
				"buttonColor":     row.ButtonColor,
				"backgroundColor": row.BackgroundColor,
				"textColor":       row.TextColor,
			},
		},
	}, nil
}

func countryResourceVersion(record countryConfigRecord) string {
	value := strings.Join([]string{
		record.CountryCode,
		record.CurrencyCode,
		record.CurrencySymbol,
		record.Timezone,
		record.DefaultLanguage,
		record.SupportedLanguages,
		record.BrandingConfig,
		record.ThemeConfig,
		record.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}, "\x00")
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:12])
}

func (s *Service) countryByID(id uint64) (countryConfigRecord, error) {
	var row countryConfigRecord
	if err := s.db.Table("game_country_configs").Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return row, rowNotFound("country")
	}
	return row, nil
}

func (s *Service) countryByCode(code string) (countryConfigRecord, error) {
	var row countryConfigRecord
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return row, rowNotFound("country")
	}
	if err := s.db.Table("game_country_configs").Where("country_code=? AND status=1 AND deleted_at IS NULL", code).Take(&row).Error; err != nil {
		return row, rowNotFound("country")
	}
	return row, nil
}

func (s *Service) defaultCountry() (countryConfigRecord, error) {
	var row countryConfigRecord
	if err := s.db.Table("game_country_configs").Where("status=1 AND deleted_at IS NULL").Order("is_default DESC,country_code").Take(&row).Error; err != nil {
		return row, rowNotFound("country")
	}
	return row, nil
}

func countryRow(record countryConfigRecord) CountryRow {
	branding := parseBranding(record.BrandingConfig)
	theme := parseTheme(record.ThemeConfig)
	supported := strings.Join(splitList(record.SupportedLanguages), ",")
	return CountryRow{
		ID: record.ID, CountryCode: record.CountryCode, ISOCountryCode: record.ISOCountryCode, CountryNameZh: record.CountryNameZh, CountryNameEn: record.CountryNameEn,
		CountryName: record.CountryNameEn, CurrencyCode: record.CurrencyCode, CurrencyNameZh: record.CurrencyNameZh, CurrencyNameEn: record.CurrencyNameEn,
		CurrencySymbol: record.CurrencySymbol, CurrencyDecimalPlaces: record.CurrencyDecimalPlaces, Timezone: record.Timezone,
		DefaultLanguage: record.DefaultLanguage, SupportedLanguages: supported, PhoneCountryCode: record.PhoneCountryCode, PhoneLength: record.PhoneLength,
		Enabled: record.Enabled, Status: record.Enabled, IsDefault: record.IsDefault,
		LogoURL: countryAssetURL(branding.LogoURL), FaviconURL: countryAssetURL(branding.FaviconURL), LoginLogoURL: countryAssetURL(branding.LoginLogoURL), HeaderLogoURL: countryAssetURL(branding.HeaderLogoURL),
		LoadingLogoURL: countryAssetURL(branding.LoadingLogoURL), AppIconURL: countryAssetURL(branding.AppIconURL), LoadingAnimationURL: countryAssetURL(branding.LoadingAnimationURL),
		LoadingBackgroundURL: countryAssetURL(branding.LoadingBackgroundURL), PrimaryColor: theme.PrimaryColor, SecondaryColor: theme.SecondaryColor,
		ButtonColor: theme.ButtonColor, BackgroundColor: theme.BackgroundColor, TextColor: theme.TextColor,
		CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
	}
}

func countryUpdates(input CountryInput, creating bool) map[string]any {
	countryCode := strings.ToUpper(strings.TrimSpace(input.CountryCode))
	isoCountryCode := strings.ToUpper(strings.TrimSpace(input.ISOCountryCode))
	if isoCountryCode == "" {
		isoCountryCode = countryCode
	}
	phoneCountryCode := strings.TrimSpace(input.PhoneCountryCode)
	if phoneCountryCode == "" {
		phoneCountryCode = "+92"
	}
	phoneLength := input.PhoneLength
	branding := countryBranding{
		LogoURL: countryAssetRef(input.LogoURL), FaviconURL: countryAssetRef(input.FaviconURL), LoginLogoURL: countryAssetRef(input.LoginLogoURL),
		HeaderLogoURL: countryAssetRef(input.HeaderLogoURL), LoadingLogoURL: countryAssetRef(input.LoadingLogoURL), AppIconURL: countryAssetRef(input.AppIconURL),
		LoadingAnimationURL: countryAssetRef(input.LoadingAnimationURL), LoadingBackgroundURL: countryAssetRef(input.LoadingBackgroundURL),
	}
	theme := countryTheme{
		PrimaryColor: strings.TrimSpace(input.PrimaryColor), SecondaryColor: strings.TrimSpace(input.SecondaryColor), ButtonColor: strings.TrimSpace(input.ButtonColor),
		BackgroundColor: strings.TrimSpace(input.BackgroundColor), TextColor: strings.TrimSpace(input.TextColor),
	}
	updates := map[string]any{
		"country_code": countryCode, "iso_country_code": isoCountryCode,
		"country_name_zh": input.CountryNameZh, "country_name_en": input.CountryNameEn, "currency_code": strings.ToUpper(input.CurrencyCode),
		"currency_name_zh": input.CurrencyNameZh, "currency_name_en": input.CurrencyNameEn, "currency_symbol": input.CurrencySymbol,
		"currency_decimal_places": input.CurrencyDecimalPlaces, "timezone": input.Timezone, "default_language": input.DefaultLanguage,
		"supported_languages": marshal(splitList(input.SupportedLanguages)), "phone_country_code": phoneCountryCode, "phone_length": phoneLength,
		"branding_config": marshal(branding), "theme_config": marshal(theme),
		"status": boolValue(boolDefault(input.Enabled, true)),
	}
	if !creating {
		delete(updates, "branding_config")
		delete(updates, "theme_config")
	}
	return updates
}

func validateCountryInput(input CountryInput, creating bool) error {
	countryCode := strings.ToUpper(strings.TrimSpace(input.CountryCode))
	if !regexp.MustCompile(`^[A-Z]{2}$`).MatchString(countryCode) {
		return NewError("INVALID_COUNTRY_CODE", "country code must be ISO alpha-2")
	}
	isoCountryCode := strings.ToUpper(strings.TrimSpace(input.ISOCountryCode))
	if isoCountryCode != "" && !regexp.MustCompile(`^[A-Z]{2}$`).MatchString(isoCountryCode) {
		return NewError("INVALID_ISO_COUNTRY_CODE", "ISO country code must be alpha-2")
	}
	if isoCountryCode != "" && isoCountryCode != countryCode {
		return NewError("INVALID_ISO_COUNTRY_CODE", "country and ISO country code must be the same")
	}
	if !regexp.MustCompile(`^[A-Z]{3}$`).MatchString(strings.ToUpper(strings.TrimSpace(input.CurrencyCode))) {
		return NewError("INVALID_CURRENCY", "currency code must have three letters")
	}
	if strings.TrimSpace(input.CountryNameEn) == "" || strings.TrimSpace(input.CurrencyCode) == "" || strings.TrimSpace(input.CurrencySymbol) == "" {
		return NewError("INVALID_REQUEST", "country name, currency code and currency symbol are required")
	}
	if !regexp.MustCompile(`^\+\d{1,4}$`).MatchString(strings.TrimSpace(input.PhoneCountryCode)) {
		return NewError("INVALID_PHONE_COUNTRY_CODE", "phone country code must start with +")
	}
	if input.PhoneLength < 0 || input.PhoneLength > 15 {
		return NewError("INVALID_PHONE_LENGTH", "phone length must be between 1 and 15, or 0 for variable length")
	}
	if input.CurrencyDecimalPlaces < 0 || input.CurrencyDecimalPlaces > 8 {
		return NewError("INVALID_CURRENCY_DECIMALS", "invalid currency decimal places")
	}
	if strings.TrimSpace(input.Timezone) == "" || strings.TrimSpace(input.DefaultLanguage) == "" {
		return NewError("INVALID_REQUEST", "timezone and default language are required")
	}
	for _, asset := range []string{input.LogoURL, input.FaviconURL, input.LoginLogoURL, input.HeaderLogoURL, input.LoadingLogoURL, input.AppIconURL, input.LoadingAnimationURL, input.LoadingBackgroundURL} {
		if err := validateRuntimeAssetURL(asset); err != nil {
			return err
		}
	}
	for _, color := range []string{input.PrimaryColor, input.SecondaryColor, input.ButtonColor, input.BackgroundColor, input.TextColor} {
		if err := validateColor(color); err != nil {
			return err
		}
	}
	return nil
}

func validateRuntimeAssetURL(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if strings.HasPrefix(value, "data:") {
		return NewError("INVALID_ASSET_URL", "base64 assets are not allowed")
	}
	if storage.IsAssetURL(value) {
		return nil
	}
	return NewError("INVALID_ASSET_URL", "asset URL must use static storage or http(s)")
}

func countryAssetURL(value string) string {
	out, err := storage.PublicURL(context.Background(), value)
	if err != nil {
		return value
	}
	return out
}

func countryAssetRef(value string) string {
	return storage.NormalizeReference(value)
}

func validateColor(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if !regexp.MustCompile(`^#[0-9a-fA-F]{3}([0-9a-fA-F]{3})?$`).MatchString(value) {
		return NewError("INVALID_COLOR", "theme colors must be hex values")
	}
	return nil
}

func parseBranding(raw string) countryBranding {
	var out countryBranding
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func parseTheme(raw string) countryTheme {
	var out countryTheme
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func splitList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{"en"}
	}
	var values []string
	if strings.HasPrefix(raw, "[") && json.Unmarshal([]byte(raw), &values) == nil {
		return cleanStringList(values)
	}
	return cleanStringList(strings.Split(raw, ","))
}

func cleanStringList(values []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	if len(out) == 0 {
		out = append(out, "en")
	}
	return out
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
