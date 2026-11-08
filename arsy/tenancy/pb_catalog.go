package tenancy

import (
	"context"
	"errors"

	"github.com/bridge-allah/a-r-s-y/core"
)

// PBCatalog resolves tenants from the control-plane `baas_tenants` collection.
// It is a read-only adapter; lifecycle writes remain in the control plane.
type PBCatalog struct {
	app core.App
}

func NewPBCatalog(app core.App) *PBCatalog { return &PBCatalog{app: app} }

func (c *PBCatalog) Get(_ context.Context, id string) (Tenant, error) {
	if c == nil || c.app == nil {
		return Tenant{}, errors.New("tenancy: catalog app is required")
	}
	record, err := c.app.FindRecordById("baas_tenants", id)
	if err != nil {
		return Tenant{}, ErrTenantNotFound
	}
	return tenantFromRecord(record), nil
}

func (c *PBCatalog) GetBySlug(_ context.Context, slug string) (Tenant, error) {
	if c == nil || c.app == nil {
		return Tenant{}, errors.New("tenancy: catalog app is required")
	}
	record, err := c.app.FindFirstRecordByData("baas_tenants", "slug", slug)
	if err != nil {
		return Tenant{}, ErrTenantNotFound
	}
	return tenantFromRecord(record), nil
}

func tenantFromRecord(r *core.Record) Tenant {
	return Tenant{
		ID:          r.Id,
		Slug:        r.GetString("slug"),
		DisplayName: r.GetString("display_name"),
		Status:      Status(r.GetString("status")),
		Plan:        r.GetString("plan"),
	}
}

var _ Catalog = (*PBCatalog)(nil)
