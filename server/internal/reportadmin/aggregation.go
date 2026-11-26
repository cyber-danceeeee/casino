package reportadmin

import (
	"fmt"
	"strings"
	"time"

	"casino/internal/tenancy"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RefreshDaily recomputes a business day from source-of-truth tables and upserts
// every summary dimension. Repeating the same day produces the same rows.
func (s *Services) RefreshDaily(date, timezone string) error {
	businessDate, err := time.Parse("2006-01-02", strings.TrimSpace(date))
	if err != nil {
		return NewError("INVALID_BUSINESS_DATE", "invalid business date")
	}
	if _, _, err = timezoneOffset(timezone); err != nil {
		return err
	}
	for _, scope := range s.dailyAggregationScopes() {
		if err := s.refreshDailyScope(businessDate, timezone, scope); err != nil {
			return err
		}
	}
	return nil
}

func (s *Services) refreshDailyScope(businessDate time.Time, timezone string, scope tenancy.Scope) error {
	filters := map[string]string{"dateFrom": businessDate.Format("2006-01-02") + " 00:00:00", "dateTo": businessDate.Format("2006-01-02") + " 23:59:59", "granularity": "day", "timezone": timezone}
	reports := map[string][]map[string]any{}
	for _, reportType := range []string{"FINANCIAL", "GAME", "PROVIDER", "USER", "CHANNEL", "AGENT", "PROMOTION"} {
		rows, e := s.Export.exportRows(reportType, filters, scope)
		if e != nil {
			return e
		}
		reports[reportType] = rows
	}
	return s.Financial.db.Transaction(func(tx *gorm.DB) error {
		for _, table := range dailySummaryTables() {
			where, args := reportScopeSQL(Query{Scope: scope}, "")
			args = append([]any{businessDate.Format("2006-01-02"), timezone}, args...)
			if err := tx.Exec("DELETE FROM "+table+" WHERE business_date=? AND timezone=?"+where, args...).Error; err != nil {
				return err
			}
		}
		for reportType, rows := range reports {
			for _, row := range rows {
				if err := upsertDailyRow(tx, reportType, businessDate, timezone, scope, row); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (s *Services) dailyAggregationScopes() []tenancy.Scope {
	db := s.Financial.db
	if db == nil || !db.Migrator().HasTable("game_saas_sites") {
		return []tenancy.Scope{{Platform: true}}
	}
	rows := []struct {
		TenantID uint64 `gorm:"column:tenant_id"`
		SiteID   uint64 `gorm:"column:site_id"`
	}{}
	_ = db.Table("game_saas_sites").Select("tenant_id,id AS site_id").Where("enabled=1 AND deleted_at IS NULL").Order("tenant_id,id").Scan(&rows).Error
	if len(rows) == 0 {
		return []tenancy.Scope{{Platform: true}}
	}
	scopes := make([]tenancy.Scope, 0, len(rows))
	for _, row := range rows {
		if row.TenantID > 0 {
			scopes = append(scopes, tenancy.Scope{TenantID: row.TenantID, SiteID: row.SiteID})
		}
	}
	if len(scopes) == 0 {
		return []tenancy.Scope{{Platform: true}}
	}
	return scopes
}

func dailySummaryTables() []string {
	return []string{
		"game_daily_financial_summaries",
		"game_daily_game_summaries",
		"game_daily_provider_summaries",
		"game_daily_user_summaries",
		"game_daily_channel_summaries",
		"game_daily_agent_summaries",
		"game_daily_promotion_summaries",
	}
}

func upsertDailyRow(tx *gorm.DB, reportType string, date time.Time, timezone string, scope tenancy.Scope, source map[string]any) error {
	base := map[string]any{"tenant_id": scope.TenantID, "site_id": scope.SiteID, "business_date": date.Format("2006-01-02"), "timezone": timezone, "created_at": time.Now(), "updated_at": time.Now()}
	var table string
	var unique []clause.Column
	var updates []string
	switch reportType {
	case "FINANCIAL":
		table = "game_daily_financial_summaries"
		assign(base, source, map[string]string{"currency": "currency", "deposit_amount": "depositAmount", "withdrawal_amount": "withdrawalAmount", "total_bet": "totalBet", "total_win": "totalWin", "bonus_cost": "bonusCost", "provider_cost": "providerCost", "payment_fee": "paymentFee", "agent_commission": "agentCommission", "adjustments": "adjustments", "active_users": "activeUsers"})
		base["country"], base["payment_channel"], base["user_channel"] = "", "", ""
		unique = columns("tenant_id", "site_id", "business_date", "timezone", "country", "currency", "payment_channel", "user_channel")
		updates = keysExcept(base, "tenant_id", "site_id", "business_date", "timezone", "country", "currency", "payment_channel", "user_channel", "created_at")
	case "GAME":
		table = "game_daily_game_summaries"
		assign(base, source, map[string]string{"game_id": "gameID", "provider_code": "provider", "game_type": "gameType", "currency": "currency", "unique_players": "uniquePlayers", "total_rounds": "rounds", "total_bet": "totalBet", "valid_bet": "validBet", "total_win": "totalWin", "maximum_win": "maximumWin"})
		base["country"], base["device"] = "", ""
		unique = columns("tenant_id", "site_id", "business_date", "timezone", "game_id", "country", "currency", "device")
		updates = keysExcept(base, "tenant_id", "site_id", "business_date", "timezone", "game_id", "country", "currency", "device", "created_at")
	case "PROVIDER":
		table = "game_daily_provider_summaries"
		assign(base, source, map[string]string{"provider_code": "provider", "currency": "currency", "game_count": "gameCount", "unique_players": "uniquePlayers", "total_bet": "totalBet", "total_win": "totalWin", "launch_count": "launchCount", "callback_count": "callbackCount", "callback_duration_ms": "averageCallbackLatency"})
		base["country"], base["game_type"], base["launch_success_count"], base["callback_success_count"] = "", "", 0, 0
		unique = columns("tenant_id", "site_id", "business_date", "timezone", "provider_code", "country", "currency", "game_type")
		updates = keysExcept(base, "tenant_id", "site_id", "business_date", "timezone", "provider_code", "country", "currency", "game_type", "created_at")
	case "USER":
		table = "game_daily_user_summaries"
		assign(base, source, map[string]string{"user_id": "userID", "country": "country", "currency": "currency", "registration_channel": "registrationChannel", "deposit_amount": "totalDeposit", "withdrawal_amount": "totalWithdrawal", "total_bet": "totalBet", "total_win": "totalWin", "bonus_amount": "bonusAmount"})
		base["agent_id"], base["active"] = 0, 1
		unique = columns("tenant_id", "site_id", "business_date", "timezone", "user_id", "currency")
		updates = keysExcept(base, "tenant_id", "site_id", "business_date", "timezone", "user_id", "currency", "created_at")
	case "CHANNEL":
		table = "game_daily_channel_summaries"
		assign(base, source, map[string]string{"channel_id": "channelID", "currency": "currency", "registrations": "registrations", "first_deposit_users": "firstDepositUsers", "depositing_users": "depositingUsers", "deposit_amount": "depositAmount", "betting_users": "bettingUsers", "total_bet": "totalBet"})
		base["agent_id"], base["country"], base["platform"], base["total_win"] = 0, "", "", 0
		unique = columns("tenant_id", "site_id", "business_date", "timezone", "channel_id", "currency")
		updates = keysExcept(base, "tenant_id", "site_id", "business_date", "timezone", "channel_id", "currency", "created_at")
	case "AGENT":
		table = "game_daily_agent_summaries"
		assign(base, source, map[string]string{"agent_id": "agentID", "currency": "currency", "attributed_users": "attributedUsers", "first_deposit_users": "firstDepositUsers", "deposit_amount": "depositAmount", "total_bet": "totalBet", "commission_amount": "commissionAmount", "settlement_amount": "settledAmount"})
		base["country"], base["total_win"] = "", 0
		unique = columns("tenant_id", "site_id", "business_date", "timezone", "agent_id", "currency")
		updates = keysExcept(base, "tenant_id", "site_id", "business_date", "timezone", "agent_id", "currency", "created_at")
	case "PROMOTION":
		table = "game_daily_promotion_summaries"
		assign(base, source, map[string]string{"promotion_id": "promotionID", "currency": "currency", "participating_users": "participatingUsers", "claim_count": "claimCount", "reward_amount": "rewardAmount", "required_wagering": "requiredWagering", "completed_wagering": "completedWagering", "completed_users": "completedUsers", "expired_users": "expiredUsers", "promotion_ggr": "promotionGGR"})
		base["country"] = ""
		unique = columns("tenant_id", "site_id", "business_date", "timezone", "promotion_id", "country", "currency")
		updates = keysExcept(base, "tenant_id", "site_id", "business_date", "timezone", "promotion_id", "country", "currency", "created_at")
	default:
		return NewError("INVALID_REPORT_TYPE", fmt.Sprintf("unsupported report type %s", reportType))
	}
	return tx.Table(table).Clauses(clause.OnConflict{Columns: unique, DoUpdates: clause.AssignmentColumns(updates)}).Create(base).Error
}

func assign(target, source map[string]any, mapping map[string]string) {
	for column, field := range mapping {
		target[column] = source[field]
	}
}
func columns(names ...string) []clause.Column {
	out := make([]clause.Column, len(names))
	for i, name := range names {
		out[i] = clause.Column{Name: name}
	}
	return out
}
func keysExcept(value map[string]any, excluded ...string) []string {
	skip := map[string]bool{}
	for _, key := range excluded {
		skip[key] = true
	}
	out := []string{}
	for key := range value {
		if !skip[key] {
			out = append(out, key)
		}
	}
	return out
}
