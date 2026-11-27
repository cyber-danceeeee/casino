package handler

import (
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"github.com/go-sql-driver/mysql"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/content"
	"casino/internal/database"
	"casino/internal/model"
)

type gameUsersAuthRequest struct {
	Phone       string `json:"phone"`
	Password    string `json:"password"`
	InviteCode  string `json:"inviteCode"`
	CountryCode string `json:"phoneCountryCode"`
}

var errH5PhoneAlreadyRegistered = errors.New("phone already registered")

func (h *gameUsersHandler) Login(c *gin.Context) {
	req := &gameUsersAuthRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		recordGameUserLogin(c, 0, "", false, "INVALID_REQUEST")
		h5Error(c, http.StatusBadRequest, "invalid request")
		return
	}

	user, wallet, err := authenticateGameUser(c, req)
	if err != nil {
		recordGameUserLogin(c, 0, req.Phone, false, "INVALID_LOGIN_IDENTIFIER")
		h5Error(c, http.StatusBadRequest, "invalid phone or password")
		return
	}
	if user.Status != 1 || user.LoginEnabled != 1 {
		recordGameUserLogin(c, user.ID, user.Phone, false, "LOGIN_DISABLED")
		h5Error(c, http.StatusForbidden, "login is disabled")
		return
	}
	if blockedGameUserDevice(c, user.ID) {
		recordGameUserLogin(c, user.ID, user.Phone, false, "DEVICE_BLOCKED")
		h5Error(c, http.StatusForbidden, "device is blocked")
		return
	}

	now := time.Now()
	db := database.GetDB()
	_ = content.ApplySite(db.Model(&model.GameUsers{}), "", h5UserContentScope(db, user.ID, h5ContentScope(c))).Where("id = ?", user.ID).Updates(map[string]interface{}{
		"last_login_at":      now,
		"last_login_ip":      c.ClientIP(),
		"failed_login_count": 0,
	}).Error
	user.LastLoginAt = &now
	upsertGameUserDevice(c, user.ID, now)
	recordGameUserLogin(c, user.ID, user.Phone, true, "")
	token, err := h5IssueGameUserSession(db, c, user.ID)
	if err != nil {
		h5Error(c, http.StatusInternalServerError, "session create failed")
		return
	}

	response.Success(c, gin.H{
		"token": token,
		"user":  h5PlayerPayload(user, wallet),
	})
}

func (h *gameUsersHandler) Register(c *gin.Context) {
	req := &gameUsersAuthRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		h5Error(c, http.StatusBadRequest, "invalid request")
		return
	}

	user, wallet, isNewUser, err := registerGameUser(c, req)
	if err != nil {
		if errors.Is(err, errH5PhoneAlreadyRegistered) {
			h5MachineError(c, http.StatusBadRequest, "PHONE_ALREADY_REGISTERED", errH5PhoneAlreadyRegistered.Error())
			return
		}
		if h5RegistrationValidationError(err) {
			h5Error(c, http.StatusBadRequest, err.Error())
			return
		}
		h5MachineError(c, http.StatusInternalServerError, "REGISTRATION_FAILED", "registration failed")
		return
	}
	now := time.Now()
	upsertGameUserDevice(c, user.ID, now)
	token, err := h5IssueGameUserSession(database.GetDB(), c, user.ID)
	if err != nil {
		h5Error(c, http.StatusInternalServerError, "session create failed")
		return
	}

	payload := gin.H{
		"token": token,
		"user":  h5PlayerPayload(user, wallet),
	}
	if isNewUser {
		if _, err := ensurePendingRegisterReward(user, wallet); err != nil {
			h5Error(c, http.StatusInternalServerError, "register reward pending failed")
			return
		}
	}

	response.Success(c, payload)
}

func (h *gameUsersHandler) Logout(c *gin.Context) {
	h5RevokeGameUserSession(database.GetDB(), c)
	response.Success(c)
}

func (h *gameUsersHandler) Me(c *gin.Context) {
	user, wallet, ok := h5AuthenticatedUser(database.GetDB(), c)
	if !ok {
		h5Error(c, http.StatusUnauthorized, "authentication required")
		return
	}
	response.Success(c, h5PlayerPayload(user, wallet))
}

func (h *gameUsersHandler) Profile(c *gin.Context) {
	db := database.GetDB()
	user, wallet, ok := h5AuthenticatedUser(db, c)
	if !ok {
		h5Error(c, http.StatusUnauthorized, "authentication required")
		return
	}
	response.Success(c, gin.H{
		"user":             h5PlayerPayload(user, wallet),
		"settings":         h5FindSettings(db, user.ID),
		"withdrawAccounts": h5FindWithdrawAccounts(db, user.ID),
		"quickMenus": []gin.H{
			{"key": "deposit", "label": "Deposit", "route": "/wallet/deposit"},
			{"key": "withdraw", "label": "Withdraw", "route": "/cash-center"},
			{"key": "records", "label": "Records", "route": "/wallet/records"},
			{"key": "security", "label": "Security", "route": "/security/bank-account"},
			{"key": "support", "label": "Support", "route": "/support"},
		},
	})
}

