package operationadmin

import (
	"gorm.io/gorm"

	"casino/internal/operationdefaults"
)

type DefaultVisibilityInput struct {
	Enabled bool   `json:"enabled"`
	Reason  string `json:"reason" binding:"max=500"`
}

func (s *Service) SetDefaultVisibility(resource string, id uint64, enabled bool, reason string, actor Actor) error {
	resource, err := operationdefaults.NormalizeResource(resource)
	if err != nil {
		return newError("INVALID_DEFAULT_RESOURCE", err.Error())
	}
	if !actor.Scope.SiteScoped() {
		return newError("TENANT_SITE_SCOPE_REQUIRED", "tenant site scope is required")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := operationdefaults.SetVisibility(tx, actor.Scope, resource, id, enabled, actor.ID); err != nil {
			if err == gorm.ErrRecordNotFound {
				return newError("OPERATION_DEFAULT_NOT_FOUND", "platform operation default was not found")
			}
			return err
		}
		return audit(tx, actor, "operation:default:visibility", "visibility_update", "operation_default", id,
			map[string]any{"resourceType": resource}, map[string]any{"resourceType": resource, "enabled": enabled}, reason)
	})
}
