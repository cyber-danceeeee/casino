package depositconfig

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/content"
	"casino/internal/model"
	"casino/internal/operationdefaults"
)

type methodRow struct {
	ID           uint64           `json:"id"`
	MethodCode   string           `json:"methodCode"`
	MethodType   string           `json:"methodType"`
	Name         string           `json:"name"`
	IconURL      string           `json:"iconURL"`
	MinAmount    *decimal.Decimal `json:"minAmount"`
	MaxAmount    *decimal.Decimal `json:"maxAmount"`
	Sort         int              `json:"sort"`
	Status       int              `json:"status"`
	ChannelCount int64            `json:"channelCount" gorm:"column:channel_count"`
	CreatedAt    time.Time        `json:"createdAt"`
	UpdatedAt    time.Time        `json:"updatedAt"`
}

type channelRow struct {
	ID           uint64           `json:"id"`
	ChannelCode  string           `json:"channelCode"`
	Name         string           `json:"name"`
	MethodID     uint64           `json:"methodId"`
	MethodCode   string           `json:"methodCode"`
	MethodName   string           `json:"methodName"`
	CountryCode  string           `json:"countryCode"`
	CurrencyCode string           `json:"currencyCode"`
	FeeRate      *decimal.Decimal `json:"feeRate"`
	MinAmount    *decimal.Decimal `json:"minAmount"`
	MaxAmount    *decimal.Decimal `json:"maxAmount"`
	Sort         int              `json:"sort"`
	Status       int              `json:"status"`
	CreatedAt    time.Time        `json:"createdAt"`
	UpdatedAt    time.Time        `json:"updatedAt"`
}

type amountOptionRow struct {
	ID                  uint64           `json:"id"`
	OptionCode          string           `json:"optionCode"`
	CurrencyCode        string           `json:"currencyCode"`
	Amount              *decimal.Decimal `json:"amount"`
	DefaultBonusReward  *decimal.Decimal `json:"defaultBonusReward" gorm:"column:bonus_reward"`
	DefaultCashReward   *decimal.Decimal `json:"defaultCashReward" gorm:"column:cash_reward"`
	DefaultRewardAmount string           `json:"defaultRewardAmount" gorm:"-"`
	Sort                int              `json:"sort"`
	Status              int              `json:"status"`
	CreatedAt           time.Time        `json:"createdAt"`
	UpdatedAt           time.Time        `json:"updatedAt"`
}

type voucherRow struct {
	model.GameDepositVouchers
	operationdefaults.Metadata
	UsageCount int64 `json:"usageCount" gorm:"-"`
	TierCount  int64 `json:"tierCount" gorm:"-"`
}

type voucherUsageRow struct {
	ID                  uint64           `json:"id"`
	OrderNo             string           `json:"orderNo"`
	UserID              uint64           `json:"userID"`
	UserNo              string           `json:"userNo"`
	Username            string           `json:"username"`
	Amount              *decimal.Decimal `json:"amount"`
	Currency            string           `json:"currency"`
	MethodCode          string           `json:"methodCode"`
	ChannelCode         string           `json:"channelCode"`
	Status              int              `json:"status"`
	VoucherBonusReward  *decimal.Decimal `json:"voucherBonusReward"`
	VoucherCashReward   *decimal.Decimal `json:"voucherCashReward"`
	VoucherRewardAmount *decimal.Decimal `json:"voucherRewardAmount"`
	CreatedAt           time.Time        `json:"createdAt"`
	PaidAt              *time.Time       `json:"paidAt"`
}

func statusFilter(db *gorm.DB, column string, value string) *gorm.DB {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "1", "TRUE", "ENABLED", "ACTIVE":
		return db.Where(column + " = 1")
	case "2", "FALSE", "DISABLED":
		return db.Where(column + " = 2")
	default:
		return db
	}
}

func (s *Service) operationLogs(targetType string, targetID uint64, scope content.Scope) []model.GameAdminOperationLogs {
	var rows []model.GameAdminOperationLogs
	db := s.db.Where("target_type = ? AND target_id = ? AND deleted_at IS NULL", targetType, targetID)
	_ = content.ApplySite(db, "", scope).Order("id DESC").Limit(100).Find(&rows).Error
	return rows
}

func (s *Service) ListMethods(q ListQuery) (PageResult, error) {
	var rows []methodRow
	db := s.db.Table("game_payment_methods m").Select("m.*, (SELECT COUNT(*) FROM game_payment_channels c WHERE c.method_id=m.id AND c.deleted_at IS NULL) channel_count").Where("m.deleted_at IS NULL")
	db = content.ApplySite(db, "m", q.Scope)
	db = contains(db, "m.method_code", q.Values["methodCode"])
	db = contains(db, "m.name", q.Values["name"])
	db = equal(db, "m.method_type", q.Values["methodType"])
	db = statusFilter(db, "m.status", q.Values["status"])
	result, err := s.paginate(db.Order("m.sort,m.id"), q, &rows)
	for i := range rows {
		signMethodAsset(&rows[i])
	}
	result.List = rows
	return result, err
}

