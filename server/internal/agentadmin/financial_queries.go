package agentadmin

import (
	"fmt"

	"casino/internal/content"
)

func (s *Service) ListCommissionRecords(q ListQuery) (PageResult, error) {
	db := s.db.Table("game_commission_records r").Where("r.deleted_at IS NULL").Select("r.*,a.agent_code,a.name agent_name,u.user_no,u.username").Joins("JOIN game_agents a ON a.id=r.agent_id").Joins("LEFT JOIN game_users u ON u.id=r.user_id")
	db = scoped(db, "r", q.Scope)
	db = scoped(db, "a", q.Scope)
	if id := parseID(val(q, "recordID")); id > 0 {
		db = db.Where("r.id=?", id)
	}
	db = like(db, "r.record_no", val(q, "recordNo"))
	if id := parseID(val(q, "agentID")); id > 0 {
		db = db.Where("r.agent_id=?", id)
	}
	db = like(db, "a.agent_code", val(q, "agentCode"))
	db = like(db, "a.name", val(q, "agentName"))
	if id := parseID(val(q, "userID")); id > 0 {
		db = db.Where("r.user_id=?", id)
	}
	db = like(db, "u.user_no", val(q, "playerNo"))
	db = eq(db, "r.commission_type", val(q, "commissionType"))
	db = eq(db, "r.business_date", val(q, "businessDate"))
	db = eq(db, "r.currency", val(q, "currency"))
	db = eq(db, "r.status", val(q, "status"))
	db = rangeTime(db, "r.created_at", val(q, "createdFrom"), val(q, "createdTo"))
	var rows []map[string]any
	return paginate(db.Order("r.created_at DESC,r.id DESC"), q, &rows)
}
func (s *Service) CommissionDetail(id uint64, scope content.Scope) (map[string]any, error) {
	var row map[string]any
	db := s.db.Table("game_commission_records r").Select("r.*,a.agent_code,a.name agent_name,u.user_no,u.username,cp.code plan_code,cp.name_en plan_name").Joins("JOIN game_agents a ON a.id=r.agent_id").Joins("LEFT JOIN game_users u ON u.id=r.user_id").Joins("LEFT JOIN game_commission_plans cp ON cp.id=r.commission_plan_id").Where("r.id=? AND r.deleted_at IS NULL", id)
	db = scoped(db, "r", scope)
	db = scoped(db, "a", scope)
	if e := db.Take(&row).Error; e != nil {
		return nil, notFound(e, "COMMISSION_RECORD_NOT_FOUND")
	}
	return map[string]any{"record": row, "adjustments": []any{}, "operationLogs": s.logsFor("commission_record", id, scope)}, nil
}

func (s *Service) ListSettlements(q ListQuery) (PageResult, error) {
	db := s.db.Table("game_agent_settlements st").Where("st.deleted_at IS NULL").Select("st.*,a.agent_code,a.name agent_name").Joins("JOIN game_agents a ON a.id=st.agent_id")
	db = scoped(db, "st", q.Scope)
	db = scoped(db, "a", q.Scope)
	if id := parseID(val(q, "settlementID")); id > 0 {
		db = db.Where("st.id=?", id)
	}
	db = like(db, "st.settlement_no", val(q, "settlementNo"))
	if id := parseID(val(q, "agentID")); id > 0 {
		db = db.Where("st.agent_id=?", id)
	}
	db = like(db, "a.agent_code", val(q, "agentCode"))
	db = like(db, "a.name", val(q, "agent"))
	db = eq(db, "st.currency", val(q, "currency"))
	db = eq(db, "st.status", val(q, "status"))
	db = eq(db, "CONCAT(st.period_start,'/',st.period_end)", val(q, "settlementCycle"))
	if min := val(q, "amountMin"); min != "" {
		db = db.Where("st.final_amount>=?", min)
	}
	if max := val(q, "amountMax"); max != "" {
		db = db.Where("st.final_amount<=?", max)
	}
	db = rangeTime(db, "st.created_at", val(q, "createdFrom"), val(q, "createdTo"))
	db = rangeTime(db, "st.completed_at", val(q, "completedFrom"), val(q, "completedTo"))
	var rows []map[string]any
	return paginate(db.Order("st.created_at DESC,st.id DESC"), q, &rows)
}
func (s *Service) SettlementDetail(id uint64, scope content.Scope) (map[string]any, error) {
	var row map[string]any
	db := s.db.Table("game_agent_settlements st").Select("st.*,a.agent_code,a.name agent_name").Joins("JOIN game_agents a ON a.id=st.agent_id").Where("st.id=? AND st.deleted_at IS NULL", id)
	db = scoped(db, "st", scope)
	db = scoped(db, "a", scope)
	if e := db.Take(&row).Error; e != nil {
		return nil, notFound(e, "SETTLEMENT_NOT_FOUND")
	}
	var records []map[string]any
	itemDB := s.db.Table("game_agent_settlement_items i").Select("r.*,i.amount settlement_amount").Joins("JOIN game_commission_records r ON r.id=i.commission_record_id").Where("i.settlement_id=? AND i.deleted_at IS NULL", id)
	itemDB = scoped(itemDB, "i", scope)
	itemDB = scoped(itemDB, "r", scope)
	_ = itemDB.Find(&records).Error
	return map[string]any{"record": row, "commissionRecords": records, "approvalHistory": s.logsFor("agent_settlement", id, scope), "operationLogs": s.logsFor("agent_settlement", id, scope)}, nil
}

