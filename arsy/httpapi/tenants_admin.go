package httpapi

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/bridge-allah/a-r-s-y/arsy/tenancy"
	"github.com/bridge-allah/a-r-s-y/core"
)

type membershipSummary struct {
	TenantID  string `json:"tenant_id"`
	Principal string `json:"principal"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	Created   string `json:"created"`
	Updated   string `json:"updated"`
}

var principalIDPattern = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func authorizeTenantAdmin(e *core.RequestEvent, tenant *core.Record) (tenancy.Role, error) {
	if e == nil || e.Auth == nil {
		return "", errors.New("unauthorized")
	}
	if e.HasAdminAuth() {
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

func requireTenantForMember(e *core.RequestEvent) (*core.Record, error) {
	slug := strings.ToLower(strings.TrimSpace(e.Request.PathValue("slug")))
	record, err := e.App.FindFirstRecordByData("baas_tenants", "slug", slug)
	if err != nil {
		return nil, errNotFound("tenant not found")
	}
	return record, nil
}

type apiErr struct {
	code    int
	message string
}

func (e apiErr) Error() string { return e.message }

func errNotFound(msg string) error  { return apiErr{code: http.StatusNotFound, message: msg} }
func errForbidden(msg string) error { return apiErr{code: http.StatusForbidden, message: msg} }
func errConflict(msg string) error  { return apiErr{code: http.StatusConflict, message: msg} }
func errBad(msg string) error       { return apiErr{code: http.StatusBadRequest, message: msg} }
func errUnauthorized(msg string) error {
	return apiErr{code: http.StatusUnauthorized, message: msg}
}

func writeAPIError(e *core.RequestEvent, err error) error {
	var ae apiErr
	if errors.As(err, &ae) {
		return e.JSON(ae.code, map[string]any{"status": "error", "message": ae.message})
	}
	return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": "request failed"})
}

func tenantUnavailable(status tenancy.Status) bool {
	return status == tenancy.StatusSuspended || status == tenancy.StatusDeleting
}

func getTenant(e *core.RequestEvent, _ EngineConfig) error {
	slug := strings.ToLower(strings.TrimSpace(e.Request.PathValue("slug")))
	record, err := e.App.FindFirstRecordByData("baas_tenants", "slug", slug)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "tenant not found"})
	}
	if e.Auth == nil {
		return e.JSON(http.StatusUnauthorized, map[string]any{"status": "error", "message": "authentication required"})
	}
	if !e.HasAdminAuth() {
		reg := tenancy.NewPBMembershipRegistry(e.App)
		if _, err := reg.Membership(e.Request.Context(), record.Id, e.Auth.Id); err != nil {
			return e.JSON(http.StatusForbidden, map[string]any{"status": "error", "message": "tenant access denied"})
		}
	}
	status := tenancy.Status(record.GetString("status"))
	if tenantUnavailable(status) {
		return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "tenant unavailable"})
	}
	return e.JSON(http.StatusOK, map[string]any{"item": tenantSummary{ID: record.Id, Slug: record.GetString("slug"), DisplayName: record.GetString("display_name"), Status: record.GetString("status"), Plan: record.GetString("plan")}})
}

func updateTenant(e *core.RequestEvent, _ EngineConfig) error {
	slug := strings.ToLower(strings.TrimSpace(e.Request.PathValue("slug")))
	record, err := e.App.FindFirstRecordByData("baas_tenants", "slug", slug)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "tenant not found"})
	}
	status := tenancy.Status(record.GetString("status"))
	if status == tenancy.StatusDeleting {
		return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "tenant unavailable"})
	}
	if _, err := authorizeTenantAdmin(e, record); err != nil {
		if err.Error() == "unauthorized" {
			return e.JSON(http.StatusUnauthorized, map[string]any{"status": "error", "message": "authentication required"})
		}
		return e.JSON(http.StatusForbidden, map[string]any{"status": "error", "message": "tenant admin access denied"})
	}
	var req struct {
		DisplayName string `json:"display_name"`
		Plan        string `json:"plan"`
		Status      string `json:"status"`
	}
	if err := decodeServerRequest(e, &req); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "invalid request"})
	}
	changed := false
	if name := strings.TrimSpace(req.DisplayName); name != "" {
		if len(name) > 120 {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "invalid display name"})
		}
		record.Set("display_name", name)
		changed = true
	}
	if req.Plan != "" {
		plan := strings.TrimSpace(req.Plan)
		if len(plan) > 64 {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "invalid plan"})
		}
		record.Set("plan", plan)
		changed = true
	}
	if req.Status != "" {
		newStatus := tenancy.Status(req.Status)
		if !isValidTenantStatus(newStatus) {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "invalid tenant status"})
		}
		if err := tenancy.ValidateTransition(status, newStatus); err != nil {
			return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "invalid tenant status transition"})
		}
		record.Set("status", string(newStatus))
		changed = true
	}
	if !changed {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "no valid fields to update"})
	}
	if err := e.App.Save(record); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "failed to update tenant"})
	}
	return e.JSON(http.StatusOK, map[string]any{"item": tenantSummary{ID: record.Id, Slug: record.GetString("slug"), DisplayName: record.GetString("display_name"), Status: record.GetString("status"), Plan: record.GetString("plan")}})
}

func isValidTenantStatus(s tenancy.Status) bool {
	switch s {
	case tenancy.StatusProvisioning, tenancy.StatusActive, tenancy.StatusSuspended, tenancy.StatusDeleting, tenancy.StatusFailed:
		return true
	default:
		return false
	}
}

func listMemberships(e *core.RequestEvent, _ EngineConfig) error {
	slug := strings.ToLower(strings.TrimSpace(e.Request.PathValue("slug")))
	record, err := e.App.FindFirstRecordByData("baas_tenants", "slug", slug)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "tenant not found"})
	}
	if tenantUnavailable(tenancy.Status(record.GetString("status"))) {
		return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "tenant unavailable"})
	}
	if _, err := authorizeTenantAdmin(e, record); err != nil {
		if err.Error() == "unauthorized" {
			return e.JSON(http.StatusUnauthorized, map[string]any{"status": "error", "message": "authentication required"})
		}
		return e.JSON(http.StatusForbidden, map[string]any{"status": "error", "message": "tenant admin access denied"})
	}
	limit, offset := tenantPagination(e, 50, 200)
	records, err := e.App.FindRecordsByFilter("baas_tenant_memberships", "tenant = {:tenant}", "principal", limit, offset, map[string]any{"tenant": record.Id})
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": "failed to list memberships"})
	}
	items := make([]membershipSummary, 0, len(records))
	for _, r := range records {
		items = append(items, membershipSummary{TenantID: record.Id, Principal: r.GetString("principal"), Role: r.GetString("role"), Status: r.GetString("status"), Created: r.GetDateTime("created").String(), Updated: r.GetDateTime("updated").String()})
	}
	return e.JSON(http.StatusOK, map[string]any{"items": items, "limit": limit, "offset": offset})
}

func addMembership(e *core.RequestEvent, _ EngineConfig) error {
	slug := strings.ToLower(strings.TrimSpace(e.Request.PathValue("slug")))
	record, err := e.App.FindFirstRecordByData("baas_tenants", "slug", slug)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "tenant not found"})
	}
	if tenantUnavailable(tenancy.Status(record.GetString("status"))) {
		return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "tenant unavailable"})
	}
	if _, err := authorizeTenantAdmin(e, record); err != nil {
		if err.Error() == "unauthorized" {
			return e.JSON(http.StatusUnauthorized, map[string]any{"status": "error", "message": "authentication required"})
		}
		return e.JSON(http.StatusForbidden, map[string]any{"status": "error", "message": "tenant admin access denied"})
	}
	var req struct {
		Principal string `json:"principal"`
		Role      string `json:"role"`
		Status    string `json:"status"`
	}
	if err := decodeServerRequest(e, &req); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "invalid request"})
	}
	req.Principal = strings.TrimSpace(req.Principal)
	if req.Principal == "" || len(req.Principal) > 64 || !principalIDPattern.MatchString(req.Principal) {
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
	if e.Auth != nil {
		m.Set("granted_by", e.Auth.Id)
	}
	if err := e.App.Save(m); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "failed to add membership"})
	}
	return e.JSON(http.StatusCreated, map[string]any{"item": membershipSummary{TenantID: record.Id, Principal: m.GetString("principal"), Role: m.GetString("role"), Status: m.GetString("status"), Created: m.GetDateTime("created").String(), Updated: m.GetDateTime("updated").String()}})
}

func updateMembership(e *core.RequestEvent, _ EngineConfig) error {
	slug := strings.ToLower(strings.TrimSpace(e.Request.PathValue("slug")))
	tenant, err := e.App.FindFirstRecordByData("baas_tenants", "slug", slug)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "tenant not found"})
	}
	if tenantUnavailable(tenancy.Status(tenant.GetString("status"))) {
		return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "tenant unavailable"})
	}
	if _, err := authorizeTenantAdmin(e, tenant); err != nil {
		if err.Error() == "unauthorized" {
			return e.JSON(http.StatusUnauthorized, map[string]any{"status": "error", "message": "authentication required"})
		}
		return e.JSON(http.StatusForbidden, map[string]any{"status": "error", "message": "tenant admin access denied"})
	}
	principal := strings.TrimSpace(e.Request.PathValue("principal"))
	if principal == "" || len(principal) > 64 || !principalIDPattern.MatchString(principal) {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "invalid principal"})
	}
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
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "invalid request"})
	}
	changed := false
	if req.Role != "" {
		if !tenancy.ValidRole(tenancy.Role(req.Role)) {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "invalid role"})
		}
		m.Set("role", req.Role)
		changed = true
	}
	if req.Status != "" {
		if !tenancy.ValidMembershipStatus(tenancy.MembershipStatus(req.Status)) {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "invalid status"})
		}
		m.Set("status", req.Status)
		changed = true
	}
	if !changed {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "no valid fields to update"})
	}
	if err := e.App.Save(m); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "failed to update membership"})
	}
	return e.JSON(http.StatusOK, map[string]any{"item": membershipSummary{TenantID: tenant.Id, Principal: m.GetString("principal"), Role: m.GetString("role"), Status: m.GetString("status"), Created: m.GetDateTime("created").String(), Updated: m.GetDateTime("updated").String()}})
}

func deleteMembership(e *core.RequestEvent, _ EngineConfig) error {
	slug := strings.ToLower(strings.TrimSpace(e.Request.PathValue("slug")))
	tenant, err := e.App.FindFirstRecordByData("baas_tenants", "slug", slug)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "tenant not found"})
	}
	if tenantUnavailable(tenancy.Status(tenant.GetString("status"))) {
		return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "tenant unavailable"})
	}
	if _, err := authorizeTenantAdmin(e, tenant); err != nil {
		if err.Error() == "unauthorized" {
			return e.JSON(http.StatusUnauthorized, map[string]any{"status": "error", "message": "authentication required"})
		}
		return e.JSON(http.StatusForbidden, map[string]any{"status": "error", "message": "tenant admin access denied"})
	}
	principal := strings.TrimSpace(e.Request.PathValue("principal"))
	if principal == "" || len(principal) > 64 || !principalIDPattern.MatchString(principal) {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "invalid principal"})
	}
	records, err := e.App.FindRecordsByFilter("baas_tenant_memberships", "tenant = {:t} && principal = {:p}", "", 1, 0, map[string]any{"t": tenant.Id, "p": principal})
	if err != nil || len(records) == 0 {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "membership not found"})
	}
	if err := e.App.Delete(records[0]); err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": "failed to delete membership"})
	}
	return e.JSON(http.StatusOK, map[string]any{"status": "deleted", "principal": principal})
}

var _ = errors.New
