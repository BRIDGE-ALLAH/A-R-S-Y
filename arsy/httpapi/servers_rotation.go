package httpapi

import (
	"net/http"
	"strings"

	"github.com/bridge-allah/a-r-s-y/arsy/config"
	"github.com/bridge-allah/a-r-s-y/arsy/servers"
	"github.com/bridge-allah/a-r-s-y/core"
)

type rotateCredentialRequest struct {
	PrivateKey string `json:"private_key"`
}

// rotateServerCredential handles POST /servers/{id}/rotate-credential.
// Encrypts and atomically replaces the server SSH private key.
func rotateServerCredential(cfg config.Config) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		serverID := e.Request.PathValue("id")
		if serverID == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "server ID required"})
		}
		var req rotateCredentialRequest
		if err := decodeServerRequest(e, &req); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": err.Error()})
		}
		req.PrivateKey = strings.TrimSpace(req.PrivateKey)
		if req.PrivateKey == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "private_key is required"})
		}
		if err := servers.RotateServerCredential(e.App, serverID, req.PrivateKey, cfg.SecretKeyEnv); err != nil {
			if strings.Contains(err.Error(), "server not found") {
				return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "server not found"})
			}
			return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": "credential rotation failed"})
		}
		_ = servers.LogAuditEvent(e.App, "credential_rotated", serverID, "", map[string]any{
			"method": "ssh_private_key",
		})
		return e.JSON(http.StatusOK, map[string]any{
			"status":    "rotated",
			"server_id": serverID,
			"message":   "Credential rotated. Run SSH test to verify the new key before the next deployment.",
		})
	}
}
