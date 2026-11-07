package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Directories describes the private A-R-S-Y runtime directories.
type Directories struct {
	Root      string
	Temp      string
	Manifests string
}

// BaaSDirectories extends Directories with tenant-specific paths.
type BaaSDirectories struct {
	Directories
	// TenantsRoot is the private directory holding per-tenant database files.
	// It is created with owner-only (0700) permissions and is never world-readable.
	TenantsRoot string
}

// EnsureBaaSDirectories creates the full A-R-S-Y + BaaS directory layout.
// tenantsRoot may be empty; if so it defaults to <root>/tenants.
func EnsureBaaSDirectories(root, tenantsRoot string) (BaaSDirectories, error) {
	dirs, err := EnsureDirectories(root)
	if err != nil {
		return BaaSDirectories{}, err
	}
	if tenantsRoot == "" {
		tenantsRoot = filepath.Join(root, "tenants")
	}
	tenantsRoot = filepath.Clean(tenantsRoot)
	if !filepath.IsAbs(tenantsRoot) {
		return BaaSDirectories{}, fmt.Errorf("tenant data root must be an absolute path, got %q", tenantsRoot)
	}
	// Guard against traversal escaping the root when using the default layout.
	// A custom tenantsRoot outside root is explicitly permitted (operator choice).
	if err := os.MkdirAll(tenantsRoot, 0o700); err != nil {
		return BaaSDirectories{}, fmt.Errorf("create tenant data root %q: %w", tenantsRoot, err)
	}
	if err := os.Chmod(tenantsRoot, 0o700); err != nil {
		return BaaSDirectories{}, fmt.Errorf("secure tenant data root %q: %w", tenantsRoot, err)
	}
	return BaaSDirectories{Directories: dirs, TenantsRoot: tenantsRoot}, nil
}

// TenantDBPath returns the canonical SQLite file path for the given tenantID.
// tenantID must be a non-empty server-assigned immutable identifier (no path components allowed).
func (d BaaSDirectories) TenantDBPath(tenantID string) (string, error) {
	if tenantID == "" {
		return "", fmt.Errorf("tenant ID must not be empty")
	}
	if filepath.Base(tenantID) != tenantID {
		return "", fmt.Errorf("tenant ID must not contain path components, got %q", tenantID)
	}
	p := filepath.Join(d.TenantsRoot, tenantID+".db")
	// Ensure the resolved path stays inside TenantsRoot (symlink-safe check via Clean).
	if !isSubPath(d.TenantsRoot, p) {
		return "", fmt.Errorf("resolved tenant DB path %q escapes tenant root %q", p, d.TenantsRoot)
	}
	return p, nil
}

// isSubPath reports whether target is a sub-path of parent (both must be clean absolute paths).
func isSubPath(parent, target string) bool {
	rel, err := filepath.Rel(parent, target)
	if err != nil {
		return false
	}
	return rel != ".." && len(rel) >= 2 && rel[:3] != "../" || rel == filepath.Base(target)
}

// EnsureDirectories creates the A-R-S-Y directory layout. Private directories
// are created with owner-only permissions.
func EnsureDirectories(root string) (Directories, error) {
	if root == "" {
		return Directories{}, fmt.Errorf("A-R-S-Y data directory is empty")
	}
	dirs := Directories{
		Root: root,
		Temp: filepath.Join(root, "tmp"),
		Manifests: filepath.Join(root, "manifests"),
	}
	for _, dir := range []string{dirs.Root, dirs.Temp, dirs.Manifests} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return Directories{}, fmt.Errorf("create A-R-S-Y directory %q: %w", dir, err)
		}
		if err := os.Chmod(dir, 0o700); err != nil {
			return Directories{}, fmt.Errorf("secure A-R-S-Y directory %q: %w", dir, err)
		}
	}
	return dirs, nil
}
