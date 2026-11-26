package riskadmin

import (
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/model"
)

type eventRow struct {
	ID                 uint64    `json:"id"`
	EventNo            string    `json:"eventID"`
	UserID             uint64    `json:"userID"`
	UserNo             string    `json:"userNo"`
	Username           string    `json:"username"`
	RuleID             uint64    `json:"ruleID"`
	RuleCode           string    `json:"rule"`
	EventType          string    `json:"eventType"`
	RiskType           string    `json:"riskType"`
	RiskLevel          string    `json:"riskLevel"`
	TriggerValue       string    `json:"triggerValue"`
	ProviderCode       string    `json:"provider"`
	GameID             uint64    `json:"gameID"`
	IPAddress          string    `json:"ip"`
	DeviceID           string    `json:"deviceID"`
	Status             string    `json:"status"`
	AssignedReviewerID uint64    `json:"assignedReviewerID"`
	AssignedReviewer   string    `json:"assignedReviewer"`
	Source             string    `json:"source"`
	Description        string    `json:"description"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}
type eventRecord struct {
	ID                 uint64
	EventNo            string
	UserID             uint64
	RuleID             uint64
	EventType          string
	RiskType           string
	RiskLevel          string
	TriggerValue       string
	ProviderCode       string
	GameID             uint64
	IPAddress          string
	DeviceID           string
	Status             string
	AssignedReviewerID uint64
	AssignedReviewer   string
	Source             string
	Description        string
	OperatorID         uint64
	OperatorUsername   string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time
}

func (eventRecord) TableName() string { return "game_user_risk_events" }

type noteRecord struct {
	ID               uint64     `json:"id"`
	TargetType       string     `json:"targetType"`
	TargetID         uint64     `json:"targetID"`
	Content          string     `json:"content"`
	OperatorID       uint64     `json:"operatorID"`
	OperatorUsername string     `json:"operator"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
	DeletedAt        *time.Time `json:"-"`
}

func (noteRecord) TableName() string { return "game_risk_internal_notes" }

type RiskEventService struct{ db *gorm.DB }

