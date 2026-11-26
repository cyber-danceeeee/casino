package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"

	"casino/internal/game/domain"
	"casino/internal/game/repository"
	"casino/internal/hotranking"
	"casino/internal/model"
	"casino/internal/tenancy"
)

const (
	sessionHeartbeatTimeout  = 2 * time.Minute
	sessionTimeoutSweepLimit = 500

	sessionStatusLaunchRequested = "LAUNCH_REQUESTED"
	sessionStatusLaunchSuccess   = "LAUNCH_SUCCESS"
	sessionStatusLaunchFailed    = "LAUNCH_FAILED"
	sessionStatusPlaying         = "PLAYING"
	sessionStatusClosed          = "CLOSED"
)

type LaunchService struct {
	registry  domain.ProviderRegistry
	repo      *repository.Repository
	launches  singleflight.Group
	hotRanker *hotranking.Store
}

type launchCountryConfig struct {
	ID                    uint64
	CountryCode           string
	CurrencyCode          string
	CurrencySymbol        string
	CurrencyDecimalPlaces int
	DefaultLanguage       string
	SupportedLanguages    string
	Timezone              string
	LanguageCode          string
}

type LaunchRequest struct {
	TenantID     uint64
	SiteID       uint64
	ProviderCode string
	GameID       string
	UserID       uint64
	Currency     string
	Language     string
	Country      string
	Device       string
	ReturnURL    string
	SessionID    string
	IPAddress    string
	Demo         bool
}

type providerTraceError interface {
	error
	ProviderRequestTrace() string
	ProviderResponseTrace() string
}

func NewLaunchService(registry domain.ProviderRegistry, repo *repository.Repository) *LaunchService {
	return &LaunchService{registry: registry, repo: repo}
}

func (s *LaunchService) SetHotRanker(rank *hotranking.Store) {
	s.hotRanker = rank
}

