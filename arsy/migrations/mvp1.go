package migrations

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tools/types"
)

const (
	projectsCollectionName       = "arsy_projects"
	serversCollectionName        = "arsy_servers"
	applicationsCollectionName   = "arsy_applications"
	deploymentsCollectionName    = "arsy_deployments"
	deploymentLogsCollectionName = "arsy_deployment_logs"
	auditLogCollectionName       = "arsy_audit_log"
	sshSessionsCollectionName    = "arsy_ssh_sessions"
)

func init() {
	core.SystemMigrations.Register(up, down, "20261107000000_arsy_init.go")
}

func up(app core.App) error {
	projects, err := ensureCollection(app, projectsCollectionName, func(c *core.Collection) {
		c.Fields.Add(
			&core.TextField{Name: "name", Required: true, Max: 120},
			&core.TextField{Name: "slug", Required: true, Max: 64, Pattern: "^[a-z0-9]+(?:-[a-z0-9]+)*$"},
			&core.TextField{Name: "description", Max: 2000},
		)
		c.Indexes = types.JSONArray[string]{
			"CREATE UNIQUE INDEX idx_arsy_projects_slug ON arsy_projects (slug)",
		}
	})
	if err != nil {
		return err
	}

	servers, err := ensureCollection(app, serversCollectionName, func(c *core.Collection) {
		c.Fields.Add(
			&core.TextField{Name: "name", Required: true, Max: 120},
			&core.TextField{Name: "host", Required: true, Max: 253},
			&core.NumberField{Name: "ssh_port", Required: true, Min: float64Pointer(1), Max: float64Pointer(65535)},
			&core.TextField{Name: "ssh_username", Required: true, Max: 64},
			&core.TextField{Name: "os_name", Max: 80},
			&core.TextField{Name: "architecture", Max: 40},
			&core.TextField{Name: "host_key_fingerprint", Max: 128},
			&core.TextField{Name: "pending_host_key_fingerprint", Hidden: true, Max: 128},
			&core.SelectField{Name: "readiness_state", Values: []string{"unknown", "pending", "ready", "not_ready", "error"}, MaxSelect: 1},
			&core.TextField{Name: "credential_ciphertext", Hidden: true, Max: 10000},
			&core.BoolField{Name: "host_key_revoked"},
		)
		c.Indexes = types.JSONArray[string]{
			"CREATE UNIQUE INDEX idx_arsy_servers_host_port ON arsy_servers (host, ssh_port)",
			"CREATE INDEX idx_arsy_servers_host_key_revoked ON arsy_servers (host_key_revoked)",
		}
	})
	if err != nil {
		return err
	}

	applications, err := ensureCollection(app, applicationsCollectionName, func(c *core.Collection) {
		c.Fields.Add(
			&core.RelationField{Name: "project", CollectionId: projects.Id, Required: true, MaxSelect: 1},
			&core.RelationField{Name: "server", CollectionId: servers.Id, Required: true, MaxSelect: 1},
			&core.TextField{Name: "name", Required: true, Max: 120},
			&core.TextField{Name: "slug", Required: true, Max: 64, Pattern: "^[a-z0-9]+(?:-[a-z0-9]+)*$"},
			&core.TextField{Name: "compose_manifest", Required: true, Max: 50000},
			&core.TextField{Name: "manifest_ciphertext", Hidden: true, Max: 100000},
			&core.SelectField{Name: "lifecycle_state", Values: []string{"draft", "ready", "deploying", "running", "stopped", "failed"}, MaxSelect: 1},
		)
		c.Indexes = types.JSONArray[string]{
			"CREATE UNIQUE INDEX idx_arsy_applications_project_slug ON arsy_applications (project, slug)",
			"CREATE UNIQUE INDEX idx_arsy_applications_server ON arsy_applications (server)",
		}
	})
	if err != nil {
		return err
	}

	deployments, err := ensureCollection(app, deploymentsCollectionName, func(c *core.Collection) {
		c.Fields.Add(
			&core.RelationField{Name: "application", CollectionId: applications.Id, Required: true, MaxSelect: 1},
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
			&core.TextField{Name: "deployment_id", Required: true, Max: 80},
			&core.TextField{Name: "idempotency_key", Max: 128},
			&core.SelectField{Name: "status", Values: []string{"queued", "preparing", "uploading", "validating", "deploying", "verifying", "succeeded", "failed", "cancelled"}, Required: true, MaxSelect: 1},
			&core.DateField{Name: "started_at"},
			&core.DateField{Name: "finished_at"},
			&core.TextField{Name: "manifest_checksum", Max: 128},
			&core.TextField{Name: "summary", Max: 2000},
			&core.TextField{Name: "previous_deployment_id", Max: 80},
		)
		c.Indexes = types.JSONArray[string]{
			"CREATE UNIQUE INDEX idx_arsy_deployments_id ON arsy_deployments (deployment_id)",
			"CREATE UNIQUE INDEX idx_arsy_deployments_idempotency ON arsy_deployments (idempotency_key) WHERE idempotency_key != ''",
		}
	})
	if err != nil {
		return err
	}
	_, err = ensureCollection(app, deploymentLogsCollectionName, func(c *core.Collection) {
		c.Fields.Add(
			&core.RelationField{Name: "deployment", CollectionId: deployments.Id, Required: true, MaxSelect: 1},
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.TextField{Name: "step", Required: true, Max: 80},
			&core.SelectField{Name: "level", Values: []string{"info", "warning", "error"}, Required: true, MaxSelect: 1},
			&core.TextField{Name: "message", Required: true, Max: 5000},
		)
		c.Indexes = types.JSONArray[string]{
			"CREATE INDEX idx_arsy_deployment_logs_deployment_created ON arsy_deployment_logs (deployment, created)",
		}
	})
	if err != nil {
		return err
	}

	_, err = ensureCollection(app, auditLogCollectionName, func(c *core.Collection) {
		c.Fields.Add(
			&core.TextField{Name: "event_type", Required: true, Max: 64},
			&core.RelationField{Name: "server", CollectionId: servers.Id, Required: true, MaxSelect: 1},
			&core.TextField{Name: "changed_by", Max: 80},
			&core.JSONField{Name: "metadata"},
			&core.AutodateField{Name: "created", OnCreate: true},
		)
		c.Indexes = types.JSONArray[string]{
			"CREATE INDEX idx_arsy_audit_log_server_created ON arsy_audit_log (server, created)",
			"CREATE INDEX idx_arsy_audit_log_event_type ON arsy_audit_log (event_type)",
		}
	})
	if err != nil {
		return err
	}

	_, err = ensureCollection(app, sshSessionsCollectionName, func(c *core.Collection) {
		c.Fields.Add(
			&core.TextField{Name: "session_id", Required: true, Max: 128},
			&core.RelationField{Name: "server", CollectionId: servers.Id, Required: true, MaxSelect: 1},
			&core.TextField{Name: "auth_method", Required: true, Max: 32},
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.DateField{Name: "expires_at"},
			&core.DateField{Name: "invalidated_at"},
		)
		c.Indexes = types.JSONArray[string]{
			"CREATE INDEX idx_arsy_ssh_sessions_server_invalidated ON arsy_ssh_sessions (server, invalidated_at)",
			"CREATE INDEX idx_arsy_ssh_sessions_expires_at ON arsy_ssh_sessions (expires_at)",
		}
	})
	return err
}

