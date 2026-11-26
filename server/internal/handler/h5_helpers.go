package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"casino/internal/content"
	"casino/internal/database"
	"casino/internal/hotranking"
	"casino/internal/model"
	"casino/internal/storage"
)

func h5HotRankingStore() (store *hotranking.Store) {
	defer func() {
		if recover() != nil {
			store = hotranking.New(nil)
		}
	}()
	return hotranking.New(database.GetRedisCli())
}

func h5Error(c *gin.Context, statusCode int, msg string) {
	c.JSON(http.StatusOK, gin.H{
		"code": statusCode,
		"msg":  msg,
		"data": nil,
	})
}

func h5MachineError(c *gin.Context, statusCode int, code string, msg string) {
	c.JSON(http.StatusOK, gin.H{
		"code": statusCode,
		"msg":  msg,
		"data": gin.H{"errorCode": code},
	})
}

func h5CurrentUser(db *gorm.DB, c *gin.Context) (model.GameUsers, model.GameUserWallets) {
	if user, wallet, ok := h5AuthenticatedUser(db, c); ok {
		return user, wallet
	}
	user := h5DefaultUser()
	return user, h5EnsureWalletForScope(db, 0, h5ContentScope(c))
}

func h5EnsureWallet(db *gorm.DB, userID uint64) model.GameUserWallets {
	return h5EnsureWalletForScope(db, userID, content.Scope{})
}

func h5EnsureWalletForScope(db *gorm.DB, userID uint64, fallback content.Scope) model.GameUserWallets {
	var wallet model.GameUserWallets
	scope := h5UserContentScope(db, userID, fallback)
	defaults := h5ScopeDefaults(db, scope)
	currency := defaults.CurrencyCode
	if userID == 0 {
		return model.GameUserWallets{

			SiteID:              scope.SiteID,
			Currency:            currency,
			CashBalance:         h5DecimalPtr(decimal.Zero),
			BonusBalance:        h5DecimalPtr(decimal.Zero),
			WithdrawableBalance: h5DecimalPtr(decimal.Zero),
			FrozenBalance:       h5DecimalPtr(decimal.Zero),
		}
	}
	err := content.ApplySite(db.Model(&model.GameUserWallets{}), "", scope).Where("user_id = ?", userID).First(&wallet).Error
	if err == nil {
		updates := map[string]any{}
		if wallet.SiteID == 0 && scope.SiteID > 0 {
			updates["site_id"] = scope.SiteID
			wallet.SiteID = scope.SiteID
			wallet.SiteID = scope.SiteID
		}
		if currency != "" && !strings.EqualFold(wallet.Currency, currency) {
			updates["currency"] = currency
			wallet.Currency = currency
		}
		if len(updates) > 0 {
			_ = db.Model(&wallet).Updates(updates).Error
		}
		if currency != "" {
			_ = content.ApplySite(db.Model(&model.GameRegisterRewardClaims{}), "", scope).
				Where("user_id = ? AND status = 1 AND deleted_at IS NULL", userID).
				Update("currency", currency).Error
		}
		return wallet
	}
	wallet = model.GameUserWallets{

		SiteID:              scope.SiteID,
		UserID:              userID,
		Currency:            currency,
		CashBalance:         h5DecimalPtr(decimal.Zero),
		BonusBalance:        h5DecimalPtr(decimal.Zero),
		WithdrawableBalance: h5DecimalPtr(decimal.Zero),
		FrozenBalance:       h5DecimalPtr(decimal.Zero),
	}
	_ = db.Create(&wallet).Error
	return wallet
}

type h5ScopeDefaultValues struct {
	CountryCode      string
	CurrencyCode     string
	Language         string
	PhoneCountryCode string
}

