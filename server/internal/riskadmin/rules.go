package riskadmin

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ruleRecord struct {
	ID              uint64     `json:"id"`
	Code            string     `json:"code"`
	NameZh          string     `json:"nameZh"`
	NameEn          string     `json:"nameEn"`
	DescriptionZh   string     `json:"descriptionZh"`
	DescriptionEn   string     `json:"descriptionEn"`
	RuleType        string     `json:"ruleType"`
	RiskLevel       string     `json:"riskLevel"`
	Priority        int        `json:"priority"`
	Enabled         bool       `json:"enabled"`
	Countries       string     `json:"countries"`
	Currencies      string     `json:"currencies"`
	Channels        string     `json:"channels"`
	TriggerCount    uint64     `json:"triggerCount"`
	LastTriggeredAt *time.Time `json:"lastTriggeredTime"`
	StartAt         *time.Time `json:"startTime"`
	EndAt           *time.Time `json:"endTime"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	DeletedAt       *time.Time `json:"-"`
}

func (ruleRecord) TableName() string { return "game_risk_rules" }

type conditionRecord struct {
	ID              uint64     `json:"id"`
	RuleID          uint64     `json:"ruleID"`
	FieldName       string     `json:"fieldName"`
	Operator        string     `json:"operator"`
	ComparisonValue string     `json:"comparisonValue"`
	WindowSeconds   uint       `json:"windowSeconds"`
	SortOrder       int        `json:"sortOrder"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	DeletedAt       *time.Time `json:"-"`
}

func (conditionRecord) TableName() string { return "game_risk_rule_conditions" }