func (s *LaunchService) Launch(ctx context.Context, req LaunchRequest) (*domain.LaunchGameResult, error) {
	req = normalizeLaunchRequest(req)
	if err := validateBasicLaunchRequest(req); err != nil {
		return nil, err
	}
	_ = s.closeTimedOutSessions(ctx, time.Now())

	scope := launchScope(req)
	providerRow, game, err := s.repo.FindLaunchGame(ctx, scope, req.ProviderCode, req.GameID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.NewWalletError(domain.WalletErrorGameNotFound, "game not found", err)
		}
		return nil, err
	}
	launchTime := time.Now()
	requestJSON := launchRequestJSON(req)
	log := newLaunchLog(scope, req, game.ID, providerRow.ID, requestJSON, launchTime)
	if err := tenancy.ApplySiteStrict(s.repo.DB().WithContext(ctx).Model(&model.GameGameLaunchLogs{}), "", scope).Create(log).Error; err != nil {
		return nil, err
	}

	provider, err := s.registry.Get(req.ProviderCode)
	if err != nil {
		_ = s.markLaunchFailed(ctx, log, "provider is not enabled", "", "")
		return nil, domain.NewWalletError(domain.WalletErrorProviderDisabled, "provider is not enabled", err)
	}
	if providerRow.Status != 1 {
		_ = s.markLaunchFailed(ctx, log, "provider is disabled", "", "")
		return nil, domain.NewWalletError(domain.WalletErrorProviderDisabled, "provider is disabled", nil)
	}
	if game.Status != 1 {
		_ = s.markLaunchFailed(ctx, log, "game is not enabled", "", "")
		return nil, domain.NewWalletError(domain.WalletErrorGameNotFound, "game is not enabled", nil)
	}
	countryConfig, err := s.resolveLaunchCountry(ctx, req)
	if err != nil {
		_ = s.markLaunchFailed(ctx, log, err.Error(), "", "")
		return nil, err
	}
	req.Country = countryConfig.CountryCode
	req.Currency = countryConfig.CurrencyCode
	req.Language = countryConfig.LanguageCode
	log.Currency = req.Currency
	log.Language = req.Language
	log.Country = req.Country
	log.CountryID = countryConfig.ID
	log.CountryCodeSnapshot = countryConfig.CountryCode
	log.CurrencyCodeSnapshot = countryConfig.CurrencyCode
	log.LanguageCodeSnapshot = countryConfig.LanguageCode
	if err := s.updateLaunchConfiguration(ctx, log); err != nil {
		_ = s.markLaunchFailed(ctx, log, err.Error(), "", "")
		return nil, err
	}
	if err := validateLaunchRequest(req); err != nil {
		_ = s.markLaunchFailed(ctx, log, err.Error(), "", "")
		return nil, err
	}
	if code, message, blocked := s.launchMaintenance(ctx, scope, providerRow.ProviderCode, game.ID, req.Country, req.Currency); blocked {
		_ = s.markLaunchFailed(ctx, log, message, "", "")
		return nil, domain.NewWalletError(code, message, nil)
	}
	if err := s.validateUserWalletCurrency(ctx, scope, req.UserID, req.Currency); err != nil {
		_ = s.markLaunchFailed(ctx, log, err.Error(), "", "")
		return nil, err
	}

	if req.Demo {
		if !game.SupportsDemo {
			_ = s.markLaunchFailed(ctx, log, "demo mode is not supported", "", "")
			return nil, domain.NewWalletError(domain.WalletErrorInvalidRequest, "demo mode is not supported", nil)
		}
		result, err := provider.LaunchDemo(ctx, domain.LaunchDemoRequest{
			ProviderCode:   req.ProviderCode,
			ProviderGameID: game.ProviderGameCode,
			Currency:       req.Currency,
			Language:       req.Language,
			Country:        req.Country,
			Device:         req.Device,
			ReturnURL:      req.ReturnURL,
			SessionID:      req.SessionID,
			IPAddress:      req.IPAddress,
		})
		if err != nil {
			requestTrace, responseTrace := launchErrorTrace(err)
			_ = s.markLaunchFailed(ctx, log, err.Error(), requestTrace, responseTrace)
			return nil, err
		}
		result.SessionID = req.SessionID
		_ = s.markLaunchEntered(ctx, log, result)
		return result, nil
	}

	player, err := s.ensureProviderPlayer(ctx, provider, req)
	if err != nil {
		requestTrace, responseTrace := launchErrorTrace(err)
		_ = s.markLaunchFailed(ctx, log, err.Error(), requestTrace, responseTrace)
		return nil, err
	}

	launchRequestKey := fmt.Sprintf("%d:%s", req.UserID, req.SessionID)
	value, err, _ := s.launches.Do(launchRequestKey, func() (any, error) {
		return provider.LaunchGame(ctx, domain.LaunchGameRequest{
			ProviderCode:           req.ProviderCode,
			ProviderGameID:         game.ProviderGameCode,
			UserID:                 fmt.Sprintf("%d", req.UserID),
			Currency:               req.Currency,
			Language:               req.Language,
			Country:                req.Country,
			Device:                 req.Device,
			ReturnURL:              req.ReturnURL,
			SessionID:              req.SessionID,
			IPAddress:              req.IPAddress,
			ProviderPlayerID:       player.ProviderPlayerID,
			ProviderPlayerPassword: player.ProviderPlayerPassword,
		})
	})
	if err != nil {
		requestTrace, responseTrace := launchErrorTrace(err)
		_ = s.markLaunchFailed(ctx, log, err.Error(), requestTrace, responseTrace)
		return nil, err
	}
	result := value.(*domain.LaunchGameResult)
	result.SessionID = req.SessionID
	_ = s.markLaunchEntered(ctx, log, result)
	return result, nil
}

func (s *LaunchService) validateUserWalletCurrency(ctx context.Context, scope tenancy.Scope, userID uint64, currency string) error {
	var wallet model.GameUserWallets
	if err := tenancy.ApplySite(s.repo.DB().WithContext(ctx).Model(&model.GameUserWallets{}), "", scope).Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		return domain.NewWalletError(domain.WalletErrorPlayerNotFound, "wallet not found", err)
	}
	if ok, err := reconcileWalletCurrency(ctx, s.repo.DB(), scope, &wallet, currency); err != nil {
		return err
	} else if !ok {
		return domain.NewWalletError(domain.WalletErrorCurrencyMismatch, "wallet currency mismatch", nil)
	}
	return nil
}

