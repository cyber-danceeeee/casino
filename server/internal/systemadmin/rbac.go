package systemadmin

import (
	"casino/internal/singlesite"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RoleRow struct {
	ID                 uint64    `json:"id"`
	TenantID           uint64    `json:"tenantID" gorm:"column:tenant_id"`
	Tenant             string    `json:"tenant"`
	Code               string    `json:"code"`
	NameZh             string    `json:"nameZh" gorm:"column:name_zh"`
	NameEn             string    `json:"nameEn" gorm:"column:name_en"`
	DescriptionZh      string    `json:"descriptionZh" gorm:"column:description_zh"`
	DescriptionEn      string    `json:"descriptionEn" gorm:"column:description_en"`
	DataScope          string    `json:"dataScope" gorm:"column:data_scope"`
	Enabled            bool      `json:"enabled" gorm:"column:status"`
	SystemProtected    bool      `json:"systemProtected" gorm:"column:system_protected"`
	AdministratorCount int64     `json:"administratorCount" gorm:"column:administrator_count"`
	PermissionCount    int64     `json:"permissionCount" gorm:"column:permission_count"`
	UpdatedAt          time.Time `json:"updatedAt" gorm:"column:updated_at"`
	CreatedAt          time.Time `json:"createdAt" gorm:"column:created_at"`
}

func (s *Service) ListRoles(q Query, actor Actor) (Result, error) {
	rows := []RoleRow{}
	db := s.db.Table("game_admin_roles r").Select(`r.id,r.tenant_id,COALESCE(t.tenant_name,'Platform') tenant,r.code,r.name_zh,r.name_en,r.description_zh,r.description_en,r.data_scope,r.status,r.system_protected,r.created_at,r.updated_at,
		(SELECT COUNT(DISTINCT ar.admin_id) FROM game_admin_account_roles ar WHERE ar.role_id=r.id AND ar.deleted_at IS NULL) administrator_count,
		(SELECT COUNT(DISTINCT rp.permission_id) FROM game_admin_role_permissions rp WHERE rp.role_id=r.id AND rp.deleted_at IS NULL) permission_count`).
		Joins("LEFT JOIN game_saas_tenants t ON t.id=r.tenant_id AND t.deleted_at IS NULL").
		Where("r.deleted_at IS NULL")
	if actor.Platform {
		db = exact(db, "r.tenant_id", value(q, "tenantID"))
		db = like(db, "t.tenant_name", value(q, "tenant"))
	} else {
		db = db.Where("r.tenant_id=?", actor.TenantID)
	}
	db = like(db, "r.code", value(q, "code"))
	name := value(q, "name")
	if name != "" {
		db = db.Where("(r.name_zh LIKE ? OR r.name_en LIKE ?)", "%"+name+"%", "%"+name+"%")
	}
	db = exact(db, "r.status", value(q, "status"))
	db = exact(db, "r.data_scope", value(q, "dataScope"))
	db = timeRange(db, "r.created_at", value(q, "createdFrom"), value(q, "createdTo"))
	return paginate(db.Order("r.system_protected DESC,r.created_at DESC,r.id DESC"), q, &rows)
}

func (s *Service) RoleDetail(id uint64, actor Actor) (map[string]any, error) {
	var role RoleRow
	db := s.db.Table("game_admin_roles r").Select(`r.id,r.tenant_id,COALESCE(t.tenant_name,'Platform') tenant,r.code,r.name_zh,r.name_en,r.description_zh,r.description_en,r.data_scope,r.status,r.system_protected,r.created_at,r.updated_at,
		(SELECT COUNT(*) FROM game_admin_account_roles ar WHERE ar.role_id=r.id AND ar.deleted_at IS NULL) administrator_count,
		(SELECT COUNT(*) FROM game_admin_role_permissions rp WHERE rp.role_id=r.id AND rp.deleted_at IS NULL) permission_count`).
		Joins("LEFT JOIN game_saas_tenants t ON t.id=r.tenant_id AND t.deleted_at IS NULL").
		Where("r.id=? AND r.deleted_at IS NULL", id)
	if !actor.Platform {
		db = db.Where("r.tenant_id=?", actor.TenantID)
	}
	if err := db.Take(&role).Error; err != nil {
		return nil, rowNotFound("role")
	}
	permissions := []map[string]any{}
	_ = s.db.Table("game_admin_role_permissions rp").Select("p.id,p.code,p.name_zh nameZh,p.name_en nameEn,p.module,p.permission_type permissionType,p.enabled").
		Joins("JOIN game_admin_permissions p ON p.id=rp.permission_id AND p.deleted_at IS NULL").Where("rp.role_id=? AND rp.deleted_at IS NULL", id).Order("p.module,p.code").Find(&permissions).Error
	menus := []map[string]any{}
	_ = s.db.Table("game_admin_role_menus rm").Select("m.id,m.code,m.name_zh nameZh,m.name_en nameEn,m.route_path routePath,m.menu_type menuType").
		Joins("JOIN game_admin_menus m ON m.id=rm.menu_id AND m.deleted_at IS NULL").Where("rm.role_id=? AND rm.deleted_at IS NULL", id).Order("m.sort_order,m.id").Find(&menus).Error
	admins := []map[string]any{}
	_ = s.db.Table("game_admin_account_roles ar").Select("a.id,a.username,a.display_name displayName,a.status").
		Joins("JOIN game_admin_accounts a ON a.id=ar.admin_id AND a.deleted_at IS NULL").Where("ar.role_id=? AND ar.deleted_at IS NULL", id).Order("a.username").Find(&admins).Error
	logs := s.operationLogs("role", id)
	return map[string]any{"overview": role, "permissions": permissions, "menus": menus, "administrators": admins, "operationLogs": logs}, nil
}

func validateRole(input RoleInput, creating bool) error {
	if creating {
		if err := validateCode(input.Code); err != nil {
			return err
		}
	}
	if strings.TrimSpace(input.NameZh) == "" || strings.TrimSpace(input.NameEn) == "" {
		return NewError("ROLE_NAME_REQUIRED", "role names are required")
	}
	switch input.DataScope {
	case "ALL", "DEPARTMENT", "DEPARTMENT_AND_CHILDREN", "SELF", "CUSTOM":
	default:
		return NewError("INVALID_DATA_SCOPE", "invalid data scope")
	}
	return nil
}

func (s *Service) roleScopedTable(actor Actor) *gorm.DB {
	db := s.db.Table("game_admin_roles")
	if !actor.Platform {
		db = db.Where("tenant_id=?", actor.TenantID)
	}
	return db
}

func (s *Service) requireRoleScope(id uint64, actor Actor) error {
	var count int64
	db := s.roleScopedTable(actor).Where("id=? AND deleted_at IS NULL", id)
	if err := db.Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return rowNotFound("role")
	}
	return nil
}

