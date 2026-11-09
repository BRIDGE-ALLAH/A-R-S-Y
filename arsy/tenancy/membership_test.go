package tenancy

import "testing"

func TestValidRoleAndPermissions(t *testing.T) {
	for _, role := range []Role{RoleOwner, RoleAdmin, RoleDeveloper, RoleViewer} {
		if !ValidRole(role) {
			t.Fatalf("expected %q valid", role)
		}
	}
	if ValidRole("member") || ValidRole("") || ValidRole("superadmin") {
		t.Fatal("unexpected valid role")
	}
	if !RoleAllowsWrite(RoleOwner) || !RoleAllowsWrite(RoleAdmin) || !RoleAllowsWrite(RoleDeveloper) {
		t.Fatal("expected owner/admin/developer to write")
	}
	if RoleAllowsWrite(RoleViewer) {
		t.Fatal("viewer must not write")
	}
	if !RoleAllowsMembershipAdmin(RoleOwner) || !RoleAllowsMembershipAdmin(RoleAdmin) {
		t.Fatal("owner/admin should manage memberships")
	}
	if RoleAllowsMembershipAdmin(RoleDeveloper) || RoleAllowsMembershipAdmin(RoleViewer) {
		t.Fatal("developer/viewer must not manage memberships")
	}
}
