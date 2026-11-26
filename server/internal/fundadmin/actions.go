package fundadmin

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/bonusaudit"
	"casino/internal/eventbus"
	"casino/internal/model"
	"casino/internal/tenancy"
)

func (s *Service) IsLargeAdjustment(id uint64, scope tenancy.Scope) (bool, error) {
	var amount decimal.Decimal
	db := s.db.Table("game_manual_adjustments").Select("amount").Where("id = ? AND deleted_at IS NULL", id)
	if err := tenancy.ApplySite(db, "", scope).Scan(&amount).Error; err != nil {
		return false, wrapInternal(err)
	}
	return amount.GreaterThanOrEqual(decimal.NewFromInt(100000)), nil
}

func (s *Service) CreateWithdrawal(userID, withdrawAccountID uint64, amount decimal.Decimal, actor Actor) (model.GameWithdrawOrders, error) {
	var result model.GameWithdrawOrders
	if !amount.GreaterThan(decimal.Zero) {
		return result, ErrInvalidAmount
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var account model.GameWithdrawAccounts
		accountQuery := tx.Model(&model.GameWithdrawAccounts{}).Where("id = ? AND user_id = ? AND status = 1 AND deleted_at IS NULL", withdrawAccountID, userID)
		if err := tenancy.ApplySite(accountQuery, "", actor.Scope).Take(&account).Error; err != nil {
			return newError("WITHDRAW_ACCOUNT_NOT_FOUND", "withdrawal account not found")
		}
		var wallet model.GameUserWallets
		walletQuery := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.GameUserWallets{}).Where("user_id = ? AND deleted_at IS NULL", userID)
		if err := tenancy.ApplySite(walletQuery, "", actor.Scope).Take(&wallet).Error; err != nil {
			return ErrWalletNotFound
		}
		if wallet.Status != walletActive {
			return ErrWalletFrozen
		}
		beforeCash := money(wallet.CashBalance)
		beforeFrozen := money(wallet.FrozenBalance)
		if beforeCash.LessThan(amount) {
			return ErrInsufficientBalance
		}
		if err := ensureWithdrawalAuditComplete(tx, actor.Scope, userID, wallet.Currency); err != nil {
			return err
		}
		afterCash := beforeCash.Sub(amount)
		afterFrozen := beforeFrozen.Add(amount)
		now := time.Now()
		result = model.GameWithdrawOrders{
			TenantID: wallet.TenantID, SiteID: wallet.SiteID,
			OrderNo: uniqueNo("W"), UserID: userID, WithdrawAccountID: withdrawAccountID,
			PaymentMethod: account.AccountType, Amount: moneyPtr(amount), FeeAmount: moneyPtr(decimal.Zero),
			ActualAmount: moneyPtr(amount), Currency: wallet.Currency, Status: 1, AuditStatus: 1,
			RiskStatus: "PENDING", KYCStatus: "NOT_SUBMITTED", SubmittedAt: &now,
		}
		if err := tx.Create(&result).Error; err != nil {
			return err
		}
		if err := tx.Model(&wallet).Updates(map[string]any{"cash_balance": afterCash, "frozen_balance": afterFrozen, "last_transaction_at": now}).Error; err != nil {
			return err
		}
		txnID, err := createLedger(tx, wallet, uniqueNo("TXN"), "WITHDRAWAL_FREEZE", "OUT", "withdrawal", result.ID, result.OrderNo, amount, beforeCash, afterCash, beforeFrozen, afterFrozen, actor, "withdrawal submitted", nullableString("withdrawal-freeze:"+result.OrderNo))
		if err != nil {
			return err
		}
		freeze := model.GameWalletFreezes{TenantID: wallet.TenantID, SiteID: wallet.SiteID, FreezeNo: uniqueNo("FRZ"), WalletID: wallet.ID, UserID: userID, Amount: moneyPtr(amount), Currency: wallet.Currency, Reason: "withdrawal submitted", ReferenceType: "withdrawal", ReferenceID: result.ID, ReferenceNo: result.OrderNo, Status: 1, FreezeTransactionID: txnID, FrozenAt: now}
		if err = tx.Create(&freeze).Error; err != nil {
			return err
		}
		result.FreezeRecordID = freeze.ID
		return tx.Model(&result).Update("freeze_record_id", freeze.ID).Error
	})
	return result, wrapInternal(err)
}

func ensureWithdrawalAuditComplete(tx *gorm.DB, scope tenancy.Scope, userID uint64, currency string) error {
	remaining, err := bonusaudit.RemainingForUser(tx, scope, userID, currency)
	if err != nil {
		return err
	}
	if remaining.GreaterThan(decimal.Zero) {
		return ErrWithdrawalAudit
	}
	return nil
}

