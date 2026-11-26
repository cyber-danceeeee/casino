package fundadmin

import (
	"encoding/json"
	"strings"

	"gorm.io/gorm"

	"casino/internal/content"
)

func addTextFilter(db *gorm.DB, column string, value string) *gorm.DB {
	if value == "" {
		return db
	}
	return db.Where(column+" LIKE ?", "%"+value+"%")
}

func addExactFilter(db *gorm.DB, column string, value string) *gorm.DB {
	if value == "" {
		return db
	}
	return db.Where(column+" = ?", value)
}

func addTimeRange(db *gorm.DB, column string, from string, to string) *gorm.DB {
	if from != "" {
		db = db.Where(column+" >= ?", from)
	}
	if to != "" {
		db = db.Where(column+" <= ?", to)
	}
	return db
}

func addMoneyRange(db *gorm.DB, column string, min string, max string) *gorm.DB {
	if min != "" {
		db = db.Where(column+" >= ?", min)
	}
	if max != "" {
		db = db.Where(column+" <= ?", max)
	}
	return db
}

func addUintFilter(db *gorm.DB, column string, value string) *gorm.DB {
	if strings.TrimSpace(value) == "" {
		return db
	}
	return db.Where(column+" = ?", strings.TrimSpace(value))
}

func phoneUsernameSQL(alias string) string {
	return "CASE WHEN NULLIF(" + alias + ".phone,'') IS NOT NULL THEN TRIM(CONCAT_WS(' ',NULLIF(" + alias + ".phone_country_code,'')," + alias + ".phone)) ELSE COALESCE(" + alias + ".username,'') END"
}

func (s *Service) ListWallets(input ListQuery) (Page[WalletRow], error) {
	q := normalizeQuery(input)
	db := s.db.Table("game_user_wallets w").
		Joins("JOIN game_users u ON u.id = w.user_id AND u.deleted_at IS NULL").
		Joins("LEFT JOIN game_user_risk_profiles rp ON rp.user_id = u.id AND rp.deleted_at IS NULL").
		Where("w.deleted_at IS NULL")
	db = content.ApplySite(db, "w", q.Scope)
	db = addTextFilter(db, "u.user_no", queryValue(q, "userNo"))
	if username := queryValue(q, "username"); username != "" {
		like := "%" + username + "%"
		db = db.Where("u.phone LIKE ? OR u.username LIKE ?", like, like)
	}
	db = addTextFilter(db, "u.phone", queryValue(q, "phone"))
	db = addUintFilter(db, "w.id", queryValue(q, "walletID"))
	db = addExactFilter(db, "w.currency", queryValue(q, "currency"))
	db = addExactFilter(db, "u.country_code", queryValue(q, "country"))
	db = addExactFilter(db, "rp.risk_status", queryValue(q, "riskStatus"))
	if value := queryValue(q, "vipLevel"); value != "" {
		db = db.Where("u.vip_level = ?", value)
	}
	if value := strings.ToUpper(queryValue(q, "status")); value != "" {
		status := 1
		if value == "FROZEN" {
			status = 2
		}
		db = db.Where("w.status = ?", status)
	}
	db = addMoneyRange(db, "w.cash_balance", queryValue(q, "availableBalanceMin"), queryValue(q, "availableBalanceMax"))
	db = addMoneyRange(db, "w.frozen_balance", queryValue(q, "frozenBalanceMin"), queryValue(q, "frozenBalanceMax"))
	db = addTimeRange(db, "w.created_at", queryValue(q, "createdFrom"), queryValue(q, "createdTo"))
	db = addTimeRange(db, "w.updated_at", queryValue(q, "updatedFrom"), queryValue(q, "updatedTo"))
	var total int64
	if err := db.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return Page[WalletRow]{}, wrapInternal(err)
	}
	rows := []WalletRow{}
	err := db.Select(`w.id, w.user_id, u.user_no, u.username, u.phone_country_code, u.phone, u.country_code, u.vip_level,
		w.currency, CAST(w.cash_balance AS CHAR) cash_balance, CAST(w.frozen_balance AS CHAR) frozen_balance,
		CAST(w.bonus_balance AS CHAR) bonus_balance, CAST(w.withdrawable_balance AS CHAR) withdrawable_balance,
		w.status, u.status account_status, COALESCE(rp.risk_status, 'CLEARED') risk_status,
		DATE_FORMAT(w.last_transaction_at, '%Y-%m-%d %H:%i:%s') last_transaction_at,
		DATE_FORMAT(w.created_at, '%Y-%m-%d %H:%i:%s') created_at,
		DATE_FORMAT(w.updated_at, '%Y-%m-%d %H:%i:%s') updated_at`).
		Order("w.updated_at DESC, w.id DESC").Offset(pageOffset(q)).Limit(q.PageSize).Scan(&rows).Error
	if err != nil {
		return Page[WalletRow]{}, wrapInternal(err)
	}
	for i := range rows {
		rows[i].Status = walletStatus(rows[i].StatusValue)
		rows[i].AccountStatus = accountStatus(rows[i].AccountStatusValue)
		rows[i].KYCStatus = "NOT_SUBMITTED"
	}
	return Page[WalletRow]{List: rows, Total: total, Page: q.Page, PageSize: q.PageSize}, nil
}

