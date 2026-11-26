package tenantledger

import (
	"time"

	"github.com/shopspring/decimal"
)

type ledgerAccount struct {
	ID               uint64          `gorm:"column:id;primaryKey" json:"id"`
	AccountNo        string          `gorm:"column:account_no;uniqueIndex:uk_ledger_account_no" json:"accountNo"`
	TenantID         uint64          `gorm:"column:tenant_id;uniqueIndex:uk_ledger_account_scope_type" json:"tenantID"`
	SiteID           uint64          `gorm:"column:site_id;uniqueIndex:uk_ledger_account_scope_type" json:"siteID"`
	Currency         string          `gorm:"column:currency;uniqueIndex:uk_ledger_account_scope_type" json:"currency"`
	AccountType      string          `gorm:"column:account_type;uniqueIndex:uk_ledger_account_scope_type" json:"accountType"`
	NormalBalance    string          `gorm:"column:normal_balance" json:"normalBalance"`
	Balance          decimal.Decimal `gorm:"column:balance;type:decimal(20,2)" json:"balance"`
	AvailableBalance decimal.Decimal `gorm:"column:available_balance;type:decimal(20,2)" json:"availableBalance"`
	FrozenBalance    decimal.Decimal `gorm:"column:frozen_balance;type:decimal(20,2)" json:"frozenBalance"`
	Status           string          `gorm:"column:status" json:"status"`
	Version          uint64          `gorm:"column:version" json:"version"`
	CreatedAt        time.Time       `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt        time.Time       `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt        *time.Time      `gorm:"column:deleted_at" json:"deletedAt"`
}

func (ledgerAccount) TableName() string { return "ledger_accounts" }

type ledgerTransaction struct {
	ID                      uint64          `gorm:"column:id;primaryKey"`
	TransactionNo           string          `gorm:"column:transaction_no;uniqueIndex:uk_ledger_transaction_no"`
	TenantID                uint64          `gorm:"column:tenant_id"`
	SiteID                  uint64          `gorm:"column:site_id"`
	Currency                string          `gorm:"column:currency"`
	EventType               string          `gorm:"column:event_type"`
	BusinessType            string          `gorm:"column:business_type"`
	BusinessOrderNo         string          `gorm:"column:business_order_no"`
	IdempotencyKey          string          `gorm:"column:idempotency_key;uniqueIndex:uk_ledger_transaction_idempotency"`
	Status                  string          `gorm:"column:status"`
	DebitAmount             decimal.Decimal `gorm:"column:debit_amount;type:decimal(20,2)"`
	CreditAmount            decimal.Decimal `gorm:"column:credit_amount;type:decimal(20,2)"`
	ReversalOfTransactionID *uint64         `gorm:"column:reversal_of_transaction_id;uniqueIndex:uk_ledger_transaction_reversal_once"`
	OperatorID              uint64          `gorm:"column:operator_id"`
	OperatorUsername        string          `gorm:"column:operator_username"`
	ApprovalID              uint64          `gorm:"column:approval_id"`
	Reason                  string          `gorm:"column:reason"`
	Metadata                string          `gorm:"column:metadata"`
	CreatedAt               time.Time       `gorm:"column:created_at"`
	UpdatedAt               time.Time       `gorm:"column:updated_at"`
	DeletedAt               *time.Time      `gorm:"column:deleted_at"`
}

func (ledgerTransaction) TableName() string { return "ledger_transactions" }

type ledgerEntry struct {
	ID            uint64          `gorm:"column:id;primaryKey"`
	TransactionID uint64          `gorm:"column:transaction_id;uniqueIndex:uk_ledger_entry_transaction_no"`
	EntryNo       uint16          `gorm:"column:entry_no;uniqueIndex:uk_ledger_entry_transaction_no"`
	AccountID     uint64          `gorm:"column:account_id"`
	TenantID      uint64          `gorm:"column:tenant_id"`
	SiteID        uint64          `gorm:"column:site_id"`
	Currency      string          `gorm:"column:currency"`
	AccountType   string          `gorm:"column:account_type"`
	Direction     string          `gorm:"column:direction"`
	Amount        decimal.Decimal `gorm:"column:amount;type:decimal(20,2)"`
	BalanceBefore decimal.Decimal `gorm:"column:balance_before;type:decimal(20,2)"`
	BalanceAfter  decimal.Decimal `gorm:"column:balance_after;type:decimal(20,2)"`
	Memo          string          `gorm:"column:memo"`
	CreatedAt     time.Time       `gorm:"column:created_at"`
	DeletedAt     *time.Time      `gorm:"column:deleted_at"`
}

func (ledgerEntry) TableName() string { return "ledger_entries" }

type ledgerBalanceSnapshot struct {
	ID               uint64          `gorm:"column:id;primaryKey"`
	AccountID        uint64          `gorm:"column:account_id;uniqueIndex:uk_ledger_snapshot_account_transaction"`
	TenantID         uint64          `gorm:"column:tenant_id"`
	SiteID           uint64          `gorm:"column:site_id"`
	Currency         string          `gorm:"column:currency"`
	AccountType      string          `gorm:"column:account_type"`
	SnapshotDate     time.Time       `gorm:"column:snapshot_date;type:date"`
	Balance          decimal.Decimal `gorm:"column:balance;type:decimal(20,2)"`
	AvailableBalance decimal.Decimal `gorm:"column:available_balance;type:decimal(20,2)"`
	FrozenBalance    decimal.Decimal `gorm:"column:frozen_balance;type:decimal(20,2)"`
	TransactionID    uint64          `gorm:"column:transaction_id;uniqueIndex:uk_ledger_snapshot_account_transaction"`
	CreatedAt        time.Time       `gorm:"column:created_at"`
}

func (ledgerBalanceSnapshot) TableName() string { return "ledger_balance_snapshots" }

type tenantCreditLimit struct {
	ID                uint64          `gorm:"column:id;primaryKey"`
	TenantID          uint64          `gorm:"column:tenant_id;uniqueIndex:uk_tenant_credit_limit_scope"`
	SiteID            uint64          `gorm:"column:site_id;uniqueIndex:uk_tenant_credit_limit_scope"`
	Currency          string          `gorm:"column:currency;uniqueIndex:uk_tenant_credit_limit_scope"`
	CreditLimit       decimal.Decimal `gorm:"column:credit_limit;type:decimal(20,2)"`
	UsedCredit        decimal.Decimal `gorm:"column:used_credit;type:decimal(20,2)"`
	Status            string          `gorm:"column:status"`
	LastTransactionID uint64          `gorm:"column:last_transaction_id"`
	UpdatedBy         uint64          `gorm:"column:updated_by"`
	CreatedAt         time.Time       `gorm:"column:created_at"`
	UpdatedAt         time.Time       `gorm:"column:updated_at"`
	DeletedAt         *time.Time      `gorm:"column:deleted_at"`
}

func (tenantCreditLimit) TableName() string { return "tenant_credit_limits" }