func h5ScopeDefaults(db *gorm.DB, scope content.Scope) h5ScopeDefaultValues {
	site := h5BootstrapSite(db, scope.SiteID)
	countryCode := strings.ToUpper(strings.TrimSpace(h5FirstNonEmpty(site.CountryCode)))
	country := h5BootstrapCountry(db, countryCode)
	countryCode = strings.ToUpper(strings.TrimSpace(h5FirstNonEmpty(site.CountryCode, country.CountryCode, "PK")))
	currencyCode := strings.ToUpper(strings.TrimSpace(h5FirstNonEmpty(site.CurrencyCode, country.CurrencyCode, "PKR")))
	language := strings.ToLower(strings.TrimSpace(h5FirstNonEmpty(site.DefaultLanguage, country.DefaultLanguage, "en")))
	return h5ScopeDefaultValues{
		CountryCode:      countryCode,
		CurrencyCode:     currencyCode,
		Language:         language,
		PhoneCountryCode: h5FirstNonEmpty(country.PhoneCountryCode, "+92"),
	}
}

func h5ContentScope(c *gin.Context) content.Scope { return content.Scope{SiteID: 1} }

func h5PlatformCatalogScope() content.Scope {
	return content.Scope{Platform: true}
}

func h5ApplyReadableScope(db *gorm.DB, alias string, scope content.Scope) *gorm.DB {
	if scope.Platform {
		return db.Where(h5Column(alias, "site_id")+" = ?", 0)
	}
	return content.ApplySite(db, alias, scope)
}

func h5FindWithPlatformFallback(base *gorm.DB, alias string, scope content.Scope, dest any, apply func(*gorm.DB) *gorm.DB) error {
	query := h5ApplyReadableScope(base.Session(&gorm.Session{}), alias, scope)
	if apply != nil {
		query = apply(query)
	}
	if err := query.Find(dest).Error; err != nil {
		return err
	}
	if h5CollectionLen(dest) > 0 || !scope.SiteScoped() {
		return nil
	}
	query = h5ApplyReadableScope(base.Session(&gorm.Session{}), alias, h5PlatformCatalogScope())
	if apply != nil {
		query = apply(query)
	}
	return query.Find(dest).Error
}

func h5CollectionLen(value any) int {
	reflected := reflect.ValueOf(value)
	if !reflected.IsValid() {
		return 0
	}
	if reflected.Kind() == reflect.Pointer {
		if reflected.IsNil() {
			return 0
		}
		reflected = reflected.Elem()
	}
	if reflected.Kind() == reflect.Slice || reflected.Kind() == reflect.Array {
		return reflected.Len()
	}
	return 0
}

func h5Column(alias string, name string) string {
	if strings.TrimSpace(alias) == "" {
		return name
	}
	return alias + "." + name
}

func h5UserContentScope(db *gorm.DB, userID uint64, fallback content.Scope) content.Scope {
	if userID == 0 {
		return fallback
	}
	var row struct {
		SiteID uint64 `gorm:"column:site_id"`
	}
	if err := db.Model(&model.GameUsers{}).Select("site_id").Where("id=? AND deleted_at IS NULL", userID).Take(&row).Error; err == nil && row.SiteID > 0 {
		return content.Scope{SiteID: row.SiteID}
	}
	return fallback
}

func h5MaintenanceBlocked(c *gin.Context, db *gorm.DB, scope content.Scope) (bool, string, string) {
	if scope.SiteID == 0 {
		return false, "", ""
	}
	maintenance := h5EffectiveMaintenance(c, db, scope)
	enabled, _ := maintenance["enabled"].(bool)
	if !enabled {
		return false, "", ""
	}
	message, _ := maintenance["message"].(string)
	if message == "" {
		message = "Site is under maintenance"
	}
	code, _ := maintenance["errorCode"].(string)
	if code == "" {
		code = "SITE_MAINTENANCE"
	}
	return true, code, message
}

