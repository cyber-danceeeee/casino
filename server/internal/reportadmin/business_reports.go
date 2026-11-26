package reportadmin

import (
	"fmt"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ChannelRow struct {
	ChannelID              uint64          `json:"channelID" gorm:"column:channel_id"`
	Channel                string          `json:"channel"`
	Agent                  string          `json:"agent"`
	Currency               string          `json:"currency"`
	Registrations          int64           `json:"registrations"`
	FirstDepositUsers      int64           `json:"firstDepositUsers" gorm:"column:first_deposit_users"`
	FirstDepositConversion *string         `json:"firstDepositConversion" gorm:"-"`
	DepositingUsers        int64           `json:"depositingUsers" gorm:"column:depositing_users"`
	DepositAmount          decimal.Decimal `json:"depositAmount" gorm:"column:deposit_amount"`
	BettingUsers           int64           `json:"bettingUsers" gorm:"column:betting_users"`
	TotalBet               decimal.Decimal `json:"totalBet" gorm:"column:total_bet"`
	GGR                    decimal.Decimal `json:"ggr"`
	NGR                    decimal.Decimal `json:"ngr"`
}

func (s *ChannelReportService) Report(q Query) (Result, error) {
	if err := validateRange(q); err != nil {
		return Result{}, err
	}
	from, to := val(q, "dateFrom"), val(q, "dateTo")
	selectSQL := `ch.id channel_id,COALESCE(NULLIF(ch.name_en,''),ch.name_zh) channel,a.name agent,w.currency,
COUNT(DISTINCT CASE WHEN ua.attributed_at>=? AND ua.attributed_at<=? THEN ua.user_id END) registrations,
COUNT(DISTINCT CASE WHEN EXISTS(SELECT 1 FROM game_deposit_orders fd WHERE fd.user_id=ua.user_id AND fd.currency=w.currency AND fd.status=2 AND fd.deleted_at IS NULL AND fd.paid_at>=? AND fd.paid_at<=? AND NOT EXISTS(SELECT 1 FROM game_deposit_orders oldfd WHERE oldfd.user_id=fd.user_id AND oldfd.status=2 AND oldfd.deleted_at IS NULL AND oldfd.paid_at<fd.paid_at)) THEN ua.user_id END) first_deposit_users,
COUNT(DISTINCT CASE WHEN EXISTS(SELECT 1 FROM game_deposit_orders d WHERE d.user_id=ua.user_id AND d.currency=w.currency AND d.status=2 AND d.deleted_at IS NULL AND d.paid_at>=? AND d.paid_at<=?) THEN ua.user_id END) depositing_users,
COALESCE(SUM((SELECT SUM(d.paid_amount) FROM game_deposit_orders d WHERE d.user_id=ua.user_id AND d.currency=w.currency AND d.status=2 AND d.deleted_at IS NULL AND d.paid_at>=? AND d.paid_at<=?)),0) deposit_amount,
COUNT(DISTINCT CASE WHEN EXISTS(SELECT 1 FROM game_bet_records b WHERE b.user_id=ua.user_id AND b.currency=w.currency AND b.bet_status='SETTLED' AND COALESCE(b.rollback_status,'')<>'ROLLED_BACK' AND b.deleted_at IS NULL AND b.settled_at>=? AND b.settled_at<=?) THEN ua.user_id END) betting_users,
COALESCE(SUM((SELECT SUM(b.bet_amount) FROM game_bet_records b WHERE b.user_id=ua.user_id AND b.currency=w.currency AND b.bet_status='SETTLED' AND COALESCE(b.rollback_status,'')<>'ROLLED_BACK' AND b.deleted_at IS NULL AND b.settled_at>=? AND b.settled_at<=?)),0) total_bet,
COALESCE(SUM((SELECT SUM(b.bet_amount-b.payout_amount) FROM game_bet_records b WHERE b.user_id=ua.user_id AND b.currency=w.currency AND b.bet_status='SETTLED' AND COALESCE(b.rollback_status,'')<>'ROLLED_BACK' AND b.deleted_at IS NULL AND b.settled_at>=? AND b.settled_at<=?)),0) ggr,
COALESCE(SUM((SELECT SUM(b.bet_amount-b.payout_amount) FROM game_bet_records b WHERE b.user_id=ua.user_id AND b.currency=w.currency AND b.bet_status='SETTLED' AND COALESCE(b.rollback_status,'')<>'ROLLED_BACK' AND b.deleted_at IS NULL AND b.settled_at>=? AND b.settled_at<=?)),0)-COALESCE(SUM((SELECT SUM(bo.amount) FROM game_bonus_records bo WHERE bo.user_id=ua.user_id AND bo.currency=w.currency AND bo.status_code IN ('ACTIVE','COMPLETED') AND bo.deleted_at IS NULL AND bo.issued_at>=? AND bo.issued_at<=?)),0) ngr`
	args := []any{from, to, from, to, from, to, from, to, from, to, from, to, from, to, from, to, from, to}
	db := s.db.Table("game_agent_channels ch").Select(selectSQL, args...).Joins("JOIN game_user_agent_attributions ua ON ua.channel_id=ch.id AND ua.active=1 AND ua.deleted_at IS NULL").Joins("JOIN game_user_wallets w ON w.user_id=ua.user_id AND w.deleted_at IS NULL").Joins("JOIN game_agents a ON a.id=ch.agent_id AND a.deleted_at IS NULL").Where("ch.deleted_at IS NULL")
	db = reportScoped(db, "ch", q)
	db = reportScoped(db, "ua", q)
	db = reportScoped(db, "w", q)
	db = reportScoped(db, "a", q)
	if v := val(q, "channel"); v != "" {
		db = db.Where("ch.id=? OR ch.code LIKE ?", v, "%"+v+"%")
	}
	db = applyEqual(db, "ch.agent_id", val(q, "agent"))
	db = applyEqual(db, "ch.country", val(q, "country"))
	db = applyEqual(db, "ch.platform", val(q, "platform"))
	db = applyEqual(db, "w.currency", val(q, "currency"))
	rows := []ChannelRow{}
	if err := db.Group("ch.id,ch.name_en,ch.name_zh,a.name,w.currency").Order("deposit_amount DESC,ch.id").Scan(&rows).Error; err != nil {
		return Result{}, err
	}
	for i := range rows {
		rows[i].FirstDepositConversion = percentage(decimal.NewFromInt(rows[i].FirstDepositUsers), decimal.NewFromInt(rows[i].Registrations))
	}
	pageRows, total, page, ps := paginateRows(rows, q)
	return Result{SummaryByCurrency: summarizeChannels(rows), List: pageRows, Total: total, Page: page, PageSize: ps, MetricBasis: metricBasis()}, nil
}

func summarizeChannels(rows []ChannelRow) []map[string]any {
	type agg struct {
		registrations, first, depositors, bettors int64
		deposit, bet, ggr, ngr                    decimal.Decimal
	}
	groups := map[string]*agg{}
	for _, r := range rows {
		a := groups[r.Currency]
		if a == nil {
			a = &agg{}
			groups[r.Currency] = a
		}
		a.registrations += r.Registrations
		a.first += r.FirstDepositUsers
		a.depositors += r.DepositingUsers
		a.bettors += r.BettingUsers
		a.deposit = a.deposit.Add(r.DepositAmount)
		a.bet = a.bet.Add(r.TotalBet)
		a.ggr = a.ggr.Add(r.GGR)
		a.ngr = a.ngr.Add(r.NGR)
	}
	out := []map[string]any{}
	for _, c := range sortedKeys(groups) {
		a := groups[c]
		out = append(out, map[string]any{"currency": c, "registrations": a.registrations, "firstDepositUsers": a.first, "depositingUsers": a.depositors, "depositAmount": decimalString(a.deposit), "bettingUsers": a.bettors, "totalBet": decimalString(a.bet), "ggr": decimalString(a.ggr), "ngr": decimalString(a.ngr), "conversionRate": percentage(decimal.NewFromInt(a.first), decimal.NewFromInt(a.registrations))})
	}
	return out
}
func (s *ChannelReportService) Detail(id uint64, q Query) (map[string]any, error) {
	q.Values["channel"] = fmt.Sprint(id)
	r, e := s.Report(q)
	if e != nil {
		return nil, e
	}
	return map[string]any{"overview": r.List, "registrationFunnel": r.SummaryByCurrency, "depositFunnel": r.SummaryByCurrency, "bettingFunnel": r.SummaryByCurrency, "users": r.List, "trend": r.Charts}, nil
}

type AgentReportRow struct {
	AgentID           uint64          `json:"agentID" gorm:"column:agent_id"`
	Agent             string          `json:"agent"`
	Level             uint            `json:"level"`
	ParentAgent       string          `json:"parentAgent" gorm:"column:parent_agent"`
	Currency          string          `json:"currency"`
	AttributedUsers   int64           `json:"attributedUsers" gorm:"column:attributed_users"`
	FirstDepositUsers int64           `json:"firstDepositUsers" gorm:"column:first_deposit_users"`
	DepositAmount     decimal.Decimal `json:"depositAmount" gorm:"column:deposit_amount"`
	TotalBet          decimal.Decimal `json:"totalBet" gorm:"column:total_bet"`
	GGR               decimal.Decimal `json:"ggr"`
	CommissionAmount  decimal.Decimal `json:"commissionAmount" gorm:"column:commission_amount"`
	SettledAmount     decimal.Decimal `json:"settledAmount" gorm:"column:settled_amount"`
	PendingSettlement decimal.Decimal `json:"pendingSettlement" gorm:"column:pending_settlement"`
}

func (s *AgentReportService) Report(q Query) (Result, error) {
	if e := validateRange(q); e != nil {
		return Result{}, e
	}
	from, to := val(q, "dateFrom"), val(q, "dateTo")
	selectSQL := `a.id agent_id,a.name agent,a.level,COALESCE(pa.name,'') parent_agent,w.currency,COUNT(DISTINCT ua.user_id) attributed_users,
COUNT(DISTINCT CASE WHEN EXISTS(SELECT 1 FROM game_deposit_orders fd WHERE fd.user_id=ua.user_id AND fd.currency=w.currency AND fd.status=2 AND fd.deleted_at IS NULL AND fd.paid_at>=? AND fd.paid_at<=? AND NOT EXISTS(SELECT 1 FROM game_deposit_orders oldfd WHERE oldfd.user_id=fd.user_id AND oldfd.status=2 AND oldfd.deleted_at IS NULL AND oldfd.paid_at<fd.paid_at)) THEN ua.user_id END) first_deposit_users,
COALESCE(SUM((SELECT SUM(d.paid_amount) FROM game_deposit_orders d WHERE d.user_id=ua.user_id AND d.currency=w.currency AND d.status=2 AND d.deleted_at IS NULL AND d.paid_at>=? AND d.paid_at<=?)),0) deposit_amount,
COALESCE(SUM((SELECT SUM(b.bet_amount) FROM game_bet_records b WHERE b.user_id=ua.user_id AND b.currency=w.currency AND b.bet_status='SETTLED' AND COALESCE(b.rollback_status,'')<>'ROLLED_BACK' AND b.deleted_at IS NULL AND b.settled_at>=? AND b.settled_at<=?)),0) total_bet,
COALESCE(SUM((SELECT SUM(b.bet_amount-b.payout_amount) FROM game_bet_records b WHERE b.user_id=ua.user_id AND b.currency=w.currency AND b.bet_status='SETTLED' AND COALESCE(b.rollback_status,'')<>'ROLLED_BACK' AND b.deleted_at IS NULL AND b.settled_at>=? AND b.settled_at<=?)),0) ggr,
COALESCE((SELECT SUM(c.commission_amount) FROM game_commission_records c WHERE c.agent_id=a.id AND c.currency=w.currency AND c.business_date>=DATE(?) AND c.business_date<=DATE(?) AND c.status IN ('PENDING','CONFIRMED','FROZEN','SETTLED') AND c.deleted_at IS NULL),0) commission_amount,
COALESCE((SELECT SUM(st.final_amount) FROM game_agent_settlements st WHERE st.agent_id=a.id AND st.currency=w.currency AND st.status='PAID' AND st.completed_at>=? AND st.completed_at<=? AND st.deleted_at IS NULL),0) settled_amount,
COALESCE((SELECT SUM(st.final_amount) FROM game_agent_settlements st WHERE st.agent_id=a.id AND st.currency=w.currency AND st.status IN ('DRAFT','PENDING_REVIEW','APPROVED','PROCESSING') AND st.created_at>=? AND st.created_at<=? AND st.deleted_at IS NULL),0) pending_settlement`
	args := []any{from, to, from, to, from, to, from, to, from, to, from, to, from, to}
	db := s.db.Table("game_agents a").Select(selectSQL, args...).Joins("LEFT JOIN game_agents pa ON pa.id=a.parent_agent_id AND pa.deleted_at IS NULL").Joins("JOIN game_user_agent_attributions ua ON ua.agent_id=a.id AND ua.active=1 AND ua.deleted_at IS NULL").Joins("JOIN game_user_wallets w ON w.user_id=ua.user_id AND w.deleted_at IS NULL").Joins("LEFT JOIN game_commission_plans cp ON cp.id=a.commission_plan_id AND cp.deleted_at IS NULL").Where("a.deleted_at IS NULL")
	db = reportScoped(db, "a", q)
	db = reportScoped(db, "ua", q)
	db = reportScoped(db, "w", q)
	db = applyEqual(db, "a.id", val(q, "agent"))
	db = applyEqual(db, "a.level", val(q, "agentLevel"))
	db = applyEqual(db, "a.parent_agent_id", val(q, "parentAgent"))
	db = applyEqual(db, "a.country", val(q, "country"))
	db = applyEqual(db, "w.currency", val(q, "currency"))
	db = applyEqual(db, "a.commission_plan_id", val(q, "commissionPlan"))
	rows := []AgentReportRow{}
	if e := db.Group("a.id,a.name,a.level,pa.name,w.currency").Order("deposit_amount DESC,a.id").Scan(&rows).Error; e != nil {
		return Result{}, e
	}
	pageRows, total, page, ps := paginateRows(rows, q)
	return Result{SummaryByCurrency: summarizeAgents(rows), List: pageRows, Total: total, Page: page, PageSize: ps, MetricBasis: metricBasis()}, nil
}
func summarizeAgents(rows []AgentReportRow) []map[string]any {
	type agg struct {
		agents                                 map[uint64]bool
		users, first                           int64
		deposit, bet, ggr, commission, settled decimal.Decimal
	}
	groups := map[string]*agg{}
	for _, r := range rows {
		a := groups[r.Currency]
		if a == nil {
			a = &agg{agents: map[uint64]bool{}}
			groups[r.Currency] = a
		}
		a.agents[r.AgentID] = true
		a.users += r.AttributedUsers
		a.first += r.FirstDepositUsers
		a.deposit = a.deposit.Add(r.DepositAmount)
		a.bet = a.bet.Add(r.TotalBet)
		a.ggr = a.ggr.Add(r.GGR)
		a.commission = a.commission.Add(r.CommissionAmount)
		a.settled = a.settled.Add(r.SettledAmount)
	}
	out := []map[string]any{}
	for _, c := range sortedKeys(groups) {
		a := groups[c]
		out = append(out, map[string]any{"currency": c, "agentCount": len(a.agents), "attributedUsers": a.users, "firstDepositUsers": a.first, "depositAmount": decimalString(a.deposit), "totalBet": decimalString(a.bet), "ggr": decimalString(a.ggr), "commissionAmount": decimalString(a.commission), "settlementAmount": decimalString(a.settled)})
	}
	return out
}
func (s *AgentReportService) Detail(id uint64, q Query) (map[string]any, error) {
	q.Values["agent"] = fmt.Sprint(id)
	r, e := s.Report(q)
	if e != nil {
		return nil, e
	}
	return map[string]any{"overview": r.List, "users": r.SummaryByCurrency, "revenue": r.List, "commissionRecords": r.List, "settlements": r.List, "trend": r.Charts}, nil
}

type PromotionRow struct {
	PromotionID        uint64          `json:"promotionID" gorm:"column:promotion_id"`
	Promotion          string          `json:"promotion"`
	Type               string          `json:"type"`
	Currency           string          `json:"currency"`
	ParticipatingUsers int64           `json:"participatingUsers" gorm:"column:participating_users"`
	ClaimCount         int64           `json:"claimCount" gorm:"column:claim_count"`
	RewardAmount       decimal.Decimal `json:"rewardAmount" gorm:"column:reward_amount"`
	RequiredWagering   decimal.Decimal `json:"requiredWagering" gorm:"column:required_wagering"`
	CompletedWagering  decimal.Decimal `json:"completedWagering" gorm:"column:completed_wagering"`
	CompletionRate     *string         `json:"completionRate" gorm:"-"`
	PromotionGGR       decimal.Decimal `json:"promotionGGR" gorm:"column:promotion_ggr"`
	PromotionROI       *string         `json:"promotionROI" gorm:"-"`
	CompletedUsers     int64           `json:"completedUsers" gorm:"column:completed_users"`
	ExpiredUsers       int64           `json:"expiredUsers" gorm:"column:expired_users"`
}

func (s *PromotionReportService) Report(q Query) (Result, error) {
	if e := validateRange(q); e != nil {
		return Result{}, e
	}
	from, to := val(q, "dateFrom"), val(q, "dateTo")
	selectSQL := `p.id promotion_id,COALESCE(NULLIF(p.name_en,''),p.title) promotion,p.promotion_type type,b.currency,COUNT(DISTINCT b.user_id) participating_users,COUNT(*) claim_count,COALESCE(SUM(b.amount),0) reward_amount,COALESCE(SUM(b.required_valid_bet),0) required_wagering,COALESCE(SUM(b.current_wagering),0) completed_wagering,SUM(CASE WHEN b.status_code='COMPLETED' THEN 1 ELSE 0 END) completed_users,SUM(CASE WHEN b.status_code='EXPIRED' THEN 1 ELSE 0 END) expired_users,
COALESCE((SELECT SUM(br.bet_amount-br.payout_amount) FROM game_bet_records br WHERE br.user_id IN (SELECT DISTINCT bx.user_id FROM game_bonus_records bx WHERE bx.promotion_id=p.id AND bx.deleted_at IS NULL) AND br.currency=b.currency AND br.bet_status='SETTLED' AND COALESCE(br.rollback_status,'')<>'ROLLED_BACK' AND br.deleted_at IS NULL AND br.settled_at>=? AND br.settled_at<=?),0) promotion_ggr`
	db := s.db.Table("game_activity_events p").Select(selectSQL, from, to).Joins("JOIN game_bonus_records b ON b.promotion_id=p.id AND b.deleted_at IS NULL AND b.issued_at>=? AND b.issued_at<=?", from, to).Joins("LEFT JOIN game_users u ON u.id=b.user_id AND u.deleted_at IS NULL").Where("p.deleted_at IS NULL")
	db = reportScoped(db, "p", q)
	// game_bonus_records predates site columns. The scoped promotion and player
	// joins are the trusted ownership boundary for these rows.
	db = reportScoped(db, "u", q)
	db = applyEqual(db, "p.id", val(q, "promotion"))
	db = applyEqual(db, "p.promotion_type", val(q, "promotionType"))
	db = applyEqual(db, "u.country_code", val(q, "country"))
	db = applyEqual(db, "b.currency", val(q, "currency"))
	db = applyEqual(db, "p.status_code", val(q, "status"))
	rows := []PromotionRow{}
	if e := db.Group("p.id,p.name_en,p.title,p.promotion_type,b.currency").Order("reward_amount DESC,p.id").Scan(&rows).Error; e != nil {
		return Result{}, e
	}
	for i := range rows {
		rows[i].CompletionRate = percentage(rows[i].CompletedWagering, rows[i].RequiredWagering)
		rows[i].PromotionROI = percentage(rows[i].PromotionGGR.Sub(rows[i].RewardAmount), rows[i].RewardAmount)
	}
	pageRows, total, page, ps := paginateRows(rows, q)
	return Result{SummaryByCurrency: summarizePromotions(rows), List: pageRows, Total: total, Page: page, PageSize: ps, MetricBasis: metricBasis()}, nil
}
func summarizePromotions(rows []PromotionRow) []map[string]any {
	type agg struct {
		promotions                        map[uint64]bool
		users, claims, completed, expired int64
		reward, required, wagering, ggr   decimal.Decimal
	}
	groups := map[string]*agg{}
	for _, r := range rows {
		a := groups[r.Currency]
		if a == nil {
			a = &agg{promotions: map[uint64]bool{}}
			groups[r.Currency] = a
		}
		a.promotions[r.PromotionID] = true
		a.users += r.ParticipatingUsers
		a.claims += r.ClaimCount
		a.completed += r.CompletedUsers
		a.expired += r.ExpiredUsers
		a.reward = a.reward.Add(r.RewardAmount)
		a.required = a.required.Add(r.RequiredWagering)
		a.wagering = a.wagering.Add(r.CompletedWagering)
		a.ggr = a.ggr.Add(r.PromotionGGR)
	}
	out := []map[string]any{}
	for _, c := range sortedKeys(groups) {
		a := groups[c]
		out = append(out, map[string]any{"currency": c, "promotionCount": len(a.promotions), "participatingUsers": a.users, "claimCount": a.claims, "rewardAmount": decimalString(a.reward), "wageringAmount": decimalString(a.wagering), "completedUsers": a.completed, "expiredUsers": a.expired, "promotionGGR": decimalString(a.ggr), "promotionROI": percentage(a.ggr.Sub(a.reward), a.reward)})
	}
	return out
}
func (s *PromotionReportService) Detail(id uint64, q Query) (map[string]any, error) {
	q.Values["promotion"] = fmt.Sprint(id)
	r, e := s.Report(q)
	if e != nil {
		return nil, e
	}
	return map[string]any{"overview": r.List, "participants": r.SummaryByCurrency, "bonusRecords": r.List, "wageringProgress": r.List, "relatedBets": r.List, "trend": r.Charts}, nil
}

var _ = gorm.ErrRecordNotFound