func (s *Service) GetMethod(id uint64, scope content.Scope) (map[string]any, error) {
	var row model.GamePaymentMethods
	if err := content.ApplySite(s.db.Model(&model.GamePaymentMethods{}), "", scope).Where("id=? AND deleted_at IS NULL", id).First(&row).Error; err != nil {
		return nil, notFound(err, "DEPOSIT_METHOD_NOT_FOUND")
	}
	signModelMethodAsset(&row)
	var channels []channelRow
	channelDB := s.db.Table("game_payment_channels c").Select("c.*,m.method_code,m.name method_name").Joins("JOIN game_payment_methods m ON m.id=c.method_id").Where("c.method_id=? AND c.deleted_at IS NULL", id)
	content.ApplySite(channelDB, "c", scope).Order("c.sort,c.id").Scan(&channels)
	return map[string]any{"overview": row, "channels": channels, "operationLogs": s.operationLogs("deposit_method", id, scope)}, nil
}

func (s *Service) SaveMethod(id uint64, in MethodInput, actor Actor) (uint64, error) {
	if err := validCode(in.MethodCode); err != nil {
		return 0, err
	}
	if in.MethodType != "deposit" && in.MethodType != "both" {
		return 0, domainError("INVALID_METHOD_TYPE", "deposit method type must be deposit or both")
	}
	if err := validStatus(in.Status); err != nil {
		return 0, err
	}
	min, err := money(in.MinAmount, "minAmount")
	if err != nil {
		return 0, err
	}
	max, err := money(in.MaxAmount, "maxAmount")
	if err != nil {
		return 0, err
	}
	if err = validRange(min, max); err != nil {
		return 0, err
	}
	var count int64
	content.ApplySite(s.db.Unscoped().Model(&model.GamePaymentMethods{}), "", actor.Scope).Where("method_code=? AND id<>?", strings.TrimSpace(in.MethodCode), id).Count(&count)
	if count > 0 {
		return 0, duplicateCodeError("deposit method")
	}
	row := model.GamePaymentMethods{SiteID: actor.Scope.SiteID, MethodCode: strings.TrimSpace(in.MethodCode), MethodType: in.MethodType, Name: strings.TrimSpace(in.Name), IconURL: depositAssetRef(in.IconURL), MinAmount: moneyPtr(min), MaxAmount: moneyPtr(max), Sort: in.Sort, Status: in.Status}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var before any = map[string]any{}
		created := id == 0
		if created {
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			id = row.ID
		} else {
			var old model.GamePaymentMethods
			oldQuery := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.GamePaymentMethods{}).Where("id=? AND deleted_at IS NULL", id)
			if err := content.ApplySite(oldQuery, "", actor.Scope).First(&old).Error; err != nil {
				return notFound(err, "DEPOSIT_METHOD_NOT_FOUND")
			}
			before = old
			if err := tx.Model(&old).Updates(map[string]any{"method_code": row.MethodCode, "method_type": row.MethodType, "name": row.Name, "icon_url": row.IconURL, "min_amount": min, "max_amount": max, "sort": row.Sort, "status": row.Status}).Error; err != nil {
				return err
			}
		}
		action := "update"
		permission := "fund:deposit-method:update"
		if created {
			action = "create"
			permission = "fund:deposit-method:create"
		}
		return s.audit(tx, actor, permission, action, "deposit_method", id, before, row)
	})
	return id, err
}

func (s *Service) DeleteMethod(id uint64, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row model.GamePaymentMethods
		if err := content.ApplySite(tx.Model(&model.GamePaymentMethods{}), "", actor.Scope).Where("id=? AND deleted_at IS NULL", id).First(&row).Error; err != nil {
			return notFound(err, "DEPOSIT_METHOD_NOT_FOUND")
		}
		var used int64
		content.ApplySite(tx.Model(&model.GamePaymentChannels{}), "", actor.Scope).Where("method_id=? AND deleted_at IS NULL", id).Count(&used)
		if used > 0 {
			return domainError("DEPOSIT_METHOD_IN_USE", "deposit method is used by channels")
		}
		content.ApplySite(tx.Model(&model.GameDepositOrders{}), "", actor.Scope).Where("method_code=?", row.MethodCode).Count(&used)
		if used > 0 {
			return domainError("DEPOSIT_METHOD_IN_USE", "deposit method is used by deposit orders")
		}
		if err := tx.Delete(&row).Error; err != nil {
			return err
		}
		return s.audit(tx, actor, "fund:deposit-method:delete", "delete", "deposit_method", id, row, map[string]any{"deleted": true})
	})
}