func (s *Service) ListWallets(q ListQuery) (PageResult, error) {
	db := s.db.Table("game_agent_wallets w").Where("w.deleted_at IS NULL").Select("w.*,a.agent_code,a.name agent_name").Joins("JOIN game_agents a ON a.id=w.agent_id")
	db = scoped(db, "w", q.Scope)
	db = scoped(db, "a", q.Scope)
	if id := parseID(val(q, "walletID")); id > 0 {
		db = db.Where("w.id=?", id)
	}
	db = like(db, "w.wallet_no", val(q, "walletNo"))
	if id := parseID(val(q, "agentID")); id > 0 {
		db = db.Where("w.agent_id=?", id)
	}
	db = like(db, "a.agent_code", val(q, "agentCode"))
	db = like(db, "a.name", val(q, "agentName"))
	db = eq(db, "w.currency", val(q, "currency"))
	db = eq(db, "w.status", val(q, "walletStatus"))
	if min := val(q, "balanceMin"); min != "" {
		db = db.Where("w.available_balance>=?", min)
	}
	if max := val(q, "balanceMax"); max != "" {
		db = db.Where("w.available_balance<=?", max)
	}
	db = rangeTime(db, "w.updated_at", val(q, "updatedFrom"), val(q, "updatedTo"))
	var rows []map[string]any
	return paginate(db.Order("w.updated_at DESC,w.id DESC"), q, &rows)
}
func (s *Service) WalletDetail(id uint64, scope content.Scope) (map[string]any, error) {
	var row map[string]any
	db := s.db.Table("game_agent_wallets w").Select("w.*,a.agent_code,a.name agent_name").Joins("JOIN game_agents a ON a.id=w.agent_id").Where("w.id=? AND w.deleted_at IS NULL", id)
	db = scoped(db, "w", scope)
	db = scoped(db, "a", scope)
	if e := db.Take(&row).Error; e != nil {
		return nil, notFound(e, "AGENT_WALLET_NOT_FOUND")
	}
	txs, _ := s.ListWalletTransactions(ListQuery{Page: 1, PageSize: 100, Values: map[string]string{"walletID": fmt.Sprint(id)}, Scope: scope})
	return map[string]any{"record": row, "balanceComposition": map[string]any{"availableBalance": row["available_balance"], "frozenBalance": row["frozen_balance"], "settledAmount": row["settled_amount"], "withdrawnAmount": row["withdrawn_amount"]}, "transactions": txs.List, "operationLogs": s.logsFor("agent_wallet", id, scope)}, nil
}
func (s *Service) ListWalletTransactions(q ListQuery) (PageResult, error) {
	db := s.db.Table("game_agent_wallet_transactions t").Where("t.deleted_at IS NULL").Select("t.*,a.agent_code,a.name agent_name,w.wallet_no").Joins("JOIN game_agents a ON a.id=t.agent_id").Joins("JOIN game_agent_wallets w ON w.id=t.wallet_id")
	db = scoped(db, "t", q.Scope)
	db = scoped(db, "a", q.Scope)
	db = scoped(db, "w", q.Scope)
	if id := parseID(val(q, "walletID")); id > 0 {
		db = db.Where("t.wallet_id=?", id)
	}
	if id := parseID(val(q, "agentID")); id > 0 {
		db = db.Where("t.agent_id=?", id)
	}
	db = eq(db, "t.transaction_type", val(q, "transactionType"))
	db = rangeTime(db, "t.created_at", val(q, "createdFrom"), val(q, "createdTo"))
	var rows []map[string]any
	return paginate(db.Order("t.created_at DESC,t.id DESC"), q, &rows)
}

func (s *Service) ListLogs(q ListQuery) (PageResult, error) {
	db := s.db.Table("game_admin_operation_logs l").Where("l.deleted_at IS NULL AND (l.target_type LIKE 'agent%' OR l.target_type IN ('commission_plan','commission_record','agent_channel'))")
	db = scoped(db, "l", q.Scope)
	db = like(db, "l.operator_username", val(q, "operator"))
	db = like(db, "l.action", val(q, "action"))
	if id := parseID(val(q, "agentID")); id > 0 {
		db = db.Where("(l.target_type='agent' AND l.target_id=?) OR l.after_state LIKE ?", id, "%\"agentId\":"+fmt.Sprint(id)+"%")
	}
	db = eq(db, "l.target_type", val(q, "targetType"))
	if id := parseID(val(q, "targetID")); id > 0 {
		db = db.Where("l.target_id=?", id)
	}
	db = rangeTime(db, "l.created_at", val(q, "createdFrom"), val(q, "createdTo"))
	var rows []map[string]any
	return paginate(db.Order("l.created_at DESC,l.id DESC"), q, &rows)
}
func (s *Service) LogDetail(id uint64, scope content.Scope) (map[string]any, error) {
	var row map[string]any
	e := scoped(s.db.Table("game_admin_operation_logs"), "", scope).Where("id=? AND deleted_at IS NULL", id).Take(&row).Error
	if e != nil {
		return nil, notFound(e, "OPERATION_LOG_NOT_FOUND")
	}
	return map[string]any{"record": row}, nil
}
