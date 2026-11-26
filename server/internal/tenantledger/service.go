package tenantledger

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"casino/internal/eventbus"
	"casino/internal/tenancy"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service { return &Service{db: db} }

func (s *Service) PostTransaction(ctx context.Context, command PostCommand) (TransactionResult, error) {
	var result TransactionResult
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := enableLedgerPosting(tx); err != nil {
			return err
		}
		defer disableLedgerPosting(tx)
		var err error
		result, err = s.postTransactionTx(tx, command)
		return err
	})
	if isDuplicateError(err) {
		existing, existingErr := s.transactionByIdempotency(ctx, strings.TrimSpace(command.IdempotencyKey))
		if existingErr == nil {
			existing.Idempotent = true
		}
		return existing, existingErr
	}
	return result, err
}

func (s *Service) CreditTenantDeposit(ctx context.Context, command TenantMoneyCommand) (TransactionResult, error) {
	command.Currency = NormalizeCurrency(command.Currency)
	if err := validateTenantMoneyCommand(command, true); err != nil {
		return TransactionResult{}, err
	}
	return s.PostTransaction(ctx, PostCommand{
		Scope:            command.Scope,
		TenantID:         command.TenantID,
		SiteID:           command.SiteID,
		Currency:         command.Currency,
		EventType:        EventTenantDeposit,
		BusinessType:     "TENANT_DEPOSIT",
		BusinessOrderNo:  command.IdempotencyKey,
		IdempotencyKey:   command.IdempotencyKey,
		Reason:           command.Reason,
		OperatorID:       command.Actor.ID,
		OperatorUsername: command.Actor.Username,
		CorrelationID:    command.Actor.CorrelationID,
		Entries: []EntryCommand{
			{AccountType: AccountTenantAvailable, Direction: DirectionDebit, Amount: command.Amount, Memo: "tenant deposit available balance"},
			{AccountType: AccountTenantDeposit, Direction: DirectionCredit, Amount: command.Amount, Memo: "tenant deposit liability"},
		},
	})
}

func (s *Service) SetCreditLimit(ctx context.Context, command CreditLimitCommand) (TransactionResult, error) {
	command.Currency = NormalizeCurrency(command.Currency)
	if err := validateCreditLimitCommand(command); err != nil {
		return TransactionResult{}, err
	}
	var result TransactionResult
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := enableLedgerPosting(tx); err != nil {
			return err
		}
		defer disableLedgerPosting(tx)
		limit, err := s.ensureCreditLimitTx(tx, command)
		if err != nil {
			return err
		}
		diff := command.CreditLimit.Sub(limit.CreditLimit)
		if diff.IsZero() {
			limit.UpdatedBy = command.Actor.ID
			if err := tx.Model(&tenantCreditLimit{}).Where("id=?", limit.ID).Updates(map[string]any{"updated_by": command.Actor.ID, "updated_at": time.Now()}).Error; err != nil {
				return err
			}
			result = TransactionResult{TenantID: command.TenantID, SiteID: command.SiteID, Currency: command.Currency, EventType: EventTenantCredit, CreditLimit: &command.CreditLimit, Idempotent: true}
			return nil
		}
		amount := diff.Abs()
		entries := []EntryCommand{
			{AccountType: AccountTenantAvailable, Direction: DirectionDebit, Amount: amount, Memo: "tenant credit available balance"},
			{AccountType: AccountTenantCredit, Direction: DirectionCredit, Amount: amount, Memo: "tenant credit limit"},
		}
		if diff.IsNegative() {
			entries = []EntryCommand{
				{AccountType: AccountTenantCredit, Direction: DirectionDebit, Amount: amount, Memo: "reduce tenant credit limit"},
				{AccountType: AccountTenantAvailable, Direction: DirectionCredit, Amount: amount, Memo: "reduce tenant available balance"},
			}
		}
		result, err = s.postTransactionTx(tx, PostCommand{
			Scope:            command.Scope,
			TenantID:         command.TenantID,
			SiteID:           command.SiteID,
			Currency:         command.Currency,
			EventType:        EventTenantCredit,
			BusinessType:     "TENANT_CREDIT_LIMIT",
			BusinessOrderNo:  command.IdempotencyKey,
			IdempotencyKey:   command.IdempotencyKey,
			Reason:           command.Reason,
			OperatorID:       command.Actor.ID,
			OperatorUsername: command.Actor.Username,
			CorrelationID:    command.Actor.CorrelationID,
			Entries:          entries,
		})
		if err != nil {
			return err
		}
		if result.Idempotent {
			return nil
		}
		if err := tx.Model(&tenantCreditLimit{}).Where("id=?", limit.ID).Updates(map[string]any{
			"credit_limit": command.CreditLimit, "last_transaction_id": result.ID, "updated_by": command.Actor.ID, "updated_at": time.Now(),
		}).Error; err != nil {
			return err
		}
		result.CreditLimit = &command.CreditLimit
		return nil
	})
	if isDuplicateError(err) {
		existing, existingErr := s.transactionByIdempotency(ctx, strings.TrimSpace(command.IdempotencyKey))
		if existingErr == nil {
			existing.Idempotent = true
		}
		return existing, existingErr
	}
	return result, err
}

