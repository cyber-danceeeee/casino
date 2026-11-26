package tenancy

import "gorm.io/gorm"

type Scope struct {
	TenantID uint64
	SiteID   uint64
	Platform bool
}

func (s Scope) TenantScoped() bool {
	return !s.Platform && s.TenantID > 0
}

func (s Scope) SiteScoped() bool {
	return s.TenantScoped() && s.SiteID > 0
}

func ApplyTenant(db *gorm.DB, alias string, scope Scope) *gorm.DB {
	if !scope.TenantScoped() {
		return db
	}
	return db.Where(column(alias, "tenant_id")+" = ?", scope.TenantID)
}

func ApplySite(db *gorm.DB, alias string, scope Scope) *gorm.DB {
	db = ApplyTenant(db, alias, scope)
	if scope.SiteScoped() {
		db = db.Where(column(alias, "site_id")+" IN (0, ?)", scope.SiteID)
	}
	return db
}

func ApplySiteStrict(db *gorm.DB, alias string, scope Scope) *gorm.DB {
	db = ApplyTenant(db, alias, scope)
	if scope.SiteScoped() {
		db = db.Where(column(alias, "site_id")+" = ?", scope.SiteID)
	}
	return db
}

func column(alias string, name string) string {
	if alias == "" {
		return name
	}
	return alias + "." + name
}
