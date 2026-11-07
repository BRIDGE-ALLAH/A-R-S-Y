 package migrations

 import (
 	"database/sql"
 	"errors"
 	"fmt"

 	"github.com/bridge-allah/a-r-s-y/core"
 	"github.com/bridge-allah/a-r-s-y/tools/types"
 )

 const (
 	tenantsCollectionName = "baas_tenants"
 )

 // Tenant lifecycle states.
 const (
 	TenantStatusProvisioning = "provisioning"
 	TenantStatusActive       = "active"
 	TenantStatusSuspended    = "suspended"
 	TenantStatusDeleting     = "deleting"
 	TenantStatusFailed       = "failed"
 )

 func init() {
 	core.SystemMigrations.Register(upBaasTenants, downBaasTenants, "20261107010000_baas_tenants.go")
 }

 func upBaasTenants(app core.App) error {
 	_, err := ensureCollection(app, tenantsCollectionName, func(c *core.Collection) {
 		c.Fields.Add(
 			// slug: immutable, server-assigned, URL-safe identifier
 			&core.TextField{
 				Name:     "slug",
 				Required: true,
 				Max:      63,
 				Pattern:  `^[a-z0-9]+(?:-[a-z0-9]+)*$`,
 			},
 			// display_name: human-readable label
 			&core.TextField{
 				Name:     "display_name",
 				Required: true,
 				Max:      120,
 			},
 			// status: one of the defined lifecycle states
 			&core.TextField{
 				Name:     "status",
 				Required: true,
 				Max:      32,
 				Pattern:  `^(provisioning|active|suspended|deleting|failed)$`,
 			},
 			// plan: subscription/plan identifier
 			&core.TextField{
 				Name: "plan",
 				Max:  64,
 			},
 		)
 		c.Indexes = types.JSONArray[string]{
 			"CREATE UNIQUE INDEX idx_baas_tenants_slug ON baas_tenants (slug)",
 		}
 	})
 	return err
 }

 func downBaasTenants(app core.App) error {
 	existing, err := app.FindCollectionByNameOrId(tenantsCollectionName)
 	if err != nil {
 		if errors.Is(err, sql.ErrNoRows) {
 			return nil
 		}
 		return fmt.Errorf("find %s: %w", tenantsCollectionName, err)
 	}
 	return app.Delete(existing)
 }
