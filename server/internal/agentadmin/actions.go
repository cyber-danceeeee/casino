package agentadmin

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func duplicate(err error) bool {
	var me *mysql.MySQLError
	return errors.As(err, &me) && me.Number == 1062
}

func (s *Service) SaveAgent(id uint64, in AgentInput, actor Actor) (uint64, error) {
	if e := requireReason(in.Reason); e != nil {
		return 0, e
	}
	in.AgentCode = strings.ToUpper(strings.TrimSpace(in.AgentCode))
	if in.AgentCode == "" || in.Name == "" {
		return 0, NewError("INVALID_REQUEST", "agent code and name are required")
	}
	if id > 0 {
		var current struct {
			AgentCode        string
			ParentAgentID    uint64
			CommissionPlanID uint64
			ActivatedAt      *time.Time
		}
		if e := actorScoped(s.db.Table("game_agents"), "", actor).Where("id=? AND deleted_at IS NULL", id).Take(&current).Error; e != nil {
			return 0, notFound(e, "AGENT_NOT_FOUND")
		}
		if current.ActivatedAt != nil && current.AgentCode != in.AgentCode {
			return 0, NewError("AGENT_CODE_IMMUTABLE", "activated agent code is immutable")
		}
		if current.ParentAgentID != in.ParentAgentID {
			return 0, NewError("USE_CHANGE_PARENT", "use the change-parent action to update agent hierarchy")
		}
		if current.CommissionPlanID != in.CommissionPlanID {
			return 0, NewError("USE_ASSIGN_PLAN", "use the assign-plan action to update commission plan")
		}
	}
	if in.ParentAgentID == id && id > 0 {
		return 0, NewError("INVALID_PARENT", "agent cannot be its own parent")
	}
	level := uint(1)
	if in.ParentAgentID > 0 {
		var parent struct {
			Level  uint
			Status string
		}
		if e := actorScoped(s.db.Table("game_agents"), "", actor).Where("id=? AND deleted_at IS NULL", in.ParentAgentID).Take(&parent).Error; e != nil {
			return 0, NewError("PARENT_AGENT_NOT_FOUND", "parent agent not found")
		}
		level = parent.Level + 1
		if level > 8 {
			return 0, NewError("MAX_HIERARCHY_DEPTH", "maximum hierarchy depth exceeded")
		}
	}
	if in.CommissionPlanID > 0 {
		var planCount int64
		if e := actorScoped(s.db.Table("game_commission_plans"), "", actor).Where("id=? AND enabled=1 AND deleted_at IS NULL", in.CommissionPlanID).Count(&planCount).Error; e != nil || planCount == 0 {
			return 0, NewError("COMMISSION_PLAN_NOT_FOUND", "enabled commission plan not found")
		}
	}
	row := map[string]any{"site_id": actor.Scope.SiteID, "agent_code": in.AgentCode, "name": strings.TrimSpace(in.Name), "user_id": in.UserID, "parent_agent_id": in.ParentAgentID, "level": level, "country": strings.ToUpper(in.Country), "currency": strings.ToUpper(in.Currency), "commission_plan_id": in.CommissionPlanID, "updated_at": time.Now()}
	e := s.db.Transaction(func(tx *gorm.DB) error {
		var before map[string]any
		if id == 0 {
			row["status"] = "ACTIVE"
			row["activated_at"] = time.Now()
			row["created_at"] = time.Now()
			if e := tx.Table("game_agents").Create(&row).Error; e != nil {
				return e
			}
			insertedID, e := lastInsertID(tx)
			if e != nil {
				return e
			}
			id = insertedID
			wallet := map[string]any{"site_id": actor.Scope.SiteID, "wallet_no": nowNo("AW"), "agent_id": id, "currency": strings.ToUpper(in.Currency), "status": "ACTIVE", "created_at": time.Now(), "updated_at": time.Now()}
			if e := tx.Table("game_agent_wallets").Create(&wallet).Error; e != nil {
				return e
			}
		} else {
			if e := actorScoped(tx.Table("game_agents"), "", actor).Where("id=?", id).Take(&before).Error; e != nil {
				return e
			}
			delete(row, "agent_code")
			delete(row, "")
			delete(row, "site_id")
			if e := actorScoped(tx.Table("game_agents"), "", actor).Where("id=?", id).Updates(row).Error; e != nil {
				return e
			}
		}
		return audit(tx, actor, permission(id == 0, "agent:create", "agent:update"), permission(id == 0, "create", "update"), "agent", id, before, row, in.Reason, nil, in.Currency)
	})
	if duplicate(e) {
		return 0, NewError("DUPLICATE_AGENT_CODE", "agent code already exists")
	}
	return id, e
}
func permission(ok bool, a string, b string) string {
	if ok {
		return a
	}
	return b
}