func (s *Service) GetWallet(id uint64, scope content.Scope) (map[string]any, error) {
	var wallet WalletRow
	db := s.db.Table("game_user_wallets w").
		Joins("JOIN game_users u ON u.id = w.user_id AND u.deleted_at IS NULL").
		Joins("LEFT JOIN game_user_risk_profiles rp ON rp.user_id = u.id AND rp.deleted_at IS NULL").
		Select(`w.id, w.user_id, u.user_no, u.username, u.phone_country_code, u.phone, u.country_code, u.vip_level, w.currency,
		CAST(w.cash_balance AS CHAR) cash_balance, CAST(w.frozen_balance AS CHAR) frozen_balance,
		CAST(w.bonus_balance AS CHAR) bonus_balance, CAST(w.withdrawable_balance AS CHAR) withdrawable_balance,
		w.status, u.status account_status, COALESCE(rp.risk_status, 'CLEARED') risk_status,
		DATE_FORMAT(w.last_transaction_at, '%Y-%m-%d %H:%i:%s') last_transaction_at,
		DATE_FORMAT(w.created_at, '%Y-%m-%d %H:%i:%s') created_at,
		DATE_FORMAT(w.updated_at, '%Y-%m-%d %H:%i:%s') updated_at`).
		Where("w.id = ? AND w.deleted_at IS NULL", id)
	err := content.ApplySite(db, "w", scope).Take(&wallet).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrWalletNotFound
		}
		return nil, wrapInternal(err)
	}
	wallet.Status = walletStatus(wallet.StatusValue)
	wallet.AccountStatus = accountStatus(wallet.AccountStatusValue)
	wallet.KYCStatus = "NOT_SUBMITTED"
	balanceComposition := map[string]any{
		"availableBalance":    wallet.CashBalance,
		"frozenBalance":       wallet.FrozenBalance,
		"bonusBalance":        wallet.BonusBalance,
		"withdrawableBalance": wallet.WithdrawableBalance,
	}
	var pending map[string]any
	if err := s.db.Raw(`SELECT
		CAST(COALESCE((SELECT SUM(amount) FROM game_withdraw_orders WHERE user_id = ? AND currency = ? AND status IN (1,6,7,8,9) AND deleted_at IS NULL), 0) AS CHAR) pending_withdrawal_amount,
		CAST(COALESCE((SELECT SUM(amount) FROM game_deposit_orders WHERE user_id = ? AND currency = ? AND status IN (1,6) AND deleted_at IS NULL), 0) AS CHAR) pending_deposit_amount`, wallet.UserID, wallet.Currency, wallet.UserID, wallet.Currency).Scan(&pending).Error; err != nil {
		return nil, wrapInternal(err)
	}
	for key, value := range pending {
		balanceComposition[key] = value
	}
	available, _ := ParseMoney(wallet.CashBalance)
	frozen, _ := ParseMoney(wallet.FrozenBalance)
	bonus, _ := ParseMoney(wallet.BonusBalance)
	balanceComposition["totalBalance"] = available.Add(frozen).Add(bonus).StringFixed(2)
	recentTransactions := []map[string]any{}
	recentDB := s.db.Table("game_wallet_transactions").Select(`txn_no transaction_id, category transaction_type,
		CAST(amount AS CHAR) amount, currency, CAST(balance_before AS CHAR) balance_before,
		CAST(balance_after AS CHAR) balance_after, source_type reference_type, order_no reference_id,
		status, DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') created_at`).
		Where("wallet_id = ? AND deleted_at IS NULL", id)
	if err := content.ApplySite(recentDB, "", scope).Order("created_at DESC, id DESC").Limit(20).Scan(&recentTransactions).Error; err != nil {
		return nil, wrapInternal(err)
	}
	freezes := []map[string]any{}
	freezeDB := s.db.Table("game_wallet_freezes").Select(`id, freeze_no, CAST(amount AS CHAR) amount, currency,
		reason, reference_type, reference_no, status, DATE_FORMAT(frozen_at, '%Y-%m-%d %H:%i:%s') frozen_at,
		DATE_FORMAT(released_at, '%Y-%m-%d %H:%i:%s') released_at, operator_username operator`).
		Where("wallet_id = ? AND deleted_at IS NULL", id)
	if err := content.ApplySite(freezeDB, "", scope).Order("created_at DESC").Limit(100).Scan(&freezes).Error; err != nil {
		return nil, wrapInternal(err)
	}
	relatedUser := map[string]any{}
	userDB := s.db.Table("game_users").Select(`id, user_no, username, country_code, vip_level,
		status, DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') registered_at, DATE_FORMAT(last_login_at, '%Y-%m-%d %H:%i:%s') last_login_at`).
		Where("id = ? AND deleted_at IS NULL", wallet.UserID)
	if err := content.ApplySite(userDB, "", scope).Scan(&relatedUser).Error; err != nil {
		return nil, wrapInternal(err)
	}
	logs, err := s.operationLogs("wallet", id, scope)
	if err != nil {
		return nil, err
	}
	return map[string]any{"wallet": wallet, "balanceComposition": balanceComposition, "recentTransactions": recentTransactions, "freezes": freezes, "relatedUser": relatedUser, "operationLogs": logs}, nil
}

