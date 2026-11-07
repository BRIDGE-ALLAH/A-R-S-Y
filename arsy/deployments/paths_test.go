package deployments

import (
	"strings"
	"testing"
)

func TestNewDeploymentIDIsRandomAndPathSafe(t *testing.T) {
	first, err := NewDeploymentID()
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewDeploymentID()
	if err != nil {
		t.Fatal(err)
	}
	if first == second || len(first) != 32 {
		t.Fatalf("deployment IDs must be unique 32-character hex values: %q, %q", first, second)
	}
	if _, err := ReleasePath("my-app", first); err != nil {
		t.Fatal(err)
	}
}

func TestReleasePathRejectsTraversalAndInvalidIdentifiers(t *testing.T) {
	tests := []struct {
		slug string
		id   string
	}{
		{"../etc", "0123456789abcdef0123456789abcdef"},
		{"app/../../tmp", "0123456789abcdef0123456789abcdef"},
		{"My-App", "0123456789abcdef0123456789abcdef"},
		{"valid-app", "../../etc/passwd"},
	}
	for _, tt := range tests {
		if path, err := ReleasePath(tt.slug, tt.id); err == nil || path != "" {
			t.Fatalf("ReleasePath(%q, %q) = %q, %v; want rejection", tt.slug, tt.id, path, err)
		}
	}
}

func TestManifestPathUsesFixedGeneratedLocation(t *testing.T) {
	path, err := ManifestPath("demo-app", "0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(path, "/opt/arsy/apps/demo-app/releases/") || !strings.HasSuffix(path, "/compose.yaml") {
		t.Fatalf("unexpected manifest path: %q", path)
	}
}
