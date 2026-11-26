package dashboardadmin

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"casino/internal/content"
)

const MoneyFormat = "EXACT_DECIMAL_STRING"

type Filter struct {
	StartTime, EndTime                                     time.Time
	Timezone, Granularity, Country, Currency, ProviderCode string
	CategoryID, ChannelID, AgentID                         uint64
	Metric                                                 string
	Limit                                                  int
	Scope                                                  content.Scope
}

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

func ParseFilter(values map[string]string) (Filter, error) {
	now := time.Now()
	f := Filter{Timezone: "+08:00", Granularity: "day", Currency: "PKR", Metric: "totalBet", Limit: 10}
	var err error
	f.StartTime, err = parseTime(values["startTime"], now.AddDate(0, 0, -6).Format("2006-01-02")+" 00:00:00")
	if err != nil {
		return f, err
	}
	f.EndTime, err = parseTime(values["endTime"], now.Format("2006-01-02")+" 23:59:59")
	if err != nil || f.EndTime.Before(f.StartTime) {
		return f, fmt.Errorf("invalid time range")
	}
	for key, target := range map[string]*string{"timezone": &f.Timezone, "granularity": &f.Granularity, "country": &f.Country, "currency": &f.Currency, "providerCode": &f.ProviderCode, "metric": &f.Metric} {
		if value := strings.TrimSpace(values[key]); value != "" {
			*target = value
		}
	}
	if strings.TrimSpace(f.Currency) == "" {
		return f, fmt.Errorf("currency is required")
	}
	if _, ok := map[string]bool{"hour": true, "day": true, "week": true, "month": true}[f.Granularity]; !ok {
		return f, fmt.Errorf("invalid granularity")
	}
	if _, err = timezoneSeconds(f.Timezone); err != nil {
		return f, err
	}
	f.CategoryID, _ = strconv.ParseUint(values["categoryId"], 10, 64)
	f.ChannelID, _ = strconv.ParseUint(values["channelId"], 10, 64)
	f.AgentID, _ = strconv.ParseUint(values["agentId"], 10, 64)
	if n, e := strconv.Atoi(values["limit"]); e == nil && (n == 10 || n == 20) {
		f.Limit = n
	}
	return f, nil
}

