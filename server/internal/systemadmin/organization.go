package systemadmin

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type DepartmentRow struct {
	ID                 uint64    `json:"id"`
	ParentID           *uint64   `json:"parentID" gorm:"column:parent_id"`
	Parent             string    `json:"parent"`
	Code               string    `json:"code"`
	NameZh             string    `json:"nameZh" gorm:"column:name_zh"`
	NameEn             string    `json:"nameEn" gorm:"column:name_en"`
	LeaderID           *uint64   `json:"leaderID" gorm:"column:leader_id"`
	Leader             string    `json:"leader"`
	AdministratorCount int64     `json:"administratorCount" gorm:"column:administrator_count"`
	ChildCount         int64     `json:"childCount" gorm:"column:child_count"`
	SortOrder          int       `json:"sortOrder" gorm:"column:sort_order"`
	Enabled            bool      `json:"enabled"`
	SystemProtected    bool      `json:"systemProtected" gorm:"column:system_protected"`
	UpdatedAt          time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

func (s *Service) ListDepartments(q Query) (Result, error) {
	rows := []DepartmentRow{}
	db := s.db.Table("game_admin_departments d").Select(`d.id,d.parent_id,COALESCE(p.name_en,'') parent,d.code,d.name_zh,d.name_en,d.leader_id,COALESCE(a.display_name,'') leader,d.sort_order,d.enabled,d.system_protected,d.updated_at,
		(SELECT COUNT(*) FROM game_admin_accounts x WHERE x.department_id=d.id AND x.deleted_at IS NULL) administrator_count,
		(SELECT COUNT(*) FROM game_admin_departments c WHERE c.parent_id=d.id AND c.deleted_at IS NULL) child_count`).
		Joins("LEFT JOIN game_admin_departments p ON p.id=d.parent_id AND p.deleted_at IS NULL").
		Joins("LEFT JOIN game_admin_accounts a ON a.id=d.leader_id AND a.deleted_at IS NULL").
		Where("d.deleted_at IS NULL")
	db = like(db, "d.code", value(q, "code"))
	name := value(q, "name")
	if name != "" {
		db = db.Where("(d.name_zh LIKE ? OR d.name_en LIKE ?)", "%"+name+"%", "%"+name+"%")
	}
	db = exact(db, "d.enabled", value(q, "status"))
	db = exact(db, "d.parent_id", value(q, "parentID"))
	return paginate(db.Order("d.parent_id,d.sort_order,d.id"), q, &rows)
}
func validateDepartment(input DepartmentInput, creating bool) error {
	if creating {
		if err := validateCode(input.Code); err != nil {
			return err
		}
	}
	if input.NameZh == "" || input.NameEn == "" {
		return NewError("DEPARTMENT_NAME_REQUIRED", "department names are required")
	}
	return nil
}
func (s *Service) ensureNoDepartmentCycle(id uint64, parentID *uint64) error {
	if parentID == nil {
		return nil
	}
	if id == *parentID {
		return NewError("DEPARTMENT_CYCLE", "department cannot be its own parent")
	}
	current := parentID
	for depth := 0; current != nil && depth < 100; depth++ {
		if *current == id {
			return NewError("DEPARTMENT_CYCLE", "department cannot be moved under its descendant")
		}
		var row struct {
			ParentID *uint64 `gorm:"column:parent_id"`
		}
		result := s.db.Table("game_admin_departments").Select("parent_id").Where("id=? AND deleted_at IS NULL", *current).Take(&row)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return rowNotFound("department")
		}
		current = row.ParentID
	}
	if current != nil {
		return NewError("DEPARTMENT_DEPTH_EXCEEDED", "department hierarchy is too deep")
	}
	return nil
}
func (s *Service) CreateDepartment(input DepartmentInput, actor Actor) (uint64, error) {
	if err := validateDepartment(input, true); err != nil {
		return 0, err
	}
	if err := s.ensureNoDepartmentCycle(0, input.ParentID); err != nil {
		return 0, err
	}
	row := map[string]any{"parent_id": input.ParentID, "code": input.Code, "name_zh": input.NameZh, "name_en": input.NameEn, "leader_id": input.LeaderID, "sort_order": input.SortOrder, "enabled": boolValue(boolDefault(input.Enabled, true)), "created_at": now(), "updated_at": now()}
	if err := s.db.Table("game_admin_departments").Create(row).Error; err != nil {
		return 0, duplicateError(err)
	}
	id := row["id"].(uint64)
	if err := audit(s.db, actor, "system:department:create", "system.department.create", "department", id, nil, row, ""); err != nil {
		return 0, err
	}
	return id, nil
}
func (s *Service) UpdateDepartment(id uint64, input DepartmentInput, reason string, actor Actor) error {
	if err := validateDepartment(input, false); err != nil {
		return err
	}
	if err := s.ensureNoDepartmentCycle(id, input.ParentID); err != nil {
		return err
	}
	var before map[string]any
	if err := s.db.Table("game_admin_departments").Where("id=? AND deleted_at IS NULL", id).Take(&before).Error; err != nil {
		return rowNotFound("department")
	}
	updates := map[string]any{"parent_id": input.ParentID, "name_zh": input.NameZh, "name_en": input.NameEn, "leader_id": input.LeaderID, "sort_order": input.SortOrder, "updated_at": now(), "version": gorm.Expr("version+1")}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_admin_departments").Where("id=?", id).Updates(updates).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:department:update", "system.department.update", "department", id, before, updates, reason)
	})
}

