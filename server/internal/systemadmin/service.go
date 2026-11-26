package systemadmin

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service { return &Service{db: db} }

func paginate(db *gorm.DB, q Query, out any) (Result, error) {
	p, size := page(q)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return Result{}, err
	}
	if err := db.Offset((p - 1) * size).Limit(size).Find(out).Error; err != nil {
		return Result{}, err
	}
	return Result{List: out, Total: total, Page: p, PageSize: size}, nil
}

func exact(db *gorm.DB, column, v string) *gorm.DB {
	if v != "" {
		return db.Where(column+" = ?", v)
	}
	return db
}
func like(db *gorm.DB, column, v string) *gorm.DB {
	if v != "" {
		return db.Where(column+" LIKE ?", "%"+v+"%")
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

func audit(tx *gorm.DB, actor Actor, permission, action, target string, targetID uint64, before, after any, reason string) error {
	return tx.Table("game_admin_operation_logs").Create(map[string]any{
		"operator_id": actor.ID, "tenant_id": actor.TenantID, "site_id": actor.SiteID, "operator_username": actor.Username, "permission": permission, "module": "system",
		"action": action, "result": "SUCCESS", "target_type": target, "target_id": targetID,
		"before_state": marshal(redact(before)), "after_state": marshal(redact(after)), "reason": strings.TrimSpace(reason),
		"ip_address": actor.IP, "user_agent": actor.UserAgent, "correlation_id": actor.CorrelationID,
		"currency": "", "created_at": now(), "updated_at": now(),
	}).Error
}

func marshal(value any) string {
	data, _ := json.Marshal(value)
	return string(data)
}

var sensitiveKeyPattern = regexp.MustCompile(`(?i)(password|secret|token|signature|private.?key|authorization|credential)`)

func redact(value any) any {
	data, _ := json.Marshal(value)
	var normalized any
	if json.Unmarshal(data, &normalized) != nil {
		return nil
	}
	return redactValue(normalized)
}
func redactValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			if sensitiveKeyPattern.MatchString(key) {
				out[key] = "******"
			} else {
				out[key] = redactValue(item)
			}
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for index, item := range typed {
			out[index] = redactValue(item)
		}
		return out
	default:
		return value
	}
}

func validateCode(value string) error {
	matched, _ := regexp.MatchString(`^[A-Za-z][A-Za-z0-9:_-]{1,127}$`, strings.TrimSpace(value))
	if !matched {
		return NewError("INVALID_CODE", "invalid code")
	}
	return nil
}
func validateSafeURL(value string) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parsed, err := url.Parse(value)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
		return NewError("UNSAFE_URL", "only http and https URLs are allowed")
	}
	return nil
}
func requireReason(reason string) error {
	if strings.TrimSpace(reason) == "" {
		return NewError("REASON_REQUIRED", "reason is required")
	}
	return nil
}
func parseID(value string) uint64 {
	id, _ := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	return id
}
func uniqueIDs(ids []uint64) []uint64 {
	seen := map[uint64]bool{}
	out := make([]uint64, 0, len(ids))
	for _, id := range ids {
		if id > 0 && !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
func duplicateError(err error) error {
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "duplicate") {
		return NewError("DUPLICATE_VALUE", "value already exists")
	}
	return err
}
func rowNotFound(kind string) error {
	return NewError(strings.ToUpper(kind)+"_NOT_FOUND", fmt.Sprintf("%s not found", kind))
}