func (s *Service) ListTransactions(input ListQuery) (Page[TransactionRow], error) {
	q := normalizeQuery(input)
	db := s.db.Table("game_wallet_transactions t").
		Joins("JOIN game_users u ON u.id = t.user_id AND u.deleted_at IS NULL").Where("t.deleted_at IS NULL")
	db = content.ApplySite(db, "t", q.Scope)
	for _, item := range []struct{ key, column string }{
		{"transactionID", "t.txn_no"}, {"externalTransactionID", "t.external_txn_no"},
		{"userNo", "u.user_no"}, {"phone", "u.phone"}, {"referenceNo", "t.order_no"},
	} {
		db = addTextFilter(db, item.column, queryValue(q, item.key))
	}
	if username := queryValue(q, "username"); username != "" {
		like := "%" + username + "%"
		db = db.Where("u.phone LIKE ? OR u.username LIKE ?", like, like)
	}
	db = addUintFilter(db, "t.wallet_id", queryValue(q, "walletID"))
	for _, item := range []struct{ key, column string }{
		{"transactionType", "t.category"}, {"direction", "t.direction"}, {"currency", "t.currency"},
		{"referenceType", "t.source_type"}, {"providerCode", "t.provider_code"},
	} {
		db = addExactFilter(db, item.column, queryValue(q, item.key))
	}
	if value := strings.ToUpper(queryValue(q, "status")); value != "" {
		status := 1
		if value != "SUCCESS" && value != "COMPLETED" {
			status = 2
		}
		db = db.Where("t.status = ?", status)
	}
	db = addTextFilter(db, "t.operator_username", queryValue(q, "operator"))
	db = addMoneyRange(db, "t.amount", queryValue(q, "amountMin"), queryValue(q, "amountMax"))
	db = addTimeRange(db, "t.created_at", queryValue(q, "createdFrom"), queryValue(q, "createdTo"))
	db = addTimeRange(db, "t.completed_at", queryValue(q, "completedFrom"), queryValue(q, "completedTo"))
	var total int64
	if err := db.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return Page[TransactionRow]{}, wrapInternal(err)
	}
	rows := []TransactionRow{}
	err := db.Select(`t.id, t.txn_no, t.external_txn_no, t.user_id, u.user_no, ` + phoneUsernameSQL("u") + ` username, u.phone_country_code, u.phone, t.wallet_id,
		t.category, t.direction, CAST(t.amount AS CHAR) amount, t.currency,
		CAST(t.balance_before AS CHAR) balance_before, CAST(t.balance_after AS CHAR) balance_after,
		CAST(t.frozen_balance_before AS CHAR) frozen_balance_before, CAST(t.frozen_balance_after AS CHAR) frozen_balance_after,
		t.source_type, t.source_id, t.order_no, t.provider_code, t.status, t.operator_username,
		COALESCE(t.idempotency_key, '') idempotency_key, t.remark,
		DATE_FORMAT(t.created_at, '%Y-%m-%d %H:%i:%s') created_at,
		DATE_FORMAT(t.completed_at, '%Y-%m-%d %H:%i:%s') completed_at`).
		Order("t.created_at DESC, t.id DESC").Offset(pageOffset(q)).Limit(q.PageSize).Scan(&rows).Error
	if err != nil {
		return Page[TransactionRow]{}, wrapInternal(err)
	}
	for i := range rows {
		rows[i].Status = transactionStatus(rows[i].StatusValue)
	}
	return Page[TransactionRow]{List: rows, Total: total, Page: q.Page, PageSize: q.PageSize}, nil
}

func (s *Service) GetTransaction(id uint64, scope content.Scope) (map[string]any, error) {
	var row TransactionRow
	db := s.db.Table("game_wallet_transactions t").Joins("JOIN game_users u ON u.id=t.user_id").
		Select(`t.id, t.txn_no, t.external_txn_no, t.user_id, u.user_no, `+phoneUsernameSQL("u")+` username, u.phone_country_code, u.phone, t.wallet_id,
		t.category, t.direction, CAST(t.amount AS CHAR) amount, t.currency,
		CAST(t.balance_before AS CHAR) balance_before, CAST(t.balance_after AS CHAR) balance_after,
		CAST(t.frozen_balance_before AS CHAR) frozen_balance_before, CAST(t.frozen_balance_after AS CHAR) frozen_balance_after,
		t.source_type, t.source_id, t.order_no, t.provider_code, t.status, t.operator_username,
		COALESCE(t.idempotency_key,'') idempotency_key, t.remark, t.raw_metadata,
		DATE_FORMAT(t.created_at, '%Y-%m-%d %H:%i:%s') created_at,
		DATE_FORMAT(t.completed_at, '%Y-%m-%d %H:%i:%s') completed_at`).
		Where("t.id=? AND t.deleted_at IS NULL", id)
	err := content.ApplySite(db, "t", scope).Take(&row).Error
	if err != nil {
		return nil, wrapInternal(err)
	}
	row.Status = transactionStatus(row.StatusValue)
	row.RawMetadata = redactMetadata(row.RawMetadata)
	metadata := map[string]any{}
	_ = json.Unmarshal([]byte(row.RawMetadata), &metadata)
	row.RawMetadata = ""
	balanceChange := map[string]any{
		"balanceBefore": row.BalanceBefore, "transactionAmount": row.Amount, "balanceAfter": row.BalanceAfter,
		"frozenBalanceBefore": row.FrozenBalanceBefore, "frozenBalanceAfter": row.FrozenBalanceAfter,
	}
	relatedBusiness := map[string]any{"referenceType": row.ReferenceType, "referenceID": row.ReferenceID, "referenceNo": row.ReferenceNo, "providerCode": row.ProviderCode}
	logs, err := s.operationLogs("wallet_transaction", id, scope)
	if err != nil {
		return nil, err
	}
	return map[string]any{"transaction": row, "balanceChange": balanceChange, "relatedBusiness": relatedBusiness, "rawMetadata": metadata, "operationLogs": logs}, nil
}

func redactMetadata(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return "{}"
	}
	value := map[string]any{}
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return "{}"
	}
	redactMap(value)
	data, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func redactMap(value map[string]any) {
	for key, item := range value {
		normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(key, "_", ""), "-", ""))
		switch normalized {
		case "token", "accesstoken", "refreshtoken", "password", "apipassword", "signature", "sign", "apikey", "secret", "salt", "saltkey", "launchurl", "accountnumber", "bankaccount":
			value[key] = "[REDACTED]"
			continue
		}
		switch nested := item.(type) {
		case map[string]any:
			redactMap(nested)
		case []any:
			for _, child := range nested {
				if object, ok := child.(map[string]any); ok {
					redactMap(object)
				}
			}
		}
	}
}

func (s *Service) operationLogs(targetType string, targetID uint64, scope content.Scope) ([]map[string]any, error) {
	rows := []map[string]any{}
	db := s.db.Table("game_admin_operation_logs").Select(`operator_username operator, permission, action,
		before_state, after_state, CAST(amount AS CHAR) amount, currency, reason, ip_address,
		correlation_id, DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') created_at`).
		Where("target_type = ? AND target_id = ? AND deleted_at IS NULL", targetType, targetID)
	err := content.ApplySite(db, "", scope).Order("created_at DESC, id DESC").Limit(100).Scan(&rows).Error
	if err != nil {
		return nil, wrapInternal(err)
	}
	return rows, nil
}