func (s *Service) SetAgentStatus(id uint64, in StatusCommand, actor Actor) error {
	if e := requireReason(in.Reason); e != nil {
		return e
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row agentRow
		if e := actorScoped(tx.Table("game_agents").Clauses(clause.Locking{Strength: "UPDATE"}), "", actor).Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; e != nil {
			return notFound(e, "AGENT_NOT_FOUND")
		}
		next := "DISABLED"
		if in.Enabled {
			next = "ACTIVE"
		}
		if row.Status == next {
			return NewError("INVALID_AGENT_STATUS_TRANSITION", "agent already has requested status")
		}
		if e := actorScoped(tx.Table("game_agents"), "", actor).Where("id=?", id).Updates(map[string]any{"status": next, "updated_at": time.Now()}).Error; e != nil {
			return e
		}
		return audit(tx, actor, "agent:status", "status.update", "agent", id, map[string]any{"status": row.Status}, map[string]any{"status": next}, in.Reason, nil, row.Currency)
	})
}

func (s *Service) ChangeParent(id uint64, in ChangeParentInput, actor Actor) error {
	if e := requireReason(in.Reason); e != nil {
		return e
	}
	if id == in.ParentAgentID {
		return NewError("INVALID_PARENT", "agent cannot be its own parent")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row agentRow
		if e := actorScoped(tx.Table("game_agents").Clauses(clause.Locking{Strength: "UPDATE"}), "", actor).Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; e != nil {
			return notFound(e, "AGENT_NOT_FOUND")
		}
		if row.ParentAgentID == in.ParentAgentID {
			return NewError("HIERARCHY_UNCHANGED", "parent agent is unchanged")
		}
		newLevel := uint(1)
		cursor := in.ParentAgentID
		seen := map[uint64]bool{id: true}
		for depth := 0; cursor > 0; depth++ {
			if seen[cursor] {
				return NewError("CIRCULAR_HIERARCHY", "circular agent hierarchy")
			}
			seen[cursor] = true
			var p struct {
				ID, ParentAgentID uint64
				Level             uint
			}
			if e := actorScoped(tx.Table("game_agents"), "", actor).Where("id=? AND deleted_at IS NULL", cursor).Take(&p).Error; e != nil {
				return NewError("PARENT_AGENT_NOT_FOUND", "parent agent not found")
			}
			if depth == 0 {
				newLevel = p.Level + 1
			}
			if newLevel > 8 {
				return NewError("MAX_HIERARCHY_DEPTH", "maximum hierarchy depth exceeded")
			}
			cursor = p.ParentAgentID
		}
		var descendants []struct {
			ID    uint64
			Level uint
		}
		rootScope, rootArgs := sqlScope(actor.Scope, "")
		childScope, childArgs := sqlScope(actor.Scope, "child")
		descendantSQL := `WITH RECURSIVE descendants AS (
			SELECT id, level FROM game_agents WHERE parent_agent_id = ?` + rootScope + ` AND deleted_at IS NULL
			UNION ALL
			SELECT child.id, child.level FROM game_agents child
			JOIN descendants parent ON child.parent_agent_id = parent.id
			WHERE child.deleted_at IS NULL` + childScope + `
		) SELECT id, level FROM descendants`
		descendantArgs := append([]any{id}, rootArgs...)
		descendantArgs = append(descendantArgs, childArgs...)
		if e := tx.Raw(descendantSQL, descendantArgs...).Scan(&descendants).Error; e != nil {
			return e
		}
		levelDelta := int(newLevel) - int(row.Level)
		for _, descendant := range descendants {
			nextLevel := int(descendant.Level) + levelDelta
			if nextLevel < 1 || nextLevel > 8 {
				return NewError("MAX_HIERARCHY_DEPTH", "maximum hierarchy depth exceeded")
			}
		}
		if e := actorScoped(tx.Table("game_agents"), "", actor).Where("id=?", id).Updates(map[string]any{"parent_agent_id": in.ParentAgentID, "level": newLevel}).Error; e != nil {
			return e
		}
		for _, descendant := range descendants {
			if e := actorScoped(tx.Table("game_agents"), "", actor).Where("id=?", descendant.ID).Update("level", int(descendant.Level)+levelDelta).Error; e != nil {
				return e
			}
		}
		history := map[string]any{"site_id": actor.Scope.SiteID, "agent_id": id, "before_parent_id": row.ParentAgentID, "after_parent_id": in.ParentAgentID, "before_level": row.Level, "after_level": newLevel, "operator_id": actor.ID, "operator_username": actor.Username, "reason": in.Reason, "created_at": time.Now(), "updated_at": time.Now()}
		if e := tx.Table("game_agent_hierarchy_history").Create(&history).Error; e != nil {
			return e
		}
		return audit(tx, actor, "agent:hierarchy:update", "hierarchy.change-parent", "agent", id, map[string]any{"parentAgentId": row.ParentAgentID, "level": row.Level}, map[string]any{"parentAgentId": in.ParentAgentID, "level": newLevel}, in.Reason, nil, row.Currency)
	})
}

