package riskadmin

import (
	"fmt"
	"strings"
	"time"

	"casino/internal/content"

	"gorm.io/gorm"
)

type associationRow struct {
	UserID                 uint64    `json:"userID"`
	UserNo                 string    `json:"userNo"`
	Username               string    `json:"username"`
	AssociationType        string    `json:"associationType"`
	AssociationValue       string    `json:"-"`
	AssociationValueMasked string    `gorm:"-" json:"associationValueMasked"`
	RelatedAccountCount    int64     `json:"relatedAccountCount"`
	RiskLevel              string    `json:"riskLevel"`
	FirstSeenAt            time.Time `json:"firstSeenAt"`
	LastSeenAt             time.Time `json:"lastSeenAt"`
}
type UserAssociationService struct{ db *gorm.DB }

func associationUnion(db *gorm.DB) *gorm.DB {
	return db.Raw(`
SELECT d.user_id,'DEVICE' association_type,CAST(d.device_id AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_general_ci association_value,MIN(d.created_at) first_seen_at,MAX(d.updated_at) last_seen_at FROM game_user_devices d WHERE d.deleted_at IS NULL AND d.device_id<>'' GROUP BY d.user_id,d.device_id
UNION ALL SELECT d.user_id,'FINGERPRINT',CAST(d.fingerprint_hash AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_general_ci,MIN(d.created_at),MAX(d.updated_at) FROM game_user_devices d WHERE d.deleted_at IS NULL AND d.fingerprint_hash<>'' GROUP BY d.user_id,d.fingerprint_hash
UNION ALL SELECT u.id,'IP',CAST(u.registered_ip AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_general_ci,u.created_at,u.updated_at FROM game_users u WHERE u.deleted_at IS NULL AND u.registered_ip<>''
UNION ALL SELECT u.id,'IP',CAST(u.last_login_ip AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_general_ci,u.created_at,u.updated_at FROM game_users u WHERE u.deleted_at IS NULL AND u.last_login_ip<>''
UNION ALL SELECT w.user_id,'WITHDRAWAL_ACCOUNT',CAST(w.account_no AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_general_ci,MIN(w.created_at),MAX(w.updated_at) FROM game_withdraw_accounts w WHERE w.deleted_at IS NULL AND w.account_no<>'' GROUP BY w.user_id,w.account_no
UNION ALL SELECT u.id,'REFERRAL',CAST(u.referrer_user_id AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_general_ci,u.created_at,u.updated_at FROM game_users u WHERE u.deleted_at IS NULL AND u.referrer_user_id>0
UNION ALL SELECT u.id,'EMAIL',CAST(LOWER(u.email) AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_general_ci,u.created_at,u.updated_at FROM game_users u WHERE u.deleted_at IS NULL AND u.email<>''
UNION ALL SELECT u.id,'MOBILE',CAST(u.phone AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_general_ci,u.created_at,u.updated_at FROM game_users u WHERE u.deleted_at IS NULL AND u.phone<>''`)
}
func (s *UserAssociationService) List(q ListQuery) (PageResult, error) {
	sub := associationUnion(s.db)
	var rows []associationRow
	relatedSQL := "(SELECT COUNT(DISTINCT a2.user_id) FROM (?) a2 JOIN game_users u2 ON u2.id=a2.user_id AND u2.deleted_at IS NULL WHERE a2.association_type=a.association_type AND a2.association_value=a.association_value"
	selectArgs := []any{associationUnion(s.db)}
	if q.Scope.SiteScoped() {
		if q.Scope.SiteScoped() {
			relatedSQL += " AND u2.site_id=?"
			selectArgs = append(selectArgs, q.Scope.SiteID)
		}
	}
	relatedSQL += ")"
	db := s.db.Table("(?) a", sub).
		Select("a.*,u.user_no,u.username,COALESCE(p.risk_level,'LOW') risk_level,"+relatedSQL+" related_account_count", selectArgs...).
		Joins("JOIN game_users u ON u.id=a.user_id AND u.deleted_at IS NULL").
		Joins("LEFT JOIN game_user_risk_profiles p ON p.user_id=a.user_id AND p.deleted_at IS NULL")
	db = content.ApplySiteStrict(db, "u", q.Scope)
	db = equal(db, "a.user_id", q.Values["userID"])
	db = contains(db, "u.username", q.Values["username"])
	db = equal(db, "a.association_type", q.Values["associationType"])
	db = equal(db, "a.association_value", q.Values["deviceID"])
	if q.Values["ip"] != "" {
		db = db.Where("a.association_type='IP' AND a.association_value=?", q.Values["ip"])
	}
	db = equal(db, "p.risk_level", q.Values["riskLevel"])
	if q.Values["relatedUserID"] != "" {
		db = db.Where("EXISTS (SELECT 1 FROM (?) ar WHERE ar.user_id=? AND ar.association_type=a.association_type AND ar.association_value=a.association_value)", associationUnion(s.db), q.Values["relatedUserID"])
	}
	res, err := paginate(db.Order("related_account_count DESC,a.last_seen_at DESC"), q, &rows)
	if err != nil {
		return res, err
	}
	for i := range rows {
		rows[i].AssociationValueMasked = maskSensitive(rows[i].AssociationValue)
	}
	res.List = rows
	return res, nil
}
func (s *UserAssociationService) Detail(userID uint64, scope content.Scope) (map[string]any, error) {
	var user map[string]any
	userQuery := content.ApplySiteStrict(s.db.Table("game_users"), "", scope)
	if err := userQuery.Select("id,user_no,username,country_code,status,risk_review_required").Where("id=? AND deleted_at IS NULL", userID).Take(&user).Error; err != nil {
		return nil, mapError(err, "USER_NOT_FOUND")
	}
	q := ListQuery{Page: 1, PageSize: 100, Scope: scope, Values: map[string]string{"userID": fmt.Sprint(userID)}}
	page, _ := s.List(q)
	associations, _ := page.List.([]associationRow)
	nodes := []map[string]any{{"id": fmt.Sprint(userID), "type": "USER", "label": user["username"]}}
	edges := []map[string]any{}
	relatedUsers := map[uint64]map[string]any{}
	for _, a := range associations {
		var ids []uint64
		related := s.db.Table("(?) a", associationUnion(s.db)).Joins("JOIN game_users ru ON ru.id=a.user_id AND ru.deleted_at IS NULL")
		related = content.ApplySiteStrict(related, "ru", scope)
		related.Where("association_type=? AND association_value=? AND user_id<>?", a.AssociationType, a.AssociationValue, userID).Distinct().Pluck("user_id", &ids)
		for _, id := range ids {
			var u map[string]any
			relatedUser := content.ApplySiteStrict(s.db.Table("game_users"), "", scope)
			relatedUser.Select("id,user_no,username,status").Where("id=? AND deleted_at IS NULL", id).Take(&u)
			relatedUsers[id] = u
			nodes = append(nodes, map[string]any{"id": fmt.Sprint(id), "type": "USER", "label": u["username"]})
			edges = append(edges, map[string]any{"from": fmt.Sprint(userID), "to": fmt.Sprint(id), "type": a.AssociationType, "value": a.AssociationValueMasked})
		}
	}
	users := make([]map[string]any, 0, len(relatedUsers))
	for _, u := range relatedUsers {
		users = append(users, u)
	}
	var devices, ips []map[string]any
	s.db.Table("game_user_devices").Select("id,device_id,platform,status,risk_status,last_ip,last_seen_at").Where("user_id=? AND deleted_at IS NULL", userID).Find(&devices)
	for i := range devices {
		if v, ok := devices[i]["device_id"].(string); ok {
			devices[i]["device_id"] = mask(v)
		}
		if v, ok := devices[i]["last_ip"].(string); ok {
			devices[i]["last_ip"] = mask(v)
		}
	}
	s.db.Table("game_user_login_records").Select("ip_address,COUNT(*) count,MAX(created_at) last_seen_at").Where("user_id=? AND deleted_at IS NULL", userID).Group("ip_address").Find(&ips)
	for i := range ips {
		if v, ok := ips[i]["ip_address"].(string); ok {
			ips[i]["ip_address"] = mask(v)
		}
	}
	return map[string]any{"overview": user, "relationshipGraph": map[string]any{"nodes": nodes, "edges": edges}, "associatedUsers": users, "devices": devices, "ips": ips, "associations": associations}, nil
}