type actionRecord struct {
	ID             uint64     `json:"id"`
	RuleID         uint64     `json:"ruleID"`
	ActionType     string     `json:"actionType"`
	ParameterValue string     `json:"parameterValue"`
	SortOrder      int        `json:"sortOrder"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	DeletedAt      *time.Time `json:"-"`
}

func (actionRecord) TableName() string { return "game_risk_rule_actions" }

type RiskRuleService struct {
	db         *gorm.DB
	evaluation *RiskEvaluationService
}
type RiskEvaluationService struct{ db *gorm.DB }

var operators = map[string]bool{"EQUAL": true, "NOT_EQUAL": true, "GREATER_THAN": true, "GREATER_THAN_OR_EQUAL": true, "LESS_THAN": true, "LESS_THAN_OR_EQUAL": true, "IN": true, "NOT_IN": true, "CONTAINS": true, "COUNT_WITHIN": true, "SUM_WITHIN": true}
var riskActions = map[string]bool{"CREATE_EVENT": true, "ADD_RISK_TAG": true, "REQUIRE_REVIEW": true, "DISABLE_LOGIN": true, "DISABLE_GAME": true, "DISABLE_WITHDRAWAL": true, "FREEZE_ACCOUNT": true, "BLOCK_DEVICE": true, "BLOCK_IP": true}

func (s *RiskRuleService) List(q ListQuery) (PageResult, error) {
	var rows []ruleRecord
	db := s.db.Model(&ruleRecord{}).Where("deleted_at IS NULL")
	db = contains(db, "code", q.Values["code"])
	db = contains(db, "name_en", q.Values["name"])
	db = equal(db, "rule_type", q.Values["ruleType"])
	db = equal(db, "risk_level", q.Values["riskLevel"])
	if v := q.Values["enabled"]; v != "" {
		db = db.Where("enabled=?", v == "true" || v == "1")
	}
	db = contains(db, "rule_type", q.Values["business"])
	db = timeRange(db, "updated_at", q.Values["updatedFrom"], q.Values["updatedTo"])
	return paginate(db.Order("priority DESC,id DESC"), q, &rows)
}
func (s *RiskRuleService) Detail(id uint64) (map[string]any, error) {
	var row ruleRecord
	if err := s.db.Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return nil, mapError(err, "RISK_RULE_NOT_FOUND")
	}
	var cs []conditionRecord
	var as []actionRecord
	var tests []map[string]any
	var triggers []map[string]any
	s.db.Where("rule_id=? AND deleted_at IS NULL", id).Order("sort_order,id").Find(&cs)
	s.db.Where("rule_id=? AND deleted_at IS NULL", id).Order("sort_order,id").Find(&as)
	s.db.Table("game_risk_rule_test_runs").Where("rule_id=? AND deleted_at IS NULL", id).Order("id DESC").Limit(20).Find(&tests)
	s.db.Table("game_user_risk_events").Select("id,event_no,user_id,risk_type,risk_level,trigger_value,status,created_at").Where("rule_id=? AND deleted_at IS NULL", id).Order("id DESC").Limit(100).Find(&triggers)
	logs, _ := operationLogs(s.db, "risk_rule", id)
	return map[string]any{"basicInformation": row, "conditions": cs, "actions": as, "scope": row, "testResults": tests, "triggerRecords": triggers, "operationLogs": logs}, nil
}

func validateRule(in RuleInput) error {
	if !valid(in.RuleType, "REGISTRATION", "LOGIN", "DEVICE", "IP", "DEPOSIT", "WITHDRAWAL", "BETTING", "BONUS", "PROVIDER_CALLBACK", "MANUAL") || !valid(in.RiskLevel, "LOW", "MEDIUM", "HIGH", "CRITICAL") {
		return newError("INVALID_RISK_RULE", "invalid rule type or risk level")
	}
	if in.StartTime != nil && in.EndTime != nil && in.EndTime.Before(*in.StartTime) {
		return newError("INVALID_TIME_RANGE", "end time must be after start time")
	}
	for _, c := range in.Conditions {
		if !operators[c.Operator] {
			return newError("INVALID_RULE_OPERATOR", "unsupported rule operator")
		}
	}
	for _, a := range in.Actions {
		if !riskActions[a.ActionType] {
			return newError("INVALID_RULE_ACTION", "unsupported risk action")
		}
	}
	return nil
}
func (s *RiskRuleService) Save(id uint64, in RuleInput, actor Actor) (uint64, error) {
	if err := validateRule(in); err != nil {
		return 0, err
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var before ruleRecord
		if id > 0 {
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&before, id).Error; err != nil {
				return mapError(err, "RISK_RULE_NOT_FOUND")
			}
		}
		row := ruleRecord{ID: id, Code: strings.TrimSpace(in.Code), NameZh: in.NameZh, NameEn: in.NameEn, DescriptionZh: in.DescriptionZh, DescriptionEn: in.DescriptionEn, RuleType: in.RuleType, RiskLevel: in.RiskLevel, Priority: in.Priority, Enabled: in.Enabled, Countries: in.Countries, Currencies: in.Currencies, Channels: in.Channels, StartAt: in.StartTime, EndAt: in.EndTime, TriggerCount: before.TriggerCount, LastTriggeredAt: before.LastTriggeredAt}
		if id == 0 {
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			id = row.ID
		} else if err := tx.Model(&ruleRecord{}).Where("id=?", id).Select("code", "name_zh", "name_en", "description_zh", "description_en", "rule_type", "risk_level", "priority", "enabled", "countries", "currencies", "channels", "start_at", "end_at").Updates(row).Error; err != nil {
			return err
		}
		if err := tx.Where("rule_id=?", id).Delete(&conditionRecord{}).Error; err != nil {
			return err
		}
		if err := tx.Where("rule_id=?", id).Delete(&actionRecord{}).Error; err != nil {
			return err
		}
		for _, c := range in.Conditions {
			if err := tx.Create(&conditionRecord{RuleID: id, FieldName: c.FieldName, Operator: c.Operator, ComparisonValue: c.ComparisonValue, WindowSeconds: c.WindowSeconds, SortOrder: c.SortOrder}).Error; err != nil {
				return err
			}
		}
		for _, a := range in.Actions {
			if err := tx.Create(&actionRecord{RuleID: id, ActionType: a.ActionType, ParameterValue: a.ParameterValue, SortOrder: a.SortOrder}).Error; err != nil {
				return err
			}
		}
		permission := "risk:rule:create"
		action := "risk_rule.create"
		if before.ID > 0 {
			permission = "risk:rule:update"
			action = "risk_rule.update"
		}
		return audit(tx, actor, permission, action, "risk_rule", id, 0, before, row, "process structured risk rule")
	})
	return id, err
}
func (s *RiskRuleService) Copy(id uint64, actor Actor) (uint64, error) {
	d, err := s.Detail(id)
	if err != nil {
		return 0, err
	}
	base := d["basicInformation"].(ruleRecord)
	cs := d["conditions"].([]conditionRecord)
	as := d["actions"].([]actionRecord)
	in := RuleInput{Code: fmt.Sprintf("%s-COPY-%d", base.Code, time.Now().Unix()), NameZh: base.NameZh, NameEn: base.NameEn, DescriptionZh: base.DescriptionZh, DescriptionEn: base.DescriptionEn, RuleType: base.RuleType, RiskLevel: base.RiskLevel, Priority: base.Priority, Enabled: false, Countries: base.Countries, Currencies: base.Currencies, Channels: base.Channels, StartTime: base.StartAt, EndTime: base.EndAt}
	for _, c := range cs {
		in.Conditions = append(in.Conditions, ConditionInput{FieldName: c.FieldName, Operator: c.Operator, ComparisonValue: c.ComparisonValue, WindowSeconds: c.WindowSeconds, SortOrder: c.SortOrder})
	}
	for _, a := range as {
		in.Actions = append(in.Actions, ActionInput{ActionType: a.ActionType, ParameterValue: a.ParameterValue, SortOrder: a.SortOrder})
	}
	return s.Save(0, in, actor)
}
func (s *RiskRuleService) Status(id uint64, cmd RuleStatusCommand, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row ruleRecord
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, id).Error; err != nil {
			return mapError(err, "RISK_RULE_NOT_FOUND")
		}
		if row.Enabled == cmd.Enabled {
			return newError("RISK_RULE_STATUS_UNCHANGED", "rule already has requested status")
		}
		if err := tx.Model(&row).Update("enabled", cmd.Enabled).Error; err != nil {
			return err
		}
		return audit(tx, actor, "risk:rule:status", "risk_rule.status", "risk_rule", id, 0, map[string]any{"enabled": row.Enabled}, map[string]any{"enabled": cmd.Enabled}, cmd.Reason)
	})
}
func (s *RiskRuleService) Delete(id uint64, reason string, actor Actor) error {
	if strings.TrimSpace(reason) == "" {
		return newError("REASON_REQUIRED", "reason is required")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row ruleRecord
		if err := tx.First(&row, id).Error; err != nil {
			return mapError(err, "RISK_RULE_NOT_FOUND")
		}
		var n int64
		tx.Table("game_user_risk_events").Where("rule_id=? AND deleted_at IS NULL", id).Count(&n)
		if n > 0 {
			return newError("RISK_RULE_IN_USE", "rule has trigger records")
		}
		now := time.Now()
		if err := tx.Model(&row).Update("deleted_at", now).Error; err != nil {
			return err
		}
		return audit(tx, actor, "risk:rule:delete", "risk_rule.delete", "risk_rule", id, 0, row, nil, reason)
	})
}
func (s *RiskRuleService) Test(id uint64, cmd RuleTestCommand, actor Actor) (map[string]any, error) {
	var cs []conditionRecord
	var as []actionRecord
	if err := s.db.Where("rule_id=? AND deleted_at IS NULL", id).Order("sort_order,id").Find(&cs).Error; err != nil {
		return nil, err
	}
	if len(cs) == 0 {
		return nil, newError("RISK_RULE_NOT_FOUND", "rule has no conditions")
	}
	s.db.Where("rule_id=? AND deleted_at IS NULL", id).Order("sort_order,id").Find(&as)
	result, matched := s.evaluation.Evaluate(cs, as, cmd.Facts)
	facts, _ := json.Marshal(cmd.Facts)
	payload, _ := json.Marshal(result)
	err := s.db.Exec("INSERT INTO game_risk_rule_test_runs(rule_id,input_facts,result,matched,operator_id,operator_username,created_at,updated_at) VALUES(?,?,?,?,?,?,NOW(3),NOW(3))", id, string(facts), string(payload), matched, actor.ID, actor.Username).Error
	return result, err
}

func (s *RiskEvaluationService) Evaluate(cs []conditionRecord, as []actionRecord, facts map[string]any) (map[string]any, bool) {
	results := make([]map[string]any, 0, len(cs))
	matched := true
	for _, c := range cs {
		actual, ok := facts[c.FieldName]
		pass := ok && compare(actual, c.Operator, c.ComparisonValue)
		if !pass {
			matched = false
		}
		results = append(results, map[string]any{"fieldName": c.FieldName, "operator": c.Operator, "expected": c.ComparisonValue, "actual": actual, "matched": pass})
	}
	actions := []string{}
	if matched {
		for _, a := range as {
			actions = append(actions, a.ActionType)
		}
	}
	return map[string]any{"matched": matched, "conditions": results, "actions": actions}, matched
}
func compare(actual any, op, expected string) bool {
	a := fmt.Sprint(actual)
	switch op {
	case "EQUAL":
		return a == expected
	case "NOT_EQUAL":
		return a != expected
	case "CONTAINS":
		return strings.Contains(a, expected)
	case "IN", "NOT_IN":
		found := false
		for _, v := range strings.Split(expected, ",") {
			if strings.TrimSpace(v) == a {
				found = true
			}
		}
		if op == "NOT_IN" {
			return !found
		}
		return found
	}
	actualDecimal, ae := decimal.NewFromString(a)
	expectedDecimal, ee := decimal.NewFromString(expected)
	if ae != nil || ee != nil {
		return false
	}
	comparison := actualDecimal.Cmp(expectedDecimal)
	switch op {
	case "GREATER_THAN":
		return comparison > 0
	case "GREATER_THAN_OR_EQUAL", "COUNT_WITHIN", "SUM_WITHIN":
		return comparison >= 0
	case "LESS_THAN":
		return comparison < 0
	case "LESS_THAN_OR_EQUAL":
		return comparison <= 0
	}
	return false
}
