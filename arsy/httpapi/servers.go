package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bridge-allah/a-r-s-y/arsy/config"
	"github.com/bridge-allah/a-r-s-y/arsy/servers"
	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tools/router"
)

const maxServerRequestBytes = 64 * 1024

type createServerRequest struct {
	Name        string `json:"name"`
	Host        string `json:"host"`
	SSHPort     int    `json:"ssh_port"`
	SSHUsername string `json:"ssh_username"`
	PrivateKey  string `json:"private_key"`
}

type updateServerRequest struct {
	Name string `json:"name"`
}

type confirmHostKeyRequest struct {
	Fingerprint string `json:"fingerprint"`
	Confirmed   bool   `json:"confirmed"`
}

func registerServerRoutes(api *router.RouterGroup[*core.RequestEvent], cfg config.Config) {
	api.GET("/servers", requireSuperuser(listServers))
	api.POST("/servers", requireSuperuser(createServer(cfg)))
	api.GET("/servers/{id}", requireSuperuser(getServer))
	api.PATCH("/servers/{id}", requireSuperuser(updateServer))
	api.POST("/servers/{id}/test", requireSuperuser(testServerConnection(cfg)))
	api.POST("/servers/{id}/host-key/confirm", requireSuperuser(confirmServerHostKey))
	api.POST("/servers/{id}/readiness", requireSuperuser(checkServerReadiness(cfg)))
	api.POST("/servers/{id}/rotate-credential", requireSuperuser(rotateServerCredential(cfg)))
	api.GET("/servers/{id}/audit-log", requireSuperuser(getServerAuditLog))
	api.POST("/servers/{id}/host-key/revoke", requireSuperuser(revokeServerHostKey))
}

func listServers(e *core.RequestEvent) error {
	limit := parseBoundedQueryInt(e.Request.URL.Query().Get("limit"), 50, 1, 100)
	offset := parseBoundedQueryInt(e.Request.URL.Query().Get("offset"), 0, 0, 100000)
	records, err := e.App.FindAllRecords("arsy_servers")
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "unable to list servers"})
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].GetString("created") > records[j].GetString("created")
	})
	if offset >= len(records) {
		records = records[:0]
	} else {
		end := offset + limit
		if end > len(records) {
			end = len(records)
		}
		records = records[offset:end]
	}
	items := make([]servers.ServerResponse, 0, len(records))
	for _, record := range records {
		items = append(items, servers.ServerResponseFromRecord(record))
	}
	return e.JSON(http.StatusOK, map[string]any{"items": items, "limit": limit, "offset": offset})
}

func createServer(cfg config.Config) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		var body createServerRequest
		if err := decodeServerRequest(e, &body); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]string{"error": "invalid server request body"})
		}
		name := strings.TrimSpace(body.Name)
		if name == "" || len([]rune(name)) > 120 {
			return e.JSON(http.StatusBadRequest, map[string]string{"error": "server name must be between 1 and 120 characters"})
		}
		input, err := servers.ValidateConnectionInput(servers.ConnectionInput{
			Host:       body.Host,
			Port:       body.SSHPort,
			Username:   body.SSHUsername,
			PrivateKey: []byte(body.PrivateKey),
		})
		if err != nil {
			return e.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		collection, err := e.App.FindCachedCollectionByNameOrId("arsy_servers")
		if err != nil {
			return e.JSON(http.StatusServiceUnavailable, map[string]string{"error": "server storage is unavailable"})
		}
		record := core.NewRecord(collection)
		record.Set("name", name)
		record.Set("host", input.Host)
		record.Set("ssh_port", input.Port)
		record.Set("ssh_username", input.Username)
		record.Set("readiness_state", "unknown")
		privateKey := input.PrivateKey
		body.PrivateKey = ""
		defer clear(privateKey)
		if err := servers.StorePrivateKey(record, privateKey, cfg.SecretKeyEnv); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]string{"error": "unable to encrypt SSH credential; verify the secret key and private key"})
		}
		if err := e.App.Save(record); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]string{"error": "unable to save server; verify that the host and SSH port are unique"})
		}
		return e.JSON(http.StatusCreated, servers.ServerResponseFromRecord(record))
	}
}

func getServer(e *core.RequestEvent) error {
	record, err := e.App.FindRecordById("arsy_servers", e.Request.PathValue("id"))
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]string{"error": "server not found"})
	}
	return e.JSON(http.StatusOK, servers.ServerResponseFromRecord(record))
}