func h5EffectiveMaintenance(c *gin.Context, db *gorm.DB, scope content.Scope) map[string]any {
	site := h5BootstrapSite(db, scope.SiteID)
	fallback := h5NormalizeMaintenance(h5BootstrapJSON(site.MaintenanceConfig), site)
	if scope.SiteID == 0 {
		return fallback
	}
	type maintenanceRow struct {
		ScopeType           string     `gorm:"column:scope_type"`
		SiteID              uint64     `gorm:"column:site_id"`
		Title               string     `gorm:"column:title"`
		MessageZh           string     `gorm:"column:message_zh"`
		MessageEn           string     `gorm:"column:message_en"`
		EstimatedRecoveryAt *time.Time `gorm:"column:estimated_recovery_at"`
		SupportURL          string     `gorm:"column:support_url"`
		SupportEmail        string     `gorm:"column:support_email"`
		SupportChat         string     `gorm:"column:support_chat"`
		AllowlistedIPs      string     `gorm:"column:allowlisted_ips"`
	}
	var row maintenanceRow
	nowValue := time.Now()
	err := db.Table("game_site_maintenance").Where("enabled = 1 AND deleted_at IS NULL AND (site_id = 0 OR site_id = ?)", scope.SiteID).Where("start_at IS NULL OR start_at<=?", nowValue).
		Where("end_at IS NULL OR end_at>?", nowValue).
		Order("site_id DESC,id DESC").Take(&row).Error
	if err != nil {
		return fallback
	}
	if h5IPAllowed(c.ClientIP(), row.AllowlistedIPs) {
		return map[string]any{"enabled": false, "bypassed": true}
	}
	language := strings.ToLower(strings.TrimSpace(c.Query("language")))
	message := row.MessageEn
	if strings.HasPrefix(language, "zh") || message == "" {
		message = row.MessageZh
	}
	if message == "" {
		message = "Site is under maintenance"
	}
	errorCode := "SITE_MAINTENANCE"
	if row.ScopeType == "SITE" || row.SiteID > 0 {
		errorCode = "SITE_MAINTENANCE"
	}
	return map[string]any{
		"enabled": true, "scope": row.ScopeType, "errorCode": errorCode, "title": row.Title, "message": message,
		"messageZh": row.MessageZh, "messageEn": row.MessageEn, "estimatedRecoveryAt": row.EstimatedRecoveryAt,
		"supportURL": row.SupportURL, "supportEmail": row.SupportEmail, "supportChat": row.SupportChat,
	}
}

func h5IPAllowed(ip string, csv string) bool {
	parsedIP := net.ParseIP(strings.TrimSpace(ip))
	if parsedIP == nil {
		return false
	}
	for _, value := range strings.Split(csv, ",") {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if candidate := net.ParseIP(value); candidate != nil && candidate.Equal(parsedIP) {
			return true
		}
		if _, network, err := net.ParseCIDR(value); err == nil && network.Contains(parsedIP) {
			return true
		}
	}
	return false
}

func h5ProviderMaintenanceHidden(db *gorm.DB, scope content.Scope, providerCode string, country string, currency string) bool {
	var count int64
	nowValue := time.Now()
	_ = db.Table("game_provider_maintenance").Where(
		"provider_code=? AND enabled=1 AND hide_from_lobby=1 AND deleted_at IS NULL", strings.ToUpper(providerCode),
	).Where("(site_id = 0) OR ((site_id = 0 OR site_id = ?))", scope.SiteID).
		Where("start_at IS NULL OR start_at<=?", nowValue).Where("end_at IS NULL OR end_at>?", nowValue).
		Where("affected_countries='' OR FIND_IN_SET('ALL',affected_countries)>0 OR FIND_IN_SET(?,affected_countries)>0", strings.ToUpper(country)).
		Where("affected_currencies='' OR FIND_IN_SET('ALL',affected_currencies)>0 OR FIND_IN_SET(?,affected_currencies)>0", strings.ToUpper(currency)).
		Count(&count).Error
	return count > 0
}

