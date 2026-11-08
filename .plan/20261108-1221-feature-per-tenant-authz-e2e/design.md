# Design

Use existing auth context (`e.Auth`, superuser check). For engine CRUD, resolve tenant as before, then authorize: allow if superuser or if caller is a verified member/owner of the tenant using control-plane records (new `baas_tenant_members` or `baas_tenants.owner` strategy — inspect current schema first; implement minimal safe strategy). Map failures to 403 without leaking other tenants' data where appropriate.

Add in-process E2E test in `arsy/httpapi` (or `arsy/e2e`) using `tests.NewTestApp` + mux + `RegisterWithEngine` + real tenant directories to satisfy "real frontend-backend wiring/persistence/E2E flow" within CI-safe environment (no external managed host).
