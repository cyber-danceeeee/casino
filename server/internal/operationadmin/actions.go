package operationadmin

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/content"
	"casino/internal/model"
	"casino/internal/notificationaudience"
	"casino/internal/notificationchannel"
	"casino/internal/operationdefaults"
)

type bonusRow struct {
	ID                 uint64           `json:"id"`
	BonusID            string           `json:"bonusID"`
	UserID             uint64           `json:"userID"`
	UserNo             string           `gorm:"-" json:"userNo"`
	Username           string           `gorm:"-" json:"username"`
	PromotionID        uint64           `json:"promotionID"`
	PromotionName      string           `gorm:"-" json:"promotion"`
	BonusType          string           `json:"bonusType"`
	Amount             *decimal.Decimal `json:"bonusAmount"`
	RequiredValidBet   *decimal.Decimal `json:"requiredWagering"`
	CurrentWagering    *decimal.Decimal `json:"currentWagering"`
	RemainingWagering  *decimal.Decimal `json:"remainingWagering"`
	WithdrawableAmount *decimal.Decimal `json:"withdrawableAmount"`
	Currency           string           `json:"currency"`
	StatusCode         string           `json:"status"`
	IssuedAt           *time.Time       `json:"issuedAt"`
	ExpiresAt          *time.Time       `json:"expiresAt"`
	CreatedAt          time.Time        `json:"createdAt"`
	UpdatedAt          time.Time        `json:"updatedAt"`
	DeletedAt          *time.Time       `json:"-"`
}

func (bonusRow) TableName() string { return "game_bonus_records" }

func (s *Service) ListBonuses(q ListQuery) (PageResult, error) {
	var rows []bonusRow
	db := s.db.Table("game_bonus_records b").Select("b.id,b.bonus_id,b.user_id,u.user_no,u.username,b.promotion_id,COALESCE(p.name_en,p.title,'') promotion_name,b.bonus_type,b.amount,b.required_valid_bet,b.current_wagering,b.remaining_wagering,b.withdrawable_amount,b.currency,b.status_code,b.issued_at,b.expires_at,b.created_at,b.updated_at").Joins("LEFT JOIN game_users u ON u.id=b.user_id").Joins("LEFT JOIN game_activity_events p ON p.id=b.promotion_id").Where("b.deleted_at IS NULL")
	db = contains(db, "b.bonus_id", q.Values["bonusID"])
	db = equal(db, "b.user_id", q.Values["userID"])
	db = contains(db, "u.username", q.Values["username"])
	db = equal(db, "b.promotion_id", q.Values["promotionID"])
	db = equal(db, "b.bonus_type", q.Values["bonusType"])
	db = equal(db, "b.currency", q.Values["currency"])
	db = equal(db, "b.status_code", q.Values["status"])
	db = rangeFilter(db, "b.created_at", q.Values["createdFrom"], q.Values["createdTo"])
	db = rangeFilter(db, "b.expires_at", q.Values["expirationFrom"], q.Values["expirationTo"])
	return page(db.Order("b.id DESC"), q, &rows)
}
func (s *Service) GetBonus(id uint64) (map[string]any, error) {
	var row bonusRow
	if err := s.db.Where("id=? AND deleted_at IS NULL", id).First(&row).Error; err != nil {
		return nil, notFound(err, "BONUS_NOT_FOUND")
	}
	var user struct {
		UserNo   string `json:"userNo"`
		Username string `json:"username"`
	}
	s.db.Table("game_users").Select("user_no,username").Where("id=?", row.UserID).Scan(&user)
	row.UserNo = user.UserNo
	row.Username = user.Username
	var history []model.GameBonusStatusHistory
	s.db.Where("bonus_record_id=? AND deleted_at IS NULL", id).Order("id DESC").Find(&history)
	var bets []map[string]any
	s.db.Table("game_bet_records").Select("platform_bet_id,bet_amount,valid_bet_amount,created_at").Where("user_id=? AND created_at>=? AND deleted_at IS NULL", row.UserID, row.CreatedAt).Order("id DESC").Limit(50).Scan(&bets)
	var source promotionRecord
	if row.PromotionID > 0 {
		_ = s.db.First(&source, row.PromotionID).Error
	}
	logs, _ := s.operationLogs("bonus", id, content.Scope{})
	return map[string]any{"overview": row, "wageringProgress": map[string]any{"currentWagering": row.CurrentWagering, "requiredWagering": row.RequiredValidBet, "remainingWagering": row.RemainingWagering, "withdrawableAmount": row.WithdrawableAmount}, "relatedBets": bets, "rewardSource": source, "statusHistory": history, "operationLogs": logs}, nil
}
func (s *Service) TransitionBonus(id uint64, target string, permission string, reason string, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row bonusRow
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=? AND deleted_at IS NULL", id).First(&row).Error; err != nil {
			return notFound(err, "BONUS_NOT_FOUND")
		}
		if row.StatusCode == target {
			return nil
		}
		allowed := map[string]map[string]bool{"PENDING": {"ACTIVE": true, "CANCELLED": true, "FORFEITED": true}, "ACTIVE": {"COMPLETED": true, "CANCELLED": true, "FORFEITED": true}}
		if !allowed[row.StatusCode][target] {
			return newError("INVALID_BONUS_STATUS_TRANSITION", "invalid bonus status transition")
		}
		before := row
		if target == "ACTIVE" {
			if err := s.changeBonusWallet(tx, row, true, actor); err != nil {
				return err
			}
		} else if row.StatusCode == "ACTIVE" && (target == "CANCELLED" || target == "FORFEITED") {
			if err := s.changeBonusWallet(tx, row, false, actor); err != nil {
				return err
			}
		}
		if err := tx.Model(&bonusRow{}).Where("id=?", id).Updates(map[string]any{"status_code": target, "status": bonusLegacyStatus(target)}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.GameBonusStatusHistory{BonusRecordID: id, BeforeStatus: row.StatusCode, AfterStatus: target, OperatorID: actor.ID, OperatorUsername: actor.Username, Reason: reason}).Error; err != nil {
			return err
		}
		row.StatusCode = target
		return audit(tx, actor, permission, "status", "bonus", id, before, row, reason)
	})
}
func bonusLegacyStatus(v string) int {
	switch v {
	case "PENDING", "ACTIVE":
		return 1
	case "COMPLETED":
		return 2
	case "EXPIRED":
		return 3
	default:
		return 4
	}
}
func (s *Service) changeBonusWallet(tx *gorm.DB, row bonusRow, credit bool, actor Actor) error {
	var wallet model.GameUserWallets
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id=? AND deleted_at IS NULL", row.UserID).First(&wallet).Error; err != nil {
		return notFound(err, "WALLET_NOT_FOUND")
	}
	amount := decimal.Zero
	if row.Amount != nil {
		amount = *row.Amount
	}
	before := decimal.Zero
	if wallet.BonusBalance != nil {
		before = *wallet.BonusBalance
	}
	after := before.Add(amount)
	direction := "credit"
	action := "activate"
	if !credit {
		direction = "debit"
		action = "reverse"
		if before.LessThan(amount) {
			return newError("BONUS_BALANCE_INSUFFICIENT", "bonus balance is insufficient for exact reversal")
		}
		after = before.Sub(amount)
	}
	key := fmt.Sprintf("bonus:%d:%s", row.ID, action)
	var existing int64
	if err := tx.Model(&model.GameWalletTransactions{}).Where("idempotency_key=?", key).Count(&existing).Error; err != nil {
		return err
	}
	if existing > 0 {
		return nil
	}
	now := time.Now()
	if err := tx.Model(&wallet).Updates(map[string]any{"bonus_balance": after, "last_transaction_at": now}).Error; err != nil {
		return err
	}
	txn := model.GameWalletTransactions{TxnNo: generatedCode("BONUS"), UserID: row.UserID, WalletID: wallet.ID, WalletType: "bonus", Direction: direction, Category: "bonus_" + action, SourceType: "bonus", SourceID: row.ID, IdempotencyKey: &key, OrderNo: row.BonusID, Amount: &amount, Currency: row.Currency, Status: 1, BalanceBefore: &before, BalanceAfter: &after, FrozenBalanceBefore: wallet.FrozenBalance, FrozenBalanceAfter: wallet.FrozenBalance, Remark: "bonus " + action, OperatorID: actor.ID, OperatorUsername: actor.Username, CompletedAt: &now, RawMetadata: "{}"}
	return tx.Create(&txn).Error
}