func (s *Service) CreateRole(input RoleInput, actor Actor) (uint64, error) {
	if err := validateRole(input, true); err != nil {
		return 0, err
	}
	row := map[string]any{"tenant_id": actor.TenantID, "code": strings.TrimSpace(input.Code), "name": input.NameEn, "name_zh": input.NameZh, "name_en": input.NameEn, "description_zh": input.DescriptionZh, "description_en": input.DescriptionEn, "data_scope": input.DataScope, "status": boolValue(boolDefault(input.Enabled, true)), "created_at": now(), "updated_at": now()}
	result := s.db.Table("game_admin_roles").Create(row)
	if result.Error != nil {
		return 0, duplicateError(result.Error)
	}
	id := row["id"].(uint64)
	if err := audit(s.db, actor, "system:role:create", "system.role.create", "role", id, nil, row, ""); err != nil {
		return 0, err
	}
	return id, nil
}
func (s *Service) UpdateRole(id uint64, input RoleInput, actor Actor) error {
	if err := validateRole(input, false); err != nil {
		return err
	}
	var before map[string]any
	if err := s.roleScopedTable(actor).Where("id=? AND deleted_at IS NULL", id).Take(&before).Error; err != nil {
		return rowNotFound("role")
	}
	updates := map[string]any{"name": input.NameEn, "name_zh": input.NameZh, "name_en": input.NameEn, "description_zh": input.DescriptionZh, "description_en": input.DescriptionEn, "data_scope": input.DataScope, "updated_at": now(), "version": gorm.Expr("version+1")}
	return s.db.Transaction(func(tx *gorm.DB) error {
		db := tx.Table("game_admin_roles").Where("id=? AND deleted_at IS NULL", id)
		if !actor.Platform {
			db = db.Where("tenant_id=?", actor.TenantID)
		}
		if err := db.Updates(updates).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:role:update", "system.role.update", "role", id, before, updates, "")
	})
}
func (s *Service) RoleStatus(id uint64, enabled bool, reason string, actor Actor) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	var row struct {
		Status          int
		SystemProtected bool
	}
	if err := s.roleScopedTable(actor).Select("status,system_protected").Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return rowNotFound("role")
	}
	if row.SystemProtected && !enabled {
		return NewError("PROTECTED_ROLE", "system-protected role cannot be disabled")
	}
	if row.Status == boolValue(enabled) {
		return NewError("INVALID_STATUS_TRANSITION", "role already has requested status")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		db := tx.Table("game_admin_roles").Where("id=?", id)
		if !actor.Platform {
			db = db.Where("tenant_id=?", actor.TenantID)
		}
		if err := db.Updates(map[string]any{"status": boolValue(enabled), "updated_at": now()}).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:role:update", "system.role.status", "role", id, row, map[string]any{"enabled": enabled}, reason)
	})
}
func replaceRelations(tx *gorm.DB, table, ownerColumn string, ownerID uint64, targetColumn string, ids []uint64) error {
	if err := tx.Table(table).Where(ownerColumn+"=? AND deleted_at IS NULL", ownerID).Update("deleted_at", now()).Error; err != nil {
		return err
	}
	for _, id := range uniqueIDs(ids) {
		row := map[string]any{ownerColumn: ownerID, targetColumn: id, "created_at": now(), "updated_at": now(), "deleted_at": nil}
		if err := tx.Table(table).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: ownerColumn}, {Name: targetColumn}}, DoUpdates: clause.Assignments(map[string]any{"deleted_at": nil, "updated_at": now()})}).Create(row).Error; err != nil {
			return err
		}
	}
	return nil
}
func (s *Service) AssignRolePermissions(id uint64, input IDsInput, actor Actor) error {
	if err := requireReason(input.Reason); err != nil {
		return err
	}
	if err := s.requireRoleScope(id, actor); err != nil {
		return err
	}
	ids := uniqueIDs(input.IDs)
	var count int64
	if len(ids) > 0 {
		if err := s.db.Table("game_admin_permissions").Where("id IN ? AND enabled=1 AND deleted_at IS NULL", ids).Count(&count).Error; err != nil {
			return err
		}
		if int(count) != len(ids) {
			return NewError("INVALID_PERMISSION", "one or more permissions are unavailable")
		}
	}
	before := []uint64{}
	_ = s.db.Table("game_admin_role_permissions").Where("role_id=? AND deleted_at IS NULL", id).Pluck("permission_id", &before).Error
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := replaceRelations(tx, "game_admin_role_permissions", "role_id", id, "permission_id", ids); err != nil {
			return err
		}
		return audit(tx, actor, "system:role:assign-permission", "system.role.permissions.assign", "role", id, before, ids, input.Reason)
	})
}
func (s *Service) AssignRoleMenus(id uint64, input IDsInput, actor Actor) error {
	if err := requireReason(input.Reason); err != nil {
		return err
	}
	if err := s.requireRoleScope(id, actor); err != nil {
		return err
	}
	ids := uniqueIDs(input.IDs)
	var count int64
	if len(ids) > 0 {
		if err := s.db.Table("game_admin_menus").Where("id IN ? AND enabled=1 AND deleted_at IS NULL", ids).Count(&count).Error; err != nil {
			return err
		}
		if int(count) != len(ids) {
			return NewError("INVALID_MENU", "one or more menus are unavailable")
		}
	}
	before := []uint64{}
	_ = s.db.Table("game_admin_role_menus").Where("role_id=? AND deleted_at IS NULL", id).Pluck("menu_id", &before).Error
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := replaceRelations(tx, "game_admin_role_menus", "role_id", id, "menu_id", ids); err != nil {
			return err
		}
		return audit(tx, actor, "system:role:assign-menu", "system.role.menus.assign", "role", id, before, ids, input.Reason)
	})
}
func (s *Service) CopyRole(id uint64, input RoleInput, actor Actor) (uint64, error) {
	if err := requireReason(input.Reason); err != nil {
		return 0, err
	}
	if err := validateRole(input, true); err != nil {
		return 0, err
	}
	var source RoleRow
	if err := s.roleScopedTable(actor).Where("id=? AND deleted_at IS NULL", id).Take(&source).Error; err != nil {
		return 0, rowNotFound("role")
	}
	var newID uint64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		row := map[string]any{"tenant_id": actor.TenantID, "code": input.Code, "name": input.NameEn, "name_zh": input.NameZh, "name_en": input.NameEn, "description_zh": input.DescriptionZh, "description_en": input.DescriptionEn, "data_scope": input.DataScope, "status": boolValue(boolDefault(input.Enabled, true)), "created_at": now(), "updated_at": now()}
		if err := tx.Table("game_admin_roles").Create(row).Error; err != nil {
			return duplicateError(err)
		}
		newID = row["id"].(uint64)
		permissionIDs := []uint64{}
		menuIDs := []uint64{}
		_ = tx.Table("game_admin_role_permissions").Where("role_id=? AND deleted_at IS NULL", id).Pluck("permission_id", &permissionIDs).Error
		_ = tx.Table("game_admin_role_menus").Where("role_id=? AND deleted_at IS NULL", id).Pluck("menu_id", &menuIDs).Error
		if err := replaceRelations(tx, "game_admin_role_permissions", "role_id", newID, "permission_id", permissionIDs); err != nil {
			return err
		}
		if err := replaceRelations(tx, "game_admin_role_menus", "role_id", newID, "menu_id", menuIDs); err != nil {
			return err
		}
		return audit(tx, actor, "system:role:create", "system.role.copy", "role", newID, map[string]any{"sourceRoleID": id}, row, input.Reason)
	})
	return newID, err
}
func (s *Service) DeleteRole(id uint64, reason string, actor Actor) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	var role struct{ SystemProtected bool }
	if err := s.roleScopedTable(actor).Select("system_protected").Where("id=? AND deleted_at IS NULL", id).Take(&role).Error; err != nil {
		return rowNotFound("role")
	}
	if role.SystemProtected {
		return NewError("PROTECTED_ROLE", "system-protected role cannot be deleted")
	}
	var assigned int64
	if err := s.db.Table("game_admin_account_roles").Where("role_id=? AND deleted_at IS NULL", id).Count(&assigned).Error; err != nil {
		return err
	}
	if assigned > 0 {
		return NewError("ROLE_IN_USE", "role is assigned to administrators")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		db := tx.Table("game_admin_roles").Where("id=?", id)
		if !actor.Platform {
			db = db.Where("tenant_id=?", actor.TenantID)
		}
		if err := db.Update("deleted_at", now()).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:role:delete", "system.role.delete", "role", id, role, nil, reason)
	})
}

