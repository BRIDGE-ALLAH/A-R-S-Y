package tenancy

import "context"

// Role is a tenant-scoped user role.
// Permission matrix (server-side, minimal implemented):
// - owner: tenant admin, membership management, engine read/write/query.
// - admin: membership management (except ownership transfer not implemented), engine read/write/query.
// - developer: engine read/write/query.
// - viewer: engine read/query only.
// Superusers are control-plane admins and bypass membership checks.
type Role string

const (
	RoleOwner     Role = "owner"
	RoleAdmin     Role = "admin"
	RoleDeveloper Role = "developer"
	RoleViewer    Role = "viewer"
)

// MembershipStatus controls access.
type MembershipStatus string

const (
	MembershipActive   MembershipStatus = "active"
	MembershipInactive MembershipStatus = "inactive"
)

// Member describes a user's membership within a tenant.
type Member struct {
	UserID   string
	TenantID string
	Role     Role
	Status   MembershipStatus
}

// MembershipRegistry is the read interface for tenant membership records.
type MembershipRegistry interface {
	// IsMember reports whether userID belongs to tenantID with an active membership.
	IsMember(ctx context.Context, tenantID, userID string) (bool, error)
	// Role returns the active role of userID within tenantID, or an error if not an active member.
	Role(ctx context.Context, tenantID, userID string) (Role, error)
	// Membership returns the active membership or ErrMembershipNotFound.
	Membership(ctx context.Context, tenantID, userID string) (Member, error)
}

func ValidRole(role Role) bool {
	switch role {
	case RoleOwner, RoleAdmin, RoleDeveloper, RoleViewer:
		return true
	default:
		return false
	}
}

func ValidMembershipStatus(status MembershipStatus) bool {
	return status == MembershipActive || status == MembershipInactive
}

func RoleAllowsWrite(role Role) bool {
	switch role {
	case RoleOwner, RoleAdmin, RoleDeveloper:
		return true
	default:
		return false
	}
}

func RoleAllowsMembershipAdmin(role Role) bool {
	return role == RoleOwner || role == RoleAdmin
}