func (s *Service) AssignPlan(id uint64, in AssignPlanInput, actor Actor) error {
	if e := requireReason(in.Reason); e != nil {
		return e
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var a agentRow
		if e := actorScoped(tx.Table("game_agents").Clauses(clause.Locking{Strength: "UPDATE"}), "", actor).Where("id=? AND deleted_at IS NULL", id).Take(&a).Error; e != nil {
			return notFound(e, "AGENT_NOT_FOUND")
		}
		var count int64
		if e := actorScoped(tx.Table("game_commission_plans"), "", actor).Where("id=? AND enabled=1 AND deleted_at IS NULL", in.CommissionPlanID).Count(&count).Error; e != nil || count == 0 {
			return NewError("COMMISSION_PLAN_NOT_FOUND", "enabled commission plan not found")
		}
		if e := actorScoped(tx.Table("game_agents"), "", actor).Where("id=?", id).Update("commission_plan_id", in.CommissionPlanID).Error; e != nil {
			return e
		}
		return audit(tx, actor, "agent:plan:assign", "plan.assign", "agent", id, map[string]any{"commissionPlanId": a.CommissionPlanID}, map[string]any{"commissionPlanId": in.CommissionPlanID}, in.Reason, nil, a.Currency)
	})
}

func (s *Service) SaveChannel(id uint64, in ChannelInput, actor Actor) (uint64, error) {
	if e := requireReason(in.Reason); e != nil {
		return 0, e
	}
	payload, _ := json.Marshal(in.TrackingParameters)
	row := map[string]any{"site_id": actor.Scope.SiteID, "code": strings.ToUpper(strings.TrimSpace(in.Code)), "name_zh": in.NameZh, "name_en": in.NameEn, "agent_id": in.AgentID, "country": strings.ToUpper(in.Country), "platform": strings.ToUpper(in.Platform), "registration_url": in.RegistrationURL, "tracking_parameters": string(payload), "enabled": in.Enabled, "updated_at": time.Now()}
	e := s.db.Transaction(func(tx *gorm.DB) error {
		var before map[string]any
		var agentCount int64
		if e := actorScoped(tx.Table("game_agents"), "", actor).Where("id=? AND deleted_at IS NULL", in.AgentID).Count(&agentCount).Error; e != nil || agentCount == 0 {
			return NewError("AGENT_NOT_FOUND", "agent not found")
		}
		if id == 0 {
			row["created_at"] = time.Now()
			if e := tx.Table("game_agent_channels").Create(&row).Error; e != nil {
				return e
			}
			insertedID, e := lastInsertID(tx)
			if e != nil {
				return e
			}
			id = insertedID
		} else {
			if e := actorScoped(tx.Table("game_agent_channels"), "", actor).Where("id=? AND deleted_at IS NULL", id).Take(&before).Error; e != nil {
				return notFound(e, "CHANNEL_NOT_FOUND")
			}
			delete(row, "code")
			delete(row, "")
			delete(row, "site_id")
			if e := actorScoped(tx.Table("game_agent_channels"), "", actor).Where("id=?", id).Updates(row).Error; e != nil {
				return e
			}
		}
		return audit(tx, actor, permission(before == nil, "agent:channel:create", "agent:channel:update"), permission(before == nil, "create", "update"), "agent_channel", id, before, row, in.Reason, nil, "")
	})
	if duplicate(e) {
		return 0, NewError("DUPLICATE_CHANNEL_CODE", "channel code already exists")
	}
	return id, e
}
func (s *Service) SetChannelStatus(id uint64, in StatusCommand, actor Actor) error {
	return s.simpleStatus("game_agent_channels", "agent_channel", id, "enabled", in.Enabled, in.Reason, "agent:channel:status", actor)
}

