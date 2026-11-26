package model

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

// GameTransactionFacts are immutable-enough normalized facts used by provider reconciliation.
type GameTransactionFacts struct {
	sgorm.Model `gorm:"embedded"`

	SiteID                      uint64    `gorm:"column:site_id;type:bigint unsigned;not null" json:"siteID"`
	ProviderID                  uint64    `gorm:"column:provider_id;type:bigint unsigned;not null" json:"providerID"`
	ProviderCode                string    `gorm:"column:provider_code;type:varchar(64);not null" json:"providerCode"`
	ProviderTransactionRowID    uint64    `gorm:"column:provider_transaction_row_id;type:bigint unsigned;not null" json:"providerTransactionRowID"`
	ProviderTransactionID       string    `gorm:"column:provider_transaction_id;type:varchar(128);not null" json:"providerTransactionID"`
	ProviderRoundID             string    `gorm:"column:provider_round_id;type:varchar(128);not null" json:"providerRoundID"`
	ProviderGameID              string    `gorm:"column:provider_game_id;type:varchar(128);not null" json:"providerGameID"`
	UserID                      uint64    `gorm:"column:user_id;type:bigint unsigned;not null" json:"userID"`
	WalletTransactionID         uint64    `gorm:"column:wallet_transaction_id;type:bigint unsigned;not null" json:"walletTransactionID"`
	BetRecordID                 uint64    `gorm:"column:bet_record_id;type:bigint unsigned;not null" json:"betRecordID"`
	RollbackRecordID            uint64    `gorm:"column:rollback_record_id;type:bigint unsigned;not null" json:"rollbackRecordID"`
	Operation                   string    `gorm:"column:operation;type:varchar(32);not null" json:"operation"`
	ProviderStatus              string    `gorm:"column:provider_status;type:varchar(32);not null" json:"providerStatus"`
	BusinessDate                time.Time `gorm:"column:business_date;type:date;not null" json:"businessDate"`
	Currency                    string    `gorm:"column:currency;type:varchar(8);not null" json:"currency"`
	ProviderAmountMinor         int64     `gorm:"column:provider_amount_minor;type:bigint;not null" json:"providerAmountMinor"`
	WalletAmountMinor           int64     `gorm:"column:wallet_amount_minor;type:bigint;not null" json:"walletAmountMinor"`
	BetAmountMinor              int64     `gorm:"column:bet_amount_minor;type:bigint;not null" json:"betAmountMinor"`
	WinAmountMinor              int64     `gorm:"column:win_amount_minor;type:bigint;not null" json:"winAmountMinor"`
	SourceType                  string    `gorm:"column:source_type;type:varchar(64);not null;uniqueIndex:uk_game_transaction_fact_source" json:"sourceType"`
	SourceID                    uint64    `gorm:"column:source_id;type:bigint unsigned;not null;uniqueIndex:uk_game_transaction_fact_source" json:"sourceID"`
	SourceCreatedAt             time.Time `gorm:"column:source_created_at;type:datetime(3);not null" json:"sourceCreatedAt"`
	LastReconciliationRunID     uint64    `gorm:"column:last_reconciliation_run_id;type:bigint unsigned;not null" json:"lastReconciliationRunID"`
	LastReconciliationStatus    string    `gorm:"column:last_reconciliation_status;type:varchar(32);not null" json:"lastReconciliationStatus"`
	LastReconciliationCheckedAt time.Time `gorm:"column:last_reconciliation_checked_at;type:datetime(3)" json:"lastReconciliationCheckedAt"`
}

func (GameTransactionFacts) TableName() string { return "game_transaction_facts" }

type ProviderReconciliationRuns struct {
	sgorm.Model `gorm:"embedded"`

	RunNo string `gorm:"column:run_no;type:varchar(64);not null;uniqueIndex:uk_provider_reconciliation_run_no" json:"runNo"`

	SiteID                   uint64     `gorm:"column:site_id;type:bigint unsigned;not null;uniqueIndex:uk_provider_reconciliation_scope" json:"siteID"`
	ProviderID               uint64     `gorm:"column:provider_id;type:bigint unsigned;not null" json:"providerID"`
	ProviderCode             string     `gorm:"column:provider_code;type:varchar(64);not null;uniqueIndex:uk_provider_reconciliation_scope" json:"providerCode"`
	BusinessDate             time.Time  `gorm:"column:business_date;type:date;not null;uniqueIndex:uk_provider_reconciliation_scope" json:"businessDate"`
	Currency                 string     `gorm:"column:currency;type:varchar(8);not null;uniqueIndex:uk_provider_reconciliation_scope" json:"currency"`
	SourceMode               string     `gorm:"column:source_mode;type:varchar(32);not null" json:"sourceMode"`
	Status                   string     `gorm:"column:status;type:varchar(32);not null" json:"status"`
	ProviderTransactionCount int64      `gorm:"column:provider_transaction_count;type:bigint;not null" json:"providerTransactionCount"`
	CallbackCount            int64      `gorm:"column:callback_count;type:bigint;not null" json:"callbackCount"`
	WalletTransactionCount   int64      `gorm:"column:wallet_transaction_count;type:bigint;not null" json:"walletTransactionCount"`
	BetRecordCount           int64      `gorm:"column:bet_record_count;type:bigint;not null" json:"betRecordCount"`
	DiscrepancyCount         int64      `gorm:"column:discrepancy_count;type:bigint;not null" json:"discrepancyCount"`
	OpenDiscrepancyCount     int64      `gorm:"column:open_discrepancy_count;type:bigint;not null" json:"openDiscrepancyCount"`
	StartedAt                time.Time  `gorm:"column:started_at;type:datetime(3);not null" json:"startedAt"`
	CompletedAt              *time.Time `gorm:"column:completed_at;type:datetime(3)" json:"completedAt"`
	FailureReason            string     `gorm:"column:failure_reason;type:varchar(500);not null" json:"failureReason"`
	CreatedBy                uint64     `gorm:"column:created_by;type:bigint unsigned;not null" json:"createdBy"`
	CreatedByUsername        string     `gorm:"column:created_by_username;type:varchar(64);not null" json:"createdByUsername"`
}

