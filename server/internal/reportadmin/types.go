package reportadmin

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"casino/internal/tenancy"

	"github.com/shopspring/decimal"
)

type Actor struct {
	ID            uint64
	Username      string
	IP            string
	UserAgent     string
	CorrelationID string
	Scope         tenancy.Scope
}

type Query struct {
	Page     int
	PageSize int
	Values   map[string]string
	Scope    tenancy.Scope
}

type ChartSeries struct {
	Name     string   `json:"name"`
	Currency string   `json:"currency,omitempty"`
	Data     []string `json:"data"`
}

type Chart struct {
	Key        string        `json:"key"`
	Categories []string      `json:"categories"`
	Series     []ChartSeries `json:"series"`
}

type Result struct {
	SummaryByCurrency []map[string]any  `json:"summaryByCurrency"`
	List              any               `json:"list"`
	Total             int               `json:"total"`
	Page              int               `json:"page"`
	PageSize          int               `json:"pageSize"`
	Charts            []Chart           `json:"charts"`
	MetricBasis       map[string]string `json:"metricBasis"`
}

type ServiceError struct {
	Code    string
	Message string
}

func (e *ServiceError) Error() string { return e.Message }
func NewError(code, message string) error {
	return &ServiceError{Code: code, Message: message}
}

type ExportCommand struct {
	ReportType     string            `json:"reportType" binding:"required"`
	Filters        map[string]string `json:"filters"`
	FileFormat     string            `json:"fileFormat" binding:"required"`
	IdempotencyKey string            `json:"idempotencyKey" binding:"required,max=128"`
}

type ExportAction struct {
	Reason string `json:"reason" binding:"required,max=500"`
}

type ExportRow struct {
	ID               uint64     `json:"id"`
	TaskNo           string     `json:"taskID" gorm:"column:task_no"`
	ReportType       string     `json:"reportType" gorm:"column:report_type"`
	FiltersSummary   string     `json:"filtersSummary" gorm:"column:filters_summary"`
	FileFormat       string     `json:"fileFormat" gorm:"column:file_format"`
	RowCount         uint64     `json:"rowCount" gorm:"column:row_count"`
	Status           string     `json:"status"`
	OperatorID       uint64     `json:"operatorID" gorm:"column:operator_id"`
	OperatorUsername string     `json:"operator" gorm:"column:operator_username"`
	FailureReason    string     `json:"failureReason" gorm:"column:failure_reason"`
	CreatedAt        time.Time  `json:"createdAt" gorm:"column:created_at"`
	CompletedAt      *time.Time `json:"completedAt" gorm:"column:completed_at"`
	ExpiresAt        *time.Time `json:"expiresAt" gorm:"column:expires_at"`
}

type MoneyRow struct {
	Currency string
	Values   map[string]decimal.Decimal
}

func val(q Query, key string) string { return strings.TrimSpace(q.Values[key]) }

func normalizedPage(q Query) (int, int) {
	p, ps := q.Page, q.PageSize
	if p < 1 {
		p = 1
	}
	if ps < 1 {
		ps = 20
	}
	if ps > 100 {
		ps = 100
	}
	return p, ps
}

func paginateRows[T any](rows []T, q Query) ([]T, int, int, int) {
	p, ps := normalizedPage(q)
	start := (p - 1) * ps
	if start >= len(rows) {
		return []T{}, len(rows), p, ps
	}
	end := start + ps
	if end > len(rows) {
		end = len(rows)
	}
	return rows[start:end], len(rows), p, ps
}

func decimalString(v decimal.Decimal) string { return v.Round(2).StringFixed(2) }

func percentage(numerator, denominator decimal.Decimal) *string {
	if denominator.IsZero() {
		return nil
	}
	value := numerator.Div(denominator).Mul(decimal.NewFromInt(100)).Round(4).StringFixed(4)
	return &value
}

func parseDecimal(value string) decimal.Decimal {
	d, err := decimal.NewFromString(strings.TrimSpace(value))
	if err != nil {
		return decimal.Zero
	}
	return d
}

func parseID(value string) uint64 {
	id, _ := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	return id
}

func marshal(value any) string {
	b, _ := json.Marshal(value)
	return string(b)
}

func unmarshalMap(value string) map[string]string {
	out := map[string]string{}
	_ = json.Unmarshal([]byte(value), &out)
	return out
}

func reportType(value string) (string, error) {
	v := strings.ToUpper(strings.TrimSpace(value))
	switch v {
	case "FINANCIAL", "GAME", "PROVIDER", "USER", "CHANNEL", "AGENT", "PROMOTION":
		return v, nil
	default:
		return "", NewError("INVALID_REPORT_TYPE", "invalid report type")
	}
}

func fileFormat(value string) (string, error) {
	v := strings.ToUpper(strings.TrimSpace(value))
	if v != "CSV" && v != "XLSX" {
		return "", NewError("INVALID_FILE_FORMAT", "invalid file format")
	}
	return v, nil
}

func granularity(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "week":
		return "week"
	case "month":
		return "month"
	default:
		return "day"
	}
}

func timezoneOffset(value string) (int, string, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		return 0, "+08:00", nil
	}
	if len(v) != 6 || (v[0] != '+' && v[0] != '-') || v[3] != ':' {
		return 0, "", NewError("INVALID_TIMEZONE", "timezone must use an offset such as +08:00")
	}
	hour, e1 := strconv.Atoi(v[1:3])
	minute, e2 := strconv.Atoi(v[4:6])
	if e1 != nil || e2 != nil || hour > 14 || minute > 59 {
		return 0, "", NewError("INVALID_TIMEZONE", "invalid timezone offset")
	}
	offset := hour*60 + minute
	if v[0] == '-' {
		offset = -offset
	}
	return offset - 480, v, nil
}

func periodExpr(column string, q Query) (string, error) {
	delta, _, err := timezoneOffset(val(q, "timezone"))
	if err != nil {
		return "", err
	}
	shifted := fmt.Sprintf("DATE_ADD(%s, INTERVAL %d MINUTE)", column, delta)
	switch granularity(val(q, "granularity")) {
	case "week":
		return fmt.Sprintf("DATE_FORMAT(%s,'%%x-W%%v')", shifted), nil
	case "month":
		return fmt.Sprintf("DATE_FORMAT(%s,'%%Y-%%m')", shifted), nil
	default:
		return fmt.Sprintf("DATE_FORMAT(%s,'%%Y-%%m-%%d')", shifted), nil
	}
}

func validateRange(q Query) error {
	from, to := val(q, "dateFrom"), val(q, "dateTo")
	if from == "" || to == "" {
		return NewError("DATE_RANGE_REQUIRED", "date range is required")
	}
	start, e1 := time.Parse("2006-01-02 15:04:05", from)
	end, e2 := time.Parse("2006-01-02 15:04:05", to)
	if e1 != nil || e2 != nil || end.Before(start) {
		return NewError("INVALID_DATE_RANGE", "invalid date range")
	}
	if end.Sub(start) > 366*24*time.Hour {
		return NewError("DATE_RANGE_TOO_LARGE", "date range cannot exceed 366 days")
	}
	return nil
}

func errorCode(err error) string {
	var serviceErr *ServiceError
	if errors.As(err, &serviceErr) {
		return serviceErr.Code
	}
	return "INTERNAL_ERROR"
}
