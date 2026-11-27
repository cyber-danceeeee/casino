package depositconfig

import (
	"fmt"
	"testing"

	"casino/internal/content"
	"casino/internal/model"

	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMoneyUsesExactDecimal(t *testing.T) {
	value, err := money("0.10", "amount")
	if err != nil {
		t.Fatalf("money returned an error: %v", err)
	}
	if !value.Equal(decimal.RequireFromString("0.10")) {
		t.Fatalf("unexpected amount: %s", value)
	}
	if _, err = money("-0.01", "amount"); err == nil {
		t.Fatal("negative amount must be rejected")
	}
}

func TestRewardTotalAndRangeValidation(t *testing.T) {
	bonus := decimal.RequireFromString("3.25")
	cash := decimal.RequireFromString("1.75")
	if got := rewardTotal(&bonus, &cash); !got.Equal(decimal.RequireFromString("5.00")) {
		t.Fatalf("unexpected reward total: %s", got)
	}
	if err := validRange(decimal.RequireFromString("100"), decimal.RequireFromString("99.99")); err == nil {
		t.Fatal("inverted amount range must be rejected")
	}
}

func TestAmountInRange(t *testing.T) {
	min := decimal.RequireFromString("100")
	max := decimal.RequireFromString("500")
	for _, test := range []struct {
		amount string
		want   bool
	}{
		{"99.99", false},
		{"100", true},
		{"500", true},
		{"500.01", false},
	} {
		if got := amountInRange(decimal.RequireFromString(test.amount), &min, &max); got != test.want {
			t.Fatalf("amountInRange(%s) = %v, want %v", test.amount, got, test.want)
		}
	}
}

func TestListContains(t *testing.T) {
	if !listContains("PK,MY;ALL", "my") {
		t.Fatal("configured item should match case-insensitively")
	}
	if listContains("PK,MY", "US") {
		t.Fatal("unconfigured item must not match")
	}
	if !listContains("", "US") {
		t.Fatal("an empty restriction should allow all values")
	}
}

func TestPaymentChannelsAreFilteredByCountryAndCurrency(t *testing.T) {
	service, methodID := newRegionService(t)
	scope := content.Scope{SiteID: 202}

	pkConfig, err := service.GetConfig(ConfigQuery{Scope: scope, Country: "pk", Currency: "pkr"})
	if err != nil {
		t.Fatalf("get PK config: %v", err)
	}
	pkChannels := pkConfig["channels"].([]map[string]any)
	if len(pkChannels) != 1 || pkChannels[0]["channelCode"] != "easypaisa-pk" {
		t.Fatalf("unexpected PK channels: %#v", pkChannels)
	}
	if pkChannels[0]["countryCode"] != "PK" || pkChannels[0]["currencyCode"] != "PKR" {
		t.Fatalf("channel region is missing from response: %#v", pkChannels[0])
	}

	mxConfig, err := service.GetConfig(ConfigQuery{Scope: scope, Country: "MX", Currency: "MXN"})
	if err != nil {
		t.Fatalf("get MX config: %v", err)
	}
	mxChannels := mxConfig["channels"].([]map[string]any)
	if len(mxChannels) != 1 || mxChannels[0]["channelCode"] != "spei-mx" {
		t.Fatalf("unexpected MX channels: %#v", mxChannels)
	}

	mismatchedConfig, err := service.GetConfig(ConfigQuery{Scope: scope, Country: "PK", Currency: "MXN"})
	if err != nil {
		t.Fatalf("get mismatched config: %v", err)
	}
	if channels := mismatchedConfig["channels"].([]map[string]any); len(channels) != 0 {
		t.Fatalf("mismatched country/currency returned channels: %#v", channels)
	}
	if methods := mismatchedConfig["methods"].([]map[string]any); len(methods) != 0 {
		t.Fatalf("method without a matching regional channel must be hidden: %#v", methods)
	}

	selection, err := service.ResolveSelection(ConfigQuery{
		Scope: scope, Country: "PK", Currency: "PKR", MethodCode: "wallet", Amount: "100",
	}, "easypaisa-pk", "")
	if err != nil {
		t.Fatalf("resolve matching selection: %v", err)
	}
	if selection.Channel.MethodID != methodID || selection.Channel.CountryCode != "PK" {
		t.Fatalf("unexpected selected channel: %#v", selection.Channel)
	}

	_, err = service.ResolveSelection(ConfigQuery{
		Scope: scope, Country: "PK", Currency: "PKR", MethodCode: "wallet", Amount: "100",
	}, "spei-mx", "")
	if err == nil {
		t.Fatal("a channel from another country must not be selectable")
	}
}

func newRegionService(t *testing.T) (*Service, uint64) {
	t.Helper()
	dsn := fmt.Sprintf("file:deposit-region-%p?mode=memory&cache=shared", t)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	statements := []string{
		`CREATE TABLE game_payment_methods (id INTEGER PRIMARY KEY AUTOINCREMENT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME, site_id INTEGER NOT NULL, method_code TEXT NOT NULL, name TEXT NOT NULL, method_type TEXT NOT NULL, icon_url TEXT NOT NULL, min_amount REAL NOT NULL, max_amount REAL NOT NULL, sort INTEGER NOT NULL, status INTEGER NOT NULL)`,

		`CREATE TABLE game_payment_channels (id INTEGER PRIMARY KEY AUTOINCREMENT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME, site_id INTEGER NOT NULL, method_id INTEGER NOT NULL, channel_code TEXT NOT NULL, name TEXT NOT NULL, country_code TEXT NOT NULL, currency_code TEXT NOT NULL, fee_rate REAL NOT NULL, min_amount REAL NOT NULL, max_amount REAL NOT NULL, sort INTEGER NOT NULL, status INTEGER NOT NULL)`,

		`CREATE TABLE game_deposit_amount_options (id INTEGER PRIMARY KEY AUTOINCREMENT, deleted_at DATETIME, site_id INTEGER NOT NULL, currency_code TEXT NOT NULL DEFAULT 'PKR', amount REAL NOT NULL, sort INTEGER NOT NULL, status INTEGER NOT NULL)`,

		`CREATE TABLE game_deposit_vouchers (id INTEGER PRIMARY KEY AUTOINCREMENT, deleted_at DATETIME, site_id INTEGER NOT NULL, start_at DATETIME, end_at DATETIME, sort INTEGER NOT NULL, status INTEGER NOT NULL)`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("create deposit config fixture: %v", err)
		}
	}

	zero := decimal.Zero
	max := decimal.NewFromInt(50000)
	method := map[string]any{
		"site_id": 0, "method_code": "wallet", "name": "Wallet", "method_type": "deposit",
		"icon_url": "", "min_amount": zero, "max_amount": max, "sort": 0, "status": 1,
	}
	if err := db.Table("game_payment_methods").Create(method).Error; err != nil {
		t.Fatalf("create method: %v", err)
	}
	var methodID uint64
	if err := db.Table("game_payment_methods").Select("id").Where("method_code = ?", "wallet").Scan(&methodID).Error; err != nil {
		t.Fatalf("load method id: %v", err)
	}
	fee := decimal.Zero
	channels := []map[string]any{
		{
			"site_id": 0, "method_id": methodID, "channel_code": "easypaisa-pk", "name": "EasyPaisa",
			"country_code": "PK", "currency_code": "PKR", "fee_rate": fee, "min_amount": zero, "max_amount": max,
			"sort": 0, "status": 1,
		},
		{
			"site_id": 0, "method_id": methodID, "channel_code": "spei-mx", "name": "SPEI",
			"country_code": "MX", "currency_code": "MXN", "fee_rate": fee, "min_amount": zero, "max_amount": max,
			"sort": 0, "status": 1,
		},
	}
	if err := db.Table("game_payment_channels").Create(channels).Error; err != nil {
		t.Fatalf("create channels: %v", err)
	}
	return New(db), methodID
}