func (s *Service) SetWalletStatus(id uint64, frozen bool, reason string, actor Actor) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	return wrapInternal(s.db.Transaction(func(tx *gorm.DB) error {
		wallet, err := lockWallet(tx, id, actor.Scope)
		if err != nil {
			return err
		}
		before := wallet.Status
		after := walletActive
		permission := "fund:wallet:unfreeze"
		action := "wallet.unfreeze"
		if frozen {
			after = walletFrozen
			permission = "fund:wallet:freeze"
			action = "wallet.freeze"
		}
		if before == after {
			return ErrInvalidState
		}
		if err := tenancy.ApplySite(tx.Model(&model.GameUserWallets{}), "", actor.Scope).Where("id=?", id).Updates(map[string]any{"status": after}).Error; err != nil {
			return err
		}
		return s.audit(tx, actor, permission, action, "wallet", id, wallet.UserID, map[string]any{"status": walletStatus(before)}, map[string]any{"status": walletStatus(after)}, reason, nil, wallet.Currency)
	}))
}

func (s *Service) CreateAdjustment(command CreateAdjustmentCommand, actor Actor) (uint64, error) {
	if strings.TrimSpace(command.IdempotencyKey) == "" {
		return 0, ErrIdempotencyRequired
	}
	if err := requireReason(command.DetailedReason); err != nil {
		return 0, err
	}
	amount, err := ParseMoney(command.Amount)
	if err != nil || !amount.GreaterThan(decimal.Zero) {
		return 0, ErrInvalidAmount
	}
	kind := strings.ToUpper(strings.TrimSpace(command.AdjustmentType))
	if kind != "CREDIT" && kind != "DEBIT" {
		return 0, newError("INVALID_ADJUSTMENT_TYPE", "adjustment type must be CREDIT or DEBIT")
	}
	if command.UserID == 0 {
		return 0, newError("USER_REQUIRED", "user is required")
	}
	var result uint64
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var existing model.GameManualAdjustments
		existingQuery := tx.Model(&model.GameManualAdjustments{}).Where("idempotency_key=? AND deleted_at IS NULL", command.IdempotencyKey)
		if err := tenancy.ApplySite(existingQuery, "", actor.Scope).Take(&existing).Error; err == nil {
			result = existing.ID
			return nil
		} else if err != gorm.ErrRecordNotFound {
			return err
		}
		var wallet model.GameUserWallets
		var e error
		if command.WalletID > 0 {
			wallet, e = lockWallet(tx, command.WalletID, actor.Scope)
		} else {
			wallet, e = lockUserWallet(tx, command.UserID, command.Currency, actor.Scope)
		}
		if e != nil {
			return e
		}
		if wallet.UserID != command.UserID {
			return newError("WALLET_USER_MISMATCH", "wallet does not belong to user")
		}
		if wallet.Status != walletActive {
			return ErrWalletFrozen
		}
		if strings.TrimSpace(command.Currency) != "" && !strings.EqualFold(wallet.Currency, command.Currency) {
			return ErrCurrencyMismatch
		}
		balance := money(wallet.CashBalance)
		if kind == "DEBIT" && balance.LessThan(amount) {
			return ErrInsufficientBalance
		}
		row := model.GameManualAdjustments{TenantID: wallet.TenantID, SiteID: wallet.SiteID, AdjustmentNo: uniqueNo("ADJ"), IdempotencyKey: command.IdempotencyKey, UserID: command.UserID, WalletID: wallet.ID, AdjustmentType: kind, Amount: moneyPtr(amount), Currency: wallet.Currency, ReasonCode: strings.TrimSpace(command.ReasonCode), DetailedReason: strings.TrimSpace(command.DetailedReason), SupportingReference: strings.TrimSpace(command.SupportingReference), InternalRemark: strings.TrimSpace(command.InternalRemark), Status: "PENDING_APPROVAL", RequestedBalance: moneyPtr(balance), BalanceBefore: moneyPtr(balance), BalanceAfter: moneyPtr(balance), ApplicantID: actor.ID, ApplicantUsername: actor.Username}
		if e = tx.Create(&row).Error; e != nil {
			return e
		}
		result = row.ID
		return s.audit(tx, actor, "fund:adjustment:create", "adjustment.create", "manual_adjustment", row.ID, row.UserID, nil, map[string]any{"status": row.Status, "type": kind}, row.DetailedReason, row.Amount, row.Currency)
	})
	return result, wrapInternal(err)
}

