package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"ganigeorgiev.com/pocketbase/dbx"
	"github.com/bridge-allah/a-r-s-y/arsy/config"
	"github.com/bridge-allah/a-r-s-y/arsy/engine/persist"
	"github.com/bridge-allah/a-r-s-y/arsy/engine/query"
	"github.com/bridge-allah/a-r-s-y/arsy/engine/registry"
	"github.com/bridge-allah/a-r-s-y/arsy/tenancy"
	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tools/router"
	_ "modernc.org/sqlite"
)

type EngineConfig struct {
	Registry    registry.ModelRegistry
	TenantsRoot string
}

func registerEngineRoutes(api *router.RouterGroup[*core.RequestEvent], _ config.Config, engine EngineConfig) {
	if engine.Registry == nil {
		return
	}
	api.GET("/engine/models", requireSuperuser(func(e *core.RequestEvent) error {
		return e.JSON(http.StatusOK, map[string]any{"items": engine.Registry.Snapshot()})
	}))
	if strings.TrimSpace(engine.TenantsRoot) == "" {
		return
	}
	tenant := func(next func(*core.RequestEvent, tenancy.Datastore) error) func(*core.RequestEvent) error {
		return requireAuthenticated(func(e *core.RequestEvent) error {
			ds, err := resolveAuthorizedTenantDatastore(e, engine)
			if err != nil {
				return writeTenantError(e, err)
			}
			defer ds.Close()
			return next(e, ds)
		})
	}
	api.POST("/engine/{model}/records", tenant(func(e *core.RequestEvent, ds tenancy.Datastore) error { return engineCreate(e, ds, engine.Registry) }))
	api.GET("/engine/{model}/records/{id}", tenant(func(e *core.RequestEvent, ds tenancy.Datastore) error { return engineGet(e, ds, engine.Registry) }))
	api.PATCH("/engine/{model}/records/{id}", tenant(func(e *core.RequestEvent, ds tenancy.Datastore) error { return engineUpdate(e, ds, engine.Registry) }))
	api.DELETE("/engine/{model}/records/{id}", tenant(func(e *core.RequestEvent, ds tenancy.Datastore) error { return engineDelete(e, ds, engine.Registry) }))
	api.POST("/engine/{model}/query", tenant(func(e *core.RequestEvent, ds tenancy.Datastore) error { return engineQuery(e, ds, engine.Registry) }))
}

func resolveAuthorizedTenantDatastore(e *core.RequestEvent, engine EngineConfig) (tenancy.Datastore, error) {
	slug := strings.TrimSpace(e.Request.Header.Get("X-Arsy-Tenant"))
	if slug == "" {
		slug = strings.TrimSpace(e.Request.URL.Query().Get("tenant"))
	}
	if slug == "" {
		return nil, errors.New("tenant is required (X-Arsy-Tenant header or tenant query)")
	}
	record, err := e.App.FindFirstRecordByData("baas_tenants", "slug", slug)
	if err != nil {
		return nil, tenancy.ErrTenantNotFound
	}
	tenant := tenancy.Tenant{ID: record.Id, Slug: record.GetString("slug"), Status: tenancy.Status(record.GetString("status"))}
	if err := authorizeEngineTenant(e, record); err != nil {
		return nil, err
	}
	switch tenant.Status {
	case tenancy.StatusSuspended:
		return nil, tenancy.ErrTenantSuspended
	case tenancy.StatusDeleting:
		return nil, tenancy.ErrTenantDeleting
	case tenancy.StatusProvisioning, tenancy.StatusFailed:
		return nil, errors.New("tenant is not active")
	}
	dirs := config.BaaSDirectories{TenantsRoot: engine.TenantsRoot}
	dbPath, err := dirs.TenantDBPath(tenant.ID)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := db.PingContext(e.Request.Context()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &sqlTenantDatastore{id: tenant.ID, db: db}, nil
}

var errEngineTenantForbidden = errors.New("engine tenant access denied")

func authorizeEngineTenant(e *core.RequestEvent, tenant *core.Record) error {
	if e == nil || e.Auth == nil {
		return errors.New("unauthorized")
	}
	if e.HasSuperuserAuth() {
		return nil
	}
	owner := strings.TrimSpace(tenant.GetString("owner"))
	if owner != "" && owner == e.Auth.Id {
		return nil
	}
	// Membership table is optional/future; owner is the minimal implemented grant.
	return errEngineTenantForbidden
}

type sqlTenantDatastore struct {
	id string
	db *sql.DB
}

func (s *sqlTenantDatastore) DB() *sql.DB                    { return s.db }
func (s *sqlTenantDatastore) TenantID() string               { return s.id }
func (s *sqlTenantDatastore) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }
func (s *sqlTenantDatastore) Close() error                   { return s.db.Close() }

var _ tenancy.Datastore = (*sqlTenantDatastore)(nil)

func writeTenantError(e *core.RequestEvent, err error) error {
	switch {
	case err != nil && err.Error() == "unauthorized":
		return e.JSON(http.StatusUnauthorized, map[string]any{"status": "unauthorized"})
	case errors.Is(err, errEngineTenantForbidden):
		return e.JSON(http.StatusForbidden, map[string]any{"status": "forbidden", "message": "tenant access denied"})
	case errors.Is(err, tenancy.ErrTenantNotFound):
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "tenant not found"})
	case errors.Is(err, tenancy.ErrTenantSuspended), errors.Is(err, tenancy.ErrTenantDeleting):
		return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "tenant unavailable"})
	case err != nil && err.Error() == "tenant is required (X-Arsy-Tenant header or tenant query)":
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": err.Error()})
	case err != nil && err.Error() == "tenant is not active":
		return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "tenant unavailable"})
	default:
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "invalid tenant request"})
	}
}

