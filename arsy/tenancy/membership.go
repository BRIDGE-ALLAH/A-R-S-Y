 package tenancy

 import "context"

 // Role is a tenant-scoped user role.
 type Role string

 const (
 	RoleOwner  Role = "owner"
 	RoleAdmin  Role = "admin"
 	RoleMember Role = "member"
 	RoleViewer Role = "viewer"
 )

 // Member describes a user's membership within a tenant.
 type Member struct {
 	UserID   string
 	TenantID string
 	Role     Role
 }

 // MembershipRegistry is the read interface for tenant membership records.
 type MembershipRegistry interface {
 	// IsMember reports whether userID belongs to tenantID with any role.
 	IsMember(ctx context.Context, tenantID, userID string) (bool, error)
 	// Role returns the role of userID within tenantID, or an error if not a member.
 	Role(ctx context.Context, tenantID, userID string) (Role, error)
 }
