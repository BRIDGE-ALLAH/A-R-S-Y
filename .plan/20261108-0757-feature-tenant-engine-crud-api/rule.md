# Rules

- No fake data/success: operations must execute real `persist.Store` + real tenant DB.
- Tenant isolation is mandatory: resolve tenant from request context/header/slug mapping already supported; never allow cross-tenant access.
- Require explicit `allowedFields` for writes/queries; never accept arbitrary column names or raw SQL.
- Superuser/auth guard required on control-plane endpoints; return sanitized errors (no internal SQL/paths/secrets).
- Keep diffs focused; do not refactor unrelated modules.
- All claims backed by command evidence.
