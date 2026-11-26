package seed

import (
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/config"
	"casino/internal/game/provider/softaggregator"
	"casino/internal/model"
)

// SeedH5 inserts baseline H5 configuration data into MySQL.
func SeedH5(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := seedGameCatalog(tx); err != nil {
			return err
		}
		if err := seedDepositConfig(tx); err != nil {
			return err
		}
		if err := seedHome(tx); err != nil {
			return err
		}
		if err := seedActivities(tx); err != nil {
			return err
		}
		if err := seedVIP(tx); err != nil {
			return err
		}
		if err := seedSupport(tx); err != nil {
			return err
		}
		if err := seedShare(tx); err != nil {
			return err
		}
		if err := seedCMS(tx); err != nil {
			return err
		}
		return seedApp(tx)
	})
}

type gameSeed struct {
	Code       string
	Name       string
	Provider   string
	Cover      string
	Categories []string
	Players    string
	Badge      string
	Status     int
}

type activitySeed struct {
	Code       string
	Slug       string
	Kind       string
	Title      string
	Subtitle   string
	Reward     string
	MaxReward  string
	Note       string
	CTA        string
	Hero       string
	Card       string
	Images     map[string]string
	Rows       []activityRowSeed
	RulesTitle string
	Rules      []string
}

type activityRowSeed struct {
	Label string
	Value string
	Extra string
}

func seedGameCatalog(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.GameGameCategories{},
		&model.GameGameCategoryRelations{},
		&model.GameProviderCategoryMappings{},
		&model.GameGames{},
	); err != nil {
		return err
	}

	categories := []model.GameGameCategories{
		{CategoryKey: "hot", Label: "Hot", ShortLabel: "H", IconURL: asset("category-hot.webp"), Sort: 1, Status: 1},
		{CategoryKey: "slots", Label: "Slots", ShortLabel: "S", IconURL: asset("category-slot.webp"), Sort: 10, Status: 1},
		{CategoryKey: "live-casino", Label: "Live Casino", ShortLabel: "LC", IconURL: asset("category-casino.webp"), Sort: 20, Status: 1},
		{CategoryKey: "table-games", Label: "Table Games", ShortLabel: "TG", IconURL: asset("category-casino.webp"), Sort: 30, Status: 1},
		{CategoryKey: "crash", Label: "Crash", ShortLabel: "CR", IconURL: asset("category-hot.webp"), Sort: 40, Status: 1},
		{CategoryKey: "fishing", Label: "Fishing", ShortLabel: "F", IconURL: asset("category-casino.webp"), Sort: 50, Status: 1},
		{CategoryKey: "sports", Label: "Sports", ShortLabel: "SP", IconURL: "", Sort: 60, Status: 1},
		{CategoryKey: "lottery", Label: "Lottery", ShortLabel: "L", IconURL: "", Sort: 70, Status: 1},
		{CategoryKey: "arcade", Label: "Arcade", ShortLabel: "A", IconURL: "", Sort: 80, Status: 1},
	}
	if err := upsert(db, &categories, []string{"category_key"}, []string{"label", "short_label", "icon_url", "sort", "status"}); err != nil {
		return err
	}
	if err := db.Model(&model.GameGameCategories{}).
		Where("category_key IN ?", []string{"chance"}).
		Update("status", 2).Error; err != nil {
		return err
	}
	if err := seedProviderCategoryMappings(db); err != nil {
		return err
	}

	return disableLegacyMockGameCatalog(db)
}

func seedProviderCategoryMappings(db *gorm.DB) error {
	var categories []model.GameGameCategories
	if err := db.Where("category_key IN ?", []string{"slots", "live-casino", "table-games", "crash", "fishing", "sports", "lottery", "arcade"}).
		Find(&categories).Error; err != nil {
		return err
	}
	categoryIDByKey := make(map[string]uint64, len(categories))
	for _, category := range categories {
		categoryIDByKey[category.CategoryKey] = category.ID
	}

	type mappingSeed struct {
		ProviderCategoryCode string
		PlatformCategoryKey  string
	}
	seeds := []mappingSeed{
		{ProviderCategoryCode: "slot", PlatformCategoryKey: "slots"},
		{ProviderCategoryCode: "slots", PlatformCategoryKey: "slots"},
		{ProviderCategoryCode: "video_slots", PlatformCategoryKey: "slots"},
		{ProviderCategoryCode: "pragmaticplaylive", PlatformCategoryKey: "live-casino"},
		{ProviderCategoryCode: "live", PlatformCategoryKey: "live-casino"},
		{ProviderCategoryCode: "live_casino", PlatformCategoryKey: "live-casino"},
		{ProviderCategoryCode: "casino", PlatformCategoryKey: "live-casino"},
		{ProviderCategoryCode: "table", PlatformCategoryKey: "table-games"},
		{ProviderCategoryCode: "table_games", PlatformCategoryKey: "table-games"},
		{ProviderCategoryCode: "card", PlatformCategoryKey: "table-games"},
		{ProviderCategoryCode: "crash", PlatformCategoryKey: "crash"},
		{ProviderCategoryCode: "instant", PlatformCategoryKey: "crash"},
		{ProviderCategoryCode: "fish", PlatformCategoryKey: "fishing"},
		{ProviderCategoryCode: "fishing", PlatformCategoryKey: "fishing"},
		{ProviderCategoryCode: "sport", PlatformCategoryKey: "sports"},
		{ProviderCategoryCode: "sports", PlatformCategoryKey: "sports"},
		{ProviderCategoryCode: "virtual_sports", PlatformCategoryKey: "sports"},
		{ProviderCategoryCode: "lottery", PlatformCategoryKey: "lottery"},
		{ProviderCategoryCode: "lotto", PlatformCategoryKey: "lottery"},
		{ProviderCategoryCode: "arcade", PlatformCategoryKey: "arcade"},
	}

	mappings := make([]model.GameProviderCategoryMappings, 0, len(seeds))
	for _, seed := range seeds {
		categoryID := categoryIDByKey[seed.PlatformCategoryKey]
		if categoryID == 0 {
			continue
		}
		mappings = append(mappings, model.GameProviderCategoryMappings{
			ProviderCode:         softaggregator.Code,
			ProviderCategoryCode: seed.ProviderCategoryCode,
			PlatformCategoryID:   categoryID,
			Status:               1,
		})
	}
	return upsert(db, &mappings, []string{"provider_code", "provider_category_code"}, []string{"platform_category_id", "status"})
}

func disableLegacyMockGameCatalog(db *gorm.DB) error {
	legacyProviderCodes := []string{"hb-sport", "jili", "pg-soft", "jdb", "cq9", "saba", "11n-sport"}

	var providerIDs []uint64
	if err := db.Model(&model.GameGameProviders{}).
		Where("provider_code IN ?", legacyProviderCodes).
		Pluck("id", &providerIDs).Error; err != nil {
		return err
	}

	if len(providerIDs) > 0 {
		if err := db.Model(&model.GameGames{}).
			Where("provider_id IN ?", providerIDs).
			Update("status", 3).Error; err != nil {
			return err
		}
	}

	return db.Model(&model.GameGameProviders{}).
		Where("provider_code IN ?", legacyProviderCodes).
		Update("status", 3).Error
}