func (ProviderReconciliationRuns) TableName() string { return "provider_reconciliation_runs" }

type ProviderReconciliationItems struct {
	sgorm.Model `gorm:"embedded"`

	RunID  uint64 `gorm:"column:run_id;type:bigint unsigned;not null" json:"runID"`
	ItemNo string `gorm:"column:item_no;type:varchar(64);not null;uniqueIndex:uk_provider_reconciliation_item_no" json:"itemNo"`

	SiteID                   uint64     `gorm:"column:site_id;type:bigint unsigned;not null" json:"siteID"`
	ProviderID               uint64     `gorm:"column:provider_id;type:bigint unsigned;not null" json:"providerID"`
	ProviderCode             string     `gorm:"column:provider_code;type:varchar(64);not null" json:"providerCode"`
	BusinessDate             time.Time  `gorm:"column:business_date;type:date;not null" json:"businessDate"`
	Currency                 string     `gorm:"column:currency;type:varchar(8);not null" json:"currency"`
	DiscrepancyType          string     `gorm:"column:discrepancy_type;type:varchar(64);not null" json:"discrepancyType"`
	Severity                 string     `gorm:"column:severity;type:varchar(16);not null" json:"severity"`
	Status                   string     `gorm:"column:status;type:varchar(32);not null" json:"status"`
	UserID                   uint64     `gorm:"column:user_id;type:bigint unsigned;not null" json:"userID"`
	ProviderTransactionRowID uint64     `gorm:"column:provider_transaction_row_id;type:bigint unsigned;not null" json:"providerTransactionRowID"`
	ProviderTransactionID    string     `gorm:"column:provider_transaction_id;type:varchar(128);not null" json:"providerTransactionID"`
	ProviderRoundID          string     `gorm:"column:provider_round_id;type:varchar(128);not null" json:"providerRoundID"`
	ProviderGameID           string     `gorm:"column:provider_game_id;type:varchar(128);not null" json:"providerGameID"`
	CallbackLogID            uint64     `gorm:"column:callback_log_id;type:bigint unsigned;not null" json:"callbackLogID"`
	WalletTransactionID      uint64     `gorm:"column:wallet_transaction_id;type:bigint unsigned;not null" json:"walletTransactionID"`
	BetRecordID              uint64     `gorm:"column:bet_record_id;type:bigint unsigned;not null" json:"betRecordID"`
	RollbackRecordID         uint64     `gorm:"column:rollback_record_id;type:bigint unsigned;not null" json:"rollbackRecordID"`
	ExpectedAmountMinor      int64      `gorm:"column:expected_amount_minor;type:bigint;not null" json:"expectedAmountMinor"`
	ActualAmountMinor        int64      `gorm:"column:actual_amount_minor;type:bigint;not null" json:"actualAmountMinor"`
	DifferenceAmountMinor    int64      `gorm:"column:difference_amount_minor;type:bigint;not null" json:"differenceAmountMinor"`
	Detail                   string     `gorm:"column:detail;type:json" json:"detail"`
	ResolvedBy               uint64     `gorm:"column:resolved_by;type:bigint unsigned;not null" json:"resolvedBy"`
	ResolvedByUsername       string     `gorm:"column:resolved_by_username;type:varchar(64);not null" json:"resolvedByUsername"`
	ResolvedAt               *time.Time `gorm:"column:resolved_at;type:datetime(3)" json:"resolvedAt"`
	ResolutionReason         string     `gorm:"column:resolution_reason;type:varchar(500);not null" json:"resolutionReason"`
}

func (ProviderReconciliationItems) TableName() string { return "provider_reconciliation_items" }
