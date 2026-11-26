// Package connections stores independent supplier accounts, never tenants.
package connections

import (
	"context"
	"errors"
	"net"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"casino/internal/config"
	"casino/internal/game/domain"
	"casino/internal/game/market"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Config struct {
	MarketCurrency      string   `json:"-" gorm:"-"`
	ProviderCode        string   `json:"providerCode" gorm:"primaryKey"`
	AdapterType         string   `json:"adapterType"`
	Name                string   `json:"name"`
	APIBaseURL          string   `json:"apiBaseUrl"`
	Environment         string   `json:"environment"`
	Enabled             bool     `json:"enabled"`
	SupportedCountries  string   `json:"supportedCountries"`
	SupportedCurrencies string   `json:"supportedCurrencies"`
	CallbackIPAllowlist []string `json:"callbackIPAllowlist" gorm:"serializer:json"`
	Version             uint64   `json:"version"`
}
type Record struct {
	Config
	CredentialsCiphertext string `json:"-"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func (Record) TableName() string { return "game_provider_connections" }

type View struct {
	Config
	Source                string          `json:"source"`
	CredentialsConfigured bool            `json:"credentialsConfigured"`
	CredentialFields      map[string]bool `json:"credentialFields"`
	MarketAllowed         bool            `json:"marketAllowed"`
	CallbackPath          string          `json:"callbackPath"`
}
type Input struct {
	Config
	Credentials      map[string]string `json:"credentials"`
	ClearCredentials bool              `json:"clearCredentials"`
	Reason           string            `json:"reason"`
}
type Store struct{ db *gorm.DB }

func New(db *gorm.DB) *Store { return &Store{db: db} }

var codePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{1,63}$`)
var ErrConflict = errors.New("provider configuration changed; reload before saving")

func legacy(code string) (Config, map[string]string, bool) {
	c := Config{ProviderCode: code, AdapterType: code, Environment: "sandbox", CallbackIPAllowlist: []string{}}
	secrets := map[string]string{}
	switch code {
	case "ag":
		s := config.Get().Game.AG
		c.Name = "AG"
		c.APIBaseURL = s.APIBaseURL
		c.Enabled = s.Enabled
		c.CallbackIPAllowlist = s.CallbackIPAllowlist
		secrets = map[string]string{"merchantCode": s.MerchantCode, "secretKey": s.SecretKey, "subMerchantCode": s.SubMerchantCode}
		if c.APIBaseURL == "" {
			c.APIBaseURL = "https://pre-api-area.agsoftgames.com"
		}
	case "softaggregator":
		s := config.Get().Game.SoftAggregator
		c.Name = "SoftAggregator"
		c.APIBaseURL = s.APIBaseURL
		c.Enabled = s.Enabled
		c.CallbackIPAllowlist = s.CallbackIPAllowlist
		secrets = map[string]string{"apiLogin": s.APILogin, "apiPassword": s.APIPassword, "saltKey": s.SaltKey}
		if c.APIBaseURL == "" {
			c.APIBaseURL = "https://api.softaggregator.com"
		}
	default:
		return c, secrets, false
	}
	return c, secrets, true
}
func (s *Store) load(code string) (Config, map[string]string, bool, error) {
	code = strings.ToLower(strings.TrimSpace(code))
	var row Record
	err := s.db.Where("provider_code=?", code).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c, values, ok := legacy(code)
		if ok {
			var metadata struct {
				Name                string
				Environment         string
				SupportedCountries  string
				SupportedCurrencies string
			}
			if err := s.db.Table("game_game_providers").Select("name,environment,supported_countries,supported_currencies").Where("provider_code=? AND deleted_at IS NULL", code).Order("site_id ASC,id ASC").Take(&metadata).Error; err == nil {
				c.SupportedCountries, c.SupportedCurrencies = metadata.SupportedCountries, metadata.SupportedCurrencies
				if metadata.Name != "" {
					c.Name = metadata.Name
				}
				if metadata.Environment != "" {
					c.Environment = metadata.Environment
				}
			}
		}
		return c, values, ok, nil
	}
	if err != nil {
		return Config{}, nil, true, errors.New("provider configuration could not be loaded")
	}
	values, err := open(code, row.AdapterType, row.CredentialsCiphertext)
	return row.Config, values, true, err
}
func (s *Store) Provider(code string) (domain.GameProvider, bool, error) {
	c, values, found, err := s.load(code)
	if err != nil || !found {
		return nil, found, err
	}
	if !c.Enabled {
		return nil, true, domain.ErrProviderDisabled
	}
	if deployment, e := market.Load(s.db); e == nil {
		c.MarketCurrency = deployment.Currency
	}
	p, _, err := build(c, values)
	return p, true, err
}
func (s *Store) Callback(code string) (domain.ProviderCallbackAdapter, bool, error) {
	c, values, found, err := s.load(code)
	if err != nil || !found {
		return nil, found, err
	}
	_, cb, err := build(c, values)
	return cb, true, err
}
func (s *Store) configs() ([]Config, error) {
	var records []Record
	if err := s.db.Order("provider_code").Find(&records).Error; err != nil {
		return nil, err
	}
	out := []Config{}
	seen := map[string]bool{}
	for _, r := range records {
		out = append(out, r.Config)
		seen[r.ProviderCode] = true
	}
	for _, code := range []string{"ag", "softaggregator"} {
		if !seen[code] {
			c, _, _ := legacy(code)
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ProviderCode < out[j].ProviderCode })
	return out, nil
}
func (s *Store) view(c Config, values map[string]string, m market.Market) View {
	fields := map[string]bool{}
	for _, key := range factories[c.AdapterType].Fields {
		fields[key] = values[key] != ""
	}
	source := "database"
	if c.Version == 0 {
		source = "environment"
	}
	return View{Config: c, Source: source, CredentialsConfigured: ready(c, values), CredentialFields: fields, MarketAllowed: m.ProviderAllowed(c.SupportedCountries, c.SupportedCurrencies), CallbackPath: "/api/v1/game/providers/" + c.ProviderCode + "/callback"}
}
func (s *Store) List() (map[string]any, error) {
	configs, err := s.configs()
	if err != nil {
		return nil, err
	}
	m, err := market.Load(s.db)
	if err != nil {
		return nil, err
	}
	rows := []View{}
	for _, c := range configs {
		effective, values, _, err := s.load(c.ProviderCode)
		if err != nil {
			return nil, err
		}
		rows = append(rows, s.view(effective, values, m))
	}
	adapters := []map[string]any{}
	for _, key := range []string{"ag", "softaggregator"} {
		f := factories[key]
		adapters = append(adapters, map[string]any{"type": key, "credentialFields": f.Fields, "requiredFields": f.Required})
	}
	var site struct{ H5Domain string }
	_ = s.db.Table("game_site_config").Select("h5_domain").Where("id=1").Take(&site).Error
	var countries []struct {
		CountryCode  string `json:"countryCode"`
		CountryName  string `json:"countryName"`
		CurrencyCode string `json:"currencyCode"`
	}
	if err := s.db.Table("game_country_configs").Select("DISTINCT country_code,country_name_en AS country_name,currency_code").Where("status=1 AND deleted_at IS NULL").Order("country_code").Scan(&countries).Error; err != nil {
		return nil, err
	}
	return map[string]any{"list": rows, "market": map[string]string{"countryCode": m.Country, "currencyCode": m.Currency}, "adapters": adapters, "callbackBaseUrl": site.H5Domain, "countries": countries}, nil
}