func h5ApplyGameMaintenance(db *gorm.DB, scope content.Scope, country string, currency string) *gorm.DB {
	nowValue := time.Now()
	providerSubquery := `NOT EXISTS(SELECT 1 FROM game_provider_maintenance AS pm WHERE pm.provider_code = gp.provider_code AND pm.enabled = 1 AND pm.hide_from_lobby = 1 AND pm.deleted_at IS NULL AND ((pm.site_id = 0) OR ((pm.site_id = 0 OR pm.site_id = ?))) AND (pm.start_at IS NULL OR pm.start_at <= ?) AND (pm.end_at IS NULL OR pm.end_at > ?) AND (pm.affected_countries = '' OR FIND_IN_SET('ALL', pm.affected_countries) > 0 OR FIND_IN_SET(?, pm.affected_countries) > 0) AND (pm.affected_currencies = '' OR FIND_IN_SET('ALL', pm.affected_currencies) > 0 OR FIND_IN_SET(?, pm.affected_currencies) > 0))`

	gameSubquery := `NOT EXISTS(SELECT 1 FROM game_game_maintenance AS gm WHERE gm.game_id = game_games.id AND gm.site_id = ? AND gm.enabled = 1 AND gm.hide_from_lobby = 1 AND gm.deleted_at IS NULL AND (gm.start_at IS NULL OR gm.start_at <= ?) AND (gm.end_at IS NULL OR gm.end_at > ?) AND (gm.affected_countries = '' OR FIND_IN_SET('ALL', gm.affected_countries) > 0 OR FIND_IN_SET(?, gm.affected_countries) > 0) AND (gm.affected_currencies = '' OR FIND_IN_SET('ALL', gm.affected_currencies) > 0 OR FIND_IN_SET(?, gm.affected_currencies) > 0))`

	return db.Where(providerSubquery, scope.SiteID, nowValue, nowValue, strings.ToUpper(country), strings.ToUpper(currency)).
		Where(gameSubquery, scope.SiteID, nowValue, nowValue, strings.ToUpper(country), strings.ToUpper(currency))
}

func h5RequestHost(c *gin.Context) string {
	host := strings.TrimSpace(c.GetHeader("X-Site-Domain"))
	if host == "" {
		host = strings.TrimSpace(c.GetHeader("X-Forwarded-Host"))
	}
	if host == "" {
		host = c.Request.Host
	}
	host = strings.TrimSpace(strings.Split(host, ",")[0])
	if idx := strings.Index(host, "://"); idx >= 0 {
		host = host[idx+3:]
	}
	host = strings.Split(host, "/")[0]
	if parsed, _, err := net.SplitHostPort(host); err == nil {
		host = parsed
	} else if strings.Count(host, ":") == 1 {
		host = strings.Split(host, ":")[0]
	}
	return strings.ToLower(strings.TrimSpace(host))
}

func h5DomainCandidates(host string) []string {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return nil
	}
	if strings.HasPrefix(host, "www.") {
		return []string{host, strings.TrimPrefix(host, "www.")}
	}
	return []string{host, "www." + host}
}

func h5PlayerPayload(user model.GameUsers, wallet model.GameUserWallets) gin.H {
	return gin.H{
		"id":               user.ID,
		"userNo":           user.UserNo,
		"phoneCountryCode": user.PhoneCountryCode,
		"phone":            user.Phone,
		"nickname":         h5FirstNonEmpty(user.Nickname, h5MaskedPhone(user.Phone)),
		"avatarUrl":        h5AssetURL(user.AvatarURL),
		"inviteCode":       user.InviteCode,
		"vipLevel":         user.VipLevel,
		"status":           user.Status,
		"wallet":           h5WalletPayload(wallet),
	}
}

func h5WalletPayload(wallet model.GameUserWallets) gin.H {
	return gin.H{
		"currency":            h5FirstNonEmpty(wallet.Currency, "PKR"),
		"cashBalance":         h5DecimalString(wallet.CashBalance),
		"bonusBalance":        h5DecimalString(wallet.BonusBalance),
		"withdrawableBalance": h5DecimalString(wallet.WithdrawableBalance),
		"frozenBalance":       h5DecimalString(wallet.FrozenBalance),
	}
}

func h5DefaultUser() model.GameUsers {
	return model.GameUsers{
		UserNo:           "GUEST",
		PhoneCountryCode: "+92",
		Phone:            "3088888889",
		Nickname:         "Guest",
		InviteCode:       "WELCOME",
		Status:           1,
	}
}

func h5FindSettings(db *gorm.DB, userID uint64) model.GameUserSettings {
	var settings model.GameUserSettings
	if userID > 0 {
		_ = db.Where("user_id = ?", userID).First(&settings).Error
	}
	if settings.UserID == 0 {
		scope := h5UserContentScope(db, userID, content.Scope{})
		settings.SiteID = scope.SiteID
		settings.SiteID = scope.SiteID
		settings.UserID = userID
		settings.Language = h5ScopeDefaults(db, scope).Language
		settings.MusicEnabled = 1
		settings.PushEnabled = 1
	} else if settings.SiteID == 0 {
		scope := h5UserContentScope(db, userID, content.Scope{})
		settings.SiteID = scope.SiteID
		settings.SiteID = scope.SiteID
		if scope.SiteID > 0 {
			_ = db.Model(&settings).Updates(map[string]any{"site_id": scope.SiteID}).Error
		}
	}
	return settings
}