type PermissionRow struct {
	ID                uint64    `json:"id"`
	Code              string    `json:"code"`
	NameZh            string    `json:"nameZh" gorm:"column:name_zh"`
	NameEn            string    `json:"nameEn" gorm:"column:name_en"`
	Module            string    `json:"module"`
	PermissionType    string    `json:"permissionType" gorm:"column:permission_type"`
	APIPath           string    `json:"apiPath" gorm:"column:api_path"`
	AssignedRoleCount int64     `json:"assignedRoleCount" gorm:"column:assigned_role_count"`
	Enabled           bool      `json:"enabled"`
	SystemProtected   bool      `json:"systemProtected" gorm:"column:system_protected"`
	UpdatedAt         time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

func (s *Service) ListPermissions(q Query) (Result, error) {
	rows := []PermissionRow{}
	db := s.db.Table("game_admin_permissions p").Select(`p.id,p.code,p.name_zh,p.name_en,p.module,p.permission_type,p.api_path,p.enabled,p.system_protected,p.updated_at,
		(SELECT COUNT(DISTINCT rp.role_id) FROM game_admin_role_permissions rp WHERE rp.permission_id=p.id AND rp.deleted_at IS NULL) assigned_role_count`).Where("p.deleted_at IS NULL")
	if singlesite.Active() {
		db = db.Where("p.code NOT LIKE ?", "system:tenant:%")
	}
	db = like(db, "p.code", value(q, "code"))
	name := value(q, "name")
	if name != "" {
		db = db.Where("(p.name_zh LIKE ? OR p.name_en LIKE ?)", "%"+name+"%", "%"+name+"%")
	}
	db = exact(db, "p.module", value(q, "module"))
	db = exact(db, "p.permission_type", value(q, "permissionType"))
	db = exact(db, "p.enabled", value(q, "status"))
	return paginate(db.Order("p.module,p.code"), q, &rows)
}
func validatePermission(input PermissionInput, creating bool) error {
	if creating {
		if err := validateCode(input.Code); err != nil {
			return err
		}
	}
	if input.NameZh == "" || input.NameEn == "" {
		return NewError("PERMISSION_NAME_REQUIRED", "permission names are required")
	}
	switch input.PermissionType {
	case "MENU", "PAGE", "ACTION", "API", "DATA":
	default:
		return NewError("INVALID_PERMISSION_TYPE", "invalid permission type")
	}
	return nil
}
func (s *Service) CreatePermission(input PermissionInput, actor Actor) (uint64, error) {
	if err := validatePermission(input, true); err != nil {
		return 0, err
	}
	row := map[string]any{"code": input.Code, "name": input.NameEn, "name_zh": input.NameZh, "name_en": input.NameEn, "module": input.Module, "permission_type": input.PermissionType, "api_path": input.APIPath, "enabled": boolValue(boolDefault(input.Enabled, true)), "created_at": now(), "updated_at": now()}
	var id uint64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_admin_permissions").Create(row).Error; err != nil {
			return duplicateError(err)
		}
		id = row["id"].(uint64)
		if err := tx.Exec(`
			INSERT INTO game_admin_role_permissions (role_id, permission_id, created_at, updated_at, deleted_at)
			SELECT id, ?, NOW(3), NOW(3), NULL
			FROM game_admin_roles
			WHERE code = 'super-admin' AND status = 1 AND system_protected = 1 AND deleted_at IS NULL
			ON DUPLICATE KEY UPDATE deleted_at = NULL, updated_at = VALUES(updated_at)`, id).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:permission:create", "system.permission.create", "permission", id, nil, row, "")
	})
	return id, err
}
func (s *Service) UpdatePermission(id uint64, input PermissionInput, actor Actor) error {
	if err := validatePermission(input, false); err != nil {
		return err
	}
	var before map[string]any
	if err := s.db.Table("game_admin_permissions").Where("id=? AND deleted_at IS NULL", id).Take(&before).Error; err != nil {
		return rowNotFound("permission")
	}
	updates := map[string]any{"name": input.NameEn, "name_zh": input.NameZh, "name_en": input.NameEn, "module": input.Module, "permission_type": input.PermissionType, "api_path": input.APIPath, "updated_at": now()}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_admin_permissions").Where("id=?", id).Updates(updates).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:permission:update", "system.permission.update", "permission", id, before, updates, "")
	})
}
func (s *Service) PermissionStatus(id uint64, enabled bool, reason string, actor Actor) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	var row struct{ Enabled, SystemProtected bool }
	if err := s.db.Table("game_admin_permissions").Select("enabled,system_protected").Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return rowNotFound("permission")
	}
	if row.SystemProtected && !enabled {
		return NewError("PROTECTED_PERMISSION", "system-protected permission cannot be disabled")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_admin_permissions").Where("id=?", id).Updates(map[string]any{"enabled": boolValue(enabled), "updated_at": now()}).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:permission:update", "system.permission.status", "permission", id, row, map[string]any{"enabled": enabled}, reason)
	})
}
func (s *Service) PermissionDetail(id uint64) (map[string]any, error) {
	var row PermissionRow
	if err := s.db.Table("game_admin_permissions p").Select(`p.id,p.code,p.name_zh,p.name_en,p.module,p.permission_type,p.api_path,p.enabled,p.system_protected,p.updated_at,
		(SELECT COUNT(*) FROM game_admin_role_permissions rp WHERE rp.permission_id=p.id AND rp.deleted_at IS NULL) assigned_role_count`).Where("p.id=? AND p.deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return nil, rowNotFound("permission")
	}
	roles := []map[string]any{}
	_ = s.db.Table("game_admin_role_permissions rp").Select("r.id,r.code,r.name_zh nameZh,r.name_en nameEn").Joins("JOIN game_admin_roles r ON r.id=rp.role_id AND r.deleted_at IS NULL").Where("rp.permission_id=? AND rp.deleted_at IS NULL", id).Order("r.code").Find(&roles).Error
	return map[string]any{"overview": row, "assignedRoles": roles, "operationLogs": s.operationLogs("permission", id)}, nil
}
func (s *Service) DeletePermission(id uint64, reason string, actor Actor) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	var row struct{ SystemProtected bool }
	if err := s.db.Table("game_admin_permissions").Select("system_protected").Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return rowNotFound("permission")
	}
	if row.SystemProtected {
		return NewError("PROTECTED_PERMISSION", "system-protected permission cannot be deleted")
	}
	var assigned int64
	_ = s.db.Table("game_admin_role_permissions").Where("permission_id=? AND deleted_at IS NULL", id).Count(&assigned).Error
	if assigned > 0 {
		return NewError("PERMISSION_IN_USE", "permission is assigned to roles")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_admin_permissions").Where("id=?", id).Update("deleted_at", now()).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:permission:delete", "system.permission.delete", "permission", id, row, nil, reason)
	})
}
func (s *Service) operationLogs(target string, id uint64) []map[string]any {
	rows := []map[string]any{}
	_ = s.db.Table("game_admin_operation_logs").Select("id,operator_username operator,permission,action,result,reason,ip_address ipAddress,created_at createdAt").
		Where("target_type=? AND target_id=? AND deleted_at IS NULL", target, id).Order("created_at DESC").Limit(100).Find(&rows).Error
	return rows
}