func down(app core.App) error {
	for _, name := range []string{
		sshSessionsCollectionName,
		auditLogCollectionName,
		deploymentLogsCollectionName,
		deploymentsCollectionName,
		applicationsCollectionName,
		serversCollectionName,
		projectsCollectionName,
	} {
		collection, err := app.FindCollectionByNameOrId(name)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return fmt.Errorf("find A-R-S-Y collection %q: %w", name, err)
		}
		if err := app.Delete(collection); err != nil {
			return fmt.Errorf("delete A-R-S-Y collection %q: %w", name, err)
		}
	}
	return nil
}

func ensureCollection(app core.App, name string, configure func(*core.Collection)) (*core.Collection, error) {
	collection, err := app.FindCollectionByNameOrId(name)
	if err == nil {
		return collection, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("find A-R-S-Y collection %q: %w", name, err)
	}

	collection = core.NewBaseCollection(name)
	collection.ListRule = nil
	collection.ViewRule = nil
	collection.CreateRule = nil
	collection.UpdateRule = nil
	collection.DeleteRule = nil
	configure(collection)

	if err := app.Save(collection); err != nil {
		return nil, fmt.Errorf("create A-R-S-Y collection %q: %w", name, err)
	}
	return collection, nil
}

func float64Pointer(value float64) *float64 {
	return &value
}
