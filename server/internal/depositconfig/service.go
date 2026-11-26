package depositconfig

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"casino/internal/model"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

func domainError(code, message string) error { return &DomainError{Code: code, Message: message} }

func normalizeQuery(q ListQuery) (ListQuery, int) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = 20
	}
	if q.PageSize > 100 {
		q.PageSize = 100
	}
	if q.Values == nil {
		q.Values = map[string]string{}
	}
	for key, value := range q.Values {
		q.Values[key] = strings.TrimSpace(value)
	}
	return q, (q.Page - 1) * q.PageSize
}

func (s *Service) paginate(db *gorm.DB, q ListQuery, out any) (PageResult, error) {
	q, offset := normalizeQuery(q)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return PageResult{}, err
	}
	if err := db.Offset(offset).Limit(q.PageSize).Find(out).Error; err != nil {
		return PageResult{}, err
	}
	return PageResult{List: out, Total: total, Page: q.Page, PageSize: q.PageSize}, nil
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

func money(value string, field string) (decimal.Decimal, error) {
	parsed, err := decimal.NewFromString(strings.TrimSpace(value))
	if err != nil || parsed.IsNegative() {
		return decimal.Zero, domainError("INVALID_AMOUNT", field+" must be a non-negative exact decimal")
	}
	return parsed, nil
}

func moneyPtr(value decimal.Decimal) *decimal.Decimal { return &value }

func validRange(min, max decimal.Decimal) error {
	if max.LessThan(min) {
		return domainError("INVALID_AMOUNT_RANGE", "maximum amount must be greater than or equal to minimum amount")
	}
	return nil
}

func validStatus(status int) error {
	if status != 1 && status != 2 {
		return domainError("INVALID_STATUS", "status must be enabled or disabled")
	}
	return nil
}

func validCode(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return domainError("INVALID_CODE", "code is required")
	}
	for _, r := range value {
		if !(r == '-' || r == '_' || r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z') {
			return domainError("INVALID_CODE", "code contains unsupported characters")
		}
	}
	return nil
}

func decimalString(value *decimal.Decimal) string {
	if value == nil {
		return "0.00"
	}
	return value.StringFixed(2)
}

func rewardTotal(bonus, cash *decimal.Decimal) decimal.Decimal {
	total := decimal.Zero
	if bonus != nil {
		total = total.Add(*bonus)
	}
	if cash != nil {
		total = total.Add(*cash)
	}
	return total
}

func jsonState(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func (s *Service) audit(tx *gorm.DB, actor Actor, permission, action, targetType string, targetID uint64, before, after any) error {
	return tx.Create(&model.GameAdminOperationLogs{
		TenantID: actor.Scope.TenantID, SiteID: actor.Scope.SiteID,
		OperatorID: actor.ID, OperatorUsername: actor.Username, Permission: permission,
		Action: action, TargetType: targetType, TargetID: targetID,
		BeforeState: jsonState(before), AfterState: jsonState(after),
		Reason: "configuration update", IPAddress: actor.IP, UserAgent: actor.UserAgent,
		CorrelationID: actor.CorrelationID,
	}).Error
}

func notFound(err error, code string) error {
	if err == gorm.ErrRecordNotFound {
		return domainError(code, strings.ToLower(strings.ReplaceAll(code, "_", " ")))
	}
	return err
}

func listContains(list, value string) bool {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return true
	}
	list = strings.TrimSpace(list)
	if list == "" {
		return true
	}
	for _, item := range strings.FieldsFunc(list, func(r rune) bool { return r == ',' || r == ';' || r == '\n' }) {
		item = strings.ToUpper(strings.TrimSpace(item))
		if item == "ALL" || item == value {
			return true
		}
	}
	return false
}

func amountInRange(amount decimal.Decimal, min, max *decimal.Decimal) bool {
	if min != nil && amount.LessThan(*min) {
		return false
	}
	return max == nil || max.IsZero() || !amount.GreaterThan(*max)
}

func generatedTitle(titleZh, titleEn string) string {
	if strings.TrimSpace(titleEn) != "" {
		return strings.TrimSpace(titleEn)
	}
	return strings.TrimSpace(titleZh)
}

func duplicateCodeError(entity string) error {
	return domainError("DUPLICATE_CODE", fmt.Sprintf("%s code already exists", entity))
}