func (s *Service) DepartmentStatus(id uint64, enabled bool, reason string, actor Actor) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	var row struct {
		Enabled         bool
		SystemProtected bool
	}
	if err := s.db.Table("game_admin_departments").Select("enabled,system_protected").Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return rowNotFound("department")
	}
	if row.SystemProtected && !enabled {
		return NewError("PROTECTED_DEPARTMENT", "system-protected department cannot be disabled")
	}
	if row.Enabled == enabled {
		return NewError("INVALID_STATUS_TRANSITION", "department already has requested status")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_admin_departments").Where("id=?", id).Updates(map[string]any{"enabled": boolValue(enabled), "version": gorm.Expr("version+1"), "updated_at": now()}).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:department:update", "system.department.status", "department", id, row, map[string]any{"enabled": enabled}, reason)
	})
}
func (s *Service) DeleteDepartment(id uint64, reason string, actor Actor) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	var row struct{ SystemProtected bool }
	if err := s.db.Table("game_admin_departments").Select("system_protected").Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return rowNotFound("department")
	}
	if row.SystemProtected {
		return NewError("PROTECTED_DEPARTMENT", "system-protected department cannot be deleted")
	}
	var children, admins int64
	_ = s.db.Table("game_admin_departments").Where("parent_id=? AND deleted_at IS NULL", id).Count(&children).Error
	_ = s.db.Table("game_admin_accounts").Where("department_id=? AND deleted_at IS NULL", id).Count(&admins).Error
	if children > 0 || admins > 0 {
		return NewError("DEPARTMENT_IN_USE", "department has active children or administrators")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_admin_departments").Where("id=?", id).Update("deleted_at", now()).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:department:delete", "system.department.delete", "department", id, row, nil, reason)
	})
}
func (s *Service) DepartmentDetail(id uint64) (map[string]any, error) {
	var row DepartmentRow
	if err := s.db.Table("game_admin_departments d").Select(`d.id,d.parent_id,COALESCE(p.name_en,'') parent,d.code,d.name_zh,d.name_en,d.leader_id,COALESCE(a.display_name,'') leader,d.sort_order,d.enabled,d.system_protected,d.updated_at,
		(SELECT COUNT(*) FROM game_admin_accounts x WHERE x.department_id=d.id AND x.deleted_at IS NULL) administrator_count,
		(SELECT COUNT(*) FROM game_admin_departments c WHERE c.parent_id=d.id AND c.deleted_at IS NULL) child_count`).
		Joins("LEFT JOIN game_admin_departments p ON p.id=d.parent_id AND p.deleted_at IS NULL").Joins("LEFT JOIN game_admin_accounts a ON a.id=d.leader_id AND a.deleted_at IS NULL").
		Where("d.id=? AND d.deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return nil, rowNotFound("department")
	}
	admins := []map[string]any{}
	_ = s.db.Table("game_admin_accounts").Select("id,username,display_name displayName,status").Where("department_id=? AND deleted_at IS NULL", id).Order("username").Find(&admins).Error
	return map[string]any{"overview": row, "administrators": admins, "operationLogs": s.operationLogs("department", id)}, nil
}

