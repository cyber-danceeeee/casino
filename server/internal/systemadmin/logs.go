package systemadmin

import (
	"encoding/json"
	"strings"
	"time"
)

type AuditLogRow struct {
	ID         uint64 `json:"logID"`
	OperatorID uint64 `json:"operatorID" gorm:"column:operator_id"`

	SiteID             uint64          `json:"siteID" gorm:"column:site_id"`
	Operator           string          `json:"operator" gorm:"column:operator_username"`
	Permission         string          `json:"permission"`
	Module             string          `json:"module"`
	Action             string          `json:"action"`
	Result             string          `json:"result"`
	TargetType         string          `json:"targetType" gorm:"column:target_type"`
	TargetID           uint64          `json:"targetID" gorm:"column:target_id"`
	BeforeState        json.RawMessage `json:"beforeState" gorm:"column:before_state;serializer:json"`
	AfterState         json.RawMessage `json:"afterState" gorm:"column:after_state;serializer:json"`
	Reason             string          `json:"reason"`
	IP                 string          `json:"ip" gorm:"column:ip_address"`
	UserAgent          string          `json:"userAgent" gorm:"column:user_agent"`
	CorrelationID      string          `json:"correlationID" gorm:"column:correlation_id"`
	RequestInformation json.RawMessage `json:"requestInformation" gorm:"column:request_info;serializer:json"`
	ErrorInformation   json.RawMessage `json:"errorInformation" gorm:"column:error_info;serializer:json"`
	CreatedAt          time.Time       `json:"createdAt" gorm:"column:created_at"`
}

func (s *Service) ListAuditLogs(q Query) (Result, error) {
	rows := []AuditLogRow{}
	moduleExpr := "COALESCE(NULLIF(module,''),NULLIF(SUBSTRING_INDEX(permission,':',1),''),'unknown')"
	if s.db.Dialector.Name() == "sqlite" {
		moduleExpr = "COALESCE(NULLIF(module,''),NULLIF(CASE WHEN INSTR(permission,':')>0 THEN SUBSTR(permission,1,INSTR(permission,':')-1) ELSE permission END,''),'unknown')"
	}
	db := s.db.Table("game_admin_operation_logs").Select("id, operator_id, site_id, operator_username, permission," + moduleExpr + " module,action,result,target_type,target_id,reason,ip_address,correlation_id,created_at").Where("deleted_at IS NULL")

	db = exact(db, "site_id", value(q, "siteID"))
	db = like(db, "operator_username", value(q, "operator"))
	if module := value(q, "module"); module != "" {
		db = db.Where(moduleExpr+" = ?", module)
	}
	db = like(db, "action", value(q, "action"))
	db = exact(db, "target_type", value(q, "targetType"))
	db = exact(db, "result", value(q, "result"))
	db = like(db, "ip_address", value(q, "ip"))
	db = exact(db, "correlation_id", value(q, "correlationID"))
	db = timeRange(db, "created_at", value(q, "createdFrom"), value(q, "createdTo"))
	return paginate(db.Order("created_at DESC,id DESC"), q, &rows)
}
func (s *Service) AuditLogDetail(id uint64) (map[string]any, error) {
	var row AuditLogRow
	if err := s.db.Table("game_admin_operation_logs").Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return nil, rowNotFound("audit log")
	}
	if strings.TrimSpace(row.Module) == "" {
		row.Module = strings.Split(strings.TrimSpace(row.Permission), ":")[0]
		if row.Module == "" {
			row.Module = "unknown"
		}
	}
	return map[string]any{
		"overview":    map[string]any{"logID": row.ID, "operator": row.Operator, "permission": row.Permission, "module": row.Module, "action": row.Action, "result": row.Result, "targetType": row.TargetType, "reason": row.Reason, "ip": row.IP, "correlationID": row.CorrelationID, "createdAt": row.CreatedAt},
		"beforeState": redactRaw(row.BeforeState), "afterState": redactRaw(row.AfterState),
		"requestInformation": redactRaw(row.RequestInformation), "errorInformation": redactRaw(row.ErrorInformation),
	}, nil
}
func redactRaw(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	var value any
	if json.Unmarshal(raw, &value) != nil {
		return nil
	}
	return redactValue(value)
}

type LoginLogRow struct {
	ID              uint64    `json:"logID"`
	AdminID         *uint64   `json:"administratorID" gorm:"column:admin_id"`
	Administrator   string    `json:"administrator" gorm:"column:username"`
	Success         bool      `json:"success"`
	LoginResult     string    `json:"loginResult" gorm:"-"`
	FailureReason   string    `json:"failureReason" gorm:"column:failure_reason"`
	IP              string    `json:"ip" gorm:"column:ip_address"`
	Country         string    `json:"country"`
	Device          string    `json:"device"`
	OperatingSystem string    `json:"operatingSystem" gorm:"column:operating_system"`
	Browser         string    `json:"browser"`
	UserAgent       string    `json:"userAgent" gorm:"column:user_agent"`
	CorrelationID   string    `json:"correlationID" gorm:"column:correlation_id"`
	CreatedAt       time.Time `json:"createdAt" gorm:"column:created_at"`
}

func (s *Service) ListLoginLogs(q Query) (Result, error) {
	rows := []LoginLogRow{}
	db := s.db.Table("game_admin_login_logs").Where("deleted_at IS NULL")
	admin := value(q, "administrator")
	if admin != "" {
		db = db.Where("(username LIKE ? OR admin_id=?)", "%"+admin+"%", parseID(admin))
	}
	switch strings.ToUpper(value(q, "loginResult")) {
	case "SUCCESS", "1", "TRUE":
		db = db.Where("success=?", true)
	case "FAILED", "0", "FALSE":
		db = db.Where("success=?", false)
	}
	db = like(db, "ip_address", value(q, "ip"))
	db = exact(db, "country", value(q, "country"))
	db = like(db, "device", value(q, "device"))
	db = like(db, "browser", value(q, "browser"))
	db = timeRange(db, "created_at", value(q, "createdFrom"), value(q, "createdTo"))
	result, err := paginate(db.Order("created_at DESC,id DESC"), q, &rows)
	for i := range rows {
		rows[i].LoginResult = "FAILED"
		if rows[i].Success {
			rows[i].LoginResult = "SUCCESS"
		}
	}
	result.List = rows
	return result, err
}
func (s *Service) LoginLogDetail(id uint64) (map[string]any, error) {
	var row LoginLogRow
	if err := s.db.Table("game_admin_login_logs").Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return nil, rowNotFound("login log")
	}
	row.LoginResult = "FAILED"
	if row.Success {
		row.LoginResult = "SUCCESS"
	}
	return map[string]any{"overview": row, "administrator": map[string]any{"id": row.AdminID, "username": row.Administrator}, "requestInformation": map[string]any{"ip": row.IP, "userAgent": row.UserAgent, "correlationID": row.CorrelationID}}, nil
}
