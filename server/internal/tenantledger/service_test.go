package tenantledger

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"casino/internal/eventbus"
	"casino/internal/tenancy"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestConcurrentIdempotencyPostsOnce(t *testing.T) {
	service, db := newLedgerTestService(t)
	actor := platformActor()
	const workers = 20
	results := make(chan TransactionResult, workers)
	errs := make(chan error, workers)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := service.CreditTenantDeposit(context.Background(), TenantMoneyCommand{
				Scope: tenancy.Scope{Platform: true}, TenantID: 7, SiteID: 3, Currency: "PKR",
				Amount: money("100"), IdempotencyKey: "deposit-once", Reason: "seed tenant deposit", Actor: actor,
			})
			if err != nil {
				errs <- err
				return
			}
			results <- result
		}()
	}
	wg.Wait()
	close(results)
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}
	var first uint64
	for result := range results {
		require.NotZero(t, result.ID)
		if first == 0 {
			first = result.ID
		}
		require.Equal(t, first, result.ID)
	}
	var count int64
	require.NoError(t, db.Model(&ledgerTransaction{}).Where("idempotency_key=?", "deposit-once").Count(&count).Error)
	require.Equal(t, int64(1), count)
	require.NoError(t, db.Table("event_outbox").Where("event_type=? AND aggregate_type=?", eventbus.EventLedgerTransactionPosted, "ledger_transaction").Count(&count).Error)
	require.Equal(t, int64(1), count)
	require.NoError(t, db.Table("event_outbox").Where("event_type=? AND aggregate_type=?", eventbus.EventTenantDepositCredited, "ledger_transaction").Count(&count).Error)
	require.Equal(t, int64(1), count)
	require.Equal(t, money("100"), accountByType(t, db, 7, 3, "PKR", AccountTenantAvailable).Balance)
}

func TestPostTransactionRejectsUnbalancedEntries(t *testing.T) {
	service, _ := newLedgerTestService(t)
	_, err := service.PostTransaction(context.Background(), PostCommand{
		Scope: tenancy.Scope{Platform: true}, TenantID: 7, SiteID: 3, Currency: "PKR",
		EventType: "TEST", BusinessType: "TEST", IdempotencyKey: "unbalanced",
		Entries: []EntryCommand{
			{AccountType: AccountTenantAvailable, Direction: DirectionDebit, Amount: money("100")},
			{AccountType: AccountTenantDeposit, Direction: DirectionCredit, Amount: money("99")},
		},
	})
	requireErrorCode(t, err, "UNBALANCED_TRANSACTION")
}

func TestPostTransactionRejectsCrossTenantAccount(t *testing.T) {
	service, db := newLedgerTestService(t)
	actor := platformActor()
	_, err := service.CreditTenantDeposit(context.Background(), TenantMoneyCommand{
		Scope: tenancy.Scope{Platform: true}, TenantID: 7, SiteID: 3, Currency: "PKR",
		Amount: money("100"), IdempotencyKey: "tenant-a", Reason: "tenant a", Actor: actor,
	})
	require.NoError(t, err)
	_, err = service.CreditTenantDeposit(context.Background(), TenantMoneyCommand{
		Scope: tenancy.Scope{Platform: true}, TenantID: 8, SiteID: 3, Currency: "PKR",
		Amount: money("100"), IdempotencyKey: "tenant-b", Reason: "tenant b", Actor: actor,
	})
	require.NoError(t, err)

	tenantAAccount := accountByType(t, db, 7, 3, "PKR", AccountTenantAvailable)
	tenantBAccount := accountByType(t, db, 8, 3, "PKR", AccountTenantDeposit)
	_, err = service.PostTransaction(context.Background(), PostCommand{
		Scope: tenancy.Scope{Platform: true}, TenantID: 8, SiteID: 3, Currency: "PKR",
		EventType: "TEST", BusinessType: "TEST", IdempotencyKey: "cross-tenant",
		Entries: []EntryCommand{
			{AccountID: tenantAAccount.ID, Direction: DirectionDebit, Amount: money("1")},
			{AccountID: tenantBAccount.ID, Direction: DirectionCredit, Amount: money("1")},
		},
	})
	requireErrorCode(t, err, "ACCOUNT_SCOPE_MISMATCH")
}

