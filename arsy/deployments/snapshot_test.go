package deployments

import (
	"errors"
	"testing"

	"github.com/bridge-allah/a-r-s-y/tests"
)

const snapshotTestSecretEnv = "ARSY_SNAPSHOT_TEST_KEY"

func setSnapshotTestKey(t *testing.T) {
	t.Helper()
	t.Setenv(snapshotTestSecretEnv, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")
}

func TestSnapshotManifestRoundTrip(t *testing.T) {
	setSnapshotTestKey(t)
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	deployment := createDeploymentFixture(t, app, "snapshot-roundtrip", StateQueued)
	deploymentID := deployment.GetString("deployment_id")
	manifest := "services:\n  web:\n    image: nginx:1.27\n"

	if err := SnapshotManifest(app, deploymentID, manifest, snapshotTestSecretEnv); err != nil {
		t.Fatalf("SnapshotManifest: %v", err)
	}

	stored, err := app.FindFirstRecordByData("arsy_deployments", "deployment_id", deploymentID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.GetString("manifest_ciphertext") == "" {
		t.Fatal("expected an encrypted snapshot to be stored")
	}
	if stored.GetString("manifest_checksum") == "" {
		t.Fatal("expected the snapshot checksum to be stored")
	}

	loaded, err := LoadDeploymentManifest(app, deploymentID, snapshotTestSecretEnv)
	if err != nil {
		t.Fatalf("LoadDeploymentManifest: %v", err)
	}
	if loaded != manifest {
		t.Fatalf("snapshot round-trip mismatch:\nwant %q\ngot  %q", manifest, loaded)
	}
}

func TestSnapshotManifestIsIdempotent(t *testing.T) {
	setSnapshotTestKey(t)
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	deployment := createDeploymentFixture(t, app, "snapshot-idempotent", StateQueued)
	deploymentID := deployment.GetString("deployment_id")

	first := "services:\n  web:\n    image: nginx:1.27\n"
	second := "services:\n  web:\n    image: nginx:1.29\n"
	if err := SnapshotManifest(app, deploymentID, first, snapshotTestSecretEnv); err != nil {
		t.Fatalf("first snapshot: %v", err)
	}
	if err := SnapshotManifest(app, deploymentID, second, snapshotTestSecretEnv); err != nil {
		t.Fatalf("second snapshot: %v", err)
	}

	loaded, err := LoadDeploymentManifest(app, deploymentID, snapshotTestSecretEnv)
	if err != nil {
		t.Fatal(err)
	}
	if loaded != first {
		t.Fatalf("snapshot was overwritten; want original %q, got %q", first, loaded)
	}
}

func TestLoadDeploymentManifestWithoutSnapshot(t *testing.T) {
	setSnapshotTestKey(t)
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	deployment := createDeploymentFixture(t, app, "snapshot-missing", StateQueued)
	if _, err := LoadDeploymentManifest(app, deployment.GetString("deployment_id"), snapshotTestSecretEnv); !errors.Is(err, ErrSnapshotUnavailable) {
		t.Fatalf("expected ErrSnapshotUnavailable, got %v", err)
	}
}

func TestApplyManifestSnapshotRestoresApplication(t *testing.T) {
	setSnapshotTestKey(t)
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	deployment := createDeploymentFixture(t, app, "snapshot-apply", StateQueued)
	applicationID := deployment.GetString("application")
	manifest := "services:\n  web:\n    image: nginx:1.27\n"

	if err := ApplyManifestSnapshot(app, applicationID, manifest, snapshotTestSecretEnv); err != nil {
		t.Fatalf("ApplyManifestSnapshot: %v", err)
	}

	application, err := app.FindRecordById("arsy_applications", applicationID)
	if err != nil {
		t.Fatal(err)
	}
	if application.GetString("compose_manifest") != "" {
		t.Fatal("restored application must not keep a plaintext manifest")
	}
	restored, err := DecryptManifest(application.GetString("manifest_ciphertext"), snapshotTestSecretEnv)
	if err != nil {
		t.Fatalf("decrypt restored manifest: %v", err)
	}
	if restored != manifest {
		t.Fatalf("restored manifest mismatch:\nwant %q\ngot  %q", manifest, restored)
	}
}

func TestApplyManifestSnapshotRejectsInvalidCompose(t *testing.T) {
	setSnapshotTestKey(t)
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	deployment := createDeploymentFixture(t, app, "snapshot-invalid", StateQueued)
	if err := ApplyManifestSnapshot(app, deployment.GetString("application"), "not-a-compose-file", snapshotTestSecretEnv); err == nil {
		t.Fatal("expected invalid Compose manifest to be rejected")
	}
}