func (s *Service) ListChannels(q ListQuery) (PageResult, error) {
	var rows []channelRow
	db := s.db.Table("game_payment_channels c").Select("c.*,m.method_code,m.name method_name").Joins("JOIN game_payment_methods m ON m.id=c.method_id AND m.deleted_at IS NULL").Where("c.deleted_at IS NULL")
	db = content.ApplySite(db, "c", q.Scope)
	db = contains(db, "c.channel_code", q.Values["channelCode"])
	db = contains(db, "c.name", q.Values["name"])
	db = equal(db, "m.method_code", q.Values["methodCode"])
	db = equal(db, "c.country_code", strings.ToUpper(q.Values["countryCode"]))
	db = equal(db, "c.currency_code", strings.ToUpper(q.Values["currencyCode"]))
	db = statusFilter(db, "c.status", q.Values["status"])
	return s.paginate(db.Order("c.sort,c.id"), q, &rows)
}

func (s *Service) GetChannel(id uint64, scope content.Scope) (map[string]any, error) {
	var row channelRow
	db := s.db.Table("game_payment_channels c").Select("c.*,m.method_code,m.name method_name").Joins("JOIN game_payment_methods m ON m.id=c.method_id").Where("c.id=? AND c.deleted_at IS NULL", id)
	if err := content.ApplySite(db, "c", scope).Scan(&row).Error; err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, domainError("DEPOSIT_CHANNEL_NOT_FOUND", "deposit channel not found")
	}
	return map[string]any{"overview": row, "method": map[string]any{"id": row.MethodID, "methodCode": row.MethodCode, "name": row.MethodName}, "operationLogs": s.operationLogs("deposit_channel", id, scope)}, nil
}

func (s *Service) SaveChannel(id uint64, in ChannelInput, actor Actor) (uint64, error) {
	if err := validCode(in.ChannelCode); err != nil {
		return 0, err
	}
	if err := validStatus(in.Status); err != nil {
		return 0, err
	}
	countryCode, currencyCode, err := s.validateChannelRegion(in.CountryCode, in.CurrencyCode)
	if err != nil {
		return 0, err
	}
	fee, err := money(in.FeeRate, "feeRate")
	if err != nil {
		return 0, err
	}
	min, err := money(in.MinAmount, "minAmount")
	if err != nil {
		return 0, err
	}
	max, err := money(in.MaxAmount, "maxAmount")
	if err != nil {
		return 0, err
	}
	if err = validRange(min, max); err != nil {
		return 0, err
	}
	var method model.GamePaymentMethods
	if err = content.ApplySite(s.db.Model(&model.GamePaymentMethods{}), "", actor.Scope).Where("id=? AND deleted_at IS NULL", in.MethodID).First(&method).Error; err != nil {
		return 0, domainError("DEPOSIT_METHOD_NOT_FOUND", "deposit method not found")
	}
	var count int64
	content.ApplySite(s.db.Unscoped().Model(&model.GamePaymentChannels{}), "", actor.Scope).
		Where("country_code=? AND currency_code=? AND channel_code=? AND id<>?", countryCode, currencyCode, strings.TrimSpace(in.ChannelCode), id).Count(&count)
	if count > 0 {
		return 0, duplicateCodeError("deposit channel")
	}
	row := model.GamePaymentChannels{SiteID: method.SiteID, MethodID: in.MethodID, ChannelCode: strings.TrimSpace(in.ChannelCode), Name: strings.TrimSpace(in.Name), CountryCode: countryCode, CurrencyCode: currencyCode, FeeRate: moneyPtr(fee), MinAmount: moneyPtr(min), MaxAmount: moneyPtr(max), Sort: in.Sort, Status: in.Status}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		before := any(map[string]any{})
		created := id == 0
		if created {
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			id = row.ID
		} else {
			var old model.GamePaymentChannels
			oldQuery := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.GamePaymentChannels{}).Where("id=? AND deleted_at IS NULL", id)
			if err := content.ApplySite(oldQuery, "", actor.Scope).First(&old).Error; err != nil {
				return notFound(err, "DEPOSIT_CHANNEL_NOT_FOUND")
			}
			before = old
			if err := tx.Model(&old).Updates(map[string]any{"method_id": row.MethodID, "channel_code": row.ChannelCode, "name": row.Name, "country_code": row.CountryCode, "currency_code": row.CurrencyCode, "fee_rate": fee, "min_amount": min, "max_amount": max, "sort": row.Sort, "status": row.Status}).Error; err != nil {
				return err
			}
		}
		action := "update"
		permission := "fund:deposit-channel:update"
		if created {
			action = "create"
			permission = "fund:deposit-channel:create"
		}
		return s.audit(tx, actor, permission, action, "deposit_channel", id, before, row)
	})
	return id, err
}

func (s *Service) validateChannelRegion(countryCode string, currencyCode string) (string, string, error) {
	countryCode = strings.ToUpper(strings.TrimSpace(countryCode))
	currencyCode = strings.ToUpper(strings.TrimSpace(currencyCode))
	if countryCode == "" || currencyCode == "" {
		return "", "", domainError("INVALID_CHANNEL_REGION", "channel country and currency are required")
	}
	var count int64
	if err := s.db.Table("game_country_configs").
		Where("country_code=? AND currency_code=? AND status=1 AND deleted_at IS NULL", countryCode, currencyCode).
		Count(&count).Error; err != nil {
		return "", "", err
	}
	if count == 0 {
		return "", "", domainError("INVALID_CHANNEL_REGION", "channel country and currency do not match the platform country catalog")
	}
	return countryCode, currencyCode, nil
}