func (h *gameUsersHandler) PendingRewards(c *gin.Context) {
	db := database.GetDB()
	user, wallet, ok := h5AuthenticatedUser(db, c)
	if !ok {
		h5Error(c, http.StatusUnauthorized, "authentication required")
		return
	}
	registerReward, err := pendingRegisterRewardPayload(db, user, wallet)
	if err != nil {
		h5Error(c, http.StatusInternalServerError, "pending rewards failed")
		return
	}
	items, totalAmount, err := pendingActivityRewardsPayload(db, user, wallet)
	if err != nil {
		h5Error(c, http.StatusInternalServerError, "pending rewards failed")
		return
	}
	response.Success(c, gin.H{
		"registerReward": registerReward,
		"items":          items,
		"count":          len(items),
		"totalAmount":    totalAmount.StringFixed(2),
		"currency":       h5FirstNonEmpty(wallet.Currency, "PKR"),
	})
}

func (h *gameUsersHandler) ClaimPendingReward(c *gin.Context) {
	recordID, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || recordID == 0 {
		h5Error(c, http.StatusBadRequest, "invalid reward")
		return
	}

	db := database.GetDB()
	user, _, ok := h5AuthenticatedUser(db, c)
	if !ok {
		h5Error(c, http.StatusUnauthorized, "authentication required")
		return
	}

	wallet, reward, err := claimPendingActivityReward(user, recordID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h5Error(c, http.StatusBadRequest, "reward is no longer available")
			return
		}
		h5Error(c, http.StatusInternalServerError, "claim reward failed")
		return
	}

	response.Success(c, gin.H{
		"reward": reward,
		"user":   h5PlayerPayload(user, wallet),
	})
}

func (h *gameUsersHandler) ClaimRegisterReward(c *gin.Context) {
	db := database.GetDB()
	user, _, ok := h5AuthenticatedUser(db, c)
	if !ok {
		h5Error(c, http.StatusUnauthorized, "authentication required")
		return
	}

	wallet, reward, err := claimPendingRegisterReward(user)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h5Error(c, http.StatusBadRequest, "no pending register reward")
			return
		}
		h5Error(c, http.StatusInternalServerError, "claim register reward failed")
		return
	}

	response.Success(c, gin.H{
		"registerReward": reward,
		"user":           h5PlayerPayload(user, wallet),
	})
}

func h5AuthenticatedUser(db *gorm.DB, c *gin.Context) (model.GameUsers, model.GameUserWallets, bool) {
	var user model.GameUsers
	token := h5BearerToken(c)
	if token == "" {
		return model.GameUsers{}, model.GameUserWallets{}, false
	}
	scope := h5ContentScope(c)
	query := db.Table("game_user_sessions AS s").
		Select("u.*").
		Joins("INNER JOIN game_users AS u ON u.id = s.user_id AND u.deleted_at IS NULL AND u.status = ? AND u.login_enabled = ?", 1, 1).
		Where("s.token_hash = ? AND s.revoked_at IS NULL AND s.expired_at > ? AND s.deleted_at IS NULL", h5UserSessionHash(token), time.Now())
	query = content.ApplySite(query, "s", scope)
	query = content.ApplySite(query, "u", scope)
	err := query.Take(&user).Error
	if err != nil || user.ID == 0 {
		return model.GameUsers{}, model.GameUserWallets{}, false
	}
	return user, h5EnsureWallet(db, user.ID), true
}

func authenticateGameUser(c *gin.Context, req *gameUsersAuthRequest) (model.GameUsers, model.GameUserWallets, error) {
	db := database.GetDB()
	scope := h5ContentScope(c)
	phone, err := h5AuthPhone(req.Phone)
	if err != nil {
		return model.GameUsers{}, model.GameUserWallets{}, err
	}
	phoneCountryCode, err := h5AuthPhoneCountryCode(req.CountryCode)
	if err != nil {
		return model.GameUsers{}, model.GameUserWallets{}, err
	}
	if strings.TrimSpace(req.Password) == "" {
		return model.GameUsers{}, model.GameUserWallets{}, errors.New("password is required")
	}
	user, err := h5FindGameUserByPhone(db, scope, phoneCountryCode, phone)
	if err != nil {
		return model.GameUsers{}, model.GameUserWallets{}, err
	}
	if !h5VerifyGameUserPassword(db, user, req.Password) {
		_ = content.ApplySiteStrict(db.Model(&model.GameUsers{}), "", scope).Where("id = ?", user.ID).Update("failed_login_count", gorm.Expr("failed_login_count + 1")).Error
		return model.GameUsers{}, model.GameUserWallets{}, errors.New("invalid phone or password")
	}

	wallet := h5EnsureWallet(db, user.ID)
	settings := h5FindSettings(db, user.ID)
	if settings.ID == 0 {
		_ = db.Create(&settings).Error
	}
	return user, wallet, nil
}