func (s *Service) SavePlan(id uint64, in CommissionPlanInput, actor Actor) (uint64, error) {
	if e := requireReason(in.Reason); e != nil {
		return 0, e
	}
	cpa, e := money(in.CPAAmount)
	if e != nil {
		return 0, e
	}
	rate, e := decimal.NewFromString(in.RevenueShareRate)
	if e != nil || rate.IsNegative() || rate.GreaterThan(decimal.NewFromInt(100)) {
		return 0, NewError("INVALID_RATE", "invalid revenue share rate")
	}
	minimum, e := money(in.MinimumSettlementAmount)
	if e != nil {
		return 0, e
	}
	row := map[string]any{"site_id": actor.Scope.SiteID, "code": strings.ToUpper(strings.TrimSpace(in.Code)), "name_zh": in.NameZh, "name_en": in.NameEn, "model": in.Model, "currency": strings.ToUpper(in.Currency), "cpa_amount": cpa, "revenue_share_rate": rate, "negative_carryover": in.NegativeCarryover, "minimum_settlement_amount": minimum, "settlement_cycle": in.SettlementCycle, "deduction_rules": jsonString(in.DeductionRules), "tier_rules": jsonString(in.TierRules), "eligible_products": jsonString(in.EligibleProducts), "enabled": in.Enabled, "updated_at": time.Now()}
	e = s.db.Transaction(func(tx *gorm.DB) error {
		var before map[string]any
		if id == 0 {
			row["created_at"] = time.Now()
			if e := tx.Table("game_commission_plans").Create(&row).Error; e != nil {
				return e
			}
			insertedID, e := lastInsertID(tx)
			if e != nil {
				return e
			}
			id = insertedID
		} else {
			if e := actorScoped(tx.Table("game_commission_plans"), "", actor).Where("id=? AND deleted_at IS NULL", id).Take(&before).Error; e != nil {
				return notFound(e, "PLAN_NOT_FOUND")
			}
			delete(row, "code")
			delete(row, "")
			delete(row, "site_id")
			if e := actorScoped(tx.Table("game_commission_plans"), "", actor).Where("id=?", id).Updates(row).Error; e != nil {
				return e
			}
		}
		return audit(tx, actor, permission(before == nil, "agent:commission-plan:create", "agent:commission-plan:update"), permission(before == nil, "create", "update"), "commission_plan", id, before, row, in.Reason, nil, in.Currency)
	})
	if duplicate(e) {
		return 0, NewError("DUPLICATE_PLAN_CODE", "commission plan code already exists")
	}
	return id, e
}
func (s *Service) SetPlanStatus(id uint64, in StatusCommand, actor Actor) error {
	return s.simpleStatus("game_commission_plans", "commission_plan", id, "enabled", in.Enabled, in.Reason, "agent:commission-plan:update", actor)
}
func (s *Service) DeletePlan(id uint64, reason string, actor Actor) error {
	if e := requireReason(reason); e != nil {
		return e
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var n int64
		if e := actorScoped(tx.Table("game_agents"), "", actor).Where("commission_plan_id=? AND deleted_at IS NULL", id).Count(&n).Error; e != nil {
			return e
		}
		if n > 0 {
			return NewError("PLAN_IN_USE", "commission plan is assigned to agents")
		}
		if e := actorScoped(tx.Table("game_commission_plans"), "", actor).Where("id=? AND deleted_at IS NULL", id).Update("deleted_at", time.Now()).Error; e != nil {
			return e
		}
		return audit(tx, actor, "agent:commission-plan:delete", "delete", "commission_plan", id, nil, map[string]any{"deleted": true}, reason, nil, "")
	})
}
func (s *Service) simpleStatus(table string, target string, id uint64, column string, value bool, reason string, perm string, actor Actor) error {
	if e := requireReason(reason); e != nil {
		return e
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row map[string]any
		if e := actorScoped(tx.Table(table), "", actor).Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; e != nil {
			return notFound(e, "RECORD_NOT_FOUND")
		}
		if databaseBool(row[column]) == value {
			return NewError("STATUS_UNCHANGED", "record already has requested status")
		}
		if e := actorScoped(tx.Table(table), "", actor).Where("id=?", id).Update(column, value).Error; e != nil {
			return e
		}
		return audit(tx, actor, perm, "status.update", target, id, map[string]any{column: row[column]}, map[string]any{column: value}, reason, nil, "")
	})
}