func engineStore(ds tenancy.Datastore, reg registry.ModelRegistry) (*persist.Store, error) {
	raw := ds.DB()
	if raw == nil {
		return nil, errors.New("tenant datastore is unavailable")
	}
	return persist.New(reg, dbx.NewFromDB(raw, "sqlite"))
}

type engineWriteRequest struct {
	Values        map[string]any `json:"values"`
	AllowedFields []string       `json:"allowed_fields"`
}

func requireAllowedWrite(values map[string]any, allowed []string) error {
	if values == nil {
		return errors.New("values is required")
	}
	if len(allowed) == 0 {
		return errors.New("allowed_fields is required")
	}
	set := make(map[string]struct{}, len(allowed))
	for _, f := range allowed {
		set[f] = struct{}{}
	}
	for k := range values {
		if _, ok := set[k]; !ok {
			return errors.New("field is not allowed")
		}
	}
	return nil
}

func engineCreate(e *core.RequestEvent, ds tenancy.Datastore, reg registry.ModelRegistry) error {
	var req engineWriteRequest
	if err := decodeServerRequest(e, &req); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": err.Error()})
	}
	model := e.Request.PathValue("model")
	if err := requireAllowedWrite(req.Values, req.AllowedFields); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": err.Error()})
	}
	store, err := engineStore(ds, reg)
	if err != nil {
		return e.JSON(http.StatusServiceUnavailable, map[string]any{"status": "error", "message": "engine store unavailable"})
	}
	id, err := store.Create(e.Request.Context(), model, req.Values)
	if err != nil {
		return writePersistError(e, err)
	}
	return e.JSON(http.StatusCreated, map[string]any{"status": "created", "id": id})
}

func engineGet(e *core.RequestEvent, ds tenancy.Datastore, reg registry.ModelRegistry) error {
	store, err := engineStore(ds, reg)
	if err != nil {
		return e.JSON(http.StatusServiceUnavailable, map[string]any{"status": "error", "message": "engine store unavailable"})
	}
	row, err := store.Get(e.Request.Context(), e.Request.PathValue("model"), e.Request.PathValue("id"))
	if err != nil {
		return writePersistError(e, err)
	}
	return e.JSON(http.StatusOK, map[string]any{"item": row})
}

func engineUpdate(e *core.RequestEvent, ds tenancy.Datastore, reg registry.ModelRegistry) error {
	var req engineWriteRequest
	if err := decodeServerRequest(e, &req); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": err.Error()})
	}
	if err := requireAllowedWrite(req.Values, req.AllowedFields); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": err.Error()})
	}
	store, err := engineStore(ds, reg)
	if err != nil {
		return e.JSON(http.StatusServiceUnavailable, map[string]any{"status": "error", "message": "engine store unavailable"})
	}
	if err := store.Update(e.Request.Context(), e.Request.PathValue("model"), e.Request.PathValue("id"), req.Values); err != nil {
		return writePersistError(e, err)
	}
	return e.JSON(http.StatusOK, map[string]any{"status": "updated"})
}

func engineDelete(e *core.RequestEvent, ds tenancy.Datastore, reg registry.ModelRegistry) error {
	store, err := engineStore(ds, reg)
	if err != nil {
		return e.JSON(http.StatusServiceUnavailable, map[string]any{"status": "error", "message": "engine store unavailable"})
	}
	if err := store.Delete(e.Request.Context(), e.Request.PathValue("model"), e.Request.PathValue("id")); err != nil {
		return writePersistError(e, err)
	}
	return e.JSON(http.StatusOK, map[string]any{"status": "deleted"})
}

func engineQuery(e *core.RequestEvent, ds tenancy.Datastore, reg registry.ModelRegistry) error {
	var req struct {
		Query         *query.Query `json:"query"`
		AllowedFields []string     `json:"allowed_fields"`
	}
	if err := decodeServerRequest(e, &req); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": err.Error()})
	}
	if len(req.AllowedFields) == 0 {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "allowed_fields is required"})
	}
	store, err := engineStore(ds, reg)
	if err != nil {
		return e.JSON(http.StatusServiceUnavailable, map[string]any{"status": "error", "message": "engine store unavailable"})
	}
	rows, err := store.Query(e.Request.Context(), e.Request.PathValue("model"), req.Query, req.AllowedFields)
	if err != nil {
		return writePersistError(e, err)
	}
	return e.JSON(http.StatusOK, map[string]any{"items": rows})
}

func writePersistError(e *core.RequestEvent, err error) error {
	var nf *persist.ErrNotFound
	var um *persist.ErrUnknownModel
	var uf *persist.ErrUnknownField
	var cv *persist.ErrConstraintViolation
	switch {
	case errors.As(err, &nf), errors.As(err, &um):
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "not found"})
	case errors.As(err, &uf):
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "invalid field or request"})
	case errors.As(err, &cv):
		return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "constraint violation"})
	default:
		return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": "engine operation failed"})
	}
}