func registerGameUser(c *gin.Context, req *gameUsersAuthRequest) (model.GameUsers, model.GameUserWallets, bool, error) {
	db := database.GetDB()
	scope := h5ContentScope(c)
	defaults := h5ScopeDefaults(db, scope)
	phone, err := h5AuthPhone(req.Phone)
	if err != nil {
		return model.GameUsers{}, model.GameUserWallets{}, false, err
	}
	phoneCountryCode, err := h5AuthPhoneCountryCode(req.CountryCode)
	if err != nil {
		return model.GameUsers{}, model.GameUserWallets{}, false, err
	}
	passwordHash, err := h5PasswordHash(req.Password)
	if err != nil {
		return model.GameUsers{}, model.GameUserWallets{}, false, err
	}

	existing, findErr := h5FindGameUserByPhone(db, scope, phoneCountryCode, phone)
	if findErr == nil {
		if !h5VerifyGameUserPassword(db, existing, req.Password) {
			return model.GameUsers{}, model.GameUserWallets{}, false, errH5PhoneAlreadyRegistered
		}
		wallet := h5EnsureWallet(db, existing.ID)
		settings := h5FindSettings(db, existing.ID)
		if settings.ID == 0 {
			_ = db.Create(&settings).Error
		}
		return existing, wallet, false, nil
	} else if !errors.Is(findErr, gorm.ErrRecordNotFound) {
		return model.GameUsers{}, model.GameUserWallets{}, false, findErr
	}

	user := model.GameUsers{

		SiteID:              scope.SiteID,
		UserNo:              h5UserNo(),
		PhoneCountryCode:    phoneCountryCode,
		Phone:               phone,
		PasswordHash:        passwordHash,
		Nickname:            h5MaskedPhone(phone),
		AvatarURL:           "",
		InviteCode:          strings.ToUpper(h5Token("I"))[:10],
		Status:              1,
		CountryCode:         defaults.CountryCode,
		Language:            defaults.Language,
		RegistrationChannel: "h5",
		RegisteredIP:        c.ClientIP(),
		LoginEnabled:        1,
		GameAccessEnabled:   1,
		DepositEnabled:      1,
		WithdrawalEnabled:   1,
	}
	if req.InviteCode != "" {
		var referrer model.GameUsers
		if content.ApplySite(db.Model(&model.GameUsers{}), "", scope).Where("invite_code = ? AND deleted_at IS NULL", strings.TrimSpace(req.InviteCode)).First(&referrer).Error == nil {
			user.ReferrerUserID = referrer.ID
		}
	}
	if err := db.Create(&user).Error; err != nil {
		if h5GameUserPhoneDuplicate(err) {
			return model.GameUsers{}, model.GameUserWallets{}, false, errH5PhoneAlreadyRegistered
		}
		return model.GameUsers{}, model.GameUserWallets{}, false, err
	}
	wallet := h5EnsureWallet(db, user.ID)
	settings := h5FindSettings(db, user.ID)
	if settings.ID == 0 {
		_ = db.Create(&settings).Error
	}
	return user, wallet, true, nil
}

func h5FindGameUserByPhone(db *gorm.DB, scope content.Scope, phoneCountryCode string, phone string) (model.GameUsers, error) {
	var user model.GameUsers
	err := content.ApplySiteStrict(db.Model(&model.GameUsers{}), "", scope).
		Where("phone_country_code = ? AND phone = ? AND deleted_at IS NULL", phoneCountryCode, phone).
		First(&user).Error
	return user, err
}

func h5AuthPhoneCountryCode(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "+92", nil
	}
	if !strings.HasPrefix(value, "+") || len(value) < 2 || len(value) > 5 {
		return "", errors.New("invalid phone country code")
	}
	for _, char := range value[1:] {
		if char < '0' || char > '9' {
			return "", errors.New("invalid phone country code")
		}
	}
	return value, nil
}