func (s *Service) Freeze(ctx context.Context, tenantID, siteID uint64, amount decimal.Decimal, currency string, actor Actor, idempotencyKey, reason string) (TransactionResult, error) {
	command := TenantMoneyCommand{Scope: actor.Scope(), TenantID: tenantID, SiteID: siteID, Amount: amount, Currency: currency, Actor: actor, IdempotencyKey: idempotencyKey, Reason: reason}
	command.Currency = NormalizeCurrency(command.Currency)
	if err := validateTenantMoneyCommand(command, true); err != nil {
		return TransactionResult{}, err
	}
	return s.PostTransaction(ctx, PostCommand{
		Scope:            command.Scope,
		TenantID:         tenantID,
		SiteID:           siteID,
		Currency:         command.Currency,
		EventType:        EventTenantFreeze,
		BusinessType:     "TENANT_FUNDS_FREEZE",
		BusinessOrderNo:  idempotencyKey,
		IdempotencyKey:   idempotencyKey,
		Reason:           reason,
		OperatorID:       actor.ID,
		OperatorUsername: actor.Username,
		CorrelationID:    actor.CorrelationID,
		Entries: []EntryCommand{
			{AccountType: AccountTenantFrozen, Direction: DirectionDebit, Amount: amount, Memo: "freeze tenant funds"},
			{AccountType: AccountTenantAvailable, Direction: DirectionCredit, Amount: amount, Memo: "freeze tenant available balance"},
		},
	})
}

func (s *Service) Unfreeze(ctx context.Context, tenantID, siteID uint64, amount decimal.Decimal, currency string, actor Actor, idempotencyKey, reason string) (TransactionResult, error) {
	command := TenantMoneyCommand{Scope: actor.Scope(), TenantID: tenantID, SiteID: siteID, Amount: amount, Currency: currency, Actor: actor, IdempotencyKey: idempotencyKey, Reason: reason}
	command.Currency = NormalizeCurrency(command.Currency)
	if err := validateTenantMoneyCommand(command, true); err != nil {
		return TransactionResult{}, err
	}
	return s.PostTransaction(ctx, PostCommand{
		Scope:            command.Scope,
		TenantID:         tenantID,
		SiteID:           siteID,
		Currency:         command.Currency,
		EventType:        EventTenantUnfreeze,
		BusinessType:     "TENANT_FUNDS_UNFREEZE",
		BusinessOrderNo:  idempotencyKey,
		IdempotencyKey:   idempotencyKey,
		Reason:           reason,
		OperatorID:       actor.ID,
		OperatorUsername: actor.Username,
		CorrelationID:    actor.CorrelationID,
		Entries: []EntryCommand{
			{AccountType: AccountTenantAvailable, Direction: DirectionDebit, Amount: amount, Memo: "unfreeze tenant funds"},
			{AccountType: AccountTenantFrozen, Direction: DirectionCredit, Amount: amount, Memo: "unfreeze tenant frozen balance"},
		},
	})
}

