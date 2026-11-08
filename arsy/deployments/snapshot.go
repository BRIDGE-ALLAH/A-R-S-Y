package deployments

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/bridge-allah/a-r-s-y/core"
)

// ErrSnapshotUnavailable indicates a deployment has no stored manifest
// snapshot, so it cannot be used as a rollback source.
var ErrSnapshotUnavailable = errors.New("deployment has no stored manifest snapshot")

// SnapshotManifest stores the exact Compose manifest bytes a deployment
// uploaded, encrypted at rest. It is idempotent: an existing snapshot is never
// overwritten, so a re-processed deployment cannot silently swap revisions.
func SnapshotManifest(app core.App, deploymentID, manifest, secretKeyEnv string) error {
	if app == nil {
		return errors.New("PocketBase app is required")
	}
	if strings.TrimSpace(deploymentID) == "" {
		return errors.New("deployment ID is required")
	}
	if strings.TrimSpace(manifest) == "" {
		return errors.New("manifest must not be empty")
	}
	record, err := app.FindFirstRecordByData("arsy_deployments", "deployment_id", deploymentID)
	if err != nil {
		return fmt.Errorf("find deployment for snapshot: %w", err)
	}
	if record.GetString("manifest_ciphertext") != "" {
		return nil
	}
	ciphertext, err := EncryptManifest(manifest, secretKeyEnv)
	if err != nil {
		return fmt.Errorf("encrypt deployment manifest snapshot: %w", err)
	}
	record.Set("manifest_ciphertext", ciphertext)
	checksum := sha256.Sum256([]byte(manifest))
	record.Set("manifest_checksum", hex.EncodeToString(checksum[:]))
	if err := app.Save(record); err != nil {
		return fmt.Errorf("save deployment manifest snapshot: %w", err)
	}
	return nil
}

// LoadDeploymentManifest returns the decrypted manifest snapshot for a
// deployment. It returns ErrSnapshotUnavailable when no snapshot exists.
func LoadDeploymentManifest(app core.App, deploymentID, secretKeyEnv string) (string, error) {
	if app == nil {
		return "", errors.New("PocketBase app is required")
	}
	if strings.TrimSpace(deploymentID) == "" {
		return "", errors.New("deployment ID is required")
	}
	record, err := app.FindFirstRecordByData("arsy_deployments", "deployment_id", deploymentID)
	if err != nil {
		return "", fmt.Errorf("find deployment manifest snapshot: %w", err)
	}
	ciphertext := record.GetString("manifest_ciphertext")
	if ciphertext == "" {
		return "", ErrSnapshotUnavailable
	}
	plaintext, err := DecryptManifest(ciphertext, secretKeyEnv)
	if err != nil {
		return "", fmt.Errorf("decrypt deployment manifest snapshot: %w", err)
	}
	return plaintext, nil
}

// ApplyManifestSnapshot writes a previously snapshotted manifest back onto the
// application record, re-encrypting it and clearing any plaintext copy so the
// application reflects the rolled-back revision.
func ApplyManifestSnapshot(app core.App, applicationID, manifest string, secretKeyEnv string) error {
	if app == nil {
		return errors.New("PocketBase app is required")
	}
	if strings.TrimSpace(applicationID) == "" {
		return errors.New("application ID is required")
	}
	if strings.TrimSpace(manifest) == "" {
		return errors.New("manifest must not be empty")
	}
	if _, err := ValidateCompose(manifest); err != nil {
		return fmt.Errorf("snapshot manifest is not valid Compose: %w", err)
	}
	application, err := app.FindRecordById("arsy_applications", applicationID)
	if err != nil {
		return fmt.Errorf("find application for snapshot restore: %w", err)
	}
	ciphertext, err := EncryptManifest(manifest, secretKeyEnv)
	if err != nil {
		return fmt.Errorf("encrypt restored application manifest: %w", err)
	}
	application.Set("manifest_ciphertext", ciphertext)
	application.Set("compose_manifest", "")
	if err := app.Save(application); err != nil {
		return fmt.Errorf("save restored application manifest: %w", err)
	}
	return nil
}