// PublicCodes is recalculated from the published market and current configuration.
// A country/currency query string can never select a different deployment market.
func (s *Store) PublicCodes(ctx context.Context) []string {
	db := s.db.WithContext(ctx)
	m, err := market.Load(db)
	if err != nil {
		return []string{}
	}
	configs, err := s.configs()
	if err != nil {
		return []string{}
	}
	out := []string{}
	for _, c := range configs {
		if !c.Enabled {
			continue
		}
		effective, values, _, err := s.load(c.ProviderCode)
		if err != nil || !ready(effective, values) {
			continue
		}
		var provider struct {
			Status              int
			SupportedCountries  string
			SupportedCurrencies string
		}
		if err := db.Table("game_game_providers").Select("status,supported_countries,supported_currencies").Where("provider_code=? AND site_id IN (0,1) AND deleted_at IS NULL", c.ProviderCode).Order("site_id ASC,id ASC").Take(&provider).Error; err != nil || provider.Status != 1 {
			continue
		}
		countries, currencies := effective.SupportedCountries, effective.SupportedCurrencies
		if effective.Version == 0 {
			countries, currencies = provider.SupportedCountries, provider.SupportedCurrencies
		}
		if m.ProviderAllowed(countries, currencies) && m.ProviderAllowed(provider.SupportedCountries, provider.SupportedCurrencies) {
			out = append(out, c.ProviderCode)
		}
	}
	return out
}
func (s *Store) Providers() []domain.GameProvider {
	configs, err := s.configs()
	if err != nil {
		return nil
	}
	out := []domain.GameProvider{}
	for _, c := range configs {
		p, _, err := s.Provider(c.ProviderCode)
		if err == nil {
			out = append(out, p)
		}
	}
	return out
}
func (s *Store) Callbacks() []domain.ProviderCallbackAdapter {
	configs, err := s.configs()
	if err != nil {
		return nil
	}
	out := []domain.ProviderCallbackAdapter{}
	for _, c := range configs {
		cb, _, err := s.Callback(c.ProviderCode)
		if err == nil {
			out = append(out, cb)
		}
	}
	return out
}

