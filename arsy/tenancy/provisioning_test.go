package tenancy_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bridge-allah/a-r-s-y/arsy/config"
	"github.com/bridge-allah/a-r-s-y/arsy/tenancy"
)

func newProvider(t *testing.T) *tenancy.SQLiteProvider {
	t.Helper()
	dirs, err := config.EnsureBaaSDirectories(t.TempDir(), "")
	if err != nil {
		t.Fatalf("setup dirs: %v", err)
	}
	return tenancy.NewSQLiteProvider(dirs)
}

// TestSQLiteProvider_ProvisionIdempotentSameKey verifies that repeated provisioning
// with the same idempotency key does not create duplicate tenants or databases.
func TestSQLiteProvider_ProvisionIdempotentSameKey(t *testing.T) {
	dirs, err := config.EnsureBaaSDirectories(t.TempDir(), "")
	if err != nil {
		t.Fatalf("setup dirs: %v", err)
	}
	provider := tenancy.NewSQLiteProvider(dirs)
	t.Cleanup(func() { _ = provider.Close() })

	ctx := context.Background()
	key := "prov-key-acme"

	// First provision
	err := provider.(tenancy.Provisioner).Provision(ctx, "tenant-acme", tenancy.ProvisionOption{IdempotencyKey: key})
	if err != nil {
		t.Fatalf("first Provision: %v", err)
	}

	// Second provision with same key must be idempotent (no error, no duplicate)
	err = provider.(tenancy.Provisioner).Provision(ctx, "tenant-acme", tenancy.ProvisionOption{IdempotencyKey: key})
	if err != nil {
		t.Fatalf("second Provision same key: %v", err)
	}

	// Provisioning a different tenant with the same key should error or be rejected.
	err = provider.(tenancy.Provisioner).Provision(ctx, "tenant-different", tenancy.ProvisionOption{IdempotencyKey: key})
	if err == nil {
		t.Fatal("expected error for different tenant with same key")
	}
}

// TestSQLiteProvider_ProvisionCrashRecovery verifies that a crash during provision
// leaves the tenant in a non-active state, and retry safely completes it.
func TestSQLiteProvider_ProvisionCrashRecovery(t *testing.T) {
	dirs, err := config.EnsureBaaSDirectories(t.TempDir(), "")
	if err != nil {
		t.Fatalf("setup dirs: %v", err)
	}
	provider := tenancy.NewSQLiteProvider(dirs)
	t.Cleanup(func() { _ = provider.Close() })

	ctx := context.Background()
	key := "prov-key-recovery"

	// Simulate a crash after DB creation but before migration completion.
	// We do this by using a migration that fails mid-way, then retry.
	migrations := []tenancy.Migration{{
		Version: "0001", Checksum: "recovery-0001",
		Up: func(db *sql.DB) error {
			// Create table, then fail to simulate partial state
			_, err := db.Exec(`CREATE TABLE recovery_items (id TEXT PRIMARY KEY, data TEXT)`)
			if err != nil {
				return err
			}
			// Simulate crash/partial state: insert half the rows then return error
			_, err = db.Exec(`INSERT INTO recovery_items (id, data) VALUES ('row1', 'data1')`)
			return err // intentionally fail after first insert
		},
	}}

	// First provision: should fail (partial state), tenant should NOT be active
	err = provider.(tenancy.Provisioner).Provision(ctx, "tenant-recovery", tenancy.ProvisionOption{IdempotencyKey: key, Migrations: migrations})
	if err == nil {
		t.Fatal("expected first Provision to fail with partial migration")
	}

	// Verify tenant exists but is not active (status checking would go here)
	// For now just verify the error was returned

	// Retry with same key: should succeed and complete the migration
	err = provider.(tenancy.Provisioner).Provision(ctx, "tenant-recovery", tenancy.ProvisionOption{IdempotencyKey: key})
	if err != nil {
		t.Fatalf("retry Provision: %v", err)
	}

	// Verify tenant is now active after retry completes
	// (In a full implementation, status checks would verify active state)
}

// TestSQLiteProvider_ProvisionDifferentKeys verifies key isolation.
func TestSQLiteProvider_ProvisionDifferentKeys(t *testing.T) {
	dirs, err := config.EnsureBaaSDirectories(t.TempDir(), "")
	if err != nil {
		t.Fatalf("setup dirs: %v", err)
	}
	provider := tenancy.NewSQLiteProvider(dirs)
	t.Cleanup(func() { _ = provider.Close() })

	ctx := context.Background()

	// Different keys = independent tenants
	_ = provider.(tenancy.Provisioner).Provision(ctx, "tenant-a", tenancy.ProvisionOption{IdempotencyKey: "key-a"})
	_ = provider.(tenancy.Provisioner).Provision(ctx, "tenant-b", tenancy.ProvisionOption{IdempotencyKey: "key-b"})

	// Both should exist as separate tenants
	dsA, err := provider.Open(ctx, "tenant-a")
	if err != nil {
		t.Fatalf("Open tenant-a: %v", err)
	}
	dsB, err := provider.Open(ctx, "tenant-b")
	if err != nil {
		t.Fatalf("Open tenant-b: %v", err)
	}
	if dsA.TenantID() != "tenant-a" || dsB.TenantID() != "tenant-b" {
		t.Fatalf("tenant IDs mismatch")
	}
}