func (s *LaunchService) resolveLaunchCountry(ctx context.Context, req LaunchRequest) (launchCountryConfig, error) {
	var user model.GameUsers
	scope := launchScope(req)
	if err := tenancy.ApplySite(s.repo.DB().WithContext(ctx).Model(&model.GameUsers{}), "", scope).Select("id", "country_code", "language").Where("id = ? AND deleted_at IS NULL", req.UserID).Take(&user).Error; err != nil {
		return launchCountryConfig{}, domain.NewWalletError(domain.WalletErrorPlayerNotFound, "player not found", err)
	}

	countryCode := strings.ToUpper(strings.TrimSpace(req.Country))
	if countryCode == "" {
		countryCode = strings.ToUpper(strings.TrimSpace(user.CountryCode))
	}

	var row struct {
		ID                    uint64 `gorm:"column:id"`
		CountryCode           string `gorm:"column:country_code"`
		CurrencyCode          string `gorm:"column:currency_code"`
		CurrencySymbol        string `gorm:"column:currency_symbol"`
		CurrencyDecimalPlaces int    `gorm:"column:currency_decimal_places"`
		DefaultLanguage       string `gorm:"column:default_language"`
		SupportedLanguages    string `gorm:"column:supported_languages"`
		Timezone              string `gorm:"column:timezone"`
	}
	countryDB := func(currentScope tenancy.Scope) *gorm.DB {
		db := s.repo.DB().WithContext(ctx).Table("game_country_configs").Where("status=1 AND deleted_at IS NULL")
		if currentScope.Platform {
			return db.Where("tenant_id=0 AND site_id=0")
		}
		return tenancy.ApplySite(db, "", currentScope)
	}
	findCountry := func(currentScope tenancy.Scope) error {
		db := countryDB(currentScope)
		query := db
		if countryCode != "" {
			query = query.Where("country_code=?", countryCode)
		} else {
			query = query.Order("is_default DESC,country_code")
		}
		if err := query.Take(&row).Error; err != nil {
			if countryCode == "" || !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			return db.Order("is_default DESC,country_code").Take(&row).Error
		}
		return nil
	}
	if err := findCountry(scope); err != nil {
		if scope.TenantScoped() && errors.Is(err, gorm.ErrRecordNotFound) {
			if platformErr := findCountry(tenancy.Scope{Platform: true}); platformErr == nil {
				err = nil
			} else {
				err = platformErr
			}
		}
		if err != nil {
			return launchCountryConfig{}, domain.NewWalletError(domain.WalletErrorInvalidRequest, "country configuration not found", err)
		}
	}

	language := strings.ToLower(strings.TrimSpace(req.Language))
	if language == "" {
		language = strings.ToLower(strings.TrimSpace(user.Language))
	}
	supported := supportedLanguages(row.SupportedLanguages)
	defaultLanguage := strings.ToLower(strings.TrimSpace(row.DefaultLanguage))
	if defaultLanguage == "" && len(supported) > 0 {
		defaultLanguage = supported[0]
	}
	if language == "" || !languageSupported(language, supported) {
		language = defaultLanguage
	}
	if language == "" {
		language = "en"
	}

	return launchCountryConfig{
		ID:                    row.ID,
		CountryCode:           strings.ToUpper(strings.TrimSpace(row.CountryCode)),
		CurrencyCode:          strings.ToUpper(strings.TrimSpace(row.CurrencyCode)),
		CurrencySymbol:        row.CurrencySymbol,
		CurrencyDecimalPlaces: row.CurrencyDecimalPlaces,
		DefaultLanguage:       defaultLanguage,
		SupportedLanguages:    row.SupportedLanguages,
		Timezone:              row.Timezone,
		LanguageCode:          language,
	}, nil
}