func (s *Service) amountOptionCurrency(inputCurrency string, scope content.Scope) (string, error) {
	if !scope.SiteScoped() {
		currency := strings.ToUpper(strings.TrimSpace(inputCurrency))
		if currency == "" {
			currency = "PKR"
		}
		return currency, nil
	}
	var row struct {
		SiteCurrency string `gorm:"column:site_currency"`

		CountryCurrency string `gorm:"column:country_currency"`
	}
	err := s.db.Table("game_site_config s").
		Select("s.currency_code site_currency,c.currency_code country_currency").
		Joins("LEFT JOIN game_country_configs c ON c.country_code=s.country_code AND c.status=1 AND c.deleted_at IS NULL").
		Where("s.id = ? AND s.deleted_at IS NULL", scope.SiteID).
		Take(&row).Error
	if err != nil {
		return "", err
	}
	currency := strings.ToUpper(strings.TrimSpace(firstNonEmpty(row.SiteCurrency, row.CountryCurrency)))
	if currency == "" {
		return "", domainError("SITE_CURRENCY_NOT_CONFIGURED", "H5 currency is not configured")
	}
	return currency, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func (s *Service) DeleteChannel(id uint64, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row model.GamePaymentChannels
		if err := content.ApplySite(tx.Model(&model.GamePaymentChannels{}), "", actor.Scope).Where("id=? AND deleted_at IS NULL", id).First(&row).Error; err != nil {
			return notFound(err, "DEPOSIT_CHANNEL_NOT_FOUND")
		}
		var used int64
		content.ApplySite(tx.Model(&model.GameDepositOrders{}), "", actor.Scope).Where("channel_code=? AND currency=?", row.ChannelCode, row.CurrencyCode).Count(&used)
		if used > 0 {
			return domainError("DEPOSIT_CHANNEL_IN_USE", "deposit channel is used by deposit orders")
		}
		if err := tx.Delete(&row).Error; err != nil {
			return err
		}
		return s.audit(tx, actor, "fund:deposit-channel:delete", "delete", "deposit_channel", id, row, map[string]any{"deleted": true})
	})
}

func (s *Service) ListAmountOptions(q ListQuery) (PageResult, error) {
	var rows []amountOptionRow
	db := s.db.Table("game_deposit_amount_options").Where("deleted_at IS NULL")
	db = content.ApplySite(db, "", q.Scope)
	db = contains(db, "option_code", q.Values["optionCode"])
	db = equal(db, "currency_code", strings.ToUpper(q.Values["currencyCode"]))
	db = statusFilter(db, "status", q.Values["status"])
	result, err := s.paginate(db.Order("sort,id"), q, &rows)
	for i := range rows {
		rows[i].DefaultRewardAmount = rewardTotal(rows[i].DefaultBonusReward, rows[i].DefaultCashReward).StringFixed(2)
	}
	result.List = rows
	return result, err
}

func (s *Service) GetAmountOption(id uint64, scope content.Scope) (map[string]any, error) {
	var row amountOptionRow
	db := s.db.Table("game_deposit_amount_options").Where("id=? AND deleted_at IS NULL", id)
	if err := content.ApplySite(db, "", scope).Scan(&row).Error; err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, domainError("DEPOSIT_OPTION_NOT_FOUND", "deposit amount option not found")
	}
	row.DefaultRewardAmount = rewardTotal(row.DefaultBonusReward, row.DefaultCashReward).StringFixed(2)
	var tiers []model.GameDepositVoucherRewards
	content.ApplySite(s.db.Model(&model.GameDepositVoucherRewards{}), "", scope).Where("amount_option_id=? AND deleted_at IS NULL", id).Order("id DESC").Find(&tiers)
	return map[string]any{"overview": row, "voucherRewardTiers": tiers, "operationLogs": s.operationLogs("deposit_amount_option", id, scope)}, nil
}

