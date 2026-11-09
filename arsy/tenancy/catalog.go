package tenancy

import (
	"context"
	"fmt"
)

// Status mirrors the lifecycle states stored in baas_tenants.status.
type Status string

const (
	StatusProvisioning Status = "provisioning"
	StatusActive       Status = "active"
	StatusSuspended    Status = "suspended"
	StatusDeleting     Status = "deleting"
	StatusFailed       Status = "failed"
)

// Tenant holds a minimal, safe in-memory view of a baas_tenants record.
// It never carries credentials or raw secrets.
type Tenant struct {
	// ID is the immutable PocketBase-assigned record ID.
	ID string
	// Slug is the URL-safe, unique tenant identifier.
	Slug string
	// DisplayName is the human-readable name.
	DisplayName string
	// Status is the current lifecycle state.
	Status Status
	// Plan is an optional subscription/plan identifier.
	Plan string
}

// SecretRef is an opaque reference to an encrypted secret stored in the
// control plane. Callers use SecretRef values to retrieve secrets via a
// SecretResolver; they never hold raw values.
type SecretRef struct {
	// Key uniquely identifies the secret within the control plane.
	Key string
}

// SecretResolver decrypts and returns a secret identified by ref.
// Implementations must not cache decrypted values in memory longer than needed.
type SecretResolver interface {
	Resolve(ctx context.Context, ref SecretRef) ([]byte, error)
}

// Catalog is a read-only view of the tenant registry.
type Catalog interface {
	// Get returns the Tenant for id, or ErrTenantNotFound.
	Get(ctx context.Context, id string) (Tenant, error)
	// GetBySlug returns the Tenant for slug, or ErrTenantNotFound.
	GetBySlug(ctx context.Context, slug string) (Tenant, error)
}

// ValidateStatus returns an error if s is not a known tenant status.
func ValidateStatus(s Status) error {
	switch s {
	case StatusProvisioning, StatusActive, StatusSuspended, StatusDeleting, StatusFailed:
		return nil
	default:
		return fmt.Errorf("tenancy: unknown tenant status %q", s)
	}
}