func (s *Service) ListVIPLevels(q ListQuery) (PageResult, error) {
	var rows []vipRecord
	db := content.ApplySite(s.db.Model(&vipRecord{}), "", q.Scope).Where("deleted_at IS NULL")
	db = equal(db, "level_no", q.Values["level"])
	db = contains(db, "code", q.Values["code"])
	if !q.Scope.SiteScoped() {
		db = boolFilter(db, "enabled", q.Values["enabled"])
	}
	if !q.Scope.SiteScoped() {
		result, err := page(db.Order("sort_order,level_no"), q, &rows)
		if err != nil {
			return result, err
		}
		for i := range rows {
			content.ApplySite(s.db.Table("game_users"), "", q.Scope).Where("vip_level=? AND deleted_at IS NULL", rows[i].LevelNo).Count(&rows[i].UserCount)
			signVIPAssets(&rows[i])
			if err := loadVIPRetention(s.db, &rows[i]); err != nil {
				return result, err
			}
		}
		result.List = rows
		return result, nil
	}
	db = boolFilter(db, "enabled", q.Values["enabled"])
	result, err := page(db.Order("sort_order,level_no"), q, &rows)
	if err != nil {
		return result, err
	}
	for i := range rows {
		rows[i].Metadata = operationdefaults.Metadata{CanEdit: true, CanDelete: true, Visible: rows[i].Enabled}
		content.ApplySite(s.db.Table("game_users"), "", q.Scope).Where("vip_level=? AND deleted_at IS NULL", rows[i].LevelNo).Count(&rows[i].UserCount)
		signVIPAssets(&rows[i])
		if err := loadVIPRetention(s.db, &rows[i]); err != nil {
			return result, err
		}
	}
	result.List = rows
	return result, nil
}

