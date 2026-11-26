package reportadmin

import (
	"fmt"
	"sort"
	"strings"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"casino/internal/tenancy"
)

type FinancialRow struct {
	Period           string          `json:"period"`
	Currency         string          `json:"currency"`
	DepositAmount    decimal.Decimal `json:"depositAmount" gorm:"column:deposit_amount"`
	WithdrawalAmount decimal.Decimal `json:"withdrawalAmount" gorm:"column:withdrawal_amount"`
	NetCashFlow      decimal.Decimal `json:"netCashFlow" gorm:"column:net_cash_flow"`
	TotalBet         decimal.Decimal `json:"totalBet" gorm:"column:total_bet"`
	TotalWin         decimal.Decimal `json:"totalWin" gorm:"column:total_win"`
	GGR              decimal.Decimal `json:"ggr" gorm:"column:ggr"`
	BonusCost        decimal.Decimal `json:"bonusCost" gorm:"column:bonus_cost"`
	ProviderCost     decimal.Decimal `json:"providerCost" gorm:"column:provider_cost"`
	PaymentFee       decimal.Decimal `json:"paymentFee" gorm:"column:payment_fee"`
	AgentCommission  decimal.Decimal `json:"agentCommission" gorm:"column:agent_commission"`
	Adjustments      decimal.Decimal `json:"adjustments"`
	NGR              decimal.Decimal `json:"ngr"`
	ActiveUsers      int64           `json:"activeUsers" gorm:"column:active_users"`
}

func userFactWhere(q Query, timeColumn, currencyColumn, userAlias, scopeAlias string, paymentFilter bool) (string, []any) {
	clauses := []string{timeColumn + " >= ?", timeColumn + " <= ?"}
	args := []any{val(q, "dateFrom"), val(q, "dateTo")}
	if q.Scope.TenantScoped() {
		scopeAlias = strings.TrimSpace(scopeAlias)
		if scopeAlias == "" || scopeAlias == timeColumn {
			scopeAlias = strings.TrimSpace(strings.Split(timeColumn, ".")[0])
		}
		if scopeAlias == "" || scopeAlias == timeColumn {
			scopeAlias = strings.TrimSpace(userAlias)
		}
		clauses = append(clauses, scopeAlias+".tenant_id = ?")
		args = append(args, q.Scope.TenantID)
		if q.Scope.SiteScoped() {
			clauses = append(clauses, scopeAlias+".site_id IN (0, ?)")
			args = append(args, q.Scope.SiteID)
		}
	}
	if value := val(q, "currency"); value != "" {
		clauses = append(clauses, currencyColumn+" = ?")
		args = append(args, value)
	}
	if value := val(q, "country"); value != "" {
		clauses = append(clauses, userAlias+".country_code = ?")
		args = append(args, value)
	}
	if value := val(q, "userChannel"); value != "" {
		clauses = append(clauses, userAlias+".registration_channel = ?")
		args = append(args, value)
	}
	if value := val(q, "paymentChannel"); value != "" && !paymentFilter {
		clauses = append(clauses, "EXISTS (SELECT 1 FROM game_deposit_orders pdc WHERE pdc.user_id="+userAlias+".id AND pdc.channel_code=? AND pdc.status=2 AND pdc.deleted_at IS NULL)")
		args = append(args, value)
	}
	return strings.Join(clauses, " AND "), args
}