func (s *LaunchService) ensureProviderPlayer(ctx context.Context, provider domain.GameProvider, req LaunchRequest) (*model.GameProviderPlayers, error) {
	scope := launchScope(req)
	existing, err := s.repo.PlayerMapping(ctx, scope, req.ProviderCode, req.UserID, req.Currency)
	if err == nil {
		return existing, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	mapping := &model.GameProviderPlayers{
		TenantID:               scope.TenantID,
		SiteID:                 scope.SiteID,
		ProviderCode:           req.ProviderCode,
		UserID:                 req.UserID,
		ProviderPlayerID:       providerPlayerID(req.ProviderCode, req.UserID),
		ProviderPlayerPassword: randomSecret(24),
		Currency:               req.Currency,
		Status:                 1,
	}
	_, err = provider.CreatePlayer(ctx, domain.CreatePlayerRequest{
		ProviderCode:           req.ProviderCode,
		UserID:                 fmt.Sprintf("%d", req.UserID),
		ExternalPlayerID:       mapping.ProviderPlayerID,
		ProviderPlayerPassword: mapping.ProviderPlayerPassword,
		Currency:               req.Currency,
	})
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreatePlayerMapping(ctx, mapping); err != nil {
		return nil, err
	}
	return mapping, nil
}

func (s *LaunchService) markLaunchEntered(ctx context.Context, log *model.GameGameLaunchLogs, result *domain.LaunchGameResult) error {
	if log == nil || log.ID == 0 {
		return nil
	}
	now := time.Now()
	responseJSON, _ := json.Marshal(result)
	requestTrace := result.ProviderRequestTrace
	responseTrace := result.ProviderResponseTrace
	if requestTrace == "" {
		requestTrace = log.ProviderLaunchRequest
	}
	if responseTrace == "" {
		responseTrace = string(responseJSON)
	}
	err := tenancy.ApplySiteStrict(s.repo.DB().WithContext(ctx).Model(&model.GameGameLaunchLogs{}), "", tenancy.Scope{TenantID: log.TenantID, SiteID: log.SiteID}).Where("id = ?", log.ID).Updates(map[string]any{
		"launch_status":            2,
		"session_status":           sessionStatusLaunchSuccess,
		"launch_success_time":      now,
		"last_heartbeat_time":      now,
		"error_message":            "",
		"provider_response":        string(responseJSON),
		"provider_launch_request":  requestTrace,
		"provider_launch_response": responseTrace,
	}).Error
	if err != nil {
		return err
	}
	s.recordTenantHotLaunch(ctx, log)
	return nil
}

func (s *LaunchService) recordTenantHotLaunch(ctx context.Context, log *model.GameGameLaunchLogs) {
	if s.hotRanker == nil || log == nil || log.TenantID == 0 || log.SiteID == 0 || log.GameID == 0 {
		return
	}

	// FindLaunchGame may intentionally fall back to the platform catalogue for
	// legacy launch compatibility. Those platform rows must not enter a
	// tenant's HOT ranking because the HOT query only exposes tenant imports.
	var game model.GameGames
	scope := tenancy.Scope{TenantID: log.TenantID, SiteID: log.SiteID}
	if err := tenancy.ApplySite(s.repo.DB().WithContext(ctx).Model(&model.GameGames{}), "", scope).
		Where("id = ? AND deleted_at IS NULL", log.GameID).
		First(&game).Error; err != nil {
		return
	}

	// Ranking is an auxiliary read model. A Redis failure must not turn a
	// successful provider launch into a failed game launch.
	_ = s.hotRanker.RecordLaunch(ctx, scope, game.ID)
}

func newLaunchLog(scope tenancy.Scope, req LaunchRequest, gameID, providerID uint64, requestJSON string, launchTime time.Time) *model.GameGameLaunchLogs {
	return &model.GameGameLaunchLogs{
		TenantID:              scope.TenantID,
		SiteID:                scope.SiteID,
		UserID:                req.UserID,
		GameID:                gameID,
		ProviderID:            providerID,
		LaunchNo:              "LCH-" + randomSecret(10),
		SessionID:             req.SessionID,
		LaunchType:            map[bool]string{true: "DEMO", false: "REAL"}[req.Demo],
		Currency:              req.Currency,
		Language:              req.Language,
		Country:               req.Country,
		Device:                req.Device,
		ProviderRequest:       requestJSON,
		ProviderLaunchRequest: requestJSON,
		CorrelationID:         req.SessionID,
		SessionToken:          req.SessionID,
		LaunchStatus:          1,
		SessionStatus:         sessionStatusLaunchRequested,
		LaunchTime:            &launchTime,
		ClientIP:              req.IPAddress,
	}
}

func (s *LaunchService) updateLaunchConfiguration(ctx context.Context, log *model.GameGameLaunchLogs) error {
	if log == nil || log.ID == 0 {
		return nil
	}
	return tenancy.ApplySiteStrict(s.repo.DB().WithContext(ctx).Model(&model.GameGameLaunchLogs{}), "", tenancy.Scope{TenantID: log.TenantID, SiteID: log.SiteID}).Where("id = ?", log.ID).Updates(map[string]any{
		"currency":               log.Currency,
		"language":               log.Language,
		"country":                log.Country,
		"country_id":             log.CountryID,
		"country_code_snapshot":  log.CountryCodeSnapshot,
		"currency_code_snapshot": log.CurrencyCodeSnapshot,
		"language_code_snapshot": log.LanguageCodeSnapshot,
	}).Error
}

func launchRequestJSON(req LaunchRequest) string {
	value := map[string]any{"providerCode": req.ProviderCode, "gameID": req.GameID, "userID": req.UserID, "currency": req.Currency, "language": req.Language, "country": req.Country, "device": req.Device, "returnURL": req.ReturnURL, "demo": req.Demo}
	raw, _ := json.Marshal(value)
	return string(raw)
}

func (s *LaunchService) markLaunchFailed(ctx context.Context, log *model.GameGameLaunchLogs, message string, requestTrace string, responseTrace string) error {
	if log == nil || log.ID == 0 {
		return nil
	}
	if len(message) > 500 {
		message = message[:500]
	}
	now := time.Now()
	return tenancy.ApplySiteStrict(s.repo.DB().WithContext(ctx).Model(&model.GameGameLaunchLogs{}), "", tenancy.Scope{TenantID: log.TenantID, SiteID: log.SiteID}).Where("id = ?", log.ID).Updates(map[string]any{
		"launch_status":            3,
		"session_status":           sessionStatusLaunchFailed,
		"close_time":               now,
		"error_message":            message,
		"provider_launch_request":  stringDefault(requestTrace, log.ProviderLaunchRequest),
		"provider_launch_response": responseTrace,
	}).Error
}

func launchErrorTrace(err error) (string, string) {
	var traced providerTraceError
	if errors.As(err, &traced) {
		return traced.ProviderRequestTrace(), traced.ProviderResponseTrace()
	}
	return "", ""
}

func (s *LaunchService) Heartbeat(ctx context.Context, scope tenancy.Scope, sessionID string, userID uint64) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" || userID == 0 {
		return domain.NewWalletError(domain.WalletErrorInvalidRequest, "invalid session", nil)
	}
	now := time.Now()
	_ = s.closeTimedOutSessions(ctx, now)
	return tenancy.ApplySite(s.repo.DB().WithContext(ctx).Model(&model.GameGameLaunchLogs{}), "", scope).
		Where("user_id=? AND (session_id=? OR session_token=?) AND deleted_at IS NULL AND session_status IN ?", userID, sessionID, sessionID, []string{sessionStatusLaunchSuccess, sessionStatusPlaying}).
		Updates(map[string]any{"session_status": sessionStatusPlaying, "last_heartbeat_time": now}).Error
}

