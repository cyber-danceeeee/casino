package agentadmin

import (
	"fmt"

	"casino/internal/content"
)

func (s *Service) ListAgents(q ListQuery) (PageResult, error) {
	db := s.db.Table("game_agents a").Where("a.deleted_at IS NULL").
		Select(`a.*,COALESCE(u.user_no,'') user_no,COALESCE(u.username,'') username,COALESCE(p.name,'') parent_agent,COALESCE(cp.name_en,'') commission_plan,
		(SELECT COUNT(*) FROM game_agents c WHERE c.parent_agent_id=a.id AND c.deleted_at IS NULL) direct_agents,
		(SELECT COUNT(*) FROM game_user_agent_attributions ua WHERE ua.agent_id=a.id AND ua.active=1 AND ua.deleted_at IS NULL) direct_users,
		COALESCE(w.available_balance,0) available_commission,COALESCE(w.frozen_balance,0) frozen_commission`).
		Joins("LEFT JOIN game_agents p ON p.id=a.parent_agent_id AND p.deleted_at IS NULL").
		Joins("LEFT JOIN game_users u ON u.id=a.user_id AND u.deleted_at IS NULL").
		Joins("LEFT JOIN game_commission_plans cp ON cp.id=a.commission_plan_id AND cp.deleted_at IS NULL").
		Joins("LEFT JOIN game_agent_wallets w ON w.agent_id=a.id AND w.currency=a.currency AND w.deleted_at IS NULL")
	db = scoped(db, "a", q.Scope)
	db = scoped(db, "u", q.Scope)
	if id := parseID(val(q, "agentID")); id > 0 {
		db = db.Where("a.id=?", id)
	}
	db = like(db, "a.agent_code", val(q, "agentCode"))
	if id := parseID(val(q, "parentAgentID")); id > 0 {
		db = db.Where("a.parent_agent_id=?", id)
	}
	db = like(db, "a.name", val(q, "agentName"))
	db = eq(db, "a.level", val(q, "agentLevel"))
	db = like(db, "p.name", val(q, "parentAgent"))
	db = eq(db, "a.status", val(q, "status"))
	db = eq(db, "a.country", val(q, "country"))
	db = eq(db, "a.currency", val(q, "currency"))
	db = like(db, "cp.code", val(q, "commissionPlan"))
	db = rangeTime(db, "a.created_at", val(q, "createdFrom"), val(q, "createdTo"))
	if channel := val(q, "registrationChannel"); channel != "" {
		db = db.Where("EXISTS (SELECT 1 FROM game_user_agent_attributions ua JOIN game_agent_channels ch ON ch.id=ua.channel_id WHERE ua.agent_id=a.id AND ch.code LIKE ? AND ua.active=1)", "%"+channel+"%")
	}
	var rows []agentRow
	result, err := paginate(db.Order("a.created_at DESC,a.id DESC"), q, &rows)
	if err != nil {
		return result, err
	}
	// One batch builds descendant totals without per-row queries.
	type node struct {
		ID, Parent uint64
		Users      int64
	}
	var nodes []node
	nodeDB := s.db.Table("game_agents a").
		Select("a.id,a.parent_agent_id parent,COUNT(ua.id) users").
		Joins("LEFT JOIN game_user_agent_attributions ua ON ua.agent_id=a.id AND ua.active=1 AND ua.deleted_at IS NULL").
		Where("a.deleted_at IS NULL").
		Group("a.id,a.parent_agent_id")
	nodeDB = scoped(nodeDB, "a", q.Scope)
	_ = nodeDB.Scan(&nodes).Error
	children := map[uint64][]uint64{}
	users := map[uint64]int64{}
	for _, n := range nodes {
		children[n.Parent] = append(children[n.Parent], n.ID)
		users[n.ID] = n.Users
	}
	var sum func(uint64, map[uint64]bool) int64
	sum = func(id uint64, seen map[uint64]bool) int64 {
		if seen[id] {
			return 0
		}
		seen[id] = true
		total := users[id]
		for _, c := range children[id] {
			total += sum(c, seen)
		}
		return total
	}
	for i := range rows {
		rows[i].TotalUsers = sum(rows[i].ID, map[uint64]bool{})
	}
	result.List = rows
	return result, nil
}

