package riskadmin

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type blacklistRecord struct {
	ID               uint64     `json:"id"`
	BlacklistNo      string     `json:"blacklistID"`
	BlacklistType    string     `json:"type"`
	ValueHash        string     `json:"-"`
	ValueMasked      string     `json:"valueMasked"`
	ValueLast4       string     `json:"-"`
	Reason           string     `json:"reason"`
	Status           string     `json:"status"`
	CreatedBy        uint64     `json:"createdByID"`
	CreatedByName    string     `json:"createdBy"`
	ExpiresAt        *time.Time `json:"expiresAt"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
	DeletedAt        *time.Time `json:"-"`
	RelatedUserCount int64      `gorm:"-" json:"relatedUserCount"`
}

func (blacklistRecord) TableName() string { return "game_risk_blacklists" }

type blacklistRelation struct {
	ID           uint64
	BlacklistID  uint64
	UserID       uint64
	RelationType string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

func (blacklistRelation) TableName() string { return "game_risk_blacklist_user_relations" }

type blacklistHistory struct {
	ID               uint64     `json:"id"`
	BlacklistID      uint64     `json:"blacklistID"`
	BeforeStatus     string     `json:"beforeStatus"`
	AfterStatus      string     `json:"afterStatus"`
	Reason           string     `json:"reason"`
	OperatorID       uint64     `json:"operatorID"`
	OperatorUsername string     `json:"operator"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
	DeletedAt        *time.Time `json:"-"`
}

func (blacklistHistory) TableName() string { return "game_risk_blacklist_histories" }

type BlacklistService struct{ db *gorm.DB }

func (s *BlacklistService) List(q ListQuery) (PageResult, error) {
	var rows []blacklistRecord
	db := s.db.Model(&blacklistRecord{}).Where("deleted_at IS NULL")
	db = equal(db, "blacklist_type", q.Values["type"])
	if v := strings.TrimSpace(q.Values["value"]); v != "" {
		normalized, err := normalizeBlacklist(q.Values["type"], v)
		if err == nil {
			db = db.Where("value_hash=?", hashBlacklist(q.Values["type"], normalized))
		} else {
			db = db.Where("1=0")
		}
	}
	db = contains(db, "reason", q.Values["reason"])
	db = equal(db, "status", q.Values["status"])
	db = contains(db, "created_by_name", q.Values["createdBy"])
	db = timeRange(db, "created_at", q.Values["createdFrom"], q.Values["createdTo"])
	db = timeRange(db, "expires_at", q.Values["expiresFrom"], q.Values["expiresTo"])
	res, err := paginate(db.Order("id DESC"), q, &rows)
	if err != nil {
		return res, err
	}
	for i := range rows {
		s.db.Table("game_risk_blacklist_user_relations").Where("blacklist_id=? AND deleted_at IS NULL", rows[i].ID).Count(&rows[i].RelatedUserCount)
	}
	res.List = rows
	return res, nil
}
func (s *BlacklistService) Detail(id uint64) (map[string]any, error) {
	var row blacklistRecord
	if err := s.db.Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return nil, mapError(err, "BLACKLIST_NOT_FOUND")
	}
	var users, triggers, history []map[string]any
	s.db.Table("game_risk_blacklist_user_relations r").Select("u.id user_id,u.user_no,u.username,r.relation_type,r.created_at").Joins("JOIN game_users u ON u.id=r.user_id AND u.deleted_at IS NULL").Where("r.blacklist_id=? AND r.deleted_at IS NULL", id).Find(&users)
	s.db.Table("game_user_risk_events").Select("id,event_no,risk_type,risk_level,status,created_at").Where("description LIKE ? AND deleted_at IS NULL", "%"+row.BlacklistNo+"%").Order("id DESC").Limit(100).Find(&triggers)
	s.db.Table("game_risk_blacklist_histories").Where("blacklist_id=? AND deleted_at IS NULL", id).Order("id DESC").Find(&history)
	logs, _ := operationLogs(s.db, "risk_blacklist", id)
	return map[string]any{"overview": row, "relatedUsers": users, "triggerRecords": triggers, "statusHistory": history, "operationLogs": logs}, nil
}
func (s *BlacklistService) Create(in BlacklistInput, actor Actor) (uint64, error) {
	normalized, err := normalizeBlacklist(in.Type, in.Value)
	if err != nil {
		return 0, err
	}
	status := "DISABLED"
	if in.Enabled {
		status = "ENABLED"
	}
	row := blacklistRecord{BlacklistNo: generatedNo("BL"), BlacklistType: in.Type, ValueHash: hashBlacklist(in.Type, normalized), ValueMasked: maskSensitive(normalized), ValueLast4: last4(normalized), Reason: in.Reason, Status: status, CreatedBy: actor.ID, CreatedByName: actor.Username, ExpiresAt: in.ExpiresAt}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		users := relatedUsersForValue(tx, in.Type, normalized)
		for _, userID := range users {
			if err := tx.Create(&blacklistRelation{BlacklistID: row.ID, UserID: userID, RelationType: in.Type}).Error; err != nil {
				return err
			}
		}
		if err := tx.Create(&blacklistHistory{BlacklistID: row.ID, BeforeStatus: "", AfterStatus: status, Reason: in.Reason, OperatorID: actor.ID, OperatorUsername: actor.Username}).Error; err != nil {
			return err
		}
		return audit(tx, actor, "risk:blacklist:create", "risk_blacklist.create", "risk_blacklist", row.ID, 0, nil, map[string]any{"type": row.BlacklistType, "valueMasked": row.ValueMasked, "status": status}, in.Reason)
	})
	return row.ID, err
}
func (s *BlacklistService) Update(id uint64, cmd BlacklistUpdateCommand, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row blacklistRecord
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
			return mapError(err, "BLACKLIST_NOT_FOUND")
		}
		before := row
		status := "DISABLED"
		if cmd.Enabled {
			status = "ENABLED"
		}
		if err := tx.Model(&row).Updates(map[string]any{"status": status, "reason": cmd.Reason, "expires_at": cmd.ExpiresAt}).Error; err != nil {
			return err
		}
		if err := tx.Create(&blacklistHistory{BlacklistID: id, BeforeStatus: row.Status, AfterStatus: status, Reason: cmd.Reason, OperatorID: actor.ID, OperatorUsername: actor.Username}).Error; err != nil {
			return err
		}
		return audit(tx, actor, "risk:blacklist:update", "risk_blacklist.update", "risk_blacklist", id, 0, before, map[string]any{"status": status, "expiresAt": cmd.ExpiresAt}, cmd.Reason)
	})
}
func (s *BlacklistService) Remove(id uint64, reason string, actor Actor) error {
	if strings.TrimSpace(reason) == "" {
		return newError("REASON_REQUIRED", "reason is required")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row blacklistRecord
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
			return mapError(err, "BLACKLIST_NOT_FOUND")
		}
		now := time.Now()
		if err := tx.Model(&row).Updates(map[string]any{"status": "REMOVED", "deleted_at": now}).Error; err != nil {
			return err
		}
		if err := tx.Create(&blacklistHistory{BlacklistID: id, BeforeStatus: row.Status, AfterStatus: "REMOVED", Reason: reason, OperatorID: actor.ID, OperatorUsername: actor.Username}).Error; err != nil {
			return err
		}
		return audit(tx, actor, "risk:blacklist:remove", "risk_blacklist.remove", "risk_blacklist", id, 0, row, map[string]any{"status": "REMOVED"}, reason)
	})
}

