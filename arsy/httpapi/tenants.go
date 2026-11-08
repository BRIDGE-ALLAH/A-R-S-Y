package httpapi

import (
	"errors"
	"net/http"
	"regexp"
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
	api.GET("/tenants", requireSuperuser(func(e *core.RequestEvent) error {
		records, err := e.App.FindRecordsByFilter("baas_tenants", "1=1", "slug", 200, 0)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": "failed to list tenants"})
		}
		items := make([]tenantSummary, 0, len(records))
		for _, r := range records {
			items = append(items, tenantSummary{ID: r.Id, Slug: r.GetString("slug"), DisplayName: r.GetString("display_name"), Status: r.GetString("status"), Plan: r.GetString("plan")})
		}
		return e.JSON(http.StatusOK, map[string]any{"items": items})
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
}

var _ = errors.New

var _ = errors.New

var _ = errors.New