func (s *Service) ListVIPUsers(q ListQuery) (PageResult, error) {
	var rows []struct {
		ID        uint64    `json:"id"`
		UserNo    string    `json:"userNo"`
		Username  string    `json:"username"`
		VIPLevel  uint      `json:"vipLevel"`
		LevelName string    `json:"levelName"`
		UpdatedAt time.Time `json:"updatedAt"`
	}
	vipJoin := "LEFT JOIN game_vip_levels v ON v.level_no=u.vip_level AND v.deleted_at IS NULL"
	vipJoinArgs := []any{}
	if q.Scope.SiteScoped() {
		vipJoin += " AND v.site_id IN (0, ?)"
		vipJoinArgs = append(vipJoinArgs, q.Scope.SiteID)
	}
	db := content.ApplySite(s.db.Table("game_users u"), "u", q.Scope).
		Select("u.id,u.user_no,u.username,u.vip_level,COALESCE(v.name_en,v.name,'') level_name,u.updated_at").
		Joins(vipJoin, vipJoinArgs...).
		Where("u.deleted_at IS NULL")
	db = equal(db, "u.id", q.Values["userID"])
	db = contains(db, "u.user_no", q.Values["playerNo"])
	db = contains(db, "u.username", q.Values["username"])
	db = equal(db, "u.vip_level", q.Values["level"])
	return page(db.Order("u.id DESC"), q, &rows)
}
func (s *Service) GetVIPLevel(id uint64, scope content.Scope) (map[string]any, error) {
	var row vipRecord
	if err := content.ApplySite(s.db.Model(&vipRecord{}), "", scope).Where("id=? AND deleted_at IS NULL", id).First(&row).Error; err != nil {
		return nil, notFound(err, "VIP_LEVEL_NOT_FOUND")
	}
	row.Metadata = operationdefaults.Metadata{CanEdit: true, CanDelete: true, Visible: row.Enabled}
	if err := loadVIPRetention(s.db, &row); err != nil {
		return nil, err
	}
	signVIPAssets(&row)
	var benefits []model.GameVIPBenefit
	var users []map[string]any
	var adjustments []model.GameVIPAdjustmentLog
	content.ApplySite(s.db.Model(&model.GameVIPBenefit{}), "", scope).Where("vip_level_id=? AND deleted_at IS NULL", id).Order("sort_order,id").Find(&benefits)
	content.ApplySite(s.db.Table("game_users"), "", scope).Select("id,user_no,username,vip_level,updated_at").Where("vip_level=? AND deleted_at IS NULL", row.LevelNo).Limit(100).Scan(&users)
	content.ApplySite(s.db.Model(&model.GameVIPAdjustmentLog{}), "", scope).Where("before_level=? OR after_level=?", row.LevelNo, row.LevelNo).Order("id DESC").Limit(100).Find(&adjustments)
	logs, _ := s.operationLogs("vip_level", id, scope)
	return map[string]any{"overview": row, "benefits": benefits, "users": users, "adjustmentLogs": adjustments, "operationLogs": logs}, nil
}

func (s *Service) SaveVIPLevel(id uint64, in VIPLevelInput, actor Actor) (uint64, error) {
	requiredDeposit, err := parseDecimal(in.RequiredDeposit, "requiredDeposit")
	if err != nil {
		return 0, err
	}
	requiredBet, err := parseDecimal(in.RequiredValidBet, "requiredValidBet")
	if err != nil {
		return 0, err
	}
	withdrawal, err := parseDecimal(in.WithdrawalLimit, "withdrawalLimit")
	if err != nil {
		return 0, err
	}
	cashback, err := parseDecimal(in.CashbackRate, "cashbackRate")
	if err != nil {
		return 0, err
	}
	birthday, err := parseDecimal(in.BirthdayReward, "birthdayReward")
	if err != nil {
		return 0, err
	}
	monthly, err := parseDecimal(in.MonthlyReward, "monthlyReward")
	if err != nil {
		return 0, err
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var before vipRecord
		action := "create"
		if id > 0 {
			if err := content.ApplySite(tx.Model(&vipRecord{}), "", actor.Scope).Where("id=? AND deleted_at IS NULL", id).First(&before).Error; err != nil {
				return notFound(err, "VIP_LEVEL_NOT_FOUND")
			}
			action = "update"
			if err := loadVIPRetention(tx, &before); err != nil {
				return err
			}
		}
		nameZh := strings.TrimSpace(in.NameZh)
		nameEn := strings.TrimSpace(in.NameEn)
		icon := strings.TrimSpace(in.Icon)
		if id > 0 {
			if nameZh == "" {
				nameZh = before.NameZh
			}
			if nameEn == "" {
				nameEn = before.NameEn
			}
			if icon == "" {
				icon = before.IconURL
			} else {
				icon = operationAssetRef(icon)
			}
		} else {
			defaultName := fmt.Sprintf("VIP %d", in.Level)
			if nameZh == "" {
				nameZh = defaultName
			}
			if nameEn == "" {
				nameEn = defaultName
			}
			icon = operationAssetRef(icon)
		}
		name := nameEn
		if name == "" {
			name = nameZh
		}
		row := vipRecord{ID: id, SiteID: actor.Scope.SiteID, LevelNo: in.Level, Code: strings.TrimSpace(in.Code), Name: name, NameZh: nameZh, NameEn: nameEn, IconURL: icon, RequiredExperience: in.RequiredExperience, RequiredDeposit: &requiredDeposit, RequiredValidBet: &requiredBet, WithdrawalLimit: &withdrawal, CashbackRate: &cashback, BirthdayReward: &birthday, MonthlyReward: &monthly, SortOrder: in.SortOrder, Status: operationStatus(in.Enabled), Enabled: in.Enabled}
		if id > 0 {
			row.SiteID = before.SiteID
		}
		if err := applyVIPRewards(&row, before, in); err != nil {
			return err
		}
		if id == 0 {
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			id = row.ID
		} else if err := content.ApplySite(tx.Model(&vipRecord{}), "", actor.Scope).Where("id=?", id).Select(
			"level_no", "code", "name", "name_zh", "name_en", "icon_url", "required_experience",
			"required_deposit", "required_valid_bet", "withdrawal_limit", "cashback_rate",
			"birthday_reward", "monthly_reward", "sort_order", "status", "enabled",
			"level_up_reward", "weekly_reward", "daily_withdraw_count",
		).Updates(row).Error; err != nil {
			return err
		}
		if err := saveVIPRequirements(tx, row, in); err != nil {
			return err
		}
		if err := loadVIPRetention(tx, &row); err != nil {
			return err
		}
		now := time.Now()
		if err := content.ApplySite(tx.Table("game_vip_benefits"), "", actor.Scope).Where("vip_level_id = ?", id).Updates(map[string]any{"deleted_at": now, "updated_at": now}).Error; err != nil {
			return err
		}
		for _, b := range in.Benefits {
			v, e := parseDecimal(b.BenefitValue, "benefitValue")
			if e != nil {
				return e
			}
			if err := tx.Create(&model.GameVIPBenefit{SiteID: actor.Scope.SiteID, VIPLevelID: id, BenefitCode: b.BenefitCode, NameZh: b.NameZh, NameEn: b.NameEn, BenefitValue: &v, ValueUnit: b.ValueUnit, Enabled: b.Enabled, SortOrder: b.SortOrder}).Error; err != nil {
				return err
			}
		}
		if err := content.ApplySite(tx.Model(&vipRecord{}), "", actor.Scope).Where("id=?", id).First(&row).Error; err != nil {
			return err
		}
		return audit(tx, actor, "operation:vip:update", action, "vip_level", id, before, row, "")
	})
	return id, err
}
func (s *Service) AdjustVIP(userID uint64, in VIPAdjustCommand, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var user model.GameUsers
		if err := content.ApplySite(tx.Model(&model.GameUsers{}), "", actor.Scope).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
			return notFound(err, "USER_NOT_FOUND")
		}
		var level vipRecord
		if err := content.ApplySite(tx.Model(&vipRecord{}), "", actor.Scope).Where("level_no=? AND enabled=1 AND deleted_at IS NULL", in.Level).First(&level).Error; err != nil {
			return notFound(err, "VIP_LEVEL_NOT_FOUND")
		}
		if user.VipLevel == in.Level {
			return nil
		}
		before := user.VipLevel
		if err := content.ApplySite(tx.Model(&model.GameUsers{}), "", actor.Scope).Where("id=?", userID).Update("vip_level", in.Level).Error; err != nil {
			return err
		}
		now := time.Now()
		zero := decimal.Zero
		if err := tx.Create(&model.GameUserVipLogs{SiteID: actor.Scope.SiteID, UserID: userID, OldLevelNo: before, NewLevelNo: in.Level, Reason: "manual_adjustment", DepositAmount: &zero, ValidBetAmount: &zero, RewardAmount: &zero, ChangedAt: &now}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.GameVIPAdjustmentLog{SiteID: actor.Scope.SiteID, UserID: userID, BeforeLevel: before, AfterLevel: in.Level, OperatorID: actor.ID, OperatorUsername: actor.Username, Reason: in.Reason}).Error; err != nil {
			return err
		}
		return audit(tx, actor, "operation:vip:adjust", "adjust", "vip_user", userID, map[string]any{"level": before}, map[string]any{"level": in.Level}, in.Reason)
	})
}

