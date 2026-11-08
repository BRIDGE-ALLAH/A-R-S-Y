package tenancy

import "errors"

// Sentinel errors for the tenancy layer.
var (
	// ErrTenantNotFound is returned when the requested tenant does not exist.
	ErrTenantNotFound = errors.New("tenancy: tenant not found")

	// ErrTenantSuspended is returned when an operation is rejected because the
	// tenant is suspended.
	ErrTenantSuspended = errors.New("tenancy: tenant is suspended")

	// ErrTenantDeleting is returned when the tenant is being deleted.
	ErrTenantDeleting = errors.New("tenancy: tenant is being deleted")

	// ErrTenantNotActive is returned when a tenant exists but is not active
	// (e.g. provisioning or failed).
	ErrTenantNotActive = errors.New("tenancy: tenant is not active")

	// ErrDatastoreClosed is returned when the datastore has already been closed.
	ErrDatastoreClosed = errors.New("tenancy: datastore is closed")

	// ErrInvalidTenantID is returned for malformed tenant identifiers.
	ErrInvalidTenantID = errors.New("tenancy: invalid tenant ID")

	// ErrInvalidChecksum is returned when a backup checksum does not match.
	ErrInvalidChecksum = errors.New("tenancy: invalid checksum")

	// ErrMembershipNotFound is returned when an active tenant membership does not exist.
	ErrMembershipNotFound = errors.New("tenancy: membership not found")

	// ErrForbidden is returned when a principal is authenticated but not authorized.
	ErrForbidden = errors.New("tenancy: forbidden")
)
