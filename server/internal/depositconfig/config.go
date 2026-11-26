package depositconfig

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"casino/internal/model"
	"casino/internal/tenancy"
)

type Selection struct {
	Method       model.GamePaymentMethods
	Channel      model.GamePaymentChannels
	Voucher      *model.GameDepositVouchers
	RewardTier   *model.GameDepositVoucherRewards
	BonusReward  decimal.Decimal
	CashReward   decimal.Decimal
	RewardAmount decimal.Decimal
}

type voucherEligibility struct {
	Available bool
	Reason    string
}

func depositPlatformScope() tenancy.Scope {
	return tenancy.Scope{Platform: true}
}

func depositApplyReadableScope(db *gorm.DB, alias string, scope tenancy.Scope) *gorm.DB {
	if scope.Platform {
		return db.Where(depositColumn(alias, "tenant_id")+" = ? AND "+depositColumn(alias, "site_id")+" = ?", 0, 0)
	}
	return tenancy.ApplySite(db, alias, scope)
}

func depositColumn(alias string, name string) string {
	if strings.TrimSpace(alias) == "" {
		return name
	}
	return alias + "." + name
}

func (s *Service) GetConfig(q ConfigQuery) (map[string]any, error) {
	q.Currency = strings.ToUpper(strings.TrimSpace(q.Currency))
	q.Country = strings.ToUpper(strings.TrimSpace(q.Country))
	q.MethodCode = strings.TrimSpace(q.MethodCode)

	var methods []model.GamePaymentMethods
	catalogScope := q.Scope
	loadMethods := func(scope tenancy.Scope) error {
		methods = []model.GamePaymentMethods{}
		methodDB := depositApplyReadableScope(s.db.Model(&model.GamePaymentMethods{}), "", scope).
			Where("deleted_at IS NULL AND status = 1 AND method_type IN ?", []string{"deposit", "both"})
		return methodDB.Order("sort ASC, id ASC").Find(&methods).Error
	}
	if err := loadMethods(catalogScope); err != nil {
		return nil, err
	}
	if len(methods) == 0 && q.Scope.TenantScoped() {
		catalogScope = depositPlatformScope()
		if err := loadMethods(catalogScope); err != nil {
			return nil, err
		}
	}

	methodIDs := make([]uint64, 0, len(methods))
	methodCodes := make(map[uint64]string, len(methods))
	methodRows := make([]map[string]any, 0, len(methods))
	for _, method := range methods {
		methodIDs = append(methodIDs, method.ID)
		methodCodes[method.ID] = method.MethodCode
		methodRows = append(methodRows, map[string]any{
			"id": method.ID, "methodCode": method.MethodCode, "name": method.Name,
			"methodType": method.MethodType, "iconURL": depositAssetURL(method.IconURL),
			"minAmount": decimalString(method.MinAmount), "maxAmount": decimalString(method.MaxAmount),
			"sort": method.Sort, "status": method.Status,
		})
	}

	var channels []model.GamePaymentChannels
	if len(methodIDs) > 0 {
		channelDB := depositApplyReadableScope(s.db.Model(&model.GamePaymentChannels{}), "", catalogScope).
			Where("deleted_at IS NULL AND status = 1 AND method_id IN ? AND country_code=? AND currency_code=?", methodIDs, q.Country, q.Currency)
		if err := channelDB.Order("sort ASC, id ASC").Find(&channels).Error; err != nil {
			return nil, err
		}
	}
	channelRows := make([]map[string]any, 0, len(channels))
	availableMethodIDs := make(map[uint64]bool, len(channels))
	for _, channel := range channels {
		availableMethodIDs[channel.MethodID] = true
		channelRows = append(channelRows, map[string]any{
			"id": channel.ID, "methodID": channel.MethodID, "methodCode": methodCodes[channel.MethodID],
			"channelCode": channel.ChannelCode, "name": channel.Name,
			"countryCode": channel.CountryCode, "currencyCode": channel.CurrencyCode,
			"feeRate": decimalString(channel.FeeRate), "minAmount": decimalString(channel.MinAmount),
			"maxAmount": decimalString(channel.MaxAmount), "sort": channel.Sort, "status": channel.Status,
		})
	}
	filteredMethodRows := make([]map[string]any, 0, len(methodRows))
	for _, method := range methodRows {
		methodID, _ := method["id"].(uint64)
		if availableMethodIDs[methodID] {
			filteredMethodRows = append(filteredMethodRows, method)
		}
	}
	methodRows = filteredMethodRows

	var options []model.GameDepositAmountOptions
	optionDB := depositApplyReadableScope(s.db.Model(&model.GameDepositAmountOptions{}), "", catalogScope).
		Where("deleted_at IS NULL AND status = 1")
	if err := optionDB.Order("sort ASC, id ASC").Find(&options).Error; err != nil {
		return nil, err
	}
	optionRows := make([]map[string]any, 0, len(options))
	for _, option := range options {
		optionRows = append(optionRows, map[string]any{
			"id": option.ID, "optionCode": option.OptionCode, "amount": decimalString(option.Amount),
			"bonusReward": decimalString(option.BonusReward), "cashReward": decimalString(option.CashReward),
			"rewardAmount": rewardTotal(option.BonusReward, option.CashReward).StringFixed(2),
			"sort":         option.Sort, "status": option.Status,
		})
	}

	now := time.Now()
	var vouchers []model.GameDepositVouchers
	voucherDB := tenancy.ApplySite(s.db.Model(&model.GameDepositVouchers{}), "", q.Scope).
		Where("deleted_at IS NULL AND status = 1 AND (start_at IS NULL OR start_at <= ?) AND (end_at IS NULL OR end_at >= ?)", now, now)
	if err := voucherDB.Order("sort ASC, id ASC").Find(&vouchers).Error; err != nil {
		return nil, err
	}
	voucherIDs := make([]uint64, 0, len(vouchers))
	for _, voucher := range vouchers {
		voucherIDs = append(voucherIDs, voucher.ID)
	}
	var tiers []model.GameDepositVoucherRewards
	if len(voucherIDs) > 0 {
		tierDB := s.db.Model(&model.GameDepositVoucherRewards{}).
			Where("deleted_at IS NULL AND status = 1 AND voucher_id IN ?", voucherIDs)
		if err := tierDB.Order("voucher_id ASC, sort ASC, id ASC").Find(&tiers).Error; err != nil {
			return nil, err
		}
	}
	tiersByVoucher := make(map[uint64][]model.GameDepositVoucherRewards, len(vouchers))
	for _, tier := range tiers {
		tiersByVoucher[tier.VoucherID] = append(tiersByVoucher[tier.VoucherID], tier)
	}

	voucherRows := make([]map[string]any, 0, len(vouchers))
	for _, voucher := range vouchers {
		eligibility, err := s.voucherEligibility(voucher, q, now)
		if err != nil {
			return nil, err
		}
		if !eligibility.Available {
			continue
		}
		rewardRows := make([]map[string]any, 0, len(tiersByVoucher[voucher.ID]))
		for _, tier := range tiersByVoucher[voucher.ID] {
			rewardRows = append(rewardRows, map[string]any{
				"id": tier.ID, "amountOptionID": tier.AmountOptionID, "amount": decimalString(tier.Amount),
				"bonusReward": decimalString(tier.BonusReward), "cashReward": decimalString(tier.CashReward),
				"rewardAmount": decimalString(tier.RewardAmount), "sort": tier.Sort, "status": tier.Status,
			})
		}
		title := generatedTitle(voucher.TitleZh, voucher.TitleEn)
		if title == "" {
			title = strings.TrimSpace(voucher.Title)
		}
		if strings.HasPrefix(strings.ToLower(q.Language), "zh") && voucher.TitleZh != "" {
			title = voucher.TitleZh
		}
		voucherRows = append(voucherRows, map[string]any{
			"id": voucher.ID, "voucherCode": voucher.VoucherCode, "voucherType": voucher.VoucherType,
			"title": title, "titleZh": voucher.TitleZh, "titleEn": voucher.TitleEn,
			"descriptionZh": voucher.DescriptionZh, "descriptionEn": voucher.DescriptionEn,
			"iconURL": depositAssetURL(voucher.IconURL), "minAmount": decimalString(voucher.MinAmount),
			"maxAmount":   decimalString(voucher.MaxAmount),
			"amountRange": fmt.Sprintf("%s-%s", decimalString(voucher.MinAmount), decimalString(voucher.MaxAmount)),
			"maxReward":   decimalString(voucher.MaxReward), "rewards": rewardRows,
			"available": true, "unavailableReason": "",
			"startTime": voucher.StartAt, "endTime": voucher.EndAt,
			"sort": voucher.Sort, "status": voucher.Status,
		})
	}

	return map[string]any{
		"methods": methodRows, "channels": channelRows, "amountOptions": optionRows, "vouchers": voucherRows,
	}, nil
}

