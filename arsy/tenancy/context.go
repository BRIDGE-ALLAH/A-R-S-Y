 package tenancy

 import "context"

 type contextKey struct{}

 // NewContext returns a derived context carrying t.
 func NewContext(parent context.Context, t Tenant) context.Context {
 	return context.WithValue(parent, contextKey{}, t)
 }

 // FromContext extracts the Tenant from ctx.
 // ok is false if no tenant has been set.
 func FromContext(ctx context.Context) (Tenant, bool) {
 	t, ok := ctx.Value(contextKey{}).(Tenant)
 	return t, ok
 }

 // MustFromContext extracts the Tenant from ctx and panics if none is set.
 // Use only in code paths that are guaranteed to run inside tenant-authenticated middleware.
 func MustFromContext(ctx context.Context) Tenant {
 	t, ok := FromContext(ctx)
 	if !ok {
 		panic("tenancy: no tenant in context")
 	}
 	return t
 }