func (s *Service) SaveAmountOption(id uint64, in AmountOptionInput, actor Actor) (uint64, error) {
	if err := validCode(in.OptionCode); err != nil {
		return 0, err
	}
	if err := validStatus(in.Status); err != nil {
		return 0, err
	}
	currencyCode, err := s.amountOptionCurrency(in.CurrencyCode, actor.Scope)
	if err != nil {
		return 0, err
	}
	amount, err := money(in.Amount, "amount")
	if err != nil || amount.IsZero() {
		return 0, domainError("INVALID_AMOUNT", "amount must be greater than zero")
	}
	bonus, err := money(in.DefaultBonusReward, "defaultBonusReward")
	if err != nil {
		return 0, err
	}
	cash, err := money(in.DefaultCashReward, "defaultCashReward")
	if err != nil {
		return 0, err
	}
	total, err := money(in.DefaultRewardAmount, "defaultRewardAmount")
	if err != nil {
		return 0, err
	}
	if !total.Equal(bonus.Add(cash)) {
		return 0, domainError("INVALID_REWARD_AMOUNT", "reward amount must equal bonus reward plus cash reward")
	}
	var count int64
	content.ApplySite(s.db.Unscoped().Model(&model.GameDepositAmountOptions{}), "", actor.Scope).Where("currency_code=? AND option_code=? AND id<>?", currencyCode, strings.TrimSpace(in.OptionCode), id).Count(&count)
	if count > 0 {
		return 0, duplicateCodeError("deposit amount option")
	}
	if in.Status == 1 {
		content.ApplySite(s.db.Model(&model.GameDepositAmountOptions{}), "", actor.Scope).Where("currency_code=? AND amount=? AND status=1 AND id<>? AND deleted_at IS NULL", currencyCode, amount, id).Count(&count)
		if count > 0 {
			return 0, domainError("DUPLICATE_ACTIVE_AMOUNT", "an active deposit amount option already uses this amount")
		}
	}
	row := model.GameDepositAmountOptions{SiteID: actor.Scope.SiteID, OptionCode: strings.TrimSpace(in.OptionCode), CurrencyCode: currencyCode, Amount: moneyPtr(amount), BonusReward: moneyPtr(bonus), CashReward: moneyPtr(cash), Sort: in.Sort, Status: in.Status}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		before := any(map[string]any{})
		created := id == 0
		if created {
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			id = row.ID
		} else {
			var old model.GameDepositAmountOptions
			oldQuery := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.GameDepositAmountOptions{}).Where("id=? AND deleted_at IS NULL", id)
			if err := content.ApplySite(oldQuery, "", actor.Scope).First(&old).Error; err != nil {
				return notFound(err, "DEPOSIT_OPTION_NOT_FOUND")
			}
			before = old
			if err := tx.Model(&old).Updates(map[string]any{"option_code": row.OptionCode, "currency_code": currencyCode, "amount": amount, "bonus_reward": bonus, "cash_reward": cash, "sort": row.Sort, "status": row.Status}).Error; err != nil {
				return err
			}
		}
		action := "update"
		permission := "fund:deposit-option:update"
		if created {
			action = "create"
			permission = "fund:deposit-option:create"
		}
		return s.audit(tx, actor, permission, action, "deposit_amount_option", id, before, row)
	})
	return id, err
}

func (s *Service) DeleteAmountOption(id uint64, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row model.GameDepositAmountOptions
		if err := content.ApplySite(tx.Model(&model.GameDepositAmountOptions{}), "", actor.Scope).Where("id=? AND deleted_at IS NULL", id).First(&row).Error; err != nil {
			return notFound(err, "DEPOSIT_OPTION_NOT_FOUND")
		}
		var used int64
		content.ApplySite(tx.Model(&model.GameDepositVoucherRewards{}), "", actor.Scope).Where("amount_option_id=? AND deleted_at IS NULL", id).Count(&used)
		if used > 0 {
			return domainError("DEPOSIT_OPTION_IN_USE", "deposit amount option is referenced by voucher reward tiers")
		}
		if err := tx.Delete(&row).Error; err != nil {
			return err
		}
		return s.audit(tx, actor, "fund:deposit-option:delete", "delete", "deposit_amount_option", id, row, map[string]any{"deleted": true})
	})
}