func (s *Service) AdjustVIPByNumber(playerNo string, in VIPAdjustCommand, actor Actor) error {
	var user struct {
		ID uint64 `gorm:"column:id"`
	}
	if err := content.ApplySite(s.db.Table("game_users"), "", actor.Scope).
		Select("id").
		Where("user_no=? AND deleted_at IS NULL", strings.TrimSpace(playerNo)).
		Take(&user).Error; err != nil {
		return notFound(err, "USER_NOT_FOUND")
	}
	return s.AdjustVIP(user.ID, in, actor)
}

func (s *Service) ListPromoCodes(q ListQuery) (PageResult, error) {
	var rows []model.GamePromoCode
	db := content.ApplySite(s.db.Model(&model.GamePromoCode{}), "", q.Scope).Where("deleted_at IS NULL")
	db = contains(db, "code", q.Values["code"])
	db = equal(db, "promotion_id", q.Values["promotionID"])
	db = equal(db, "reward_type", q.Values["rewardType"])
	db = boolFilter(db, "enabled", q.Values["enabled"])
	db = rangeFilter(db, "used_count", q.Values["usageMin"], q.Values["usageMax"])
	db = rangeFilter(db, "created_at", q.Values["createdFrom"], q.Values["createdTo"])
	db = rangeFilter(db, "expires_at", q.Values["expirationFrom"], q.Values["expirationTo"])
	return page(db.Order("id DESC"), q, &rows)
}
func (s *Service) GetPromoCode(id uint64, scope content.Scope) (map[string]any, error) {
	var row model.GamePromoCode
	if err := content.ApplySite(s.db.Model(&model.GamePromoCode{}), "", scope).Where("id=? AND deleted_at IS NULL", id).First(&row).Error; err != nil {
		return nil, notFound(err, "PROMO_CODE_NOT_FOUND")
	}
	var usage []model.GamePromoCodeRedemption
	content.ApplySite(s.db.Model(&model.GamePromoCodeRedemption{}), "", scope).Where("promo_code_id=? AND deleted_at IS NULL", id).Order("id DESC").Limit(200).Find(&usage)
	logs, _ := s.operationLogs("promo_code", id, scope)
	return map[string]any{"overview": row, "usageRecords": usage, "operationLogs": logs}, nil
}