func financialFacts(q Query, withPeriod bool) (string, []any, error) {
	depositPeriod, err := periodExpr("d.paid_at", q)
	if err != nil {
		return "", nil, err
	}
	withdrawPeriod, _ := periodExpr("w.completed_at", q)
	betPeriod, _ := periodExpr("b.settled_at", q)
	bonusPeriod, _ := periodExpr("bo.issued_at", q)
	commissionPeriod, _ := periodExpr("c.business_date", q)
	if !withPeriod {
		depositPeriod, withdrawPeriod, betPeriod, bonusPeriod, commissionPeriod = "''", "''", "''", "''", "''"
	}

	dw, da := userFactWhere(q, "d.paid_at", "d.currency", "u", "d", true)
	if value := val(q, "paymentChannel"); value != "" {
		dw += " AND d.channel_code = ?"
		da = append(da, value)
	}
	ww, wa := userFactWhere(q, "w.completed_at", "w.currency", "u", "w", false)
	bw, ba := userFactWhere(q, "b.settled_at", "b.currency", "u", "b", false)
	// Legacy bonus records have no tenant columns. Scope them through their player.
	bow, boa := userFactWhere(q, "bo.issued_at", "bo.currency", "u", "u", false)
	cw, ca := userFactWhere(q, "c.business_date", "c.currency", "u", "c", false)

	branches := []string{
		fmt.Sprintf(`SELECT %s period,d.currency,d.user_id,d.paid_amount deposit_amount,0 withdrawal_amount,0 total_bet,0 total_win,0 bonus_cost,0 provider_cost,d.fee_amount payment_fee,0 agent_commission,0 adjustments FROM game_deposit_orders d JOIN game_users u ON u.id=d.user_id AND u.deleted_at IS NULL WHERE d.status=2 AND d.deleted_at IS NULL AND %s`, depositPeriod, dw),
		fmt.Sprintf(`SELECT %s period,w.currency,w.user_id,0,w.actual_amount,0,0,0,0,w.fee_amount,0,0 FROM game_withdraw_orders w JOIN game_users u ON u.id=w.user_id AND u.deleted_at IS NULL WHERE w.status=2 AND w.deleted_at IS NULL AND %s`, withdrawPeriod, ww),
		fmt.Sprintf(`SELECT %s period,b.currency,b.user_id,0,0,b.bet_amount,b.payout_amount,0,0,0,0,0 FROM game_bet_records b JOIN game_users u ON u.id=b.user_id AND u.deleted_at IS NULL WHERE b.bet_status='SETTLED' AND COALESCE(b.rollback_status,'')<>'ROLLED_BACK' AND b.deleted_at IS NULL AND %s`, betPeriod, bw),
		fmt.Sprintf(`SELECT %s period,bo.currency,bo.user_id,0,0,0,0,bo.amount,0,0,0,0 FROM game_bonus_records bo JOIN game_users u ON u.id=bo.user_id AND u.deleted_at IS NULL WHERE bo.status_code IN ('ACTIVE','COMPLETED') AND bo.deleted_at IS NULL AND %s`, bonusPeriod, bow),
		fmt.Sprintf(`SELECT %s period,c.currency,c.user_id,0,0,0,0,0,0,0,c.commission_amount,0 FROM game_commission_records c LEFT JOIN game_users u ON u.id=c.user_id AND u.deleted_at IS NULL WHERE c.status IN ('CONFIRMED','FROZEN','SETTLED') AND c.deleted_at IS NULL AND %s`, commissionPeriod, cw),
	}
	args := append(append(append(append(da, wa...), ba...), boa...), ca...)
	return strings.Join(branches, " UNION ALL "), args, nil
}

func financialSelect(periodSelect, groupBy string) string {
	return `SELECT ` + periodSelect + `,currency,
COALESCE(SUM(deposit_amount),0) deposit_amount,COALESCE(SUM(withdrawal_amount),0) withdrawal_amount,
COALESCE(SUM(deposit_amount-withdrawal_amount),0) net_cash_flow,
COALESCE(SUM(total_bet),0) total_bet,COALESCE(SUM(total_win),0) total_win,
COALESCE(SUM(total_bet-total_win),0) ggr,COALESCE(SUM(bonus_cost),0) bonus_cost,
COALESCE(SUM(provider_cost),0) provider_cost,COALESCE(SUM(payment_fee),0) payment_fee,
COALESCE(SUM(agent_commission),0) agent_commission,COALESCE(SUM(adjustments),0) adjustments,
COALESCE(SUM(total_bet-total_win-bonus_cost-provider_cost-payment_fee-agent_commission-adjustments),0) ngr,
COUNT(DISTINCT user_id) active_users FROM facts GROUP BY ` + groupBy
}

func (s *FinancialReportService) Report(q Query) (Result, error) {
	if err := validateRange(q); err != nil {
		return Result{}, err
	}
	facts, args, err := financialFacts(q, true)
	if err != nil {
		return Result{}, err
	}
	rows := []FinancialRow{}
	if err = s.db.Raw("WITH facts AS ("+facts+") "+financialSelect("period", "period,currency")+" ORDER BY period DESC,currency", args...).Scan(&rows).Error; err != nil {
		return Result{}, err
	}
	summaryFacts, summaryArgs, _ := financialFacts(q, false)
	summaryRows := []FinancialRow{}
	if err = s.db.Raw("WITH facts AS ("+summaryFacts+") "+financialSelect("'' AS period", "currency")+" ORDER BY currency", summaryArgs...).Scan(&summaryRows).Error; err != nil {
		return Result{}, err
	}
	pageRows, total, page, pageSize := paginateRows(rows, q)
	return Result{SummaryByCurrency: financialSummaries(summaryRows), List: pageRows, Total: total, Page: page, PageSize: pageSize, Charts: financialCharts(rows), MetricBasis: metricBasis()}, nil
}