func seedDepositConfig(db *gorm.DB) error {
	if err := db.AutoMigrate(&model.GameDepositVoucherRewards{}); err != nil {
		return err
	}

	methods := []model.GamePaymentMethods{
		{MethodCode: "easypaisa", Name: "EasyPaisa", MethodType: "deposit", IconURL: asset("deposit/payment-method.webp"), MinAmount: money("100"), MaxAmount: money("50000"), Sort: 1, Status: 1},
		{MethodCode: "jazzcash", Name: "JazzCash", MethodType: "deposit", IconURL: asset("deposit/payment-method.webp"), MinAmount: money("100"), MaxAmount: money("50000"), Sort: 2, Status: 1},
		{MethodCode: "qr_scan", Name: "QR Scan", MethodType: "deposit", IconURL: asset("deposit/payment-method.webp"), MinAmount: money("100"), MaxAmount: money("50000"), Sort: 3, Status: 1},
	}
	if err := upsert(db, &methods, []string{"method_code"}, []string{"name", "method_type", "icon_url", "min_amount", "max_amount", "sort", "status"}); err != nil {
		return err
	}

	var dbMethods []model.GamePaymentMethods
	if err := db.Where("method_code IN ?", []string{"easypaisa", "jazzcash", "qr_scan"}).Find(&dbMethods).Error; err != nil {
		return err
	}
	methodIDByCode := map[string]uint64{}
	for _, method := range dbMethods {
		methodIDByCode[method.MethodCode] = method.ID
	}

	channels := []model.GamePaymentChannels{
		{MethodID: methodIDByCode["easypaisa"], ChannelCode: "tarspay-ep", Name: "tarspay ep", CountryCode: "PK", CurrencyCode: "PKR", FeeRate: money("0"), MinAmount: money("100"), MaxAmount: money("50000"), Sort: 1, Status: 1},
		{MethodID: methodIDByCode["easypaisa"], ChannelCode: "sodapay-ep", Name: "sodapay ep", CountryCode: "PK", CurrencyCode: "PKR", FeeRate: money("0"), MinAmount: money("100"), MaxAmount: money("50000"), Sort: 2, Status: 1},
	}
	if err := upsert(db, &channels, []string{"tenant_id", "site_id", "country_code", "currency_code", "channel_code"}, []string{"method_id", "name", "fee_rate", "min_amount", "max_amount", "sort", "status"}); err != nil {
		return err
	}
	if err := db.Model(&model.GamePaymentChannels{}).
		Where("channel_code IN ?", []string{"npay-easypaisa", "univepay-easypaisa", "mcbpay-easypaisa"}).
		Update("status", 0).Error; err != nil {
		return err
	}

	amounts := []model.GameDepositAmountOptions{
		{OptionCode: "deposit-100", Amount: money("100"), BonusReward: money("30"), CashReward: money("20"), Sort: 1, Status: 1},
		{OptionCode: "deposit-300", Amount: money("300"), BonusReward: money("100"), CashReward: money("40"), Sort: 2, Status: 1},
		{OptionCode: "deposit-500", Amount: money("500"), BonusReward: money("170"), CashReward: money("80"), Sort: 3, Status: 1},
		{OptionCode: "deposit-1000", Amount: money("1000"), BonusReward: money("300"), CashReward: money("200"), Sort: 4, Status: 1},
		{OptionCode: "deposit-3000", Amount: money("3000"), BonusReward: money("1000"), CashReward: money("500"), Sort: 5, Status: 1},
		{OptionCode: "deposit-5000", Amount: money("5000"), BonusReward: money("1700"), CashReward: money("800"), Sort: 6, Status: 1},
		{OptionCode: "deposit-10000", Amount: money("10000"), BonusReward: money("3500"), CashReward: money("1500"), Sort: 7, Status: 1},
		{OptionCode: "deposit-30000", Amount: money("30000"), BonusReward: money("10000"), CashReward: money("5000"), Sort: 8, Status: 1},
		{OptionCode: "deposit-50000", Amount: money("50000"), BonusReward: money("17000"), CashReward: money("8000"), Sort: 9, Status: 1},
	}
	if err := upsert(db, &amounts, []string{"option_code"}, []string{"amount", "bonus_reward", "cash_reward", "sort", "status"}); err != nil {
		return err
	}

	vouchers := []model.GameDepositVouchers{
		{VoucherCode: "first-deposit-rewards", Title: "First Deposit Rewards", IconURL: asset("deposit/first_deposit_gift.webp"), MinAmount: money("100"), MaxAmount: money("50000"), MaxReward: money("25000"), BonusReward: money("30"), CashReward: money("20"), Sort: 1, Status: 1},
		{VoucherCode: "weekly-deposit-bonus", Title: "Weekly Deposit Bonus", IconURL: asset("deposit/first_deposit_gift.webp"), MinAmount: money("100"), MaxAmount: money("50000"), MaxReward: money("70"), BonusReward: money("0"), CashReward: money("0"), Sort: 2, Status: 1},
	}
	if err := upsert(db, &vouchers, []string{"voucher_code"}, []string{"title", "icon_url", "min_amount", "max_amount", "max_reward", "bonus_reward", "cash_reward", "sort", "status"}); err != nil {
		return err
	}

	var dbVouchers []model.GameDepositVouchers
	if err := db.Where("voucher_code IN ?", []string{"first-deposit-rewards", "weekly-deposit-bonus"}).Find(&dbVouchers).Error; err != nil {
		return err
	}
	voucherIDByCode := map[string]uint64{}
	for _, voucher := range dbVouchers {
		voucherIDByCode[voucher.VoucherCode] = voucher.ID
	}

	rewards := []model.GameDepositVoucherRewards{
		{VoucherID: voucherIDByCode["first-deposit-rewards"], Amount: money("100"), BonusReward: money("30"), CashReward: money("20"), Sort: 1, Status: 1},
		{VoucherID: voucherIDByCode["first-deposit-rewards"], Amount: money("300"), BonusReward: money("100"), CashReward: money("40"), Sort: 2, Status: 1},
		{VoucherID: voucherIDByCode["first-deposit-rewards"], Amount: money("500"), BonusReward: money("170"), CashReward: money("80"), Sort: 3, Status: 1},
		{VoucherID: voucherIDByCode["first-deposit-rewards"], Amount: money("1000"), BonusReward: money("300"), CashReward: money("200"), Sort: 4, Status: 1},
		{VoucherID: voucherIDByCode["first-deposit-rewards"], Amount: money("3000"), BonusReward: money("1000"), CashReward: money("500"), Sort: 5, Status: 1},
		{VoucherID: voucherIDByCode["first-deposit-rewards"], Amount: money("5000"), BonusReward: money("1700"), CashReward: money("800"), Sort: 6, Status: 1},
		{VoucherID: voucherIDByCode["first-deposit-rewards"], Amount: money("10000"), BonusReward: money("3500"), CashReward: money("1500"), Sort: 7, Status: 1},
		{VoucherID: voucherIDByCode["first-deposit-rewards"], Amount: money("30000"), BonusReward: money("10000"), CashReward: money("5000"), Sort: 8, Status: 1},
		{VoucherID: voucherIDByCode["first-deposit-rewards"], Amount: money("50000"), BonusReward: money("17000"), CashReward: money("8000"), Sort: 9, Status: 1},
		{VoucherID: voucherIDByCode["weekly-deposit-bonus"], Amount: money("100"), BonusReward: money("0"), CashReward: money("0"), Sort: 1, Status: 1},
		{VoucherID: voucherIDByCode["weekly-deposit-bonus"], Amount: money("300"), BonusReward: money("0"), CashReward: money("0"), Sort: 2, Status: 1},
		{VoucherID: voucherIDByCode["weekly-deposit-bonus"], Amount: money("500"), BonusReward: money("5"), CashReward: money("0"), Sort: 3, Status: 1},
		{VoucherID: voucherIDByCode["weekly-deposit-bonus"], Amount: money("1000"), BonusReward: money("10"), CashReward: money("0"), Sort: 4, Status: 1},
		{VoucherID: voucherIDByCode["weekly-deposit-bonus"], Amount: money("3000"), BonusReward: money("20"), CashReward: money("0"), Sort: 5, Status: 1},
		{VoucherID: voucherIDByCode["weekly-deposit-bonus"], Amount: money("5000"), BonusReward: money("30"), CashReward: money("0"), Sort: 6, Status: 1},
		{VoucherID: voucherIDByCode["weekly-deposit-bonus"], Amount: money("10000"), BonusReward: money("70"), CashReward: money("0"), Sort: 7, Status: 1},
		{VoucherID: voucherIDByCode["weekly-deposit-bonus"], Amount: money("30000"), BonusReward: money("70"), CashReward: money("0"), Sort: 8, Status: 1},
		{VoucherID: voucherIDByCode["weekly-deposit-bonus"], Amount: money("50000"), BonusReward: money("70"), CashReward: money("0"), Sort: 9, Status: 1},
	}
	return upsert(db, &rewards, []string{"voucher_id", "amount"}, []string{"bonus_reward", "cash_reward", "sort", "status"})
}