func (s *Service) ApproveAdjustment(id uint64, command ReasonCommand, actor Actor) error {
	if err := requireReason(command.Reason); err != nil {
		return err
	}
	return wrapInternal(s.db.Transaction(func(tx *gorm.DB) error {
		var row model.GameManualAdjustments
		rowQuery := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.GameManualAdjustments{}).Where("id=? AND deleted_at IS NULL", id)
		if err := tenancy.ApplySite(rowQuery, "", actor.Scope).Take(&row).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrAdjustmentNotFound
			}
			return err
		}
		if err := validateAdjustmentApproval(row.Status, row.ApplicantID, actor.ID, actor.IsSuperAdmin()); err != nil {
			return err
		}
		wallet, err := lockWallet(tx, row.WalletID, actor.Scope)
		if err != nil {
			return err
		}
		if wallet.Status != walletActive {
			return ErrWalletFrozen
		}
		if !strings.EqualFold(wallet.Currency, row.Currency) {
			return ErrCurrencyMismatch
		}
		amount := money(row.Amount)
		before := money(wallet.CashBalance)
		after, direction, err := adjustedBalance(row.AdjustmentType, before, amount)
		if err != nil {
			return err
		}
		now := time.Now()
		if err = tenancy.ApplySite(tx.Model(&model.GameUserWallets{}), "", actor.Scope).Where("id=?", wallet.ID).Updates(map[string]any{"cash_balance": after, "last_transaction_at": now}).Error; err != nil {
			return err
		}
		txnID, err := createLedger(tx, wallet, uniqueNo("TXN"), "MANUAL_ADJUSTMENT", direction, "manual_adjustment", row.ID, row.AdjustmentNo, amount, before, after, money(wallet.FrozenBalance), money(wallet.FrozenBalance), actor, row.DetailedReason, nullableString("adjustment:"+row.IdempotencyKey))
		if err != nil {
			return err
		}
		updates := map[string]any{"status": "COMPLETED", "balance_before": before, "balance_after": after, "approver_id": actor.ID, "approver_username": actor.Username, "approval_reason": strings.TrimSpace(command.Reason), "wallet_transaction_id": txnID, "approved_at": now, "completed_at": now}
		if err = tenancy.ApplySite(tx.Model(&model.GameManualAdjustments{}), "", actor.Scope).Where("id=?", id).Updates(updates).Error; err != nil {
			return err
		}
		approval := model.GameManualAdjustmentApprovals{TenantID: row.TenantID, SiteID: row.SiteID, AdjustmentID: id, OperatorID: actor.ID, OperatorUsername: actor.Username, Decision: "APPROVED", BeforeStatus: row.Status, AfterStatus: "COMPLETED", Reason: strings.TrimSpace(command.Reason)}
		if err = tx.Create(&approval).Error; err != nil {
			return err
		}
		return s.audit(tx, actor, "fund:adjustment:approve", "adjustment.approve", "manual_adjustment", id, row.UserID, map[string]any{"status": row.Status, "balance": before.StringFixed(2)}, map[string]any{"status": "COMPLETED", "balance": after.StringFixed(2)}, command.Reason, row.Amount, row.Currency)
	}))
}

func (a Actor) IsSuperAdmin() bool {
	for _, permission := range a.Permissions {
		if permission == "*" {
			return true
		}
	}
	return false
}

func validateAdjustmentApproval(status string, applicantID, approverID uint64, allowSelfApproval bool) error {
	if status != "PENDING_APPROVAL" {
		return ErrInvalidState
	}
	if applicantID == approverID && !allowSelfApproval {
		return ErrMakerChecker
	}
	return nil
}

func adjustedBalance(kind string, balance, amount decimal.Decimal) (decimal.Decimal, string, error) {
	if kind == "CREDIT" {
		return balance.Add(amount), "IN", nil
	}
	if kind != "DEBIT" {
		return balance, "", newError("INVALID_ADJUSTMENT_TYPE", "invalid adjustment type")
	}
	if balance.LessThan(amount) {
		return balance, "", ErrInsufficientBalance
	}
	return balance.Sub(amount), "OUT", nil
}

func validateDepositConfirmation(status int) error {
	if status == 2 {
		return nil
	}
	if status != 1 && status != 6 {
		return ErrInvalidState
	}
	return nil
}

func withdrawalReviewTarget(status int, stage string) (int, error) {
	if stage == "RISK" && status == 1 {
		return 6, nil
	}
	if stage == "FINANCE" && status == 6 {
		return 7, nil
	}
	return 0, ErrInvalidState
}

func (s *Service) RejectAdjustment(id uint64, command ReasonCommand, actor Actor) error {
	return s.finishAdjustmentWithoutPosting(id, "REJECTED", "fund:adjustment:reject", "adjustment.reject", command, actor)
}
func (s *Service) CancelAdjustment(id uint64, command ReasonCommand, actor Actor) error {
	return s.finishAdjustmentWithoutPosting(id, "CANCELLED", "fund:adjustment:cancel", "adjustment.cancel", command, actor)
}
func (s *Service) finishAdjustmentWithoutPosting(id uint64, status, permission, action string, command ReasonCommand, actor Actor) error {
	if err := requireReason(command.Reason); err != nil {
		return err
	}
	return wrapInternal(s.db.Transaction(func(tx *gorm.DB) error {
		var row model.GameManualAdjustments
		rowQuery := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.GameManualAdjustments{}).Where("id=? AND deleted_at IS NULL", id)
		if err := tenancy.ApplySite(rowQuery, "", actor.Scope).Take(&row).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrAdjustmentNotFound
			}
			return err
		}
		if row.Status != "PENDING_APPROVAL" {
			return ErrInvalidState
		}
		if status == "CANCELLED" && row.ApplicantID != actor.ID {
			return newError("ONLY_APPLICANT_CAN_CANCEL", "only applicant can cancel pending adjustment")
		}
		if err := tenancy.ApplySite(tx.Model(&model.GameManualAdjustments{}), "", actor.Scope).Where("id=?", id).Updates(map[string]any{"status": status, "approver_id": actor.ID, "approver_username": actor.Username, "approval_reason": strings.TrimSpace(command.Reason)}).Error; err != nil {
			return err
		}
		approval := model.GameManualAdjustmentApprovals{TenantID: row.TenantID, SiteID: row.SiteID, AdjustmentID: id, OperatorID: actor.ID, OperatorUsername: actor.Username, Decision: status, BeforeStatus: row.Status, AfterStatus: status, Reason: command.Reason}
		if err := tx.Create(&approval).Error; err != nil {
			return err
		}
		return s.audit(tx, actor, permission, action, "manual_adjustment", id, row.UserID, map[string]any{"status": row.Status}, map[string]any{"status": status}, command.Reason, row.Amount, row.Currency)
	}))
}