func updateServer(e *core.RequestEvent) error {
	var body updateServerRequest
	if err := decodeServerRequest(e, &body); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "invalid server update body"})
	}
	name := strings.TrimSpace(body.Name)
	if name == "" || len([]rune(name)) > 120 {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "server name must be between 1 and 120 characters"})
	}
	record, err := e.App.FindRecordById("arsy_servers", e.Request.PathValue("id"))
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]string{"error": "server not found"})
	}
	record.Set("name", name)
	if err := e.App.Save(record); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "unable to update server"})
	}
	return e.JSON(http.StatusOK, servers.ServerResponseFromRecord(record))
}

func testServerConnection(cfg config.Config) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		record, err := e.App.FindRecordById("arsy_servers", e.Request.PathValue("id"))
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]string{"error": "server not found"})
		}
		host := record.GetString("host")
		port := record.GetInt("ssh_port")
		address := net.JoinHostPort(host, strconv.Itoa(port))
		ctx, cancel := context.WithTimeout(e.Request.Context(), 15*time.Second)
		defer cancel()

		if record.GetString("host_key_fingerprint") == "" {
			fingerprint, err := servers.ProbeHostKey(ctx, address, 10*time.Second)
			if err != nil {
				return e.JSON(http.StatusBadGateway, map[string]string{"error": "unable to retrieve SSH host key"})
			}
			record.Set("pending_host_key_fingerprint", fingerprint)
			if err := e.App.Save(record); err != nil {
				return e.JSON(http.StatusInternalServerError, map[string]string{"error": "unable to store pending host-key fingerprint"})
			}
			return e.JSON(http.StatusOK, map[string]any{
				"code":        "host_key_confirmation_required",
				"fingerprint": fingerprint,
				"message":     "Verify this fingerprint out of band, then explicitly confirm it before authenticating.",
			})
		}

		privateKey, err := servers.LoadPrivateKey(record, cfg.SecretKeyEnv)
		if err != nil {
			return e.JSON(http.StatusServiceUnavailable, map[string]string{"error": "stored SSH credential is unavailable"})
		}
		defer clear(privateKey)
		result, err := servers.TestSSHConnection(ctx, host, port, record.GetString("ssh_username"), privateKey, record.GetString("host_key_fingerprint"), 10*time.Second)
		if err != nil {
			return e.JSON(http.StatusBadGateway, result)
		}
		return e.JSON(http.StatusOK, result)
	}
}

func confirmServerHostKey(e *core.RequestEvent) error {
	var body confirmHostKeyRequest
	if err := decodeServerRequest(e, &body); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "invalid host-key confirmation body"})
	}
	if !body.Confirmed {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "explicit fingerprint confirmation is required"})
	}
	record, err := e.App.FindRecordById("arsy_servers", e.Request.PathValue("id"))
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]string{"error": "server not found"})
	}
	if err := servers.ConfirmHostKey(record, body.Fingerprint); err != nil {
		return e.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
	}
	if err := e.App.Save(record); err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "unable to save host-key confirmation"})
	}
	return e.JSON(http.StatusOK, servers.ServerResponseFromRecord(record))
}

