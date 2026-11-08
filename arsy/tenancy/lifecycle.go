package tenancy

import (
	"context"
	"fmt"
)

// LifecycleManager handles tenant state transitions.
// State machine: provisioning → active ↔ suspended; active/suspended → deleting → (removed)
type LifecycleManager struct {
	provider Provider
}

// NewLifecycleManager creates a LifecycleManager backed by provider.
func NewLifecycleManager(provider Provider) *LifecycleManager {
	return &LifecycleManager{provider: provider}
}

// validTransition returns true when moving from `from` to `to` is allowed.
func validTransition(from, to Status) bool {
	allowed := map[Status][]Status{
		StatusProvisioning: {StatusActive, StatusFailed},
		StatusActive:       {StatusSuspended, StatusDeleting},
		StatusSuspended:    {StatusActive, StatusDeleting},
		StatusDeleting:     {StatusFailed}, // terminal; only failure is a back-transition
		StatusFailed:       {StatusProvisioning},
	}
	for _, s := range allowed[from] {
		if s == to {
			return true
		}
	}
	return false
}

// ValidateTransition returns nil if transitioning from → to is permitted.
func ValidateTransition(from, to Status) error {
	if !validTransition(from, to) {
		return fmt.Errorf("tenancy: invalid state transition %s → %s", from, to)
	}
	return nil
}

// Suspend marks a tenant as suspended.
// The datastore is not closed; existing connections are preserved.
func (l *LifecycleManager) Suspend(_ context.Context, t *Tenant) error {
	if err := ValidateTransition(t.Status, StatusSuspended); err != nil {
		return err
	}
	t.Status = StatusSuspended
	return nil
}

// Reinstate moves a suspended tenant back to active.
func (l *LifecycleManager) Reinstate(_ context.Context, t *Tenant) error {
	if err := ValidateTransition(t.Status, StatusActive); err != nil {
		return err
	}
	t.Status = StatusActive
	return nil
}

// BeginDeletion marks a tenant for deletion and evicts its datastore handle.
// Eviction is idempotent and safe for unknown tenants.
func (l *LifecycleManager) BeginDeletion(ctx context.Context, t *Tenant) error {
	if err := ValidateTransition(t.Status, StatusDeleting); err != nil {
		return err
	}
	t.Status = StatusDeleting
	if err := l.provider.CloseTenant(ctx, t.ID); err != nil && err != ErrDatastoreClosed {
		return err
	}
	return nil
}