func (s *Service) ResolveSelection(q ConfigQuery, channelCode, voucherCode string) (Selection, error) {
	amount, err := money(q.Amount, "amount")
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return Selection{}, domainError("INVALID_AMOUNT", "amount must be greater than zero")
	}

	var method model.GamePaymentMethods
	catalogScope := q.Scope
	methodCode := strings.TrimSpace(q.MethodCode)
	methodDB := func(scope tenancy.Scope) *gorm.DB {
		return depositApplyReadableScope(s.db.Model(&model.GamePaymentMethods{}), "", scope).
			Where("deleted_at IS NULL AND status = 1 AND method_code = ? AND method_type IN ?", methodCode, []string{"deposit", "both"})
	}
	err = methodDB(catalogScope).First(&method).Error
	if err != nil && q.Scope.TenantScoped() && err == gorm.ErrRecordNotFound {
		catalogScope = depositPlatformScope()
		err = methodDB(catalogScope).First(&method).Error
	}
	if err != nil {
		return Selection{}, notFound(err, "DEPOSIT_METHOD_NOT_FOUND")
	}
	if !amountInRange(amount, method.MinAmount, method.MaxAmount) {
		return Selection{}, domainError("METHOD_AMOUNT_OUT_OF_RANGE", "amount is outside the deposit method range")
	}

	var channel model.GamePaymentChannels
	channelDB := depositApplyReadableScope(s.db.Model(&model.GamePaymentChannels{}), "", catalogScope).
		Where("deleted_at IS NULL AND status = 1 AND method_id = ? AND channel_code = ? AND country_code=? AND currency_code=?", method.ID, strings.TrimSpace(channelCode), strings.ToUpper(strings.TrimSpace(q.Country)), strings.ToUpper(strings.TrimSpace(q.Currency)))
	if err := channelDB.First(&channel).Error; err != nil {
		return Selection{}, notFound(err, "DEPOSIT_CHANNEL_NOT_FOUND")
	}
	if !amountInRange(amount, channel.MinAmount, channel.MaxAmount) {
		return Selection{}, domainError("CHANNEL_AMOUNT_OUT_OF_RANGE", "amount is outside the deposit channel range")
	}

	selection := Selection{Method: method, Channel: channel}
	if strings.TrimSpace(voucherCode) == "" {
		var option model.GameDepositAmountOptions
		optionDB := depositApplyReadableScope(s.db.Model(&model.GameDepositAmountOptions{}), "", catalogScope).
			Where("deleted_at IS NULL AND status = 1 AND amount = ?", amount)
		if err := optionDB.First(&option).Error; err == nil {
			selection.BonusReward = valueOrZero(option.BonusReward)
			selection.CashReward = valueOrZero(option.CashReward)
			selection.RewardAmount = selection.BonusReward.Add(selection.CashReward)
		} else if err != gorm.ErrRecordNotFound {
			return Selection{}, err
		}
		return selection, nil
	}

	now := time.Now()
	var voucher model.GameDepositVouchers
	voucherDB := tenancy.ApplySite(s.db.Model(&model.GameDepositVouchers{}), "", q.Scope).
		Where("deleted_at IS NULL AND status = 1 AND voucher_code = ? AND (start_at IS NULL OR start_at <= ?) AND (end_at IS NULL OR end_at >= ?)", strings.TrimSpace(voucherCode), now, now)
	if err := voucherDB.First(&voucher).Error; err != nil {
		return Selection{}, notFound(err, "VOUCHER_NOT_FOUND")
	}
	if !amountInRange(amount, voucher.MinAmount, voucher.MaxAmount) {
		return Selection{}, domainError("VOUCHER_AMOUNT_OUT_OF_RANGE", "amount is outside the voucher range")
	}
	eligibility, err := s.voucherEligibility(voucher, q, now)
	if err != nil {
		return Selection{}, err
	}
	if !eligibility.Available {
		return Selection{}, domainError("VOUCHER_NOT_AVAILABLE", eligibility.Reason)
	}

	var tier model.GameDepositVoucherRewards
	tierDB := s.db.Model(&model.GameDepositVoucherRewards{}).
		Where("deleted_at IS NULL AND status = 1 AND voucher_id = ? AND amount = ?", voucher.ID, amount)
	if err := tierDB.First(&tier).Error; err != nil {
		return Selection{}, notFound(err, "VOUCHER_REWARD_TIER_NOT_FOUND")
	}
	selection.Voucher = &voucher
	selection.RewardTier = &tier
	selection.BonusReward = valueOrZero(tier.BonusReward)
	selection.CashReward = valueOrZero(tier.CashReward)
	selection.RewardAmount = valueOrZero(tier.RewardAmount)
	return selection, nil
}

