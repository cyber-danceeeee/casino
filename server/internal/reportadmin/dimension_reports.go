package reportadmin

import (
	"fmt"
	"sort"
	"strings"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type GameRow struct {
	GameID         uint64          `json:"gameID" gorm:"column:game_id"`
	GameCode       string          `json:"gameCode" gorm:"column:game_code"`
	Game           string          `json:"game"`
	Provider       string          `json:"provider"`
	GameType       string          `json:"gameType" gorm:"column:game_type"`
	Currency       string          `json:"currency"`
	UniquePlayers  int64           `json:"uniquePlayers" gorm:"column:unique_players"`
	Rounds         int64           `json:"rounds"`
	TotalBet       decimal.Decimal `json:"totalBet" gorm:"column:total_bet"`
	ValidBet       decimal.Decimal `json:"validBet" gorm:"column:valid_bet"`
	TotalWin       decimal.Decimal `json:"totalWin" gorm:"column:total_win"`
	GGR            decimal.Decimal `json:"ggr"`
	ActualRTP      *string         `json:"actualRTP" gorm:"-"`
	AverageBet     decimal.Decimal `json:"averageBet" gorm:"column:average_bet"`
	MaximumWin     decimal.Decimal `json:"maximumWin" gorm:"column:maximum_win"`
	ProviderGameID string          `json:"providerGameID" gorm:"column:provider_game_id"`
}

func reportBetBase(db *gorm.DB, q Query) *gorm.DB {
	db = db.Table("game_bet_records b").Joins("JOIN game_games g ON g.id=b.game_id AND g.deleted_at IS NULL").Joins("LEFT JOIN game_users u ON u.id=b.user_id AND u.deleted_at IS NULL").Where("b.deleted_at IS NULL AND b.bet_status='SETTLED' AND COALESCE(b.rollback_status,'')<>'ROLLED_BACK'").Where("b.settled_at>=? AND b.settled_at<=?", val(q, "dateFrom"), val(q, "dateTo"))
	db = reportScoped(db, "b", q)
	db = applyEqual(db, "b.provider_code", val(q, "provider"))
	db = applyEqual(db, "b.game_id", val(q, "gameID"))
	db = applyEqual(db, "g.game_code", val(q, "gameCode"))
	if name := val(q, "gameName"); name != "" {
		db = db.Where("g.name LIKE ?", "%"+name+"%")
	}
	db = applyEqual(db, "b.provider_game_id", val(q, "providerGameID"))
	db = applyEqual(db, "b.game_type", val(q, "gameType"))
	db = applyEqual(db, "b.currency", val(q, "currency"))
	db = applyEqual(db, "u.country_code", val(q, "country"))
	if category := val(q, "platformCategory"); category != "" {
		db = db.Where("g.platform_category_code=? OR EXISTS (SELECT 1 FROM game_game_category_relations cr JOIN game_game_categories gc ON gc.id=cr.category_id WHERE cr.game_id=g.id AND gc.category_code=? AND cr.deleted_at IS NULL)", category, category)
	}
	if device := val(q, "device"); device != "" {
		db = db.Where("EXISTS (SELECT 1 FROM game_game_launch_logs gl WHERE gl.user_id=b.user_id AND gl.game_id=b.game_id AND gl.device=? AND gl.deleted_at IS NULL)", device)
	}
	return db
}

func (s *GameReportService) Report(q Query) (Result, error) {
	if err := validateRange(q); err != nil {
		return Result{}, err
	}
	rows := []GameRow{}
	db := reportBetBase(s.db, q).Select(`b.game_id,g.game_code,g.name game,b.provider_code provider,b.game_type,b.currency,b.provider_game_id,
COUNT(DISTINCT b.user_id) unique_players,COUNT(DISTINCT b.provider_round_id) rounds,
COALESCE(SUM(b.bet_amount),0) total_bet,COALESCE(SUM(b.valid_bet_amount),0) valid_bet,
COALESCE(SUM(b.payout_amount),0) total_win,COALESCE(SUM(b.bet_amount-b.payout_amount),0) ggr,
COALESCE(AVG(b.bet_amount),0) average_bet,COALESCE(MAX(b.payout_amount),0) maximum_win`).
		Group("b.game_id,g.game_code,g.name,b.provider_code,b.game_type,b.currency,b.provider_game_id").Order("total_bet DESC,b.game_id")
	if err := db.Scan(&rows).Error; err != nil {
		return Result{}, err
	}
	for i := range rows {
		rows[i].ActualRTP = percentage(rows[i].TotalWin, rows[i].TotalBet)
	}
	pageRows, total, page, pageSize := paginateRows(rows, q)
	summary := summarizeGames(rows)
	trend, err := gameTrend(s.db, q)
	if err != nil {
		return Result{}, err
	}
	return Result{SummaryByCurrency: summary, List: pageRows, Total: total, Page: page, PageSize: pageSize, Charts: []Chart{trend}, MetricBasis: metricBasis()}, nil
}

func summarizeGames(rows []GameRow) []map[string]any {
	type aggregate struct {
		games         map[uint64]bool
		players       int64
		rounds        int64
		bet, win, max decimal.Decimal
	}
	groups := map[string]*aggregate{}
	for _, row := range rows {
		a := groups[row.Currency]
		if a == nil {
			a = &aggregate{games: map[uint64]bool{}}
			groups[row.Currency] = a
		}
		a.games[row.GameID] = true
		a.players += row.UniquePlayers
		a.rounds += row.Rounds
		a.bet = a.bet.Add(row.TotalBet)
		a.win = a.win.Add(row.TotalWin)
		if row.MaximumWin.GreaterThan(a.max) {
			a.max = row.MaximumWin
		}
	}
	currencies := sortedKeys(groups)
	out := make([]map[string]any, 0, len(currencies))
	for _, currency := range currencies {
		a := groups[currency]
		out = append(out, map[string]any{"currency": currency, "gameCount": len(a.games), "uniquePlayers": a.players, "totalRounds": a.rounds, "totalBet": decimalString(a.bet), "totalWin": decimalString(a.win), "ggr": decimalString(a.bet.Sub(a.win)), "actualRTP": percentage(a.win, a.bet), "averageBet": decimalString(a.bet.Div(decimal.NewFromInt(maxInt64(a.rounds, 1)))), "maximumWin": decimalString(a.max)})
	}
	return out
}

func gameTrend(db *gorm.DB, q Query) (Chart, error) {
	expr, err := periodExpr("b.settled_at", q)
	if err != nil {
		return Chart{}, err
	}
	rows := []struct {
		Period, Currency   string
		TotalBet, TotalWin decimal.Decimal
	}{}
	err = reportBetBase(db, q).Select(fmt.Sprintf("%s period,b.currency,COALESCE(SUM(b.bet_amount),0) total_bet,COALESCE(SUM(b.payout_amount),0) total_win", expr)).Group("period,b.currency").Order("period,b.currency").Scan(&rows).Error
	if err != nil {
		return Chart{}, err
	}
	return chartFromMetrics("gameTrend", rows, func(r struct {
		Period, Currency   string
		TotalBet, TotalWin decimal.Decimal
	}) map[string]decimal.Decimal {
		return map[string]decimal.Decimal{"totalBet": r.TotalBet, "totalWin": r.TotalWin}
	}), nil
}

func (s *GameReportService) Detail(id uint64, q Query) (map[string]any, error) {
	q.Values["gameID"] = fmt.Sprint(id)
	result, err := s.Report(q)
	if err != nil {
		return nil, err
	}
	var game map[string]any
	if err = reportScoped(s.db.Table("game_games g"), "g", q).Select("g.id,g.name,g.game_code,g.provider_game_code,g.game_type,g.platform_category_code,g.supports_demo,g.supports_free_spins").Where("g.id=? AND g.deleted_at IS NULL", id).Take(&game).Error; err != nil {
		return nil, err
	}
	return map[string]any{"overview": game, "trend": result.Charts, "users": result.SummaryByCurrency, "bettingRecords": result.List, "roundRecords": result.List}, nil
}

type ProviderRow struct {
	Provider               string          `json:"provider"`
	Currency               string          `json:"currency"`
	GameCount              int64           `json:"gameCount" gorm:"column:game_count"`
	UniquePlayers          int64           `json:"uniquePlayers" gorm:"column:unique_players"`
	TotalBet               decimal.Decimal `json:"totalBet" gorm:"column:total_bet"`
	TotalWin               decimal.Decimal `json:"totalWin" gorm:"column:total_win"`
	GGR                    decimal.Decimal `json:"ggr"`
	ActualRTP              *string         `json:"actualRTP" gorm:"-"`
	LaunchCount            int64           `json:"launchCount" gorm:"-"`
	LaunchSuccessRate      *string         `json:"launchSuccessRate" gorm:"-"`
	CallbackCount          int64           `json:"callbackCount" gorm:"-"`
	CallbackSuccessRate    *string         `json:"callbackSuccessRate" gorm:"-"`
	AverageCallbackLatency decimal.Decimal `json:"averageCallbackLatency" gorm:"-"`
}

type countMetric struct {
	Count, Success int64
	Latency        decimal.Decimal
}

func (s *ProviderReportService) Report(q Query) (Result, error) {
	if err := validateRange(q); err != nil {
		return Result{}, err
	}
	rows := []ProviderRow{}
	if err := reportBetBase(s.db, q).Select(`b.provider_code provider,b.currency,COUNT(DISTINCT b.game_id) game_count,COUNT(DISTINCT b.user_id) unique_players,COALESCE(SUM(b.bet_amount),0) total_bet,COALESCE(SUM(b.payout_amount),0) total_win,COALESCE(SUM(b.bet_amount-b.payout_amount),0) ggr`).Group("b.provider_code,b.currency").Order("total_bet DESC,b.provider_code").Scan(&rows).Error; err != nil {
		return Result{}, err
	}
	launches, callbacks, err := s.providerMetrics(q)
	if err != nil {
		return Result{}, err
	}
	for i := range rows {
		key := rows[i].Provider + "|" + rows[i].Currency
		rows[i].ActualRTP = percentage(rows[i].TotalWin, rows[i].TotalBet)
		rows[i].LaunchCount = launches[key].Count
		rows[i].LaunchSuccessRate = percentage(decimal.NewFromInt(launches[key].Success), decimal.NewFromInt(launches[key].Count))
		rows[i].CallbackCount = callbacks[key].Count
		rows[i].CallbackSuccessRate = percentage(decimal.NewFromInt(callbacks[key].Success), decimal.NewFromInt(callbacks[key].Count))
		rows[i].AverageCallbackLatency = callbacks[key].Latency
	}
	pageRows, total, page, pageSize := paginateRows(rows, q)
	summary := summarizeProviders(rows)
	trend, err := gameTrend(s.db, q)
	if err != nil {
		return Result{}, err
	}
	return Result{SummaryByCurrency: summary, List: pageRows, Total: total, Page: page, PageSize: pageSize, Charts: []Chart{trend}, MetricBasis: metricBasis()}, nil
}

func (s *ProviderReportService) providerMetrics(q Query) (map[string]countMetric, map[string]countMetric, error) {
	launchRows := []struct {
		Provider, Currency string
		Count, Success     int64
	}{}
	db := s.db.Table("game_game_launch_logs l").Select("p.provider_code provider,l.currency,COUNT(*) count,SUM(CASE WHEN l.launch_status=2 THEN 1 ELSE 0 END) success").Joins("JOIN game_game_providers p ON p.id=l.provider_id").Where("l.deleted_at IS NULL AND l.created_at>=? AND l.created_at<=?", val(q, "dateFrom"), val(q, "dateTo"))
	db = reportScoped(db, "l", q)
	db = applyEqual(db, "p.provider_code", val(q, "provider"))
	db = applyEqual(db, "l.currency", val(q, "currency"))
	if err := db.Group("p.provider_code,l.currency").Scan(&launchRows).Error; err != nil {
		return nil, nil, err
	}
	callbackRows := []struct {
		Provider, Currency string
		Count, Success     int64
		Latency            decimal.Decimal
	}{}
	db = s.db.Table("game_provider_callback_logs c").Select("c.provider_code provider,c.currency,COUNT(*) count,SUM(CASE WHEN c.processing_result='SUCCESS' THEN 1 ELSE 0 END) success,COALESCE(AVG(c.duration_ms),0) latency").Where("c.deleted_at IS NULL AND c.created_at>=? AND c.created_at<=?", val(q, "dateFrom"), val(q, "dateTo"))
	db = reportScoped(db, "c", q)
	db = applyEqual(db, "c.provider_code", val(q, "provider"))
	db = applyEqual(db, "c.currency", val(q, "currency"))
	if err := db.Group("c.provider_code,c.currency").Scan(&callbackRows).Error; err != nil {
		return nil, nil, err
	}
	launches, callbacks := map[string]countMetric{}, map[string]countMetric{}
	for _, r := range launchRows {
		launches[r.Provider+"|"+r.Currency] = countMetric{Count: r.Count, Success: r.Success}
	}
	for _, r := range callbackRows {
		callbacks[r.Provider+"|"+r.Currency] = countMetric{Count: r.Count, Success: r.Success, Latency: r.Latency}
	}
	return launches, callbacks, nil
}

func summarizeProviders(rows []ProviderRow) []map[string]any {
	type agg struct {
		providers                           map[string]bool
		games, players, launches, callbacks int64
		bet, win                            decimal.Decimal
		launchSuccess, callbackSuccess      decimal.Decimal
	}
	groups := map[string]*agg{}
	for _, r := range rows {
		a := groups[r.Currency]
		if a == nil {
			a = &agg{providers: map[string]bool{}}
			groups[r.Currency] = a
		}
		a.providers[r.Provider] = true
		a.games += r.GameCount
		a.players += r.UniquePlayers
		a.bet = a.bet.Add(r.TotalBet)
		a.win = a.win.Add(r.TotalWin)
		a.launches += r.LaunchCount
		a.callbacks += r.CallbackCount
		if r.LaunchSuccessRate != nil {
			a.launchSuccess = a.launchSuccess.Add(parseDecimal(*r.LaunchSuccessRate).Mul(decimal.NewFromInt(r.LaunchCount)).Div(decimal.NewFromInt(100)))
		}
		if r.CallbackSuccessRate != nil {
			a.callbackSuccess = a.callbackSuccess.Add(parseDecimal(*r.CallbackSuccessRate).Mul(decimal.NewFromInt(r.CallbackCount)).Div(decimal.NewFromInt(100)))
		}
	}
	out := []map[string]any{}
	for _, currency := range sortedKeys(groups) {
		a := groups[currency]
		out = append(out, map[string]any{"currency": currency, "providerCount": len(a.providers), "totalGames": a.games, "uniquePlayers": a.players, "totalBet": decimalString(a.bet), "totalWin": decimalString(a.win), "ggr": decimalString(a.bet.Sub(a.win)), "actualRTP": percentage(a.win, a.bet), "launchSuccessRate": percentage(a.launchSuccess, decimal.NewFromInt(a.launches)), "callbackSuccessRate": percentage(a.callbackSuccess, decimal.NewFromInt(a.callbacks))})
	}
	return out
}

func (s *ProviderReportService) Detail(code string, q Query) (map[string]any, error) {
	q.Values["provider"] = code
	r, e := s.Report(q)
	if e != nil {
		return nil, e
	}
	return map[string]any{"overview": r.SummaryByCurrency, "games": r.List, "bettingTrend": r.Charts, "callbackMetrics": r.List, "launchMetrics": r.List, "errors": []any{}}, nil
}

type UserRow struct {
	UserID              uint64          `json:"userID" gorm:"column:user_id"`
	PlayerNo            string          `json:"playerNo" gorm:"column:player_no"`
	Username            string          `json:"username"`
	Country             string          `json:"country"`
	Currency            string          `json:"currency"`
	RegistrationChannel string          `json:"registrationChannel" gorm:"column:registration_channel"`
	Agent               string          `json:"agent"`
	TotalDeposit        decimal.Decimal `json:"totalDeposit" gorm:"column:total_deposit"`
	TotalWithdrawal     decimal.Decimal `json:"totalWithdrawal" gorm:"column:total_withdrawal"`
	TotalBet            decimal.Decimal `json:"totalBet" gorm:"column:total_bet"`
	TotalWin            decimal.Decimal `json:"totalWin" gorm:"column:total_win"`
	UserNetResult       decimal.Decimal `json:"userNetResult" gorm:"column:user_net_result"`
	BonusAmount         decimal.Decimal `json:"bonusAmount" gorm:"column:bonus_amount"`
	LastActiveAt        string          `json:"lastActiveAt" gorm:"column:last_active_at"`
}

func (s *UserReportService) Report(q Query) (Result, error) {
	if e := validateRange(q); e != nil {
		return Result{}, e
	}
	from, to := val(q, "dateFrom"), val(q, "dateTo")
	selectSQL := `u.id user_id,u.user_no player_no,u.username,u.country_code country,w.currency,u.registration_channel,COALESCE(a.name,'') agent,
COALESCE((SELECT SUM(d.paid_amount) FROM game_deposit_orders d WHERE d.user_id=u.id AND d.currency=w.currency AND d.status=2 AND d.deleted_at IS NULL AND d.paid_at>=? AND d.paid_at<=?),0) total_deposit,
COALESCE((SELECT SUM(x.actual_amount) FROM game_withdraw_orders x WHERE x.user_id=u.id AND x.currency=w.currency AND x.status=2 AND x.deleted_at IS NULL AND x.completed_at>=? AND x.completed_at<=?),0) total_withdrawal,
COALESCE((SELECT SUM(b.bet_amount) FROM game_bet_records b WHERE b.user_id=u.id AND b.currency=w.currency AND b.bet_status='SETTLED' AND COALESCE(b.rollback_status,'')<>'ROLLED_BACK' AND b.deleted_at IS NULL AND b.settled_at>=? AND b.settled_at<=?),0) total_bet,
COALESCE((SELECT SUM(b.payout_amount) FROM game_bet_records b WHERE b.user_id=u.id AND b.currency=w.currency AND b.bet_status='SETTLED' AND COALESCE(b.rollback_status,'')<>'ROLLED_BACK' AND b.deleted_at IS NULL AND b.settled_at>=? AND b.settled_at<=?),0) total_win,
COALESCE((SELECT SUM(b.payout_amount-b.bet_amount) FROM game_bet_records b WHERE b.user_id=u.id AND b.currency=w.currency AND b.bet_status='SETTLED' AND COALESCE(b.rollback_status,'')<>'ROLLED_BACK' AND b.deleted_at IS NULL AND b.settled_at>=? AND b.settled_at<=?),0) user_net_result,
COALESCE((SELECT SUM(bo.amount) FROM game_bonus_records bo WHERE bo.user_id=u.id AND bo.currency=w.currency AND bo.status_code IN ('ACTIVE','COMPLETED') AND bo.deleted_at IS NULL AND bo.issued_at>=? AND bo.issued_at<=?),0) bonus_amount,
DATE_FORMAT(GREATEST(COALESCE(u.last_login_at,u.created_at),COALESCE(w.last_transaction_at,u.created_at)),'%Y-%m-%d %H:%i:%s') last_active_at`
	args := []any{from, to, from, to, from, to, from, to, from, to, from, to}
	db := s.db.Table("game_users u").Select(selectSQL, args...).Joins("JOIN game_user_wallets w ON w.user_id=u.id AND w.deleted_at IS NULL").Joins("LEFT JOIN game_user_agent_attributions ua ON ua.user_id=u.id AND ua.active=1 AND ua.deleted_at IS NULL").Joins("LEFT JOIN game_agents a ON a.id=ua.agent_id AND a.deleted_at IS NULL").Joins("LEFT JOIN game_user_risk_profiles rp ON rp.user_id=u.id AND rp.deleted_at IS NULL").Where("u.deleted_at IS NULL")
	db = reportScoped(db, "u", q)
	db = reportScoped(db, "w", q)
	db = applyEqual(db, "u.id", val(q, "userID"))
	db = applyEqual(db, "u.user_no", val(q, "playerNo"))
	if v := val(q, "username"); v != "" {
		db = db.Where("u.username LIKE ?", "%"+v+"%")
	}
	db = applyEqual(db, "u.country_code", val(q, "country"))
	db = applyEqual(db, "w.currency", val(q, "currency"))
	db = applyEqual(db, "u.vip_level", val(q, "vipLevel"))
	db = applyEqual(db, "u.registration_channel", val(q, "registrationChannel"))
	db = applyEqual(db, "a.id", val(q, "agent"))
	db = applyEqual(db, "rp.risk_status", val(q, "riskStatus"))
	rows := []UserRow{}
	if e := db.Order("total_bet DESC,u.id").Scan(&rows).Error; e != nil {
		return Result{}, e
	}
	pageRows, total, page, ps := paginateRows(rows, q)
	summary := summarizeUsers(rows, from, to, s.db, q)
	return Result{SummaryByCurrency: summary, List: pageRows, Total: total, Page: page, PageSize: ps, Charts: []Chart{}, MetricBasis: metricBasis()}, nil
}

func summarizeUsers(rows []UserRow, from string, to string, db *gorm.DB, q Query) []map[string]any {
	type agg struct {
		users               map[uint64]bool
		depositors, bettors int64
	}
	groups := map[string]*agg{}
	for _, r := range rows {
		a := groups[r.Currency]
		if a == nil {
			a = &agg{users: map[uint64]bool{}}
			groups[r.Currency] = a
		}
		a.users[r.UserID] = true
		if r.TotalDeposit.GreaterThan(decimal.Zero) {
			a.depositors++
		}
		if r.TotalBet.GreaterThan(decimal.Zero) {
			a.bettors++
		}
	}
	var registrations int64
	reportScoped(db.Table("game_users"), "", q).Where("deleted_at IS NULL AND created_at>=? AND created_at<=?", from, to).Count(&registrations)
	out := []map[string]any{}
	for _, currency := range sortedKeys(groups) {
		a := groups[currency]
		out = append(out, map[string]any{"currency": currency, "registeredUsers": registrations, "activeUsers": len(a.users), "depositingUsers": a.depositors, "firstDepositUsers": a.depositors, "bettingUsers": a.bettors, "returningUsers": 0, "churnedUsers": 0})
	}
	return out
}

func (s *UserReportService) Detail(id uint64, q Query) (map[string]any, error) {
	q.Values["userID"] = fmt.Sprint(id)
	r, e := s.Report(q)
	if e != nil {
		return nil, e
	}
	if r.Total == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return map[string]any{"overview": r.List, "financialSummary": r.SummaryByCurrency, "bettingSummary": r.List, "bonusSummary": r.List, "activityTrend": r.Charts}, nil
}

func maxInt64(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func sortedKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func chartFromMetrics[T any](key string, rows []T, metrics func(T) map[string]decimal.Decimal) Chart {
	categories := []string{}
	seriesMap := map[string][]string{}
	for _, row := range rows {
		value := any(row)
		period := ""
		currency := ""
		switch r := value.(type) {
		case struct {
			Period, Currency   string
			TotalBet, TotalWin decimal.Decimal
		}:
			period = r.Period
			currency = r.Currency
		}
		categories = append(categories, period)
		for metric, amount := range metrics(row) {
			seriesMap[metric+"|"+currency] = append(seriesMap[metric+"|"+currency], decimalString(amount))
		}
	}
	series := []ChartSeries{}
	for _, name := range sortedKeys(seriesMap) {
		parts := strings.Split(name, "|")
		series = append(series, ChartSeries{Name: parts[0], Currency: parts[1], Data: seriesMap[name]})
	}
	return Chart{Key: key, Categories: categories, Series: series}
}
