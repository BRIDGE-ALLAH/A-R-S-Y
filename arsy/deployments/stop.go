package deployments

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/bridge-allah/a-r-s-y/arsy/servers"
	"github.com/bridge-allah/a-r-s-y/core"
)

// StopApplicationError carries the HTTP-friendly reason an application cannot be
// stopped, so the API layer can map it to the correct status without guessing.
type StopApplicationError struct {
	Code    string
	Message string
}

func (e *StopApplicationError) Error() string { return e.Message }

// StopApplication brings a running application's Compose project down on its
// target host and records the lifecycle transition. It never reports success
// unless the remote `docker compose down` actually exits zero.
//
// Guards (each returns a *StopApplicationError):
//   - lifecycle must be "running"
//   - no deployment may be active
//   - the server must have a pinned, non-revoked host key
//   - a succeeded deployment must exist (it defines the release path)
func StopApplication(ctx context.Context, app core.App, applicationID, secretKeyEnv string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if app == nil {
		return "", errors.New("PocketBase app is required")
	}
	if applicationID == "" {
		return "", &StopApplicationError{Code: "invalid_request", Message: "application ID is required"}
	}
	application, err := app.FindRecordById("arsy_applications", applicationID)
	if err != nil {
		return "", &StopApplicationError{Code: "not_found", Message: "application not found"}
	}
	if application.GetString("lifecycle_state") != "running" {
		return "", &StopApplicationError{Code: "not_running", Message: "application is not running"}
	}
	active, err := HasActiveDeployment(app, applicationID)
	if err != nil {
		return "", fmt.Errorf("check active deployments: %w", err)
	}
	if active {
		return "", &StopApplicationError{Code: "conflict", Message: "a deployment is in progress for this application"}
	}
	server, err := app.FindRecordById("arsy_servers", application.GetString("server"))
	if err != nil {
		return "", &StopApplicationError{Code: "not_found", Message: "server not found"}
	}
	if server.GetBool("host_key_revoked") {
		return "", &StopApplicationError{Code: "conflict", Message: "host key revoked; cannot run remote commands"}
	}
	pinnedFingerprint := server.GetString("host_key_fingerprint")
	if pinnedFingerprint == "" {
		return "", &StopApplicationError{Code: "conflict", Message: "target server has no confirmed SSH host-key fingerprint"}
	}

	records, err := app.FindRecordsByFilter(
		"arsy_deployments",
		`application = {:appID} && status = "succeeded"`,
		"-created",
		1, 0,
		map[string]any{"appID": applicationID},
	)
	if err != nil {
		return "", fmt.Errorf("find latest succeeded deployment: %w", err)
	}
	if len(records) == 0 {
		return "", &StopApplicationError{Code: "conflict", Message: "no succeeded deployment to stop"}
	}
	deploymentID := records[0].GetString("deployment_id")
	manifestPath, err := ManifestPath(application.GetString("slug"), deploymentID)
	if err != nil {
		return "", fmt.Errorf("resolve release manifest path: %w", err)
	}

	privateKey, err := servers.LoadPrivateKey(server, secretKeyEnv)
	if err != nil {
		return "", &StopApplicationError{Code: "service_unavailable", Message: "stored SSH credential is unavailable"}
	}
	defer clear(privateKey)

	address := net.JoinHostPort(server.GetString("host"), strconv.Itoa(server.GetInt("ssh_port")))
	sshCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	client, err := servers.NewSSHClient(sshCtx, address, server.GetString("ssh_username"), privateKey, pinnedFingerprint, 10*time.Second)
	if err != nil {
		return "", &StopApplicationError{Code: "bad_gateway", Message: "SSH connection or pinned host-key verification failed"}
	}
	defer client.Close()

	executor := servers.Executor{Timeout: 60 * time.Second, MaxOutputBytes: 32 * 1024}
	result, err := executor.StopComposeManifest(sshCtx, client, manifestPath)
	if err != nil || result.ExitCode != 0 {
		return "", &StopApplicationError{Code: "bad_gateway", Message: "remote docker compose down failed"}
	}

	application.Set("lifecycle_state", "stopped")
	if err := app.Save(application); err != nil {
		return "", fmt.Errorf("save stopped application state: %w", err)
	}
	_ = servers.LogAuditEvent(app, "application_stopped", server.Id, "", map[string]any{
		"application_id":   applicationID,
		"application_slug": application.GetString("slug"),
		"deployment_id":    deploymentID,
	})
	return "stopped", nil
}