func (s *Service) walletTransactionResult(id uint64, scope content.Scope) (map[string]any, error) {
	if id == 0 {
		return map[string]any{}, nil
	}
	row := map[string]any{}
	db := s.db.Table("game_wallet_transactions").Select(`txn_no transaction_id, category transaction_type,
		direction, CAST(amount AS CHAR) amount, currency, CAST(balance_before AS CHAR) balance_before,
		CAST(balance_after AS CHAR) balance_after, CAST(frozen_balance_before AS CHAR) frozen_balance_before,
		CAST(frozen_balance_after AS CHAR) frozen_balance_after, source_type reference_type, order_no reference_id,
		status, DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') created_at,
		DATE_FORMAT(completed_at, '%Y-%m-%d %H:%i:%s') completed_at`).
		Where("id = ? AND deleted_at IS NULL", id)
	err := content.ApplySite(db, "", scope).Scan(&row).Error
	if err != nil {
		return nil, wrapInternal(err)
	}
	return row, nil
}

func (s *Service) ListAdjustments(input ListQuery) (Page[AdjustmentRow], error) {
	q := normalizeQuery(input)
	db := s.db.Table("game_manual_adjustments a").Joins("JOIN game_users u ON u.id=a.user_id AND u.deleted_at IS NULL").Where("a.deleted_at IS NULL")
	db = content.ApplySite(db, "a", q.Scope)
	for _, item := range []struct{ key, column string }{{"adjustmentID", "a.adjustment_no"}, {"userNo", "u.user_no"}, {"phone", "u.phone"}, {"applicant", "a.applicant_username"}, {"approver", "a.approver_username"}} {
		db = addTextFilter(db, item.column, queryValue(q, item.key))
	}
	if username := queryValue(q, "username"); username != "" {
		like := "%" + username + "%"
		db = db.Where("u.phone LIKE ? OR u.username LIKE ?", like, like)
	}
	db = addUintFilter(db, "a.wallet_id", queryValue(q, "walletID"))
	for _, item := range []struct{ key, column string }{{"adjustmentType", "a.adjustment_type"}, {"currency", "a.currency"}, {"reasonCode", "a.reason_code"}, {"status", "a.status"}} {
		db = addExactFilter(db, item.column, queryValue(q, item.key))
	}
	db = addMoneyRange(db, "a.amount", queryValue(q, "amountMin"), queryValue(q, "amountMax"))
	db = addTimeRange(db, "a.created_at", queryValue(q, "createdFrom"), queryValue(q, "createdTo"))
	db = addTimeRange(db, "a.completed_at", queryValue(q, "completedFrom"), queryValue(q, "completedTo"))
	return scanAdjustmentPage(db, q)
}

func scanAdjustmentPage(db *gorm.DB, q ListQuery) (Page[AdjustmentRow], error) {
	var total int64
	if err := db.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return Page[AdjustmentRow]{}, wrapInternal(err)
	}
	rows := []AdjustmentRow{}
	err := db.Select(`a.id, a.adjustment_no, a.user_id, u.user_no, ` + phoneUsernameSQL("u") + ` username, u.phone_country_code, u.phone, a.wallet_id, a.adjustment_type,
		CAST(a.amount AS CHAR) amount, a.currency, a.reason_code, a.detailed_reason, a.supporting_reference,
		a.internal_remark, a.status, CAST(a.requested_balance AS CHAR) requested_balance,
		CAST(a.balance_before AS CHAR) balance_before, CAST(a.balance_after AS CHAR) balance_after,
		a.applicant_username, a.applicant_id, a.approver_username, a.approval_reason,
		a.wallet_transaction_id, a.failure_reason,
		DATE_FORMAT(a.created_at, '%Y-%m-%d %H:%i:%s') created_at,
		DATE_FORMAT(a.approved_at, '%Y-%m-%d %H:%i:%s') approved_at,
		DATE_FORMAT(a.completed_at, '%Y-%m-%d %H:%i:%s') completed_at`).
		Order("a.created_at DESC, a.id DESC").Offset(pageOffset(q)).Limit(q.PageSize).Scan(&rows).Error
	if err != nil {
		return Page[AdjustmentRow]{}, wrapInternal(err)
	}
	return Page[AdjustmentRow]{List: rows, Total: total, Page: q.Page, PageSize: q.PageSize}, nil
}

func (s *Service) GetAdjustment(id uint64, scope content.Scope) (map[string]any, error) {
	var row AdjustmentRow
	db := s.db.Table("game_manual_adjustments a").Joins("JOIN game_users u ON u.id=a.user_id").
		Select(`a.id, a.adjustment_no, a.user_id, u.user_no, `+phoneUsernameSQL("u")+` username, u.phone_country_code, u.phone, a.wallet_id, a.adjustment_type,
		CAST(a.amount AS CHAR) amount, a.currency, a.reason_code, a.detailed_reason, a.supporting_reference,
		a.internal_remark, a.status, CAST(a.requested_balance AS CHAR) requested_balance,
		CAST(a.balance_before AS CHAR) balance_before, CAST(a.balance_after AS CHAR) balance_after,
		a.applicant_username, a.applicant_id, a.approver_username, a.approval_reason,
		a.wallet_transaction_id, a.failure_reason, DATE_FORMAT(a.created_at,'%Y-%m-%d %H:%i:%s') created_at,
		DATE_FORMAT(a.approved_at,'%Y-%m-%d %H:%i:%s') approved_at,
		DATE_FORMAT(a.completed_at,'%Y-%m-%d %H:%i:%s') completed_at`).
		Where("a.id=? AND a.deleted_at IS NULL", id)
	err := content.ApplySite(db, "a", scope).Take(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAdjustmentNotFound
		}
		return nil, wrapInternal(err)
	}
	approvals := []map[string]any{}
	approvalDB := s.db.Table("game_manual_adjustment_approvals").Select(`operator_username operator, decision,
		before_status, after_status, reason, DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') created_at`).
		Where("adjustment_id=? AND deleted_at IS NULL", id)
	if err := content.ApplySite(approvalDB, "", scope).Order("created_at").Scan(&approvals).Error; err != nil {
		return nil, wrapInternal(err)
	}
	walletResult, err := s.walletTransactionResult(row.WalletTransactionID, scope)
	if err != nil {
		return nil, err
	}
	logs, err := s.operationLogs("manual_adjustment", id, scope)
	if err != nil {
		return nil, err
	}
	return map[string]any{"adjustment": row, "approvals": approvals, "walletResult": walletResult, "operationLogs": logs}, nil
}