type MenuRow struct {
	ID              uint64    `json:"id"`
	ParentID        *uint64   `json:"parentID" gorm:"column:parent_id"`
	Parent          string    `json:"parent"`
	MenuType        string    `json:"menuType" gorm:"column:menu_type"`
	Code            string    `json:"code"`
	NameZh          string    `json:"nameZh" gorm:"column:name_zh"`
	NameEn          string    `json:"nameEn" gorm:"column:name_en"`
	RoutePath       string    `json:"routePath" gorm:"column:route_path"`
	ComponentPath   string    `json:"componentPath" gorm:"column:component_path"`
	Icon            string    `json:"icon"`
	PermissionCode  string    `json:"permissionCode" gorm:"column:permission_code"`
	SortOrder       int       `json:"sortOrder" gorm:"column:sort_order"`
	Visible         bool      `json:"visible"`
	Enabled         bool      `json:"enabled"`
	Cacheable       bool      `json:"cacheable"`
	ExternalLink    string    `json:"externalLink" gorm:"column:external_link"`
	SystemProtected bool      `json:"systemProtected" gorm:"column:system_protected"`
	UpdatedAt       time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

func (s *Service) ListMenus(q Query) (Result, error) {
	rows := []MenuRow{}
	db := s.db.Table("game_admin_menus m").Select("m.id,m.parent_id,COALESCE(p.name_en,'') parent,m.menu_type,m.code,m.name_zh,m.name_en,m.route_path,m.component_path,m.icon,m.permission_code,m.sort_order,m.visible,m.enabled,m.cacheable,m.external_link,m.system_protected,m.updated_at").
		Joins("LEFT JOIN game_admin_menus p ON p.id=m.parent_id AND p.deleted_at IS NULL").Where("m.deleted_at IS NULL")
	db = like(db, "m.code", value(q, "code"))
	name := value(q, "name")
	if name != "" {
		db = db.Where("(m.name_zh LIKE ? OR m.name_en LIKE ?)", "%"+name+"%", "%"+name+"%")
	}
	db = exact(db, "m.menu_type", value(q, "menuType"))
	db = exact(db, "m.enabled", value(q, "status"))
	db = exact(db, "m.permission_code", value(q, "permissionCode"))
	return paginate(db.Order("m.parent_id,m.sort_order,m.id"), q, &rows)
}
func validateMenu(input MenuInput, creating bool) error {
	if creating {
		if err := validateCode(input.Code); err != nil {
			return err
		}
	}
	if input.NameZh == "" || input.NameEn == "" {
		return NewError("MENU_NAME_REQUIRED", "menu names are required")
	}
	switch input.MenuType {
	case "DIRECTORY", "MENU", "BUTTON", "EXTERNAL_LINK":
	default:
		return NewError("INVALID_MENU_TYPE", "invalid menu type")
	}
	if input.MenuType == "EXTERNAL_LINK" {
		if err := validateSafeURL(input.ExternalLink); err != nil {
			return err
		}
	} else if input.ExternalLink != "" {
		return NewError("INVALID_EXTERNAL_LINK", "external link is only allowed for external-link menus")
	}
	return nil
}
func (s *Service) ensureNoMenuCycle(id uint64, parentID *uint64) error {
	if parentID == nil {
		return nil
	}
	if id == *parentID {
		return NewError("MENU_CYCLE", "menu cannot be its own parent")
	}
	current := parentID
	for depth := 0; current != nil && depth < 100; depth++ {
		if *current == id {
			return NewError("MENU_CYCLE", "menu cannot be moved under its descendant")
		}
		var row struct {
			ParentID *uint64 `gorm:"column:parent_id"`
		}
		result := s.db.Table("game_admin_menus").Select("parent_id").Where("id=? AND deleted_at IS NULL", *current).Take(&row)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return rowNotFound("menu")
		}
		current = row.ParentID
	}
	if current != nil {
		return NewError("MENU_DEPTH_EXCEEDED", "menu hierarchy is too deep")
	}
	return nil
}
func menuMap(input MenuInput, creating bool) map[string]any {
	row := map[string]any{"parent_id": input.ParentID, "menu_type": input.MenuType, "name_zh": input.NameZh, "name_en": input.NameEn, "route_path": strings.TrimSpace(input.RoutePath), "component_path": strings.TrimSpace(input.ComponentPath), "icon": input.Icon, "permission_code": strings.TrimSpace(input.PermissionCode), "sort_order": input.SortOrder, "visible": boolValue(boolDefault(input.Visible, true)), "enabled": boolValue(boolDefault(input.Enabled, true)), "cacheable": boolValue(boolDefault(input.Cacheable, false)), "external_link": strings.TrimSpace(input.ExternalLink), "updated_at": now()}
	if creating {
		row["code"] = input.Code
		row["created_at"] = now()
	}
	return row
}
func (s *Service) CreateMenu(input MenuInput, actor Actor) (uint64, error) {
	if err := validateMenu(input, true); err != nil {
		return 0, err
	}
	if err := s.ensureNoMenuCycle(0, input.ParentID); err != nil {
		return 0, err
	}
	row := menuMap(input, true)
	if err := s.db.Table("game_admin_menus").Create(row).Error; err != nil {
		return 0, duplicateError(err)
	}
	id := row["id"].(uint64)
	if err := audit(s.db, actor, "system:menu:create", "system.menu.create", "menu", id, nil, row, ""); err != nil {
		return 0, err
	}
	return id, nil
}
func (s *Service) UpdateMenu(id uint64, input MenuInput, reason string, actor Actor) error {
	if err := validateMenu(input, false); err != nil {
		return err
	}
	if err := s.ensureNoMenuCycle(id, input.ParentID); err != nil {
		return err
	}
	var before map[string]any
	if err := s.db.Table("game_admin_menus").Where("id=? AND deleted_at IS NULL", id).Take(&before).Error; err != nil {
		return rowNotFound("menu")
	}
	updates := menuMap(input, false)
	delete(updates, "enabled")
	updates["version"] = gorm.Expr("version+1")
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_admin_menus").Where("id=?", id).Updates(updates).Error; err != nil {
			return duplicateError(err)
		}
		return audit(tx, actor, "system:menu:update", "system.menu.update", "menu", id, before, updates, reason)
	})
}

