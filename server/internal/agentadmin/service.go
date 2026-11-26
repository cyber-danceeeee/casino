package agentadmin

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"casino/internal/content"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

func page(q ListQuery) (int, int, int) {
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
func val(q ListQuery, key string) string { return strings.TrimSpace(q.Values[key]) }
func like(db *gorm.DB, col string, value string) *gorm.DB {
	if value != "" {
		return db.Where(col+" LIKE ?", "%"+value+"%")
	}
	return db
}
func eq(db *gorm.DB, col string, value string) *gorm.DB {
	if value != "" {
		return db.Where(col+" = ?", value)
	}
	return db
}
func scoped(db *gorm.DB, alias string, scope content.Scope) *gorm.DB {
	return content.ApplySite(db, alias, scope)
}
func actorScoped(db *gorm.DB, alias string, actor Actor) *gorm.DB {
	return scoped(db, alias, actor.Scope)
}
func sqlScope(scope content.Scope, alias string) (string, []any) {
	if !scope.SiteScoped() {
		return "", nil
	}
	prefix := ""
	if strings.TrimSpace(alias) != "" {
		prefix = alias + "."
	}
	clause := ""
	args := []any{}
	if scope.SiteScoped() {
		clause += " AND " + prefix + "site_id IN (0, ?)"
		args = append(args, scope.SiteID)
	}
	return clause, args
}
func rangeTime(db *gorm.DB, col string, from string, to string) *gorm.DB {
	if from != "" {
		db = db.Where(col+">=?", from)
	}
	if to != "" {
		db = db.Where(col+"<=?", to)
	}
	return db
}
func parseID(v string) uint64 { n, _ := strconv.ParseUint(strings.TrimSpace(v), 10, 64); return n }
func money(v string) (decimal.Decimal, error) {
	d, e := decimal.NewFromString(strings.TrimSpace(v))
	if e != nil || d.IsNegative() {
		return decimal.Zero, NewError("INVALID_AMOUNT", "invalid amount")
	}
	return d.Round(2), nil
}
func jsonString(v any) string { b, _ := json.Marshal(v); return string(b) }
func jsonValue(v string) any {
	var out any
	if json.Unmarshal([]byte(v), &out) == nil {
		return out
	}
	return v
}
func nowNo(prefix string) string { return fmt.Sprintf("%s%d", prefix, time.Now().UnixNano()) }

func lastInsertID(tx *gorm.DB) (uint64, error) {
	var id uint64
	if err := tx.Raw("SELECT LAST_INSERT_ID()").Scan(&id).Error; err != nil {
		return 0, err
	}
	if id == 0 {
		return 0, errors.New("database did not return an inserted id")
	}
	return id, nil
}

func sameDate(left time.Time, right time.Time) bool {
	ly, lm, ld := left.Date()
	ry, rm, rd := right.Date()
	return ly == ry && lm == rm && ld == rd
}

func databaseBool(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case int64:
		return typed != 0
	case uint64:
		return typed != 0
	case []byte:
		return string(typed) == "1" || strings.EqualFold(string(typed), "true")
	default:
		text := fmt.Sprint(value)
		return text == "1" || strings.EqualFold(text, "true")
	}
}

func paginate(db *gorm.DB, q ListQuery, out any) (PageResult, error) {
	p, ps, off := page(q)
	value := reflect.ValueOf(out)
	if value.Kind() == reflect.Pointer && value.Elem().Kind() == reflect.Slice && value.Elem().IsNil() {
		value.Elem().Set(reflect.MakeSlice(value.Elem().Type(), 0, 0))
	}
	var total int64
	if e := db.Count(&total).Error; e != nil {
		return PageResult{}, e
	}
	if e := db.Offset(off).Limit(ps).Scan(out).Error; e != nil {
		return PageResult{}, e
	}
	return PageResult{List: out, Total: total, Page: p, PageSize: ps}, nil
}