func (s *Service) ListDeposits(input ListQuery) (Page[DepositRow], error) {
	q := normalizeQuery(input)
	db := s.db.Table("game_deposit_orders o").Joins("JOIN game_users u ON u.id=o.user_id AND u.deleted_at IS NULL").Where("o.deleted_at IS NULL")
	db = content.ApplySite(db, "o", q.Scope)
	for _, item := range []struct{ key, column string }{{"platformOrderID", "o.order_no"}, {"channelOrderID", "o.channel_order_no"}, {"userNo", "u.user_no"}, {"username", "u.username"}, {"phone", "u.phone"}} {
		db = addTextFilter(db, item.column, queryValue(q, item.key))
	}
	for _, item := range []struct{ key, column string }{{"paymentChannel", "o.channel_code"}, {"paymentMethod", "o.method_code"}, {"currency", "o.currency"}} {
		db = addExactFilter(db, item.column, queryValue(q, item.key))
	}
	if value := queryValue(q, "status"); value != "" {
		db = db.Where("o.status=?", depositStatusValue(value))
	}
	db = addMoneyRange(db, "o.amount", queryValue(q, "amountMin"), queryValue(q, "amountMax"))
	db = addTimeRange(db, "o.created_at", queryValue(q, "createdFrom"), queryValue(q, "createdTo"))
	db = addTimeRange(db, "o.paid_at", queryValue(q, "paidFrom"), queryValue(q, "paidTo"))
	db = addTimeRange(db, "o.callback_at", queryValue(q, "callbackFrom"), queryValue(q, "callbackTo"))
	var total int64
	if err := db.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return Page[DepositRow]{}, wrapInternal(err)
	}
	rows := []DepositRow{}
	err := db.Select(`o.id,o.order_no,o.channel_order_no,o.user_id,u.user_no,u.username,o.channel_code,o.method_code,
		CAST(o.amount AS CHAR) amount,CAST(o.fee_amount AS CHAR) fee_amount,CAST(o.paid_amount AS CHAR) paid_amount,o.currency,o.status,
		o.wallet_transaction_id,o.reviewer_username,o.review_reason,o.evidence_reference,o.channel_response_code,
		o.channel_response_message,o.fail_reason,DATE_FORMAT(o.created_at,'%Y-%m-%d %H:%i:%s') created_at,
		DATE_FORMAT(o.expired_at,'%Y-%m-%d %H:%i:%s') expired_at,DATE_FORMAT(o.paid_at,'%Y-%m-%d %H:%i:%s') paid_at,
		DATE_FORMAT(o.callback_at,'%Y-%m-%d %H:%i:%s') callback_at,DATE_FORMAT(o.completed_at,'%Y-%m-%d %H:%i:%s') completed_at`).
		Order("o.created_at DESC,o.id DESC").Offset(pageOffset(q)).Limit(q.PageSize).Scan(&rows).Error
	if err != nil {
		return Page[DepositRow]{}, wrapInternal(err)
	}
	for i := range rows {
		rows[i].Status = depositStatus(rows[i].StatusValue)
	}
	return Page[DepositRow]{List: rows, Total: total, Page: q.Page, PageSize: q.PageSize}, nil
}

func depositStatusValue(value string) int {
	switch strings.ToUpper(value) {
	case "SUCCESS":
		return 2
	case "FAILED":
		return 3
	case "CANCELLED":
		return 4
	case "EXPIRED":
		return 5
	case "MANUAL_REVIEW":
		return 6
	default:
		return 1
	}
}

func (s *Service) GetDeposit(id uint64, scope content.Scope) (map[string]any, error) {
	var row DepositRow
	db := s.db.Table("game_deposit_orders o").Joins("JOIN game_users u ON u.id=o.user_id").Select(`o.id,o.order_no,o.channel_order_no,o.user_id,u.user_no,u.username,o.channel_code,o.method_code,CAST(o.amount AS CHAR) amount,CAST(o.fee_amount AS CHAR) fee_amount,CAST(o.paid_amount AS CHAR) paid_amount,o.currency,o.status,o.wallet_transaction_id,o.reviewer_username,o.review_reason,o.evidence_reference,o.channel_response_code,o.channel_response_message,o.fail_reason,DATE_FORMAT(o.created_at,'%Y-%m-%d %H:%i:%s') created_at,DATE_FORMAT(o.expired_at,'%Y-%m-%d %H:%i:%s') expired_at,DATE_FORMAT(o.paid_at,'%Y-%m-%d %H:%i:%s') paid_at,DATE_FORMAT(o.callback_at,'%Y-%m-%d %H:%i:%s') callback_at,DATE_FORMAT(o.completed_at,'%Y-%m-%d %H:%i:%s') completed_at`).Where("o.id=? AND o.deleted_at IS NULL", id)
	err := content.ApplySite(db, "o", scope).Take(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrDepositNotFound
		}
		return nil, wrapInternal(err)
	}
	row.Status = depositStatus(row.StatusValue)
	paymentInformation := map[string]any{
		"paymentChannel": row.ChannelCode, "paymentMethod": row.MethodCode,
		"channelOrderID": row.ChannelOrderNo, "channelResponseCode": row.ChannelResponseCode,
		"channelResponseMessage": row.ChannelResponseMessage,
	}
	walletCredit, err := s.walletTransactionResult(row.WalletTransactionID, scope)
	if err != nil {
		return nil, err
	}
	callbacks := []map[string]any{}
	callbackDB := s.db.Table("game_deposit_callback_logs").Select(`DATE_FORMAT(callback_time, '%Y-%m-%d %H:%i:%s') callback_time,
		source_ip, signature_valid, channel_status, processing_result,
		CASE WHEN http_response IS NULL OR http_response = '' THEN '' ELSE '[REDACTED]' END http_response, retry_count`).
		Where("deposit_order_id=? AND deleted_at IS NULL", id)
	if err := content.ApplySite(callbackDB, "", scope).Order("callback_time DESC").Scan(&callbacks).Error; err != nil {
		return nil, wrapInternal(err)
	}
	logs, err := s.operationLogs("deposit", id, scope)
	if err != nil {
		return nil, err
	}
	return map[string]any{"order": row, "paymentInformation": paymentInformation, "walletCredit": walletCredit, "callbacks": callbacks, "operationLogs": logs}, nil
}

