package systemadmin

import (
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

type BrandDeliveryRow struct {
	ID                   uint64    `json:"id"`
	TenantID             uint64    `json:"tenantID" gorm:"column:tenant_id"`
	TenantName           string    `json:"tenantName" gorm:"column:tenant_name"`
	TenantCode           string    `json:"tenantCode" gorm:"column:tenant_code"`
	SiteID               uint64    `json:"siteID" gorm:"column:site_id"`
	SiteName             string    `json:"siteName" gorm:"column:site_name"`
	AdminID              uint64    `json:"adminID" gorm:"column:admin_id"`
	AdminUsername        string    `json:"adminUsername" gorm:"column:admin_username"`
	PrimaryDomain        string    `json:"primaryDomain" gorm:"column:primary_domain"`
	H5URL                string    `json:"h5Url" gorm:"column:h5_url"`
	AdminURL             string    `json:"adminUrl" gorm:"column:admin_url"`
	DomainStatus         string    `json:"domainStatus" gorm:"column:domain_status"`
	InitialConfigVersion uint64    `json:"initialConfigVersion" gorm:"column:initial_config_version"`
	Status               string    `json:"status"`
	DeliveredBy          uint64    `json:"deliveredBy" gorm:"column:delivered_by"`
	DeliveredByName      string    `json:"deliveredByName" gorm:"column:delivered_by_name"`
	DeliveredAt          time.Time `json:"deliveredAt" gorm:"column:delivered_at"`
}

type PlatformFinanceSummary struct {
	Currency   string `json:"currency"`
	TotalBet   string `json:"totalBet" gorm:"column:total_bet"`
	TotalWin   string `json:"totalWin" gorm:"column:total_win"`
	GGR        string `json:"ggr"`
	NGR        string `json:"ngr"`
	Deposit    string `json:"deposit" gorm:"column:deposit_amount"`
	Withdrawal string `json:"withdrawal" gorm:"column:withdrawal_amount"`
}

type TenantPackageRow struct {
	ID          uint64     `json:"id"`
	Code        string     `json:"code" gorm:"column:package_code"`
	NameZh      string     `json:"nameZh" gorm:"column:name_zh"`
	NameEn      string     `json:"nameEn" gorm:"column:name_en"`
	Sites       int        `json:"sites" gorm:"column:site_limit"`
	Games       int        `json:"games" gorm:"column:game_limit"`
	Providers   int        `json:"providers" gorm:"column:provider_limit"`
	StorageGB   int        `json:"storageGB" gorm:"column:storage_gb"`
	Players     int64      `json:"players" gorm:"column:player_limit"`
	Status      string     `json:"status"`
	SortOrder   int        `json:"sortOrder" gorm:"column:sort_order"`
	TenantCount int64      `json:"tenantCount" gorm:"column:tenant_count"`
	CreatedAt   time.Time  `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt   time.Time  `json:"updatedAt" gorm:"column:updated_at"`
	DeletedAt   *time.Time `json:"-" gorm:"column:deleted_at"`
}

type TenantPackageInput struct {
	Code      string `json:"code"`
	NameZh    string `json:"nameZh"`
	NameEn    string `json:"nameEn"`
	Sites     int    `json:"sites"`
	Games     int    `json:"games"`
	Providers int    `json:"providers"`
	StorageGB int    `json:"storageGB"`
	Players   int64  `json:"players"`
	Status    string `json:"status"`
	SortOrder int    `json:"sortOrder"`
	Reason    string `json:"reason"`
}

func (s *Service) ListBrandDeliveries(q Query, actor Actor) (Result, error) {
	if !actor.Platform {
		return Result{}, NewError("PERMISSION_DENIED", "permission denied")
	}
	rows := []BrandDeliveryRow{}
	selectColumns := `d.id,d.tenant_id,COALESCE(t.tenant_name,'') tenant_name,COALESCE(t.tenant_code,'') tenant_code,d.site_id,COALESCE(s.site_name,'') site_name,d.admin_id,d.admin_username,d.primary_domain,d.h5_url,d.admin_url,d.domain_status,d.initial_config_version,d.status,d.delivered_by,COALESCE(a.username,'') delivered_by_name,d.delivered_at`
	db := s.db.Table("game_saas_brand_deliveries d").
		Select(selectColumns).
		Joins("LEFT JOIN game_saas_tenants t ON t.id=d.tenant_id AND t.deleted_at IS NULL").
		Joins("LEFT JOIN game_saas_sites s ON s.id=d.site_id AND s.deleted_at IS NULL").
		Joins("LEFT JOIN game_admin_accounts a ON a.id=d.delivered_by AND a.deleted_at IS NULL").
		Where("d.deleted_at IS NULL")
	hasLiveDomain := s.db.Migrator().HasColumn("game_saas_domains", "domain") && s.db.Migrator().HasColumn("game_saas_domains", "status")
	if hasLiveDomain {
		db = db.Select(strings.Replace(selectColumns, "d.domain_status", "COALESCE(dm.status,d.domain_status) domain_status", 1)).
			Joins("LEFT JOIN game_saas_domains dm ON dm.tenant_id=d.tenant_id AND dm.site_id=d.site_id AND dm.domain=d.primary_domain AND dm.deleted_at IS NULL")
	}
	db = exact(db, "d.tenant_id", value(q, "tenantID"))
	db = like(db, "t.tenant_name", value(q, "tenantName"))
	db = like(db, "t.tenant_code", value(q, "tenantCode"))
	db = like(db, "d.admin_username", value(q, "adminUsername"))
	db = like(db, "d.primary_domain", value(q, "domain"))
	if status := strings.ToUpper(value(q, "domainStatus")); status != "" {
		if hasLiveDomain {
			db = db.Where("COALESCE(dm.status,d.domain_status)=?", status)
		} else {
			db = db.Where("d.domain_status=?", status)
		}
	}
	db = exact(db, "d.status", strings.ToUpper(value(q, "status")))
	db = timeRange(db, "d.delivered_at", value(q, "createdFrom"), value(q, "createdTo"))
	return paginate(db.Order("d.delivered_at DESC,d.id DESC"), q, &rows)
}

func (s *Service) TenantPackages(actor Actor) (map[string]any, error) {
	if !actor.Platform {
		return nil, NewError("PERMISSION_DENIED", "permission denied")
	}
	if !s.db.Migrator().HasTable("game_saas_tenant_packages") {
		return map[string]any{"list": []TenantPackageRow{}}, nil
	}
	rows := []TenantPackageRow{}
	err := s.db.Table("game_saas_tenant_packages p").
		Select(`p.id,p.package_code,p.name_zh,p.name_en,p.site_limit,p.game_limit,p.provider_limit,p.storage_gb,p.player_limit,p.status,p.sort_order,p.created_at,p.updated_at,
			(SELECT COUNT(*) FROM game_saas_tenants t WHERE t.package_code=p.package_code AND t.deleted_at IS NULL) tenant_count`).
		Where("p.deleted_at IS NULL").Order("p.sort_order ASC,p.id ASC").Find(&rows).Error
	return map[string]any{"list": rows}, err
}

func normalizeTenantPackageInput(input *TenantPackageInput, creating bool) error {
	input.Code = strings.ToUpper(strings.TrimSpace(input.Code))
	input.NameZh = strings.TrimSpace(input.NameZh)
	input.NameEn = strings.TrimSpace(input.NameEn)
	input.Status = strings.ToUpper(strings.TrimSpace(defaultString(input.Status, "AVAILABLE")))
	if input.Code == "" || input.NameZh == "" || input.NameEn == "" {
		return NewError("PACKAGE_REQUIRED", "package code and names are required")
	}
	if creating && !regexpSaaSPackage.MatchString(input.Code) {
		return NewError("INVALID_PACKAGE_CODE", "invalid package code")
	}
	if !map[string]bool{"AVAILABLE": true, "CONTACT": true, "LEGACY": true, "DISABLED": true}[input.Status] {
		return NewError("INVALID_PACKAGE_STATUS", "invalid package status")
	}
	if !creating {
		return requireReason(input.Reason)
	}
	return nil
}

var regexpSaaSPackage = regexp.MustCompile(`^[A-Z][A-Z0-9_]{1,63}$`)

func tenantPackageValues(input TenantPackageInput, actorID uint64) map[string]any {
	return map[string]any{"package_code": input.Code, "name_zh": input.NameZh, "name_en": input.NameEn,
		"site_limit": input.Sites, "game_limit": input.Games, "provider_limit": input.Providers,
		"storage_gb": input.StorageGB, "player_limit": input.Players, "status": input.Status,
		"sort_order": input.SortOrder, "updated_by": actorID, "updated_at": now()}
}

func (s *Service) CreateTenantPackage(input TenantPackageInput, actor Actor) (uint64, error) {
	if !actor.Platform {
		return 0, NewError("PERMISSION_DENIED", "permission denied")
	}
	if err := normalizeTenantPackageInput(&input, true); err != nil {
		return 0, err
	}
	values := tenantPackageValues(input, actor.ID)
	values["created_by"], values["created_at"] = actor.ID, now()
	var id uint64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_saas_tenant_packages").Create(values).Error; err != nil {
			return duplicateError(err)
		}
		if err := tx.Table("game_saas_tenant_packages").Select("id").Where("package_code=? AND deleted_at IS NULL", input.Code).Take(&id).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:tenant:update", "system.tenant_package.create", "tenant_package", id, nil, values, input.Reason)
	})
	return id, err
}

func (s *Service) UpdateTenantPackage(id uint64, input TenantPackageInput, actor Actor) error {
	if !actor.Platform {
		return NewError("PERMISSION_DENIED", "permission denied")
	}
	if err := normalizeTenantPackageInput(&input, false); err != nil {
		return err
	}
	var before TenantPackageRow
	if err := s.db.Table("game_saas_tenant_packages").Where("id=? AND deleted_at IS NULL", id).Take(&before).Error; err != nil {
		return rowNotFound("tenant package")
	}
	values := tenantPackageValues(input, actor.ID)
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_saas_tenant_packages").Where("id=? AND deleted_at IS NULL", id).Updates(values).Error; err != nil {
			return duplicateError(err)
		}
		if before.Code != input.Code {
			if err := tx.Table("game_saas_tenants").Where("package_code=? AND deleted_at IS NULL", before.Code).Update("package_code", input.Code).Error; err != nil {
				return err
			}
		}
		return audit(tx, actor, "system:tenant:update", "system.tenant_package.update", "tenant_package", id, before, values, input.Reason)
	})
}

func (s *Service) DeleteTenantPackage(id uint64, reason string, actor Actor) error {
	if !actor.Platform {
		return NewError("PERMISSION_DENIED", "permission denied")
	}
	if err := requireReason(reason); err != nil {
		return err
	}
	var before TenantPackageRow
	if err := s.db.Table("game_saas_tenant_packages").Where("id=? AND deleted_at IS NULL", id).Take(&before).Error; err != nil {
		return rowNotFound("tenant package")
	}
	var tenants int64
	if err := s.db.Table("game_saas_tenants").Where("package_code=? AND deleted_at IS NULL", before.Code).Count(&tenants).Error; err != nil {
		return err
	}
	if tenants > 0 {
		return NewError("PACKAGE_IN_USE", "tenant package is in use")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{"deleted_at": now(), "updated_by": actor.ID, "updated_at": now()}
		if err := tx.Table("game_saas_tenant_packages").Where("id=?", id).Updates(updates).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:tenant:update", "system.tenant_package.delete", "tenant_package", id, before, updates, reason)
	})
}

func (s *Service) TenantAudit(tenantID uint64, q Query, actor Actor) (Result, error) {
	if !actor.Platform {
		return Result{}, NewError("PERMISSION_DENIED", "permission denied")
	}
	var tenantCount int64
	if err := s.db.Table("game_saas_tenants").Where("id=? AND deleted_at IS NULL", tenantID).Count(&tenantCount).Error; err != nil {
		return Result{}, err
	}
	if tenantCount == 0 {
		return Result{}, rowNotFound("tenant")
	}
	rows := []AuditLogRow{}
	moduleExpr := "COALESCE(NULLIF(l.module,''),NULLIF(SUBSTRING_INDEX(l.permission,':',1),''),'unknown')"
	if s.db.Dialector.Name() == "sqlite" {
		moduleExpr = "COALESCE(NULLIF(l.module,''),'unknown')"
	}
	db := s.db.Table("game_admin_operation_logs l").Select("l.id,l.operator_id,l.tenant_id,l.site_id,l.operator_username,l.permission,"+moduleExpr+" module,l.action,l.result,l.target_type,l.target_id,l.reason,l.ip_address,l.correlation_id,l.created_at").Where("l.deleted_at IS NULL").
		Where(`(l.tenant_id=? OR (l.target_type='tenant' AND l.target_id=?) OR (l.target_type='site' AND l.target_id IN (SELECT id FROM game_saas_sites WHERE tenant_id=?)) OR (l.target_type='domain' AND l.target_id IN (SELECT id FROM game_saas_domains WHERE tenant_id=?)) OR (l.target_type='administrator' AND l.target_id IN (SELECT id FROM game_admin_accounts WHERE tenant_id=?)))`, tenantID, tenantID, tenantID, tenantID, tenantID)
	db = like(db, "l.operator_username", value(q, "operator"))
	db = like(db, "l.action", value(q, "action"))
	db = exact(db, "l.target_type", value(q, "targetType"))
	db = exact(db, "l.result", value(q, "result"))
	db = exact(db, "l.correlation_id", value(q, "correlationID"))
	db = timeRange(db, "l.created_at", value(q, "createdFrom"), value(q, "createdTo"))
	return paginate(db.Order("l.created_at DESC,l.id DESC"), q, &rows)
}

func (s *Service) PlatformDashboard(actor Actor) (map[string]any, error) {
	if !actor.Platform {
		return nil, NewError("PERMISSION_DENIED", "permission denied")
	}
	var totalTenants, activeTenants, totalSites, totalPlayers, pendingDomains, runningJobs int64
	_ = s.db.Table("game_saas_tenants").Where("deleted_at IS NULL").Count(&totalTenants).Error
	_ = s.db.Table("game_saas_tenants").Where("enabled=1 AND status='ACTIVE' AND deleted_at IS NULL").Count(&activeTenants).Error
	_ = s.db.Table("game_saas_sites").Where("deleted_at IS NULL").Count(&totalSites).Error
	_ = s.db.Table("game_users").Where("deleted_at IS NULL").Count(&totalPlayers).Error
	_ = s.db.Table("game_saas_domains").Where("status='PENDING' AND deleted_at IS NULL").Count(&pendingDomains).Error
	_ = s.db.Table("game_system_tasks").Where("running=1 AND deleted_at IS NULL").Count(&runningJobs).Error

	providers := []map[string]any{}
	_ = s.db.Table("game_game_providers").Select("provider_code providerCode,name,api_status apiStatus,status,last_sync_at lastSyncAt,last_tested_at lastTestedAt").Where("tenant_id=0 AND site_id=0 AND deleted_at IS NULL").Order("sort,id").Find(&providers).Error

	finance := []PlatformFinanceSummary{}
	_ = s.db.Table("game_daily_financial_summaries").
		Select("currency,CAST(SUM(total_bet) AS CHAR) total_bet,CAST(SUM(total_win) AS CHAR) total_win,CAST(SUM(total_bet-total_win) AS CHAR) ggr,CAST(SUM(total_bet-total_win-bonus_cost-provider_cost-payment_fee-agent_commission+adjustments) AS CHAR) ngr,CAST(SUM(deposit_amount) AS CHAR) deposit_amount,CAST(SUM(withdrawal_amount) AS CHAR) withdrawal_amount").
		Where("business_date=? AND deleted_at IS NULL", now().Format("2006-01-02")).Group("currency").Order("currency").Find(&finance).Error

	recent := []TenantRow{}
	_ = s.db.Table("game_saas_tenants t").Select(tenantSelect()).Where("t.deleted_at IS NULL").Order("t.created_at DESC,t.id DESC").Limit(8).Find(&recent).Error
	return map[string]any{
		"sampledAt":         now(),
		"overview":          map[string]any{"totalTenants": totalTenants, "activeTenants": activeTenants, "totalSites": totalSites, "totalPlayers": totalPlayers, "pendingDomains": pendingDomains, "runningJobs": runningJobs},
		"financeByCurrency": finance, "providers": providers, "recentTenants": recent,
	}, nil
}