func (s *Service) ListVouchers(q ListQuery) (PageResult, error) {
	var rows []voucherRow
	base := func(db *gorm.DB) *gorm.DB {
		db = contains(db, "v.voucher_code", q.Values["voucherCode"])
		if title := strings.TrimSpace(q.Values["title"]); title != "" {
			db = db.Where("(v.title_zh LIKE ? OR v.title_en LIKE ? OR v.title LIKE ?)", "%"+title+"%", "%"+title+"%", "%"+title+"%")
		}
		db = equal(db, "v.voucher_type", q.Values["voucherType"])
		switch strings.ToUpper(strings.TrimSpace(q.Values["availability"])) {
		case "AVAILABLE":
			now := time.Now()
			db = db.Where("(v.start_at IS NULL OR v.start_at<=?) AND (v.end_at IS NULL OR v.end_at>=?)", now, now)
		case "UNAVAILABLE":
			now := time.Now()
			db = db.Where("(v.start_at IS NOT NULL AND v.start_at>?) OR (v.end_at IS NOT NULL AND v.end_at<?)", now, now)
		}
		if v := q.Values["amountMin"]; v != "" {
			db = db.Where("v.max_amount>=?", v)
		}
		if v := q.Values["amountMax"]; v != "" {
			db = db.Where("v.min_amount<=?", v)
		}
		if v := q.Values["startFrom"]; v != "" {
			db = db.Where("v.start_at>=?", v)
		}
		if v := q.Values["endTo"]; v != "" {
			db = db.Where("v.end_at<=?", v)
		}
		return db
	}
	db := s.db.Table("game_deposit_vouchers v").
		Select("v.*, (SELECT COUNT(*) FROM game_deposit_voucher_rewards r WHERE r.voucher_id=v.id AND r.deleted_at IS NULL) tier_count, (SELECT COUNT(*) FROM game_deposit_orders d WHERE d.voucher_id=v.id AND d.status=2 AND d.deleted_at IS NULL) usage_count").
		Where("v.deleted_at IS NULL")
	db = content.ApplySite(db, "v", q.Scope)
	db = base(db)
	if !q.Scope.SiteScoped() {
		db = statusFilter(db, "v.status", q.Values["status"])
		result, err := s.paginate(db.Order("v.sort,v.id"), q, &rows)
		for i := range rows {
			signVoucherAsset(&rows[i].GameDepositVouchers)
		}
		result.List = rows
		return result, err
	}
	db = statusFilter(db, "v.status", q.Values["status"])
	result, err := s.paginate(db.Order("v.sort,v.id"), q, &rows)
	if err != nil {
		return result, err
	}
	for i := range rows {
		rows[i].Metadata = operationdefaults.Metadata{CanEdit: true, CanDelete: true, Visible: rows[i].Status == 1}
		signVoucherAsset(&rows[i].GameDepositVouchers)
	}
	result.List = rows
	return result, nil
}

func (s *Service) GetVoucher(id uint64, scope content.Scope) (map[string]any, error) {
	var row model.GameDepositVouchers
	if err := content.ApplySite(s.db.Model(&model.GameDepositVouchers{}), "", scope).Where("id=? AND deleted_at IS NULL", id).First(&row).Error; err != nil {
		return nil, notFound(err, "VOUCHER_NOT_FOUND")
	}
	signVoucherAsset(&row)
	overview := voucherRow{GameDepositVouchers: row, Metadata: operationdefaults.Metadata{CanEdit: true, CanDelete: true, Visible: row.Status == 1}}
	var tiers []model.GameDepositVoucherRewards
	content.ApplySite(s.db.Model(&model.GameDepositVoucherRewards{}), "", scope).Where("voucher_id=? AND deleted_at IS NULL", id).Order("sort,id").Find(&tiers)
	var usage []map[string]any
	usageDB := s.db.Table("game_deposit_orders d").Select("d.id,d.order_no,d.user_id,u.user_no,u.username,d.amount,d.currency,d.method_code,d.channel_code,d.status,d.created_at,d.paid_at,d.voucher_bonus_reward,d.voucher_cash_reward,d.voucher_reward_amount").Joins("LEFT JOIN game_users u ON u.id=d.user_id").Where("d.voucher_id=? AND d.deleted_at IS NULL", id)
	content.ApplySite(usageDB, "d", scope).Order("d.id DESC").Limit(100).Scan(&usage)
	var successful int64
	var reward decimal.Decimal
	content.ApplySite(s.db.Model(&model.GameDepositOrders{}), "", scope).Where("voucher_id=? AND status=2", id).Count(&successful)
	content.ApplySite(s.db.Model(&model.GameDepositOrders{}), "", scope).Select("COALESCE(SUM(voucher_reward_amount),0)").Where("voucher_id=? AND status=2", id).Scan(&reward)
	return map[string]any{"overview": overview, "rewardTiers": tiers, "eligibilityRules": map[string]any{"newUsersOnly": row.NewUsersOnly, "firstDepositOnly": row.FirstDepositOnly, "dailyLimit": row.DailyLimit, "weeklyLimit": row.WeeklyLimit, "perUserLimit": row.PerUserLimit, "countries": row.Countries, "currencies": row.Currencies, "methodCodes": row.MethodCodes}, "usageRecords": usage, "statistics": map[string]any{"usageCount": successful, "rewardAmount": reward.StringFixed(2)}, "operationLogs": s.operationLogs("deposit_voucher", id, scope)}, nil
}

func (s *Service) ListVoucherUsage(id uint64, q ListQuery) (PageResult, error) {
	var voucher model.GameDepositVouchers
	if err := content.ApplySite(s.db.Model(&model.GameDepositVouchers{}), "", q.Scope).Select("id").Where("id=? AND deleted_at IS NULL", id).First(&voucher).Error; err != nil {
		return PageResult{}, notFound(err, "VOUCHER_NOT_FOUND")
	}
	var rows []voucherUsageRow
	db := s.db.Table("game_deposit_orders d").
		Select("d.id,d.order_no,d.user_id,u.user_no,u.username,d.amount,d.currency,d.method_code,d.channel_code,d.status,d.voucher_bonus_reward,d.voucher_cash_reward,d.voucher_reward_amount,d.created_at,d.paid_at").
		Joins("LEFT JOIN game_users u ON u.id=d.user_id").
		Where("d.voucher_id=? AND d.deleted_at IS NULL", id)
	db = content.ApplySite(db, "d", q.Scope)
	db = equal(db, "d.user_id", q.Values["userID"])
	db = contains(db, "u.username", q.Values["username"])
	db = equal(db, "d.status", q.Values["status"])
	return s.paginate(db.Order("d.id DESC"), q, &rows)
}