func (s *Service) voucherEligibility(voucher model.GameDepositVouchers, q ConfigQuery, now time.Time) (voucherEligibility, error) {
	if !listContains(voucher.Countries, q.Country) {
		return voucherEligibility{Reason: "country is not eligible"}, nil
	}
	if !listContains(voucher.Currencies, q.Currency) {
		return voucherEligibility{Reason: "currency is not eligible"}, nil
	}
	if !listContains(voucher.MethodCodes, q.MethodCode) {
		return voucherEligibility{Reason: "deposit method is not eligible"}, nil
	}
	if strings.TrimSpace(q.Amount) != "" {
		amount, err := money(q.Amount, "amount")
		if err != nil {
			return voucherEligibility{}, err
		}
		if !amountInRange(amount, voucher.MinAmount, voucher.MaxAmount) {
			return voucherEligibility{Reason: "amount is outside the voucher range"}, nil
		}
	}
	if q.UserID == 0 {
		if voucher.NewUsersOnly || voucher.FirstDepositOnly {
			return voucherEligibility{Reason: "login is required"}, nil
		}
		return voucherEligibility{Available: true}, nil
	}

	var successfulDeposits int64
	if voucher.NewUsersOnly || voucher.FirstDepositOnly {
		depositDB := tenancy.ApplySite(s.db.Model(&model.GameDepositOrders{}), "", q.Scope).
			Where("user_id = ? AND status = 2 AND deleted_at IS NULL", q.UserID)
		if err := depositDB.Count(&successfulDeposits).Error; err != nil {
			return voucherEligibility{}, err
		}
		if successfulDeposits > 0 {
			return voucherEligibility{Reason: "voucher is limited to users without a successful deposit"}, nil
		}
	}

	if voucher.PerUserLimit > 0 {
		var count int64
		depositDB := tenancy.ApplySite(s.db.Model(&model.GameDepositOrders{}), "", q.Scope).
			Where("user_id = ? AND voucher_id = ? AND status = 2 AND deleted_at IS NULL", q.UserID, voucher.ID)
		if err := depositDB.Count(&count).Error; err != nil {
			return voucherEligibility{}, err
		}
		if count >= int64(voucher.PerUserLimit) {
			return voucherEligibility{Reason: "per-user voucher limit reached"}, nil
		}
	}
	if voucher.DailyLimit > 0 {
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		var count int64
		depositDB := tenancy.ApplySite(s.db.Model(&model.GameDepositOrders{}), "", q.Scope).
			Where("user_id = ? AND voucher_id = ? AND status = 2 AND created_at >= ? AND deleted_at IS NULL", q.UserID, voucher.ID, start)
		if err := depositDB.Count(&count).Error; err != nil {
			return voucherEligibility{}, err
		}
		if count >= int64(voucher.DailyLimit) {
			return voucherEligibility{Reason: "daily voucher limit reached"}, nil
		}
	}
	if voucher.WeeklyLimit > 0 {
		weekday := (int(now.Weekday()) + 6) % 7
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -weekday)
		var count int64
		depositDB := tenancy.ApplySite(s.db.Model(&model.GameDepositOrders{}), "", q.Scope).
			Where("user_id = ? AND voucher_id = ? AND status = 2 AND created_at >= ? AND deleted_at IS NULL", q.UserID, voucher.ID, start)
		if err := depositDB.Count(&count).Error; err != nil {
			return voucherEligibility{}, err
		}
		if count >= int64(voucher.WeeklyLimit) {
			return voucherEligibility{Reason: "weekly voucher limit reached"}, nil
		}
	}
	return voucherEligibility{Available: true}, nil
}

func valueOrZero(value *decimal.Decimal) decimal.Decimal {
	if value == nil {
		return decimal.Zero
	}
	return *value
}