func (s *Service) OperationLogs(q ListQuery) (PageResult, error) {
	var rows []map[string]any
	db := s.db.Table("game_admin_operation_logs").Where("deleted_at IS NULL AND permission LIKE 'risk:%'")
	db = contains(db, "operator_username", q.Values["operatorName"])
	db = contains(db, "action", q.Values["action"])
	db = equal(db, "target_type", q.Values["targetType"])
	db = equal(db, "target_id", q.Values["targetID"])
	if level := strings.TrimSpace(q.Values["riskLevel"]); level != "" {
		db = db.Where("JSON_UNQUOTE(JSON_EXTRACT(before_state, '$.riskLevel')) = ? OR JSON_UNQUOTE(JSON_EXTRACT(after_state, '$.riskLevel')) = ?", level, level)
	}
	db = timeRange(db, "created_at", q.Values["createdFrom"], q.Values["createdTo"])
	return paginate(db.Order("id DESC"), q, &rows)
}
func (s *Service) OperationLogDetail(id uint64) (map[string]any, error) {
	var row map[string]any
	if err := s.db.Table("game_admin_operation_logs").Where("id=? AND deleted_at IS NULL AND permission LIKE 'risk:%'", id).Take(&row).Error; err != nil {
		return nil, mapError(err, "RISK_OPERATION_LOG_NOT_FOUND")
	}
	return map[string]any{"overview": row}, nil
}
func associationTypes() string {
	return strings.Join([]string{"DEVICE", "FINGERPRINT", "IP", "WITHDRAWAL_ACCOUNT", "REFERRAL", "EMAIL", "MOBILE"}, ",")
}
