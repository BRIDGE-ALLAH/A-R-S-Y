package pgadapter_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bridge-allah/a-r-s-y/arsy/tenancy"
	"github.com/bridge-allah/a-r-s-y/arsy/tenancy/pgadapter"
)

// TestPostgresProvider_ConfigValidation verifies the adapter rejects bad config
// without touching a database.
func TestPostgresProvider_ConfigValidation(t *testing.T) {
	if _, err := pgadapter.NewPostgresProvider(pgadapter.Config{}); !errors.Is(err, pgadapter.ErrPGDSNRequired) {
		t.Fatalf("NewPostgresProvider error = %v, want ErrPGDSNRequired", err)
	}
	if _, err := pgadapter.NewPostgresProvider(pgadapter.Config{DSN: "postgres://x"}); err != nil {
		t.Fatalf("NewPostgresProvider with DSN: %v", err)
	}
}

// TestPostgresProvider_ImplementsContract is a runtime assertion that the
// adapter satisfies the engine-neutral tenancy contract. It does not require
// a running PostgreSQL service.
func TestPostgresProvider_ImplementsContract(t *testing.T) {
	p, err := pgadapter.NewPostgresProvider(pgadapter.Config{DSN: "postgres://x"})
	if err != nil {
		t.Fatalf("NewPostgresProvider: %v", err)
	}
	var provider tenancy.Provider = p
	var checker tenancy.HealthChecker = p

	ctx := context.Background()
	if _, err := provider.Open(ctx, ""); !errors.Is(err, pgadapter.ErrPGInvalidTenantID) {
		t.Fatalf("Open error = %v, want ErrPGInvalidTenantID", err)
	}
	if err := provider.CloseTenant(ctx, "unknown"); err != nil {
		t.Fatalf("CloseTenant unknown: %v", err)
	}
	if err := provider.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, err := provider.Open(ctx, "after-close"); !errors.Is(err, pgadapter.ErrPGClosed) {
		t.Fatalf("Open after Close error = %v, want ErrPGClosed", err)
	}
	if status := checker.Health(ctx, "closed"); status.Healthy {
		t.Fatal("Health after Close must not be healthy")
	}
	_ = time.Now
}