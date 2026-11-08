package deployments

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/bridge-allah/a-r-s-y/arsy/servers"
	"github.com/bridge-allah/a-r-s-y/core"
)

// NewSSHDeploymentProcessor builds the MVP-1 deployment processor. It validates
// the manifest locally, uploads it via SFTP, validates it with Compose on the
// target, deploys, and verifies that every Compose service is running.
func NewSSHDeploymentProcessor(app core.App, secretKeyEnv string) DeploymentProcessor {
	return func(ctx context.Context, deployment *core.Record) (processErr error) {
		if ctx == nil || deployment == nil {
			return errors.New("deployment context and record are required")
		}
		if app == nil {
			return errors.New("PocketBase app is required")
		}
		application, err := app.FindRecordById("arsy_applications", deployment.GetString("application"))
		if err != nil {
			return fmt.Errorf("load deployment application: %w", err)
		}
		application.Set("lifecycle_state", "deploying")
		if err := app.Save(application); err != nil {
			return fmt.Errorf("mark application as deploying: %w", err)
		}
		defer func() {
			if processErr != nil {
				application.Set("lifecycle_state", "failed")
				_ = app.Save(application)
			}
		}()

		server, err := app.FindRecordById("arsy_servers", application.GetString("server"))
		if err != nil {
			return fmt.Errorf("load application server: %w", err)
		}
		if server.GetString("readiness_state") != "ready" {
			return errors.New("target server has not passed readiness checks")
		}
		pinnedFingerprint := server.GetString("host_key_fingerprint")
		if pinnedFingerprint == "" {
			return errors.New("target server has no confirmed SSH host-key fingerprint")
		}
		applicationSlug := application.GetString("slug")
		deploymentID := deployment.GetString("deployment_id")
		if !deploymentIDPattern.MatchString(deploymentID) {
			return errors.New("deployment ID is invalid")
		}
		manifest := []byte(application.GetString("compose_manifest"))
		ciphertext := application.GetString("manifest_ciphertext")
		if ciphertext != "" {
			plaintext, err := DecryptManifest(ciphertext, secretKeyEnv)
			if err != nil {
				return fmt.Errorf("decrypt application manifest: %w", err)
			}
			manifest = []byte(plaintext)
		}
		if _, err := ValidateCompose(string(manifest)); err != nil {
			return fmt.Errorf("application Compose manifest is invalid: %w", err)
		}
		if err := SnapshotManifest(app, deploymentID, string(manifest), secretKeyEnv); err != nil {
			return fmt.Errorf("store deployment manifest snapshot: %w", err)
		}
		manifestPath, err := ManifestPath(applicationSlug, deploymentID)
		if err != nil {
			return err
		}
		checksum := sha256.Sum256(manifest)
		deployment.Set("manifest_checksum", hex.EncodeToString(checksum[:]))
		if err := app.Save(deployment); err != nil {
			return fmt.Errorf("save deployment manifest checksum: %w", err)
		}
		_ = AppendDeploymentLog(app, deployment.Id, "preparing", "info", "Deployment preparation started.")

		privateKey, err := servers.LoadPrivateKey(server, secretKeyEnv)
		if err != nil {
			return errors.New("stored SSH credential could not be decrypted")
		}
		defer clear(privateKey)
		address := net.JoinHostPort(server.GetString("host"), strconv.Itoa(server.GetInt("ssh_port")))
		sshCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		client, err := servers.NewSSHClient(sshCtx, address, server.GetString("ssh_username"), privateKey, pinnedFingerprint, 10*time.Second)
		cancel()
		if err != nil {
			return errors.New("SSH authentication or pinned host-key verification failed")
		}
		defer client.Close()

		executor := servers.Executor{Timeout: 45 * time.Second, MaxOutputBytes: 32 * 1024}
		if _, err := UpdateDeploymentState(app, deployment.Id, StateUploading, "Uploading Compose manifest over SFTP."); err != nil {
			return err
		}
		if err := UploadComposeManifestWithTimeout(ctx, client, applicationSlug, deploymentID, manifest, 2*time.Minute); err != nil {
			_ = AppendDeploymentLog(app, deployment.Id, "upload", "error", "SFTP manifest upload failed.")
			return errors.New("SFTP manifest upload failed")
		}
		_ = AppendDeploymentLog(app, deployment.Id, "upload", "info", "Compose manifest uploaded with owner-only permissions.")

		if _, err := UpdateDeploymentState(app, deployment.Id, StateValidating, "Validating Compose manifest on target host."); err != nil {
			return err
		}
		validation, err := executor.ValidateComposeManifest(ctx, client, manifestPath)
		if err != nil || validation.ExitCode != 0 {
			_ = AppendDeploymentLog(app, deployment.Id, "validate", "error", "Target-side Docker Compose validation failed.")
			return errors.New("target-side Docker Compose validation failed")
		}
		_ = AppendDeploymentLog(app, deployment.Id, "validate", "info", "Target-side Docker Compose validation passed.")

		if _, err := UpdateDeploymentState(app, deployment.Id, StateDeploying, "Starting Compose services."); err != nil {
			return err
		}
		deployResult, err := executor.DeployComposeManifest(ctx, client, manifestPath)
		if err != nil || deployResult.ExitCode != 0 {
			_ = AppendDeploymentLog(app, deployment.Id, "deploy", "error", "Docker Compose up failed.")
			return errors.New("Docker Compose deployment command failed")
		}
		_ = AppendDeploymentLog(app, deployment.Id, "deploy", "info", "Docker Compose up command completed.")

		if _, err := UpdateDeploymentState(app, deployment.Id, StateVerifying, "Checking Compose service state."); err != nil {
			return err
		}
		inspection, err := executor.InspectComposeManifest(ctx, client, manifestPath)
		if err != nil || inspection.ExitCode != 0 {
			_ = AppendDeploymentLog(app, deployment.Id, "verify", "error", "Unable to inspect deployed Compose services.")
			return errors.New("unable to inspect deployed Compose services")
		}
		if err := AllComposeContainersRunning(inspection.Stdout); err != nil {
			_ = AppendDeploymentLog(app, deployment.Id, "verify", "error", "One or more Compose services are not running.")
			return err
		}
		application.Set("lifecycle_state", "running")
		if err := app.Save(application); err != nil {
			return fmt.Errorf("save application running state: %w", err)
		}
		_ = AppendDeploymentLog(app, deployment.Id, "verify", "info", "All Compose services are running.")
		return nil
	}
}

// AllComposeContainersRunning rejects empty or non-running Compose service
// sets, delegating the parsing to the shared inspection parser.
func AllComposeContainersRunning(output string) error {
	containers, err := ParseComposePS(output)
	if err != nil {
		return err
	}
	for _, container := range containers {
		if container.State != "running" {
			return errors.New("one or more Compose services are not running")
		}
	}
	return nil
}