func normalizeBlacklist(kind, value string) (string, error) {
	kind = strings.ToUpper(strings.TrimSpace(kind))
	value = strings.TrimSpace(value)
	if !valid(kind, "USER", "DEVICE", "IP", "EMAIL", "MOBILE", "BANK_ACCOUNT", "WALLET_ADDRESS", "PAYMENT_ACCOUNT") {
		return "", newError("INVALID_BLACKLIST_TYPE", "unsupported blacklist type")
	}
	switch kind {
	case "EMAIL":
		value = strings.ToLower(value)
		if !strings.Contains(value, "@") {
			return "", newError("INVALID_BLACKLIST_VALUE", "invalid email")
		}
	case "MOBILE":
		value = regexp.MustCompile(`\D`).ReplaceAllString(value, "")
	case "IP":
		parsed := net.ParseIP(value)
		if parsed == nil {
			return "", newError("INVALID_BLACKLIST_VALUE", "invalid IP address")
		}
		value = parsed.String()
	case "BANK_ACCOUNT", "WALLET_ADDRESS", "PAYMENT_ACCOUNT":
		value = strings.ToUpper(regexp.MustCompile(`\s+`).ReplaceAllString(value, ""))
	case "USER", "DEVICE":
		value = strings.ToLower(value)
	}
	if value == "" {
		return "", newError("INVALID_BLACKLIST_VALUE", "blacklist value is required")
	}
	return value, nil
}
func hashBlacklist(kind, value string) string {
	sum := sha256.Sum256([]byte(strings.ToUpper(kind) + ":" + value))
	return hex.EncodeToString(sum[:])
}
func maskSensitive(value string) string { return mask(value) }
func last4(value string) string {
	r := []rune(value)
	if len(r) <= 4 {
		return value
	}
	return string(r[len(r)-4:])
}
func relatedUsersForValue(tx *gorm.DB, kind, value string) []uint64 {
	var ids []uint64
	switch kind {
	case "USER":
		tx.Table("game_users").Where("(LOWER(user_no)=? OR LOWER(username)=?) AND deleted_at IS NULL", value, value).Pluck("id", &ids)
	case "DEVICE":
		tx.Table("game_user_devices").Where("LOWER(device_id)=? AND deleted_at IS NULL", value).Distinct().Pluck("user_id", &ids)
	case "IP":
		tx.Table("game_users").Where("(registered_ip=? OR last_login_ip=?) AND deleted_at IS NULL", value, value).Pluck("id", &ids)
	case "EMAIL":
		tx.Table("game_users").Where("LOWER(email)=? AND deleted_at IS NULL", value).Pluck("id", &ids)
	case "MOBILE":
		tx.Table("game_users").Where("phone=? AND deleted_at IS NULL", value).Pluck("id", &ids)
	case "BANK_ACCOUNT", "WALLET_ADDRESS", "PAYMENT_ACCOUNT":
		tx.Table("game_withdraw_accounts").Where("UPPER(REPLACE(account_no,' ',''))=? AND deleted_at IS NULL", value).Distinct().Pluck("user_id", &ids)
	}
	return ids
}