func (s *Service) AgentDetail(id uint64, scope content.Scope) (map[string]any, error) {
	q := ListQuery{Page: 1, PageSize: 1, Values: map[string]string{"agentID": fmt.Sprint(id)}, Scope: scope}
	list, e := s.ListAgents(q)
	if e != nil {
		return nil, e
	}
	rows := list.List.([]agentRow)
	if len(rows) == 0 {
		return nil, NewError("AGENT_NOT_FOUND", "agent not found")
	}
	children, _ := s.ListAgents(ListQuery{Page: 1, PageSize: 100, Values: map[string]string{"parentAgentID": fmt.Sprint(id)}, Scope: scope})
	users, _ := s.AgentUsers(id, ListQuery{Page: 1, PageSize: 20, Values: map[string]string{}, Scope: scope})
	commissions, _ := s.ListCommissionRecords(ListQuery{Page: 1, PageSize: 20, Values: map[string]string{"agentID": fmt.Sprint(id)}, Scope: scope})
	settlements, _ := s.ListSettlements(ListQuery{Page: 1, PageSize: 20, Values: map[string]string{"agentID": fmt.Sprint(id)}, Scope: scope})
	wallets, _ := s.ListWallets(ListQuery{Page: 1, PageSize: 20, Values: map[string]string{"agentID": fmt.Sprint(id)}, Scope: scope})
	logs, _ := s.ListLogs(ListQuery{Page: 1, PageSize: 50, Values: map[string]string{"agentID": fmt.Sprint(id)}, Scope: scope})
	var plan any
	if rows[0].CommissionPlanID > 0 {
		if detail, err := s.PlanDetail(rows[0].CommissionPlanID, scope); err == nil {
			plan = detail["record"]
		}
	}
	return map[string]any{"record": rows[0], "hierarchy": map[string]any{"parentAgentId": rows[0].ParentAgentID, "parentAgent": rows[0].ParentAgent, "level": rows[0].Level}, "children": children.List, "users": users.List, "commissionPlan": plan, "commissionRecords": commissions.List, "settlements": settlements.List, "wallets": wallets.List, "operationLogs": logs.List}, nil
}

func (s *Service) AgentUsers(id uint64, q ListQuery) (PageResult, error) {
	db := s.db.Table("game_user_agent_attributions ua").
		Select("u.id,u.user_no userNo,u.username,u.phone,u.country_code country,u.created_at createdAt,ua.attributed_at attributedAt,ua.attribution_source attributionSource,ch.code channelCode").
		Joins("JOIN game_agents a ON a.id=ua.agent_id AND a.deleted_at IS NULL").
		Joins("JOIN game_users u ON u.id=ua.user_id AND u.deleted_at IS NULL").
		Joins("LEFT JOIN game_agent_channels ch ON ch.id=ua.channel_id").
		Where("ua.agent_id=? AND ua.active=1 AND ua.deleted_at IS NULL", id)
	db = scoped(db, "ua", q.Scope)
	db = scoped(db, "a", q.Scope)
	db = scoped(db, "u", q.Scope)
	db = like(db, "u.username", val(q, "username"))
	db = like(db, "ch.code", val(q, "channel"))
	return paginate(db.Order("ua.attributed_at DESC"), q, &[]map[string]any{})
}

func (s *Service) Hierarchy(q ListQuery) (PageResult, error) {
	result, e := s.ListAgents(q)
	if e != nil {
		return result, e
	}
	return result, nil
}

