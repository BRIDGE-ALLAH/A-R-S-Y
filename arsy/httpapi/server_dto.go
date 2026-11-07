package httpapi

import (
	"encoding/json"

	"github.com/bridge-allah/a-r-s-y/core"
)

// ServerResponse is the only supported JSON representation for server records.
// Keep credential-bearing fields out of this explicit allowlist.
type ServerResponse struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Host               string `json:"host"`
	SSHPort            int    `json:"ssh_port"`
	SSHUsername        string `json:"ssh_username"`
	OSName             string `json:"os_name"`
	OSVersion          string `json:"os_version"`
	HostKeyFingerprint string `json:"host_key_fingerprint"`
	ReadinessState     string `json:"readiness_state"`
	Created            string `json:"created"`
	Updated            string `json:"updated"`
}

// ServerResponseFromRecord copies only fields approved for an administrator API response.
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
		OSVersion:          record.GetString("os_version"),
		HostKeyFingerprint: record.GetString("host_key_fingerprint"),
		ReadinessState:     record.GetString("readiness_state"),
		Created:            record.GetString("created"),
		Updated:            record.GetString("updated"),
	}
}

// MarshalServerResponse serializes the allowlisted server fields.
func MarshalServerResponse(record *core.Record) ([]byte, error) {
	return json.Marshal(ServerResponseFromRecord(record))
}