func seedHome(db *gorm.DB) error {
	banners := []model.GameBanners{
		{BannerCode: "checkin", Position: "home", Title: "Daily Check-in Rewards", ImageURL: asset("banner-checkin.webp"), LinkType: "activity", ActivitySlug: "daily-sign-in", Sort: 1, Status: 1},
		{BannerCode: "red-packet", Position: "home", Title: "Red Packet Rain", ImageURL: asset("banner-red-packet.webp"), LinkType: "activity", ActivitySlug: "red-packet-rain", Sort: 2, Status: 1},
		{BannerCode: "refer", Position: "home", Title: "Refer & Earn", ImageURL: asset("banner-refer.webp"), LinkType: "route", LinkURL: "/share?tab=social", Sort: 3, Status: 1},
		{BannerCode: "app-download", Position: "home", Title: "App Download Bonus", ImageURL: asset("banner-app-download.webp"), LinkType: "activity", ActivitySlug: "app-download", Sort: 4, Status: 1},
		{BannerCode: "recharge", Position: "home", Title: "Recharge Bonus", ImageURL: asset("banner-recharge.webp"), LinkType: "activity", ActivitySlug: "deposit-bonus", Sort: 5, Status: 1},
	}
	if err := upsert(db, &banners, []string{"banner_code"}, []string{"position", "title", "image_url", "link_type", "link_url", "activity_slug", "sort", "status", "start_at", "end_at"}); err != nil {
		return err
	}

	features := []model.GameHomeFeatureCards{
		{FeatureCode: "invite-spin", Title: "Invite Spin", ImageURL: asset("share-invite-spin.webp"), BackgroundStyle: "linear-gradient(232deg,#f9efa5,#faabc8,#e87bf9,#a73ef2,#4f16de)", LinkType: "activity", LinkURL: "/activity/invitation-wheel", Sort: 1, Status: 1},
		{FeatureCode: "invite-chest", Title: "Invite Chest", ImageURL: asset("share-invite-chest.webp"), BackgroundStyle: "linear-gradient(232deg,#fab9ff,#cdaeff,#8d64fe,#6744e6,#242e94)", LinkType: "activity", LinkURL: "/activity/invitation-chest", Sort: 2, Status: 1},
		{FeatureCode: "referrals-rewards", Title: "Referrals Rewards", ImageURL: asset("share-referrals-rewards.webp"), BackgroundStyle: "linear-gradient(232deg,#f6a78d,#f9b371,#faa926,#f0842d,#ed422d)", LinkType: "route", LinkURL: "/share?tab=agent", Sort: 3, Status: 1},
	}
	if err := upsert(db, &features, []string{"feature_code"}, []string{"title", "image_url", "background_style", "link_type", "link_url", "countdown_end_at", "sort", "status"}); err != nil {
		return err
	}

	if err := db.Unscoped().Where("period_type = ?", "weekly").Delete(&model.GameRebateJackpots{}).Error; err != nil {
		return err
	}
	jackpot := model.GameRebateJackpots{PeriodType: "weekly", JackpotAmount: money("500343"), DigitText: "0000500343", Status: 1, PeriodStart: timePtr(time.Now().AddDate(0, 0, -3)), PeriodEnd: timePtr(time.Now().AddDate(0, 0, 4))}
	return db.Create(&jackpot).Error
}

