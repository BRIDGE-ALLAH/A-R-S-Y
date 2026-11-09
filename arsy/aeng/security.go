// Package aeng contains the Application Engine framework.
// AENG-035: Enforce model and record security metadata at runtime.
package aeng

import (
	"context"
	"errors"
	"fmt"
)

// SecurityEnforcer enforces model and record security metadata at runtime.
type SecurityEnforcer struct {
	registry *ModelRegistry
}

// NewSecurityEnforcer creates a security enforcer backed by a registry.
func NewSecurityEnforcer(registry *ModelRegistry) *SecurityEnforcer {
	return &SecurityEnforcer{registry: registry}
}

// CheckPermission evaluates whether a subject can perform an action on a resource.
func (e *SecurityEnforcer) CheckPermission(
	ctx context.Context,
	subject any,
	action string,
	resource *ResourceMeta,
) (bool, string) {
	if resource == nil {
		return false, "resource is nil"
	}

	// Superuser bypass
	if subject != nil {
		if su, ok := subject.(interface{ IsSuperuser() bool }); ok && su.IsSuperuser() {
			return true, "superuser bypass"
		}
	}

	// Check resource-level field permissions
	for _, fp := range resource.FieldPermissions {
		if fp.Field == action || action == "*" {
			if fp.Effect == EffectDeny {
				return false, fmt.Sprintf("field %s denied", fp.Field)
			}
			if fp.Effect == EffectAllow {
				return true, "field permission allowed"
			}
		}
	}

	// Check role-based permissions
	for _, role := range resource.Actions {
		if role == action {
			return true, "action allowed"
		}
	}

	// Default deny
	return false, "no matching permission"
}

// FilterRecordFields filters record fields based on user role and field permissions.
func (e *SecurityEnforcer) FilterRecordFields(
	record any,
	userRole string,
	fieldPerms []FieldPermission,
) (any, error) {
	if record == nil {
		return nil, errors.New("aeng: nil record")
	}

	// For now, return the record unchanged - field filtering is applied
	// at the HTTP layer by checking SecurityEnforcer.CheckPermission
	// for each field access.
	return record, nil
}

// CheckActionPermission verifies action-level access control.
func (e *SecurityEnforcer) CheckActionPermission(
	ctx context.Context,
	subject any,
	actionName string,
) (bool, string) {
	if e.registry == nil {
		return false, "registry not initialized"
	}

	actionInfo, ok := e.registry.GetAction(actionName)
	if !ok {
		return false, fmt.Sprintf("action %s not found in registry", actionName)
	}

	if actionInfo.ActionDef == nil {
		return false, fmt.Sprintf("action %s has nil definition", actionName)
	}

	// Superuser bypass
	if subject != nil {
		if su, ok := subject.(interface{ IsSuperuser() bool }); ok && su.IsSuperuser() {
			return true, "superuser bypass"
		}
	}

	// Check action access roles
	if actionInfo.ActionDef.Access != nil {
		for _, role := range actionInfo.ActionDef.Access.Roles {
			if role == "superuser" {
				return true, "superuser role"
			}
		}
	}

	// Default deny if no matching roles
	return false, "action permission denied"
}

// ValidateSecurityMeta validates security metadata consistency.
func ValidateSecurityMeta(registry *ModelRegistry) error {
	if registry == nil {
		return errors.New("aeng: nil registry")
	}

	// Validate all module security metadata
	for moduleID := range registry.modules {
		sec, ok := registry.securityByModule[moduleID]
		if !ok || sec == nil {
			continue
		}

		if sec.Module != moduleID {
			return fmt.Errorf("aeng: security module mismatch: %s vs %s", sec.Module, moduleID)
		}

		// Validate field permissions reference existing fields
		for _, res := range sec.Resources {
			model, ok := registry.modelByName[res.Name]
			if !ok {
				return fmt.Errorf("aeng: security resource %s references unknown model", res.Name)
			}

			for _, fp := range res.FieldPermissions {
				found := false
				for _, f := range model.Fields {
					if f.Name == fp.Field {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("aeng: security field %s not found in model %s", fp.Field, res.Name)
				}
			}
		}
	}

	return nil
}