func (s *Service) ListPromoCodeUsage(id uint64, q ListQuery) (PageResult, error) {
	var codeCount int64
	if err := content.ApplySite(s.db.Model(&model.GamePromoCode{}), "", q.Scope).Where("id=? AND deleted_at IS NULL", id).Count(&codeCount).Error; err != nil {
		return PageResult{}, err
	}
	if codeCount == 0 {
		return PageResult{}, newError("PROMO_CODE_NOT_FOUND", "record not found")
	}
	var rows []model.GamePromoCodeRedemption
	db := content.ApplySite(s.db.Model(&model.GamePromoCodeRedemption{}), "", q.Scope).Where("promo_code_id=? AND deleted_at IS NULL", id)
	return page(db.Order("id DESC"), q, &rows)
}
func (s *Service) SavePromoCode(id uint64, in PromoCodeInput, actor Actor) (uint64, error) {
	if err := validateTime(in.StartTime, in.ExpiresAt); err != nil {
		return 0, err
	}
	value, err := parseDecimal(in.RewardValue, "rewardValue")
	if err != nil {
		return 0, err
	}
	in.Code = strings.ToUpper(strings.TrimSpace(in.Code))
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var before model.GamePromoCode
		action := "create"
		permission := "operation:promo-code:create"
		used := uint(0)
		if id > 0 {
			if err := content.ApplySite(tx.Model(&model.GamePromoCode{}), "", actor.Scope).Where("id=? AND deleted_at IS NULL", id).First(&before).Error; err != nil {
				return notFound(err, "PROMO_CODE_NOT_FOUND")
			}
			action = "update"
			permission = "operation:promo-code:update"
			used = before.UsedCount
		}
		row := model.GamePromoCode{ID: id, SiteID: actor.Scope.SiteID, Code: in.Code, PromotionID: in.PromotionID, RewardType: in.RewardType, RewardValue: &value, Currency: in.Currency, TotalUsageLimit: in.TotalUsageLimit, UsedCount: used, PerUserLimit: in.PerUserLimit, Countries: in.Countries, NewUsersOnly: in.NewUsersOnly, Enabled: in.Enabled, StartAt: in.StartTime, ExpiresAt: in.ExpiresAt}
		if id == 0 {
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			id = row.ID
		} else if err := content.ApplySite(tx.Model(&model.GamePromoCode{}), "", actor.Scope).Where("id=?", id).Select(
			"code", "promotion_id", "reward_type", "reward_value", "currency", "total_usage_limit",
			"per_user_limit", "countries", "new_users_only", "enabled", "start_at", "expires_at",
		).Updates(row).Error; err != nil {
			return err
		}
		if err := content.ApplySite(tx.Model(&model.GamePromoCode{}), "", actor.Scope).Where("id=?", id).First(&row).Error; err != nil {
			return err
		}
		return audit(tx, actor, permission, action, "promo_code", id, before, row, "")
	})
	return id, err
}
func (s *Service) DeletePromoCode(id uint64, actor Actor) error {
	var count int64
	content.ApplySite(s.db.Model(&model.GamePromoCodeRedemption{}), "", actor.Scope).Where("promo_code_id=?", id).Count(&count)
	if count > 0 {
		return newError("PROMO_CODE_IN_USE", "promo code with usage records cannot be deleted")
	}
	return s.softDelete("game_promo_codes", "promo_code", id, "operation:promo-code:delete", actor)
}
func (s *Service) RedeemPromoCode(in RedeemPromoCodeCommand) (*model.GamePromoCodeRedemption, error) {
	if strings.TrimSpace(in.IdempotencyKey) == "" {
		return nil, newError("IDEMPOTENCY_KEY_REQUIRED", "idempotency key is required")
	}
	var result model.GamePromoCodeRedemption
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var user model.GameUsers
		if err := tx.Where("id=? AND deleted_at IS NULL", in.UserID).First(&user).Error; err != nil {
			return notFound(err, "USER_NOT_FOUND")
		}
		scope := content.Scope{SiteID: user.SiteID}
		if err := content.ApplySite(tx.Model(&model.GamePromoCodeRedemption{}), "", scope).Where("idempotency_key=?", in.IdempotencyKey).First(&result).Error; err == nil {
			return nil
		} else if err != gorm.ErrRecordNotFound {
			return err
		}
		var code model.GamePromoCode
		if err := content.ApplySite(tx.Model(&model.GamePromoCode{}), "", scope).Clauses(clause.Locking{Strength: "UPDATE"}).Where("code=? AND deleted_at IS NULL", strings.ToUpper(strings.TrimSpace(in.Code))).First(&code).Error; err != nil {
			return notFound(err, "PROMO_CODE_NOT_FOUND")
		}
		now := time.Now()
		if !code.Enabled || (code.StartAt != nil && now.Before(*code.StartAt)) || (code.ExpiresAt != nil && now.After(*code.ExpiresAt)) {
			return newError("PROMO_CODE_UNAVAILABLE", "promo code is unavailable")
		}
		if code.TotalUsageLimit > 0 && code.UsedCount >= code.TotalUsageLimit {
			return newError("PROMO_CODE_LIMIT_REACHED", "promo code usage limit reached")
		}
		if !containsCode(code.Countries, user.CountryCode) {
			return newError("PROMO_CODE_COUNTRY_NOT_ELIGIBLE", "promo code is not available in this country")
		}
		if code.NewUsersOnly {
			var completedDeposits int64
			if err := content.ApplySite(tx.Model(&model.GameDepositOrders{}), "", scope).Where("user_id=? AND completed_at IS NOT NULL AND deleted_at IS NULL", in.UserID).Count(&completedDeposits).Error; err != nil {
				return err
			}
			if completedDeposits > 0 {
				return newError("PROMO_CODE_NEW_USERS_ONLY", "promo code is only available to new users")
			}
		}
		var userCount int64
		if err := content.ApplySite(tx.Model(&model.GamePromoCodeRedemption{}), "", scope).Where("promo_code_id=? AND user_id=? AND status='COMPLETED'", code.ID, in.UserID).Count(&userCount).Error; err != nil {
			return err
		}
		if uint(userCount) >= code.PerUserLimit {
			return newError("PROMO_CODE_USER_LIMIT_REACHED", "promo code user limit reached")
		}
		var wallet model.GameUserWallets
		if err := content.ApplySite(tx.Model(&model.GameUserWallets{}), "", scope).Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id=? AND deleted_at IS NULL", in.UserID).First(&wallet).Error; err != nil {
			return notFound(err, "WALLET_NOT_FOUND")
		}
		if !strings.EqualFold(wallet.Currency, code.Currency) {
			return newError("PROMO_CODE_CURRENCY_MISMATCH", "promo code currency does not match wallet currency")
		}
		if code.RewardType != "CASH" && code.RewardType != "BONUS" {
			return newError("PROMO_CODE_REWARD_TYPE_UNSUPPORTED", "promo code reward type is not supported for wallet redemption")
		}
		amount := decimal.Zero
		if code.RewardValue != nil {
			amount = *code.RewardValue
		}
		before := decimal.Zero
		walletField := "bonus_balance"
		walletType := "bonus"
		if code.RewardType == "CASH" {
			walletField = "cash_balance"
			walletType = "cash"
			before = *wallet.CashBalance
		} else {
			before = *wallet.BonusBalance
		}
		after := before.Add(amount)
		if err := content.ApplySite(tx.Model(&model.GameUserWallets{}), "", scope).Where("id=?", wallet.ID).Updates(map[string]any{walletField: after, "last_transaction_at": now}).Error; err != nil {
			return err
		}
		key := "promo:" + in.IdempotencyKey
		txn := model.GameWalletTransactions{SiteID: scope.SiteID, TxnNo: generatedCode("PROMO"), UserID: in.UserID, WalletID: wallet.ID, WalletType: walletType, Direction: "credit", Category: "promo_code", SourceType: "promo_code", SourceID: code.ID, IdempotencyKey: &key, OrderNo: code.Code, Amount: &amount, Currency: code.Currency, Status: 1, BalanceBefore: &before, BalanceAfter: &after, FrozenBalanceBefore: wallet.FrozenBalance, FrozenBalanceAfter: wallet.FrozenBalance, Remark: "promo code redemption", CompletedAt: &now, RawMetadata: "{}"}
		if err := tx.Create(&txn).Error; err != nil {
			return err
		}
		result = model.GamePromoCodeRedemption{SiteID: scope.SiteID, PromoCodeID: code.ID, UserID: in.UserID, IdempotencyKey: in.IdempotencyKey, RewardType: code.RewardType, RewardValue: &amount, Currency: code.Currency, WalletTransactionID: txn.ID, Status: "COMPLETED"}
		if err := tx.Create(&result).Error; err != nil {
			return err
		}
		return content.ApplySite(tx.Model(&model.GamePromoCode{}), "", scope).Where("id=?", code.ID).Update("used_count", gorm.Expr("used_count + 1")).Error
	})
	return &result, err
}

