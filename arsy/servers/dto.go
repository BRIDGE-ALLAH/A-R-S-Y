package servers

import (
	"strings"

	"github.com/bridge-allah/a-r-s-y/core"
)

// ServerResponse is the allow-listed public representation of a server.
// Credential ciphertext and decrypted private-key material are never included.
type ServerResponse struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Host               string `json:"host"`
	SSHPort            int    `json:"ssh_port"`
	SSHUsername        string `json:"ssh_username"`
	OSName             string `json:"os_name"`
	Architecture       string `json:"architecture"`
	HostKeyFingerprint string `json:"host_key_fingerprint"`
	ReadinessState     string `json:"readiness_state"`
}

// ServerResponseFromRecord converts a credential-bearing server record to an
// explicit response DTO containing only allow-listed non-secret fields.
func ServerResponseFromRecord(record *core.Record) ServerResponse {
	if record == nil {
		return ServerResponse{}
	}
	return ServerResponse{
		ID:                 record.Id,
		Name:               record.GetString("name"),
		Host:               record.GetString("host"),
		SSHPort:            record.GetInt("ssh_port"),
		SSHUsername:        record.GetString("ssh_username"),
		OSName:             record.GetString("os_name"),
		Architecture:       record.GetString("architecture"),
		HostKeyFingerprint: record.GetString("host_key_fingerprint"),
		ReadinessState:     record.GetString("readiness_state"),
	}
}

// RedactSecrets removes known secret values from error text before it is
// returned to the console or stored in deployment diagnostics.
func RedactSecrets(message string, secrets ...string) string {
	for _, secret := range secrets {
		if secret != "" {
			message = strings.ReplaceAll(message, secret, "[REDACTED]")
		}
	}
	return message
}