func (s *Service) ListWithdrawals(input ListQuery) (Page[WithdrawalRow], error) {
	q := normalizeQuery(input)
	db := s.db.Table("game_withdraw_orders o").Joins("JOIN game_users u ON u.id=o.user_id AND u.deleted_at IS NULL").Joins("LEFT JOIN game_withdraw_accounts a ON a.id=o.withdraw_account_id AND a.deleted_at IS NULL").Where("o.deleted_at IS NULL")
	db = content.ApplySite(db, "o", q.Scope)
	for _, item := range []struct{ key, column string }{{"platformOrderID", "o.order_no"}, {"channelOrderID", "o.channel_order_no"}, {"userNo", "u.user_no"}, {"username", "u.username"}, {"phone", "u.phone"}} {
		db = addTextFilter(db, item.column, queryValue(q, item.key))
	}
	for _, item := range []struct{ key, column string }{{"currency", "o.currency"}, {"paymentMethod", "o.payment_method"}, {"riskStatus", "o.risk_status"}, {"kycStatus", "o.kyc_status"}, {"reviewer", "o.reviewer_username"}} {
		db = addExactFilter(db, item.column, queryValue(q, item.key))
	}
	if value := queryValue(q, "status"); value != "" {
		db = db.Where("o.status=?", withdrawalStatusValue(value))
	}
	db = addMoneyRange(db, "o.amount", queryValue(q, "amountMin"), queryValue(q, "amountMax"))
	db = addTimeRange(db, "o.created_at", queryValue(q, "createdFrom"), queryValue(q, "createdTo"))
	db = addTimeRange(db, "o.reviewed_at", queryValue(q, "reviewedFrom"), queryValue(q, "reviewedTo"))
	db = addTimeRange(db, "o.completed_at", queryValue(q, "completedFrom"), queryValue(q, "completedTo"))
	var total int64
	if err := db.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return Page[WithdrawalRow]{}, wrapInternal(err)
	}
	rows := []WithdrawalRow{}
	err := db.Select(`o.id,o.order_no,o.channel_order_no,o.user_id,u.user_no,u.username,CAST(o.amount AS CHAR) amount,CAST(o.fee_amount AS CHAR) fee_amount,CAST(o.actual_amount AS CHAR) actual_amount,o.currency,o.payment_method,
		CASE WHEN a.account_no IS NULL OR a.account_no='' THEN '' WHEN CHAR_LENGTH(a.account_no)<=4 THEN a.account_no ELSE CONCAT('****',RIGHT(a.account_no,4)) END account_masked,
		CASE WHEN a.account_name IS NULL OR a.account_name='' THEN '' WHEN CHAR_LENGTH(a.account_name)<=2 THEN CONCAT(LEFT(a.account_name,1),'*') ELSE CONCAT(LEFT(a.account_name,1),'***',RIGHT(a.account_name,1)) END account_name_masked,
		COALESCE(a.bank_name,'') bank_name,o.kyc_status,o.risk_status,o.status,o.audit_status,o.reviewer_username,o.review_remark,o.freeze_record_id,o.final_debit_transaction_id,o.refund_transaction_id,o.failure_reason,DATE_FORMAT(o.created_at,'%Y-%m-%d %H:%i:%s') created_at,DATE_FORMAT(o.submitted_at,'%Y-%m-%d %H:%i:%s') submitted_at,DATE_FORMAT(o.reviewed_at,'%Y-%m-%d %H:%i:%s') reviewed_at,DATE_FORMAT(o.approved_at,'%Y-%m-%d %H:%i:%s') approved_at,DATE_FORMAT(o.completed_at,'%Y-%m-%d %H:%i:%s') completed_at`).Order("o.created_at DESC,o.id DESC").Offset(pageOffset(q)).Limit(q.PageSize).Scan(&rows).Error
	if err != nil {
		return Page[WithdrawalRow]{}, wrapInternal(err)
	}
	for i := range rows {
		rows[i].Status = withdrawalStatus(rows[i].StatusValue)
		rows[i].AuditStatus = auditStatus(rows[i].AuditStatusValue)
	}
	return Page[WithdrawalRow]{List: rows, Total: total, Page: q.Page, PageSize: q.PageSize}, nil
}

func withdrawalStatusValue(value string) int {
	switch strings.ToUpper(value) {
	case "SUCCESS":
		return 2
	case "FAILED":
		return 3
	case "REJECTED":
		return 4
	case "CANCELLED":
		return 5
	case "RISK_REVIEW":
		return 6
	case "FINANCE_REVIEW":
		return 7
	case "APPROVED":
		return 8
	case "PROCESSING":
		return 9
	default:
		return 1
	}
}
func auditStatus(value int) string {
	switch value {
	case 2:
		return "APPROVED"
	case 3:
		return "REJECTED"
	default:
		return "PENDING"
	}
}