func audit(tx *gorm.DB, actor Actor, permission string, action string, target string, targetID uint64, before any, after any, reason string, amount *decimal.Decimal, currency string) error {
	row := map[string]any{"site_id": actor.Scope.SiteID, "operator_id": actor.ID, "operator_username": actor.Username, "permission": permission, "action": action, "target_type": target, "target_id": targetID, "before_state": jsonString(before), "after_state": jsonString(after), "reason": reason, "ip_address": actor.IP, "user_agent": actor.UserAgent, "correlation_id": actor.CorrelationID, "currency": currency, "created_at": time.Now(), "updated_at": time.Now()}
	if amount != nil {
		row["amount"] = amount
	}
	return tx.Table("game_admin_operation_logs").Create(row).Error
}
func notFound(err error, code string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NewError(code, "record not found")
	}
	return err
}
func requireReason(reason string) error {
	if strings.TrimSpace(reason) == "" {
		return NewError("REASON_REQUIRED", "operator reason is required")
	}
	return nil
}

type agentRow struct {
	ID                  uint64          `json:"id"`
	AgentCode           string          `json:"agentCode"`
	Name                string          `json:"name"`
	UserID              uint64          `json:"userId"`
	UserNo              string          `json:"userNo"`
	Username            string          `json:"username"`
	ParentAgentID       uint64          `json:"parentAgentId"`
	ParentAgent         string          `json:"parentAgent"`
	Level               uint            `json:"level"`
	Country             string          `json:"country"`
	Currency            string          `json:"currency"`
	Status              string          `json:"status"`
	CommissionPlanID    uint64          `json:"commissionPlanId"`
	CommissionPlan      string          `json:"commissionPlan"`
	DirectUsers         int64           `json:"directUsers"`
	DirectAgents        int64           `json:"directAgents"`
	TotalUsers          int64           `json:"totalUsers"`
	AvailableCommission decimal.Decimal `json:"availableCommission"`
	FrozenCommission    decimal.Decimal `json:"frozenCommission"`
	CreatedAt           time.Time       `json:"createdAt"`
	UpdatedAt           time.Time       `json:"updatedAt"`
}
type channelRow struct {
	ID                 uint64          `json:"id"`
	Code               string          `json:"code"`
	NameZh             string          `json:"nameZh"`
	NameEn             string          `json:"nameEn"`
	AgentID            uint64          `json:"agentId"`
	AgentName          string          `json:"agentName"`
	Country            string          `json:"country"`
	Platform           string          `json:"platform"`
	RegistrationURL    string          `json:"registrationUrl"`
	TrackingParameters string          `json:"-"`
	Enabled            bool            `json:"enabled"`
	RegisteredUsers    int64           `json:"registeredUsers"`
	FirstDepositUsers  int64           `json:"firstDepositUsers"`
	DepositAmount      decimal.Decimal `json:"depositAmount"`
	CreatedAt          time.Time       `json:"createdAt"`
	UpdatedAt          time.Time       `json:"updatedAt"`
}
type planRow struct {
	ID                      uint64          `json:"id"`
	Code                    string          `json:"code"`
	NameZh                  string          `json:"nameZh"`
	NameEn                  string          `json:"nameEn"`
	Model                   string          `json:"model"`
	Currency                string          `json:"currency"`
	CPAAmount               decimal.Decimal `json:"cpaAmount"`
	RevenueShareRate        decimal.Decimal `json:"revenueShareRate"`
	NegativeCarryover       bool            `json:"negativeCarryover"`
	MinimumSettlementAmount decimal.Decimal `json:"minimumSettlementAmount"`
	SettlementCycle         string          `json:"settlementCycle"`
	DeductionRules          string          `json:"-"`
	TierRules               string          `json:"-"`
	EligibleProducts        string          `json:"-"`
	Enabled                 bool            `json:"enabled"`
	AssignedAgentCount      int64           `json:"assignedAgentCount"`
	CreatedAt               time.Time       `json:"createdAt"`
	UpdatedAt               time.Time       `json:"updatedAt"`
}