func (s *Service) ReviewDeposit(id uint64, command ReasonCommand, actor Actor) error {
	if err := requireReason(command.Reason); err != nil {
		return err
	}
	return wrapInternal(s.db.Transaction(func(tx *gorm.DB) error {
		var row model.GameDepositOrders
		rowQuery := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.GameDepositOrders{}).Where("id=? AND deleted_at IS NULL", id)
		if err := tenancy.ApplySite(rowQuery, "", actor.Scope).Take(&row).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrDepositNotFound
			}
			return err
		}
		if row.Status != 1 {
			return ErrInvalidState
		}
		if err := tx.Model(&row).Updates(map[string]any{"status": 6, "reviewer_id": actor.ID, "reviewer_username": actor.Username, "review_reason": strings.TrimSpace(command.Reason)}).Error; err != nil {
			return err
		}
		return s.audit(tx, actor, "fund:deposit:review", "deposit.review", "deposit", id, row.UserID, map[string]any{"status": "PENDING"}, map[string]any{"status": "MANUAL_REVIEW"}, command.Reason, nil, row.Currency)
	}))
}

func (s *Service) ConfirmDeposit(id uint64, command DepositConfirmCommand, actor Actor) error {
	if err := requireReason(command.Reason); err != nil {
		return err
	}
	received, err := ParseMoney(command.ReceivedAmount)
	if err != nil || !received.GreaterThan(decimal.Zero) {
		return ErrInvalidAmount
	}
	return wrapInternal(s.db.Transaction(func(tx *gorm.DB) error {
		var row model.GameDepositOrders
		rowQuery := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.GameDepositOrders{}).Where("id=? AND deleted_at IS NULL", id)
		if err := tenancy.ApplySite(rowQuery, "", actor.Scope).Take(&row).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrDepositNotFound
			}
			return err
		}
		if err := validateDepositConfirmation(row.Status); err != nil {
			return err
		}
		if row.Status == 2 {
			return nil
		}
		var wallet model.GameUserWallets
		walletQuery := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.GameUserWallets{}).Where("user_id=? AND currency=? AND deleted_at IS NULL", row.UserID, row.Currency)
		if err := tenancy.ApplySite(walletQuery, "", actor.Scope).Take(&wallet).Error; err != nil {
			return ErrWalletNotFound
		}
		if wallet.Status != walletActive {
			return ErrWalletFrozen
		}
		before := money(wallet.CashBalance)
		after := before.Add(received)
		now := time.Now()
		if err := tx.Model(&wallet).Updates(map[string]any{"cash_balance": after, "last_transaction_at": now}).Error; err != nil {
			return err
		}
		txnID, err := createLedger(tx, wallet, uniqueNo("TXN"), "DEPOSIT", "IN", "deposit", row.ID, row.OrderNo, received, before, after, money(wallet.FrozenBalance), money(wallet.FrozenBalance), actor, command.Reason, nullableString("deposit:"+row.OrderNo))
		if err != nil {
			return err
		}
		updates := map[string]any{"status": 2, "paid_amount": received, "wallet_transaction_id": txnID, "reviewer_id": actor.ID, "reviewer_username": actor.Username, "review_reason": command.Reason, "evidence_reference": strings.TrimSpace(command.EvidenceReference), "channel_order_no": strings.TrimSpace(command.ChannelOrderNo), "paid_at": now, "completed_at": now}
		if err = tx.Model(&row).Updates(updates).Error; err != nil {
			return err
		}
		var depositSequence int64
		if err := tx.Model(&model.GameDepositOrders{}).Where(
			"tenant_id=? AND site_id=? AND user_id=? AND status=2 AND deleted_at IS NULL", row.TenantID, row.SiteID, row.UserID,
		).Count(&depositSequence).Error; err != nil {
			return err
		}
		var country string
		if err := tx.Model(&model.GameUsers{}).Select("country_code").Where(
			"id=? AND tenant_id=? AND site_id=? AND deleted_at IS NULL", row.UserID, row.TenantID, row.SiteID,
		).Scan(&country).Error; err != nil {
			return err
		}
		correlationID := strings.TrimSpace(actor.CorrelationID)
		if correlationID == "" {
			correlationID = row.OrderNo
		}
		if _, err := eventbus.New(tx).PublishTx(tx, eventbus.Event{
			Scope: tenancy.Scope{TenantID: row.TenantID, SiteID: row.SiteID}, TenantID: row.TenantID, SiteID: row.SiteID,
			EventType: eventbus.EventPlayerDepositSucceeded, AggregateType: "deposit_order", AggregateID: fmt.Sprint(row.ID),
			Source: "fundadmin", IdempotencyKey: "deposit:" + row.OrderNo + ":succeeded", TraceID: correlationID, CorrelationID: correlationID,
			OccurredAt: now, TargetConsumers: []string{eventbus.ConsumerPromotionDeposit},
			Payload: map[string]any{
				"depositOrderID": row.ID, "orderNo": row.OrderNo, "userID": row.UserID, "amount": received.StringFixed(2),
				"currency": strings.ToUpper(row.Currency), "country": strings.ToUpper(country), "depositSequence": depositSequence,
			},
			Metadata: map[string]any{"operatorID": actor.ID, "operatorUsername": actor.Username},
		}); err != nil {
			return err
		}
		return s.audit(tx, actor, "fund:deposit:confirm", "deposit.confirm", "deposit", id, row.UserID, map[string]any{"status": depositStatus(row.Status)}, map[string]any{"status": "SUCCESS", "receivedAmount": received.StringFixed(2)}, command.Reason, moneyPtr(received), row.Currency)
	}))
}