func checkServerReadiness(cfg config.Config) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		record, err := e.App.FindRecordById("arsy_servers", e.Request.PathValue("id"))
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]string{"error": "server not found"})
		}
		pinned := record.GetString("host_key_fingerprint")
		if pinned == "" {
			return e.JSON(http.StatusConflict, map[string]string{"error": "confirm the SSH host-key fingerprint before checking readiness"})
		}
		privateKey, err := servers.LoadPrivateKey(record, cfg.SecretKeyEnv)
		if err != nil {
			return e.JSON(http.StatusServiceUnavailable, map[string]string{"error": "stored SSH credential is unavailable"})
		}
		defer clear(privateKey)

		address := net.JoinHostPort(record.GetString("host"), strconv.Itoa(record.GetInt("ssh_port")))
		ctx, cancel := context.WithTimeout(e.Request.Context(), 60*time.Second)
		defer cancel()
		client, err := servers.NewSSHClient(ctx, address, record.GetString("ssh_username"), privateKey, pinned, 10*time.Second)
		if err != nil {
			record.Set("readiness_state", "error")
			_ = e.App.Save(record)
			return e.JSON(http.StatusBadGateway, map[string]string{"error": "SSH connection or host-key verification failed"})
		}
		defer client.Close()

		executor := servers.Executor{Timeout: 15 * time.Second, MaxOutputBytes: 16 * 1024}
		osResult, osErr := executor.Execute(ctx, client, servers.RemoteOSFacts)
		sshStatus := servers.ComponentStatus{Ready: true, Code: "ready", Message: "SSH authentication succeeded"}
		hostKeyStatus := servers.ComponentStatus{Ready: true, Code: "ready", Message: "Pinned host key matched"}
		osName, architecture, parseErr := servers.ParseLinuxFacts(osResult.Stdout)
		osStatus := servers.ComponentStatus{Ready: parseErr == nil && osErr == nil && osResult.ExitCode == 0, Code: "unsupported_os", Message: "Managed host must be a supported Linux system"}
		if osStatus.Ready {
			osStatus.Code = "ready"
			osStatus.Message = "Linux host and architecture are supported"
			osStatus.Version = architecture
			record.Set("os_name", osName)
			record.Set("architecture", architecture)
		} else if osErr != nil {
			osStatus.Code = "check_failed"
			osStatus.Message = "Unable to inspect managed host operating system"
		}

		dockerResult, dockerErr := executor.Execute(ctx, client, servers.RemoteDockerVersion)
		dockerStatus := servers.DockerReadiness(dockerResult)
		if dockerErr != nil {
			dockerStatus = servers.ComponentStatus{Code: "check_failed", Message: "Unable to check Docker Engine"}
		}
		composeResult, composeErr := executor.Execute(ctx, client, servers.RemoteComposeVersion)
		composeStatus := servers.ComposeReadiness(composeResult)
		if composeErr != nil {
			composeStatus = servers.ComponentStatus{Code: "check_failed", Message: "Unable to check Docker Compose plugin"}
		}

		summary := servers.BuildReadinessSummary(sshStatus, hostKeyStatus, osStatus, dockerStatus, composeStatus)
		if summary.Ready {
			record.Set("readiness_state", "ready")
		} else {
			record.Set("readiness_state", "not_ready")
		}
		if err := e.App.Save(record); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{"error": "unable to save server readiness"})
		}
		status := http.StatusOK
		if !summary.Ready {
			status = http.StatusConflict
		}
		return e.JSON(status, summary)
	}
}

func decodeServerRequest(e *core.RequestEvent, destination any) error {
	e.Request.Body = http.MaxBytesReader(e.Response, e.Request.Body, maxServerRequestBytes)
	decoder := json.NewDecoder(e.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("request must contain a single JSON object")
	}
	return nil
}

func parseBoundedQueryInt(value string, fallback, minimum, maximum int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	if parsed < minimum {
		return minimum
	}
	if parsed > maximum {
		return maximum
	}
	return parsed
}

func getServerAuditLog(e *core.RequestEvent) error {
	serverID := e.Request.PathValue("id")
	if serverID == "" {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "server ID required"})
	}
	limit := parseBoundedQueryInt(e.Request.URL.Query().Get("limit"), 50, 1, 100)
	offset := parseBoundedQueryInt(e.Request.URL.Query().Get("offset"), 0, 0, 100000)
	eventType := e.Request.URL.Query().Get("event_type")

	records, err := servers.GetServerAuditLog(e.App, serverID, eventType, limit, offset)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
	}

	items := make([]map[string]any, 0, len(records))
	for _, r := range records {
		items = append(items, map[string]any{
			"event_type": r.GetString("event_type"),
			"server_id":  r.GetString("server"),
			"changed_by": r.GetString("changed_by"),
			"created_at": r.GetDateTime("created").String(),
			"metadata":   r.GetString("metadata"),
		})
	}
	return e.JSON(http.StatusOK, map[string]any{"items": items, "limit": limit, "offset": offset})
}

func revokeServerHostKey(e *core.RequestEvent) error {
	serverID := e.Request.PathValue("id")
	if serverID == "" {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "server ID required"})
	}

	if err := servers.RevokeServerHostKey(e.App, serverID, ""); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "server not found"})
		}
		return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": "revocation failed"})
	}

	return e.JSON(http.StatusOK, map[string]any{
		"status":     "host_key_revoked",
		"server_id":  serverID,
		"revoked_at": time.Now().UTC().Format(time.RFC3339),
	})
}