func TestEmptyPaymentCatalogStillReturnsCurrencyPresetsAndVouchers(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:deposit-empty-%p?mode=memory&cache=shared", t)), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	statements := []string{
		`CREATE TABLE game_payment_methods (id INTEGER PRIMARY KEY AUTOINCREMENT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME, site_id INTEGER, method_code TEXT, name TEXT, method_type TEXT, icon_url TEXT, min_amount NUMERIC, max_amount NUMERIC, sort INTEGER, status INTEGER)`,
		`CREATE TABLE game_payment_channels (id INTEGER PRIMARY KEY AUTOINCREMENT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME, site_id INTEGER, method_id INTEGER, channel_code TEXT, name TEXT, country_code TEXT, currency_code TEXT, fee_rate NUMERIC, min_amount NUMERIC, max_amount NUMERIC, sort INTEGER, status INTEGER)`,
		`CREATE TABLE game_deposit_amount_options (id INTEGER PRIMARY KEY AUTOINCREMENT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME, site_id INTEGER, option_code TEXT, currency_code TEXT, amount NUMERIC, bonus_reward NUMERIC, cash_reward NUMERIC, sort INTEGER, status INTEGER)`,
		`CREATE TABLE game_deposit_vouchers (id INTEGER PRIMARY KEY AUTOINCREMENT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME, site_id INTEGER, voucher_code TEXT, title TEXT, title_zh TEXT, title_en TEXT, description_zh TEXT, description_en TEXT, voucher_type TEXT, icon_url TEXT, min_amount NUMERIC, max_amount NUMERIC, max_reward NUMERIC, bonus_reward NUMERIC, cash_reward NUMERIC, sort INTEGER, status INTEGER, start_at DATETIME, end_at DATETIME, new_users_only INTEGER, first_deposit_only INTEGER, daily_limit INTEGER, weekly_limit INTEGER, per_user_limit INTEGER, countries TEXT, currencies TEXT, method_codes TEXT)`,
		`CREATE TABLE game_deposit_voucher_rewards (id INTEGER PRIMARY KEY AUTOINCREMENT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME, site_id INTEGER, voucher_id INTEGER, amount_option_id INTEGER, amount NUMERIC, bonus_reward NUMERIC, cash_reward NUMERIC, reward_amount NUMERIC, sort INTEGER, status INTEGER)`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	service := New(db)
	amount, zero := decimal.NewFromInt(10000), decimal.Zero
	for _, currency := range []string{"IDR", "PKR"} {
		option := model.GameDepositAmountOptions{SiteID: 1, OptionCode: "preset-" + currency, CurrencyCode: currency, Amount: &amount, BonusReward: &zero, CashReward: &zero, Status: 1}
		if err := db.Create(&option).Error; err != nil {
			t.Fatal(err)
		}
	}
	voucher := model.GameDepositVouchers{SiteID: 1, VoucherCode: "operator-voucher", TitleEn: "Deposit Offer", Currencies: "IDR", MinAmount: &zero, MaxAmount: &amount, MaxReward: &zero, BonusReward: &zero, CashReward: &zero, Status: 1}
	if err := db.Create(&voucher).Error; err != nil {
		t.Fatal(err)
	}
	wrongCurrency := voucher
	wrongCurrency.ID = 0
	wrongCurrency.VoucherCode = "pkr-voucher"
	wrongCurrency.Currencies = "PKR"
	if err := db.Create(&wrongCurrency).Error; err != nil {
		t.Fatal(err)
	}
	q := ConfigQuery{Scope: content.Scope{SiteID: 1}, Country: "ID", Currency: "idr"}
	config, err := service.GetConfig(q)
	if err != nil {
		t.Fatalf("empty methods must not break deposit config: %v", err)
	}
	for _, key := range []string{"methods", "channels"} {
		if len(config[key].([]map[string]any)) != 0 {
			t.Fatalf("unexpected %s", key)
		}
	}
	options := config["amountOptions"].([]map[string]any)
	if len(options) != 1 || options[0]["currencyCode"] != "IDR" {
		t.Fatalf("currency leakage: %#v", options)
	}
	vouchers := config["vouchers"].([]map[string]any)
	if len(vouchers) != 1 || vouchers[0]["voucherCode"] != "operator-voucher" || vouchers[0]["available"] != false || vouchers[0]["unavailableReason"] == "" {
		t.Fatalf("incomplete voucher must remain visible with explanation: %#v", vouchers)
	}
	tier := model.GameDepositVoucherRewards{VoucherID: voucher.ID, Amount: &amount, BonusReward: &zero, CashReward: &zero, RewardAmount: &zero, Status: 1}
	if err := db.Create(&tier).Error; err != nil {
		t.Fatal(err)
	}
	config, err = service.GetConfig(q)
	if err != nil {
		t.Fatal(err)
	}
	vouchers = config["vouchers"].([]map[string]any)
	if len(vouchers) != 1 || vouchers[0]["available"] != true || len(vouchers[0]["rewards"].([]map[string]any)) != 1 {
		t.Fatalf("configured voucher must become available: %#v", vouchers)
	}
	// The same nominal PKR amount must never grant a reward to an IDR order.
	if err := db.Model(&model.GameDepositAmountOptions{}).Where("currency_code = ?", "PKR").Update("bonus_reward", 900).Error; err != nil {
		t.Fatal(err)
	}
	method := model.GamePaymentMethods{SiteID: 1, MethodCode: "qris", Name: "QRIS", MethodType: "deposit", Status: 1, MinAmount: &zero, MaxAmount: &amount}
	if err := db.Create(&method).Error; err != nil {
		t.Fatal(err)
	}
	channel := model.GamePaymentChannels{SiteID: 1, MethodID: method.ID, ChannelCode: "fixture-idr", Name: "Fixture", CountryCode: "ID", CurrencyCode: "IDR", FeeRate: &zero, MinAmount: &zero, MaxAmount: &amount, Status: 1}
	if err := db.Create(&channel).Error; err != nil {
		t.Fatal(err)
	}
	q.Currency = "IDR"
	q.MethodCode = "qris"
	q.Amount = "10000"
	selection, err := service.ResolveSelection(q, "fixture-idr", "")
	if err != nil {
		t.Fatal(err)
	}
	if !selection.RewardAmount.IsZero() {
		t.Fatalf("foreign currency reward leaked: %s", selection.RewardAmount)
	}

}