func (s *Service) CancelDeposit(id uint64, command ReasonCommand, actor Actor) error {
	if err := requireReason(command.Reason); err != nil {
		return err
	}
	return wrapInternal(s.db.Transaction(func(tx *gorm.DB) error {
		var row model.GameDepositOrders
		rowQuery := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.GameDepositOrders{}).Where("id=? AND deleted_at IS NULL", id)
		if err := tenancy.ApplySite(rowQuery, "", actor.Scope).Take(&row).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrDepositNotFound
			}
			return err
		}
		if row.Status != 1 && row.Status != 6 {
			return ErrInvalidState
		}
		now := time.Now()
		if err := tx.Model(&row).Updates(map[string]any{"status": 4, "reviewer_id": actor.ID, "reviewer_username": actor.Username, "review_reason": command.Reason, "completed_at": now}).Error; err != nil {
			return err
		}
		return s.audit(tx, actor, "fund:deposit:cancel", "deposit.cancel", "deposit", id, row.UserID, map[string]any{"status": depositStatus(row.Status)}, map[string]any{"status": "CANCELLED"}, command.Reason, nil, row.Currency)
	}))
}

func (s *Service) ReviewWithdrawal(id uint64, stage string, command ReasonCommand, actor Actor) error {
	if err := requireReason(command.Reason); err != nil {
		return err
	}
	stage = strings.ToUpper(stage)
	return wrapInternal(s.db.Transaction(func(tx *gorm.DB) error {
		var row model.GameWithdrawOrders
		rowQuery := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.GameWithdrawOrders{}).Where("id=? AND deleted_at IS NULL", id)
		if err := tenancy.ApplySite(rowQuery, "", actor.Scope).Take(&row).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrWithdrawalNotFound
			}
			return err
		}
		next, transitionErr := withdrawalReviewTarget(row.Status, stage)
		permission := ""
		if transitionErr != nil {
			return transitionErr
		}
		if stage == "RISK" {
			permission = "fund:withdrawal:risk-review"
		} else if stage == "FINANCE" {
			permission = "fund:withdrawal:finance-review"
		}
		now := time.Now()
		if err := tx.Model(&row).Updates(map[string]any{"status": next, "reviewer_id": actor.ID, "reviewer_username": actor.Username, "review_remark": command.Reason, "reviewed_at": now}).Error; err != nil {
			return err
		}
		approval := model.GameWithdrawalApprovals{TenantID: row.TenantID, SiteID: row.SiteID, WithdrawalOrderID: id, OperatorID: actor.ID, OperatorUsername: actor.Username, Stage: stage, Decision: "PASSED", BeforeStatus: withdrawalStatus(row.Status), AfterStatus: withdrawalStatus(next), Reason: command.Reason}
		if err := tx.Create(&approval).Error; err != nil {
			return err
		}
		return s.audit(tx, actor, permission, "withdrawal."+strings.ToLower(stage)+"-review", "withdrawal", id, row.UserID, map[string]any{"status": withdrawalStatus(row.Status)}, map[string]any{"status": withdrawalStatus(next)}, command.Reason, nil, row.Currency)
	}))
}

