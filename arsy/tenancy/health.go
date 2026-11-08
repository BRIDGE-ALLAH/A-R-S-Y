package tenancy

import (
	"context"
	"database/sql"
	"os"
	"time"
)

// TenantHealthReport is the safe, redacted health summary for a single tenant database.
// It never contains filesystem paths, DSNs, credentials or secret references.
type TenantHealthReport struct {
	TenantID      string `json:"tenant_id"`
	Healthy       bool   `json:"healthy"`
	Error         string `json:"error,omitempty"`
	SchemaVersion string `json:"schema_version,omitempty"`
	FileSizeBytes int64  `json:"file_size_bytes,omitempty"`
	ConnectionOK  bool   `json:"connection_ok"`
	CheckedAt     string `json:"checked_at"`
}

// CheckTenant probes a single tenant database and returns a redacted report.
func CheckTenant(ctx context.Context, provider Provider, dirs configDirs, tenantID string) TenantHealthReport {
	report := TenantHealthReport{TenantID: tenantID, CheckedAt: time.Now().UTC().Format(time.RFC3339)}
	status := HealthStatus{}
	if hc, ok := provider.(HealthChecker); ok {
		status = hc.Health(ctx, tenantID)
	} else {
		ds, err := provider.Open(ctx, tenantID)
		if err != nil {
			status = HealthStatus{TenantID: tenantID, Healthy: false, Err: err}
		} else {
			if err := ds.Ping(ctx); err != nil {
				status = HealthStatus{TenantID: tenantID, Healthy: false, Err: err}
			} else {
				status = HealthStatus{TenantID: tenantID, Healthy: true}
			}
			_ = ds.Close()
		}
	}
	report.Healthy = status.Healthy
	report.ConnectionOK = status.Healthy
	if status.Err != nil {
		report.Error = status.Err.Error()
		return report
	}
	ds, err := provider.Open(ctx, tenantID)
	if err != nil {
		report.Healthy = false
		report.Error = err.Error()
		report.ConnectionOK = false
		return report
	}
	defer ds.Close()
	var version string
	if err := ds.DB().QueryRowContext(ctx, `SELECT version FROM arsy_schema_versions ORDER BY version DESC LIMIT 1`).Scan(&version); err == nil {
		report.SchemaVersion = version
	} else if err != sql.ErrNoRows {
		// table missing or other error: keep schema_version empty
	}
	if p, err := dirs.TenantDBPath(tenantID); err == nil {
		if info, err := os.Stat(p); err == nil {
			report.FileSizeBytes = info.Size()
		}
	}
	return report
}

type configDirs interface {
	TenantDBPath(tenantID string) (string, error)
}
