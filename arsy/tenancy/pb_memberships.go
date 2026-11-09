package tenancy

import (
	"context"
	"errors"
	"strings"

	"github.com/bridge-allah/a-r-s-y/core"
)

// PBMembershipRegistry resolves tenant memberships from the control-plane
// `baas_tenant_memberships` collection. Authorization decisions must always be
// made server-side using the resolved tenant ID and authenticated principal ID.
type PBMembershipRegistry struct {
	app core.App
}

func NewPBMembershipRegistry(app core.App) *PBMembershipRegistry {
	return &PBMembershipRegistry{app: app}
}

func (r *PBMembershipRegistry) IsMember(ctx context.Context, tenantID, userID string) (bool, error) {
	_, err := r.Membership(ctx, tenantID, userID)
	if err != nil {
		if errors.Is(err, ErrMembershipNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *PBMembershipRegistry) Role(ctx context.Context, tenantID, userID string) (Role, error) {
	m, err := r.Membership(ctx, tenantID, userID)
	if err != nil {
		return "", err
	}
	return m.Role, nil
}

func (r *PBMembershipRegistry) Membership(ctx context.Context, tenantID, userID string) (Member, error) {
	if r == nil || r.app == nil {
		return Member{}, errors.New("tenancy: membership registry app is required")
	}
	tenantID = strings.TrimSpace(tenantID)
	userID = strings.TrimSpace(userID)
	if tenantID == "" || userID == "" {
		return Member{}, ErrMembershipNotFound
	}
	records, err := r.app.FindRecordsByFilter(
		"baas_tenant_memberships",
		"tenant = {:tenant} && principal = {:principal} && status = {:status}",
		"",
		1,
		0,
		map[string]any{"tenant": tenantID, "principal": userID, "status": string(MembershipActive)},
	)
	if err != nil || len(records) == 0 {
		return Member{}, ErrMembershipNotFound
	}
	_ = ctx
	record := records[0]
	role := Role(record.GetString("role"))
	if !ValidRole(role) {
		return Member{}, ErrMembershipNotFound
	}
	status := MembershipStatus(record.GetString("status"))
	if !ValidMembershipStatus(status) || status != MembershipActive {
		return Member{}, ErrMembershipNotFound
	}
	return Member{UserID: userID, TenantID: tenantID, Role: role, Status: status}, nil
}

var _ MembershipRegistry = (*PBMembershipRegistry)(nil)