func eventSelect(db *gorm.DB) *gorm.DB {
	return db.Table("game_user_risk_events e").Select("e.*,u.user_no,u.username,COALESCE(r.code,'') rule_code").Joins("LEFT JOIN game_users u ON u.id=e.user_id AND u.deleted_at IS NULL").Joins("LEFT JOIN game_risk_rules r ON r.id=e.rule_id AND r.deleted_at IS NULL").Where("e.deleted_at IS NULL")
}
func (s *RiskEventService) List(q ListQuery) (PageResult, error) {
	var rows []eventRow
	db := eventSelect(s.db)
	db = contains(db, "e.event_no", q.Values["eventID"])
	db = equal(db, "e.user_id", q.Values["userID"])
	db = contains(db, "u.username", q.Values["username"])
	db = contains(db, "r.code", q.Values["rule"])
	db = equal(db, "e.risk_type", q.Values["riskType"])
	db = equal(db, "e.risk_level", q.Values["riskLevel"])
	db = equal(db, "e.status", q.Values["status"])
	db = equal(db, "e.provider_code", q.Values["provider"])
	db = equal(db, "e.game_id", q.Values["gameID"])
	db = contains(db, "e.ip_address", q.Values["ip"])
	db = contains(db, "e.device_id", q.Values["deviceID"])
	db = timeRange(db, "e.created_at", q.Values["createdFrom"], q.Values["createdTo"])
	return paginate(db.Order("e.id DESC"), q, &rows)
}
func (s *RiskEventService) Detail(id uint64) (map[string]any, error) {
	var row eventRow
	if err := eventSelect(s.db).Where("e.id=?", id).Take(&row).Error; err != nil {
		return nil, mapError(err, "RISK_EVENT_NOT_FOUND")
	}
	var evidence, history, notes, devices, ips, accounts, deposits, withdrawals, bets []map[string]any
	s.db.Table("game_risk_event_evidence").Where("event_id=? AND deleted_at IS NULL", id).Order("id").Find(&evidence)
	s.db.Table("game_risk_review_histories h").Joins("JOIN game_risk_reviews r ON r.id=h.review_id").Where("r.business_type='USER' AND r.business_id=? AND h.deleted_at IS NULL", id).Order("h.id DESC").Find(&history)
	s.db.Table("game_risk_internal_notes").Where("target_type='RISK_EVENT' AND target_id=? AND deleted_at IS NULL", id).Order("id DESC").Find(&notes)
	s.db.Table("game_user_devices").Select("id,device_id,platform,status,risk_status,last_ip,last_seen_at").Where("user_id=? AND deleted_at IS NULL", row.UserID).Find(&devices)
	s.db.Table("game_user_login_records").Select("ip_address,COUNT(*) count,MAX(created_at) last_seen_at").Where("user_id=? AND deleted_at IS NULL", row.UserID).Group("ip_address").Find(&ips)
	s.db.Table("game_users").Select("id,user_no,username,status,risk_review_required").Where("referrer_user_id=? AND deleted_at IS NULL", row.UserID).Limit(50).Find(&accounts)
	s.db.Table("game_deposit_orders").Select("id,order_no,amount,currency,status,created_at").Where("user_id=? AND deleted_at IS NULL", row.UserID).Order("id DESC").Limit(50).Find(&deposits)
	s.db.Table("game_withdraw_orders").Select("id,order_no,amount,currency,status,risk_status,created_at").Where("user_id=? AND deleted_at IS NULL", row.UserID).Order("id DESC").Limit(50).Find(&withdrawals)
	s.db.Table("game_bet_records").Select("id,bet_no,bet_amount_minor,win_amount_minor,currency,bet_status status,created_at").Where("user_id=? AND deleted_at IS NULL", row.UserID).Order("id DESC").Limit(50).Find(&bets)
	logs, _ := operationLogs(s.db, "risk_event", id)
	row.IPAddress = mask(row.IPAddress)
	row.DeviceID = mask(row.DeviceID)
	return map[string]any{"riskSummary": row, "triggerData": evidence, "userInformation": map[string]any{"userID": row.UserID, "username": row.Username}, "relatedIPs": ips, "relatedDevices": devices, "relatedAccounts": accounts, "relatedDeposits": deposits, "relatedWithdrawals": withdrawals, "relatedBets": bets, "reviewHistory": history, "internalNotes": notes, "operationLogs": logs}, nil
}
func (s *RiskEventService) Create(userID uint64, riskType string, riskLevel string, reason string, actor Actor) (uint64, error) {
	if userID == 0 || !valid(riskLevel, "LOW", "MEDIUM", "HIGH", "CRITICAL") {
		return 0, newError("INVALID_REQUEST", "invalid user or risk level")
	}
	row := eventRecord{EventNo: generatedNo("RE"), UserID: userID, EventType: riskType, RiskType: riskType, RiskLevel: riskLevel, Status: "PENDING", Source: "MANUAL", Description: reason, OperatorID: actor.ID, OperatorUsername: actor.Username}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		if err := upsertRiskProfile(tx, userID, riskLevel, "OBSERVED", reason, actor); err != nil {
			return err
		}
		return audit(tx, actor, "risk:event:review", "risk_event.create", "risk_event", row.ID, userID, nil, row, reason)
	})
	return row.ID, err
}
func (s *RiskEventService) Assign(id uint64, cmd AssignmentCommand, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row eventRecord
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
			return mapError(err, "RISK_EVENT_NOT_FOUND")
		}
		if !valid(row.Status, "PENDING", "UNDER_REVIEW") {
			return newError("INVALID_RISK_EVENT_TRANSITION", "event cannot be assigned")
		}
		before := row
		if err := tx.Model(&row).Updates(map[string]any{"assigned_reviewer_id": cmd.ReviewerID, "assigned_reviewer": cmd.ReviewerName, "status": "UNDER_REVIEW"}).Error; err != nil {
			return err
		}
		return audit(tx, actor, "risk:event:assign", "risk_event.assign", "risk_event", id, row.UserID, before, map[string]any{"status": "UNDER_REVIEW", "reviewer": cmd.ReviewerName}, cmd.Reason)
	})
}
func (s *RiskEventService) Transition(id uint64, target string, cmd EventActionCommand, permission string, actor Actor) error {
	target = strings.ToUpper(target)
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row eventRecord
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
			return mapError(err, "RISK_EVENT_NOT_FOUND")
		}
		if !eventTransition(row.Status, target) {
			return newError("INVALID_RISK_EVENT_TRANSITION", "invalid risk event state transition")
		}
		before := row
		updates := map[string]any{"status": target, "operator_id": actor.ID, "operator_username": actor.Username}
		if target == "UNDER_REVIEW" {
			updates["assigned_reviewer_id"] = actor.ID
			updates["assigned_reviewer"] = actor.Username
		}
		if err := tx.Model(&row).Updates(updates).Error; err != nil {
			return err
		}
		profileStatus := target
		if target == "CONFIRMED" {
			profileStatus = "UNDER_REVIEW"
		}
		if err := upsertRiskProfile(tx, row.UserID, row.RiskLevel, profileStatus, cmd.Reason, actor); err != nil {
			return err
		}
		if target == "RESTRICTED" {
			if err := tx.Table("game_users").Where("id=?", row.UserID).Updates(map[string]any{"withdrawal_enabled": 0, "risk_review_required": 1}).Error; err != nil {
				return err
			}
		}
		if target == "BLOCKED" {
			if err := tx.Table("game_users").Where("id=?", row.UserID).Updates(map[string]any{"status": 2, "login_enabled": 0, "game_access_enabled": 0, "deposit_enabled": 0, "withdrawal_enabled": 0, "risk_review_required": 1}).Error; err != nil {
				return err
			}
		}
		if cmd.RiskTags != "" {
			if err := addRiskTags(tx, row.UserID, cmd.RiskTags); err != nil {
				return err
			}
		}
		return audit(tx, actor, permission, "risk_event."+strings.ToLower(target), "risk_event", id, row.UserID, before, map[string]any{"status": target}, cmd.Reason)
	})
}
func eventTransition(from string, to string) bool {
	switch to {
	case "UNDER_REVIEW":
		return from == "PENDING"
	case "CONFIRMED", "CLEARED":
		return from == "PENDING" || from == "UNDER_REVIEW"
	case "RESTRICTED", "BLOCKED":
		return from == "PENDING" || from == "UNDER_REVIEW" || from == "CONFIRMED"
	}
	return false
}
func (s *RiskEventService) AddNote(id uint64, cmd NoteCommand, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row eventRecord
		if err := tx.Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
			return mapError(err, "RISK_EVENT_NOT_FOUND")
		}
		note := noteRecord{TargetType: "RISK_EVENT", TargetID: id, Content: cmd.Content, OperatorID: actor.ID, OperatorUsername: actor.Username}
		if err := tx.Create(&note).Error; err != nil {
			return err
		}
		return audit(tx, actor, "risk:event:note", "risk_event.note", "risk_event", id, row.UserID, nil, map[string]any{"noteID": note.ID}, "internal note added")
	})
}
func upsertRiskProfile(tx *gorm.DB, userID uint64, level string, status string, reason string, actor Actor) error {
	return tx.Exec("INSERT INTO game_user_risk_profiles (user_id, risk_level, risk_status, trigger_count, associated_accounts, assigned_reviewer_id, assigned_reviewer, latest_reason, created_at, updated_at) VALUES (?, ?, ?, 1, 0, ?, ?, ?, NOW(3), NOW(3)) ON DUPLICATE KEY UPDATE risk_level = VALUES(risk_level), risk_status = VALUES(risk_status), trigger_count = trigger_count + 1, assigned_reviewer_id = VALUES(assigned_reviewer_id), assigned_reviewer = VALUES(assigned_reviewer), latest_reason = VALUES(latest_reason), updated_at = NOW(3)", userID, level, status, actor.ID, actor.Username, reason).Error
}
func addRiskTags(tx *gorm.DB, userID uint64, tags string) error {
	for _, name := range strings.Split(tags, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		tag := model.GameUserTags{Name: name, TagType: "RISK"}
		if err := tx.Where("name=? AND tag_type='RISK'", name).FirstOrCreate(&tag).Error; err != nil {
			return err
		}
		rel := model.GameUserTagRelations{UserID: userID, TagID: tag.ID}
		if err := tx.Where("user_id=? AND tag_id=?", userID, tag.ID).FirstOrCreate(&rel).Error; err != nil {
			return err
		}
	}
	return nil
}
