package httpapi

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/bridge-allah/a-r-s-y/arsy/tenancy"
	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tools/types"
)

type membershipSummary struct {
	TenantID  string `json:"tenant_id"`
	Principal string `json:"principal"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	Created   string `json:"created"`
	Updated   string `json:"updated"`
}

func authorizeTenantAdmin(e *core.RequestEvent, tenant *core.Record) (tenancy.Role, error) {
	if e == nil || e.Auth == nil {
		return "", errors.New("unauthorized")
	}
	if e.HasSuperuserAuth() {
		return tenancy.RoleOwner, nil
	}
	reg := tenancy.NewPBMembershipRegistry(e.App)
	member, err := reg.Membership(e.Request.Context(), tenant.Id, e.Auth.Id)
	if err != nil {
		return "", tenancy.ErrForbidden
	}
	if !tenancy.RoleAllowsMembershipAdmin(member.Role) {
		return "", tenancy.ErrForbidden
	}
	return member.Role, nil
}

func getTenant(e *core.RequestEvent, engine EngineConfig) error {
	slug := strings.ToLower(strings.TrimSpace(e.Request.PathValue("slug")))
	record, err := e.App.FindFirstRecordByData("baas_tenants", "slug", slug)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "tenant not found"})
	}
	if !e.HasSuperuserAuth() {
		reg := tenancy.NewPBMembershipRegistry(e.App)
		if _, err := reg.Membership(e.Request.Context(), record.Id, e.Auth.Id); err != nil {
			return e.JSON(http.StatusForbidden, map[string]any{"status": "error", "message": "tenant access denied"})
		}
	}
	status := tenancy.Status(record.GetString("status"))
	if status == tenancy.StatusSuspended || status == tenancy.StatusDeleting {
		return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "tenant unavailable"})
	}
	return e.JSON(http.StatusOK, map[string]any{"item": tenantSummary{ID: record.Id, Slug: record.GetString("slug"), DisplayName: record.GetString("display_name"), Status: record.GetString("status"), Plan: record.GetString("plan")}})
}

func updateTenant(e *core.RequestEvent, engine EngineConfig) error {
	slug := strings.ToLower(strings.TrimSpace(e.Request.PathValue("slug")))
	record, err := e.App.FindFirstRecordByData("baas_tenants", "slug", slug)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "tenant not found"})
	}
	if _, err := authorizeTenantAdmin(e, record); err != nil {
		return e.JSON(http.StatusForbidden, map[string]any{"status": "error", "message": "tenant admin access denied"})
	}
	var req struct {
		DisplayName string `json:"display_name"`
		Plan        string `json:"plan"`
		Status      string `json:"status"`
	}
	if err := decodeServerRequest(e, &req); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": err.Error()})
	}
	if name := strings.TrimSpace(req.DisplayName); name != "" {
		if len(name) > 120 {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "invalid display name"})
		}
		record.Set("display_name", name)
	}
	if req.Plan != "" {
		record.Set("plan", strings.TrimSpace(req.Plan))
	}
	if req.Status != "" {
		newStatus := tenancy.Status(req.Status)
		if err := tenancy.ValidateTransition(tenancy.Status(record.GetString("status")), newStatus); err != nil {
			return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": err.Error()})
		}
		record.Set("status", string(newStatus))
	}
	if err := e.App.Save(record); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "failed to update tenant"})
	}
	return e.JSON(http.StatusOK, map[string]any{"item": tenantSummary{ID: record.Id, Slug: record.GetString("slug"), DisplayName: record.GetString("display_name"), Status: record.GetString("status"), Plan: record.GetString("plan")}})
}

func listMemberships(e *core.RequestEvent, engine EngineConfig) error {
	slug := strings.ToLower(strings.TrimSpace(e.Request.PathValue("slug")))
	record, err := e.App.FindFirstRecordByData("baas_tenants", "slug", slug)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "tenant not found"})
	}
	if _, err := authorizeTenantAdmin(e, record); err != nil {
		return e.JSON(http.StatusForbidden, map[string]any{"status": "error", "message": "tenant admin access denied"})
	}
	records, err := e.App.FindRecordsByFilter("baas_tenant_memberships", "tenant = {:tenant}", "principal", 200, 0, map[string]any{"tenant": record.Id})
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": "failed to list memberships"})
	}
	items := make([]membershipSummary, 0, len(records))
	for _, r := range records {
		items = append(items, membershipSummary{TenantID: record.Id, Principal: r.GetString("principal"), Role: r.GetString("role"), Status: r.GetString("status"), Created: r.GetDateTime("created").String(), Updated: r.GetDateTime("updated").String()})
	}
	return e.JSON(http.StatusOK, map[string]any{"items": items})
}

func addMembership(e *core.RequestEvent, engine EngineConfig) error {
	slug := strings.ToLower(strings.TrimSpace(e.Request.PathValue("slug")))
	record, err := e.App.FindFirstRecordByData("baas_tenants", "slug", slug)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "tenant not found"})
	}
	if _, err := authorizeTenantAdmin(e, record); err != nil {
		return e.JSON(http.StatusForbidden, map[string]any{"status": "error", "message": "tenant admin access denied"})
	}
	var req struct {
		Principal string `json:"principal"`
		Role      string `json:"role"`
		Status    string `json:"status"`
	}
	if err := decodeServerRequest(e, &req); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": err.Error()})
	}
	req.Principal = strings.TrimSpace(req.Principal)
	if req.Principal == "" || !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(req.Principal) {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "invalid principal"})
	}
	if !tenancy.ValidRole(tenancy.Role(req.Role)) {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "invalid role"})
	}
	if req.Status == "" {
		req.Status = string(tenancy.MembershipActive)
	}
	if !tenancy.ValidMembershipStatus(tenancy.MembershipStatus(req.Status)) {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "invalid status"})
	}
	existing, err := e.App.FindRecordsByFilter("baas_tenant_memberships", "tenant = {:t} && principal = {:p}", "", 1, 0, map[string]any{"t": record.Id, "p": req.Principal})
	if err == nil && len(existing) > 0 {
		return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "membership already exists"})
	}
	collection, err := e.App.FindCollectionByNameOrId("baas_tenant_memberships")
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": "membership store unavailable"})
	}
	m := core.NewRecord(collection)
	m.Set("tenant", record.Id)
	m.Set("principal", req.Principal)
	m.Set("role", req.Role)
	m.Set("status", req.Status)
	m.Set("granted_by", e.Auth.Id)
	if err := e.App.Save(m); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "failed to add membership"})
	}
	return e.JSON(http.StatusCreated, map[string]any{"item": membershipSummary{TenantID: record.Id, Principal: m.GetString("principal"), Role: m.GetString("role"), Status: m.GetString("status"), Created: m.GetDateTime("created").String(), Updated: m.GetDateTime("updated").String()}})
}

func updateMembership(e *core.RequestEvent, engine EngineConfig) error {
	slug := strings.ToLower(strings.TrimSpace(e.Request.PathValue("slug")))
	tenant, err := e.App.FindFirstRecordByData("baas_tenants", "slug", slug)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "tenant not found"})
	}
	if _, err := authorizeTenantAdmin(e, tenant); err != nil {
		return e.JSON(http.StatusForbidden, map[string]any{"status": "error", "message": "tenant admin access denied"})
	}
	principal := strings.TrimSpace(e.Request.PathValue("principal"))
	records, err := e.App.FindRecordsByFilter("baas_tenant_memberships", "tenant = {:t} && principal = {:p}", "", 1, 0, map[string]any{"t": tenant.Id, "p": principal})
	if err != nil || len(records) == 0 {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "membership not found"})
	}
	m := records[0]
	var req struct {
		Role   string `json:"role"`
		Status string `json:"status"`
	}
	if err := decodeServerRequest(e, &req); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": err.Error()})
	}
	if req.Role != "" {
		if !tenancy.ValidRole(tenancy.Role(req.Role)) {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "invalid role"})
		}
		m.Set("role", req.Role)
	}
	if req.Status != "" {
		if !tenancy.ValidMembershipStatus(tenancy.MembershipStatus(req.Status)) {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "invalid status"})
		}
		m.Set("status", req.Status)
	}
	if err := e.App.Save(m); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "failed to update membership"})
	}
	return e.JSON(http.StatusOK, map[string]any{"item": membershipSummary{TenantID: tenant.Id, Principal: m.GetString("principal"), Role: m.GetString("role"), Status: m.GetString("status"), Created: m.GetDateTime("created").String(), Updated: m.GetDateTime("updated").String()}})
}

func deleteMembership(e *core.RequestEvent, engine EngineConfig) error {
	slug := strings.ToLower(strings.TrimSpace(e.Request.PathValue("slug")))
	tenant, err := e.App.FindFirstRecordByData("baas_tenants", "slug", slug)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "tenant not found"})
	}
	if _, err := authorizeTenantAdmin(e, tenant); err != nil {
		return e.JSON(http.StatusForbidden, map[string]any{"status": "error", "message": "tenant admin access denied"})
	}
	principal := strings.TrimSpace(e.Request.PathValue("principal"))
	records, err := e.App.FindRecordsByFilter("baas_tenant_memberships", "tenant = {:t} && principal = {:p}", "", 1, 0, map[string]any{"t": tenant.Id, "p": principal})
	if err != nil || len(records) == 0 {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "membership not found"})
	}
	if err := e.App.Delete(records[0]); err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": "failed to delete membership"})
	}
	return e.JSON(http.StatusOK, map[string]any{"status": "deleted", "principal": principal})
}

func containsMigrationIndex(indexes types.JSONArray[string], name string) bool {
	for _, idx := range indexes {
		if strings.Contains(idx, name) {
			return true
		}
	}
	return false
}

var _ = errors.New