func containsCode(csv string, value string) bool {
	csv = strings.TrimSpace(csv)
	if csv == "" {
		return true
	}
	for _, item := range strings.Split(csv, ",") {
		if strings.EqualFold(strings.TrimSpace(item), "ALL") || strings.EqualFold(strings.TrimSpace(item), strings.TrimSpace(value)) {
			return true
		}
	}
	return false
}

func (s *Service) ListNotifications(q ListQuery) (PageResult, error) {
	var rows []model.GameNotification
	db := content.ApplySite(s.db.Model(&model.GameNotification{}), "", q.Scope).Where("deleted_at IS NULL")
	db = contains(db, "notification_id", q.Values["notificationID"])
	db = contains(db, "title_zh", q.Values["title"])
	if channel := strings.ToUpper(strings.TrimSpace(q.Values["channel"])); channel != "" {
		db = db.Where("FIND_IN_SET(?, channel)>0", channel)
	}
	db = equal(db, "audience_type", q.Values["audienceType"])
	db = equal(db, "status", q.Values["status"])
	db = rangeFilter(db, "scheduled_at", q.Values["scheduledFrom"], q.Values["scheduledTo"])
	db = rangeFilter(db, "sent_at", q.Values["sentFrom"], q.Values["sentTo"])
	return page(db.Order("id DESC"), q, &rows)
}

