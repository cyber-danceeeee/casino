package service

import (
	"context"
	"fmt"
	"testing"

	"casino/internal/model"
	"casino/internal/tenancy"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newWalletCurrencyTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`CREATE TABLE game_saas_tenants (id INTEGER PRIMARY KEY, default_currency TEXT, country_code TEXT, deleted_at DATETIME)`,
		`CREATE TABLE game_saas_sites (id INTEGER PRIMARY KEY, tenant_id INTEGER, currency_code TEXT, country_code TEXT, deleted_at DATETIME)`,
		`CREATE TABLE game_country_configs (country_code TEXT, currency_code TEXT, status INTEGER, deleted_at DATETIME)`,
		`CREATE TABLE game_user_wallets (id INTEGER PRIMARY KEY, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME, tenant_id INTEGER, site_id INTEGER, user_id INTEGER, currency TEXT, cash_balance DECIMAL, bonus_balance DECIMAL, withdrawable_balance DECIMAL, frozen_balance DECIMAL, status INTEGER, last_transaction_at DATETIME)`,
		`INSERT INTO game_saas_tenants (id,default_currency,country_code) VALUES (7,'USD','US')`,
		`INSERT INTO game_saas_sites (id,tenant_id,currency_code,country_code) VALUES (70,7,'USD','US')`,
		`INSERT INTO game_country_configs (country_code,currency_code,status) VALUES ('US','USD',1)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func TestReconcileWalletCurrencyAdoptsConfiguredSiteCurrency(t *testing.T) {
	db := newWalletCurrencyTestDB(t)
	zero := decimal.Zero
	wallet := model.GameUserWallets{
		TenantID:            7,
		SiteID:              70,
		UserID:              99,
		Currency:            "PKR",
		CashBalance:         &zero,
		BonusBalance:        &zero,
		WithdrawableBalance: &zero,
		FrozenBalance:       &zero,
		Status:              1,
	}
	wallet.ID = 1
	if err := db.Create(&wallet).Error; err != nil {
		t.Fatal(err)
	}

	ok, err := reconcileWalletCurrency(context.Background(), db, tenancy.Scope{TenantID: 7, SiteID: 70}, &wallet, "USD")
	if err != nil {
		t.Fatal(err)
	}
	if !ok || wallet.Currency != "USD" {
		t.Fatalf("wallet currency was not reconciled: ok=%v currency=%q", ok, wallet.Currency)
	}
	var stored string
	if err := db.Table("game_user_wallets").Select("currency").Where("id=1").Take(&stored).Error; err != nil {
		t.Fatal(err)
	}
	if stored != "USD" {
		t.Fatalf("stored wallet currency = %q, want USD", stored)
	}
}

func TestReconcileWalletCurrencyRejectsUnconfiguredCurrency(t *testing.T) {
	db := newWalletCurrencyTestDB(t)
	wallet := model.GameUserWallets{TenantID: 7, SiteID: 70, UserID: 99, Currency: "PKR"}
	wallet.ID = 1
	if err := db.Create(&wallet).Error; err != nil {
		t.Fatal(err)
	}

	ok, err := reconcileWalletCurrency(context.Background(), db, tenancy.Scope{TenantID: 7, SiteID: 70}, &wallet, "EUR")
	if err != nil {
		t.Fatal(err)
	}
	if ok || wallet.Currency != "PKR" {
		t.Fatalf("unexpected wallet currency reconciliation: ok=%v currency=%q", ok, wallet.Currency)
	}
}
