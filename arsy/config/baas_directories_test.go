 package config_test

 import (
 	"os"
 	"path/filepath"
 	"strings"
 	"testing"

 	"github.com/bridge-allah/a-r-s-y/arsy/config"
 )

 func TestEnsureBaaSDirectories_Defaults(t *testing.T) {
 	root := t.TempDir()
 	d, err := config.EnsureBaaSDirectories(root, "")
 	if err != nil {
 		t.Fatalf("unexpected error: %v", err)
 	}
 	if d.TenantsRoot != filepath.Join(root, "tenants") {
 		t.Errorf("expected default tenants root, got %q", d.TenantsRoot)
 	}
 	info, err := os.Stat(d.TenantsRoot)
 	if err != nil {
 		t.Fatalf("tenants root not created: %v", err)
 	}
	_ = info // permission check is OS-specific on Windows; existence is sufficient
 }

 func TestEnsureBaaSDirectories_CustomRoot(t *testing.T) {
 	root := t.TempDir()
 	custom := filepath.Join(t.TempDir(), "custom-tenants")
 	d, err := config.EnsureBaaSDirectories(root, custom)
 	if err != nil {
 		t.Fatalf("unexpected error: %v", err)
 	}
 	if d.TenantsRoot != custom {
 		t.Errorf("expected custom tenants root %q, got %q", custom, d.TenantsRoot)
 	}
 }

func TestEnsureBaaSDirectories_RelativeTenantRootResolved(t *testing.T) {
	root := t.TempDir()
	d, err := config.EnsureBaaSDirectories(root, "relative-tenants")
	if err != nil {
		t.Fatalf("unexpected error for relative tenant root: %v", err)
	}
	if !filepath.IsAbs(d.TenantsRoot) {
		t.Fatalf("expected resolved absolute tenants root, got %q", d.TenantsRoot)
	}
	if filepath.Base(d.TenantsRoot) != "relative-tenants" {
		t.Fatalf("unexpected resolved tenants root: %q", d.TenantsRoot)
	}
}

 func TestTenantDBPath_Valid(t *testing.T) {
 	root := t.TempDir()
 	d, _ := config.EnsureBaaSDirectories(root, "")
 	p, err := d.TenantDBPath("abc123")
 	if err != nil {
 		t.Fatalf("unexpected error: %v", err)
 	}
 	if !strings.HasSuffix(p, "abc123.db") {
 		t.Errorf("expected path ending in abc123.db, got %q", p)
 	}
 }

 func TestTenantDBPath_EmptyIDRejected(t *testing.T) {
 	root := t.TempDir()
 	d, _ := config.EnsureBaaSDirectories(root, "")
 	if _, err := d.TenantDBPath(""); err == nil {
 		t.Fatal("expected error for empty tenant ID")
 	}
 }

 func TestTenantDBPath_PathComponentRejected(t *testing.T) {
 	root := t.TempDir()
 	d, _ := config.EnsureBaaSDirectories(root, "")
 	for _, bad := range []string{"../escape", "sub/dir", "a/b"} {
 		if _, err := d.TenantDBPath(bad); err == nil {
 			t.Errorf("expected error for tenant ID %q, got nil", bad)
 		}
 	}
 }