func (s *Service) ListChannels(q ListQuery) (PageResult, error) {
	db := s.db.Table("game_agent_channels c").Where("c.deleted_at IS NULL").Select(`c.*,a.name agent_name,
	(SELECT COUNT(*) FROM game_user_agent_attributions ua WHERE ua.channel_id=c.id AND ua.active=1 AND ua.deleted_at IS NULL) registered_users,
		(SELECT COUNT(DISTINCT d.user_id) FROM game_deposit_orders d JOIN game_user_agent_attributions ua ON ua.user_id=d.user_id AND ua.channel_id=c.id WHERE d.status=2 AND d.deleted_at IS NULL) first_deposit_users,
		COALESCE((SELECT SUM(d.paid_amount) FROM game_deposit_orders d JOIN game_user_agent_attributions ua ON ua.user_id=d.user_id AND ua.channel_id=c.id WHERE d.status=2 AND d.deleted_at IS NULL),0) deposit_amount`).Joins("JOIN game_agents a ON a.id=c.agent_id AND a.deleted_at IS NULL")
	db = scoped(db, "c", q.Scope)
	db = scoped(db, "a", q.Scope)
	db = like(db, "c.code", val(q, "channelCode"))
	db = like(db, "CONCAT(c.name_zh,c.name_en)", val(q, "channelName"))
	db = like(db, "a.name", val(q, "agent"))
	db = eq(db, "c.country", val(q, "country"))
	db = eq(db, "c.platform", val(q, "platform"))
	if v := val(q, "status"); v != "" {
		db = db.Where("c.enabled=?", v == "ENABLED" || v == "true")
	}
	db = rangeTime(db, "c.created_at", val(q, "createdFrom"), val(q, "createdTo"))
	var rows []channelRow
	return paginate(db.Order("c.created_at DESC"), q, &rows)
}
func (s *Service) ChannelDetail(id uint64, scope content.Scope) (map[string]any, error) {
	var r channelRow
	db := s.db.Table("game_agent_channels c").Select(`c.*,a.name agent_name,
	(SELECT COUNT(*) FROM game_user_agent_attributions ua WHERE ua.channel_id=c.id AND ua.active=1 AND ua.deleted_at IS NULL) registered_users,
	(SELECT COUNT(DISTINCT d.user_id) FROM game_deposit_orders d JOIN game_user_agent_attributions ua ON ua.user_id=d.user_id AND ua.channel_id=c.id WHERE d.status=2 AND d.deleted_at IS NULL) first_deposit_users,
	COALESCE((SELECT SUM(d.paid_amount) FROM game_deposit_orders d JOIN game_user_agent_attributions ua ON ua.user_id=d.user_id AND ua.channel_id=c.id WHERE d.status=2 AND d.deleted_at IS NULL),0) deposit_amount`).Joins("JOIN game_agents a ON a.id=c.agent_id").Where("c.id=? AND c.deleted_at IS NULL", id)
	db = scoped(db, "c", scope)
	db = scoped(db, "a", scope)
	if e := db.Take(&r).Error; e != nil {
		return nil, notFound(e, "CHANNEL_NOT_FOUND")
	}
	return map[string]any{"record": r, "statistics": map[string]any{"registeredUsers": r.RegisteredUsers, "firstDepositUsers": r.FirstDepositUsers, "depositAmount": r.DepositAmount}, "trackingParameters": jsonValue(r.TrackingParameters), "operationLogs": s.logsFor("agent_channel", id, scope)}, nil
}

func (s *Service) ListPlans(q ListQuery) (PageResult, error) {
	db := s.db.Table("game_commission_plans p").Where("p.deleted_at IS NULL").Select("p.*,(SELECT COUNT(*) FROM game_agents a WHERE a.commission_plan_id=p.id AND a.deleted_at IS NULL) assigned_agent_count")
	db = scoped(db, "p", q.Scope)
	db = like(db, "p.code", val(q, "planCode"))
	db = like(db, "CONCAT(p.name_zh,p.name_en)", val(q, "planName"))
	db = eq(db, "p.model", val(q, "commissionModel"))
	db = eq(db, "p.currency", val(q, "currency"))
	if v := val(q, "status"); v != "" {
		db = db.Where("p.enabled=?", v == "ENABLED" || v == "true")
	}
	db = rangeTime(db, "p.created_at", val(q, "createdFrom"), val(q, "createdTo"))
	var rows []planRow
	return paginate(db.Order("p.created_at DESC"), q, &rows)
}
func (s *Service) PlanDetail(id uint64, scope content.Scope) (map[string]any, error) {
	var r planRow
	db := s.db.Table("game_commission_plans p").Select("p.*,(SELECT COUNT(*) FROM game_agents a WHERE a.commission_plan_id=p.id AND a.deleted_at IS NULL) assigned_agent_count").Where("p.id=? AND p.deleted_at IS NULL", id)
	db = scoped(db, "p", scope)
	if e := db.Take(&r).Error; e != nil {
		return nil, notFound(e, "PLAN_NOT_FOUND")
	}
	var agents []agentRow
	agentDB := scoped(s.db.Table("game_agents"), "", scope).Where("commission_plan_id=? AND deleted_at IS NULL", id)
	_ = agentDB.Scan(&agents).Error
	return map[string]any{"record": r, "cpaRules": map[string]any{"cpaAmount": r.CPAAmount}, "revenueShareRules": map[string]any{"revenueShareRate": r.RevenueShareRate, "negativeCarryover": r.NegativeCarryover, "deductionRules": jsonValue(r.DeductionRules)}, "deductionRules": jsonValue(r.DeductionRules), "tierRules": jsonValue(r.TierRules), "eligibleProducts": jsonValue(r.EligibleProducts), "assignedAgents": agents, "operationLogs": s.logsFor("commission_plan", id, scope)}, nil
}

func (s *Service) logsFor(target string, id uint64, scope content.Scope) []map[string]any {
	rows := make([]map[string]any, 0)
	db := scoped(s.db.Table("game_admin_operation_logs"), "", scope).Where("target_type=? AND target_id=? AND deleted_at IS NULL", target, id)
	_ = db.Order("created_at DESC").Limit(100).Find(&rows).Error
	return rows
}