func (s *Service) SaveVoucher(id uint64, in VoucherInput, actor Actor) (uint64, error) {
	if err := validCode(in.VoucherCode); err != nil {
		return 0, err
	}
	if err := validStatus(in.Status); err != nil {
		return 0, err
	}
	if in.StartTime != nil && in.EndTime != nil && in.EndTime.Before(*in.StartTime) {
		return 0, domainError("INVALID_TIME_RANGE", "end time must be after start time")
	}
	allowed := map[string]bool{"FIRST_DEPOSIT": true, "DEPOSIT_BONUS": true, "WEEKLY_DEPOSIT": true, "DAILY_DEPOSIT": true, "PROMOTION": true}
	if !allowed[in.VoucherType] {
		return 0, domainError("INVALID_VOUCHER_TYPE", "unsupported voucher type")
	}
	min, err := money(in.MinAmount, "minAmount")
	if err != nil {
		return 0, err
	}
	max, err := money(in.MaxAmount, "maxAmount")
	if err != nil {
		return 0, err
	}
	if err = validRange(min, max); err != nil {
		return 0, err
	}
	maximum, err := money(in.MaxReward, "maxReward")
	if err != nil {
		return 0, err
	}
	var count int64
	content.ApplySite(s.db.Unscoped().Model(&model.GameDepositVouchers{}), "", actor.Scope).Where("voucher_code=? AND id<>?", strings.TrimSpace(in.VoucherCode), id).Count(&count)
	if count > 0 {
		return 0, duplicateCodeError("voucher")
	}
	zero := decimal.Zero
	row := model.GameDepositVouchers{SiteID: actor.Scope.SiteID, VoucherCode: strings.TrimSpace(in.VoucherCode), Title: generatedTitle(in.TitleZh, in.TitleEn), TitleZh: strings.TrimSpace(in.TitleZh), TitleEn: strings.TrimSpace(in.TitleEn), DescriptionZh: in.DescriptionZh, DescriptionEn: in.DescriptionEn, VoucherType: in.VoucherType, IconURL: depositAssetRef(in.IconURL), MinAmount: moneyPtr(min), MaxAmount: moneyPtr(max), MaxReward: moneyPtr(maximum), BonusReward: &zero, CashReward: &zero, Sort: in.Sort, Status: in.Status, StartAt: in.StartTime, EndAt: in.EndTime, NewUsersOnly: in.NewUsersOnly, FirstDepositOnly: in.FirstDepositOnly, DailyLimit: in.DailyLimit, WeeklyLimit: in.WeeklyLimit, PerUserLimit: in.PerUserLimit, Countries: strings.TrimSpace(in.Countries), Currencies: strings.TrimSpace(in.Currencies), MethodCodes: strings.TrimSpace(in.MethodCodes)}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		before := any(map[string]any{})
		created := id == 0
		if created {
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			id = row.ID
		} else {
			var old model.GameDepositVouchers
			oldQuery := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.GameDepositVouchers{}).Where("id=? AND deleted_at IS NULL", id)
			if err := content.ApplySite(oldQuery, "", actor.Scope).First(&old).Error; err != nil {
				return notFound(err, "VOUCHER_NOT_FOUND")
			}
			before = old
			if err := tx.Model(&old).Updates(map[string]any{"title": row.Title, "title_zh": row.TitleZh, "title_en": row.TitleEn, "description_zh": row.DescriptionZh, "description_en": row.DescriptionEn, "voucher_type": row.VoucherType, "icon_url": row.IconURL, "min_amount": min, "max_amount": max, "max_reward": maximum, "sort": row.Sort, "status": row.Status, "start_at": row.StartAt, "end_at": row.EndAt, "new_users_only": row.NewUsersOnly, "first_deposit_only": row.FirstDepositOnly, "daily_limit": row.DailyLimit, "weekly_limit": row.WeeklyLimit, "per_user_limit": row.PerUserLimit, "countries": row.Countries, "currencies": row.Currencies, "method_codes": row.MethodCodes}).Error; err != nil {
				return err
			}
		}
		action := "update"
		permission := "operation:voucher:update"
		if created {
			action = "create"
			permission = "operation:voucher:create"
		}
		return s.audit(tx, actor, permission, action, "deposit_voucher", id, before, row)
	})
	return id, err
}