func TestPostTransactionRejectsMixedCurrencyAccount(t *testing.T) {
	service, db := newLedgerTestService(t)
	actor := platformActor()
	_, err := service.CreditTenantDeposit(context.Background(), TenantMoneyCommand{
		Scope: tenancy.Scope{Platform: true}, TenantID: 7, SiteID: 3, Currency: "USD",
		Amount: money("100"), IdempotencyKey: "usd", Reason: "usd", Actor: actor,
	})
	require.NoError(t, err)
	_, err = service.CreditTenantDeposit(context.Background(), TenantMoneyCommand{
		Scope: tenancy.Scope{Platform: true}, TenantID: 7, SiteID: 3, Currency: "PKR",
		Amount: money("100"), IdempotencyKey: "pkr", Reason: "pkr", Actor: actor,
	})
	require.NoError(t, err)

	usdAccount := accountByType(t, db, 7, 3, "USD", AccountTenantAvailable)
	pkrAccount := accountByType(t, db, 7, 3, "PKR", AccountTenantDeposit)
	_, err = service.PostTransaction(context.Background(), PostCommand{
		Scope: tenancy.Scope{Platform: true}, TenantID: 7, SiteID: 3, Currency: "PKR",
		EventType: "TEST", BusinessType: "TEST", IdempotencyKey: "mixed-currency",
		Entries: []EntryCommand{
			{AccountID: usdAccount.ID, Direction: DirectionDebit, Amount: money("1")},
			{AccountID: pkrAccount.ID, Direction: DirectionCredit, Amount: money("1")},
		},
	})
	requireErrorCode(t, err, "CURRENCY_MISMATCH")
}

func TestFreezeAndUnfreezeBalanceGuards(t *testing.T) {
	service, db := newLedgerTestService(t)
	actor := platformActor()
	_, err := service.CreditTenantDeposit(context.Background(), TenantMoneyCommand{
		Scope: tenancy.Scope{Platform: true}, TenantID: 7, SiteID: 3, Currency: "PKR",
		Amount: money("100"), IdempotencyKey: "deposit", Reason: "deposit", Actor: actor,
	})
	require.NoError(t, err)

	_, err = service.Freeze(context.Background(), 7, 3, money("150"), "PKR", actor, "freeze-too-much", "freeze too much")
	requireErrorCode(t, err, "INSUFFICIENT_BALANCE")
	require.Equal(t, money("100"), accountByType(t, db, 7, 3, "PKR", AccountTenantAvailable).Balance)
	requireNoAccount(t, db, 7, 3, "PKR", AccountTenantFrozen)

	_, err = service.Freeze(context.Background(), 7, 3, money("40"), "PKR", actor, "freeze-ok", "freeze ok")
	require.NoError(t, err)
	require.Equal(t, money("60"), accountByType(t, db, 7, 3, "PKR", AccountTenantAvailable).Balance)
	require.Equal(t, money("40"), accountByType(t, db, 7, 3, "PKR", AccountTenantFrozen).Balance)

	_, err = service.Unfreeze(context.Background(), 7, 3, money("50"), "PKR", actor, "unfreeze-too-much", "unfreeze too much")
	requireErrorCode(t, err, "INSUFFICIENT_FROZEN_BALANCE")
	require.Equal(t, money("60"), accountByType(t, db, 7, 3, "PKR", AccountTenantAvailable).Balance)
	require.Equal(t, money("40"), accountByType(t, db, 7, 3, "PKR", AccountTenantFrozen).Balance)
}

func TestSetCreditLimitUsesLedgerAndCanReduce(t *testing.T) {
	service, db := newLedgerTestService(t)
	actor := platformActor()
	result, err := service.SetCreditLimit(context.Background(), CreditLimitCommand{
		Scope: tenancy.Scope{Platform: true}, TenantID: 7, SiteID: 3, Currency: "PKR",
		CreditLimit: money("500"), IdempotencyKey: "credit-500", Reason: "set credit", Actor: actor,
	})
	require.NoError(t, err)
	require.NotZero(t, result.ID)
	require.Equal(t, money("500"), accountByType(t, db, 7, 3, "PKR", AccountTenantAvailable).Balance)
	require.Equal(t, money("500"), accountByType(t, db, 7, 3, "PKR", AccountTenantCredit).Balance)

	result, err = service.SetCreditLimit(context.Background(), CreditLimitCommand{
		Scope: tenancy.Scope{Platform: true}, TenantID: 7, SiteID: 3, Currency: "PKR",
		CreditLimit: money("300"), IdempotencyKey: "credit-300", Reason: "reduce credit", Actor: actor,
	})
	require.NoError(t, err)
	require.NotZero(t, result.ID)
	require.Equal(t, money("300"), accountByType(t, db, 7, 3, "PKR", AccountTenantAvailable).Balance)
	require.Equal(t, money("300"), accountByType(t, db, 7, 3, "PKR", AccountTenantCredit).Balance)
}