func h5GameUserPhoneDuplicate(err error) bool {
	var mysqlErr *mysql.MySQLError
	if !errors.As(err, &mysqlErr) || mysqlErr.Number != 1062 {
		return false
	}
	message := strings.ToLower(mysqlErr.Message)
	return strings.Contains(message, "uk_game_users_scope_phone") || strings.Contains(message, "uk_game_users_phone")
}

func h5RegistrationValidationError(err error) bool {
	if err == nil {
		return false
	}
	switch err.Error() {
	case "phone is required", "password is required", "invalid phone country code", "invalid phone number":
		return true
	default:
		return false
	}
}

func h5AuthPhone(value string) (string, error) {
	phone := h5CleanPhone(value)
	if phone == "" {
		return "", errors.New("phone is required")
	}
	if len(phone) < 6 || len(phone) > 15 {
		return "", errors.New("invalid phone number")
	}
	for _, char := range phone {
		if char < '0' || char > '9' {
			return "", errors.New("invalid phone number")
		}
	}
	return phone, nil
}

func h5PasswordHash(password string) (string, error) {
	password = strings.TrimSpace(password)
	if password == "" {
		return "", errors.New("password is required")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func h5VerifyGameUserPassword(db *gorm.DB, user model.GameUsers, password string) bool {
	password = strings.TrimSpace(password)
	stored := strings.TrimSpace(user.PasswordHash)
	if password == "" || stored == "" {
		return false
	}
	if bcrypt.CompareHashAndPassword([]byte(stored), []byte(password)) == nil {
		return true
	}
	if stored == password {
		if hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost); err == nil {
			_ = db.Model(&model.GameUsers{}).Where("id = ?", user.ID).Update("password_hash", string(hash)).Error
		}
		return true
	}
	return false
}

func h5BearerToken(c *gin.Context) string {
	value := strings.TrimSpace(c.GetHeader("Authorization"))
	if value == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(value), "bearer ") {
		return strings.TrimSpace(value[7:])
	}
	return ""
}

func h5UserSessionHash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func h5IssueGameUserSession(db *gorm.DB, c *gin.Context, userID uint64) (string, error) {
	token := "user_" + h5RandomHex(32)
	expiredAt := time.Now().Add(30 * 24 * time.Hour)
	scope := h5UserContentScope(db, userID, h5ContentScope(c))
	session := model.GameUserSessions{

		SiteID:    scope.SiteID,
		UserID:    userID,
		TokenHash: h5UserSessionHash(token),
		DeviceID:  h5Truncate(strings.TrimSpace(h5FirstNonEmpty(c.GetHeader("X-Device-ID"), c.GetHeader("X-Device-Id"))), 128),
		ClientIP:  h5Truncate(c.ClientIP(), 64),
		UserAgent: h5Truncate(c.GetHeader("User-Agent"), 500),
		ExpiredAt: &expiredAt,
	}
	return token, db.Create(&session).Error
}

func h5RevokeGameUserSession(db *gorm.DB, c *gin.Context) {
	token := h5BearerToken(c)
	if token == "" {
		return
	}
	now := time.Now()
	_ = content.ApplySite(db.Model(&model.GameUserSessions{}), "", h5ContentScope(c)).
		Where("token_hash = ? AND revoked_at IS NULL", h5UserSessionHash(token)).
		Updates(map[string]interface{}{"revoked_at": &now, "updated_at": now}).Error
}

func h5Truncate(value string, max int) string {
	if max <= 0 || len(value) <= max {
		return value
	}
	return value[:max]
}