func (s *LaunchService) Exit(ctx context.Context, scope tenancy.Scope, sessionID string, userID uint64, reason string) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" || userID == 0 {
		return domain.NewWalletError(domain.WalletErrorInvalidRequest, "invalid session", nil)
	}
	reason = normalizeExitReason(reason)
	return s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var log model.GameGameLaunchLogs
		if err := tenancy.ApplySite(tx.Model(&model.GameGameLaunchLogs{}), "", scope).Where("user_id=? AND (session_id=? OR session_token=?) AND deleted_at IS NULL", userID, sessionID, sessionID).Take(&log).Error; err != nil {
			return domain.NewWalletError(domain.WalletErrorInvalidRequest, "session not found", err)
		}
		if log.SessionStatus == sessionStatusClosed {
			return nil
		}
		now := time.Now()
		startedAt := log.LaunchTime
		if log.LaunchSuccessTime != nil {
			startedAt = log.LaunchSuccessTime
		}
		duration := uint64(0)
		if startedAt != nil && now.After(*startedAt) {
			duration = uint64(now.Sub(*startedAt).Seconds())
		}
		return tx.Model(&log).Updates(map[string]any{
			"session_status":   sessionStatusClosed,
			"exit_time":        now,
			"close_time":       now,
			"duration_seconds": duration,
			"exit_reason":      reason,
		}).Error
	})
}