func (s *Service) ApproveWithdrawal(id uint64, command ReasonCommand, actor Actor) error {
	if err := requireReason(command.Reason); err != nil {
		return err
	}
	return wrapInternal(s.db.Transaction(func(tx *gorm.DB) error {
		var row model.GameWithdrawOrders
		rowQuery := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.GameWithdrawOrders{}).Where("id=? AND deleted_at IS NULL", id)
		if err := tenancy.ApplySite(rowQuery, "", actor.Scope).Take(&row).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrWithdrawalNotFound
			}
			return err
		}
		if row.Status != 7 {
			return ErrInvalidState
		}
		now := time.Now()
		if err := tx.Model(&row).Updates(map[string]any{"status": 9, "audit_status": 2, "reviewer_id": actor.ID, "reviewer_username": actor.Username, "review_remark": command.Reason, "approved_at": now}).Error; err != nil {
			return err
		}
		approval := model.GameWithdrawalApprovals{TenantID: row.TenantID, SiteID: row.SiteID, WithdrawalOrderID: id, OperatorID: actor.ID, OperatorUsername: actor.Username, Stage: "FINANCE", Decision: "APPROVED", BeforeStatus: "FINANCE_REVIEW", AfterStatus: "PROCESSING", Reason: command.Reason}
		if err := tx.Create(&approval).Error; err != nil {
			return err
		}
		payout := model.GameWithdrawalPayoutLogs{TenantID: row.TenantID, SiteID: row.SiteID, WithdrawalOrderID: id, RequestResult: "NOT_CONFIGURED", RetryCount: 0, FailureReason: "payout provider is not configured", RawMetadata: "{}", RequestedAt: now}
		if err := tx.Create(&payout).Error; err != nil {
			return err
		}
		return s.audit(tx, actor, "fund:withdrawal:approve", "withdrawal.approve", "withdrawal", id, row.UserID, map[string]any{"status": "FINANCE_REVIEW"}, map[string]any{"status": "PROCESSING"}, command.Reason, row.Amount, row.Currency)
	}))
}

func (s *Service) RejectWithdrawal(id uint64, command ReasonCommand, actor Actor) error {
	if err := requireReason(command.Reason); err != nil {
		return err
	}
	return wrapInternal(s.db.Transaction(func(tx *gorm.DB) error {
		var row model.GameWithdrawOrders
		rowQuery := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.GameWithdrawOrders{}).Where("id=? AND deleted_at IS NULL", id)
		if err := tenancy.ApplySite(rowQuery, "", actor.Scope).Take(&row).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrWithdrawalNotFound
			}
			return err
		}
		if row.Status != 1 && row.Status != 6 && row.Status != 7 {
			return ErrInvalidState
		}
		if err := s.releaseWithdrawalFreeze(tx, &row, actor, command.Reason); err != nil {
			return err
		}
		now := time.Now()
		if err := tx.Model(&row).Updates(map[string]any{"status": 4, "audit_status": 3, "reviewer_id": actor.ID, "reviewer_username": actor.Username, "review_remark": command.Reason, "completed_at": now}).Error; err != nil {
			return err
		}
		approval := model.GameWithdrawalApprovals{TenantID: row.TenantID, SiteID: row.SiteID, WithdrawalOrderID: id, OperatorID: actor.ID, OperatorUsername: actor.Username, Stage: "FINAL", Decision: "REJECTED", BeforeStatus: withdrawalStatus(row.Status), AfterStatus: "REJECTED", Reason: command.Reason}
		if err := tx.Create(&approval).Error; err != nil {
			return err
		}
		return s.audit(tx, actor, "fund:withdrawal:reject", "withdrawal.reject", "withdrawal", id, row.UserID, map[string]any{"status": withdrawalStatus(row.Status)}, map[string]any{"status": "REJECTED"}, command.Reason, row.Amount, row.Currency)
	}))
}

func (s *Service) releaseWithdrawalFreeze(tx *gorm.DB, row *model.GameWithdrawOrders, actor Actor, reason string) error {
	if row.FreezeRecordID == 0 {
		return nil
	}
	var freeze model.GameWalletFreezes
	freezeQuery := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.GameWalletFreezes{}).Where("id=?", row.FreezeRecordID)
	if err := tenancy.ApplySite(freezeQuery, "", actor.Scope).Take(&freeze).Error; err != nil {
		return err
	}
	if freeze.Status != 1 {
		return ErrInvalidState
	}
	wallet, err := lockWallet(tx, freeze.WalletID, actor.Scope)
	if err != nil {
		return err
	}
	amount := money(freeze.Amount)
	beforeCash := money(wallet.CashBalance)
	beforeFrozen := money(wallet.FrozenBalance)
	if beforeFrozen.LessThan(amount) {
		return newError("FROZEN_BALANCE_MISMATCH", "wallet frozen balance is inconsistent")
	}
	afterCash := beforeCash.Add(amount)
	afterFrozen := beforeFrozen.Sub(amount)
	now := time.Now()
	if err = tx.Model(&wallet).Updates(map[string]any{"cash_balance": afterCash, "frozen_balance": afterFrozen, "last_transaction_at": now}).Error; err != nil {
		return err
	}
	txnID, err := createLedger(tx, wallet, uniqueNo("TXN"), "WITHDRAWAL_RELEASE", "IN", "withdrawal", row.ID, row.OrderNo, amount, beforeCash, afterCash, beforeFrozen, afterFrozen, actor, reason, nullableString("withdrawal-release:"+row.OrderNo))
	if err != nil {
		return err
	}
	return tx.Model(&freeze).Updates(map[string]any{"status": 2, "release_transaction_id": txnID, "released_at": now}).Error
}

