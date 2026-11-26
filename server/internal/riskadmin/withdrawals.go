package riskadmin

import (
	"strings"

	"gorm.io/gorm"

	"casino/internal/fundadmin"
	"casino/internal/tenancy"
)

type WithdrawalRiskService struct {
	db    *gorm.DB
	funds *fundadmin.Service
}

func (s *WithdrawalRiskService) List(q ListQuery) (PageResult, error) {
	var rows []map[string]any
	db := s.db.Table("game_withdraw_orders w").Select("w.id,w.order_no withdrawal_order_id,w.user_id,u.user_no,u.username,w.amount,w.currency,w.kyc_status,w.risk_status,COALESCE(p.risk_level,'LOW') risk_level,COALESCE(p.risk_status,'OBSERVED') review_status,w.status,w.created_at,w.updated_at,(SELECT COUNT(*) FROM game_withdraw_accounts wa WHERE wa.user_id=w.user_id AND wa.deleted_at IS NULL) associated_account_count").Joins("JOIN game_users u ON u.id=w.user_id AND u.deleted_at IS NULL").Joins("LEFT JOIN game_user_risk_profiles p ON p.user_id=w.user_id AND p.deleted_at IS NULL").Where("w.deleted_at IS NULL")
	db = tenancy.ApplySite(db, "w", q.Scope)
	db = tenancy.ApplySite(db, "u", q.Scope)
	db = tenancy.ApplySite(db, "p", q.Scope)
	db = contains(db, "w.order_no", q.Values["withdrawalOrderID"])
	db = equal(db, "w.user_id", q.Values["userID"])
	db = contains(db, "u.username", q.Values["username"])
	db = equal(db, "p.risk_level", q.Values["riskLevel"])
	db = equal(db, "w.risk_status", q.Values["riskStatus"])
	db = equal(db, "w.kyc_status", q.Values["kycStatus"])
	db = equal(db, "u.country_code", q.Values["country"])
	db = equal(db, "w.currency", q.Values["currency"])
	if q.Values["amountMin"] != "" {
		db = db.Where("w.amount>=?", q.Values["amountMin"])
	}
	if q.Values["amountMax"] != "" {
		db = db.Where("w.amount<=?", q.Values["amountMax"])
	}
	db = timeRange(db, "w.created_at", q.Values["createdFrom"], q.Values["createdTo"])
	return paginate(db.Order("FIELD(p.risk_level,'CRITICAL','HIGH','MEDIUM','LOW'),w.id DESC"), q, &rows)
}
func (s *WithdrawalRiskService) Detail(id uint64, scope tenancy.Scope) (map[string]any, error) {
	detail, err := s.funds.GetWithdrawal(id, scope)
	if err != nil {
		return nil, err
	}
	order, _ := detail["order"].(fundadmin.WithdrawalRow)
	userID := order.UserID
	var profile map[string]any
	tenancy.ApplySite(s.db.Table("game_user_risk_profiles"), "", scope).Where("user_id=? AND deleted_at IS NULL", userID).Take(&profile)
	var history, bets, bonuses, devices, accounts []map[string]any
	tenancy.ApplySite(s.db.Table("game_withdrawal_approvals"), "", scope).Where("withdrawal_order_id=? AND deleted_at IS NULL", id).Order("id DESC").Find(&history)
	tenancy.ApplySite(s.db.Table("game_bet_records"), "", scope).Select("COUNT(*) bet_count,SUM(bet_amount_minor) total_bet_minor,SUM(win_amount_minor) total_win_minor").Where("user_id=? AND deleted_at IS NULL", userID).Find(&bets)
	s.db.Table("game_user_bonuses").Select("status,COUNT(*) count,SUM(bonus_amount) amount").Where("user_id=? AND deleted_at IS NULL", userID).Group("status").Find(&bonuses)
	tenancy.ApplySite(s.db.Table("game_user_devices"), "", scope).Select("device_id,platform,risk_status,last_ip,last_seen_at").Where("user_id=? AND deleted_at IS NULL", userID).Find(&devices)
	tenancy.ApplySite(s.db.Table("game_withdraw_accounts"), "", scope).Select("account_type,bank_name,wallet_type,CONCAT('**',RIGHT(account_no,4)) account_masked,status").Where("user_id=? AND deleted_at IS NULL", userID).Find(&accounts)
	logs, _ := operationLogs(s.db, "withdrawal_risk", id)
	return map[string]any{"withdrawalSummary": detail, "userRiskProfile": profile, "depositAndWithdrawalHistory": detail, "bettingSummary": bets, "bonusUsage": bonuses, "devicesAndIPs": devices, "associatedAccounts": accounts, "reviewHistory": history, "operationLogs": logs}, nil
}
func (s *WithdrawalRiskService) Decision(id uint64, cmd WithdrawalDecisionCommand, actor Actor) error {
	decision := strings.ToUpper(cmd.Decision)
	fundActor := fundadmin.Actor{ID: actor.ID, Username: actor.Username, Scope: actor.Scope, IP: actor.IP, UserAgent: actor.UserAgent, CorrelationID: actor.CorrelationID}
	reason := fundadmin.ReasonCommand{Reason: cmd.Reason, InternalRemark: cmd.InternalRemark}
	switch decision {
	case "APPROVE":
		return s.funds.ReviewWithdrawal(id, "RISK", reason, fundActor)
	case "REJECT":
		return s.funds.RejectWithdrawal(id, reason, fundActor)
	case "ADDITIONAL_REVIEW":
		return s.additionalReview(id, cmd, actor)
	case "RESTRICT_WITHDRAWAL":
		return s.restrict(id, cmd, actor, false)
	case "FREEZE_ACCOUNT":
		return s.restrict(id, cmd, actor, true)
	default:
		return newError("INVALID_WITHDRAWAL_RISK_DECISION", "unsupported withdrawal risk decision")
	}
}
func (s *WithdrawalRiskService) additionalReview(id uint64, cmd WithdrawalDecisionCommand, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row struct {
			ID         uint64
			UserID     uint64
			RiskStatus string
		}
		if err := tx.Table("game_withdraw_orders").Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
			return mapError(err, "WITHDRAWAL_NOT_FOUND")
		}
		if err := tx.Table("game_withdraw_orders").Where("id=?", id).Update("risk_status", "ADDITIONAL_REVIEW").Error; err != nil {
			return err
		}
		reviews := &RiskReviewService{db: tx}
		reviewID, err := reviews.Create(tx, "WITHDRAWAL", id, row.UserID, "HIGH", cmd.Reason, actor)
		if err != nil {
			return err
		}
		return audit(tx, actor, "risk:withdrawal:restrict", "withdrawal.additional_review", "withdrawal_risk", id, row.UserID, map[string]any{"riskStatus": row.RiskStatus}, map[string]any{"riskStatus": "ADDITIONAL_REVIEW", "reviewID": reviewID}, cmd.Reason)
	})
}
func (s *WithdrawalRiskService) restrict(id uint64, cmd WithdrawalDecisionCommand, actor Actor, freeze bool) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row struct {
			ID         uint64
			UserID     uint64
			RiskStatus string
		}
		if err := tx.Table("game_withdraw_orders").Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
			return mapError(err, "WITHDRAWAL_NOT_FOUND")
		}
		updates := map[string]any{"withdrawal_enabled": 0, "risk_review_required": 1}
		decision := "RESTRICT_WITHDRAWAL"
		if freeze {
			updates["status"] = 3
			decision = "FREEZE_ACCOUNT"
		}
		if err := tx.Table("game_users").Where("id=?", row.UserID).Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.Table("game_withdraw_orders").Where("id=?", id).Update("risk_status", decision).Error; err != nil {
			return err
		}
		if cmd.RiskTags != "" {
			if err := addRiskTags(tx, row.UserID, cmd.RiskTags); err != nil {
				return err
			}
		}
		return audit(tx, actor, "risk:withdrawal:restrict", "withdrawal."+strings.ToLower(decision), "withdrawal_risk", id, row.UserID, map[string]any{"riskStatus": row.RiskStatus}, map[string]any{"riskStatus": decision}, cmd.Reason)
	})
}