func h5FindWithdrawAccounts(db *gorm.DB, userID uint64) []model.GameWithdrawAccounts {
	var accounts []model.GameWithdrawAccounts
	if userID > 0 {
		_ = content.ApplySite(db.Model(&model.GameWithdrawAccounts{}), "", h5UserContentScope(db, userID, content.Scope{})).
			Where("user_id = ? AND status <> ?", userID, 2).
			Order("is_default DESC, id DESC").
			Find(&accounts).Error
	}
	return accounts
}

func h5GameType(category string) string {
	switch strings.ToLower(category) {
	case "slots", "slot":
		return "slot"
	case "casino", "chance", "live":
		return "chance"
	case "live-casino", "live_casino", "live casino":
		return "live"
	case "table-games", "table_games", "table games", "table":
		return "table"
	case "crash", "instant":
		return "crash"
	case "fishing", "fish":
		return "fish"
	case "sports", "sport":
		return "sports"
	default:
		return strings.ToLower(category)
	}
}

func h5Limit(c *gin.Context, defaultLimit int) int {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(defaultLimit)))
	if err != nil || limit <= 0 {
		return defaultLimit
	}
	if limit > 200 {
		return 200
	}
	return limit
}

func h5QueryUint64(c *gin.Context, key string) uint64 {
	v, _ := strconv.ParseUint(strings.TrimSpace(c.Query(key)), 10, 64)
	return v
}

func h5HeaderUint64(c *gin.Context, key string) uint64 {
	v, _ := strconv.ParseUint(strings.TrimSpace(c.GetHeader(key)), 10, 64)
	return v
}

func h5PathUint64(c *gin.Context, key string) uint64 {
	v, _ := strconv.ParseUint(strings.TrimSpace(c.Param(key)), 10, 64)
	return v
}

func h5ParseMoney(value string) (decimal.Decimal, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return decimal.Zero, errors.New("empty amount")
	}
	return decimal.NewFromString(value)
}

func h5DecimalPtr(value decimal.Decimal) *decimal.Decimal {
	v := value
	return &v
}

func h5DecimalValue(value *decimal.Decimal) decimal.Decimal {
	if value == nil {
		return decimal.Zero
	}
	return *value
}

func h5DecimalString(value *decimal.Decimal) string {
	if value == nil {
		return "0.00"
	}
	return value.StringFixed(2)
}

func h5TimePtr(value time.Time) *time.Time {
	v := value
	return &v
}