func (s *Service) ReverseTransaction(ctx context.Context, transactionID uint64, reason string, actor Actor) (TransactionResult, error) {
	if strings.TrimSpace(reason) == "" {
		return TransactionResult{}, ErrReasonRequired
	}
	var result TransactionResult
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := enableLedgerPosting(tx); err != nil {
			return err
		}
		defer disableLedgerPosting(tx)
		var original ledgerTransaction
		query := lockForUpdate(tx).Where("id=? AND status=? AND deleted_at IS NULL", transactionID, StatusPosted)
		if !actor.Platform {
			query = query.Where("tenant_id=?", actor.TenantID)
			if actor.SiteID > 0 {
				query = query.Where("site_id IN (0, ?)", actor.SiteID)
			}
		}
		if err := query.Take(&original).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrTransactionNotFound
			}
			return err
		}
		var count int64
		if err := tx.Model(&ledgerTransaction{}).Where("reversal_of_transaction_id=? AND deleted_at IS NULL", transactionID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrTransactionReversed
		}
		var originalEntries []ledgerEntry
		if err := tx.Where("transaction_id=? AND deleted_at IS NULL", transactionID).Order("entry_no ASC").Find(&originalEntries).Error; err != nil {
			return err
		}
		entries := make([]EntryCommand, 0, len(originalEntries))
		for _, entry := range originalEntries {
			direction := DirectionDebit
			if entry.Direction == DirectionDebit {
				direction = DirectionCredit
			}
			entries = append(entries, EntryCommand{AccountID: entry.AccountID, AccountType: entry.AccountType, Direction: direction, Amount: entry.Amount, Memo: "reverse " + original.TransactionNo})
		}
		reversalOf := original.ID
		var err error
		result, err = s.postTransactionTx(tx, PostCommand{
			Scope:                   actor.Scope(),
			TenantID:                original.TenantID,
			SiteID:                  original.SiteID,
			Currency:                original.Currency,
			EventType:               EventReversal,
			BusinessType:            "LEDGER_REVERSAL",
			BusinessOrderNo:         original.TransactionNo,
			IdempotencyKey:          "reversal:" + original.TransactionNo,
			Reason:                  reason,
			OperatorID:              actor.ID,
			OperatorUsername:        actor.Username,
			CorrelationID:           actor.CorrelationID,
			ReversalOfTransactionID: &reversalOf,
			Entries:                 entries,
		})
		return err
	})
	if isDuplicateError(err) {
		return TransactionResult{}, ErrTransactionReversed
	}
	return result, err
}

func (s *Service) GetBalance(ctx context.Context, accountID uint64) (AccountBalance, error) {
	var account ledgerAccount
	if err := s.db.WithContext(ctx).Where("id=? AND deleted_at IS NULL", accountID).Take(&account).Error; err != nil {
		return AccountBalance{}, err
	}
	return accountBalance(account), nil
}

func (s *Service) GetTransaction(ctx context.Context, transactionID uint64, scope tenancy.Scope) (TransactionResult, error) {
	var row ledgerTransaction
	db := s.db.WithContext(ctx).Where("id=? AND deleted_at IS NULL", transactionID)
	if !scope.Platform {
		db = db.Where("tenant_id=?", scope.TenantID)
		if scope.SiteID > 0 {
			db = db.Where("site_id IN (0, ?)", scope.SiteID)
		}
	}
	if err := db.Take(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return TransactionResult{}, ErrTransactionNotFound
		}
		return TransactionResult{}, err
	}
	return s.transactionResult(ctx, row)
}

func (s *Service) ListAccounts(ctx context.Context, scope tenancy.Scope, tenantID, siteID uint64, currency string) ([]AccountBalance, error) {
	if err := validateScope(scope, tenantID, siteID); err != nil {
		return nil, err
	}
	var accounts []ledgerAccount
	db := s.db.WithContext(ctx).Where("tenant_id=? AND deleted_at IS NULL", tenantID)
	if siteID > 0 {
		db = db.Where("site_id=?", siteID)
	}
	if currency = NormalizeCurrency(currency); currency != "" {
		db = db.Where("currency=?", currency)
	}
	if err := db.Order("site_id ASC,currency ASC,account_type ASC").Find(&accounts).Error; err != nil {
		return nil, err
	}
	out := make([]AccountBalance, 0, len(accounts))
	for _, account := range accounts {
		out = append(out, accountBalance(account))
	}
	return out, nil
}