func (s *Service) RetryWithdrawal(id uint64, command ReasonCommand, actor Actor) error {
	if err := requireReason(command.Reason); err != nil {
		return err
	}
	return wrapInternal(s.db.Transaction(func(tx *gorm.DB) error {
		var row model.GameWithdrawOrders
		rowQuery := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.GameWithdrawOrders{}).Where("id=? AND deleted_at IS NULL", id)
		if err := tenancy.ApplySite(rowQuery, "", actor.Scope).Take(&row).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrWithdrawalNotFound
			}
			return err
		}
		if row.Status != 3 {
			return ErrInvalidState
		}
		now := time.Now()
		if err := tx.Model(&row).Updates(map[string]any{"status": 9, "payout_retry_count": gorm.Expr("payout_retry_count+1"), "failure_reason": ""}).Error; err != nil {
			return err
		}
		payout := model.GameWithdrawalPayoutLogs{TenantID: row.TenantID, SiteID: row.SiteID, WithdrawalOrderID: id, RequestResult: "RETRY_NOT_CONFIGURED", RetryCount: row.PayoutRetryCount + 1, FailureReason: "payout provider is not configured", RawMetadata: "{}", RequestedAt: now}
		if err := tx.Create(&payout).Error; err != nil {
			return err
		}
		return s.audit(tx, actor, "fund:withdrawal:retry", "withdrawal.retry", "withdrawal", id, row.UserID, map[string]any{"status": "FAILED"}, map[string]any{"status": "PROCESSING"}, command.Reason, row.Amount, row.Currency)
	}))
}

func (s *Service) GenerateReconciliation(command GenerateReconciliationCommand, actor Actor) (uint64, error) {
	date, err := time.Parse("2006-01-02", strings.TrimSpace(command.BusinessDate))
	if err != nil {
		return 0, newError("INVALID_BUSINESS_DATE", "businessDate must be YYYY-MM-DD")
	}
	business := strings.ToUpper(strings.TrimSpace(command.BusinessType))
	if business != "DEPOSIT" && business != "WITHDRAWAL" && business != "ADJUSTMENT" {
		return 0, newError("INVALID_BUSINESS_TYPE", "businessType is invalid")
	}
	currency := strings.ToUpper(strings.TrimSpace(command.Currency))
	if currency == "" {
		return 0, ErrCurrencyMismatch
	}
	var result uint64
	err = s.db.Transaction(func(tx *gorm.DB) error {
		source := "deposit"
		orderTable := "game_deposit_orders"
		amountColumn := "paid_amount"
		successStatus := 2
		if business == "WITHDRAWAL" {
			source = "withdrawal"
			orderTable = "game_withdraw_orders"
			amountColumn = "actual_amount"
		} else if business == "ADJUSTMENT" {
			source = "manual_adjustment"
			orderTable = "game_manual_adjustments"
			amountColumn = "amount"
			successStatus = 0
		}
		start := date
		end := date.Add(24 * time.Hour)
		orderQuery := tx.Table(orderTable).Where("created_at>=? AND created_at<? AND currency=? AND deleted_at IS NULL", start, end, currency)
		orderQuery = tenancy.ApplySite(orderQuery, "", actor.Scope)
		if business == "ADJUSTMENT" {
			orderQuery = orderQuery.Where("status=?", "COMPLETED")
		} else {
			orderQuery = orderQuery.Where("status=?", successStatus)
		}
		if command.ChannelCode != "" && business == "DEPOSIT" {
			orderQuery = orderQuery.Where("channel_code=?", command.ChannelCode)
		}
		var orderCount int64
		var orderAmount decimal.Decimal
		if e := orderQuery.Count(&orderCount).Error; e != nil {
			return e
		}
		if e := orderQuery.Select("COALESCE(SUM(" + amountColumn + "),0)").Scan(&orderAmount).Error; e != nil {
			return e
		}
		ledgerQuery := tx.Table("game_wallet_transactions").Where("source_type=? AND created_at>=? AND created_at<? AND currency=? AND status=1 AND deleted_at IS NULL", source, start, end, currency)
		ledgerQuery = tenancy.ApplySite(ledgerQuery, "", actor.Scope)
		var ledgerCount int64
		var ledgerAmount decimal.Decimal
		if e := ledgerQuery.Count(&ledgerCount).Error; e != nil {
			return e
		}
		if e := ledgerQuery.Select("COALESCE(SUM(amount),0)").Scan(&ledgerAmount).Error; e != nil {
			return e
		}
		difference := orderAmount.Sub(ledgerAmount)
		absDifference := difference.Abs()
		diffCount := orderCount - ledgerCount
		if diffCount < 0 {
			diffCount = -diffCount
		}
		status := "MATCHED"
		if !difference.IsZero() || diffCount != 0 {
			status = "DIFFERENCE"
		}
		now := time.Now()
		record := model.GameReconciliationRecords{TenantID: actor.Scope.TenantID, SiteID: actor.Scope.SiteID, ReconciliationNo: uniqueNo("REC"), BusinessDate: date, BusinessType: business, ChannelCode: strings.TrimSpace(command.ChannelCode), Currency: currency, ReconciliationMode: "INTERNAL", PlatformOrderCount: uint(orderCount), PlatformAmount: moneyPtr(orderAmount), ChannelOrderCount: uint(ledgerCount), ChannelAmount: moneyPtr(ledgerAmount), DifferenceCount: uint(diffCount), DifferenceAmount: moneyPtr(absDifference), Status: status, GeneratedByID: actor.ID, GeneratedBy: actor.Username, GeneratedAt: now, CompletedAt: &now}
		if e := tx.Create(&record).Error; e != nil {
			return e
		}
		result = record.ID
		if status == "DIFFERENCE" {
			diff := model.GameReconciliationDifferences{TenantID: actor.Scope.TenantID, SiteID: actor.Scope.SiteID, ReconciliationID: record.ID, PlatformStatus: "COMPLETED", ChannelStatus: "LEDGER_POSTED", PlatformAmount: moneyPtr(orderAmount), ChannelAmount: moneyPtr(ledgerAmount), DifferenceType: "AMOUNT_OR_COUNT_MISMATCH", DifferenceAmount: moneyPtr(absDifference), ResolutionStatus: "OPEN", DetectedAt: now}
			if e := tx.Create(&diff).Error; e != nil {
				return e
			}
		}
		log := model.GameReconciliationLogs{TenantID: actor.Scope.TenantID, SiteID: actor.Scope.SiteID, ReconciliationID: record.ID, Level: "INFO", Message: "Internal reconciliation generated from platform orders and immutable wallet ledger"}
		if e := tx.Create(&log).Error; e != nil {
			return e
		}
		return s.audit(tx, actor, "fund:reconciliation:generate", "reconciliation.generate", "reconciliation", record.ID, 0, nil, map[string]any{"status": status, "differenceAmount": absDifference.StringFixed(2)}, "generate internal reconciliation", nil, currency)
	})
	return result, wrapInternal(err)
}

