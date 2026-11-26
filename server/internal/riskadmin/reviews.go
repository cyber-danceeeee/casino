package riskadmin

import (
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type reviewRecord struct {
	ID                 uint64     `json:"id"`
	ReviewNo           string     `json:"reviewID"`
	BusinessType       string     `json:"businessType"`
	BusinessID         uint64     `json:"businessID"`
	UserID             uint64     `json:"userID"`
	RiskLevel          string     `json:"riskLevel"`
	Reason             string     `json:"reason"`
	Status             string     `json:"status"`
	AssignedReviewerID uint64     `json:"assignedReviewerID"`
	AssignedReviewer   string     `json:"assignedReviewer"`
	CreatedBy          uint64     `json:"createdBy"`
	Decision           string     `json:"decision"`
	InternalRemark     string     `json:"internalRemark"`
	RiskTags           string     `json:"riskTags"`
	Version            uint       `json:"version"`
	CompletedAt        *time.Time `json:"completedAt"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
	DeletedAt          *time.Time `json:"-"`
}

func (reviewRecord) TableName() string { return "game_risk_reviews" }

type reviewRow struct {
	reviewRecord
	UserNo   string `json:"userNo"`
	Username string `json:"username"`
}
type historyRecord struct {
	ID               uint64     `json:"id"`
	ReviewID         uint64     `json:"reviewID"`
	BeforeStatus     string     `json:"beforeStatus"`
	AfterStatus      string     `json:"afterStatus"`
	Decision         string     `json:"decision"`
	Reason           string     `json:"reason"`
	InternalRemark   string     `json:"internalRemark"`
	RiskTags         string     `json:"riskTags"`
	OperatorID       uint64     `json:"operatorID"`
	OperatorUsername string     `json:"operator"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
	DeletedAt        *time.Time `json:"-"`
}

func (historyRecord) TableName() string { return "game_risk_review_histories" }

type RiskReviewService struct{ db *gorm.DB }

func reviewSelect(db *gorm.DB) *gorm.DB {
	return db.Table("game_risk_reviews r").Select("r.*,u.user_no,u.username").Joins("LEFT JOIN game_users u ON u.id=r.user_id AND u.deleted_at IS NULL").Where("r.deleted_at IS NULL")
}
func (s *RiskReviewService) List(q ListQuery) (PageResult, error) {
	var rows []reviewRow
	db := reviewSelect(s.db)
	db = contains(db, "r.review_no", q.Values["reviewID"])
	db = equal(db, "r.business_type", q.Values["businessType"])
	db = equal(db, "r.user_id", q.Values["userID"])
	db = contains(db, "u.username", q.Values["username"])
	db = equal(db, "r.risk_level", q.Values["riskLevel"])
	db = equal(db, "r.status", q.Values["status"])
	db = contains(db, "r.assigned_reviewer", q.Values["reviewer"])
	db = timeRange(db, "r.created_at", q.Values["createdFrom"], q.Values["createdTo"])
	return paginate(db.Order("FIELD(r.risk_level,'CRITICAL','HIGH','MEDIUM','LOW'),r.id ASC"), q, &rows)
}
func (s *RiskReviewService) Detail(id uint64) (map[string]any, error) {
	var row reviewRow
	if err := reviewSelect(s.db).Where("r.id=?", id).Take(&row).Error; err != nil {
		return nil, mapError(err, "RISK_REVIEW_NOT_FOUND")
	}
	var history, notes []map[string]any
	s.db.Table("game_risk_review_histories").Where("review_id=? AND deleted_at IS NULL", id).Order("id DESC").Find(&history)
	s.db.Table("game_risk_internal_notes").Where("target_type='RISK_REVIEW' AND target_id=? AND deleted_at IS NULL", id).Order("id DESC").Find(&notes)
	logs, _ := operationLogs(s.db, "risk_review", id)
	return map[string]any{"overview": row, "reviewHistory": history, "internalNotes": notes, "operationLogs": logs}, nil
}
func (s *RiskReviewService) Create(tx *gorm.DB, businessType string, businessID uint64, userID uint64, level string, reason string, actor Actor) (uint64, error) {
	row := reviewRecord{ReviewNo: generatedNo("RR"), BusinessType: businessType, BusinessID: businessID, UserID: userID, RiskLevel: level, Reason: reason, Status: "PENDING", CreatedBy: actor.ID, Version: 1}
	if err := tx.Create(&row).Error; err != nil {
		return 0, err
	}
	return row.ID, nil
}
func (s *RiskReviewService) Claim(id uint64, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row reviewRecord
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
			return mapError(err, "RISK_REVIEW_NOT_FOUND")
		}
		if row.Status != "PENDING" {
			return newError("RISK_REVIEW_ALREADY_CLAIMED", "review is no longer available")
		}
		before := row
		if err := tx.Model(&row).Updates(map[string]any{"status": "ASSIGNED", "assigned_reviewer_id": actor.ID, "assigned_reviewer": actor.Username, "version": gorm.Expr("version+1")}).Error; err != nil {
			return err
		}
		if err := addReviewHistory(tx, row.ID, row.Status, "ASSIGNED", "CLAIM", "review claimed", "", "", actor); err != nil {
			return err
		}
		return audit(tx, actor, "risk:review:claim", "risk_review.claim", "risk_review", id, row.UserID, before, map[string]any{"status": "ASSIGNED", "reviewer": actor.Username}, "review claimed")
	})
}
func (s *RiskReviewService) Assign(id uint64, cmd AssignmentCommand, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row reviewRecord
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
			return mapError(err, "RISK_REVIEW_NOT_FOUND")
		}
		if !valid(row.Status, "PENDING", "ASSIGNED") {
			return newError("INVALID_RISK_REVIEW_TRANSITION", "review cannot be assigned")
		}
		before := row
		if err := tx.Model(&row).Updates(map[string]any{"status": "ASSIGNED", "assigned_reviewer_id": cmd.ReviewerID, "assigned_reviewer": cmd.ReviewerName, "version": gorm.Expr("version+1")}).Error; err != nil {
			return err
		}
		if err := addReviewHistory(tx, row.ID, row.Status, "ASSIGNED", "ASSIGN", cmd.Reason, "", "", actor); err != nil {
			return err
		}
		return audit(tx, actor, "risk:review:assign", "risk_review.assign", "risk_review", id, row.UserID, before, map[string]any{"status": "ASSIGNED", "reviewer": cmd.ReviewerName}, cmd.Reason)
	})
}
func (s *RiskReviewService) Decision(id uint64, cmd ReviewDecisionCommand, actor Actor) error {
	decision := strings.ToUpper(cmd.Decision)
	target := map[string]string{"APPROVE": "APPROVED", "REJECT": "REJECTED", "CLEAR": "CLEARED", "ESCALATE": "ESCALATED", "UNDER_REVIEW": "UNDER_REVIEW"}[decision]
	if target == "" {
		return newError("INVALID_REVIEW_DECISION", "unsupported review decision")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row reviewRecord
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
			return mapError(err, "RISK_REVIEW_NOT_FOUND")
		}
		if valid(row.Status, "APPROVED", "REJECTED", "CLEARED") {
			return newError("RISK_REVIEW_ALREADY_DECIDED", "review decision is immutable")
		}
		if actor.ID == row.CreatedBy && decision != "UNDER_REVIEW" {
			return newError("SELF_APPROVAL_FORBIDDEN", "review creator cannot approve own review")
		}
		if row.AssignedReviewerID != 0 && row.AssignedReviewerID != actor.ID {
			return newError("RISK_REVIEW_ASSIGNED_TO_OTHER", "review is assigned to another operator")
		}
		before := row
		updates := map[string]any{"status": target, "decision": decision, "internal_remark": cmd.InternalRemark, "risk_tags": cmd.RiskTags, "assigned_reviewer_id": actor.ID, "assigned_reviewer": actor.Username, "version": gorm.Expr("version+1")}
		if valid(target, "APPROVED", "REJECTED", "CLEARED") {
			updates["completed_at"] = time.Now()
		}
		result := tx.Model(&reviewRecord{}).Where("id=? AND version=?", id, row.Version).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return newError("RISK_REVIEW_CONFLICT", "review was updated concurrently")
		}
		if err := addReviewHistory(tx, row.ID, row.Status, target, decision, cmd.Reason, cmd.InternalRemark, cmd.RiskTags, actor); err != nil {
			return err
		}
		if cmd.RiskTags != "" && row.UserID > 0 {
			if err := addRiskTags(tx, row.UserID, cmd.RiskTags); err != nil {
				return err
			}
		}
		permission := "risk:review:decision"
		if decision == "ESCALATE" {
			permission = "risk:review:escalate"
		}
		return audit(tx, actor, permission, "risk_review."+strings.ToLower(decision), "risk_review", id, row.UserID, before, map[string]any{"status": target, "decision": decision}, cmd.Reason)
	})
}
func (s *RiskReviewService) AddNote(id uint64, cmd NoteCommand, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row reviewRecord
		if err := tx.Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
			return mapError(err, "RISK_REVIEW_NOT_FOUND")
		}
		n := noteRecord{TargetType: "RISK_REVIEW", TargetID: id, Content: cmd.Content, OperatorID: actor.ID, OperatorUsername: actor.Username}
		if err := tx.Create(&n).Error; err != nil {
			return err
		}
		return audit(tx, actor, "risk:review:decision", "risk_review.note", "risk_review", id, row.UserID, nil, map[string]any{"noteID": n.ID}, "internal note added")
	})
}
func addReviewHistory(tx *gorm.DB, id uint64, before string, after string, decision string, reason string, remark string, tags string, actor Actor) error {
	return tx.Create(&historyRecord{ReviewID: id, BeforeStatus: before, AfterStatus: after, Decision: decision, Reason: reason, InternalRemark: remark, RiskTags: tags, OperatorID: actor.ID, OperatorUsername: actor.Username}).Error
}
