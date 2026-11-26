package systemadmin

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AdminRow struct {
	ID            uint64     `json:"administratorID"`
	Username      string     `json:"username"`
	DisplayName   string     `json:"displayName" gorm:"column:display_name"`
	TenantID      uint64     `json:"tenantID" gorm:"column:tenant_id"`
	Tenant        string     `json:"tenant"`
	SiteID        uint64     `json:"siteID" gorm:"column:site_id"`
	Site          string     `json:"site"`
	DepartmentID  *uint64    `json:"departmentID" gorm:"column:department_id"`
	Department    string     `json:"department"`
	Roles         string     `json:"roles"`
	Mobile        string     `json:"mobile"`
	Email         string     `json:"email"`
	Status        bool       `json:"status"`
	MFAEnabled    bool       `json:"mfaEnabled" gorm:"column:mfa_enabled"`
	LastLoginIP   string     `json:"lastLoginIP" gorm:"column:last_login_ip"`
	LastLoginTime *time.Time `json:"lastLoginTime" gorm:"column:last_login_at"`
	CreatedAt     time.Time  `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt     time.Time  `json:"updatedAt" gorm:"column:updated_at"`
}

func adminSelect() string {
	return `a.id,a.username,a.display_name,a.tenant_id,COALESCE(t.tenant_name,'Platform') tenant,a.site_id,COALESCE(st.site_name,'') site,a.department_id,COALESCE(d.name_en,'') department,
		COALESCE(GROUP_CONCAT(DISTINCT r.name_en ORDER BY r.name_en SEPARATOR ', '),'') roles,
		a.mobile,a.email,a.status,a.mfa_enabled,a.last_login_ip,a.last_login_at,a.created_at,a.updated_at`
}

func (s *Service) ListAdmins(q Query, actor Actor) (Result, error) {
	rows := []AdminRow{}
	db := s.db.Table("game_admin_accounts a").
		Select(adminSelect()).
		Joins("LEFT JOIN game_saas_tenants t ON t.id=a.tenant_id AND t.deleted_at IS NULL").
		Joins("LEFT JOIN game_saas_sites st ON st.id=a.site_id AND st.deleted_at IS NULL").
		Joins("LEFT JOIN game_admin_departments d ON d.id=a.department_id AND d.deleted_at IS NULL").
		Joins("LEFT JOIN game_admin_account_roles ar ON ar.admin_id=a.id AND ar.deleted_at IS NULL").
		Joins("LEFT JOIN game_admin_roles r ON r.id=ar.role_id AND r.deleted_at IS NULL").
		Where("a.deleted_at IS NULL").
		Group("a.id,t.tenant_name,st.site_name,d.name_en")
	if actor.Platform {
		db = exact(db, "a.tenant_id", value(q, "tenantID"))
		db = exact(db, "a.site_id", value(q, "siteID"))
		db = like(db, "t.tenant_name", value(q, "tenant"))
		db = like(db, "st.site_name", value(q, "site"))
	} else {
		db = db.Where("a.tenant_id=?", actor.TenantID)
	}
	db = exact(db, "a.id", value(q, "administratorID"))
	db = like(db, "a.username", value(q, "username"))
	db = like(db, "a.display_name", value(q, "displayName"))
	db = exact(db, "a.department_id", value(q, "department"))
	if role := value(q, "role"); role != "" {
		db = db.Where("EXISTS (SELECT 1 FROM game_admin_account_roles x WHERE x.admin_id=a.id AND x.role_id=? AND x.deleted_at IS NULL)", role)
	}
	db = exact(db, "a.status", value(q, "status"))
	db = like(db, "a.mobile", value(q, "mobile"))
	db = like(db, "a.email", value(q, "email"))
	db = timeRange(db, "a.created_at", value(q, "createdFrom"), value(q, "createdTo"))
	db = timeRange(db, "a.last_login_at", value(q, "lastLoginFrom"), value(q, "lastLoginTo"))
	return paginate(db.Order("a.created_at DESC,a.id DESC"), q, &rows)
}

func (s *Service) AdminDetail(id uint64, actor Actor) (map[string]any, error) {
	var row AdminRow
	db := s.db.Table("game_admin_accounts a").Select(adminSelect()).
		Joins("LEFT JOIN game_saas_tenants t ON t.id=a.tenant_id AND t.deleted_at IS NULL").
		Joins("LEFT JOIN game_saas_sites st ON st.id=a.site_id AND st.deleted_at IS NULL").
		Joins("LEFT JOIN game_admin_departments d ON d.id=a.department_id AND d.deleted_at IS NULL").
		Joins("LEFT JOIN game_admin_account_roles ar ON ar.admin_id=a.id AND ar.deleted_at IS NULL").
		Joins("LEFT JOIN game_admin_roles r ON r.id=ar.role_id AND r.deleted_at IS NULL").
		Where("a.id=? AND a.deleted_at IS NULL", id).Group("a.id,t.tenant_name,st.site_name,d.name_en")
	if !actor.Platform {
		db = db.Where("a.tenant_id=?", actor.TenantID)
	}
	if err := db.Take(&row).Error; err != nil {
		return nil, rowNotFound("administrator")
	}
	roles := []map[string]any{}
	_ = s.db.Table("game_admin_account_roles ar").Select("r.id,r.code,r.name_zh nameZh,r.name_en nameEn,r.data_scope dataScope").
		Joins("JOIN game_admin_roles r ON r.id=ar.role_id AND r.deleted_at IS NULL").
		Where("ar.admin_id=? AND ar.deleted_at IS NULL", id).Order("r.code").Find(&roles).Error
	permissions := []map[string]any{}
	_ = s.db.Table("game_admin_account_roles ar").
		Select("DISTINCT p.id,p.code,p.name_zh nameZh,p.name_en nameEn,p.module,p.permission_type permissionType,p.enabled").
		Joins("JOIN game_admin_role_permissions rp ON rp.role_id=ar.role_id AND rp.deleted_at IS NULL").
		Joins("JOIN game_admin_permissions p ON p.id=rp.permission_id AND p.enabled=1 AND p.deleted_at IS NULL").
		Where("ar.admin_id=? AND ar.deleted_at IS NULL", id).Order("p.module,p.permission_type,p.code").Find(&permissions).Error
	loginLogs := []map[string]any{}
	_ = s.db.Table("game_admin_login_logs").Select("id,success,failure_reason failureReason,ip_address ipAddress,country,device,operating_system operatingSystem,browser,created_at createdAt").
		Where("admin_id=? AND deleted_at IS NULL", id).Order("created_at DESC").Limit(50).Find(&loginLogs).Error
	operationLogs := []map[string]any{}
	_ = s.db.Table("game_admin_operation_logs").Select("id,permission,module,action,result,target_type targetType,target_id targetID,reason,ip_address ipAddress,created_at createdAt").
		Where("operator_id=? AND deleted_at IS NULL", id).Order("created_at DESC").Limit(50).Find(&operationLogs).Error
	return map[string]any{
		"overview":            row,
		"rolesAndPermissions": map[string]any{"roles": roles, "permissions": permissions},
		"departments":         map[string]any{"departmentID": row.DepartmentID, "department": row.Department},
		"loginSecurity":       map[string]any{"mfaEnabled": row.MFAEnabled, "lastLoginIP": row.LastLoginIP, "lastLoginTime": row.LastLoginTime},
		"loginLogs":           loginLogs, "operationLogs": operationLogs,
	}, nil
}

func validateAdminInput(input AdminInput, creating bool) error {
	if creating && validateCode(input.Username) != nil {
		return NewError("INVALID_USERNAME", "invalid username")
	}
	if strings.TrimSpace(input.DisplayName) == "" {
		return NewError("DISPLAY_NAME_REQUIRED", "display name is required")
	}
	if creating && len(input.Password) < 10 {
		return NewError("WEAK_PASSWORD", "password must contain at least 10 characters")
	}
	return nil
}

func (s *Service) normalizeAdminScope(input *AdminInput, actor Actor) error {
	if !actor.Platform {
		input.TenantID = actor.TenantID
		if input.SiteID == 0 {
			input.SiteID = actor.SiteID
		}
	}
	if input.TenantID == 0 {
		input.SiteID = 0
		return nil
	}
	var tenantCount int64
	if err := s.db.Table("game_saas_tenants").Where("id=? AND enabled=1 AND deleted_at IS NULL", input.TenantID).Count(&tenantCount).Error; err != nil {
		return err
	}
	if tenantCount == 0 {
		return NewError("TENANT_NOT_FOUND", "tenant not found")
	}
	if input.SiteID > 0 {
		var siteCount int64
		if err := s.db.Table("game_saas_sites").Where("id=? AND tenant_id=? AND enabled=1 AND deleted_at IS NULL", input.SiteID, input.TenantID).Count(&siteCount).Error; err != nil {
			return err
		}
		if siteCount == 0 {
			return NewError("SITE_NOT_FOUND", "site not found")
		}
	}
	return nil
}

func (s *Service) requireAdminScope(id uint64, actor Actor) error {
	var target struct {
		TenantID uint64 `gorm:"column:tenant_id"`
	}
	if err := s.db.Table("game_admin_accounts").Select("tenant_id").Where("id=? AND deleted_at IS NULL", id).Take(&target).Error; err != nil {
		return rowNotFound("administrator")
	}
	if !actor.Platform && target.TenantID != actor.TenantID {
		return NewError("PERMISSION_DENIED", "permission denied")
	}
	return nil
}

func (s *Service) CreateAdmin(input AdminInput, actor Actor) (uint64, error) {
	if err := validateAdminInput(input, true); err != nil {
		return 0, err
	}
	if err := s.normalizeAdminScope(&input, actor); err != nil {
		return 0, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	row := map[string]any{
		"username": strings.TrimSpace(input.Username), "display_name": strings.TrimSpace(input.DisplayName), "tenant_id": input.TenantID, "site_id": input.SiteID,
		"password_hash": string(hash), "department_id": input.DepartmentID, "mobile": strings.TrimSpace(input.Mobile),
		"email": strings.TrimSpace(input.Email), "status": boolValue(boolDefault(input.Status, true)),
		"last_password_changed_at": now(), "must_change_password": 1, "created_at": now(), "updated_at": now(),
	}
	var id uint64
	err = s.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Table("game_admin_accounts").Create(row)
		if result.Error != nil {
			return duplicateError(result.Error)
		}
		if err := tx.Table("game_admin_accounts").Select("id").Where("tenant_id=? AND username=? AND deleted_at IS NULL", input.TenantID, row["username"]).Take(&id).Error; err != nil {
			return err
		}
		if err := assignAdminRoles(tx, id, input.TenantID, input.RoleIDs); err != nil {
			return err
		}
		return audit(tx, actor, "system:admin:create", "system.admin.create", "administrator", id, nil, map[string]any{"username": input.Username, "roles": input.RoleIDs}, "")
	})
	return id, err
}

func (s *Service) UpdateAdmin(id uint64, input AdminInput, actor Actor) error {
	if err := validateAdminInput(input, false); err != nil {
		return err
	}
	if err := s.normalizeAdminScope(&input, actor); err != nil {
		return err
	}
	var before map[string]any
	if err := s.db.Table("game_admin_accounts").Where("id=? AND deleted_at IS NULL", id).Take(&before).Error; err != nil {
		return rowNotFound("administrator")
	}
	updates := map[string]any{"display_name": strings.TrimSpace(input.DisplayName), "tenant_id": input.TenantID, "site_id": input.SiteID, "department_id": input.DepartmentID, "mobile": strings.TrimSpace(input.Mobile), "email": strings.TrimSpace(input.Email), "updated_at": now(), "version": gorm.Expr("version+1")}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_admin_accounts").Where("id=? AND version=? AND deleted_at IS NULL", id, before["version"]).Updates(updates).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:admin:update", "system.admin.update", "administrator", id, before, updates, "")
	})
}

func assignAdminRoles(tx *gorm.DB, adminID, tenantID uint64, roleIDs []uint64) error {
	roleIDs = uniqueIDs(roleIDs)
	if len(roleIDs) > 0 {
		var count int64
		if err := tx.Table("game_admin_roles").Where("id IN ? AND tenant_id=? AND status=1 AND deleted_at IS NULL", roleIDs, tenantID).Count(&count).Error; err != nil {
			return err
		}
		if int(count) != len(roleIDs) {
			return NewError("INVALID_ROLE", "one or more roles are unavailable")
		}
	}
	if err := tx.Table("game_admin_account_roles").Where("admin_id=? AND deleted_at IS NULL", adminID).Update("deleted_at", now()).Error; err != nil {
		return err
	}
	for _, roleID := range roleIDs {
		row := map[string]any{"admin_id": adminID, "role_id": roleID, "created_at": now(), "updated_at": now(), "deleted_at": nil}
		if err := tx.Table("game_admin_account_roles").Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "admin_id"}, {Name: "role_id"}},
			DoUpdates: clause.Assignments(map[string]any{"deleted_at": nil, "updated_at": now()}),
		}).Create(row).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) AssignAdminRoles(id uint64, input IDsInput, actor Actor) error {
	if err := requireReason(input.Reason); err != nil {
		return err
	}
	if actor.ID == id {
		var protected int64
		_ = s.db.Table("game_admin_roles").Where("id IN ? AND system_protected=1 AND deleted_at IS NULL", uniqueIDs(input.IDs)).Count(&protected).Error
		if protected > 0 && !s.actorIsSuper(actor.ID) {
			return NewError("SELF_PRIVILEGE_ESCALATION", "cannot escalate own privileges")
		}
	}
	before := []uint64{}
	_ = s.db.Table("game_admin_account_roles").Where("admin_id=? AND deleted_at IS NULL", id).Pluck("role_id", &before).Error
	var target struct {
		TenantID uint64 `gorm:"column:tenant_id"`
	}
	if err := s.db.Table("game_admin_accounts").Select("tenant_id").Where("id=? AND deleted_at IS NULL", id).Take(&target).Error; err != nil {
		return rowNotFound("administrator")
	}
	if !actor.Platform && target.TenantID != actor.TenantID {
		return NewError("PERMISSION_DENIED", "permission denied")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := assignAdminRoles(tx, id, target.TenantID, input.IDs); err != nil {
			return err
		}
		return audit(tx, actor, "system:admin:assign-role", "system.admin.roles.assign", "administrator", id, before, input.IDs, input.Reason)
	})
}

func (s *Service) actorIsSuper(id uint64) bool {
	var count int64
	_ = s.db.Table("game_admin_account_roles ar").Joins("JOIN game_admin_roles r ON r.id=ar.role_id AND r.code='super-admin' AND r.tenant_id=0 AND r.status=1 AND r.deleted_at IS NULL").Where("ar.admin_id=? AND ar.deleted_at IS NULL", id).Count(&count).Error
	return count > 0
}
func (s *Service) isLastSuper(id uint64) bool {
	var assigned int64
	_ = s.db.Table("game_admin_account_roles ar").Joins("JOIN game_admin_roles r ON r.id=ar.role_id AND r.code='super-admin' AND r.tenant_id=0 AND r.deleted_at IS NULL").Where("ar.admin_id=? AND ar.deleted_at IS NULL", id).Count(&assigned).Error
	if assigned == 0 {
		return false
	}
	var active int64
	_ = s.db.Table("game_admin_accounts a").Joins("JOIN game_admin_account_roles ar ON ar.admin_id=a.id AND ar.deleted_at IS NULL").Joins("JOIN game_admin_roles r ON r.id=ar.role_id AND r.code='super-admin' AND r.tenant_id=0 AND r.deleted_at IS NULL").Where("a.status=1 AND a.tenant_id=0 AND a.deleted_at IS NULL").Distinct("a.id").Count(&active).Error
	return active <= 1
}

func (s *Service) AdminStatus(id uint64, enabled bool, reason string, actor Actor) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	if actor.ID == id && !enabled {
		return NewError("SELF_DISABLE_NOT_ALLOWED", "cannot disable own account")
	}
	if !enabled && s.isLastSuper(id) {
		return NewError("LAST_SUPER_ADMIN", "cannot disable the last super administrator")
	}
	if err := s.requireAdminScope(id, actor); err != nil {
		return err
	}
	var before struct{ Status int }
	if err := s.db.Table("game_admin_accounts").Select("status").Where("id=? AND deleted_at IS NULL", id).Take(&before).Error; err != nil {
		return rowNotFound("administrator")
	}
	if before.Status == boolValue(enabled) {
		return NewError("INVALID_STATUS_TRANSITION", "administrator is already in requested status")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		db := tx.Table("game_admin_accounts").Where("id=?", id)
		if !actor.Platform {
			db = db.Where("tenant_id=?", actor.TenantID)
		}
		if err := db.Updates(map[string]any{"status": boolValue(enabled), "version": gorm.Expr("version+1"), "updated_at": now()}).Error; err != nil {
			return err
		}
		if !enabled {
			if err := tx.Table("game_admin_sessions").Where("admin_id=? AND revoked_at IS NULL", id).Updates(map[string]any{"revoked_at": now(), "updated_at": now()}).Error; err != nil {
				return err
			}
		}
		return audit(tx, actor, "system:admin:status", "system.admin.status", "administrator", id, before, map[string]any{"status": enabled}, reason)
	})
}

func randomPassword() (string, error) {
	data := make([]byte, 12)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return "A9!" + hex.EncodeToString(data)[:16], nil
}
func (s *Service) ResetPassword(id uint64, reason string, actor Actor) (string, error) {
	if err := requireReason(reason); err != nil {
		return "", err
	}
	if err := s.requireAdminScope(id, actor); err != nil {
		return "", err
	}
	password, err := randomPassword()
	if err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var previous struct{ PasswordHash string }
		if err := tx.Table("game_admin_accounts").Select("password_hash").Where("id=? AND deleted_at IS NULL", id).Take(&previous).Error; err != nil {
			return rowNotFound("administrator")
		}
		if previous.PasswordHash != "" {
			if err := tx.Table("game_admin_password_history").Create(map[string]any{"admin_id": id, "password_hash": previous.PasswordHash, "created_at": now()}).Error; err != nil {
				return err
			}
		}
		db := tx.Table("game_admin_accounts").Where("id=? AND deleted_at IS NULL", id)
		if !actor.Platform {
			db = db.Where("tenant_id=?", actor.TenantID)
		}
		result := db.Updates(map[string]any{"password_hash": string(hash), "last_password_changed_at": now(), "must_change_password": 1, "version": gorm.Expr("version+1"), "updated_at": now()})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return rowNotFound("administrator")
		}
		if err := tx.Table("game_admin_sessions").Where("admin_id=? AND revoked_at IS NULL", id).Updates(map[string]any{"revoked_at": now(), "updated_at": now()}).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:admin:reset-password", "system.admin.password.reset", "administrator", id, nil, map[string]any{"passwordReset": true}, reason)
	})
	return password, err
}
func (s *Service) ResetMFA(id uint64, reason string, actor Actor) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	if err := s.requireAdminScope(id, actor); err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		db := tx.Table("game_admin_accounts").Where("id=? AND deleted_at IS NULL", id)
		if !actor.Platform {
			db = db.Where("tenant_id=?", actor.TenantID)
		}
		result := db.Updates(map[string]any{"mfa_enabled": 0, "mfa_secret_hash": "", "mfa_secret_encrypted": "", "version": gorm.Expr("version+1"), "updated_at": now()})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return rowNotFound("administrator")
		}
		if err := tx.Table("game_admin_mfa_recovery_codes").Where("admin_id=? AND deleted_at IS NULL", id).Update("deleted_at", now()).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:admin:reset-mfa", "system.admin.mfa.reset", "administrator", id, map[string]any{"mfa": "configured"}, map[string]any{"mfa": "reset"}, reason)
	})
}
func (s *Service) ForceLogout(id uint64, reason string, actor Actor) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	if err := s.requireAdminScope(id, actor); err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Table("game_admin_accounts").Where("id=? AND deleted_at IS NULL", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return rowNotFound("administrator")
		}
		result := tx.Table("game_admin_sessions").Where("admin_id=? AND revoked_at IS NULL AND deleted_at IS NULL", id).Updates(map[string]any{"revoked_at": now(), "updated_at": now()})
		if result.Error != nil {
			return result.Error
		}
		return audit(tx, actor, "system:admin:force-logout", "system.admin.force_logout", "administrator", id, nil, map[string]any{"revokedSessions": result.RowsAffected}, reason)
	})
}

func (s *Service) MoveAdminDepartment(id uint64, departmentID *uint64, reason string, actor Actor) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	if err := s.requireAdminScope(id, actor); err != nil {
		return err
	}
	var before struct{ DepartmentID *uint64 }
	if err := s.db.Table("game_admin_accounts").Select("department_id").Where("id=? AND deleted_at IS NULL", id).Take(&before).Error; err != nil {
		return rowNotFound("administrator")
	}
	if departmentID != nil {
		var count int64
		db := s.db.Table("game_admin_departments").Where("id=? AND enabled=1 AND deleted_at IS NULL", *departmentID)
		if !actor.Platform {
			db = db.Where("tenant_id=?", actor.TenantID)
		}
		if err := db.Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return rowNotFound("department")
		}
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		db := tx.Table("game_admin_accounts").Where("id=?", id)
		if !actor.Platform {
			db = db.Where("tenant_id=?", actor.TenantID)
		}
		if err := db.Updates(map[string]any{"department_id": departmentID, "updated_at": now()}).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:admin:update", "system.admin.department.move", "administrator", id, before, map[string]any{"departmentID": departmentID}, reason)
	})
}

func (s *Service) adminExists(id uint64) error {
	var count int64
	if err := s.db.Table("game_admin_accounts").Where("id=? AND deleted_at IS NULL", id).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return rowNotFound("administrator")
	}
	return nil
}

func temporaryPasswordResponse(password string) map[string]any {
	return map[string]any{"temporaryPassword": password, "message": fmt.Sprintf("temporary password generated at %s", now().Format(time.RFC3339))}
}

var _ = clause.Assignment{}
