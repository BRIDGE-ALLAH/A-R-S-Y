package tenancy

import (
	"context"
	"fmt"
)

// Router resolves the Datastore for the tenant stored in ctx.
// It combines a Catalog (to validate tenant state) and a Provider (to open the DB).
type Router struct {
	catalog  Catalog
	provider Provider
}

// NewRouter creates a Router from catalog and provider.
func NewRouter(catalog Catalog, provider Provider) *Router {
	return &Router{catalog: catalog, provider: provider}
}

// Datastore returns the tenant Datastore for the tenant in ctx.
// The tenant state is re-read from the catalog on every call so a stale
// request context cannot keep access to a suspended or deleted tenant.
// Returns ErrTenantNotFound, ErrTenantSuspended, ErrTenantDeleting or
// ErrTenantNotActive as appropriate.
func (r *Router) Datastore(ctx context.Context) (Datastore, error) {
	t, ok := FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenancy: no tenant in context")
	}
	if r.catalog != nil {
		current, err := r.catalog.Get(ctx, t.ID)
		if err != nil {
			return nil, err
		}
		t = current
	}
	switch t.Status {
	case StatusActive:
		return r.provider.Open(ctx, t.ID)
	case StatusSuspended:
		return nil, ErrTenantSuspended
	case StatusDeleting:
		return nil, ErrTenantDeleting
	default:
		return nil, ErrTenantNotActive
	}
}
