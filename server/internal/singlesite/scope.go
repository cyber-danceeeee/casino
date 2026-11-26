// Package singlesite defines the only H5 configuration in this installation.
package singlesite

import (
	"casino/internal/content"
	"strings"
)

func Scope() (content.Scope, bool) { return content.Scope{SiteID: 1}, true }

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