func (s *Service) GetWithdrawal(id uint64, scope content.Scope) (map[string]any, error) {
	var row WithdrawalRow
	db := s.db.Table("game_withdraw_orders o").Joins("JOIN game_users u ON u.id=o.user_id").Joins("LEFT JOIN game_withdraw_accounts a ON a.id=o.withdraw_account_id AND a.deleted_at IS NULL").
		Select(`o.id,o.order_no,o.channel_order_no,o.user_id,u.user_no,u.username,CAST(o.amount AS CHAR) amount,CAST(o.fee_amount AS CHAR) fee_amount,CAST(o.actual_amount AS CHAR) actual_amount,o.currency,o.payment_method,
		CASE WHEN a.account_no IS NULL OR a.account_no='' THEN '' WHEN CHAR_LENGTH(a.account_no)<=4 THEN a.account_no ELSE CONCAT('****',RIGHT(a.account_no,4)) END account_masked,
		CASE WHEN a.account_name IS NULL OR a.account_name='' THEN '' WHEN CHAR_LENGTH(a.account_name)<=2 THEN CONCAT(LEFT(a.account_name,1),'*') ELSE CONCAT(LEFT(a.account_name,1),'***',RIGHT(a.account_name,1)) END account_name_masked,
		COALESCE(a.bank_name,'') bank_name,o.kyc_status,o.risk_status,o.status,o.audit_status,o.reviewer_username,o.review_remark,o.freeze_record_id,o.final_debit_transaction_id,o.refund_transaction_id,o.failure_reason,DATE_FORMAT(o.created_at,'%Y-%m-%d %H:%i:%s') created_at,DATE_FORMAT(o.submitted_at,'%Y-%m-%d %H:%i:%s') submitted_at,DATE_FORMAT(o.reviewed_at,'%Y-%m-%d %H:%i:%s') reviewed_at,DATE_FORMAT(o.approved_at,'%Y-%m-%d %H:%i:%s') approved_at,DATE_FORMAT(o.completed_at,'%Y-%m-%d %H:%i:%s') completed_at`).
		Where("o.id=? AND o.deleted_at IS NULL", id)
	err := content.ApplySite(db, "o", scope).Take(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrWithdrawalNotFound
		}
		return nil, wrapInternal(err)
	}
	row.Status = withdrawalStatus(row.StatusValue)
	row.AuditStatus = auditStatus(row.AuditStatusValue)
	userReview := map[string]any{}
	userDB := s.db.Table("game_users u").Joins("LEFT JOIN game_user_wallets w ON w.user_id = u.id AND w.currency = ? AND w.deleted_at IS NULL", row.Currency).
		Joins("LEFT JOIN game_user_risk_profiles rp ON rp.user_id = u.id AND rp.deleted_at IS NULL").
		Select(`u.user_no, u.username, u.status account_status, u.country_code, u.vip_level,
		COALESCE(rp.risk_status, 'CLEARED') risk_status, CAST(COALESCE(w.cash_balance, 0) AS CHAR) available_balance,
		CAST(COALESCE(w.frozen_balance, 0) AS CHAR) frozen_balance,
		DATE_FORMAT(u.created_at, '%Y-%m-%d %H:%i:%s') registration_time,
		DATE_FORMAT(u.last_login_at, '%Y-%m-%d %H:%i:%s') last_login_time`).
		Where("u.id = ? AND u.deleted_at IS NULL", row.UserID)
	if err := content.ApplySite(userDB, "u", scope).Scan(&userReview).Error; err != nil {
		return nil, wrapInternal(err)
	}
	userReview["kyc_status"] = row.KYCStatus
	userReview["withdrawal_risk_status"] = row.RiskStatus
	fundsAndFreeze := map[string]any{
		"freezeRecordID": row.FreezeRecordID, "finalDebitTransactionID": row.FinalDebitTransactionID,
		"refundTransactionID": row.RefundTransactionID,
	}
	if row.FreezeRecordID != 0 {
		freeze := map[string]any{}
		freezeDB := s.db.Table("game_wallet_freezes").Select(`freeze_no, CAST(amount AS CHAR) frozen_amount, currency,
			status freeze_status, freeze_transaction_id, release_transaction_id,
			DATE_FORMAT(frozen_at, '%Y-%m-%d %H:%i:%s') freeze_time,
			DATE_FORMAT(released_at, '%Y-%m-%d %H:%i:%s') release_time`).
			Where("id = ? AND deleted_at IS NULL", row.FreezeRecordID)
		if err := content.ApplySite(freezeDB, "", scope).Scan(&freeze).Error; err != nil {
			return nil, wrapInternal(err)
		}
		for key, value := range freeze {
			fundsAndFreeze[key] = value
		}
	}
	paymentInformation := map[string]any{"paymentMethod": row.PaymentMethod, "recipientNameMasked": row.RecipientNameMasked, "recipientAccountMasked": row.RecipientAccountMasked, "bankName": row.BankName, "channelOrderID": row.ChannelOrderNo}
	approvals := []map[string]any{}
	payouts := []map[string]any{}
	approvalDB := s.db.Table("game_withdrawal_approvals").Select(`operator_username reviewer, stage, decision,
		before_status, after_status, reason, internal_remark,
		DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') created_at`).
		Where("withdrawal_order_id=? AND deleted_at IS NULL", id)
	if err := content.ApplySite(approvalDB, "", scope).Order("created_at").Scan(&approvals).Error; err != nil {
		return nil, wrapInternal(err)
	}
	payoutDB := s.db.Table("game_withdrawal_payout_logs").Select(`channel_code, request_result, channel_order_no,
		query_result, callback_result, retry_count, failure_reason,
		DATE_FORMAT(requested_at, '%Y-%m-%d %H:%i:%s') requested_at`).
		Where("withdrawal_order_id=? AND deleted_at IS NULL", id)
	if err := content.ApplySite(payoutDB, "", scope).Order("created_at DESC").Scan(&payouts).Error; err != nil {
		return nil, wrapInternal(err)
	}
	logs, err := s.operationLogs("withdrawal", id, scope)
	if err != nil {
		return nil, err
	}
	return map[string]any{"order": row, "userReview": userReview, "fundsAndFreeze": fundsAndFreeze, "paymentInformation": paymentInformation, "approvals": approvals, "payouts": payouts, "operationLogs": logs}, nil
}