func (s *Service) NotificationAudienceOptions(audienceType string, keyword string, selected string, scope content.Scope) ([]NotificationAudienceOption, error) {
	switch strings.ToUpper(strings.TrimSpace(audienceType)) {
	case "SELECTED_USERS":
		type userOption struct {
			ID               uint64
			UserNo           string
			Username         string
			Nickname         string
			PhoneCountryCode string
			Phone            string
			VIPLevel         uint
		}
		var rows []userOption
		query := content.ApplySite(s.db.Table("game_users"), "", scope).
			Select("id,user_no,username,nickname,phone_country_code,phone,vip_level").
			Where("deleted_at IS NULL")
		if values := notificationaudience.Values(selected); len(values) > 0 {
			query = query.Where("id IN ?", values)
		} else if keyword = strings.TrimSpace(keyword); keyword != "" {
			like := "%" + keyword + "%"
			query = query.Where("phone LIKE ? OR user_no LIKE ? OR username LIKE ? OR nickname LIKE ?", like, like, like, like)
		}
		if err := query.Order("id DESC").Limit(50).Scan(&rows).Error; err != nil {
			return nil, err
		}
		options := make([]NotificationAudienceOption, 0, len(rows))
		for _, row := range rows {
			phone := strings.TrimSpace(row.PhoneCountryCode + " " + row.Phone)
			name := strings.TrimSpace(row.Nickname)
			if name == "" {
				name = strings.TrimSpace(row.Username)
			}
			label := phone
			if name != "" {
				label += " - " + name
			}
			options = append(options, NotificationAudienceOption{Label: label, Value: fmt.Sprint(row.ID), Phone: phone, UserNo: row.UserNo, VIPLevel: row.VIPLevel})
		}
		return options, nil
	case "VIP_LEVEL":
		type vipOption struct {
			Level  uint
			NameZh string
			NameEn string
		}
		var rows []vipOption
		if err := content.ApplySite(s.db.Table("game_vip_levels"), "", scope).
			Select("level_no AS level,name_zh,name_en").
			Where("deleted_at IS NULL AND status=1").
			Order("level_no ASC").Scan(&rows).Error; err != nil {
			return nil, err
		}
		options := make([]NotificationAudienceOption, 0, len(rows))
		for _, row := range rows {
			options = append(options, NotificationAudienceOption{Label: fmt.Sprintf("VIP %d - %s / %s", row.Level, row.NameZh, row.NameEn), Value: fmt.Sprint(row.Level), VIPLevel: row.Level, NameZh: row.NameZh, NameEn: row.NameEn})
		}
		return options, nil
	case "USER_TAG":
		type tagOption struct {
			ID      uint64
			Name    string
			TagType string
		}
		var rows []tagOption
		query := content.ApplySite(s.db.Table("game_user_tags"), "", scope).
			Select("id,name,tag_type").
			Where("deleted_at IS NULL")
		if values := notificationaudience.Values(selected); len(values) > 0 {
			query = query.Where("id IN ? OR name IN ?", values, values)
		} else if keyword = strings.TrimSpace(keyword); keyword != "" {
			like := "%" + keyword + "%"
			query = query.Where("name LIKE ? OR tag_type LIKE ?", like, like)
		}
		if err := query.Order("tag_type ASC, name ASC").Limit(100).Scan(&rows).Error; err != nil {
			return nil, err
		}
		options := make([]NotificationAudienceOption, 0, len(rows))
		for _, row := range rows {
			options = append(options, NotificationAudienceOption{
				Label:   row.Name,
				Value:   fmt.Sprint(row.ID),
				TagType: row.TagType,
			})
		}
		return options, nil
	default:
		return []NotificationAudienceOption{}, nil
	}
}

func (s *Service) GetNotification(id uint64, scope content.Scope) (map[string]any, error) {
	var row model.GameNotification
	if err := content.ApplySite(s.db.Model(&model.GameNotification{}), "", scope).Where("id=? AND deleted_at IS NULL", id).First(&row).Error; err != nil {
		return nil, notFound(err, "NOTIFICATION_NOT_FOUND")
	}
	var jobs []model.GameNotificationJob
	content.ApplySite(s.db.Model(&model.GameNotificationJob{}), "", scope).Where("notification_id=? AND deleted_at IS NULL", id).Order("id DESC").Find(&jobs)
	var deliveries []model.GameNotificationDelivery
	content.ApplySite(s.db.Model(&model.GameNotificationDelivery{}), "", scope).Where("notification_id=? AND deleted_at IS NULL", id).Order("id DESC").Limit(200).Find(&deliveries)
	logs, _ := s.operationLogs("notification", id, scope)
	return map[string]any{"basicInformation": row, "chineseContent": map[string]any{"title": row.TitleZh, "content": row.ContentZh}, "englishContent": map[string]any{"title": row.TitleEn, "content": row.ContentEn}, "audience": map[string]any{"type": row.AudienceType, "value": row.AudienceValue, "targetUserCount": row.TargetUserCount}, "schedule": map[string]any{"scheduledAt": row.ScheduledAt, "sentAt": row.SentAt}, "sendingResults": map[string]any{"successCount": row.SuccessCount, "failureCount": row.FailureCount, "jobs": jobs, "deliveries": deliveries}, "operationLogs": logs}, nil
}
func (s *Service) SaveNotification(id uint64, in NotificationInput, actor Actor) (uint64, error) {
	if !notificationchannel.Valid(in.Channel) {
		return 0, newError("UNSUPPORTED_NOTIFICATION_CHANNEL", "notification channel is not supported")
	}
	in.Channel = notificationchannel.Normalize(in.Channel)
	if err := validateNotificationAudience(s.db, in.AudienceType, in.AudienceValue, actor.Scope); err != nil {
		return 0, err
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var before model.GameNotification
		action := "create"
		permission := "operation:notification:create"
		status := "DRAFT"
		notificationID := generatedCode("NTF")
		if id > 0 {
			if err := content.ApplySite(tx.Model(&model.GameNotification{}), "", actor.Scope).Where("id=? AND deleted_at IS NULL", id).First(&before).Error; err != nil {
				return notFound(err, "NOTIFICATION_NOT_FOUND")
			}
			if before.Status == "SENDING" || before.Status == "SENT" {
				return newError("NOTIFICATION_NOT_EDITABLE", "notification cannot be edited")
			}
			action = "update"
			permission = "operation:notification:update"
			status = before.Status
			notificationID = before.NotificationID
		}
		row := model.GameNotification{ID: id, SiteID: actor.Scope.SiteID, NotificationID: notificationID, TitleZh: in.TitleZh, TitleEn: in.TitleEn, ContentZh: in.ContentZh, ContentEn: in.ContentEn, Channel: in.Channel, AudienceType: in.AudienceType, AudienceValue: in.AudienceValue, Status: status, ScheduledAt: in.ScheduledAt}
		if id == 0 {
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			id = row.ID
		} else if err := content.ApplySite(tx.Model(&model.GameNotification{}), "", actor.Scope).Where("id=?", id).Select(
			"title_zh", "title_en", "content_zh", "content_en", "channel", "audience_type", "audience_value", "scheduled_at",
		).Updates(row).Error; err != nil {
			return err
		}
		if err := content.ApplySite(tx.Model(&model.GameNotification{}), "", actor.Scope).Where("id=?", id).First(&row).Error; err != nil {
			return err
		}
		return audit(tx, actor, permission, action, "notification", id, before, row, "")
	})
	return id, err
}
func audienceValues(value string) []string {
	return notificationaudience.Values(value)
}