func (s *Service) SaveVoucherTiers(voucherID uint64, inputs []RewardTierInput, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var voucher model.GameDepositVouchers
		if err := content.ApplySite(tx.Model(&model.GameDepositVouchers{}), "", actor.Scope).Where("id=? AND deleted_at IS NULL", voucherID).First(&voucher).Error; err != nil {
			return notFound(err, "VOUCHER_NOT_FOUND")
		}
		var before []model.GameDepositVoucherRewards
		content.ApplySite(tx.Model(&model.GameDepositVoucherRewards{}), "", actor.Scope).Where("voucher_id=? AND deleted_at IS NULL", voucherID).Order("sort,id").Find(&before)
		seen := map[string]bool{}
		keep := make([]uint64, 0, len(inputs))
		for _, in := range inputs {
			if err := validStatus(in.Status); err != nil {
				return err
			}
			amount, err := money(in.Amount, "amount")
			if err != nil || amount.IsZero() {
				return domainError("INVALID_AMOUNT", "tier amount must be greater than zero")
			}
			bonus, err := money(in.BonusReward, "bonusReward")
			if err != nil {
				return err
			}
			cash, err := money(in.CashReward, "cashReward")
			if err != nil {
				return err
			}
			total, err := money(in.RewardAmount, "rewardAmount")
			if err != nil {
				return err
			}
			if !total.Equal(bonus.Add(cash)) {
				return domainError("INVALID_REWARD_AMOUNT", "reward amount must equal bonus reward plus cash reward")
			}
			key := amount.StringFixed(2)
			if seen[key] {
				return domainError("DUPLICATE_TIER_AMOUNT", "voucher reward tier amount must be unique")
			}
			seen[key] = true
			if in.AmountOptionID > 0 {
				var option model.GameDepositAmountOptions
				if err := content.ApplySite(tx.Model(&model.GameDepositAmountOptions{}), "", actor.Scope).Where("id=? AND deleted_at IS NULL", in.AmountOptionID).First(&option).Error; err != nil {
					return domainError("DEPOSIT_OPTION_NOT_FOUND", "deposit amount option not found")
				}
				if !amount.Equal(*option.Amount) {
					return domainError("AMOUNT_OPTION_MISMATCH", "tier amount must match the selected deposit amount option")
				}
			}
			row := model.GameDepositVoucherRewards{SiteID: voucher.SiteID, VoucherID: voucherID, AmountOptionID: in.AmountOptionID, Amount: moneyPtr(amount), BonusReward: moneyPtr(bonus), CashReward: moneyPtr(cash), RewardAmount: moneyPtr(total), Sort: in.Sort, Status: in.Status}
			if in.ID > 0 {
				var existing model.GameDepositVoucherRewards
				existingQuery := tx.Model(&model.GameDepositVoucherRewards{}).Where("id=? AND voucher_id=? AND deleted_at IS NULL", in.ID, voucherID)
				if err := content.ApplySite(existingQuery, "", actor.Scope).First(&existing).Error; err != nil {
					return notFound(err, "VOUCHER_TIER_NOT_FOUND")
				}
				if err := tx.Model(&existing).Updates(map[string]any{"amount_option_id": row.AmountOptionID, "amount": amount, "bonus_reward": bonus, "cash_reward": cash, "reward_amount": total, "sort": row.Sort, "status": row.Status}).Error; err != nil {
					return err
				}
				keep = append(keep, in.ID)
			} else {
				if err := tx.Create(&row).Error; err != nil {
					return err
				}
				keep = append(keep, row.ID)
			}
		}
		deleteQuery := content.ApplySite(tx.Model(&model.GameDepositVoucherRewards{}), "", actor.Scope).Where("voucher_id=? AND deleted_at IS NULL", voucherID)
		if len(keep) > 0 {
			deleteQuery = deleteQuery.Where("id NOT IN ?", keep)
		}
		if err := deleteQuery.Delete(&model.GameDepositVoucherRewards{}).Error; err != nil {
			return err
		}
		return s.audit(tx, actor, "operation:voucher:tier:update", "configure_tiers", "deposit_voucher", voucherID, before, inputs)
	})
}

func (s *Service) DeleteVoucher(id uint64, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row model.GameDepositVouchers
		if err := content.ApplySite(tx.Model(&model.GameDepositVouchers{}), "", actor.Scope).Where("id=? AND deleted_at IS NULL", id).First(&row).Error; err != nil {
			return notFound(err, "VOUCHER_NOT_FOUND")
		}
		var used int64
		content.ApplySite(tx.Model(&model.GameDepositOrders{}), "", actor.Scope).Where("voucher_id=?", id).Count(&used)
		if used > 0 {
			return domainError("VOUCHER_IN_USE", "voucher has usage records and cannot be deleted")
		}
		deleteRewards := content.ApplySite(tx.Model(&model.GameDepositVoucherRewards{}), "", actor.Scope).Where("voucher_id=?", id)
		if err := deleteRewards.Delete(&model.GameDepositVoucherRewards{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&row).Error; err != nil {
			return err
		}
		return s.audit(tx, actor, "operation:voucher:delete", "delete", "deposit_voucher", id, row, map[string]any{"deleted": true})
	})
}