func ensurePendingRegisterReward(user model.GameUsers, wallet model.GameUserWallets) (*model.GameRegisterRewardClaims, error) {
	db := database.GetDB()
	scope := h5UserContentScope(db, user.ID, content.Scope{})
	amount, enabled := registerRewardAmount(db, scope)
	if !enabled || !amount.GreaterThan(decimal.Zero) {
		return nil, nil
	}

	idempotencyKey := fmt.Sprintf("register_reward:%d", user.ID)
	var existing model.GameRegisterRewardClaims
	if err := content.ApplySite(db.Model(&model.GameRegisterRewardClaims{}), "", scope).Where("idempotency_key = ? AND deleted_at IS NULL", idempotencyKey).Take(&existing).Error; err == nil {
		return &existing, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	before := h5DecimalValue(wallet.CashBalance)
	claim := model.GameRegisterRewardClaims{

		SiteID:         scope.SiteID,
		UserID:         user.ID,
		Amount:         h5DecimalPtr(amount),
		Currency:       h5FirstNonEmpty(wallet.Currency, h5ScopeDefaults(db, scope).CurrencyCode),
		WalletType:     "CASH",
		Status:         1,
		CashBefore:     h5DecimalPtr(before),
		CashAfter:      h5DecimalPtr(before.Add(amount)),
		IdempotencyKey: idempotencyKey,
		RawMetadata:    "{}",
	}
	if err := db.Create(&claim).Error; err != nil {
		return nil, err
	}
	return &claim, nil
}

func pendingRegisterRewardPayload(db *gorm.DB, user model.GameUsers, wallet model.GameUserWallets) (gin.H, error) {
	if user.ID == 0 {
		return nil, nil
	}

	var claim model.GameRegisterRewardClaims
	err := content.ApplySite(db.Model(&model.GameRegisterRewardClaims{}), "", h5UserContentScope(db, user.ID, content.Scope{})).Where("user_id = ? AND status = 1 AND deleted_at IS NULL", user.ID).
		Order("id ASC").
		Take(&claim).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	before := h5DecimalValue(wallet.CashBalance)
	amount := h5DecimalValue(claim.Amount)
	return registerRewardPayload(claim, before, before.Add(amount), 0), nil
}

type pendingActivityRewardRow struct {
	ID           uint64
	ActivityID   uint64
	ActivityCode string
	Slug         string
	Title        string
	IconURL      string
	RewardCash   *decimal.Decimal
	RewardBonus  *decimal.Decimal
	CreatedAt    time.Time
}

func pendingActivityRewardsPayload(db *gorm.DB, user model.GameUsers, wallet model.GameUserWallets) ([]gin.H, decimal.Decimal, error) {
	items := make([]gin.H, 0)
	totalAmount := decimal.Zero
	if user.ID == 0 {
		return items, totalAmount, nil
	}

	var rows []pendingActivityRewardRow
	err := db.Table("game_user_activity_records AS reward").
		Select(`reward.id, reward.activity_id, activity.activity_code, activity.slug, activity.title,
			COALESCE(NULLIF(activity.card_image_url, ''), activity.hero_url) AS icon_url,
			reward.reward_cash, reward.reward_bonus, reward.created_at`).
		Joins("LEFT JOIN game_activity_events AS activity ON activity.id = reward.activity_id AND activity.deleted_at IS NULL").
		Where("reward.user_id = ? AND reward.status = 1 AND reward.deleted_at IS NULL", user.ID).
		Order("reward.created_at ASC, reward.id ASC").
		Find(&rows).Error
	if err != nil {
		return nil, totalAmount, err
	}

	scope := h5UserContentScope(db, user.ID, content.Scope{})
	currency := h5FirstNonEmpty(wallet.Currency, h5ScopeDefaults(db, scope).CurrencyCode)
	for _, row := range rows {
		cashAmount := h5DecimalValue(row.RewardCash)
		bonusAmount := h5DecimalValue(row.RewardBonus)
		amount := cashAmount.Add(bonusAmount)
		if !amount.GreaterThan(decimal.Zero) {
			continue
		}
		rewardType := "cash"
		if cashAmount.IsZero() && bonusAmount.GreaterThan(decimal.Zero) {
			rewardType = "bonus"
		} else if cashAmount.GreaterThan(decimal.Zero) && bonusAmount.GreaterThan(decimal.Zero) {
			rewardType = "mixed"
		}
		totalAmount = totalAmount.Add(amount)
		items = append(items, gin.H{
			"id":           row.ID,
			"activityID":   row.ActivityID,
			"activityCode": row.ActivityCode,
			"slug":         row.Slug,
			"title":        h5FirstNonEmpty(row.Title, "Bonus"),
			"iconURL":      h5AssetURL(row.IconURL),
			"rewardType":   rewardType,
			"cashAmount":   cashAmount.StringFixed(2),
			"bonusAmount":  bonusAmount.StringFixed(2),
			"amount":       amount.StringFixed(2),
			"currency":     currency,
			"status":       1,
			"createdAt":    row.CreatedAt,
		})
	}
	return items, totalAmount, nil
}

func claimPendingActivityReward(user model.GameUsers, recordID uint64) (model.GameUserWallets, gin.H, error) {
	db := database.GetDB()
	var updatedWallet model.GameUserWallets
	var reward gin.H
	err := db.Transaction(func(tx *gorm.DB) error {
		var record model.GameUserActivityRecords
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ? AND status = 1 AND deleted_at IS NULL", recordID, user.ID).
			Take(&record).Error; err != nil {
			return err
		}

		var wallet model.GameUserWallets
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND deleted_at IS NULL", user.ID).
			Order("id ASC").
			Take(&wallet).Error; err != nil {
			return err
		}

		cashAmount := h5DecimalValue(record.RewardCash)
		bonusAmount := h5DecimalValue(record.RewardBonus)
		if !cashAmount.Add(bonusAmount).GreaterThan(decimal.Zero) {
			return gorm.ErrRecordNotFound
		}

		cashBefore := h5DecimalValue(wallet.CashBalance)
		cashAfter := cashBefore.Add(cashAmount)
		bonusBefore := h5DecimalValue(wallet.BonusBalance)
		bonusAfter := bonusBefore.Add(bonusAmount)
		now := time.Now()
		if err := tx.Model(&model.GameUserWallets{}).
			Where("id = ?", wallet.ID).
			Updates(map[string]any{
				"cash_balance":        cashAfter,
				"bonus_balance":       bonusAfter,
				"last_transaction_at": now,
			}).Error; err != nil {
			return err
		}

		transactionIDs := make([]uint64, 0, 2)
		createLedger := func(walletType string, amount, before, after decimal.Decimal) error {
			if !amount.GreaterThan(decimal.Zero) {
				return nil
			}
			key := fmt.Sprintf("activity_reward:%d:%s", record.ID, strings.ToLower(walletType))
			ledger := model.GameWalletTransactions{
				TxnNo:               h5OrderNo("ACT"),
				UserID:              user.ID,
				WalletID:            wallet.ID,
				WalletType:          walletType,
				Direction:           "IN",
				Category:            "ACTIVITY_REWARD",
				SourceType:          "activity_reward",
				SourceID:            record.ID,
				IdempotencyKey:      &key,
				OrderNo:             h5FirstNonEmpty(record.OrderNo, h5OrderNo("AR")),
				Amount:              h5DecimalPtr(amount),
				Currency:            h5FirstNonEmpty(wallet.Currency, h5ScopeDefaults(tx, h5UserContentScope(tx, user.ID, content.Scope{})).CurrencyCode),
				Status:              1,
				BalanceBefore:       h5DecimalPtr(before),
				BalanceAfter:        h5DecimalPtr(after),
				FrozenBalanceBefore: h5DecimalPtr(h5DecimalValue(wallet.FrozenBalance)),
				FrozenBalanceAfter:  h5DecimalPtr(h5DecimalValue(wallet.FrozenBalance)),
				Remark:              h5FirstNonEmpty(record.Remark, "activity reward"),
				CompletedAt:         &now,
				RawMetadata:         "{}",
			}
			if err := tx.Create(&ledger).Error; err != nil {
				return err
			}
			transactionIDs = append(transactionIDs, ledger.ID)
			return nil
		}

		if err := createLedger("CASH", cashAmount, cashBefore, cashAfter); err != nil {
			return err
		}
		if err := createLedger("BONUS", bonusAmount, bonusBefore, bonusAfter); err != nil {
			return err
		}
		if err := tx.Model(&model.GameUserActivityRecords{}).
			Where("id = ? AND status = 1", record.ID).
			Updates(map[string]any{"status": 2, "claimed_at": now}).Error; err != nil {
			return err
		}

		wallet.CashBalance = h5DecimalPtr(cashAfter)
		wallet.BonusBalance = h5DecimalPtr(bonusAfter)
		wallet.LastTransactionAt = &now
		updatedWallet = wallet
		reward = gin.H{
			"id":             record.ID,
			"cashAmount":     cashAmount.StringFixed(2),
			"bonusAmount":    bonusAmount.StringFixed(2),
			"amount":         cashAmount.Add(bonusAmount).StringFixed(2),
			"currency":       h5FirstNonEmpty(wallet.Currency, h5ScopeDefaults(tx, h5UserContentScope(tx, user.ID, content.Scope{})).CurrencyCode),
			"status":         2,
			"transactionIDs": transactionIDs,
		}
		return nil
	})
	if err != nil {
		return model.GameUserWallets{}, nil, err
	}
	return updatedWallet, reward, nil
}

func claimPendingRegisterReward(user model.GameUsers) (model.GameUserWallets, gin.H, error) {
	db := database.GetDB()
	var updatedWallet model.GameUserWallets
	var reward gin.H
	err := db.Transaction(func(tx *gorm.DB) error {
		var claim model.GameRegisterRewardClaims
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND status = 1 AND deleted_at IS NULL", user.ID).
			Order("id ASC").
			Take(&claim).Error; err != nil {
			return err
		}

		var wallet model.GameUserWallets
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND deleted_at IS NULL", user.ID).
			Order("id ASC").
			Take(&wallet).Error; err != nil {
			return err
		}

		amount := h5DecimalValue(claim.Amount)
		before := h5DecimalValue(wallet.CashBalance)
		after := before.Add(amount)
		now := time.Now()
		if err := tx.Model(&model.GameUserWallets{}).
			Where("id = ?", wallet.ID).
			Updates(map[string]any{
				"cash_balance":        after,
				"last_transaction_at": now,
			}).Error; err != nil {
			return err
		}

		key := claim.IdempotencyKey
		ledger := model.GameWalletTransactions{
			TxnNo:               h5OrderNo("REG"),
			UserID:              user.ID,
			WalletID:            wallet.ID,
			WalletType:          "CASH",
			Direction:           "IN",
			Category:            "REGISTER_REWARD",
			SourceType:          "register_reward",
			SourceID:            claim.ID,
			IdempotencyKey:      &key,
			OrderNo:             h5OrderNo("RG"),
			Amount:              h5DecimalPtr(amount),
			Currency:            h5FirstNonEmpty(wallet.Currency, h5ScopeDefaults(tx, h5UserContentScope(tx, user.ID, content.Scope{})).CurrencyCode),
			Status:              1,
			BalanceBefore:       h5DecimalPtr(before),
			BalanceAfter:        h5DecimalPtr(after),
			FrozenBalanceBefore: h5DecimalPtr(h5DecimalValue(wallet.FrozenBalance)),
			FrozenBalanceAfter:  h5DecimalPtr(h5DecimalValue(wallet.FrozenBalance)),
			Remark:              "register reward",
			CompletedAt:         &now,
			RawMetadata:         "{}",
		}
		if err := tx.Create(&ledger).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.GameRegisterRewardClaims{}).
			Where("id = ?", claim.ID).
			Updates(map[string]any{
				"status":                2,
				"cash_before":           before,
				"cash_after":            after,
				"wallet_transaction_id": ledger.ID,
				"claimed_at":            now,
			}).Error; err != nil {
			return err
		}

		wallet.CashBalance = h5DecimalPtr(after)
		wallet.LastTransactionAt = &now
		updatedWallet = wallet
		claim.Status = 2
		claim.CashBefore = h5DecimalPtr(before)
		claim.CashAfter = h5DecimalPtr(after)
		claim.WalletTransactionID = ledger.ID
		claim.ClaimedAt = &now
		reward = registerRewardPayload(claim, before, after, ledger.ID)
		return nil
	})
	if err != nil {
		return model.GameUserWallets{}, nil, err
	}
	return updatedWallet, reward, nil
}