func (s *Service) MenuStatus(id uint64, enabled bool, reason string, actor Actor) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	var row struct {
		Enabled         bool
		SystemProtected bool
	}
	if err := s.db.Table("game_admin_menus").Select("enabled,system_protected").Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return rowNotFound("menu")
	}
	if row.SystemProtected && !enabled {
		return NewError("PROTECTED_MENU", "system-protected menu cannot be disabled")
	}
	if row.Enabled == enabled {
		return NewError("INVALID_STATUS_TRANSITION", "menu already has requested status")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_admin_menus").Where("id=?", id).Updates(map[string]any{"enabled": boolValue(enabled), "version": gorm.Expr("version+1"), "updated_at": now()}).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:menu:update", "system.menu.status", "menu", id, row, map[string]any{"enabled": enabled}, reason)
	})
}
func (s *Service) CopyMenu(id uint64, input MenuInput, actor Actor) (uint64, error) {
	if err := requireReason(input.Reason); err != nil {
		return 0, err
	}
	if err := validateMenu(input, true); err != nil {
		return 0, err
	}
	var count int64
	if err := s.db.Table("game_admin_menus").Where("id=? AND deleted_at IS NULL", id).Count(&count).Error; err != nil || count == 0 {
		return 0, rowNotFound("menu")
	}
	newID, err := s.CreateMenu(input, actor)
	if err == nil {
		_ = audit(s.db, actor, "system:menu:create", "system.menu.copy", "menu", newID, map[string]any{"sourceMenuID": id}, input, input.Reason)
	}
	return newID, err
}
func (s *Service) DeleteMenu(id uint64, reason string, actor Actor) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	var row struct{ SystemProtected bool }
	if err := s.db.Table("game_admin_menus").Select("system_protected").Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return rowNotFound("menu")
	}
	if row.SystemProtected {
		return NewError("PROTECTED_MENU", "system-protected menu cannot be deleted")
	}
	var children, assigned int64
	_ = s.db.Table("game_admin_menus").Where("parent_id=? AND deleted_at IS NULL", id).Count(&children).Error
	_ = s.db.Table("game_admin_role_menus").Where("menu_id=? AND deleted_at IS NULL", id).Count(&assigned).Error
	if children > 0 || assigned > 0 {
		return NewError("MENU_IN_USE", "menu has children or role assignments")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_admin_menus").Where("id=?", id).Update("deleted_at", now()).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:menu:delete", "system.menu.delete", "menu", id, row, nil, reason)
	})
}
func (s *Service) MenuDetail(id uint64) (map[string]any, error) {
	var row MenuRow
	if err := s.db.Table("game_admin_menus m").Select("m.id,m.parent_id,COALESCE(p.name_en,'') parent,m.menu_type,m.code,m.name_zh,m.name_en,m.route_path,m.component_path,m.icon,m.permission_code,m.sort_order,m.visible,m.enabled,m.cacheable,m.external_link,m.system_protected,m.updated_at").
		Joins("LEFT JOIN game_admin_menus p ON p.id=m.parent_id AND p.deleted_at IS NULL").Where("m.id=? AND m.deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return nil, rowNotFound("menu")
	}
	roles := []map[string]any{}
	_ = s.db.Table("game_admin_role_menus rm").Select("r.id,r.code,r.name_zh nameZh,r.name_en nameEn").Joins("JOIN game_admin_roles r ON r.id=rm.role_id AND r.deleted_at IS NULL").Where("rm.menu_id=? AND rm.deleted_at IS NULL", id).Find(&roles).Error
	return map[string]any{"overview": row, "assignedRoles": roles, "operationLogs": s.operationLogs("menu", id)}, nil
}