func (s *Service) GenerateSettlement(in GenerateSettlementInput, actor Actor) (uint64, error) {
	if e := requireReason(in.Reason); e != nil {
		return 0, e
	}
	if in.PeriodEnd.Before(in.PeriodStart) {
		return 0, NewError("INVALID_PERIOD", "period end must not precede period start")
	}
	var id uint64
	in.Currency = strings.ToUpper(strings.TrimSpace(in.Currency))
	e := s.db.Transaction(func(tx *gorm.DB) error {
		var existing struct {
			ID                     uint64
			AgentID                uint64
			PeriodStart, PeriodEnd time.Time
			Currency               string
		}
		if e := actorScoped(tx.Table("game_agent_settlements"), "", actor).Where("idempotency_key=?", in.IdempotencyKey).Take(&existing).Error; e == nil {
			if existing.AgentID != in.AgentID || existing.Currency != in.Currency || !sameDate(existing.PeriodStart, in.PeriodStart) || !sameDate(existing.PeriodEnd, in.PeriodEnd) {
				return NewError("IDEMPOTENCY_CONFLICT", "idempotency key was already used for another settlement")
			}
			id = existing.ID
			return nil
		}
		var agent agentRow
		if e := actorScoped(tx.Table("game_agents"), "", actor).Where("id=? AND status='ACTIVE' AND deleted_at IS NULL", in.AgentID).Take(&agent).Error; e != nil {
			return NewError("AGENT_NOT_FOUND", "active agent not found")
		}
		var records []struct {
			ID               uint64
			CommissionAmount decimal.Decimal
		}
		if e := actorScoped(tx.Table("game_commission_records").Clauses(clause.Locking{Strength: "UPDATE"}), "", actor).Select("id,commission_amount").Where("agent_id=? AND currency=? AND business_date BETWEEN ? AND ? AND status='CONFIRMED' AND settlement_id=0 AND deleted_at IS NULL", in.AgentID, in.Currency, in.PeriodStart, in.PeriodEnd).Find(&records).Error; e != nil {
			return e
		}
		gross := decimal.Zero
		for _, r := range records {
			gross = gross.Add(r.CommissionAmount)
		}
		row := map[string]any{"site_id": actor.Scope.SiteID, "settlement_no": nowNo("ST"), "agent_id": in.AgentID, "period_start": in.PeriodStart, "period_end": in.PeriodEnd, "gross_commission": gross, "adjustments": decimal.Zero, "final_amount": gross, "currency": strings.ToUpper(in.Currency), "status": "DRAFT", "creator_id": actor.ID, "creator_username": actor.Username, "idempotency_key": in.IdempotencyKey, "created_at": time.Now(), "updated_at": time.Now()}
		if e := tx.Table("game_agent_settlements").Create(&row).Error; e != nil {
			return e
		}
		insertedID, e := lastInsertID(tx)
		if e != nil {
			return e
		}
		id = insertedID
		for _, r := range records {
			item := map[string]any{"site_id": actor.Scope.SiteID, "settlement_id": id, "commission_record_id": r.ID, "amount": r.CommissionAmount, "created_at": time.Now(), "updated_at": time.Now()}
			if e := tx.Table("game_agent_settlement_items").Create(&item).Error; e != nil {
				return e
			}
		}
		if len(records) > 0 {
			ids := make([]uint64, len(records))
			for i, r := range records {
				ids[i] = r.ID
			}
			if e := actorScoped(tx.Table("game_commission_records"), "", actor).Where("id IN ?", ids).Updates(map[string]any{"status": "FROZEN", "settlement_id": id}).Error; e != nil {
				return e
			}
		}
		return audit(tx, actor, "agent:settlement:generate", "generate", "agent_settlement", id, nil, map[string]any{"agentId": in.AgentID, "grossCommission": gross.StringFixed(2)}, in.Reason, &gross, in.Currency)
	})
	if duplicate(e) {
		return 0, NewError("DUPLICATE_SETTLEMENT", "settlement already exists for period")
	}
	return id, e
}