func parseTime(value string, fallback string) (time.Time, error) {
	if strings.TrimSpace(value) == "" {
		value = fallback
	}
	for _, layout := range []string{"2006-01-02 15:04:05", time.RFC3339, "2006-01-02"} {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time")
}

func timezoneSeconds(value string) (int, error) {
	if len(value) != 6 || (value[0] != '+' && value[0] != '-') || value[3] != ':' {
		return 0, fmt.Errorf("invalid timezone")
	}
	h, e1 := strconv.Atoi(value[1:3])
	m, e2 := strconv.Atoi(value[4:6])
	if e1 != nil || e2 != nil || h > 14 || m > 59 {
		return 0, fmt.Errorf("invalid timezone")
	}
	seconds := h*3600 + m*60
	if value[0] == '-' {
		seconds = -seconds
	}
	return seconds, nil
}

func bucket(column string, f Filter) string {
	seconds, _ := timezoneSeconds(f.Timezone)
	converted := fmt.Sprintf("DATE_ADD(%s, INTERVAL %d SECOND)", column, seconds-8*3600)
	switch f.Granularity {
	case "hour":
		return "DATE_FORMAT(" + converted + ",'%Y-%m-%d %H:00')"
	case "week":
		return "DATE_FORMAT(DATE_SUB(DATE(" + converted + "), INTERVAL WEEKDAY(" + converted + ") DAY),'%Y-%m-%d')"
	case "month":
		return "DATE_FORMAT(" + converted + ",'%Y-%m')"
	default:
		return "DATE_FORMAT(" + converted + ",'%Y-%m-%d')"
	}
}

func decimalString(value decimal.Decimal) string { return value.StringFixed(2) }

type moneyRow struct {
	Period string          `gorm:"column:period"`
	Value  decimal.Decimal `gorm:"column:value"`
	Value2 decimal.Decimal `gorm:"column:value2"`
	Value3 decimal.Decimal `gorm:"column:value3"`
	Value4 decimal.Decimal `gorm:"column:value4"`
}

func (s *Service) Filters(scope content.Scope) (map[string]any, error) {
	var countries, currencies, providers, categories, channels, agents []map[string]any
	queries := []struct {
		dest any
		sql  string
		args []any
	}{
		dashboardOptionSQL(&countries, scope, "game_users", "", "SELECT DISTINCT country_code value,country_code label FROM game_users WHERE deleted_at IS NULL AND country_code<>''", " ORDER BY country_code"),
		dashboardOptionSQL(&currencies, scope, "game_user_wallets", "", "SELECT DISTINCT currency value,currency label FROM game_user_wallets WHERE deleted_at IS NULL AND currency<>''", " ORDER BY currency"),
		dashboardOptionSQL(&providers, scope, "game_game_providers", "", "SELECT provider_code value,name label FROM game_game_providers WHERE deleted_at IS NULL AND status=1", " ORDER BY sort,id"),
		dashboardOptionSQL(&categories, scope, "game_game_categories", "", "SELECT id value,COALESCE(NULLIF(name_en,''),label) label FROM game_game_categories WHERE deleted_at IS NULL AND status=1", " ORDER BY sort,id"),
		dashboardOptionSQL(&channels, scope, "game_agent_channels", "", "SELECT id value,COALESCE(NULLIF(name_en,''),code) label FROM game_agent_channels WHERE deleted_at IS NULL AND enabled=1", " ORDER BY id"),
		dashboardOptionSQL(&agents, scope, "game_agents", "", "SELECT id value,name label FROM game_agents WHERE deleted_at IS NULL AND status='ACTIVE'", " ORDER BY id"),
	}
	for _, query := range queries {
		if err := s.db.Raw(query.sql, query.args...).Scan(query.dest).Error; err != nil {
			return nil, err
		}
	}
	if len(currencies) == 0 {
		currencies = []map[string]any{{"value": "PKR", "label": "PKR"}}
	}
	return map[string]any{"countries": countries, "currencies": currencies, "providers": providers, "categories": categories, "channels": channels, "agents": agents}, nil
}

func dashboardOptionSQL(dest any, scope content.Scope, _ string, alias string, base string, order string) struct {
	dest any
	sql  string
	args []any
} {
	where, args := dashboardScopeSQL(scope, alias)
	return struct {
		dest any
		sql  string
		args []any
	}{dest: dest, sql: base + where + order, args: args}
}

func dashboardScopeSQL(scope content.Scope, alias string) (string, []any) {
	if !scope.SiteScoped() {
		return "", nil
	}
	prefix := ""
	if strings.TrimSpace(alias) != "" {
		prefix = alias + "."
	}
	where := ""
	args := []any{}
	if scope.SiteScoped() {
		where += " AND " + prefix + "site_id IN (0, ?)"
		args = append(args, scope.SiteID)
	}
	return where, args
}

func userScope(f Filter, alias string) (string, []any) {
	where, args := dashboardScopeSQL(f.Scope, alias)
	if f.Country != "" {
		where += " AND " + alias + ".country_code=?"
		args = append(args, f.Country)
	}
	if f.ChannelID > 0 {
		where += " AND EXISTS (SELECT 1 FROM game_user_agent_attributions dua WHERE dua.user_id=" + alias + ".id AND dua.channel_id=? AND dua.active=1 AND dua.deleted_at IS NULL)"
		args = append(args, f.ChannelID)
	}
	if f.AgentID > 0 {
		where += " AND EXISTS (SELECT 1 FROM game_user_agent_attributions dua WHERE dua.user_id=" + alias + ".id AND dua.agent_id=? AND dua.active=1 AND dua.deleted_at IS NULL)"
		args = append(args, f.AgentID)
	}
	return where, args
}

func betDimensionScope(f Filter, alias string) (string, []any) {
	where := ""
	args := []any{}
	if f.ProviderCode != "" {
		where += " AND " + alias + ".provider_code=?"
		args = append(args, f.ProviderCode)
	}
	if f.CategoryID > 0 {
		where += " AND EXISTS (SELECT 1 FROM game_game_category_relations dcr WHERE dcr.game_id=" + alias + ".game_id AND dcr.category_id=? AND dcr.deleted_at IS NULL)"
		args = append(args, f.CategoryID)
	}
	return where, args
}

func (s *Service) Overview(f Filter) (map[string]any, error) {
	duration := f.EndTime.Sub(f.StartTime)
	previousEnd := f.StartTime.Add(-time.Millisecond)
	previousStart := previousEnd.Add(-duration)
	current, err := s.overviewPeriod(f, f.StartTime, f.EndTime)
	if err != nil {
		return nil, err
	}
	previous, err := s.overviewPeriod(f, previousStart, previousEnd)
	if err != nil {
		return nil, err
	}
	keys := []string{"depositCount", "depositAmount", "withdrawalCount", "withdrawalAmount", "betCount", "totalBet", "ggr", "onlineUsers", "registeredUsers", "depositSuccessRate", "withdrawalSuccessRate", "pendingWithdrawalCount", "activeUsers", "firstDepositUsers", "payingUsers", "netCashFlow", "totalWin", "ngr", "actualRTP", "bonusCost", "pendingWithdrawalAmount"}
	cards := make([]map[string]any, 0, len(keys))
	for _, key := range keys {
		cv, pv := current[key], previous[key]
		cards = append(cards, map[string]any{"key": key, "value": cv, "previousValue": pv, "changePercent": percentChange(cv, pv), "format": metricFormat(key), "currency": f.Currency})
	}
	return map[string]any{"cards": cards, "currency": f.Currency, "moneyFormat": MoneyFormat, "metricBasis": metricBasis()}, nil
}

func metricFormat(key string) string {
	if key == "actualRTP" || strings.HasSuffix(key, "Rate") {
		return "percent"
	}
	lower := strings.ToLower(key)
	if strings.Contains(lower, "user") || strings.Contains(lower, "count") {
		return "count"
	}
	return "money"
}
func numeric(v any) decimal.Decimal { d, _ := decimal.NewFromString(fmt.Sprint(v)); return d }
func percentChange(current any, previous any) any {
	p := numeric(previous)
	if p.IsZero() {
		return nil
	}
	return numeric(current).Sub(p).Div(p.Abs()).Mul(decimal.NewFromInt(100)).Round(2).StringFixed(2)
}
func metricBasis() map[string]string {
	return map[string]string{"deposit": "SUCCESS_PAID_AT", "withdrawal": "SUCCESS_COMPLETED_AT", "betting": "SETTLED_AT_EXCLUDING_ROLLBACK", "registration": "USER_CREATED_AT", "moneyFormat": MoneyFormat}
}

func (s *Service) overviewPeriod(f Filter, start time.Time, end time.Time) (map[string]any, error) {
	scope, scopeArgs := userScope(f, "u")
	betDimensions, betDimensionArgs := betDimensionScope(f, "b")
	commissionScope, commissionScopeArgs := dashboardScopeSQL(f.Scope, "c")
	withdrawScope, withdrawScopeArgs := dashboardScopeSQL(f.Scope, "")
	sessionScope, sessionScopeArgs := dashboardScopeSQL(f.Scope, "")
	type agg struct {
		RegisteredUsers, ActiveUsers, FirstDepositUsers, PayingUsers, OnlineUsers                                                    int64
		DepositCount, DepositAttemptCount, WithdrawalCount, WithdrawalAttemptCount, BetCount, PendingWithdrawalCount                 int64
		DepositAmount, WithdrawalAmount, TotalBet, TotalWin, BonusCost, ProviderCost, PaymentFee, AgentCommission, PendingWithdrawal decimal.Decimal
	}
	a := agg{}
	queries := []struct {
		sql  string
		args []any
		dest any
	}{
		{"SELECT COUNT(*) registered_users FROM game_users u WHERE u.deleted_at IS NULL AND u.created_at BETWEEN ? AND ?" + scope, append([]any{start, end}, scopeArgs...), &a},
		{"SELECT COUNT(DISTINCT b.user_id) active_users FROM game_bet_records b JOIN game_users u ON u.id=b.user_id AND u.deleted_at IS NULL WHERE b.deleted_at IS NULL AND b.settled_at BETWEEN ? AND ? AND b.bet_status='SETTLED' AND COALESCE(b.rollback_status,'')<>'ROLLED_BACK' AND b.currency=?" + scope, append([]any{start, end, f.Currency}, scopeArgs...), &a},
		{"SELECT COUNT(DISTINCT d.user_id) first_deposit_users FROM game_deposit_orders d JOIN game_users u ON u.id=d.user_id AND u.deleted_at IS NULL WHERE d.deleted_at IS NULL AND d.status=2 AND d.paid_at BETWEEN ? AND ? AND d.currency=? AND NOT EXISTS(SELECT 1 FROM game_deposit_orders x WHERE x.user_id=d.user_id AND x.status=2 AND x.deleted_at IS NULL AND x.paid_at<d.paid_at)" + scope, append([]any{start, end, f.Currency}, scopeArgs...), &a},
		{"SELECT COUNT(DISTINCT d.user_id) paying_users FROM game_deposit_orders d JOIN game_users u ON u.id=d.user_id AND u.deleted_at IS NULL WHERE d.deleted_at IS NULL AND d.status=2 AND d.paid_at BETWEEN ? AND ? AND d.currency=?" + scope, append([]any{start, end, f.Currency}, scopeArgs...), &a},
		{"SELECT COUNT(*) deposit_attempt_count,COALESCE(SUM(CASE WHEN d.status=2 THEN 1 ELSE 0 END),0) deposit_count FROM game_deposit_orders d JOIN game_users u ON u.id=d.user_id AND u.deleted_at IS NULL WHERE d.deleted_at IS NULL AND d.created_at BETWEEN ? AND ? AND d.currency=?" + scope, append([]any{start, end, f.Currency}, scopeArgs...), &a},
		{"SELECT COALESCE(SUM(d.paid_amount),0) deposit_amount,COALESCE(SUM(d.fee_amount),0) payment_fee FROM game_deposit_orders d JOIN game_users u ON u.id=d.user_id AND u.deleted_at IS NULL WHERE d.deleted_at IS NULL AND d.status=2 AND d.paid_at BETWEEN ? AND ? AND d.currency=?" + scope, append([]any{start, end, f.Currency}, scopeArgs...), &a},
		{"SELECT COUNT(*) withdrawal_attempt_count,COALESCE(SUM(CASE WHEN w.status=2 THEN 1 ELSE 0 END),0) withdrawal_count FROM game_withdraw_orders w JOIN game_users u ON u.id=w.user_id AND u.deleted_at IS NULL WHERE w.deleted_at IS NULL AND w.created_at BETWEEN ? AND ? AND w.currency=?" + scope, append([]any{start, end, f.Currency}, scopeArgs...), &a},
		{"SELECT COALESCE(SUM(w.actual_amount),0) withdrawal_amount FROM game_withdraw_orders w JOIN game_users u ON u.id=w.user_id AND u.deleted_at IS NULL WHERE w.deleted_at IS NULL AND w.status=2 AND w.completed_at BETWEEN ? AND ? AND w.currency=?" + scope, append([]any{start, end, f.Currency}, scopeArgs...), &a},
		{"SELECT COUNT(*) bet_count,COALESCE(SUM(b.bet_amount),0) total_bet,COALESCE(SUM(b.payout_amount),0) total_win FROM game_bet_records b JOIN game_users u ON u.id=b.user_id AND u.deleted_at IS NULL WHERE b.deleted_at IS NULL AND b.bet_status='SETTLED' AND COALESCE(b.rollback_status,'')<>'ROLLED_BACK' AND b.settled_at BETWEEN ? AND ? AND b.currency=?" + scope + betDimensions, append(append([]any{start, end, f.Currency}, scopeArgs...), betDimensionArgs...), &a},
		{"SELECT COALESCE(SUM(bo.amount),0) bonus_cost FROM game_bonus_records bo JOIN game_users u ON u.id=bo.user_id AND u.deleted_at IS NULL WHERE bo.deleted_at IS NULL AND bo.status_code IN ('ACTIVE','COMPLETED') AND bo.issued_at BETWEEN ? AND ? AND bo.currency=?" + scope, append([]any{start, end, f.Currency}, scopeArgs...), &a},
		{"SELECT COALESCE(SUM(c.commission_amount),0) agent_commission FROM game_commission_records c WHERE c.deleted_at IS NULL AND c.status IN ('CONFIRMED','FROZEN','SETTLED') AND c.created_at BETWEEN ? AND ? AND c.currency=?" + commissionScope, append([]any{start, end, f.Currency}, commissionScopeArgs...), &a},
		{"SELECT COUNT(*) pending_withdrawal_count,COALESCE(SUM(actual_amount),0) pending_withdrawal FROM game_withdraw_orders WHERE deleted_at IS NULL AND status<>2 AND status<>3 AND created_at BETWEEN ? AND ? AND currency=?" + withdrawScope, append([]any{start, end, f.Currency}, withdrawScopeArgs...), &a},
		{"SELECT COUNT(DISTINCT user_id) online_users FROM game_user_sessions WHERE deleted_at IS NULL AND revoked_at IS NULL AND expired_at>NOW(3)" + sessionScope, sessionScopeArgs, &a},
	}
	for _, q := range queries {
		if err := s.db.Raw(q.sql, q.args...).Scan(q.dest).Error; err != nil {
			return nil, err
		}
	}
	ggr := a.TotalBet.Sub(a.TotalWin)
	ngr := ggr.Sub(a.BonusCost).Sub(a.ProviderCost).Sub(a.PaymentFee).Sub(a.AgentCommission)
	rtp := decimal.Zero
	if !a.TotalBet.IsZero() {
		rtp = a.TotalWin.Div(a.TotalBet).Mul(decimal.NewFromInt(100))
	}
	depositRate := decimal.Zero
	if a.DepositAttemptCount > 0 {
		depositRate = decimal.NewFromInt(a.DepositCount).Div(decimal.NewFromInt(a.DepositAttemptCount)).Mul(decimal.NewFromInt(100))
	}
	withdrawalRate := decimal.Zero
	if a.WithdrawalAttemptCount > 0 {
		withdrawalRate = decimal.NewFromInt(a.WithdrawalCount).Div(decimal.NewFromInt(a.WithdrawalAttemptCount)).Mul(decimal.NewFromInt(100))
	}
	return map[string]any{"registeredUsers": a.RegisteredUsers, "activeUsers": a.ActiveUsers, "firstDepositUsers": a.FirstDepositUsers, "payingUsers": a.PayingUsers, "onlineUsers": a.OnlineUsers, "depositCount": a.DepositCount, "withdrawalCount": a.WithdrawalCount, "betCount": a.BetCount, "pendingWithdrawalCount": a.PendingWithdrawalCount, "depositSuccessRate": depositRate.Round(2).StringFixed(2), "withdrawalSuccessRate": withdrawalRate.Round(2).StringFixed(2), "depositAmount": decimalString(a.DepositAmount), "withdrawalAmount": decimalString(a.WithdrawalAmount), "netCashFlow": decimalString(a.DepositAmount.Sub(a.WithdrawalAmount)), "totalBet": decimalString(a.TotalBet), "totalWin": decimalString(a.TotalWin), "ggr": decimalString(ggr), "ngr": decimalString(ngr), "actualRTP": rtp.Round(2).StringFixed(2), "bonusCost": decimalString(a.BonusCost), "pendingWithdrawalAmount": decimalString(a.PendingWithdrawal)}, nil
}

func chart(dimensions []string, metrics []string, rows any, f Filter) map[string]any {
	return map[string]any{"dimensions": dimensions, "metrics": metrics, "rows": rows, "currency": f.Currency, "moneyFormat": MoneyFormat}
}

func (s *Service) FinancialTrend(f Filter) (map[string]any, error) {
	return s.moneyTrend(f, "financial")
}
func (s *Service) BettingTrend(f Filter) (map[string]any, error) { return s.moneyTrend(f, "betting") }
func (s *Service) moneyTrend(f Filter, kind string) (map[string]any, error) {
	rows := []map[string]any{}
	scope, scopeArgs := userScope(f, "u")
	if kind == "financial" {
		type row struct {
			Period              string
			Deposit, Withdrawal decimal.Decimal
		}
		data := []row{}
		sql := `WITH facts AS (SELECT ` + bucket("d.paid_at", f) + ` period,d.paid_amount deposit_amount,0 withdrawal_amount FROM game_deposit_orders d JOIN game_users u ON u.id=d.user_id AND u.deleted_at IS NULL WHERE d.deleted_at IS NULL AND d.status=2 AND d.paid_at BETWEEN ? AND ? AND d.currency=?` + scope + ` UNION ALL SELECT ` + bucket("w.completed_at", f) + `,0,w.actual_amount FROM game_withdraw_orders w JOIN game_users u ON u.id=w.user_id AND u.deleted_at IS NULL WHERE w.deleted_at IS NULL AND w.status=2 AND w.completed_at BETWEEN ? AND ? AND w.currency=?` + scope + `) SELECT period,COALESCE(SUM(deposit_amount),0) deposit,COALESCE(SUM(withdrawal_amount),0) withdrawal FROM facts GROUP BY period ORDER BY period`
		args := append(append(append([]any{f.StartTime, f.EndTime, f.Currency}, scopeArgs...), f.StartTime, f.EndTime, f.Currency), scopeArgs...)
		if err := s.db.Raw(sql, args...).Scan(&data).Error; err != nil {
			return nil, err
		}
		for _, v := range data {
			rows = append(rows, map[string]any{"time": v.Period, "depositAmount": decimalString(v.Deposit), "withdrawalAmount": decimalString(v.Withdrawal), "netCashFlow": decimalString(v.Deposit.Sub(v.Withdrawal))})
		}
		return chart([]string{"time"}, []string{"depositAmount", "withdrawalAmount", "netCashFlow"}, rows, f), nil
	}
	type row struct {
		Period                           string
		Bet, Win, Bonus, Fee, Commission decimal.Decimal
	}
	data := []row{}
	betScope, betArgs := betDimensionScope(f, "b")
	commissionScope, commissionArgs := dashboardScopeSQL(f.Scope, "c")
	sql := `WITH facts AS (SELECT ` + bucket("b.settled_at", f) + ` period,b.bet_amount bet,b.payout_amount win,0 bonus,0 fee,0 commission FROM game_bet_records b JOIN game_users u ON u.id=b.user_id AND u.deleted_at IS NULL WHERE b.deleted_at IS NULL AND b.bet_status='SETTLED' AND COALESCE(b.rollback_status,'')<>'ROLLED_BACK' AND b.settled_at BETWEEN ? AND ? AND b.currency=?` + scope + betScope + ` UNION ALL SELECT ` + bucket("bo.issued_at", f) + `,0,0,bo.amount,0,0 FROM game_bonus_records bo JOIN game_users u ON u.id=bo.user_id AND u.deleted_at IS NULL WHERE bo.deleted_at IS NULL AND bo.status_code IN ('ACTIVE','COMPLETED') AND bo.issued_at BETWEEN ? AND ? AND bo.currency=?` + scope + ` UNION ALL SELECT ` + bucket("d.paid_at", f) + `,0,0,0,d.fee_amount,0 FROM game_deposit_orders d JOIN game_users u ON u.id=d.user_id AND u.deleted_at IS NULL WHERE d.deleted_at IS NULL AND d.status=2 AND d.paid_at BETWEEN ? AND ? AND d.currency=?` + scope + ` UNION ALL SELECT ` + bucket("c.created_at", f) + `,0,0,0,0,c.commission_amount FROM game_commission_records c LEFT JOIN game_users u ON u.id=c.user_id AND u.deleted_at IS NULL WHERE c.deleted_at IS NULL AND c.status IN ('CONFIRMED','FROZEN','SETTLED') AND c.created_at BETWEEN ? AND ? AND c.currency=?` + commissionScope + `) SELECT period,SUM(bet) bet,SUM(win) win,SUM(bonus) bonus,SUM(fee) fee,SUM(commission) commission FROM facts GROUP BY period ORDER BY period`
	args := append(append([]any{f.StartTime, f.EndTime, f.Currency}, scopeArgs...), betArgs...)
	args = append(append(args, f.StartTime, f.EndTime, f.Currency), scopeArgs...)
	args = append(append(args, f.StartTime, f.EndTime, f.Currency), scopeArgs...)
	args = append(append(args, f.StartTime, f.EndTime, f.Currency), commissionArgs...)
	if err := s.db.Raw(sql, args...).Scan(&data).Error; err != nil {
		return nil, err
	}
	for _, v := range data {
		ggr := v.Bet.Sub(v.Win)
		rows = append(rows, map[string]any{"time": v.Period, "totalBet": decimalString(v.Bet), "totalWin": decimalString(v.Win), "ggr": decimalString(ggr), "ngr": decimalString(ggr.Sub(v.Bonus).Sub(v.Fee).Sub(v.Commission))})
	}
	return chart([]string{"time"}, []string{"totalBet", "totalWin", "ggr", "ngr"}, rows, f), nil
}

func (s *Service) UserTrend(f Filter) (map[string]any, error) {
	type row struct {
		Period                                    string
		Registered, Active, FirstDeposit, Betting int64
	}
	data := []row{}
	userWhere, userArgs := dashboardScopeSQL(f.Scope, "")
	betWhere, betArgs := dashboardScopeSQL(f.Scope, "")
	depositWhere, depositArgs := dashboardScopeSQL(f.Scope, "d")
	sql := `WITH facts AS (SELECT ` + bucket("created_at", f) + ` period,COUNT(*) registered,0 active,0 first_deposit,0 betting FROM game_users WHERE deleted_at IS NULL AND created_at BETWEEN ? AND ?` + userWhere + ` GROUP BY 1 UNION ALL SELECT ` + bucket("settled_at", f) + `,0,COUNT(DISTINCT user_id),0,COUNT(DISTINCT user_id) FROM game_bet_records WHERE deleted_at IS NULL AND bet_status='SETTLED' AND settled_at BETWEEN ? AND ? AND currency=?` + betWhere + ` GROUP BY 1 UNION ALL SELECT ` + bucket("paid_at", f) + `,0,0,COUNT(DISTINCT user_id),0 FROM game_deposit_orders d WHERE deleted_at IS NULL AND status=2 AND paid_at BETWEEN ? AND ? AND currency=?` + depositWhere + ` AND NOT EXISTS(SELECT 1 FROM game_deposit_orders x WHERE x.user_id=d.user_id AND x.status=2 AND x.deleted_at IS NULL AND x.paid_at<d.paid_at) GROUP BY 1) SELECT period,SUM(registered) registered,SUM(active) active,SUM(first_deposit) first_deposit,SUM(betting) betting FROM facts GROUP BY period ORDER BY period`
	args := append([]any{f.StartTime, f.EndTime}, userArgs...)
	args = append(append(args, f.StartTime, f.EndTime, f.Currency), betArgs...)
	args = append(append(args, f.StartTime, f.EndTime, f.Currency), depositArgs...)
	if err := s.db.Raw(sql, args...).Scan(&data).Error; err != nil {
		return nil, err
	}
	rows := []map[string]any{}
	for _, v := range data {
		rows = append(rows, map[string]any{"time": v.Period, "registeredUsers": v.Registered, "activeUsers": v.Active, "firstDepositUsers": v.FirstDeposit, "bettingUsers": v.Betting})
	}
	return chart([]string{"time"}, []string{"registeredUsers", "activeUsers", "firstDepositUsers", "bettingUsers"}, rows, f), nil
}

func (s *Service) UserFunnel(f Filter) (map[string]any, error) {
	type counts struct{ Registered, LoggedIn, Depositing, FirstDeposit, Betting, Returning int64 }
	v := counts{}
	userWhere, userArgs := dashboardScopeSQL(f.Scope, "")
	loginWhere, loginArgs := dashboardScopeSQL(f.Scope, "")
	depositWhere, depositArgs := dashboardScopeSQL(f.Scope, "")
	depositAliasWhere, depositAliasArgs := dashboardScopeSQL(f.Scope, "d")
	betWhere, betArgs := dashboardScopeSQL(f.Scope, "")
	queries := []struct {
		sql  string
		args []any
	}{
		{"SELECT COUNT(*) registered FROM game_users WHERE deleted_at IS NULL AND created_at BETWEEN ? AND ?" + userWhere, append([]any{f.StartTime, f.EndTime}, userArgs...)},
		{"SELECT COUNT(DISTINCT user_id) logged_in FROM game_user_login_records WHERE created_at BETWEEN ? AND ? AND success=1 AND deleted_at IS NULL" + loginWhere, append([]any{f.StartTime, f.EndTime}, loginArgs...)},
		{"SELECT COUNT(DISTINCT user_id) depositing FROM game_deposit_orders WHERE paid_at BETWEEN ? AND ? AND status=2 AND currency=? AND deleted_at IS NULL" + depositWhere, append([]any{f.StartTime, f.EndTime, f.Currency}, depositArgs...)},
		{"SELECT COUNT(DISTINCT d.user_id) first_deposit FROM game_deposit_orders d WHERE d.paid_at BETWEEN ? AND ? AND d.status=2 AND d.currency=? AND d.deleted_at IS NULL" + depositAliasWhere + " AND NOT EXISTS(SELECT 1 FROM game_deposit_orders x WHERE x.user_id=d.user_id AND x.status=2 AND x.deleted_at IS NULL AND x.paid_at<d.paid_at)", append([]any{f.StartTime, f.EndTime, f.Currency}, depositAliasArgs...)},
		{"SELECT COUNT(DISTINCT user_id) betting FROM game_bet_records WHERE settled_at BETWEEN ? AND ? AND bet_status='SETTLED' AND currency=? AND deleted_at IS NULL" + betWhere, append([]any{f.StartTime, f.EndTime, f.Currency}, betArgs...)},
		{"SELECT COUNT(DISTINCT user_id) returning FROM game_user_login_records WHERE created_at BETWEEN ? AND ? AND success=1 AND deleted_at IS NULL" + loginWhere + " AND user_id IN(SELECT user_id FROM game_user_login_records WHERE created_at<? AND success=1 AND deleted_at IS NULL" + loginWhere + ")", append(append([]any{f.StartTime, f.EndTime}, loginArgs...), append([]any{f.StartTime}, loginArgs...)...)},
	}
	for _, q := range queries {
		if err := s.db.Raw(q.sql, q.args...).Scan(&v).Error; err != nil {
			return nil, err
		}
	}
	rows := []map[string]any{{"stage": "registeredUsers", "value": v.Registered}, {"stage": "loggedInUsers", "value": v.LoggedIn}, {"stage": "depositingUsers", "value": v.Depositing}, {"stage": "firstDepositUsers", "value": v.FirstDeposit}, {"stage": "bettingUsers", "value": v.Betting}, {"stage": "returningUsers", "value": v.Returning}}
	return chart([]string{"stage"}, []string{"value"}, rows, f), nil
}

func (s *Service) OrderStatus(f Filter) (map[string]any, error) {
	rows := []map[string]any{}
	type row struct {
		Status int
		Value  int64
	}
	dep, with := []row{}, []row{}
	depositWhere, depositArgs := dashboardScopeSQL(f.Scope, "")
	withdrawWhere, withdrawArgs := dashboardScopeSQL(f.Scope, "")
	if err := s.db.Raw("SELECT status,COUNT(*) value FROM game_deposit_orders WHERE deleted_at IS NULL AND created_at BETWEEN ? AND ? AND currency=?"+depositWhere+" GROUP BY status", append([]any{f.StartTime, f.EndTime, f.Currency}, depositArgs...)...).Scan(&dep).Error; err != nil {
		return nil, err
	}
	if err := s.db.Raw("SELECT status,COUNT(*) value FROM game_withdraw_orders WHERE deleted_at IS NULL AND created_at BETWEEN ? AND ? AND currency=?"+withdrawWhere+" GROUP BY status", append([]any{f.StartTime, f.EndTime, f.Currency}, withdrawArgs...)...).Scan(&with).Error; err != nil {
		return nil, err
	}
	for _, v := range dep {
		rows = append(rows, map[string]any{"orderType": "deposit", "status": statusKey("deposit", v.Status), "value": v.Value})
	}
	for _, v := range with {
		rows = append(rows, map[string]any{"orderType": "withdrawal", "status": statusKey("withdrawal", v.Status), "value": v.Value})
	}
	return chart([]string{"orderType", "status"}, []string{"value"}, rows, f), nil
}
func statusKey(kind string, status int) string {
	if kind == "deposit" {
		return map[int]string{0: "pending", 1: "processing", 2: "success", 3: "failed", 4: "expired", 5: "cancelled"}[status]
	}
	return map[int]string{0: "pending", 1: "processing", 2: "success", 3: "rejected", 4: "failed"}[status]
}

func metricColumn(metric string) string {
	return map[string]string{"ggr": "SUM(s.total_bet-s.total_win)", "uniquePlayers": "SUM(s.unique_players)", "roundCount": "SUM(s.total_rounds)", "maximumWin": "MAX(s.maximum_win)"}[metric]
}
func (s *Service) GameCategories(f Filter) (map[string]any, error) {
	col := metricColumn(f.Metric)
	if col == "" {
		col = "SUM(s.total_bet)"
	}
	rows := []map[string]any{}
	scopeWhere, scopeArgs := dashboardScopeSQL(f.Scope, "s")
	sql := `SELECT COALESCE(c.category_key,'other') name,` + col + ` value FROM game_daily_game_summaries s LEFT JOIN game_game_category_relations r ON r.game_id=s.game_id AND r.deleted_at IS NULL LEFT JOIN game_game_categories c ON c.id=r.category_id AND c.deleted_at IS NULL WHERE s.deleted_at IS NULL AND s.business_date BETWEEN DATE(?) AND DATE(?) AND s.currency=?` + scopeWhere
	args := append([]any{f.StartTime, f.EndTime, f.Currency}, scopeArgs...)
	if f.ProviderCode != "" {
		sql += " AND s.provider_code=?"
		args = append(args, f.ProviderCode)
	}
	if f.CategoryID > 0 {
		sql += " AND c.id=?"
		args = append(args, f.CategoryID)
	}
	sql += " GROUP BY COALESCE(c.category_key,'other') ORDER BY value DESC"
	if err := s.db.Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return chart([]string{"name"}, []string{"value"}, rows, f), nil
}

func (s *Service) TopGames(f Filter) (map[string]any, error) {
	col := metricColumn(f.Metric)
	if col == "" {
		col = "SUM(s.total_bet)"
	}
	rows := []map[string]any{}
	scopeWhere, scopeArgs := dashboardScopeSQL(f.Scope, "s")
	sql := `SELECT CAST(g.id AS CHAR) id,g.name,p.provider_code AS providerCode,` + col + ` value,CASE WHEN SUM(s.total_bet)=0 THEN NULL ELSE ROUND(SUM(s.total_win)/SUM(s.total_bet)*100,2) END secondaryValue FROM game_daily_game_summaries s JOIN game_games g ON g.id=s.game_id AND g.deleted_at IS NULL JOIN game_game_providers p ON p.id=g.provider_id AND p.deleted_at IS NULL WHERE s.deleted_at IS NULL AND s.business_date BETWEEN DATE(?) AND DATE(?) AND s.currency=?` + scopeWhere
	args := append([]any{f.StartTime, f.EndTime, f.Currency}, scopeArgs...)
	if f.ProviderCode != "" {
		sql += " AND p.provider_code=?"
		args = append(args, f.ProviderCode)
	}
	sql += " GROUP BY g.id,g.name,p.provider_code ORDER BY value DESC LIMIT ?"
	args = append(args, f.Limit)
	if err := s.db.Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return chart([]string{"name"}, []string{"value", "secondaryValue"}, rows, f), nil
}

func (s *Service) ProviderPerformance(f Filter) (map[string]any, error) {
	rows := []map[string]any{}
	scopeWhere, scopeArgs := dashboardScopeSQL(f.Scope, "s")
	sql := `SELECT s.provider_code id,s.provider_code name,SUM(s.total_bet) totalBet,SUM(s.total_win) totalWin,SUM(s.total_bet-s.total_win) ggr,SUM(s.unique_players) uniquePlayers,SUM(s.launch_count) launchCount,CASE WHEN SUM(s.launch_count)=0 THEN NULL ELSE ROUND(SUM(s.launch_success_count)/SUM(s.launch_count)*100,2) END launchSuccessRate,SUM(s.callback_count) callbackCount,CASE WHEN SUM(s.callback_count)=0 THEN NULL ELSE ROUND(SUM(s.callback_success_count)/SUM(s.callback_count)*100,2) END callbackSuccessRate,CASE WHEN SUM(s.callback_count)=0 THEN NULL ELSE ROUND(SUM(s.callback_duration_ms)/SUM(s.callback_count),2) END averageCallbackLatency,CASE WHEN SUM(s.total_bet)=0 THEN NULL ELSE ROUND(SUM(s.total_win)/SUM(s.total_bet)*100,2) END actualRTP FROM game_daily_provider_summaries s WHERE s.deleted_at IS NULL AND s.business_date BETWEEN DATE(?) AND DATE(?) AND s.currency=?` + scopeWhere
	args := append([]any{f.StartTime, f.EndTime, f.Currency}, scopeArgs...)
	if f.ProviderCode != "" {
		sql += " AND s.provider_code=?"
		args = append(args, f.ProviderCode)
	}
	sql += " GROUP BY s.provider_code ORDER BY SUM(s.total_bet) DESC"
	if err := s.db.Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return chart([]string{"name"}, []string{"totalBet", "totalWin", "ggr", "actualRTP", "uniquePlayers", "launchSuccessRate", "callbackSuccessRate", "averageCallbackLatency"}, rows, f), nil
}

func (s *Service) CountryDistribution(f Filter) (map[string]any, error) {
	col := map[string]string{"registeredUsers": "COUNT(DISTINCT u.id)", "activeUsers": "SUM(s.active)", "depositAmount": "SUM(s.deposit_amount)", "totalBet": "SUM(s.total_bet)", "ggr": "SUM(s.total_bet-s.total_win)"}[f.Metric]
	if col == "" {
		col = "SUM(s.total_bet-s.total_win)"
	}
	rows := []map[string]any{}
	scopeWhere, scopeArgs := dashboardScopeSQL(f.Scope, "s")
	sql := `SELECT COALESCE(NULLIF(s.country,''),'other') name,` + col + ` value FROM game_daily_user_summaries s LEFT JOIN game_users u ON u.id=s.user_id WHERE s.deleted_at IS NULL AND s.business_date BETWEEN DATE(?) AND DATE(?) AND s.currency=?` + scopeWhere + ` GROUP BY COALESCE(NULLIF(s.country,''),'other') ORDER BY value DESC LIMIT 12`
	if err := s.db.Raw(sql, append([]any{f.StartTime, f.EndTime, f.Currency}, scopeArgs...)...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return chart([]string{"name"}, []string{"value"}, rows, f), nil
}

func (s *Service) ChannelPerformance(f Filter) (map[string]any, error) {
	return s.summaryRanking(f, "channel")
}
func (s *Service) AgentPerformance(f Filter) (map[string]any, error) {
	return s.summaryRanking(f, "agent")
}
func (s *Service) summaryRanking(f Filter, kind string) (map[string]any, error) {
	rows := []map[string]any{}
	var sql string
	scopeWhere, scopeArgs := dashboardScopeSQL(f.Scope, "s")
	args := append([]any{f.StartTime, f.EndTime, f.Currency}, scopeArgs...)
	if kind == "channel" {
		sql = `SELECT CAST(s.channel_id AS CHAR) id,COALESCE(NULLIF(c.name_en,''),c.code) name,SUM(s.registrations) registrations,SUM(s.first_deposit_users) firstDepositUsers,CASE WHEN SUM(s.registrations)=0 THEN NULL ELSE ROUND(SUM(s.first_deposit_users)/SUM(s.registrations)*100,2) END conversionRate,SUM(s.deposit_amount) depositAmount,SUM(s.total_bet-s.total_win) ggr FROM game_daily_channel_summaries s LEFT JOIN game_agent_channels c ON c.id=s.channel_id WHERE s.deleted_at IS NULL AND s.business_date BETWEEN DATE(?) AND DATE(?) AND s.currency=?` + scopeWhere
		if f.ChannelID > 0 {
			sql += " AND s.channel_id=?"
			args = append(args, f.ChannelID)
		}
		sql += " GROUP BY s.channel_id,c.name_en,c.code ORDER BY ggr DESC LIMIT ?"
	} else {
		sql = `SELECT CAST(s.agent_id AS CHAR) id,a.name,SUM(s.attributed_users) attributedUsers,SUM(s.first_deposit_users) firstDepositUsers,SUM(s.deposit_amount) depositAmount,SUM(s.total_bet) totalBet,SUM(s.total_bet-s.total_win) ggr,SUM(s.commission_amount) commissionAmount,SUM(s.total_bet-s.total_win-s.commission_amount) ngrAfterCommission FROM game_daily_agent_summaries s LEFT JOIN game_agents a ON a.id=s.agent_id WHERE s.deleted_at IS NULL AND s.business_date BETWEEN DATE(?) AND DATE(?) AND s.currency=?` + scopeWhere
		if f.AgentID > 0 {
			sql += " AND s.agent_id=?"
			args = append(args, f.AgentID)
		}
		sql += " GROUP BY s.agent_id,a.name ORDER BY ggr DESC LIMIT ?"
	}
	args = append(args, f.Limit)
	if err := s.db.Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return chart([]string{"name"}, []string{"value"}, rows, f), nil
}

func (s *Service) PromotionPerformance(f Filter) (map[string]any, error) {
	rows := []map[string]any{}
	scopeWhere, scopeArgs := dashboardScopeSQL(f.Scope, "s")
	sql := `SELECT CAST(s.promotion_id AS CHAR) id,COALESCE(NULLIF(a.name_en,''),a.title) name,SUM(s.participating_users) participatingUsers,SUM(s.reward_amount) rewardCost,SUM(s.completed_wagering) completedWagering,SUM(s.promotion_ggr) promotionGGR,CASE WHEN SUM(s.reward_amount)=0 THEN NULL ELSE ROUND((SUM(s.promotion_ggr)-SUM(s.reward_amount))/SUM(s.reward_amount)*100,2) END promotionROI FROM game_daily_promotion_summaries s LEFT JOIN game_activity_events a ON a.id=s.promotion_id WHERE s.deleted_at IS NULL AND s.business_date BETWEEN DATE(?) AND DATE(?) AND s.currency=?` + scopeWhere + ` GROUP BY s.promotion_id,a.name_en,a.title ORDER BY promotionGGR DESC`
	if err := s.db.Raw(sql, append([]any{f.StartTime, f.EndTime, f.Currency}, scopeArgs...)...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return chart([]string{"name"}, []string{"participatingUsers", "rewardCost", "completedWagering", "promotionGGR", "promotionROI"}, rows, f), nil
}

func (s *Service) RiskOverview(f Filter) (map[string]any, error) {
	type kpi struct{ PendingReviews, HighRiskUsers, CriticalEvents, SuspiciousWithdrawals, BonusAbuseEvents, BlockedAccounts, BlockedDevices, BlockedIPs int64 }
	v := kpi{}
	scopeWhere, scopeArgs := dashboardScopeSQL(f.Scope, "")
	queries := []struct {
		sql      string
		timeArgs bool
	}{{"SELECT COUNT(*) pending_reviews FROM game_risk_reviews WHERE deleted_at IS NULL AND status IN ('PENDING','UNDER_REVIEW')" + scopeWhere, false}, {"SELECT COUNT(*) high_risk_users FROM game_user_risk_profiles WHERE deleted_at IS NULL AND risk_level IN ('HIGH','CRITICAL') AND risk_status<>'CLEARED'" + scopeWhere, false}, {"SELECT COUNT(*) critical_events FROM game_user_risk_events WHERE deleted_at IS NULL AND risk_level='CRITICAL' AND created_at BETWEEN ? AND ?" + scopeWhere, true}, {"SELECT COUNT(*) suspicious_withdrawals FROM game_risk_reviews WHERE deleted_at IS NULL AND business_type='WITHDRAWAL' AND status IN ('PENDING','UNDER_REVIEW')" + scopeWhere, false}, {"SELECT COUNT(*) bonus_abuse_events FROM game_user_risk_events WHERE deleted_at IS NULL AND risk_type LIKE '%BONUS%' AND created_at BETWEEN ? AND ?" + scopeWhere, true}, {"SELECT COUNT(*) blocked_accounts FROM game_user_risk_profiles WHERE deleted_at IS NULL AND risk_status='BLOCKED'" + scopeWhere, false}, {"SELECT COUNT(*) blocked_devices FROM game_risk_blacklists WHERE deleted_at IS NULL AND blacklist_type='DEVICE' AND status='ACTIVE'" + scopeWhere, false}, {"SELECT COUNT(*) blocked_ips FROM game_risk_blacklists WHERE deleted_at IS NULL AND blacklist_type='IP' AND status='ACTIVE'" + scopeWhere, false}}
	for _, q := range queries {
		args := []any{}
		if q.timeArgs {
			args = append(args, f.StartTime, f.EndTime)
		}
		args = append(args, scopeArgs...)
		if err := s.db.Raw(q.sql, args...).Scan(&v).Error; err != nil {
			return nil, err
		}
	}
	byType, byLevel, trend, reviews := []map[string]any{}, []map[string]any{}, []map[string]any{}, []map[string]any{}
	args := append([]any{f.StartTime, f.EndTime}, scopeArgs...)
	s.db.Raw("SELECT risk_type name,COUNT(*) value FROM game_user_risk_events WHERE deleted_at IS NULL AND created_at BETWEEN ? AND ?"+scopeWhere+" GROUP BY risk_type ORDER BY value DESC", args...).Scan(&byType)
	s.db.Raw("SELECT risk_level name,COUNT(*) value FROM game_user_risk_events WHERE deleted_at IS NULL AND created_at BETWEEN ? AND ?"+scopeWhere+" GROUP BY risk_level", args...).Scan(&byLevel)
	s.db.Raw("SELECT "+bucket("created_at", f)+" time,COUNT(*) value FROM game_user_risk_events WHERE deleted_at IS NULL AND created_at BETWEEN ? AND ?"+scopeWhere+" GROUP BY time ORDER BY time", args...).Scan(&trend)
	s.db.Raw("SELECT status name,COUNT(*) value FROM game_risk_reviews WHERE deleted_at IS NULL AND created_at BETWEEN ? AND ?"+scopeWhere+" GROUP BY status", args...).Scan(&reviews)
	return map[string]any{"cards": []map[string]any{{"key": "pendingRiskReviews", "value": v.PendingReviews}, {"key": "highRiskUsers", "value": v.HighRiskUsers}, {"key": "criticalRiskEvents", "value": v.CriticalEvents}, {"key": "suspiciousWithdrawals", "value": v.SuspiciousWithdrawals}, {"key": "bonusAbuseEvents", "value": v.BonusAbuseEvents}, {"key": "blockedAccounts", "value": v.BlockedAccounts}, {"key": "blockedDevices", "value": v.BlockedDevices}, {"key": "blockedIPs", "value": v.BlockedIPs}}, "byType": byType, "byLevel": byLevel, "trend": trend, "reviewResults": reviews}, nil
}

func (s *Service) Realtime(f Filter) (map[string]any, error) {
	rows := []map[string]any{}
	scopeWhere, scopeArgs := dashboardScopeSQL(f.Scope, "")
	queries := []struct {
		key, sql string
		args     []any
	}{{"onlineUsers", "SELECT COUNT(DISTINCT user_id) value FROM game_user_sessions WHERE deleted_at IS NULL AND revoked_at IS NULL AND expired_at>NOW(3)" + scopeWhere, scopeArgs}, {"usersInGames", "SELECT COUNT(DISTINCT user_id) value FROM game_game_launch_logs WHERE deleted_at IS NULL AND session_status IN ('LAUNCH_SUCCESS','PLAYING') AND created_at>DATE_SUB(NOW(3),INTERVAL 30 MINUTE)" + scopeWhere, scopeArgs}, {"betCallbacksPerMinute", "SELECT COUNT(*) value FROM game_provider_callback_logs WHERE deleted_at IS NULL AND operation IN ('DEBIT','CREDIT') AND created_at>DATE_SUB(NOW(3),INTERVAL 1 MINUTE)" + scopeWhere, scopeArgs}, {"pendingDeposits", "SELECT COUNT(*) value FROM game_deposit_orders WHERE deleted_at IS NULL AND status IN (0,1)" + scopeWhere, scopeArgs}, {"pendingWithdrawals", "SELECT COUNT(*) value FROM game_withdraw_orders WHERE deleted_at IS NULL AND status IN (0,1)" + scopeWhere, scopeArgs}, {"providerFailures", "SELECT COUNT(*) value FROM game_provider_callback_logs WHERE deleted_at IS NULL AND processing_result<>'SUCCESS' AND created_at>DATE_SUB(NOW(3),INTERVAL 5 MINUTE)" + scopeWhere, scopeArgs}, {"gameLaunchSuccessRate", "SELECT CASE WHEN COUNT(*)=0 THEN NULL ELSE ROUND(SUM(launch_status=2)/COUNT(*)*100,2) END value FROM game_game_launch_logs WHERE deleted_at IS NULL AND created_at>DATE_SUB(NOW(3),INTERVAL 5 MINUTE)" + scopeWhere, scopeArgs}, {"walletCallbackSuccessRate", "SELECT CASE WHEN COUNT(*)=0 THEN NULL ELSE ROUND(SUM(processing_result='SUCCESS')/COUNT(*)*100,2) END value FROM game_provider_callback_logs WHERE deleted_at IS NULL AND created_at>DATE_SUB(NOW(3),INTERVAL 5 MINUTE)" + scopeWhere, scopeArgs}}
	for _, q := range queries {
		var v struct{ Value string }
		if err := s.db.Raw(q.sql, q.args...).Scan(&v).Error; err != nil {
			return nil, err
		}
		rows = append(rows, map[string]any{"key": q.key, "value": v.Value})
	}
	return map[string]any{"rows": rows, "refreshedAt": time.Now()}, nil
}

func (s *Service) Alerts(f Filter) (map[string]any, error) {
	rows := []map[string]any{}
	eventScope, eventArgs := dashboardScopeSQL(f.Scope, "e")
	callbackScope, callbackArgs := dashboardScopeSQL(f.Scope, "c")
	sql := `SELECT CONCAT('risk-',e.id) id,CASE e.risk_level WHEN 'CRITICAL' THEN 'critical' WHEN 'HIGH' THEN 'high' WHEN 'MEDIUM' THEN 'medium' ELSE 'low' END level,e.risk_type alert_type,COALESCE(NULLIF(e.description,''),e.trigger_value) summary,e.provider_code related_provider,e.user_id related_user,'' related_order,e.created_at detected_time,e.status FROM game_user_risk_events e WHERE e.deleted_at IS NULL AND e.created_at BETWEEN ? AND ?` + eventScope + ` UNION ALL SELECT CONCAT('callback-',c.id),CASE WHEN c.processing_result='FAILED' THEN 'high' ELSE 'medium' END,'PROVIDER_CALLBACK_FAILURE',c.failure_reason,c.provider_code,c.user_id,c.provider_transaction_id,c.created_at,c.processing_result FROM game_provider_callback_logs c WHERE c.deleted_at IS NULL AND c.processing_result<>'SUCCESS' AND c.created_at BETWEEN ? AND ?` + callbackScope + ` ORDER BY detected_time DESC LIMIT 50`
	args := append([]any{f.StartTime, f.EndTime}, eventArgs...)
	args = append(args, f.StartTime, f.EndTime)
	args = append(args, callbackArgs...)
	if err := s.db.Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return map[string]any{"rows": rows}, nil
}
