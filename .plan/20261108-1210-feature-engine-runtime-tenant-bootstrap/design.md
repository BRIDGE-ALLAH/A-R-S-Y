# Design

## Context
Engine CRUD handlers and `RegisterWithEngine` exist but are not wired in `examples/base/main.go`; no runtime registry, no tenant provisioning API, and no schema for starter models in tenant DBs.

## Approach
- Add `arsy/engine/runtime` (or `arsy/runtime`) bootstrap: build `registry.InMemoryRegistry`, register starter descriptors (e.g. `base.company` -> `base_company`), and provide a tenant schema migrator that creates required tables/columns idempotently for registered models.
- Use existing `config.EnsureBaaSDirectories` and `tenancy` migration patterns where practical; keep schema creation explicit and safe (validated identifiers only).
- Add HTTP handlers under `/tenants` (superuser) that manage `baas_tenants` records and call provisioning.
- Wire entrypoint: ensure BaaS dirs, build engine config, `httpapi.RegisterWithEngine`.

## Trade-offs
- Starter schema is intentionally narrow and idempotent; dynamic arbitrary model schema lifecycle remains future work.
- Tenant auth remains superuser-scoped in this slice.