func (s *LaunchService) closeTimedOutSessions(ctx context.Context, now time.Time) error {
	cutoff := now.Add(-sessionHeartbeatTimeout)
	activeStatuses := []string{sessionStatusLaunchRequested, sessionStatusLaunchSuccess, sessionStatusPlaying}
	var rows []model.GameGameLaunchLogs
	if err := s.repo.DB().WithContext(ctx).
		Select("id", "launch_time", "launch_success_time", "last_heartbeat_time", "created_at").
		Where("deleted_at IS NULL AND session_status IN ?", activeStatuses).
		Where(`(last_heartbeat_time IS NOT NULL AND last_heartbeat_time < ?)
			OR (last_heartbeat_time IS NULL AND launch_success_time IS NOT NULL AND launch_success_time < ?)
			OR (last_heartbeat_time IS NULL AND launch_success_time IS NULL AND launch_time IS NOT NULL AND launch_time < ?)
			OR (last_heartbeat_time IS NULL AND launch_success_time IS NULL AND launch_time IS NULL AND created_at < ?)`, cutoff, cutoff, cutoff, cutoff).
		Order("id ASC").
		Limit(sessionTimeoutSweepLimit).
		Find(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		startedAt := row.CreatedAt
		if row.LaunchTime != nil {
			startedAt = *row.LaunchTime
		}
		if row.LaunchSuccessTime != nil {
			startedAt = *row.LaunchSuccessTime
		}
		duration := uint64(0)
		if !startedAt.IsZero() && now.After(startedAt) {
			duration = uint64(now.Sub(startedAt).Seconds())
		}
		if err := s.repo.DB().WithContext(ctx).Model(&model.GameGameLaunchLogs{}).
			Where("id=? AND session_status IN ?", row.ID, activeStatuses).
			Updates(map[string]any{
				"session_status":   sessionStatusClosed,
				"exit_time":        now,
				"close_time":       now,
				"duration_seconds": duration,
				"exit_reason":      "TIMEOUT",
			}).Error; err != nil {
			return err
		}
	}
	return nil
}

func normalizeLaunchRequest(req LaunchRequest) LaunchRequest {
	req.ProviderCode = strings.ToLower(strings.TrimSpace(req.ProviderCode))
	req.GameID = strings.TrimSpace(req.GameID)
	req.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))
	req.Language = strings.ToLower(strings.TrimSpace(req.Language))
	req.Country = strings.ToUpper(strings.TrimSpace(req.Country))
	req.Device = strings.ToLower(strings.TrimSpace(req.Device))
	if req.Device == "" {
		req.Device = "web"
	}
	if req.SessionID == "" {
		req.SessionID = "launch-" + randomSecret(12)
	}
	return req
}

func launchScope(req LaunchRequest) tenancy.Scope {
	return tenancy.Scope{TenantID: req.TenantID, SiteID: req.SiteID}
}

func validateBasicLaunchRequest(req LaunchRequest) error {
	if req.ProviderCode == "" || req.GameID == "" || req.UserID == 0 {
		return domain.NewWalletError(domain.WalletErrorInvalidRequest, "invalid launch request", nil)
	}
	return nil
}

func validateLaunchRequest(req LaunchRequest) error {
	if err := validateBasicLaunchRequest(req); err != nil {
		return err
	}
	if !validCurrency(req.Currency) {
		return domain.NewWalletError(domain.WalletErrorCurrencyMismatch, "invalid currency", nil)
	}
	if len(req.Language) < 2 || len(req.Language) > 8 {
		return domain.NewWalletError(domain.WalletErrorInvalidRequest, "invalid language", nil)
	}
	return nil
}

func validCurrency(currency string) bool {
	if len(currency) != 3 {
		return false
	}
	for _, r := range currency {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

func supportedLanguages(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var values []string
	if strings.HasPrefix(raw, "[") {
		_ = json.Unmarshal([]byte(raw), &values)
	} else {
		values = strings.Split(raw, ",")
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}

func languageSupported(language string, supported []string) bool {
	if len(supported) == 0 {
		return true
	}
	for _, item := range supported {
		if strings.EqualFold(language, item) {
			return true
		}
	}
	return false
}

func normalizeExitReason(reason string) string {
	reason = strings.ToUpper(strings.TrimSpace(reason))
	switch reason {
	case "USER_EXIT", "TIMEOUT", "PROVIDER_EXIT", "NETWORK_ERROR", "SYSTEM_CLOSE", "UNKNOWN":
		return reason
	default:
		return "UNKNOWN"
	}
}

func providerPlayerID(providerCode string, userID uint64) string {
	switch providerCode {
	case "softaggregator":
		return fmt.Sprintf("sa_%d", userID)
	default:
		return fmt.Sprintf("%s_%d", providerCode, userID)
	}
}

func randomSecret(byteLen int) string {
	buf := make([]byte, byteLen)
	if _, err := rand.Read(buf); err != nil {
		return decimal.NewFromInt(int64(byteLen)).String() + "-fallback"
	}
	return hex.EncodeToString(buf)
}