func (s *Service) postTransactionTx(tx *gorm.DB, command PostCommand) (TransactionResult, error) {
	if err := validatePostCommand(command); err != nil {
		return TransactionResult{}, err
	}
	if existing, err := s.transactionByIdempotencyTx(tx, command.IdempotencyKey); err == nil {
		existing.Idempotent = true
		return existing, nil
	} else if err != gorm.ErrRecordNotFound {
		return TransactionResult{}, err
	}
	accounts, resolvedEntries, err := s.resolveAccountsTx(tx, command)
	if err != nil {
		return TransactionResult{}, err
	}
	debitTotal, creditTotal := decimal.Zero, decimal.Zero
	for _, entry := range resolvedEntries {
		if entry.Direction == DirectionDebit {
			debitTotal = debitTotal.Add(entry.Amount)
		} else {
			creditTotal = creditTotal.Add(entry.Amount)
		}
	}
	if !debitTotal.Equal(creditTotal) {
		return TransactionResult{}, ErrUnbalancedTransaction
	}
	now := time.Now()
	row := ledgerTransaction{
		TransactionNo:           uniqueNo("LTX"),
		TenantID:                command.TenantID,
		SiteID:                  command.SiteID,
		Currency:                NormalizeCurrency(command.Currency),
		EventType:               strings.ToUpper(strings.TrimSpace(command.EventType)),
		BusinessType:            strings.ToUpper(strings.TrimSpace(command.BusinessType)),
		BusinessOrderNo:         strings.TrimSpace(command.BusinessOrderNo),
		IdempotencyKey:          strings.TrimSpace(command.IdempotencyKey),
		Status:                  StatusPosted,
		DebitAmount:             debitTotal,
		CreditAmount:            creditTotal,
		ReversalOfTransactionID: command.ReversalOfTransactionID,
		OperatorID:              command.OperatorID,
		OperatorUsername:        strings.TrimSpace(command.OperatorUsername),
		ApprovalID:              command.ApprovalID,
		Reason:                  strings.TrimSpace(command.Reason),
		Metadata:                marshal(command.Metadata),
		CreatedAt:               now,
		UpdatedAt:               now,
	}
	if err := tx.Create(&row).Error; err != nil {
		return TransactionResult{}, err
	}
	results := make([]EntryResult, 0, len(resolvedEntries))
	for index, entry := range resolvedEntries {
		account := accounts[entry.AccountID]
		before := account.Balance
		after, err := applyEntryBalance(account, entry.Direction, entry.Amount)
		if err != nil {
			return TransactionResult{}, err
		}
		updates := map[string]any{"balance": after, "version": gorm.Expr("version+1"), "updated_at": now}
		if account.AccountType == AccountTenantAvailable {
			updates["available_balance"] = after
		}
		if account.AccountType == AccountTenantFrozen {
			updates["frozen_balance"] = after
		}
		if err := tx.Model(&ledgerAccount{}).Where("id=?", account.ID).Updates(updates).Error; err != nil {
			return TransactionResult{}, err
		}
		account.Balance = after
		if account.AccountType == AccountTenantAvailable {
			account.AvailableBalance = after
		}
		if account.AccountType == AccountTenantFrozen {
			account.FrozenBalance = after
		}
		accounts[entry.AccountID] = account
		entryRow := ledgerEntry{
			TransactionID: row.ID,
			EntryNo:       uint16(index + 1),
			AccountID:     account.ID,
			TenantID:      account.TenantID,
			SiteID:        account.SiteID,
			Currency:      account.Currency,
			AccountType:   account.AccountType,
			Direction:     entry.Direction,
			Amount:        entry.Amount,
			BalanceBefore: before,
			BalanceAfter:  after,
			Memo:          strings.TrimSpace(entry.Memo),
			CreatedAt:     now,
		}
		if err := tx.Create(&entryRow).Error; err != nil {
			return TransactionResult{}, err
		}
		if err := tx.Create(&ledgerBalanceSnapshot{
			AccountID:        account.ID,
			TenantID:         account.TenantID,
			SiteID:           account.SiteID,
			Currency:         account.Currency,
			AccountType:      account.AccountType,
			SnapshotDate:     time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC),
			Balance:          account.Balance,
			AvailableBalance: account.AvailableBalance,
			FrozenBalance:    account.FrozenBalance,
			TransactionID:    row.ID,
			CreatedAt:        now,
		}).Error; err != nil {
			return TransactionResult{}, err
		}
		results = append(results, EntryResult{ID: entryRow.ID, AccountID: account.ID, AccountType: account.AccountType, Direction: entry.Direction, Amount: entry.Amount, BalanceBefore: before, BalanceAfter: after})
	}
	if err := s.audit(tx, command, row.ID, debitTotal); err != nil {
		return TransactionResult{}, err
	}
	if err := s.publishLedgerEvent(tx, command, row, results); err != nil {
		return TransactionResult{}, err
	}
	return TransactionResult{ID: row.ID, TransactionNo: row.TransactionNo, TenantID: row.TenantID, SiteID: row.SiteID, Currency: row.Currency, EventType: row.EventType, DebitAmount: row.DebitAmount, CreditAmount: row.CreditAmount, ReversalOf: row.ReversalOfTransactionID, Entries: results, CreatedAt: row.CreatedAt}, nil
}