func normalize(input *Input) error {
	input.ProviderCode = strings.ToLower(strings.TrimSpace(input.ProviderCode))
	input.AdapterType = strings.ToLower(strings.TrimSpace(input.AdapterType))
	input.Name = strings.TrimSpace(input.Name)
	input.APIBaseURL = strings.TrimRight(strings.TrimSpace(input.APIBaseURL), "/")
	if !codePattern.MatchString(input.ProviderCode) || len(input.Name) == 0 || len(input.Name) > 128 {
		return errors.New("supplier code and name are required; codes use lowercase letters, digits, hyphens or underscores")
	}
	f, ok := factories[input.AdapterType]
	if !ok {
		return errors.New("unsupported provider adapter")
	}
	if input.Environment != "sandbox" && input.Environment != "production" {
		return errors.New("environment must be sandbox or production")
	}
	u, err := url.Parse(input.APIBaseURL)
	if err != nil || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Scheme != "https" && u.Scheme != "http") || (input.Environment == "production" && u.Scheme != "https") {
		return errors.New("provide a valid API base URL; production connections require HTTPS")
	}
	if len(input.APIBaseURL) > 500 {
		return errors.New("API base URL is too long")
	}
	for _, r := range []struct {
		value *string
		size  int
	}{{&input.SupportedCountries, 2}, {&input.SupportedCurrencies, 3}} {
		vs := market.Values(*r.value)
		if len(vs) == 0 {
			return errors.New("explicit supported countries and currencies are required")
		}
		for _, v := range vs {
			if !market.Letters(v, r.size) {
				return errors.New("invalid country or currency code")
			}
		}
		sort.Strings(vs)
		*r.value = strings.Join(vs, ",")
		if len(*r.value) > 500 {
			return errors.New("country or currency list is too long")
		}
	}
	if input.CallbackIPAllowlist == nil {
		input.CallbackIPAllowlist = []string{}
	}
	if len(input.CallbackIPAllowlist) > 100 {
		return errors.New("too many callback IP rules")
	}
	for i, v := range input.CallbackIPAllowlist {
		v = strings.TrimSpace(v)
		input.CallbackIPAllowlist[i] = v
		if net.ParseIP(v) == nil {
			if _, _, err := net.ParseCIDR(v); err != nil {
				return errors.New("invalid callback IP or CIDR")
			}
		}
	}
	allowed := map[string]bool{}
	for _, key := range f.Fields {
		allowed[key] = true
	}
	for key, value := range input.Credentials {
		if !allowed[key] || len(value) > 4096 {
			return errors.New("invalid provider credential field")
		}
	}
	if strings.TrimSpace(input.Reason) == "" || len(input.Reason) > 500 {
		return errors.New("a change reason is required")
	}
	return nil
}
func (s *Store) Save(input Input) (View, error) {
	if err := normalize(&input); err != nil {
		return View{}, err
	}
	var saved Config
	var savedValues map[string]string
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var before Record
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("provider_code=?", input.ProviderCode).Take(&before).Error
		existing := err == nil
		if !existing {
			if previous, _, reserved := legacy(input.ProviderCode); reserved && previous.AdapterType != input.AdapterType {
				return errors.New("built-in connection codes keep their original API protocol; create a new connection code")
			}
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if input.Version != before.Version {
			return ErrConflict
		}
		if existing && before.AdapterType != input.AdapterType {
			return errors.New("an existing connection cannot change API protocol; create a new connection code")
		}
		values := map[string]string{}
		if existing {
			values, err = open(before.ProviderCode, before.AdapterType, before.CredentialsCiphertext)
			if err != nil {
				return err
			}
		} else if c, v, ok := legacy(input.ProviderCode); ok && c.AdapterType == input.AdapterType {
			values = v
		}
		if input.ClearCredentials {
			values = map[string]string{}
		}
		for key, value := range input.Credentials {
			if value != "" && value != "***" {
				values[key] = value
			}
		}
		if input.Enabled && !ready(input.Config, values) {
			return errors.New("complete all required credentials before enabling the supplier")
		}
		ciphertext, err := seal(input.ProviderCode, input.AdapterType, values)
		if err != nil {
			return err
		}
		input.Version = before.Version + 1
		record := Record{Config: input.Config, CredentialsCiphertext: ciphertext, CreatedAt: time.Now(), UpdatedAt: time.Now()}
		if existing {
			record.CreatedAt = before.CreatedAt
		}
		if err := tx.Save(&record).Error; err != nil {
			return err
		}
		var count int64
		if err := tx.Table("game_game_providers").Where("provider_code=? AND site_id=0 AND deleted_at IS NULL", input.ProviderCode).Count(&count).Error; err != nil {
			return err
		}
		metadata := map[string]any{"name": input.Name, "api_base_url": input.APIBaseURL, "environment": input.Environment, "supported_countries": input.SupportedCountries, "supported_currencies": input.SupportedCurrencies, "callback_url": "/api/v1/game/providers/" + input.ProviderCode + "/callback", "updated_at": time.Now()}
		if count == 0 {
			row := map[string]any{}
			for k, v := range metadata {
				row[k] = v
			}
			row["provider_code"], row["site_id"], row["status"], row["wallet_mode"], row["created_at"] = input.ProviderCode, 0, 3, "SEAMLESS", time.Now()
			if input.Enabled {
				row["status"] = 1
			}
			if err := tx.Table("game_game_providers").Create(row).Error; err != nil {
				return err
			}
		}
		if input.Enabled {
			metadata["status"] = gorm.Expr("CASE WHEN status=2 THEN 2 ELSE 1 END")
		} else {
			metadata["status"] = 3
		}
		if err := tx.Table("game_game_providers").Where("provider_code=? AND deleted_at IS NULL", input.ProviderCode).Updates(metadata).Error; err != nil {
			return err
		}
		saved, savedValues = input.Config, values
		return nil
	})
	if err != nil {
		return View{}, err
	}
	m, err := market.Load(s.db)
	if err != nil {
		return View{}, err
	}
	return s.view(saved, savedValues, m), nil
}
func (s *Store) SanitizeError(code string, err error) string {
	if err == nil {
		return ""
	}
	_, values, _, loadErr := s.load(code)
	if loadErr != nil {
		return "provider connection failed"
	}
	message := err.Error()
	for _, value := range values {
		if value != "" {
			message = strings.ReplaceAll(message, value, "[redacted]")
		}
	}
	if len(message) > 500 {
		message = message[:500]
	}
	return message
}
func AdapterTypes() []string { return []string{"ag", "softaggregator"} }
func (s *Store) Description(code string) (string, string) {
	c, _, _, err := s.load(code)
	if err != nil {
		return code, ""
	}
	return c.Name, c.APIBaseURL
}
