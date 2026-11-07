# Application Engine Architecture Overview

**Status:** Planned / In Development

## Purpose

The Application Engine (AENG) provides a metadata-driven, multi-tenant application runtime built on PocketBase. It enables declarative model definitions, security rules, migrations, and extensibility through add-ons while maintaining strict tenant isolation.

## Architecture Layers

Dependency flows bottom-up. Each layer depends only on layers below it.

1. **PocketBase Storage & Core Services** - SQLite, Collections, Records, Auth, Hooks
2. **Tenant Isolation Boundary (BAAS)** - Dedicated database per tenant, secure routing
3. **Schema Migration & Persistence** - Unit of Work, ORM, Query Compiler
4. **Module & Model Registry** - Field/Relation metadata, validation
5. **Security Layer** - Model permissions, field visibility, record rules
6. **Application Runtime** - Views, Actions, Workflows
7. **Add-on Lifecycle** - Registry, dependency resolution, seed data

## Glossary

### Application Engine Terms

- **Model:** A typed entity with fields, relations, and validation rules
- **Field:** A typed attribute of a model
- **Relation:** A typed link between models (one-to-one, one-to-many, many-to-many)
- **Module:** A self-contained unit of application functionality with models and security
- **Add-on:** An installable, versioned module
- **View:** Declarative UI metadata defining forms, lists, menus
- **Action:** A server-side operation bound to a model
- **Seed Data:** Idempotent initialization data
- **Migration:** A versioned, ordered schema change
- **Query AST:** A validated, type-safe representation of filter/sort logic
- **Record Rule:** A server-side filter limiting which records a user can access
- **Unit of Work:** A transaction boundary grouping multiple operations

### BaaS (Tenancy) Terms

- **Tenant:** An isolated customer/organization with its own database and data
- **Control Plane:** Shared PocketBase database storing tenant registry and memberships
- **Data Plane:** Set of isolated tenant databases (one SQLite file per tenant)
- **Tenant Context:** Authenticated tenant identity and user role resolved from request
- **Membership:** Link between user and tenant with a role (owner, admin, developer, viewer)
- **Provisioning:** Idempotent workflow to create tenant, database, and initial schema
- **Datastore Provider:** Engine-neutral interface for tenant database access
- **Isolation Boundary:** Enforcement point ensuring cross-tenant data access is prevented

### Related Terms

- **Company:** A business entity within a tenant's data (NOT the same as a tenant)
- **User:** An authenticated principal; users can be members of multiple tenants
- **Collection:** PocketBase term for a table; A-R-S-Y uses "model" at application level
- **Record:** PocketBase term for a row; A-R-S-Y uses "record" at persistence layer

## PocketBase Reuse vs. A-R-S-Y Ownership

### What A-R-S-Y Reuses from PocketBase

- SQLite storage engine
- Collection/Record CRUD operations
- Built-in authentication (password, OAuth2, MFA, OTP)
- File uploads and storage
- Real-time subscriptions
- Admin UI for development
- Hooks (lifecycle events)
- Migration runner for system tables

### What A-R-S-Y Adds

- Multi-Tenancy: Dedicated database per tenant
- Model Registry: Static Go structs mapped to metadata
- Declarative Security: Model and field permissions
- Add-on System: Module manifests and dependency resolution
- Typed Relations: One-to-one, one-to-many, many-to-many with cascade rules
- Application Views: Declarative UI metadata
- Safe Query Compiler: AST-based query builder
- Versioned Migrations: Application-level schema evolution
- Unit of Work: Transaction boundaries

### What A-R-S-Y Does NOT Do

- Does NOT fork or modify PocketBase source code
- Does NOT replace PocketBase migrations for core tables
- Does NOT introduce a generic ORM that bypasses PocketBase
- Does NOT partition a shared database for tenants
- Does NOT support PostgreSQL multi-tenancy in MVP-1

## Implementation Status

- **Model Registry:** ✅ Completed (#124-126)
- **Tenant Isolation:** 📋 Planned (#77-94)
- **Security Layer:** 📋 Planned (#116, #127, #136)
- **Application Runtime:** 📋 Planned (#117, #118, #137)
- **Add-on Lifecycle:** 📋 Planned (#112-115, #134, #135)

## Related Documentation

- BaaS Architecture: docs/baas-architecture.md
- MVP-1 Plan: .plan/task-mvp-1.md
- AENG Issues: GitHub issues #97-137
- BAAS Issues: GitHub issues #77-96

---

**Last Updated:** 2026-11-07
**Status:** Living Document
