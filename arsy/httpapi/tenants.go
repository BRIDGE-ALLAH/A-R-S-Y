package httpapi

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/bridge-allah/a-r-s-y/arsy/config"
	"github.com/bridge-allah/a-r-s-y/arsy/engine/bootstrap"
	"github.com/bridge-allah/a-r-s-y/arsy/tenancy"
	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tools/router"
)

var tenantSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type tenantSummary struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	DisplayName string `json:"display_name"`
	Status      string `json:"status"`
	Plan        string `json:"plan,omitempty"`
}

func registerTenantRoutes(api *router.RouterGroup[*core.RequestEvent], _ config.Config, engine EngineConfig) {
	if strings.TrimSpace(engine.TenantsRoot) == "" {
		return
	}
	api.GET("/tenants", requireAuthenticated(func(e *core.RequestEvent) error {
		limit, offset := tenantPagination(e, 50, 200)
		if e.HasSuperuserAuth() {
			records, err := e.App.FindRecordsByFilter("baas_tenants", "1=1", "slug", limit, offset)
			if err != nil {
				return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": "failed to list tenants"})
			}
			return e.JSON(http.StatusOK, map[string]any{"items": tenantSummaries(records), "limit": limit, "offset": offset})
		}
		members, err := e.App.FindRecordsByFilter("baas_tenant_memberships", "principal = {:p} && status = {:s}", "", 500, 0, map[string]any{"p": e.Auth.Id, "s": string(tenancy.MembershipActive)})
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": "failed to list tenants"})
		}
		items := make([]tenantSummary, 0, len(members))
		for _, m := range members {
			r, err := e.App.FindRecordById("baas_tenants", m.GetString("tenant"))
			if err != nil {
				continue
			}
			status := tenancy.Status(r.GetString("status"))
			if status == tenancy.StatusSuspended || status == tenancy.StatusDeleting {
				continue
			}
			items = append(items, tenantSummary{ID: r.Id, Slug: r.GetString("slug"), DisplayName: r.GetString("display_name"), Status: r.GetString("status"), Plan: r.GetString("plan")})
		}
		if offset > len(items) {
			offset = len(items)
		}
		end := offset + limit
		if end > len(items) {
			end = len(items)
		}
		return e.JSON(http.StatusOK, map[string]any{"items": items[offset:end], "limit": limit, "offset": offset})
	}))

	api.POST("/tenants", requireSuperuser(func(e *core.RequestEvent) error {
		var req struct {
			Slug        string `json:"slug"`
			DisplayName string `json:"display_name"`
			Plan        string `json:"plan"`
		}
		if err := decodeServerRequest(e, &req); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "invalid request"})
		}
		req.Slug = strings.ToLower(strings.TrimSpace(req.Slug))
		req.DisplayName = strings.TrimSpace(req.DisplayName)
		if !tenantSlugPattern.MatchString(req.Slug) || req.Slug == "" || len(req.Slug) > 63 {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "invalid tenant slug"})
		}
		if req.DisplayName == "" || len(req.DisplayName) > 120 {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "invalid display name"})
		}
		if _, err := e.App.FindFirstRecordByData("baas_tenants", "slug", req.Slug); err == nil {
			return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "tenant slug already exists"})
		}
		collection, err := e.App.FindCollectionByNameOrId("baas_tenants")
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": "tenant store unavailable"})
		}
		record := core.NewRecord(collection)
		record.Set("slug", req.Slug)
		record.Set("display_name", req.DisplayName)
		record.Set("status", string(tenancy.StatusActive))
		record.Set("plan", strings.TrimSpace(req.Plan))
		if err := e.App.Save(record); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "failed to create tenant"})
		}
		if engine.Registry != nil {
			dirs := config.BaaSDirectories{TenantsRoot: engine.TenantsRoot}
			if err := bootstrap.ApplySchema(e.Request.Context(), dirs, record.Id, engine.Registry); err != nil {
				record.Set("status", string(tenancy.StatusFailed))
				_ = e.App.Save(record)
				return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": "tenant provisioning failed"})
			}
		}
		return e.JSON(http.StatusCreated, map[string]any{"item": tenantSummary{ID: record.Id, Slug: record.GetString("slug"), DisplayName: record.GetString("display_name"), Status: record.GetString("status"), Plan: record.GetString("plan")}})
	}))

	api.POST("/tenants/{slug}/provision", requireSuperuser(func(e *core.RequestEvent) error {
		slug := strings.ToLower(strings.TrimSpace(e.Request.PathValue("slug")))
		record, err := e.App.FindFirstRecordByData("baas_tenants", "slug", slug)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "tenant not found"})
		}
		status := tenancy.Status(record.GetString("status"))
		if status == tenancy.StatusDeleting {
			return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "tenant unavailable"})
		}
		if engine.Registry == nil {
			return e.JSON(http.StatusServiceUnavailable, map[string]any{"status": "error", "message": "engine registry unavailable"})
		}
		dirs := config.BaaSDirectories{TenantsRoot: engine.TenantsRoot}
		if err := bootstrap.ApplySchema(e.Request.Context(), dirs, record.Id, engine.Registry); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": "tenant provisioning failed"})
		}
		if status == tenancy.StatusProvisioning || status == tenancy.StatusFailed {
			record.Set("status", string(tenancy.StatusActive))
			_ = e.App.Save(record)
		}
		return e.JSON(http.StatusOK, map[string]any{"status": "provisioned", "item": tenantSummary{ID: record.Id, Slug: record.GetString("slug"), DisplayName: record.GetString("display_name"), Status: record.GetString("status"), Plan: record.GetString("plan")}})
	}))

	api.GET("/tenants/{slug}", requireAuthenticated(func(e *core.RequestEvent) error { return getTenant(e, engine) }))
	api.PATCH("/tenants/{slug}", requireAuthenticated(func(e *core.RequestEvent) error { return updateTenant(e, engine) }))
	api.GET("/tenants/{slug}/memberships", requireAuthenticated(func(e *core.RequestEvent) error { return listMemberships(e, engine) }))
	api.POST("/tenants/{slug}/memberships", requireAuthenticated(func(e *core.RequestEvent) error { return addMembership(e, engine) }))
	api.PATCH("/tenants/{slug}/memberships/{principal}", requireAuthenticated(func(e *core.RequestEvent) error { return updateMembership(e, engine) }))
	api.DELETE("/tenants/{slug}/memberships/{principal}", requireAuthenticated(func(e *core.RequestEvent) error { return deleteMembership(e, engine) }))
}

func tenantPagination(e *core.RequestEvent, def, max int) (int, int) {
	limit := def
	offset := 0
	if raw := strings.TrimSpace(e.Request.URL.Query().Get("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			limit = v
		}
	}
	if raw := strings.TrimSpace(e.Request.URL.Query().Get("perPage")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			limit = v
		}
	}
	if limit > max {
		limit = max
	}
	if raw := strings.TrimSpace(e.Request.URL.Query().Get("offset")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v >= 0 {
			offset = v
		}
	}
	if raw := strings.TrimSpace(e.Request.URL.Query().Get("page")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			offset = (v - 1) * limit
		}
	}
	return limit, offset
}

func tenantSummaries(records []*core.Record) []tenantSummary {
	items := make([]tenantSummary, 0, len(records))
	for _, r := range records {
		items = append(items, tenantSummary{ID: r.Id, Slug: r.GetString("slug"), DisplayName: r.GetString("display_name"), Status: r.GetString("status"), Plan: r.GetString("plan")})
	}
	return items
}