func h5FirstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func h5CleanPhone(value string) string {
	var b strings.Builder
	for _, r := range value {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func h5MaskedPhone(phone string) string {
	phone = h5CleanPhone(phone)
	if len(phone) <= 4 {
		return phone
	}
	return phone[:1] + strings.Repeat("*", len(phone)-3) + phone[len(phone)-2:]
}

func h5OrderNo(prefix string) string {
	return fmt.Sprintf("%s%s%s", prefix, time.Now().Format("20060102150405"), h5RandomHex(3))
}

func h5UserNo() string {
	return fmt.Sprintf("U%s%s", time.Now().Format("060102"), h5RandomHex(3))
}

func h5Token(prefix string) string {
	return prefix + h5RandomHex(12)
}

func h5RandomHex(byteLen int) string {
	buf := make([]byte, byteLen)
	if _, err := rand.Read(buf); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(buf)
}

func h5FrontendOrigin(c *gin.Context) string {
	if origin := c.GetHeader("Origin"); origin != "" {
		return origin
	}
	host := c.Request.Host
	if host == "" {
		host = "localhost"
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + host
}

func h5Title(value string) string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == '-' || r == '_' || r == '/'
	})
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func h5ConfigJSON(c *gin.Context, db *gorm.DB, key string, fallback interface{}) interface{} {
	var configs []model.GameSystemConfigs
	if err := h5FindWithPlatformFallback(db.Model(&model.GameSystemConfigs{}), "", h5ContentScope(c), &configs, func(query *gorm.DB) *gorm.DB {
		return query.Where("config_key = ? AND status = ?", key, 1).Order("site_id DESC, id DESC").Limit(1)
	}); err != nil || len(configs) == 0 {
		return fallback
	}
	var out interface{}
	if err := json.Unmarshal([]byte(configs[0].ConfigValue), &out); err != nil {
		return fallback
	}
	return out
}

func h5ConfigJSONAssets(c *gin.Context, db *gorm.DB, key string, fallback interface{}) interface{} {
	return h5SignJSONAssetReferences(h5ConfigJSON(c, db, key, fallback))
}

func h5SignJSONAssetReferences(value interface{}) interface{} {
	switch typed := value.(type) {
	case []interface{}:
		out := make([]interface{}, 0, len(typed))
		for _, item := range typed {
			out = append(out, h5SignJSONAssetReferences(item))
		}
		return out
	case []gin.H:
		out := make([]gin.H, 0, len(typed))
		for _, item := range typed {
			converted, _ := h5SignJSONAssetReferences(item).(gin.H)
			out = append(out, converted)
		}
		return out
	case map[string]interface{}:
		out := gin.H{}
		for key, item := range typed {
			out[key] = h5SignJSONAssetReferences(item)
		}
		return out
	case gin.H:
		out := gin.H{}
		for key, item := range typed {
			out[key] = h5SignJSONAssetReferences(item)
		}
		return out
	case string:
		if h5LooksLikeAssetReference(typed) {
			return h5AssetURL(typed)
		}
		return typed
	default:
		return value
	}
}

func h5LooksLikeAssetReference(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || strings.HasPrefix(value, "data:") {
		return false
	}
	if strings.HasPrefix(value, "/static/") || strings.HasPrefix(value, "oss://") {
		return true
	}
	cleanValue := strings.ToLower(strings.SplitN(value, "?", 2)[0])
	for _, extension := range []string{".avif", ".webp", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".json", ".mp3", ".wav", ".ogg", ".woff", ".woff2"} {
		if strings.HasSuffix(cleanValue, extension) {
			return true
		}
	}
	return false
}

func h5SocialLinks(c *gin.Context, db *gorm.DB) []gin.H {
	var rows []model.GameSocialLinks
	scope := h5ContentScope(c)
	if err := content.ApplySite(db.Model(&model.GameSocialLinks{}), "", scope).Where("deleted_at IS NULL").Order("site_id DESC,sort ASC,id ASC").Find(&rows).Error; err != nil {
		return []gin.H{}
	}

	// Current H5 settings override imported defaults, including disabled links.
	// Select the effective row before applying status so defaults cannot reappear.
	seen := make(map[string]bool, len(rows))
	effective := make([]model.GameSocialLinks, 0, len(rows))
	for _, row := range rows {
		code := strings.ToLower(strings.TrimSpace(row.PlatformCode))
		if code == "" || seen[code] {
			continue
		}
		seen[code] = true
		if row.Status != 1 {
			continue
		}
		row.PlatformCode = code
		effective = append(effective, row)
	}
	sort.SliceStable(effective, func(i, j int) bool {
		if effective[i].Sort == effective[j].Sort {
			return effective[i].ID < effective[j].ID
		}
		return effective[i].Sort < effective[j].Sort
	})
	links := make([]gin.H, 0, len(effective))
	for _, row := range effective {
		links = append(links, gin.H{
			"id":               row.PlatformCode,
			"platformCode":     row.PlatformCode,
			"label":            row.Label,
			"icon":             h5AssetURL(row.IconURL),
			"image":            h5AssetURL(row.IconURL),
			"url":              row.ShareURL,
			"shareURL":         row.ShareURL,
			"shareTitle":       row.ShareTitle,
			"shareText":        row.ShareText,
			"shareDescription": row.ShareDescription,
			"sort":             row.Sort,
		})
	}
	return links
}

func h5AssetURL(value string) string {
	out, err := storage.PublicURL(context.Background(), value)
	if err != nil {
		return value
	}
	return out
}