func registerRewardPayload(claim model.GameRegisterRewardClaims, cashBefore decimal.Decimal, cashAfter decimal.Decimal, transactionID uint64) gin.H {
	return gin.H{
		"id":            claim.ID,
		"amount":        h5DecimalValue(claim.Amount).StringFixed(2),
		"currency":      h5FirstNonEmpty(claim.Currency, "PKR"),
		"walletType":    strings.ToLower(h5FirstNonEmpty(claim.WalletType, "CASH")),
		"cashBefore":    cashBefore.StringFixed(2),
		"cashAfter":     cashAfter.StringFixed(2),
		"transactionID": transactionID,
		"status":        claim.Status,
	}
}

func registerRewardAmount(db *gorm.DB, scope content.Scope) (decimal.Decimal, bool) {
	if !systemConfigBool(db, "operation.register_reward_enabled", true, scope) {
		return decimal.Zero, false
	}

	minAmount := systemConfigDecimal(db, "operation.register_reward_min_amount", decimal.NewFromInt(1), scope)
	maxAmount := systemConfigDecimal(db, "operation.register_reward_max_amount", decimal.NewFromInt(20), scope)
	if minAmount.LessThan(decimal.Zero) {
		minAmount = decimal.Zero
	}
	if maxAmount.LessThan(minAmount) {
		maxAmount = minAmount
	}

	minCents := minAmount.Mul(decimal.NewFromInt(100)).IntPart()
	maxCents := maxAmount.Mul(decimal.NewFromInt(100)).IntPart()
	if maxCents <= minCents {
		return decimal.NewFromInt(minCents).Div(decimal.NewFromInt(100)), true
	}

	n, err := crand.Int(crand.Reader, big.NewInt(maxCents-minCents+1))
	if err != nil {
		return decimal.NewFromInt(minCents).Div(decimal.NewFromInt(100)), true
	}
	return decimal.NewFromInt(minCents + n.Int64()).Div(decimal.NewFromInt(100)), true
}

