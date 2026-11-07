 package tenancy_test

 import (
 	"context"
 	"testing"

 	"github.com/bridge-allah/a-r-s-y/arsy/tenancy"
 )

 func TestContextRoundtrip(t *testing.T) {
 	tenant := tenancy.Tenant{ID: "t1", Slug: "acme", Status: tenancy.StatusActive}
 	ctx := tenancy.NewContext(context.Background(), tenant)

 	got, ok := tenancy.FromContext(ctx)
 	if !ok {
 		t.Fatal("expected tenant in context, got false")
 	}
 	if got.ID != tenant.ID || got.Slug != tenant.Slug {
 		t.Errorf("got %+v, want %+v", got, tenant)
 	}
 }

 func TestFromContext_Missing(t *testing.T) {
 	_, ok := tenancy.FromContext(context.Background())
 	if ok {
 		t.Fatal("expected ok=false for empty context")
 	}
 }

 func TestValidateStatus(t *testing.T) {
 	for _, s := range []tenancy.Status{
 		tenancy.StatusProvisioning, tenancy.StatusActive,
 		tenancy.StatusSuspended, tenancy.StatusDeleting, tenancy.StatusFailed,
 	} {
 		if err := tenancy.ValidateStatus(s); err != nil {
 			t.Errorf("valid status %q returned error: %v", s, err)
 		}
 	}
 	if err := tenancy.ValidateStatus("unknown"); err == nil {
 		t.Error("expected error for unknown status")
 	}
 }
