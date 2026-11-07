 package tenancy

 import (
 	"context"
 	"fmt"
 )

 // Provisioner creates a new tenant database and runs its initial migrations.
 // It is idempotent: if the database already exists it just runs pending migrations.
 type Provisioner struct {
 	provider Provider
 	runner   *MigrationRunner
 }

 // NewProvisioner creates a Provisioner using the given provider and migrations.
 func NewProvisioner(provider Provider, migrations []Migration) *Provisioner {
 	return &Provisioner{
 		provider: provider,
 		runner:   NewMigrationRunner(migrations),
 	}
 }

 // Provision opens (or creates) the tenant database and applies pending migrations.
 // Returns ErrInvalidTenantID for a blank tenantID.
 func (p *Provisioner) Provision(ctx context.Context, tenantID string) error {
 	if tenantID == "" {
 		return ErrInvalidTenantID
 	}
 	ds, err := p.provider.Open(ctx, tenantID)
 	if err != nil {
 		return fmt.Errorf("tenancy provision: open datastore for %q: %w", tenantID, err)
 	}
 	if err := p.runner.Migrate(ctx, ds); err != nil {
 		return fmt.Errorf("tenancy provision: migrate %q: %w", tenantID, err)
 	}
 	return nil
 }
