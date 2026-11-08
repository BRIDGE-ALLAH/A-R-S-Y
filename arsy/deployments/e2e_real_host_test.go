package deployments_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/bridge-allah/a-r-s-y/arsy/deployments"
	"github.com/bridge-allah/a-r-s-y/arsy/servers"
)

// TestE2EDeploymentWorkflowAgainstRealHost exercises the deployment workflow
// against a real managed Linux host. It is intentionally skipped unless the
// operator provides explicit infrastructure credentials.
//
// Required env:
//   ARSY_E2E_HOST, ARSY_E2E_SSH_USER, ARSY_E2E_SSH_PRIVATE_KEY_PATH,
//   ARSY_E2E_HOST_KEY_FINGERPRINT, ARSY_E2E_COMPOSE_MANIFEST_PATH
//
// This test never runs in CI by default and performs no fake/simulated deploy.
func TestE2EDeploymentWorkflowAgainstRealHost(t *testing.T) {
	required := []string{
		"ARSY_E2E_HOST",
		"ARSY_E2E_SSH_USER",
		"ARSY_E2E_SSH_PRIVATE_KEY_PATH",
		"ARSY_E2E_HOST_KEY_FINGERPRINT",
		"ARSY_E2E_COMPOSE_MANIFEST_PATH",
		"ARSY_E2E_REMOTE_MANIFEST_PATH",
	}
	for _, k := range required {
		if os.Getenv(k) == "" {
			t.Skipf("real E2E infra not configured (missing %s); set ARSY_E2E_* to run", k)
		}
	}

	key, err := os.ReadFile(os.Getenv("ARSY_E2E_SSH_PRIVATE_KEY_PATH"))
	if err != nil {
		t.Fatalf("read SSH private key: %v", err)
	}
	manifest, err := os.ReadFile(os.Getenv("ARSY_E2E_COMPOSE_MANIFEST_PATH"))
	if err != nil {
		t.Fatalf("read compose manifest: %v", err)
	}
	if _, err := deployments.ValidateCompose(string(manifest)); err != nil {
		t.Fatalf("compose manifest invalid: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	client, err := servers.NewSSHClient(ctx, os.Getenv("ARSY_E2E_HOST"), os.Getenv("ARSY_E2E_SSH_USER"), key, os.Getenv("ARSY_E2E_HOST_KEY_FINGERPRINT"), 15*time.Second)
	if err != nil {
		t.Fatalf("real SSH connection failed (no simulation): %v", err)
	}
	defer client.Close()

	executor := servers.Executor{Timeout: 60 * time.Second, MaxOutputBytes: 64 * 1024}
	remoteManifest := os.Getenv("ARSY_E2E_REMOTE_MANIFEST_PATH")
	res, err := executor.ValidateComposeManifest(ctx, client, remoteManifest)
	if err != nil {
		t.Fatalf("real compose validation exec failed: %v", err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("real compose validation failed: stdout=%q stderr=%q", res.Stdout, res.Stderr)
	}
}