func (s *Service) SettlementAction(id uint64, action string, reason string, actor Actor) error {
	if e := requireReason(reason); e != nil {
		return e
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row struct {
			ID, AgentID, CreatorID uint64
			Status, Currency       string
			FinalAmount            decimal.Decimal
		}
		if e := actorScoped(tx.Table("game_agent_settlements").Clauses(clause.Locking{Strength: "UPDATE"}), "", actor).Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; e != nil {
			return notFound(e, "SETTLEMENT_NOT_FOUND")
		}
		next := ""
		perm := "agent:settlement:" + action
		updates := map[string]any{"updated_at": time.Now()}
		switch action {
		case "submit":
			if row.Status != "DRAFT" {
				return NewError("INVALID_SETTLEMENT_TRANSITION", "only draft settlement can be submitted")
			}
			next = "PENDING_REVIEW"
			updates["submitted_at"] = time.Now()
		case "approve":
			if row.Status != "PENDING_REVIEW" {
				return NewError("INVALID_SETTLEMENT_TRANSITION", "only pending settlement can be approved")
			}
			if row.CreatorID == actor.ID {
				return NewError("MAKER_CHECKER_VIOLATION", "creator cannot approve own settlement")
			}
			next = "APPROVED"
			updates["reviewed_at"] = time.Now()
			updates["reviewer_id"] = actor.ID
			updates["reviewer_username"] = actor.Username
			updates["review_reason"] = reason
		case "reject":
			if row.Status != "PENDING_REVIEW" {
				return NewError("INVALID_SETTLEMENT_TRANSITION", "only pending settlement can be rejected")
			}
			next = "REJECTED"
			updates["reviewed_at"] = time.Now()
			updates["reviewer_id"] = actor.ID
			updates["reviewer_username"] = actor.Username
			updates["review_reason"] = reason
		case "cancel":
			if row.Status != "DRAFT" && row.Status != "REJECTED" {
				return NewError("INVALID_SETTLEMENT_TRANSITION", "settlement cannot be cancelled")
			}
			next = "CANCELLED"
		case "mark-paid":
			if row.Status == "PAID" {
				return nil
			}
			if row.Status != "APPROVED" {
				return NewError("INVALID_SETTLEMENT_TRANSITION", "only approved settlement can be paid")
			}
			next = "PAID"
			updates["completed_at"] = time.Now()
			if e := s.creditSettlement(tx, row.AgentID, row.Currency, row.FinalAmount, id, reason, actor); e != nil {
				return e
			}
		default:
			return NewError("INVALID_ACTION", "unsupported settlement action")
		}
		updates["status"] = next
		if e := actorScoped(tx.Table("game_agent_settlements"), "", actor).Where("id=?", id).Updates(updates).Error; e != nil {
			return e
		}
		if next == "PAID" {
			if e := actorScoped(tx.Table("game_commission_records"), "", actor).Where("settlement_id=?", id).Updates(map[string]any{"status": "SETTLED"}).Error; e != nil {
				return e
			}
		} else if next == "CANCELLED" {
			if e := actorScoped(tx.Table("game_commission_records"), "", actor).Where("settlement_id=?", id).Updates(map[string]any{"status": "CONFIRMED", "settlement_id": 0}).Error; e != nil {
				return e
			}
		}
		return audit(tx, actor, perm, "settlement."+action, "agent_settlement", id, map[string]any{"status": row.Status}, map[string]any{"status": next}, reason, &row.FinalAmount, row.Currency)
	})
}

