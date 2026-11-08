package tenancy_test

import (
	"context"
	"testing"

	"github.com/bridge-allah/a-r-s-y/arsy/config"
	"github.com/bridge-allah/a-r-s-y/arsy/tenancy"
)

func TestCheckTenant_Healthy(t *testing.T) {
	root := t.TempDir()
	dirs, _ := config.EnsureBaaSDirectories(root, "")
	provider := tenancy.NewSQLiteProvider(dirs)
	defer provider.Close()
	ctx := context.Background()
	if _, err := provider.Open(ctx, "health-t1"); err != nil {
		t.Fatal(err)
	}
	report := tenancy.CheckTenant(ctx, provider, dirs, "health-t1")
	if !report.Healthy || !report.ConnectionOK {
		t.Fatalf("expected healthy report: %+v", report)
	}
	if report.TenantID != "health-t1" {
		t.Fatalf("tenant id mismatch: %+v", report)
	}
	if report.Error != "" {
		t.Fatalf("unexpected error: %+v", report)
	}
}

func TestCheckTenant_UnhealthyMissing(t *testing.T) {
	root := t.TempDir()
	dirs, _ := config.EnsureBaaSDirectories(root, "")
	provider := tenancy.NewSQLiteProvider(dirs)
	defer provider.Close()
	ctx := context.Background()
	// Open+close a DB file path won't exist in a fresh provider state for this id except Open creates it;
	// use an invalid provider state by closing provider to force unhealthy.
	provider.Close()
	report := tenancy.CheckTenant(ctx, provider, dirs, "missing-t1")
	if report.Healthy || report.ConnectionOK {
		t.Fatalf("expected unhealthy report: %+v", report)
	}
}
