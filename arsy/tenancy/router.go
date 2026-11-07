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
 // Returns ErrTenantNotFound, ErrTenantSuspended or ErrTenantDeleting as appropriate.
 func (r *Router) Datastore(ctx context.Context) (Datastore, error) {
 	t, ok := FromContext(ctx)
 	if !ok {
 		return nil, fmt.Errorf("tenancy: no tenant in context")
 	}
 	switch t.Status {
 	case StatusSuspended:
 		return nil, ErrTenantSuspended
 	case StatusDeleting:
 		return nil, ErrTenantDeleting
 	}
 	return r.provider.Open(ctx, t.ID)
 }
