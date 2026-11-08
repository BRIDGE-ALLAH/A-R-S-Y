package migrations

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tools/types"
)

const tenantDatabasesCollectionName = "baas_tenant_databases"

func init() {
	core.SystemMigrations.Register(upBaasTenantDatabases, downBaasTenantDatabases, "20261108160000_baas_tenant_databases.go")
}

func upBaasTenantDatabases(app core.App) error {
	tenants, err := app.FindCollectionByNameOrId(tenantsCollectionName)
	if err != nil {
		return fmt.Errorf("baas_tenant_databases migration: tenants collection required: %w", err)
	}
	_, err = ensureCollection(app, tenantDatabasesCollectionName, func(c *core.Collection) {
		c.Fields.Add(
			&core.RelationField{
				Name:          "tenant",
				Required:      true,
				MaxSelect:     1,
				CollectionId:  tenants.Id,
				CascadeDelete: false,
			},
			&core.TextField{Name: "engine", Required: true, Max: 32, Pattern: `^(sqlite|postgres)$`},
			&core.TextField{Name: "placement_key", Max: 120},
			&core.TextField{Name: "database_id", Required: true, Max: 120},
			&core.TextField{Name: "schema_version", Max: 64},
			&core.TextField{Name: "state", Required: true, Max: 32, Pattern: `^(provisioning|active|suspended|deleting|failed|retired)$`},
			&core.TextField{Name: "secret_ref", Max: 255},
			&core.BoolField{Name: "is_primary"},
		)
		c.Indexes = types.JSONArray[string]{
			"CREATE UNIQUE INDEX idx_baas_tenant_databases_tenant_primary ON baas_tenant_databases (tenant) WHERE is_primary = TRUE",
			"CREATE UNIQUE INDEX idx_baas_tenant_databases_database_id ON baas_tenant_databases (database_id)",
			"CREATE INDEX idx_baas_tenant_databases_tenant ON baas_tenant_databases (tenant)",
		}
	})
	return err
}

func downBaasTenantDatabases(app core.App) error {
	existing, err := app.FindCollectionByNameOrId(tenantDatabasesCollectionName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("find %s: %w", tenantDatabasesCollectionName, err)
	}
	return app.Delete(existing)
}