func TestReverseTransactionOnlyOnceAndDoesNotModifyOriginal(t *testing.T) {
	service, db := newLedgerTestService(t)
	actor := platformActor()
	original, err := service.CreditTenantDeposit(context.Background(), TenantMoneyCommand{
		Scope: tenancy.Scope{Platform: true}, TenantID: 7, SiteID: 3, Currency: "PKR",
		Amount: money("100"), IdempotencyKey: "reverse-me", Reason: "deposit", Actor: actor,
	})
	require.NoError(t, err)

	reversal, err := service.ReverseTransaction(context.Background(), original.ID, "manual reversal", actor)
	require.NoError(t, err)
	require.NotZero(t, reversal.ID)
	require.Equal(t, &original.ID, reversal.ReversalOf)
	require.Equal(t, money("0"), accountByType(t, db, 7, 3, "PKR", AccountTenantAvailable).Balance)

	var storedOriginal ledgerTransaction
	require.NoError(t, db.Take(&storedOriginal, original.ID).Error)
	require.Equal(t, StatusPosted, storedOriginal.Status)
	require.Nil(t, storedOriginal.ReversalOfTransactionID)

	_, err = service.ReverseTransaction(context.Background(), original.ID, "second reversal", actor)
	requireErrorCode(t, err, "TRANSACTION_ALREADY_REVERSED")
}

func TestFailedTransactionLeavesAccountBalanceUnchanged(t *testing.T) {
	service, db := newLedgerTestService(t)
	actor := platformActor()
	_, err := service.CreditTenantDeposit(context.Background(), TenantMoneyCommand{
		Scope: tenancy.Scope{Platform: true}, TenantID: 7, SiteID: 3, Currency: "PKR",
		Amount: money("100"), IdempotencyKey: "rollback-seed", Reason: "seed", Actor: actor,
	})
	require.NoError(t, err)

	_, err = service.PostTransaction(context.Background(), PostCommand{
		Scope: tenancy.Scope{Platform: true}, TenantID: 7, SiteID: 3, Currency: "PKR",
		EventType: "TEST", BusinessType: "TEST", IdempotencyKey: "rollback-fail",
		Entries: []EntryCommand{
			{AccountType: AccountTenantFrozen, Direction: DirectionDebit, Amount: money("120")},
			{AccountType: AccountTenantAvailable, Direction: DirectionCredit, Amount: money("120")},
		},
	})
	requireErrorCode(t, err, "INSUFFICIENT_BALANCE")
	require.Equal(t, money("100"), accountByType(t, db, 7, 3, "PKR", AccountTenantAvailable).Balance)
	requireNoAccount(t, db, 7, 3, "PKR", AccountTenantFrozen)
}

func TestTenantLedgerMigrationGuardsImmutableRowsAndDirectBalanceUpdates(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "migrations", "20260719_tenant_ledger.sql"))
	require.NoError(t, err)
	migration := string(data)

	for _, name := range []string{
		"trg_ledger_accounts_balance_guard",
		"trg_ledger_transactions_no_update",
		"trg_ledger_transactions_no_delete",
		"trg_ledger_entries_no_update",
		"trg_ledger_entries_no_delete",
	} {
		require.Contains(t, migration, name)
	}
	require.Contains(t, migration, "uk_ledger_account_scope_type (tenant_id, site_id, currency, account_type)")
	require.Contains(t, migration, "uk_ledger_transaction_idempotency (idempotency_key)")
	require.Contains(t, migration, "uk_ledger_transaction_reversal_once (reversal_of_transaction_id)")
	require.Contains(t, migration, "chk_ledger_transaction_balanced CHECK (debit_amount = credit_amount)")
}

func newLedgerTestService(t *testing.T) (*Service, *gorm.DB) {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	require.NoError(t, db.AutoMigrate(&ledgerAccount{}, &ledgerTransaction{}, &ledgerEntry{}, &ledgerBalanceSnapshot{}, &tenantCreditLimit{}))
	require.NoError(t, eventbus.Migrate(db))
	return New(db), db
}

func platformActor() Actor {
	return Actor{ID: 1, Username: "platform-admin", Platform: true}
}

func money(value string) decimal.Decimal {
	amount, err := decimal.NewFromString(value)
	if err != nil {
		panic(err)
	}
	return amount
}

func accountByType(t *testing.T, db *gorm.DB, tenantID, siteID uint64, currency, accountType string) ledgerAccount {
	t.Helper()
	var account ledgerAccount
	require.NoError(t, db.Where("tenant_id=? AND site_id=? AND currency=? AND account_type=? AND deleted_at IS NULL", tenantID, siteID, currency, accountType).Take(&account).Error)
	return account
}

func requireNoAccount(t *testing.T, db *gorm.DB, tenantID, siteID uint64, currency, accountType string) {
	t.Helper()
	var account ledgerAccount
	err := db.Where("tenant_id=? AND site_id=? AND currency=? AND account_type=? AND deleted_at IS NULL", tenantID, siteID, currency, accountType).Take(&account).Error
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound), "expected account to be absent, got %v", err)
}

func requireErrorCode(t *testing.T, err error, code string) {
	t.Helper()
	require.Error(t, err)
	require.Equal(t, code, ErrorCode(err))
}
