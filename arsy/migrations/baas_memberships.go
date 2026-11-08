package migrations

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tools/types"
)

const tenantMembershipsCollectionName = "baas_tenant_memberships"

func init() {
	core.SystemMigrations.Register(upBaasMemberships, downBaasMemberships, "20261108150000_baas_tenant_memberships.go")
}

func upBaasMemberships(app core.App) error {
	tenants, err := app.FindCollectionByNameOrId(tenantsCollectionName)
	if err != nil {
		return fmt.Errorf("find %s for memberships migration: %w", tenantsCollectionName, err)
	}
	_, err = ensureCollection(app, tenantMembershipsCollectionName, func(c *core.Collection) {
		c.Fields.Add(
			&core.RelationField{Name: "tenant", CollectionId: tenants.Id, Required: true, MaxSelect: 1, CascadeDelete: true},
			&core.TextField{Name: "principal", Required: true, Max: 64, Pattern: `^[a-zA-Z0-9]+$`},
			&core.TextField{Name: "role", Required: true, Max: 16, Pattern: `^(owner|admin|developer|viewer)$`},
			&core.TextField{Name: "status", Required: true, Max: 16, Pattern: `^(active|inactive)$`},
			&core.TextField{Name: "granted_by", Max: 64},
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
		)
		c.Indexes = types.JSONArray[string]{
			"CREATE UNIQUE INDEX idx_baas_tenant_memberships_tenant_principal ON baas_tenant_memberships (tenant, principal)",
			"CREATE INDEX idx_baas_tenant_memberships_principal ON baas_tenant_memberships (principal)",
			"CREATE INDEX idx_baas_tenant_memberships_tenant_status ON baas_tenant_memberships (tenant, status)",
		}
	})
	return err
}

func downBaasMemberships(app core.App) error {
	existing, err := app.FindCollectionByNameOrId(tenantMembershipsCollectionName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("find %s: %w", tenantMembershipsCollectionName, err)
	}
	return app.Delete(existing)
}

