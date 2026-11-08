# Rules

1. Additive, backward-safe behavior: existing superuser workflows must keep working.
2. Never leak tenant existence/data across tenants to unauthorized callers; prefer safe not-found/forbidden responses per existing patterns.
3. No secrets/credentials in responses or logs.
4. Add regression/feature tests for authz decisions and E2E flow.
5. Do not commit/push/tag/release/deploy.