func (s *Service) ListReconciliations(input ListQuery) (Page[ReconciliationRow], error) {
	q := normalizeQuery(input)
	db := s.db.Table("game_reconciliation_records r").Where("r.deleted_at IS NULL")
	db = content.ApplySite(db, "r", q.Scope)
	for _, item := range []struct{ key, column string }{{"reconciliationID", "r.reconciliation_no"}, {"businessDate", "r.business_date"}, {"businessType", "r.business_type"}, {"channel", "r.channel_code"}, {"currency", "r.currency"}, {"status", "r.status"}} {
		if item.key == "reconciliationID" {
			db = addTextFilter(db, item.column, queryValue(q, item.key))
		} else {
			db = addExactFilter(db, item.column, queryValue(q, item.key))
		}
	}
	db = addTimeRange(db, "r.generated_at", queryValue(q, "generatedFrom"), queryValue(q, "generatedTo"))
	for _, item := range []struct{ key, column string }{{"differenceType", "d.difference_type"}, {"platformOrderID", "d.platform_order_id"}, {"channelOrderID", "d.channel_order_id"}} {
		if value := queryValue(q, item.key); value != "" {
			db = db.Where("EXISTS (SELECT 1 FROM game_reconciliation_differences d WHERE d.reconciliation_id = r.id AND d.deleted_at IS NULL AND "+item.column+" LIKE ?)", "%"+value+"%")
		}
	}
	var total int64
	if err := db.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return Page[ReconciliationRow]{}, wrapInternal(err)
	}
	rows := []ReconciliationRow{}
	err := db.Select(`r.id,r.reconciliation_no,DATE_FORMAT(r.business_date,'%Y-%m-%d') business_date,r.business_type,r.channel_code,r.currency,r.reconciliation_mode,CAST(r.platform_order_count AS UNSIGNED) platform_order_count,CAST(r.platform_amount AS CHAR) platform_amount,CAST(r.channel_order_count AS UNSIGNED) channel_order_count,CAST(r.channel_amount AS CHAR) channel_amount,CAST(r.difference_amount AS CHAR) difference_amount,CAST(r.difference_count AS UNSIGNED) difference_count,r.status,r.generated_by,DATE_FORMAT(r.generated_at,'%Y-%m-%d %H:%i:%s') generated_at,DATE_FORMAT(r.completed_at,'%Y-%m-%d %H:%i:%s') completed_at`).Order("r.generated_at DESC,r.id DESC").Offset(pageOffset(q)).Limit(q.PageSize).Scan(&rows).Error
	if err != nil {
		return Page[ReconciliationRow]{}, wrapInternal(err)
	}
	return Page[ReconciliationRow]{List: rows, Total: total, Page: q.Page, PageSize: q.PageSize}, nil
}

func (s *Service) GetReconciliation(id uint64, scope content.Scope) (map[string]any, error) {
	var row ReconciliationRow
	db := s.db.Table("game_reconciliation_records r").Select(`r.id,r.reconciliation_no,DATE_FORMAT(r.business_date,'%Y-%m-%d') business_date,r.business_type,r.channel_code,r.currency,r.reconciliation_mode,r.platform_order_count,CAST(r.platform_amount AS CHAR) platform_amount,r.channel_order_count,CAST(r.channel_amount AS CHAR) channel_amount,CAST(r.difference_amount AS CHAR) difference_amount,r.difference_count,r.status,r.generated_by,DATE_FORMAT(r.generated_at,'%Y-%m-%d %H:%i:%s') generated_at,DATE_FORMAT(r.completed_at,'%Y-%m-%d %H:%i:%s') completed_at`).Where("r.id=? AND r.deleted_at IS NULL", id)
	err := content.ApplySite(db, "r", scope).Take(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrReconciliationAbsent
		}
		return nil, wrapInternal(err)
	}
	differences := []map[string]any{}
	logs := []map[string]any{}
	resolutionHistory := []map[string]any{}
	diffDB := s.db.Table("game_reconciliation_differences").Select(`id, platform_order_id, channel_order_id,
		platform_status, channel_status, CAST(platform_amount AS CHAR) platform_amount,
		CAST(channel_amount AS CHAR) channel_amount, difference_type,
		CAST(difference_amount AS CHAR) difference_amount, resolution_status,
		DATE_FORMAT(detected_at, '%Y-%m-%d %H:%i:%s') detected_at,
		DATE_FORMAT(resolved_at, '%Y-%m-%d %H:%i:%s') resolved_at`).
		Where("reconciliation_id=? AND deleted_at IS NULL", id)
	if err := content.ApplySite(diffDB, "", scope).Order("id").Scan(&differences).Error; err != nil {
		return nil, wrapInternal(err)
	}
	logDB := s.db.Table("game_reconciliation_logs").Select(`level, message,
		DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') created_at`).
		Where("reconciliation_id=? AND deleted_at IS NULL", id)
	if err := content.ApplySite(logDB, "", scope).Order("id").Scan(&logs).Error; err != nil {
		return nil, wrapInternal(err)
	}
	resolutionDB := s.db.Table("game_reconciliation_resolutions").Select(`operator_username operator, action, reason,
		DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') created_at`).
		Where("reconciliation_id=? AND deleted_at IS NULL", id)
	if err := content.ApplySite(resolutionDB, "", scope).Order("id").Scan(&resolutionHistory).Error; err != nil {
		return nil, wrapInternal(err)
	}
	auditLogs, err := s.operationLogs("reconciliation", id, scope)
	if err != nil {
		return nil, err
	}
	return map[string]any{"record": row, "matchedRecords": []map[string]any{}, "differences": differences, "logs": logs, "resolutionHistory": resolutionHistory, "operationLogs": auditLogs}, nil
}