func (s *Service) ResolveDifference(reconciliationID, differenceID uint64, command ReasonCommand, actor Actor) error {
	if err := requireReason(command.Reason); err != nil {
		return err
	}
	return wrapInternal(s.db.Transaction(func(tx *gorm.DB) error {
		var diff model.GameReconciliationDifferences
		diffQuery := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&model.GameReconciliationDifferences{}).Where("id=? AND reconciliation_id=? AND deleted_at IS NULL", differenceID, reconciliationID)
		if err := tenancy.ApplySite(diffQuery, "", actor.Scope).Take(&diff).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrDifferenceNotFound
			}
			return err
		}
		if diff.ResolutionStatus != "OPEN" {
			return ErrInvalidState
		}
		now := time.Now()
		if err := tx.Model(&diff).Updates(map[string]any{"resolution_status": "RESOLVED", "resolution_reason": command.Reason, "resolved_by_id": actor.ID, "resolved_by": actor.Username, "resolved_at": now}).Error; err != nil {
			return err
		}
		resolution := model.GameReconciliationResolutions{TenantID: diff.TenantID, SiteID: diff.SiteID, ReconciliationID: reconciliationID, DifferenceID: differenceID, OperatorID: actor.ID, OperatorUsername: actor.Username, Action: "RESOLVE", Reason: command.Reason}
		if err := tx.Create(&resolution).Error; err != nil {
			return err
		}
		var open int64
		openQuery := tx.Table("game_reconciliation_differences").Where("reconciliation_id=? AND resolution_status='OPEN' AND deleted_at IS NULL", reconciliationID)
		if err := tenancy.ApplySite(openQuery, "", actor.Scope).Count(&open).Error; err != nil {
			return err
		}
		if open == 0 {
			if err := tenancy.ApplySite(tx.Model(&model.GameReconciliationRecords{}), "", actor.Scope).Where("id=?", reconciliationID).Update("status", "RESOLVED").Error; err != nil {
				return err
			}
		}
		return s.audit(tx, actor, "fund:reconciliation:resolve", "reconciliation.resolve", "reconciliation", reconciliationID, 0, map[string]any{"differenceStatus": "OPEN"}, map[string]any{"differenceStatus": "RESOLVED"}, command.Reason, nil, "")
	}))
}