func (s *Service) creditSettlement(tx *gorm.DB, agentID uint64, currency string, amount decimal.Decimal, settlementID uint64, reason string, actor Actor) error {
	var w struct {
		ID                              uint64
		AvailableBalance, SettledAmount decimal.Decimal
		Version                         uint64
	}
	if e := actorScoped(tx.Table("game_agent_wallets").Clauses(clause.Locking{Strength: "UPDATE"}), "", actor).Where("agent_id=? AND currency=? AND deleted_at IS NULL", agentID, currency).Take(&w).Error; e != nil {
		return e
	}
	after := w.AvailableBalance.Add(amount)
	if e := actorScoped(tx.Table("game_agent_wallets"), "", actor).Where("id=? AND version=?", w.ID, w.Version).Updates(map[string]any{"available_balance": after, "settled_amount": w.SettledAmount.Add(amount), "version": w.Version + 1}).Error; e != nil {
		return e
	}
	tr := map[string]any{"site_id": actor.Scope.SiteID, "transaction_no": nowNo("AT"), "wallet_id": w.ID, "agent_id": agentID, "transaction_type": "SETTLEMENT_CREDIT", "amount": amount, "balance_before": w.AvailableBalance, "balance_after": after, "currency": currency, "reference_type": "SETTLEMENT", "reference_id": settlementID, "idempotency_key": fmt.Sprintf("settlement:%d", settlementID), "operator_id": actor.ID, "operator_username": actor.Username, "reason": reason, "created_at": time.Now(), "updated_at": time.Now()}
	return tx.Table("game_agent_wallet_transactions").Create(&tr).Error
}

func (s *Service) SetWalletFrozen(id uint64, frozen bool, reason string, actor Actor) error {
	if e := requireReason(reason); e != nil {
		return e
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var w struct {
			ID, AgentID      uint64
			Status, Currency string
			AvailableBalance decimal.Decimal
		}
		if e := actorScoped(tx.Table("game_agent_wallets").Clauses(clause.Locking{Strength: "UPDATE"}), "", actor).Where("id=? AND deleted_at IS NULL", id).Take(&w).Error; e != nil {
			return notFound(e, "AGENT_WALLET_NOT_FOUND")
		}
		next := "ACTIVE"
		typ := "UNFREEZE"
		if frozen {
			next = "FROZEN"
			typ = "FREEZE"
		}
		if w.Status == next {
			return NewError("INVALID_WALLET_STATUS", "wallet already has requested status")
		}
		if e := actorScoped(tx.Table("game_agent_wallets"), "", actor).Where("id=?", id).Update("status", next).Error; e != nil {
			return e
		}
		zero := decimal.Zero
		tr := map[string]any{"site_id": actor.Scope.SiteID, "transaction_no": nowNo("AT"), "wallet_id": id, "agent_id": w.AgentID, "transaction_type": typ, "amount": zero, "balance_before": w.AvailableBalance, "balance_after": w.AvailableBalance, "currency": w.Currency, "idempotency_key": nowNo("freeze:"), "operator_id": actor.ID, "operator_username": actor.Username, "reason": reason, "created_at": time.Now(), "updated_at": time.Now()}
		if e := tx.Table("game_agent_wallet_transactions").Create(&tr).Error; e != nil {
			return e
		}
		return audit(tx, actor, "agent:wallet:freeze", "wallet."+strings.ToLower(typ), "agent_wallet", id, map[string]any{"status": w.Status}, map[string]any{"status": next}, reason, nil, w.Currency)
	})
}

func (s *Service) AdjustWallet(id uint64, in WalletAdjustmentInput, actor Actor) error {
	if e := requireReason(in.Reason); e != nil {
		return e
	}
	amount, e := money(in.Amount)
	if e != nil || amount.IsZero() {
		return NewError("INVALID_AMOUNT", "amount must be greater than zero")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var exists int64
		if e := actorScoped(tx.Table("game_agent_wallet_transactions"), "", actor).Where("idempotency_key=?", in.IdempotencyKey).Count(&exists).Error; e != nil {
			return e
		}
		if exists > 0 {
			return nil
		}
		var w struct {
			ID, AgentID, Version uint64
			AvailableBalance     decimal.Decimal
			Currency, Status     string
		}
		if e := actorScoped(tx.Table("game_agent_wallets").Clauses(clause.Locking{Strength: "UPDATE"}), "", actor).Where("id=? AND deleted_at IS NULL", id).Take(&w).Error; e != nil {
			return notFound(e, "AGENT_WALLET_NOT_FOUND")
		}
		if w.Status != "ACTIVE" {
			return NewError("WALLET_FROZEN", "agent wallet is not active")
		}
		after := w.AvailableBalance
		if in.Type == "MANUAL_CREDIT" {
			after = after.Add(amount)
		} else {
			after = after.Sub(amount)
			if after.IsNegative() {
				return NewError("INSUFFICIENT_BALANCE", "insufficient agent wallet balance")
			}
		}
		if e := actorScoped(tx.Table("game_agent_wallets"), "", actor).Where("id=? AND version=?", id, w.Version).Updates(map[string]any{"available_balance": after, "version": w.Version + 1}).Error; e != nil {
			return e
		}
		tr := map[string]any{"site_id": actor.Scope.SiteID, "transaction_no": nowNo("AT"), "wallet_id": id, "agent_id": w.AgentID, "transaction_type": in.Type, "amount": amount, "balance_before": w.AvailableBalance, "balance_after": after, "currency": w.Currency, "reference_type": "ADMIN_ADJUSTMENT", "idempotency_key": in.IdempotencyKey, "operator_id": actor.ID, "operator_username": actor.Username, "reason": in.Reason, "created_at": time.Now(), "updated_at": time.Now()}
		if e := tx.Table("game_agent_wallet_transactions").Create(&tr).Error; e != nil {
			return e
		}
		return audit(tx, actor, "agent:wallet:adjust", "wallet.adjust", "agent_wallet", id, map[string]any{"balance": w.AvailableBalance.StringFixed(2)}, map[string]any{"balance": after.StringFixed(2), "type": in.Type}, in.Reason, &amount, w.Currency)
	})
}

