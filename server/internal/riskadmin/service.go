package riskadmin

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"casino/internal/fundadmin"
	"casino/internal/model"
)

type Service struct {
	db           *gorm.DB
	Rules        *RiskRuleService
	Evaluation   *RiskEvaluationService
	Events       *RiskEventService
	Reviews      *RiskReviewService
	Blacklists   *BlacklistService
	Associations *UserAssociationService
	Withdrawals  *WithdrawalRiskService
}

func New(db *gorm.DB) *Service {
	s := &Service{db: db}
	s.Evaluation = &RiskEvaluationService{db: db}
	s.Rules = &RiskRuleService{db: db, evaluation: s.Evaluation}
	s.Events = &RiskEventService{db: db}
	s.Reviews = &RiskReviewService{db: db}
	s.Blacklists = &BlacklistService{db: db}
	s.Associations = &UserAssociationService{db: db}
	s.Withdrawals = &WithdrawalRiskService{db: db, funds: fundadmin.New(db)}
	return s
}

func normalizedPage(q ListQuery) (int, int, int) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = 20
	}
	if q.PageSize > 100 {
		q.PageSize = 100
	}
	return q.Page, q.PageSize, (q.Page - 1) * q.PageSize
}

func paginate(db *gorm.DB, q ListQuery, out any) (PageResult, error) {
	p, size, offset := normalizedPage(q)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return PageResult{}, err
	}
	if err := db.Offset(offset).Limit(size).Find(out).Error; err != nil {
		return PageResult{}, err
	}
	return PageResult{List: out, Total: total, Page: p, PageSize: size}, nil
}

func contains(db *gorm.DB, column, value string) *gorm.DB {
	if value = strings.TrimSpace(value); value != "" {
		return db.Where(column+" LIKE ?", "%"+value+"%")
	}
	return db
}
func equal(db *gorm.DB, column, value string) *gorm.DB {
	if value = strings.TrimSpace(value); value != "" {
		return db.Where(column+" = ?", value)
	}
	return db
}
func timeRange(db *gorm.DB, column, from, to string) *gorm.DB {
	if from != "" {
		db = db.Where(column+" >= ?", from)
	}
	if to != "" {
		db = db.Where(column+" <= ?", to)
	}
	return db
}
func generatedNo(prefix string) string { return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano()) }

func audit(tx *gorm.DB, actor Actor, permission, action, target string, id, userID uint64, before, after any, reason string) error {
	b, _ := json.Marshal(before)
	a, _ := json.Marshal(after)
	return tx.Create(&model.GameAdminOperationLogs{OperatorID: actor.ID, OperatorUsername: actor.Username, Permission: permission, Action: action, TargetType: target, TargetID: id, TargetUserID: userID, BeforeState: string(b), AfterState: string(a), Reason: reason, IPAddress: actor.IP, UserAgent: actor.UserAgent, CorrelationID: actor.CorrelationID}).Error
}

func operationLogs(db *gorm.DB, target string, id uint64) ([]model.GameAdminOperationLogs, error) {
	var rows []model.GameAdminOperationLogs
	err := db.Where("target_type=? AND target_id=? AND deleted_at IS NULL", target, id).Order("id DESC").Limit(100).Find(&rows).Error
	return rows, err
}

func mapError(err error, code string) error {
	if err == gorm.ErrRecordNotFound {
		return newError(code, "record not found")
	}
	return err
}

func valid(value string, allowed ...string) bool {
	for _, item := range allowed {
		if value == item {
			return true
		}
	}
	return false
}

type countPoint struct {
	Name  string `json:"name" gorm:"column:name"`
	Value int64  `json:"value" gorm:"column:value"`
}

func (s *Service) Dashboard(from, to string) (map[string]any, error) {
	events := s.db.Table("game_user_risk_events").Where("deleted_at IS NULL")
	if from != "" {
		events = events.Where("created_at >= ?", from)
	}
	if to != "" {
		events = events.Where("created_at <= ?", to)
	}
	count := func(db *gorm.DB) int64 { var n int64; _ = db.Count(&n).Error; return n }
	var byDay, byType, byLevel, countries, ips, devices []countPoint
	events.Session(&gorm.Session{}).Select("DATE(created_at) name,COUNT(*) value").Group("DATE(created_at)").Order("name").Scan(&byDay)
	events.Session(&gorm.Session{}).Select("risk_type name,COUNT(*) value").Group("risk_type").Order("value DESC").Scan(&byType)
	events.Session(&gorm.Session{}).Select("risk_level name,COUNT(*) value").Group("risk_level").Order("value DESC").Scan(&byLevel)
	s.db.Table("game_users").Select("country_code name,COUNT(*) value").Where("risk_review_required=1 AND deleted_at IS NULL").Group("country_code").Order("value DESC").Limit(10).Scan(&countries)
	events.Session(&gorm.Session{}).Select("ip_address name,COUNT(*) value").Where("ip_address<>''").Group("ip_address").Order("value DESC").Limit(10).Scan(&ips)
	events.Session(&gorm.Session{}).Select("device_id name,COUNT(*) value").Where("device_id<>''").Group("device_id").Order("value DESC").Limit(10).Scan(&devices)
	var reviewResults []countPoint
	s.db.Table("game_risk_reviews").Select("status name,COUNT(*) value").Where("deleted_at IS NULL").Group("status").Scan(&reviewResults)
	return map[string]any{
		"metrics": map[string]int64{
			"pendingReviews":        count(s.db.Table("game_risk_reviews").Where("deleted_at IS NULL AND status IN ?", []string{"PENDING", "ASSIGNED", "UNDER_REVIEW"})),
			"highRiskUsers":         count(s.db.Table("game_user_risk_profiles").Where("deleted_at IS NULL AND risk_level IN ?", []string{"HIGH", "CRITICAL"})),
			"criticalEvents":        count(events.Session(&gorm.Session{}).Where("risk_level='CRITICAL'")),
			"blockedAccounts":       count(s.db.Table("game_users").Where("deleted_at IS NULL AND (status=2 OR login_enabled=0)")),
			"blockedDevices":        count(s.db.Table("game_user_devices").Where("deleted_at IS NULL AND status=2")),
			"blockedIPs":            count(s.db.Table("game_risk_blacklists").Where("deleted_at IS NULL AND blacklist_type='IP' AND status='ENABLED'")),
			"suspiciousWithdrawals": count(s.db.Table("game_withdraw_orders").Where("deleted_at IS NULL AND risk_status NOT IN ('CLEARED','PASSED')")),
			"bonusAbuse":            count(events.Session(&gorm.Session{}).Where("risk_type='BONUS'")),
			"callbackAnomalies":     count(events.Session(&gorm.Session{}).Where("risk_type='PROVIDER_CALLBACK'")),
		},
		"eventsByDay": byDay, "eventsByType": byType, "eventsByLevel": byLevel, "reviewResults": reviewResults,
		"topCountries": countries, "topIPs": maskPoints(ips), "topDevices": maskPoints(devices),
	}, nil
}

func maskPoints(points []countPoint) []countPoint {
	for i := range points {
		points[i].Name = mask(points[i].Name)
	}
	return points
}
func mask(value string) string {
	r := []rune(value)
	if len(r) <= 4 {
		return strings.Repeat("*", len(r))
	}
	return string(r[:2]) + strings.Repeat("*", len(r)-4) + string(r[len(r)-2:])
}
