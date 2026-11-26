// Package singlesite binds one deployment to one H5 site. The IDs are legacy
// storage keys, never a client-selectable tenant or a deployment mode switch.
package singlesite

import (
	"strings"
	"sync/atomic"

	"casino/internal/tenancy"
)

var current atomic.Pointer[tenancy.Scope]

func Scope() (tenancy.Scope, bool) {
	value := current.Load()
	if value == nil {
		return tenancy.Scope{}, false
	}
	return *value, true
}
func Active() bool { _, ok := Scope(); return ok }

// Bind is called only after the database initializer validates the singleton.
func Bind(scope tenancy.Scope) {
	if scope.TenantID == 0 || scope.SiteID == 0 {
		panic("single-site storage keys must be positive")
	}
	scope.Platform = false
	current.Store(&scope)
}

// Restore supports isolated package tests without changing production settings.
func Restore(scope tenancy.Scope, active bool) {
	if active {
		Bind(scope)
	} else {
		current.Store(nil)
	}
}

func SaaSPath(path string) bool {
	for _, prefix := range []string{"/admin/tenants", "/admin/system/tenants", "/admin/system/tenant-packages", "/admin/system/tenant-deliveries", "/admin/system/platform/dashboard"} {
		if strings.HasSuffix(path, prefix) || strings.Contains(path, prefix+"/") {
			return true
		}
	}
	return false
}

// ControlResource retains permission-controlled administration of accounts,
// roles, country dictionaries and system settings within this one installation.
func ControlResource(path string) bool {
	for _, prefix := range []string{"admins", "roles", "permissions", "departments", "menus", "configs", "countries", "tasks", "events", "monitoring", "popup-components", "backups"} {
		base := "/admin/system/" + prefix
		if strings.HasSuffix(path, base) || strings.Contains(path, base+"/") {
			return true
		}
	}
	return false
}

// CatalogResource manages the upstream provider catalog; H5 game selection and
// all player/business data continue to use the bound site's scope.
func CatalogResource(path string) bool {
	for _, prefix := range []string{"game-providers", "game-sync", "game-sync-logs", "game-category-mapping"} {
		base := "/admin/" + prefix
		if strings.HasSuffix(path, base) || strings.Contains(path, base+"/") {
			return true
		}
	}
	return false
}