func financialSummaries(rows []FinancialRow) []map[string]any {
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]any{
			"currency": row.Currency, "depositAmount": decimalString(row.DepositAmount), "withdrawalAmount": decimalString(row.WithdrawalAmount),
			"netCashFlow": decimalString(row.NetCashFlow), "totalBet": decimalString(row.TotalBet), "totalWin": decimalString(row.TotalWin),
			"ggr": decimalString(row.GGR), "bonusCost": decimalString(row.BonusCost), "providerCost": decimalString(row.ProviderCost),
			"paymentFee": decimalString(row.PaymentFee), "agentCommission": decimalString(row.AgentCommission), "ngr": decimalString(row.NGR), "activeUsers": row.ActiveUsers,
		})
	}
	return out
}

func financialCharts(rows []FinancialRow) []Chart {
	ordered := append([]FinancialRow(nil), rows...)
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].Period == ordered[j].Period {
			return ordered[i].Currency < ordered[j].Currency
		}
		return ordered[i].Period < ordered[j].Period
	})
	return []Chart{
		financialChart("cashFlow", ordered, []string{"depositAmount", "withdrawalAmount"}),
		financialChart("revenue", ordered, []string{"ggr", "ngr"}),
		financialChart("betting", ordered, []string{"totalBet", "totalWin"}),
	}
}

func financialChart(key string, rows []FinancialRow, metrics []string) Chart {
	categories := []string{}
	categoryIndexes := map[string]int{}
	seriesData := map[string][]string{}
	for _, row := range rows {
		if _, exists := categoryIndexes[row.Period]; !exists {
			categoryIndexes[row.Period] = len(categories)
			categories = append(categories, row.Period)
		}
	}
	for _, metric := range metrics {
		for _, row := range rows {
			seriesKey := metric + "|" + row.Currency
			if _, exists := seriesData[seriesKey]; !exists {
				seriesData[seriesKey] = make([]string, len(categories))
				for index := range seriesData[seriesKey] {
					seriesData[seriesKey][index] = "0.00"
				}
			}
			seriesData[seriesKey][categoryIndexes[row.Period]] = financialMetric(row, metric)
		}
	}
	keys := make([]string, 0, len(seriesData))
	for key := range seriesData {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	series := make([]ChartSeries, 0, len(keys))
	for _, seriesKey := range keys {
		parts := strings.Split(seriesKey, "|")
		series = append(series, ChartSeries{Name: parts[0], Currency: parts[1], Data: seriesData[seriesKey]})
	}
	return Chart{Key: key, Categories: categories, Series: series}
}

func financialMetric(row FinancialRow, key string) string {
	switch key {
	case "depositAmount":
		return decimalString(row.DepositAmount)
	case "withdrawalAmount":
		return decimalString(row.WithdrawalAmount)
	case "ggr":
		return decimalString(row.GGR)
	case "ngr":
		return decimalString(row.NGR)
	case "totalBet":
		return decimalString(row.TotalBet)
	default:
		return decimalString(row.TotalWin)
	}
}

func (s *FinancialReportService) Detail(q Query) (map[string]any, error) {
	result, err := s.Report(q)
	if err != nil {
		return nil, err
	}
	return map[string]any{"summary": result.SummaryByCurrency, "deposits": result.List, "withdrawals": result.List, "betting": result.List, "costs": result.List, "adjustments": []any{}, "metricBasis": result.MetricBasis}, nil
}

func reportScoped(db *gorm.DB, alias string, q Query) *gorm.DB {
	return tenancy.ApplySite(db, alias, q.Scope)
}

func reportScopeSQL(q Query, alias string) (string, []any) {
	if !q.Scope.TenantScoped() {
		return "", nil
	}
	prefix := ""
	if strings.TrimSpace(alias) != "" {
		prefix = strings.TrimSpace(alias) + "."
	}
	where := " AND " + prefix + "tenant_id=?"
	args := []any{q.Scope.TenantID}
	if q.Scope.SiteScoped() {
		where += " AND " + prefix + "site_id IN (0, ?)"
		args = append(args, q.Scope.SiteID)
	}
	return where, args
}

func applyEqual(db *gorm.DB, column, value string) *gorm.DB {
	if strings.TrimSpace(value) != "" {
		return db.Where(column+" = ?", strings.TrimSpace(value))
	}
	return db
}
