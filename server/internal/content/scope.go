// Package content distinguishes the imported catalog from published H5 content.
package content

import "gorm.io/gorm"

type Scope struct {
	SiteID   uint64
	Platform bool
}

func (s Scope) SiteScoped() bool { return !s.Platform && s.SiteID > 0 }
func ApplySite(db *gorm.DB, alias string, scope Scope) *gorm.DB {
	if scope.SiteScoped() {
		return db.Where(column(alias, "site_id")+" IN (0, ?)", scope.SiteID)
	}
	return db
}
func ApplySiteStrict(db *gorm.DB, alias string, scope Scope) *gorm.DB {
	if scope.SiteScoped() {
		return db.Where(column(alias, "site_id")+" = ?", scope.SiteID)
	}
	return db
}
func column(alias, name string) string {
	if alias == "" {
		return name
	}
	return alias + "." + name
}