func validatePostCommand(command PostCommand) error {
	if err := validateScope(command.Scope, command.TenantID, command.SiteID); err != nil {
		return err
	}
	if NormalizeCurrency(command.Currency) == "" {
		return ErrCurrencyRequired
	}
	if strings.TrimSpace(command.IdempotencyKey) == "" {
		return ErrIdempotencyRequired
	}
	if len(command.Entries) < 2 {
		return ErrUnbalancedTransaction
	}
	for _, entry := range command.Entries {
		if entry.Amount.LessThanOrEqual(decimal.Zero) {
			return ErrInvalidAmount
		}
		if normalizeDirection(entry.Direction) == "" {
			return ErrInvalidDirection
		}
		if entry.AccountID == 0 && normalBalance(entry.AccountType) == "" {
			return ErrInvalidAccountType
		}
	}
	return nil
}

func validateTenantMoneyCommand(command TenantMoneyCommand, platformOnly bool) error {
	if platformOnly && !command.Actor.Platform {
		return ErrPermissionDenied
	}
	if err := validateScope(command.Scope, command.TenantID, command.SiteID); err != nil {
		return err
	}
	if NormalizeCurrency(command.Currency) == "" {
		return ErrCurrencyRequired
	}
	if command.Amount.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidAmount
	}
	if strings.TrimSpace(command.IdempotencyKey) == "" {
		return ErrIdempotencyRequired
	}
	if strings.TrimSpace(command.Reason) == "" {
		return ErrReasonRequired
	}
	return nil
}

func validateCreditLimitCommand(command CreditLimitCommand) error {
	if !command.Actor.Platform {
		return ErrPermissionDenied
	}
	if err := validateScope(command.Scope, command.TenantID, command.SiteID); err != nil {
		return err
	}
	if NormalizeCurrency(command.Currency) == "" {
		return ErrCurrencyRequired
	}
	if command.CreditLimit.IsNegative() {
		return ErrInvalidAmount
	}
	if strings.TrimSpace(command.IdempotencyKey) == "" {
		return ErrIdempotencyRequired
	}
	if strings.TrimSpace(command.Reason) == "" {
		return ErrReasonRequired
	}
	return nil
}

func validateScope(scope tenancy.Scope, tenantID, siteID uint64) error {
	if tenantID == 0 {
		return ErrInvalidScope
	}
	if scope.Platform {
		return nil
	}
	if scope.TenantID != tenantID || scope.TenantID == 0 {
		return ErrInvalidScope
	}
	if scope.SiteID > 0 && siteID > 0 && scope.SiteID != siteID {
		return ErrInvalidScope
	}
	return nil
}