func seedActivities(db *gorm.DB) error {
	activities := []activitySeed{
		{
			Code: "first-deposit", Slug: "first-deposit-bonus", Kind: "deposit", Title: "First Deposit Super Rewards", Subtitle: "Get it on your first deposit", Reward: "+100% Rewards", MaxReward: "Maximum Reward: +200", Note: "First deposit only", CTA: "Deposit Now", Hero: source("activity__first-deposit-bonus__first_deposit_bundle.webp"), Card: asset("offers-first-deposit.png"), RulesTitle: "Reward Details",
			Images: map[string]string{"pageBg": source("activity__first-deposit-bonus__page_bgs.webp"), "bundle": source("activity__first-deposit-bonus__first_deposit_bundle.webp"), "heroBox": source("activity__first-deposit-bonus__hero_box.webp"), "giftBox": source("activity__first-deposit-bonus__gift_box.webp"), "treasure": source("activity__first-deposit-bonus__treasure.webp"), "coin": source("activity__first-deposit-bonus__holographic_coin.webp"), "banner": source("activity__first-deposit-bonus__treasure_banner.webp"), "rulesBanner": source("activity__first-deposit-bonus__rules_banner.webp")},
			Rows:   []activityRowSeed{{"100.00-299.99", "+30.00", "+20.00"}, {"300.00-499.99", "+80.00", "+60.00"}, {"500.00-999.99", "+150.00", "+100.00"}, {"1,000.00-2,999.99", "+300.00", "+200.00"}, {"3,000.00-4,999.99", "+900.00", "+600.00"}, {"5,000.00-9,999.99", "+1,750.00", "+750.00"}, {"10,000.00-29,999.99", "+3,500.00", "+1,500.00"}, {"30,000.00-49,999.99", "+12,000.00", "+3,000.00"}, {"50,000.00+", "+20,000.00", "+5,000.00"}},
			Rules:  []string{"This promotion is available when you make a deposit using the <strong>First Deposit Rewards Coupon</strong>. The coupon can only be used once.", "After using the coupon, you can receive both <strong>Bonus</strong> and <strong>Cash</strong> rewards based on your deposit amount.", "Bonus rewards will be credited to your <strong>Bonus Wallet</strong> and can be unlocked step by step through betting activity.", "Cash rewards will be credited to your <strong>Cash Wallet</strong> and require <strong>3x valid bets</strong> before withdrawal.", "Each device can receive this reward for only 1 account.", "Phone number verification is required before claiming the reward.", "Rewards may be cancelled in cases of duplicate accounts, abnormal activity, or any violation of promotion rules."},
		},
		{
			Code: "deposit-bonus", Slug: "deposit-bonus", Kind: "deposit", Title: "Recharge Rewards", Subtitle: "Get rewards on every deposit", Reward: "Bonus Rewards", MaxReward: "Maximum Reward Single Reward", Note: "Daily recharge bonus", CTA: "Deposit Now", Hero: source("activity__deposit-bonus__daily_deposit_bundle.webp"), Card: asset("offers-recharge.png"), RulesTitle: "Recharge Details",
			Images: map[string]string{"pageBg": source("activity__deposit-bonus__content_bg.webp"), "pile": source("activity__deposit-bonus__coin_pile.webp"), "chest": source("activity__deposit-bonus__treasure_chest.webp"), "coin": source("activity__deposit-bonus__gold_coin.webp"), "bundle": source("activity__deposit-bonus__daily_deposit_bundle.webp"), "rulesBanner": source("activity__deposit-bonus__rules_banner.webp")},
			Rows:   []activityRowSeed{{"Single Reward", "Bonus Rewards", "10,000.00"}, {"Claim Rewards", "Cash Rewards Claimed: 0", "Bonus Rewards Claimed: 0"}, {"100.00-499.99", "+18.00", "+8%"}, {"500.00-999.99", "+58.00", "+10%"}, {"1,000.00-4,999.99", "+128.00", "+12%"}, {"5,000.00+", "+500.00", "+15%"}},
			Rules:  []string{"This promotion is available when you make a qualifying deposit during the event.", "After depositing, you can receive a random Bonus reward based on your deposit tier.", "The actual Bonus amount will be randomly issued within the reward range shown on the promotion page.", "Bonus rewards will be credited to your Bonus Wallet and can be unlocked step by step through betting activity.", "Each device can receive this reward for only 1 account.", "Phone number verification is required before claiming the reward.", "Rewards may be cancelled in cases of duplicate accounts, abnormal activity, or any violation of promotion rules."},
		},
		{
			Code: "daily-deposit-cashback", Slug: "daily-deposit-cashback", Kind: "deposit", Title: "DEPOSIT REBATE EVERYDAY!", Subtitle: "The more you deposit, the more you earn - no requirements!", Reward: "Increase Rebate", MaxReward: "Daily Deposit Rebate", Note: "APP only", CTA: "Increase Rebate Rate", Hero: source("activity__daily-deposit-cashback__hero_main.webp"), Card: asset("offers-daily-rebate.png"), RulesTitle: "Event Rules",
			Images: map[string]string{"pageBg": source("activity__daily-deposit-cashback__page_bg.webp"), "main": source("activity__daily-deposit-cashback__hero_main.webp"), "button": source("activity__daily-deposit-cashback__increase_rebate_btn.webp"), "left": source("activity__daily-deposit-cashback__card_corner_left.webp"), "right": source("activity__daily-deposit-cashback__card_corner_right.webp")},
			Rows:   []activityRowSeed{{"3,000.00+", "0.00%", "0.50%"}, {"5,000.00+", "0.00%", "1.00%"}, {"10,000.00+", "0.00%", "1.50%"}, {"30,000.00+", "0.00%", "2.00%"}, {"50,000.00+", "0.00%", "2.50%"}, {"100,000.00+", "0.00%", "3.00%"}, {"300,000.00+", "0.00%", "3.50%"}, {"500,000.00+", "0.00%", "4.00%"}},
			Rules:  []string{"This promotion is available for qualifying daily deposits during the event.", "You can receive a Bonus rebate based on your total successful deposit amount for the day.", "The higher your daily deposit amount, the higher your Bonus rebate rate.", "Bonus rewards will be credited to your Bonus Wallet and can be unlocked step by step through betting activity.", "This reward is available only in the 56MVP APP. Please claim it from the APP to enjoy the benefit.", "Each device can receive this reward for only 1 account.", "Rewards may be cancelled in cases of duplicate accounts, abnormal activity, or any violation of promotion rules."},
		},
		{Code: "invitation-wheel", Slug: "invitation-wheel", Kind: "spin", Title: "Round 1", Subtitle: "Rs 1,500.00", Reward: "Notice", MaxReward: "Official Benefits", Note: "Earn spins by sharing daily", CTA: "Start Withdraw", Hero: source("activity__invitation-wheel__wheel_5.webp"), Card: asset("offers-invite-spin.png"), RulesTitle: "Event Rules", Images: map[string]string{"top": source("activity__invitation-wheel__top_decor.webp"), "wheel": source("activity__invitation-wheel__wheel_5.webp"), "ring": source("activity__invitation-wheel__wheel_13.webp"), "pointer": source("activity__invitation-wheel__wheel_16.webp"), "pedestal": source("activity__invitation-wheel__spin_pedestal.svg"), "progressCheck": source("activity__invitation-wheel__progress_check.svg")}, Rows: []activityRowSeed{{"1 Valid Invite", "1 Spin", "Random prize"}, {"3 Valid Invites", "3 Spins", "Bonus chance"}, {"5 Valid Invites", "6 Spins", "Chest chance"}, {"10 Valid Invites", "12 Spins", "Big reward"}}, Rules: []string{"You can get spin chances by joining the event, sharing the event, or inviting new members.", "Each valid invited member can help you get more spin chances and speed up your reward progress.", "Complete the required spin progress to unlock the Cash reward.", "Rewards may be reviewed before being issued."}},
		{Code: "invitation-chest", Slug: "invitation-chest", Kind: "chest", Title: "Invite Chest", Subtitle: "Open treasure with your team", Reward: "Treasure Bonus", MaxReward: "Big cash surprise inside", Note: "Invite to open chests", CTA: "Invite Now", Hero: source("activity__invitation-chest__treasure_box.webp"), Card: asset("offers-invite-chest.png"), RulesTitle: "Chest Rules", Images: map[string]string{"stage": source("activity__invitation-chest__stage_bg.webp"), "hostess": source("activity__invitation-chest__hostess.webp"), "treasure": source("activity__invitation-chest__treasure_box.webp"), "blueChest": source("activity__invitation-chest__blue_chest.webp"), "titleFrame": source("activity__invitation-chest__title_frame.webp")}, Rows: []activityRowSeed{{"Share to Get Chests", "Rs 100.00 Chest", "Your friends can receive Rs 10.00 rewards"}, {"How to Get?", "Invite 1 registrations", "Get 0 / Opened 0 / Earned 0.00"}, {"How to Open?", "Deposit >= Rs 500.00", "Valid bets >= Rs 3,000.00"}}, Rules: []string{"Chest progress increases after valid invited users are confirmed.", "Rewards can be claimed after the chest opens.", "Duplicate or risk-controlled invitations do not count."}},
		{Code: "daily-sign-in", Slug: "daily-sign-in", Kind: "checkin", Title: "Daily Check-in Rewards", Subtitle: "Log in daily, boost your rewards", Reward: "7 Day Rewards", MaxReward: "Rewards increase every day", Note: "Check in once per day", CTA: "Check In", Hero: asset("offers-daily-checkin.png"), Card: asset("offers-daily-checkin.png"), RulesTitle: "Check-in Rules", Images: map[string]string{"treasure": source("activity__daily-sign-in__treasure.webp"), "lit": source("activity__daily-sign-in__check_badge_lit.webp"), "dim": source("activity__daily-sign-in__check_badge_dim.webp")}, Rows: []activityRowSeed{{"Day 1", "Rs 1.00", "Bonus"}, {"Day 2", "Rs 2.00", "Bonus"}, {"Day 3", "Rs 5.00", "Bonus"}, {"Day 7", "Rs 30.00", "Cash"}}, Rules: []string{"Each account can check in once per day.", "Continuous sign-in rewards increase with the streak.", "Missed days may reset the consecutive sign-in progress."}},
		{Code: "red-packet-rain", Slug: "red-packet-rain", Kind: "red_packet", Title: "Red Packet Rain", Subtitle: "Catch lucky packets while they fall", Reward: "Lucky Cash Rain", MaxReward: "Random packets every round", Note: "Open during event time", CTA: "Start Catching", Hero: source("activity__red-packet-rain__hero_packet.webp"), Card: asset("offers-red-packet.png"), RulesTitle: "Packet Rules", Images: map[string]string{"heroBg": source("activity__red-packet-rain__hero_bg.webp"), "packet": source("activity__red-packet-rain__hero_packet.webp"), "packets": source("activity__red-packet-rain__decor_packets.webp"), "stage": source("activity__red-packet-rain__stage_bg.webp")}, Rows: []activityRowSeed{{"Small Packet", "Random", "Cash"}, {"Lucky Packet", "Random", "Bonus"}, {"Mega Packet", "Up to 77.00", "Cash"}, {"Round Time", "Limited", "Timer"}}, Rules: []string{"Tap red packets during the active rain round.", "Each packet reward is random and subject to activity status.", "Rewards are issued after the round ends and the claim is confirmed."}},
		{Code: "welfare-card", Slug: "welfare-card", Kind: "card", Title: "Benefit Card", Subtitle: "Unlock exclusive daily benefits", Reward: "Daily Welfare", MaxReward: "Multiple bonuses available", Note: "Limited welfare card", CTA: "Activate Now", Hero: source("activity__welfare-card__hero_main.webp"), Card: asset("offers-benefit-card.png"), RulesTitle: "Benefit Rules", Images: map[string]string{"heroBg": source("activity__welfare-card__hero_bg.webp"), "main": source("activity__welfare-card__hero_main.webp"), "coins": source("activity__welfare-card__hero_coins.webp")}, Rows: []activityRowSeed{{"Weekly Card", "Rs 125.00", "Instant 20.00 / Daily 15.00 / Price 100.00"}, {"Monthly Card", "Rs 420.00", "Instant 60.00 / Daily 12.00 / Price 300.00"}}, Rules: []string{"Benefits refresh after the welfare card is activated.", "Daily benefits must be claimed before reset.", "Card rewards follow the same audit requirements as bonus wallet rewards."}},
		{Code: "sub-deposit-reward", Slug: "sub-deposit-reward", Kind: "deposit", Title: "Invite Deposit Cashback", Subtitle: "Earn when invited users deposit", Reward: "Invite Cashback", MaxReward: "Unlimited commission style bonus", Note: "Invite deposit reward", CTA: "Share Now", Hero: source("activity__sub-deposit-reward__hero_main.webp"), Card: asset("offers-invite-cashback.png"), RulesTitle: "Cashback Rules", Images: map[string]string{"main": source("activity__sub-deposit-reward__hero_main.webp"), "ribbon": source("activity__sub-deposit-reward__ribbon_cycle_range.webp")}, Rows: []activityRowSeed{{"100.00", "0.01%", "Cycle Total Recharge (Direct)"}, {"1,000.00", "0.02%", "Cycle Total Recharge (Direct)"}, {"10,000.00", "0.02%", "Cycle Total Recharge (Direct)"}, {"100,000.00", "0.03%", "Cycle Total Recharge (Direct)"}}, Rules: []string{"Cashback is calculated from invited users successful deposits.", "Settlement data is displayed after platform review.", "Duplicate or risk-controlled accounts do not count."}},
		{Code: "daily-relief", Slug: "daily-relief", Kind: "deposit", Title: "Daily Rescue Cash", Subtitle: "Get rescued after difficult rounds", Reward: "Relief Bonus", MaxReward: "Maximum Rescue: +200", Note: "Daily loss relief", CTA: "Claim Rescue", Hero: source("activity__daily-relief__hero_main.webp"), Card: asset("offers-daily-rescue.png"), RulesTitle: "Rescue Rules", Images: map[string]string{"main": source("activity__daily-relief__hero_main.webp"), "ribbon": source("activity__daily-relief__ribbon_banner.webp")}, Rows: []activityRowSeed{{"100.00+", "0.50%", "Valid bets"}, {"1,000.00+", "1.00%", "Valid bets"}, {"5,000.00+", "1.50%", "Valid bets"}, {"10,000.00+", "2.00%", "Valid bets"}}, Rules: []string{"Daily rescue is calculated from qualifying valid losses.", "The reward can be claimed after settlement.", "Abnormal gameplay may be excluded from the event."}},
		{Code: "rebate-ranking", Slug: "rebate-ranking", Kind: "card", Title: "Cashback Leaderboard", Subtitle: "Rank up for bigger rebates", Reward: "Top Ranking Rewards", MaxReward: "Daily leaderboard settlement", Note: "Rank up for rebate", CTA: "View Ranking", Hero: asset("offers-cashback-leaderboard.png"), Card: asset("offers-cashback-leaderboard.png"), RulesTitle: "Leaderboard Rules", Images: map[string]string{"top1": source("activity__rebate-ranking__ranking-top1-badge.webp"), "top2": source("activity__rebate-ranking__ranking-top2-badge.webp"), "top3": source("activity__rebate-ranking__ranking-top3-badge.webp")}, Rows: []activityRowSeed{{"Top 1", "Rs 888.00", "Cash"}, {"Top 2", "Rs 588.00", "Cash"}, {"Top 3", "Rs 388.00", "Cash"}, {"Top 10", "Rs 88.00", "Cash"}}, Rules: []string{"Ranking is calculated from valid rebate activity during the cycle.", "Rewards are settled after the leaderboard closes.", "Final ranking is subject to platform review."}},
		{Code: "phone-verification", Slug: "phone-verification", Kind: "card", Title: "Phone Verification Bonus", Subtitle: "Secure your account and claim rewards", Reward: "+123.00 Bonus", MaxReward: "Verify once, claim once", Note: "Bind phone to claim", CTA: "Verify Now", Hero: asset("offers-phone-verification.png"), Card: asset("offers-phone-verification.png"), RulesTitle: "Verification Rules", Images: map[string]string{"main": source("bonus__icon_cash_bonus.webp")}, Rows: []activityRowSeed{{"Phone Bind", "Rs 123.00", "Bonus"}, {"Security", "Improved", "Account"}}, Rules: []string{"Each account can receive the verification reward once.", "The phone number must pass platform verification.", "The reward is issued after binding succeeds."}},
		{Code: "app-download", Slug: "app-download", Kind: "card", Title: "Download App Bonus", Subtitle: "For a better experience and enhanced security", Reward: "+77.00 Bonus", MaxReward: "Download and claim bonus", Note: "Install app to claim", CTA: "Download App", Hero: asset("offers-app-download.png"), Card: asset("offers-app-download.png"), RulesTitle: "Download Rules", Images: map[string]string{"main": source("activity__guide-flow-dialogs__bg_glow.webp")}, Rows: []activityRowSeed{{"App Install", "+44.43", "Bonus"}, {"Secure Login", "+10.00", "Extra"}, {"Push Enabled", "+12.00", "Reward"}, {"First Claim", "+77.00", "Max"}}, Rules: []string{"Download the official app and log in with the same account.", "The app reward can be claimed once per eligible account.", "If the device or account is flagged as abnormal, the reward may require review."}},
	}

	rows := make([]model.GameActivityEvents, 0, len(activities))
	for index, activity := range activities {
		rows = append(rows, model.GameActivityEvents{ActivityCode: activity.Code, Slug: activity.Slug, Kind: activity.Kind, Title: activity.Title, Subtitle: activity.Subtitle, RewardText: activity.Reward, MaxRewardText: activity.MaxReward, Note: activity.Note, CtaText: activity.CTA, HeroURL: activity.Hero, CardImageURL: activity.Card, Status: 1, Sort: index + 1})
	}
	if err := upsert(db, &rows, []string{"activity_code"}, []string{"slug", "kind", "title", "subtitle", "reward_text", "max_reward_text", "note", "cta_text", "hero_url", "card_image_url", "status", "sort", "start_at", "end_at"}); err != nil {
		return err
	}

	var dbActivities []model.GameActivityEvents
	if err := db.Where("activity_code IN ?", activityCodes(activities)).Find(&dbActivities).Error; err != nil {
		return err
	}
	byCode := map[string]model.GameActivityEvents{}
	ids := make([]uint64, 0, len(dbActivities))
	for _, activity := range dbActivities {
		byCode[activity.ActivityCode] = activity
		ids = append(ids, activity.ID)
	}
	if len(ids) > 0 {
		if err := db.Unscoped().Where("activity_id IN ?", ids).Delete(&model.GameActivityAssets{}).Error; err != nil {
			return err
		}
		if err := db.Unscoped().Where("activity_id IN ?", ids).Delete(&model.GameActivityRules{}).Error; err != nil {
			return err
		}
		if err := db.Unscoped().Where("activity_id IN ?", ids).Delete(&model.GameActivityRewardRules{}).Error; err != nil {
			return err
		}
	}
	for _, activity := range activities {
		dbActivity := byCode[activity.Code]
		assets := make([]model.GameActivityAssets, 0, len(activity.Images))
		sort := 1
		for key, url := range activity.Images {
			assets = append(assets, model.GameActivityAssets{ActivityID: dbActivity.ID, AssetKey: key, AssetURL: url, Sort: sort})
			sort++
		}
		if len(assets) > 0 {
			if err := db.Create(&assets).Error; err != nil {
				return err
			}
		}
		if err := db.Create(&model.GameActivityRules{ActivityID: dbActivity.ID, Title: activity.RulesTitle, ContentHTML: rulesHTML(activity.Rules), Sort: 1, Status: 1}).Error; err != nil {
			return err
		}
		rewardRules := make([]model.GameActivityRewardRules, 0, len(activity.Rows))
		for index, row := range activity.Rows {
			rewardRules = append(rewardRules, model.GameActivityRewardRules{ActivityID: dbActivity.ID, Label: row.Label, RewardValue: row.Value, BonusText: row.Extra, MinAmount: money("0"), MaxAmount: money("0"), RewardCash: money("0"), RewardBonus: money("0"), RequirementMultiplier: money("1"), Sort: index + 1, Status: 1})
		}
		if len(rewardRules) > 0 {
			if err := db.Create(&rewardRules).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func seedVIP(db *gorm.DB) error {
	deposits := []string{"100.00", "500.00", "1,000.00", "3,000.00", "5,000.00", "10,000.00", "20,000.00", "50,000.00", "100,000.00", "300,000.00", "500,000.00", "1,000,000.00", "2,000,000.00", "5,000,000.00", "10,000,000.00"}
	bets := []string{"300.00", "1,500.00", "3,000.00", "10,000.00", "18,000.00", "35,000.00", "80,000.00", "200,000.00", "400,000.00", "1,200,000.00", "2,000,000.00", "4,000,000.00", "8,000,000.00", "20,000,000.00", "40,000,000.00"}
	levels := make([]model.GameVipLevels, 0, len(deposits))
	for i := range deposits {
		level := uint(i + 1)
		levels = append(levels, model.GameVipLevels{LevelNo: level, Name: fmt.Sprintf("VIP %d", level), IconURL: asset(fmt.Sprintf("offers-vip-%d.webp", level)), LevelUpReward: money(fmt.Sprintf("%d", 5*(i+1))), WeeklyReward: money(fmt.Sprintf("%d", i+1)), MonthlyReward: money(fmt.Sprintf("%d", 3*(i+1))), DailyWithdrawCount: level, Status: 1})
	}
	if err := upsert(db, &levels, []string{"level_no"}, []string{"name", "icon_url", "level_up_reward", "weekly_reward", "monthly_reward", "daily_withdraw_count", "status"}); err != nil {
		return err
	}
	var dbLevels []model.GameVipLevels
	if err := db.Where("level_no BETWEEN ? AND ?", 1, len(deposits)).Order("level_no ASC").Find(&dbLevels).Error; err != nil {
		return err
	}
	levelIDs := make([]uint64, 0, len(dbLevels))
	for _, level := range dbLevels {
		levelIDs = append(levelIDs, level.ID)
	}
	if len(levelIDs) > 0 {
		if err := db.Unscoped().Where("vip_level_id IN ?", levelIDs).Delete(&model.GameVipRequirements{}).Error; err != nil {
			return err
		}
	}
	requirements := make([]model.GameVipRequirements, 0, len(dbLevels)*2)
	for _, level := range dbLevels {
		index := int(level.LevelNo) - 1
		requirements = append(requirements,
			model.GameVipRequirements{VipLevelID: level.ID, RequirementType: "level_up", TargetLevelNo: level.LevelNo, PeriodDays: 0, DepositAmount: money(deposits[index]), ValidBetAmount: money(bets[index])},
			model.GameVipRequirements{VipLevelID: level.ID, RequirementType: "retain", TargetLevelNo: level.LevelNo, PeriodDays: 30, DepositAmount: money(deposits[index]), ValidBetAmount: money(bets[index])},
		)
	}
	if err := db.Create(&requirements).Error; err != nil {
		return err
	}
	return upsertConfig(db, "h5.vip.benefits", []map[string]string{
		{"label": "Level Up Reward", "value": "Rs 5.00", "delta": "+5.00", "type": "reward"},
		{"label": "Weekly Reward", "value": "Rs 1.00", "delta": "+1.00", "type": "reward"},
		{"label": "Monthly Reward", "value": "Rs 3.00", "delta": "+3.00", "type": "reward"},
		{"label": "Additional Daily Withdrawal Count", "value": "1", "delta": "+1", "type": "privilege"},
	}, "H5 VIP benefit display rows")
}

func seedSupport(db *gorm.DB) error {
	categories := []model.GameSupportCategories{
		{CategoryCode: "deposit", Name: "Deposit", Sort: 1, Status: 1},
		{CategoryCode: "other", Name: "Other", Sort: 2, Status: 1},
		{CategoryCode: "withdrawal", Name: "Withdrawal", Sort: 3, Status: 1},
	}
	if err := upsert(db, &categories, []string{"category_code"}, []string{"name", "sort", "status"}); err != nil {
		return err
	}
	categoryIDs, err := supportCategoryIDMap(db)
	if err != nil {
		return err
	}
	if err := db.Unscoped().Where("category_id IN ?", mapValues(categoryIDs)).Delete(&model.GameSupportQuestions{}).Error; err != nil {
		return err
	}
	type faq struct{ cat, q, a string }
	faqs := []faq{
		{"deposit", "How do I make a deposit?", "<p>Go to <strong>Profile &gt; Deposit</strong>, add your bank account, then choose the amount and payment method to complete your deposit.</p><p><strong>Minimum Deposit:</strong> 100 PKR</p><p><strong>Maximum Deposit (per transaction):</strong> 50,000 PKR</p>"},
		{"other", "How can I become an agent?", "<p>Tap the <strong>Share</strong> button at the bottom of the game homepage to get your <strong>exclusive referral link</strong> and <strong>QR code</strong>.</p><p>When your friends register through your link and make a deposit, you will earn referral commissions.</p>"},
		{"withdrawal", "How do I withdraw my funds?", "<p>Go to <strong>Profile &gt; Withdrawal</strong>, add your bank account, enter the withdrawal amount, and submit your request.</p><ul><li><p><strong>Minimum Withdrawal:</strong> 100 PKR</p></li><li><p><strong>Maximum Withdrawal:</strong> 50,000 PKR</p></li></ul>"},
		{"deposit", "How long does it take for my deposit to be credited?", "<p>Most deposits are credited within <strong>1-2 minutes</strong>. Some transactions may take up to <strong>24 hours</strong>.</p>"},
		{"withdrawal", "What are the withdrawal requirements?", "<p>Normally, you need to complete the required <strong>valid betting turnover</strong> before making a withdrawal.</p>"},
		{"other", "How can I change my bank account?", "<p>For account security, bank account information cannot be changed directly. Please contact <strong>Online Customer Support</strong>.</p>"},
		{"deposit", "Why did my deposit fail?", "<p>The payment channel may be temporarily busy. Please wait a few minutes and try again.</p>"},
		{"withdrawal", "Why did my withdrawal fail and the funds were returned?", "<p>Your withdrawal was returned by the bank and credited back to your game account. Please verify your bank details and submit again.</p>"},
		{"other", "The game is not working properly.", "<p>Please contact <strong>Online Customer Support</strong> and send a screenshot of the issue.</p>"},
		{"withdrawal", "My withdrawal has been rejected multiple times.", "<p>Please check your account name, bank account number, and bank details. Contact support if the problem continues.</p>"},
		{"deposit", "My deposit has not been credited after 24 hours.", "<p>Please contact support and provide payment receipt, screenshot, or bank transaction record for verification.</p>"},
		{"withdrawal", "Why is my withdrawal still processing?", "<p>Some banks may take longer to process transactions. Please wait or contact support for status.</p>"},
		{"deposit", "I still can’t make a deposit after several attempts.", "<p>Try a different amount or payment method. If the problem continues, send a screenshot to support.</p>"},
		{"withdrawal", "Why can’t I withdraw my funds?", "<p>Your withdrawable balance may be insufficient, turnover may be incomplete, or you may have reached the withdrawal limit.</p>"},
		{"withdrawal", "What does “Withdrawal Returned” mean?", "<p>If the bank cannot complete the transfer, your withdrawal is returned to your game account automatically.</p>"},
		{"withdrawal", "My withdrawal has not been received.", "<p>Check your withdrawal order status. If it has not updated for a long time, contact Online Customer Support.</p>"},
		{"withdrawal", "Why is my withdrawal under review?", "<p>Due to many withdrawal requests, verification may take longer than usual. Thank you for your patience.</p>"},
	}
	rows := make([]model.GameSupportQuestions, 0, len(faqs))
	for index, item := range faqs {
		rows = append(rows, model.GameSupportQuestions{CategoryID: categoryIDs[item.cat], Question: item.q, AnswerHTML: item.a, Sort: index + 1, Status: 1})
	}
	return db.Create(&rows).Error
}

func seedShare(db *gorm.DB) error {
	shareText := "Join me on 56MVP and claim daily rewards. Register here: {url}"
	shareTitle := "Join 56MVP"
	shareDescription := "Invite friends, unlock rewards, and enjoy games together."
	socialLinks := []model.GameSocialLinks{
		{PlatformCode: "facebook", Label: "Facebook", IconURL: asset("social-facebook.webp"), ShareURL: "https://www.facebook.com/sharer/sharer.php?u={url}&quote={text}", ShareTitle: shareTitle, ShareText: shareText, ShareDescription: shareDescription, Sort: 1, Status: 1},
		{PlatformCode: "messenger", Label: "Messenger", IconURL: asset("social-messenger.webp"), ShareURL: "fb-messenger://share?link={url}", ShareTitle: shareTitle, ShareText: shareText, ShareDescription: shareDescription, Sort: 2, Status: 1},
		{PlatformCode: "whatsapp", Label: "WhatsApp", IconURL: asset("social-whatsapp.webp"), ShareURL: "https://wa.me/?text={text}", ShareTitle: shareTitle, ShareText: shareText, ShareDescription: shareDescription, Sort: 3, Status: 1},
		{PlatformCode: "telegram", Label: "Telegram", IconURL: asset("social-telegram.webp"), ShareURL: "https://t.me/share/url?url={url}&text={text}", ShareTitle: shareTitle, ShareText: shareText, ShareDescription: shareDescription, Sort: 4, Status: 1},
		{PlatformCode: "instagram", Label: "Instagram", IconURL: asset("social-instagram.webp"), ShareURL: "https://www.instagram.com/", ShareTitle: shareTitle, ShareText: "Copy my invite link and share it on Instagram: {url}", ShareDescription: shareDescription, Sort: 5, Status: 1},
		{PlatformCode: "tiktok", Label: "TikTok", IconURL: asset("social-tiktok.webp"), ShareURL: "https://www.tiktok.com/", ShareTitle: shareTitle, ShareText: "Copy my invite link and share it on TikTok: {url}", ShareDescription: shareDescription, Sort: 6, Status: 1},
		{PlatformCode: "x", Label: "X", IconURL: asset("social-x.webp"), ShareURL: "https://twitter.com/intent/tweet?url={url}&text={text}", ShareTitle: shareTitle, ShareText: shareText, ShareDescription: shareDescription, Sort: 7, Status: 1},
	}
	if err := upsert(db, &socialLinks, []string{"platform_code"}, []string{"label", "icon_url", "share_url", "share_title", "share_text", "share_description", "sort", "status"}); err != nil {
		return err
	}

	configs := []struct {
		key   string
		value interface{}
		desc  string
	}{
		{"h5.share.actions", []map[string]string{{"label": "Generate Poster", "tone": "violet"}, {"label": "Copy Link", "tone": "blue"}, {"label": "Share Out", "tone": "rose"}}, "H5 share action buttons"},
		{"h5.share.home_entries", []map[string]string{
			{"id": "invite-spin", "title": "Invite Spin", "image": asset("share-invite-spin.webp"), "background": "linear-gradient(232deg,#f9efa5,#faabc8,#e87bf9,#a73ef2,#4f16de)"},
			{"id": "invite-chest", "title": "Invite Chest", "image": asset("share-invite-chest.webp"), "background": "linear-gradient(232deg,#fab9ff,#cdaeff,#8d64fe,#6744e6,#242e94)"},
			{"id": "referrals-rewards", "title": "Referrals Rewards", "image": asset("share-referrals-rewards.webp"), "background": "linear-gradient(232deg,#f6a78d,#f9b371,#faa926,#f0842d,#ed422d)"},
		}, "H5 share entry cards"},
		{"h5.invite.chest_rewards", []string{"Rs 1,500.00", "Rs 3,000.00", "Rs 9,000.00", "Rs 15,000.00", "Rs 30,000.00", "Rs 90,000.00", "Rs 150,000.00", "Rs 300,000.00"}, "Invitation chest reward labels"},
	}
	for _, config := range configs {
		if err := upsertConfig(db, config.key, config.value, config.desc); err != nil {
			return err
		}
	}

	agentRules := []model.GameAgentCommissionRules{
		{LevelNo: 1, MinReferrals: 1, MinValidBet: money("0"), CommissionRate: money("0.0070"), BonusRate: money("0.1500"), Status: 1},
		{LevelNo: 2, MinReferrals: 3, MinValidBet: money("10000"), CommissionRate: money("0.0071"), BonusRate: money("0.1500"), Status: 1},
		{LevelNo: 3, MinReferrals: 8, MinValidBet: money("100000"), CommissionRate: money("0.0073"), BonusRate: money("0.1500"), Status: 1},
		{LevelNo: 4, MinReferrals: 15, MinValidBet: money("3000000"), CommissionRate: money("0.0075"), BonusRate: money("0.1500"), Status: 1},
		{LevelNo: 5, MinReferrals: 21, MinValidBet: money("5000000"), CommissionRate: money("0.0077"), BonusRate: money("0.1500"), Status: 1},
		{LevelNo: 6, MinReferrals: 55, MinValidBet: money("10000000"), CommissionRate: money("0.0080"), BonusRate: money("0.1500"), Status: 1},
		{LevelNo: 7, MinReferrals: 89, MinValidBet: money("30000000"), CommissionRate: money("0.0090"), BonusRate: money("0.1500"), Status: 1},
		{LevelNo: 8, MinReferrals: 110, MinValidBet: money("50000000"), CommissionRate: money("0.0100"), BonusRate: money("0.1500"), Status: 1},
		{LevelNo: 9, MinReferrals: 144, MinValidBet: money("70000000"), CommissionRate: money("0.0115"), BonusRate: money("0.1500"), Status: 1},
		{LevelNo: 10, MinReferrals: 377, MinValidBet: money("150000000"), CommissionRate: money("0.0150"), BonusRate: money("0.1500"), Status: 1},
		{LevelNo: 11, MinReferrals: 777, MinValidBet: money("300000000"), CommissionRate: money("0.0200"), BonusRate: money("0.1500"), Status: 1},
	}
	if err := upsert(db, &agentRules, []string{"level_no"}, []string{"min_referrals", "min_valid_bet", "commission_rate", "bonus_rate", "status"}); err != nil {
		return err
	}

	chests := []model.GameInvitationChestConfigs{
		{ChestCode: "bronze", Title: "Bronze Chest", RequiredInvites: 1, RewardCash: money("10"), RewardBonus: money("0"), IconURL: asset("small_chest_closed.webp"), Sort: 1, Status: 1},
		{ChestCode: "silver", Title: "Silver Chest", RequiredInvites: 3, RewardCash: money("50"), RewardBonus: money("0"), IconURL: asset("large_chest_closed.webp"), Sort: 2, Status: 1},
		{ChestCode: "gold", Title: "Gold Chest", RequiredInvites: 5, RewardCash: money("120"), RewardBonus: money("0"), IconURL: asset("large_chest_closed.webp"), Sort: 3, Status: 1},
		{ChestCode: "mega", Title: "Mega Chest", RequiredInvites: 10, RewardCash: money("300"), RewardBonus: money("0"), IconURL: asset("small_chest_open.webp"), Sort: 4, Status: 1},
	}
	if err := upsert(db, &chests, []string{"chest_code"}, []string{"title", "required_invites", "reward_cash", "reward_bonus", "icon_url", "sort", "status"}); err != nil {
		return err
	}

	prizes := []model.GameInvitationWheelPrizes{
		{PrizeCode: "cash-1500", Label: "1.5k", RewardType: "cash", MinAmount: money("1500"), MaxAmount: money("1500"), Weight: 1, IconURL: source("activity__invitation-wheel__prize_1.webp"), Sort: 1, Status: 1},
		{PrizeCode: "cash-15", Label: "15", RewardType: "cash", MinAmount: money("15"), MaxAmount: money("15"), Weight: 20, IconURL: source("activity__invitation-wheel__prize_2.webp"), Sort: 2, Status: 1},
		{PrizeCode: "cash-001-75", Label: "0.01-75", RewardType: "cash", MinAmount: money("0.01"), MaxAmount: money("75"), Weight: 40, IconURL: source("activity__invitation-wheel__prize_3.webp"), Sort: 3, Status: 1},
		{PrizeCode: "cash-30", Label: "30", RewardType: "cash", MinAmount: money("30"), MaxAmount: money("30"), Weight: 30, IconURL: source("activity__invitation-wheel__prize_4.webp"), Sort: 4, Status: 1},
		{PrizeCode: "cash-01-30", Label: "0.1-30", RewardType: "cash", MinAmount: money("0.1"), MaxAmount: money("30"), Weight: 30, IconURL: source("activity__invitation-wheel__prize_5.webp"), Sort: 5, Status: 1},
		{PrizeCode: "cash-150", Label: "150", RewardType: "cash", MinAmount: money("150"), MaxAmount: money("150"), Weight: 10, IconURL: source("activity__invitation-wheel__prize_6.webp"), Sort: 6, Status: 1},
		{PrizeCode: "cash-11-14k", Label: "11-1.4k", RewardType: "cash", MinAmount: money("11"), MaxAmount: money("1400"), Weight: 5, IconURL: source("activity__invitation-wheel__prize_7.webp"), Sort: 7, Status: 1},
		{PrizeCode: "cash-10-15k", Label: "10-1.5k", RewardType: "cash", MinAmount: money("10"), MaxAmount: money("1500"), Weight: 5, IconURL: source("activity__invitation-wheel__prize_8.webp"), Sort: 8, Status: 1},
	}
	if err := upsert(db, &prizes, []string{"prize_code"}, []string{"label", "reward_type", "min_amount", "max_amount", "weight", "icon_url", "sort", "status"}); err != nil {
		return err
	}

	yesterday := time.Now().AddDate(0, 0, -1)
	period := model.GameRankingPeriods{RankingType: "commission", Title: "This Period's Commission & Rewards", RankingDate: timePtr(yesterday), PeriodStart: timePtr(yesterday.Add(-24 * time.Hour)), PeriodEnd: timePtr(yesterday), Status: 2}
	if err := db.Where("ranking_type = ? AND ranking_date = ?", period.RankingType, yesterday.Format("2006-01-02")).Assign(period).FirstOrCreate(&period).Error; err != nil {
		return err
	}
	if err := db.Unscoped().Where("ranking_period_id = ?", period.ID).Delete(&model.GameRankingRecords{}).Error; err != nil {
		return err
	}
	rankings := []model.GameRankingRecords{
		{RankingPeriodID: period.ID, RankNo: 1, DisplayName: "R******8", AvatarURL: asset("default-avatar-016.webp"), TotalAmount: money("11944.89"), RewardAmount: money("888"), Status: 2},
		{RankingPeriodID: period.ID, RankNo: 2, DisplayName: "Q******9", AvatarURL: asset("default-avatar-011.webp"), TotalAmount: money("11773.49"), RewardAmount: money("588"), Status: 2},
		{RankingPeriodID: period.ID, RankNo: 3, DisplayName: "U******7", AvatarURL: asset("default-avatar-013.webp"), TotalAmount: money("11248.34"), RewardAmount: money("388"), Status: 2},
	}
	for i, name := range []string{"U******4", "H******4", "E******5", "Y******4", "M******0", "C******4", "Y******0"} {
		rankings = append(rankings, model.GameRankingRecords{RankingPeriodID: period.ID, RankNo: uint(i + 4), DisplayName: name, AvatarURL: asset("default-avatar-018.webp"), TotalAmount: money(fmt.Sprintf("%d.25", 11208-(i*370))), RewardAmount: money("88"), Status: 2})
	}
	return db.Create(&rankings).Error
}

func seedCMS(db *gorm.DB) error {
	pages := []model.GameCmsPages{
		{PageCode: "about-us", Title: "About Us", ContentHTML: "<p>56MVP provides online entertainment services, rich game categories, and 24/7 professional support.</p><p>Please play responsibly and follow local laws.</p>", VersionNo: 1, Status: 1, EffectiveAt: timePtr(time.Now())},
		{PageCode: "user-agreement", Title: "User Agreement", ContentHTML: "<p>Welcome to 56MVP. By using our service, you confirm that you meet the legal age requirement and agree to follow platform rules.</p><p>Do not use the service for illegal activity, cheating, duplicate accounts, or abuse of promotions.</p>", VersionNo: 1, Status: 1, EffectiveAt: timePtr(time.Now())},
		{PageCode: "faq", Title: "FAQ", ContentHTML: "<p>Please open Support to view frequently asked questions or contact Online Customer Support.</p>", VersionNo: 1, Status: 1, EffectiveAt: timePtr(time.Now())},
	}
	return upsert(db, &pages, []string{"page_code", "version_no"}, []string{"title", "content_html", "status", "effective_at"})
}

func seedApp(db *gorm.DB) error {
	app := model.GameAppVersions{Platform: "android", VersionName: "1.0.3", VersionCode: 103, DownloadURL: "https://56mvp.top/", ForceUpdate: 0, ReleaseNote: "H5 seed version", Status: 1}
	return upsert(db, &app, []string{"platform", "version_code"}, []string{"version_name", "download_url", "force_update", "release_note", "status"})
}

func upsert(db *gorm.DB, value interface{}, keys []string, updates []string) error {
	columns := make([]clause.Column, 0, len(keys))
	for _, key := range keys {
		columns = append(columns, clause.Column{Name: key})
	}
	return db.Clauses(clause.OnConflict{
		Columns:   columns,
		DoUpdates: clause.AssignmentColumns(updates),
	}).Create(value).Error
}

func upsertConfig(db *gorm.DB, key string, value interface{}, description string) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	row := model.GameSystemConfigs{ConfigKey: key, ConfigValue: string(payload), Description: description, Status: 1}
	return upsert(db, &row, []string{"config_key"}, []string{"config_value", "description", "status"})
}

func jsonText(value interface{}) string {
	payload, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(payload)
}

func asset(name string) string {
	return staticAsset("h5/images/" + name)
}

func source(name string) string {
	return staticAsset("h5/images/source-activity/" + name)
}

func staticAsset(relative string) string {
	relative = strings.TrimPrefix(strings.ReplaceAll(strings.TrimSpace(relative), "\\", "/"), "/")
	if relative == "" {
		return ""
	}
	settings := config.StorageConfig()
	if strings.EqualFold(strings.TrimSpace(settings.Driver), "oss") {
		prefix := strings.Trim(strings.TrimSpace(settings.OSS.ObjectPrefix), "/")
		if prefix != "" {
			return path.Join(prefix, relative)
		}
		return relative
	}
	return "/static/" + relative
}

func providerCode(name string) string {
	return strings.NewReplacer(" ", "-", ".", "").Replace(strings.ToLower(name))
}

func providerIDMap(db *gorm.DB) (map[string]uint64, error) {
	var providers []model.GameGameProviders
	if err := db.Find(&providers).Error; err != nil {
		return nil, err
	}
	ids := map[string]uint64{}
	for _, provider := range providers {
		ids[provider.ProviderCode] = provider.ID
	}
	return ids, nil
}

func categoryIDMap(db *gorm.DB) (map[string]uint64, error) {
	var categories []model.GameGameCategories
	if err := db.Find(&categories).Error; err != nil {
		return nil, err
	}
	ids := map[string]uint64{}
	for _, category := range categories {
		ids[category.CategoryKey] = category.ID
	}
	return ids, nil
}

func supportCategoryIDMap(db *gorm.DB) (map[string]uint64, error) {
	var categories []model.GameSupportCategories
	if err := db.Where("category_code IN ?", []string{"deposit", "other", "withdrawal"}).Find(&categories).Error; err != nil {
		return nil, err
	}
	ids := map[string]uint64{}
	for _, category := range categories {
		ids[category.CategoryCode] = category.ID
	}
	return ids, nil
}

func mapValues(values map[string]uint64) []uint64 {
	out := make([]uint64, 0, len(values))
	for _, value := range values {
		out = append(out, value)
	}
	return out
}

func gameCodes(games []gameSeed) []string {
	codes := make([]string, 0, len(games))
	for _, game := range games {
		codes = append(codes, game.Code)
	}
	return codes
}

func activityCodes(activities []activitySeed) []string {
	codes := make([]string, 0, len(activities))
	for _, activity := range activities {
		codes = append(codes, activity.Code)
	}
	return codes
}

func primaryGameType(categories []string) string {
	for _, category := range categories {
		if category == "sports" {
			return "sports"
		}
	}
	for _, category := range categories {
		if category == "slots" {
			return "slot"
		}
	}
	for _, category := range categories {
		if category == "chance" {
			return "chance"
		}
	}
	return "slot"
}

func hotScore(categories []string, index int) int {
	for _, category := range categories {
		if category == "hot" {
			return 1000 - index
		}
	}
	return 0
}

func parsePlayers(value string) uint {
	value = strings.TrimSpace(strings.ToLower(value))
	multiplier := 1.0
	if strings.HasSuffix(value, "k") {
		multiplier = 1000
		value = strings.TrimSuffix(value, "k")
	}
	f, _ := strconv.ParseFloat(value, 64)
	return uint(f * multiplier)
}

var moneyCleaner = regexp.MustCompile(`[^0-9.\-]`)

func money(value string) *decimal.Decimal {
	clean := moneyCleaner.ReplaceAllString(strings.ReplaceAll(value, ",", ""), "")
	if clean == "" || clean == "." || clean == "-" {
		clean = "0"
	}
	v, err := decimal.NewFromString(clean)
	if err != nil {
		v = decimal.Zero
	}
	return &v
}

func timePtr(value time.Time) *time.Time {
	v := value
	return &v
}

func rulesHTML(rules []string) string {
	var b strings.Builder
	b.WriteString("<ol>")
	for _, rule := range rules {
		b.WriteString("<li><p>")
		b.WriteString(rule)
		b.WriteString("</p></li>")
	}
	b.WriteString("</ol>")
	return b.String()
}