func (s *Service) ChangeAttribution(in AttributionInput, actor Actor) error {
	if e := requireReason(in.Reason); e != nil {
		return e
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var userCount int64
		if e := actorScoped(tx.Table("game_users"), "", actor).Where("id=? AND deleted_at IS NULL", in.UserID).Count(&userCount).Error; e != nil || userCount == 0 {
			return NewError("USER_NOT_FOUND", "user not found")
		}
		var agentCount int64
		if e := actorScoped(tx.Table("game_agents"), "", actor).Where("id=? AND status='ACTIVE' AND deleted_at IS NULL", in.AgentID).Count(&agentCount).Error; e != nil || agentCount == 0 {
			return NewError("AGENT_NOT_FOUND", "active agent not found")
		}
		if in.ChannelID > 0 {
			var channelCount int64
			if e := actorScoped(tx.Table("game_agent_channels"), "", actor).Where("id=? AND agent_id=? AND enabled=1 AND deleted_at IS NULL", in.ChannelID, in.AgentID).Count(&channelCount).Error; e != nil || channelCount == 0 {
				return NewError("CHANNEL_NOT_FOUND", "enabled channel does not belong to agent")
			}
		}
		var current struct{ ID, AgentID, ChannelID uint64 }
		_ = actorScoped(tx.Table("game_user_agent_attributions").Clauses(clause.Locking{Strength: "UPDATE"}), "", actor).Where("user_id=? AND active=1 AND deleted_at IS NULL", in.UserID).Take(&current).Error
		if current.AgentID == in.AgentID && current.ChannelID == in.ChannelID {
			return NewError("ATTRIBUTION_UNCHANGED", "user attribution is unchanged")
		}
		if current.ID > 0 {
			if e := actorScoped(tx.Table("game_user_agent_attributions"), "", actor).Where("id=?", current.ID).Updates(map[string]any{"agent_id": in.AgentID, "channel_id": in.ChannelID, "attribution_source": in.Source, "attributed_at": time.Now(), "updated_at": time.Now()}).Error; e != nil {
				return e
			}
		} else {
			row := map[string]any{"site_id": actor.Scope.SiteID, "user_id": in.UserID, "agent_id": in.AgentID, "channel_id": in.ChannelID, "attribution_source": in.Source, "active": true, "attributed_at": time.Now(), "created_at": time.Now(), "updated_at": time.Now()}
			if e := tx.Table("game_user_agent_attributions").Create(&row).Error; e != nil {
				return e
			}
		}
		history := map[string]any{"site_id": actor.Scope.SiteID, "user_id": in.UserID, "before_agent_id": current.AgentID, "after_agent_id": in.AgentID, "before_channel_id": current.ChannelID, "after_channel_id": in.ChannelID, "attribution_source": in.Source, "operator_id": actor.ID, "operator_username": actor.Username, "reason": in.Reason, "created_at": time.Now(), "updated_at": time.Now()}
		if e := tx.Table("game_user_agent_attribution_history").Create(&history).Error; e != nil {
			return e
		}
		return audit(tx, actor, "agent:attribution:update", "attribution.change", "agent", in.AgentID, map[string]any{"userId": in.UserID, "agentId": current.AgentID, "channelId": current.ChannelID}, map[string]any{"userId": in.UserID, "agentId": in.AgentID, "channelId": in.ChannelID}, in.Reason, nil, "")
	})
}