func (s *Service) resolveAccountsTx(tx *gorm.DB, command PostCommand) (map[uint64]ledgerAccount, []EntryCommand, error) {
	resolved := make([]EntryCommand, 0, len(command.Entries))
	accountIDs := map[uint64]bool{}
	for _, entry := range command.Entries {
		entry.Direction = normalizeDirection(entry.Direction)
		entry.AccountType = strings.ToUpper(strings.TrimSpace(entry.AccountType))
		if entry.AccountID == 0 {
			account, err := s.ensureAccountTx(tx, command.TenantID, command.SiteID, NormalizeCurrency(command.Currency), entry.AccountType)
			if err != nil {
				return nil, nil, err
			}
			entry.AccountID = account.ID
		}
		resolved = append(resolved, entry)
		accountIDs[entry.AccountID] = true
	}
	ids := make([]uint64, 0, len(accountIDs))
	for id := range accountIDs {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	var accounts []ledgerAccount
	if err := lockForUpdate(tx).Where("id IN ? AND deleted_at IS NULL", ids).Order("id ASC").Find(&accounts).Error; err != nil {
		return nil, nil, err
	}
	if len(accounts) != len(ids) {
		return nil, nil, ErrInvalidAccountType
	}
	accountByID := make(map[uint64]ledgerAccount, len(accounts))
	for _, account := range accounts {
		if !strings.EqualFold(account.Currency, command.Currency) {
			return nil, nil, ErrCurrencyMismatch
		}
		if account.TenantID != command.TenantID || account.SiteID != command.SiteID {
			return nil, nil, ErrAccountScopeMismatch
		}
		accountByID[account.ID] = account
	}
	return accountByID, resolved, nil
}

func (s *Service) ensureAccountTx(tx *gorm.DB, tenantID, siteID uint64, currency, accountType string) (ledgerAccount, error) {
	accountType = strings.ToUpper(strings.TrimSpace(accountType))
	normal := normalBalance(accountType)
	if normal == "" {
		return ledgerAccount{}, ErrInvalidAccountType
	}
	now := time.Now()
	row := ledgerAccount{
		AccountNo:     uniqueNo("LAC"),
		TenantID:      tenantID,
		SiteID:        siteID,
		Currency:      NormalizeCurrency(currency),
		AccountType:   accountType,
		NormalBalance: normal,
		Status:        StatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error; err != nil {
		return ledgerAccount{}, err
	}
	var account ledgerAccount
	if err := lockForUpdate(tx).Where("tenant_id=? AND site_id=? AND currency=? AND account_type=? AND deleted_at IS NULL", tenantID, siteID, NormalizeCurrency(currency), accountType).Take(&account).Error; err != nil {
		return ledgerAccount{}, err
	}
	return account, nil
}

func (s *Service) ensureCreditLimitTx(tx *gorm.DB, command CreditLimitCommand) (tenantCreditLimit, error) {
	now := time.Now()
	row := tenantCreditLimit{
		TenantID:    command.TenantID,
		SiteID:      command.SiteID,
		Currency:    NormalizeCurrency(command.Currency),
		Status:      StatusActive,
		UpdatedBy:   command.Actor.ID,
		CreatedAt:   now,
		UpdatedAt:   now,
		CreditLimit: decimal.Zero,
		UsedCredit:  decimal.Zero,
	}
	if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error; err != nil {
		return tenantCreditLimit{}, err
	}
	var limit tenantCreditLimit
	if err := lockForUpdate(tx).Where("tenant_id=? AND site_id=? AND currency=? AND deleted_at IS NULL", command.TenantID, command.SiteID, NormalizeCurrency(command.Currency)).Take(&limit).Error; err != nil {
		return tenantCreditLimit{}, err
	}
	return limit, nil
}

func applyEntryBalance(account ledgerAccount, direction string, amount decimal.Decimal) (decimal.Decimal, error) {
	after := account.Balance
	if direction == account.NormalBalance {
		after = after.Add(amount)
	} else {
		after = after.Sub(amount)
	}
	if after.IsNegative() {
		if account.AccountType == AccountTenantFrozen {
			return account.Balance, ErrInsufficientFrozen
		}
		return account.Balance, ErrInsufficientBalance
	}
	return after, nil
}

func normalBalance(accountType string) string {
	return accountNormalBalance[strings.ToUpper(strings.TrimSpace(accountType))]
}

func normalizeDirection(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == DirectionDebit || value == DirectionCredit {
		return value
	}
	return ""
}

func (s *Service) transactionByIdempotency(ctx context.Context, key string) (TransactionResult, error) {
	var row ledgerTransaction
	if err := s.db.WithContext(ctx).Where("idempotency_key=? AND deleted_at IS NULL", strings.TrimSpace(key)).Take(&row).Error; err != nil {
		return TransactionResult{}, err
	}
	return s.transactionResult(ctx, row)
}

func (s *Service) transactionByIdempotencyTx(tx *gorm.DB, key string) (TransactionResult, error) {
	var row ledgerTransaction
	if err := lockForUpdate(tx).Where("idempotency_key=? AND deleted_at IS NULL", strings.TrimSpace(key)).Take(&row).Error; err != nil {
		return TransactionResult{}, err
	}
	return s.transactionResultTx(tx, row)
}

func (s *Service) transactionResult(ctx context.Context, row ledgerTransaction) (TransactionResult, error) {
	return s.transactionResultTx(s.db.WithContext(ctx), row)
}

func (s *Service) transactionResultTx(tx *gorm.DB, row ledgerTransaction) (TransactionResult, error) {
	var entries []ledgerEntry
	if err := tx.Where("transaction_id=? AND deleted_at IS NULL", row.ID).Order("entry_no ASC").Find(&entries).Error; err != nil {
		return TransactionResult{}, err
	}
	out := TransactionResult{ID: row.ID, TransactionNo: row.TransactionNo, TenantID: row.TenantID, SiteID: row.SiteID, Currency: row.Currency, EventType: row.EventType, DebitAmount: row.DebitAmount, CreditAmount: row.CreditAmount, ReversalOf: row.ReversalOfTransactionID, CreatedAt: row.CreatedAt}
	for _, entry := range entries {
		out.Entries = append(out.Entries, EntryResult{ID: entry.ID, AccountID: entry.AccountID, AccountType: entry.AccountType, Direction: entry.Direction, Amount: entry.Amount, BalanceBefore: entry.BalanceBefore, BalanceAfter: entry.BalanceAfter})
	}
	return out, nil
}

func accountBalance(account ledgerAccount) AccountBalance {
	return AccountBalance{ID: account.ID, AccountNo: account.AccountNo, TenantID: account.TenantID, SiteID: account.SiteID, Currency: account.Currency, AccountType: account.AccountType, NormalBalance: account.NormalBalance, Balance: account.Balance, AvailableBalance: account.AvailableBalance, FrozenBalance: account.FrozenBalance, Status: account.Status, UpdatedAt: account.UpdatedAt}
}

func lockForUpdate(db *gorm.DB) *gorm.DB {
	if db.Dialector != nil && db.Dialector.Name() == "sqlite" {
		return db
	}
	return db.Clauses(clause.Locking{Strength: "UPDATE"})
}

func enableLedgerPosting(tx *gorm.DB) error {
	if tx.Dialector == nil || tx.Dialector.Name() != "mysql" {
		return nil
	}
	return tx.Exec("SET @ledger_posting = 1").Error
}

func disableLedgerPosting(tx *gorm.DB) {
	if tx.Dialector != nil && tx.Dialector.Name() == "mysql" {
		_ = tx.Exec("SET @ledger_posting = 0").Error
	}
}

func marshal(value any) string {
	if value == nil {
		return "{}"
	}
	data, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func uniqueNo(prefix string) string {
	return fmt.Sprintf("%s%s", prefix, time.Now().UTC().Format("20060102150405.000000000"))
}

func isDuplicateError(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "duplicate")
}

func (s *Service) audit(tx *gorm.DB, command PostCommand, transactionID uint64, amount decimal.Decimal) error {
	if !tx.Migrator().HasTable("game_admin_operation_logs") {
		return nil
	}
	return tx.Table("game_admin_operation_logs").Create(map[string]any{
		"tenant_id": command.TenantID, "site_id": command.SiteID,
		"operator_id": command.OperatorID, "operator_username": command.OperatorUsername,
		"permission": "system:tenant:ledger:" + strings.ToLower(strings.TrimPrefix(command.EventType, "TENANT_")),
		"module":     "system", "action": "tenant.ledger." + strings.ToLower(command.EventType), "result": "SUCCESS",
		"target_type": "ledger_transaction", "target_id": transactionID,
		"after_state": marshal(map[string]any{"eventType": command.EventType, "businessType": command.BusinessType, "idempotencyKey": command.IdempotencyKey}),
		"reason":      strings.TrimSpace(command.Reason), "amount": amount, "currency": NormalizeCurrency(command.Currency),
		"correlation_id": command.CorrelationID, "created_at": time.Now(), "updated_at": time.Now(),
	}).Error
}

func (s *Service) publishLedgerEvent(tx *gorm.DB, command PostCommand, row ledgerTransaction, entries []EntryResult) error {
	payloadEntries := make([]map[string]any, 0, len(entries))
	for _, entry := range entries {
		payloadEntries = append(payloadEntries, map[string]any{
			"accountID": entry.AccountID, "accountType": entry.AccountType, "direction": entry.Direction,
			"amount": entry.Amount.StringFixed(2), "balanceBefore": entry.BalanceBefore.StringFixed(2), "balanceAfter": entry.BalanceAfter.StringFixed(2),
		})
	}
	payload := map[string]any{
		"transactionID": row.ID, "transactionNo": row.TransactionNo, "ledgerEventType": row.EventType, "businessEventType": ledgerBusEventType(row),
		"businessType": row.BusinessType, "businessOrderNo": row.BusinessOrderNo,
		"currency": row.Currency, "debitAmount": row.DebitAmount.StringFixed(2), "creditAmount": row.CreditAmount.StringFixed(2),
		"reversalOfTransactionID": row.ReversalOfTransactionID, "entries": payloadEntries,
	}
	metadata := map[string]any{"operatorID": row.OperatorID, "operatorUsername": row.OperatorUsername, "approvalID": row.ApprovalID}
	traceID := command.CorrelationID
	bus := eventbus.New(tx)
	baseEvent, err := bus.PublishTx(tx, eventbus.Event{
		Scope:           command.Scope,
		TenantID:        row.TenantID,
		SiteID:          row.SiteID,
		EventType:       eventbus.EventLedgerTransactionPosted,
		AggregateType:   "ledger_transaction",
		AggregateID:     fmt.Sprint(row.ID),
		Source:          "tenantledger",
		IdempotencyKey:  "ledger:" + row.TransactionNo + ":posted",
		TraceID:         traceID,
		CorrelationID:   command.CorrelationID,
		TargetConsumers: []string{eventbus.ConsumerLedgerAudit},
		Payload:         payload,
		Metadata:        metadata,
	})
	if err != nil {
		return err
	}
	businessEventType := ledgerBusEventType(row)
	if businessEventType == "" || businessEventType == eventbus.EventLedgerTransactionPosted {
		return nil
	}
	_, err = bus.PublishTx(tx, eventbus.Event{
		Scope:           command.Scope,
		TenantID:        row.TenantID,
		SiteID:          row.SiteID,
		EventType:       businessEventType,
		AggregateType:   "ledger_transaction",
		AggregateID:     fmt.Sprint(row.ID),
		Source:          "tenantledger",
		IdempotencyKey:  "ledger:" + row.TransactionNo + ":" + businessEventType,
		TraceID:         traceID,
		CorrelationID:   command.CorrelationID,
		CausationID:     baseEvent.EventID,
		TargetConsumers: ledgerBusinessTargetConsumers(businessEventType),
		Payload:         payload,
		Metadata:        metadata,
	})
	return err
}

func ledgerBusinessTargetConsumers(eventType string) []string {
	switch eventType {
	case eventbus.EventTenantDepositCredited, eventbus.EventTenantCreditChanged, eventbus.EventTenantFundsFrozen, eventbus.EventTenantFundsUnfrozen:
		return []string{eventbus.ConsumerDashboardDailyStats}
	default:
		return nil
	}
}

func ledgerBusEventType(row ledgerTransaction) string {
	if row.ReversalOfTransactionID != nil || row.EventType == EventReversal {
		return eventbus.EventTenantLedgerReversed
	}
	switch row.EventType {
	case EventTenantDeposit:
		return eventbus.EventTenantDepositCredited
	case EventTenantCredit:
		return eventbus.EventTenantCreditChanged
	case EventTenantFreeze:
		return eventbus.EventTenantFundsFrozen
	case EventTenantUnfreeze:
		return eventbus.EventTenantFundsUnfrozen
	default:
		return eventbus.EventTenantLedgerPosted
	}
}