func notificationAudienceCount(tx *gorm.DB, row model.GameNotification) (int64, error) {
	scope := content.Scope{SiteID: row.SiteID}
	if err := validateNotificationAudience(tx, row.AudienceType, row.AudienceValue, scope); err != nil {
		return 0, err
	}
	users, err := notificationaudience.UsersScoped(tx, scope, row.AudienceType, row.AudienceValue)
	if err != nil {
		return 0, err
	}
	var count int64
	if err := users.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func validateNotificationAudience(tx *gorm.DB, audienceType string, audienceValue string, scope content.Scope) error {
	_, err := notificationaudience.UsersScoped(tx, scope, audienceType, audienceValue)
	if errors.Is(err, notificationaudience.ErrAudienceRequired) {
		return newError("NOTIFICATION_AUDIENCE_REQUIRED", err.Error())
	}
	if errors.Is(err, notificationaudience.ErrUnsupportedAudience) {
		return newError("UNSUPPORTED_NOTIFICATION_AUDIENCE", err.Error())
	}
	return err
}

func (s *Service) QueueNotification(id uint64, scheduled *time.Time, reason string, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row model.GameNotification
		if err := content.ApplySite(tx.Model(&model.GameNotification{}), "", actor.Scope).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=? AND deleted_at IS NULL", id).First(&row).Error; err != nil {
			return notFound(err, "NOTIFICATION_NOT_FOUND")
		}
		if row.Status == "SENT" || row.Status == "SENDING" {
			return nil
		}
		targetUserCount, err := notificationAudienceCount(tx, row)
		if err != nil {
			return err
		}
		before := row
		status := "SCHEDULED"
		when := scheduled
		if when == nil {
			now := time.Now()
			when = &now
			status = "SENDING"
		}
		key := fmt.Sprintf("notification:%d:%d", id, time.Now().UnixNano())
		job := model.GameNotificationJob{SiteID: row.SiteID, NotificationID: id, JobType: "SEND", Status: "PENDING", ScheduledAt: when, IdempotencyKey: key}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&job).Error; err != nil {
			return err
		}
		row.Status = status
		row.ScheduledAt = when
		row.TargetUserCount = uint(targetUserCount)
		if err := content.ApplySite(tx.Model(&model.GameNotification{}), "", actor.Scope).Where("id=?", id).Updates(map[string]any{"status": status, "scheduled_at": when, "target_user_count": row.TargetUserCount}).Error; err != nil {
			return err
		}
		return audit(tx, actor, "operation:notification:send", "queue", "notification", id, before, row, reason)
	})
}
func (s *Service) CancelNotification(id uint64, reason string, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row model.GameNotification
		if err := content.ApplySite(tx.Model(&model.GameNotification{}), "", actor.Scope).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=? AND deleted_at IS NULL", id).First(&row).Error; err != nil {
			return notFound(err, "NOTIFICATION_NOT_FOUND")
		}
		if row.Status == "CANCELLED" {
			return nil
		}
		if row.Status == "SENT" {
			return newError("NOTIFICATION_ALREADY_SENT", "sent notification cannot be cancelled")
		}
		before := row
		row.Status = "CANCELLED"
		if err := content.ApplySite(tx.Model(&model.GameNotification{}), "", actor.Scope).Where("id=?", id).Update("status", row.Status).Error; err != nil {
			return err
		}
		if err := content.ApplySite(tx.Model(&model.GameNotificationJob{}), "", actor.Scope).Where("notification_id=? AND status='PENDING'", id).Update("status", "CANCELLED").Error; err != nil {
			return err
		}
		return audit(tx, actor, "operation:notification:cancel", "cancel", "notification", id, before, row, reason)
	})
}
func (s *Service) DeleteNotification(id uint64, actor Actor) error {
	var row model.GameNotification
	if err := content.ApplySite(s.db.Model(&model.GameNotification{}), "", actor.Scope).Where("id=? AND deleted_at IS NULL", id).First(&row).Error; err != nil {
		return notFound(err, "NOTIFICATION_NOT_FOUND")
	}
	if row.Status != "DRAFT" && row.Status != "CANCELLED" {
		return newError("NOTIFICATION_NOT_DELETABLE", "only draft or cancelled notifications can be deleted")
	}
	return s.softDelete("game_notifications", "notification", id, "operation:notification:update", actor)
}