func systemConfigBool(db *gorm.DB, key string, fallback bool, scope content.Scope) bool {
	value := strings.TrimSpace(systemConfigValue(db, key, scope))
	if value == "" {
		return fallback
	}
	return strings.EqualFold(value, "true") || value == "1" || strings.EqualFold(value, "yes")
}

func systemConfigDecimal(db *gorm.DB, key string, fallback decimal.Decimal, scope content.Scope) decimal.Decimal {
	value := strings.TrimSpace(systemConfigValue(db, key, scope))
	if value == "" {
		return fallback
	}
	parsed, err := decimal.NewFromString(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func systemConfigValue(db *gorm.DB, key string, scope content.Scope) string {
	var value string
	_ = content.ApplySite(db.Table("game_system_configs"), "", scope).
		Select("config_value").
		Where("config_key = ? AND status = 1 AND deleted_at IS NULL", key).
		Scan(&value).Error
	return value
}

func recordGameUserLogin(c *gin.Context, userID uint64, identifier string, success bool, failureReason string) {
	now := time.Now()
	db := database.GetDB()
	scope := h5UserContentScope(db, userID, h5ContentScope(c))
	record := model.GameUserLoginRecords{

		SiteID:          scope.SiteID,
		UserID:          userID,
		LoginIdentifier: h5MaskedPhone(h5CleanPhone(identifier)),
		Success:         boolInt(success),
		FailureReason:   failureReason,
		IPAddress:       c.ClientIP(),
		Country:         strings.TrimSpace(c.GetHeader("X-Country-Code")),
		Region:          strings.TrimSpace(c.GetHeader("X-Region")),
		DeviceID:        strings.TrimSpace(c.GetHeader("X-Device-ID")),
		DeviceType:      strings.TrimSpace(c.GetHeader("X-Device-Type")),
		Platform:        strings.TrimSpace(c.GetHeader("X-Platform")),
		OS:              strings.TrimSpace(c.GetHeader("X-OS")),
		Browser:         strings.TrimSpace(c.GetHeader("X-Browser")),
		AppVersion:      strings.TrimSpace(c.GetHeader("X-App-Version")),
		UserAgent:       c.Request.UserAgent(),
		OccurredAt:      &now,
	}
	_ = db.Create(&record).Error
}

func upsertGameUserDevice(c *gin.Context, userID uint64, now time.Time) {
	deviceID := strings.TrimSpace(c.GetHeader("X-Device-ID"))
	if deviceID == "" {
		return
	}
	values := map[string]any{
		"platform":         strings.TrimSpace(c.GetHeader("X-Platform")),
		"device_model":     strings.TrimSpace(c.GetHeader("X-Device-Model")),
		"os":               strings.TrimSpace(c.GetHeader("X-OS")),
		"browser":          strings.TrimSpace(c.GetHeader("X-Browser")),
		"app_version":      strings.TrimSpace(c.GetHeader("X-App-Version")),
		"fingerprint_hash": strings.TrimSpace(c.GetHeader("X-Fingerprint-Hash")),
		"last_ip":          c.ClientIP(),
		"client_ip":        c.ClientIP(),
		"user_agent":       c.Request.UserAgent(),
		"last_seen_at":     now,
		"login_count":      gorm.Expr("login_count + 1"),
	}
	device := model.GameUserDevices{}
	db := database.GetDB()
	scope := h5UserContentScope(db, userID, h5ContentScope(c))
	err := content.ApplySite(db.Model(&model.GameUserDevices{}), "", scope).Where("user_id = ? AND device_id = ? AND deleted_at IS NULL", userID, deviceID).Take(&device).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		device = model.GameUserDevices{
			SiteID: scope.SiteID,
			UserID: userID, DeviceID: deviceID, Platform: firstHeader(c, "X-Platform", "h5"),
			DeviceModel: c.GetHeader("X-Device-Model"), OS: c.GetHeader("X-OS"), Browser: c.GetHeader("X-Browser"),
			AppVersion: c.GetHeader("X-App-Version"), FingerprintHash: c.GetHeader("X-Fingerprint-Hash"),
			FirstIP: c.ClientIP(), LastIP: c.ClientIP(), ClientIP: c.ClientIP(), UserAgent: c.Request.UserAgent(),
			FirstSeenAt: &now, LastSeenAt: &now, Status: 1, TrustStatus: 1, RiskStatus: 1, LoginCount: 1,
		}
		_ = db.Create(&device).Error
		return
	}
	if err == nil {
		_ = db.Model(&device).Updates(values).Error
	}
}

func blockedGameUserDevice(c *gin.Context, userID uint64) bool {
	deviceID := strings.TrimSpace(c.GetHeader("X-Device-ID"))
	if deviceID == "" {
		return false
	}
	var count int64
	db := database.GetDB()
	_ = content.ApplySite(db.Model(&model.GameUserDevices{}), "", h5UserContentScope(db, userID, h5ContentScope(c))).
		Where("user_id = ? AND device_id = ? AND status = 2 AND deleted_at IS NULL", userID, deviceID).
		Count(&count).Error
	return count > 0
}

func firstHeader(c *gin.Context, key string, fallback string) string {
	value := strings.TrimSpace(c.GetHeader(key))
	if value == "" {
		return fallback
	}
	return value
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
