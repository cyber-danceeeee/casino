package model

import "time"

// ProviderWalletHistory serializes the lifecycle of a single seamless-wallet
// history. A cancellation may arrive before a bet or after final settlement.
type ProviderWalletHistory struct {
	ID                uint64 `gorm:"primaryKey"`
	SiteID            uint64
	ProviderCode      string `gorm:"uniqueIndex:provider_history,priority:1;size:64"`
	UserID            uint64 `gorm:"uniqueIndex:provider_history,priority:2"`
	ProviderHistoryID string `gorm:"uniqueIndex:provider_history,priority:3;size:32"`
	ProviderRoundID   string `gorm:"size:128"`
	ProviderGameID    string `gorm:"size:128"`
	Status            string `gorm:"size:16"`
	BetMinor          int64
	WinMinor          int64
	TransactionCount  uint
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (ProviderWalletHistory) TableName() string { return "game_provider_wallet_histories" }

// ProviderWalletReceipt enforces transactionCode uniqueness across users and
// operation types, independently of the older operation-specific ledger key.
type ProviderWalletReceipt struct {
	ProviderCode      string `gorm:"primaryKey;size:64"`
	TransactionCode   string `gorm:"primaryKey;size:128"`
	UserID            uint64
	ProviderHistoryID string `gorm:"size:32"`
	CreatedAt         time.Time
}

func (ProviderWalletReceipt) TableName() string { return "game_provider_wallet_receipts" }
